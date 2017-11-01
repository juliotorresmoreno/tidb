package layer

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/juju/errors"
	"github.com/pingcap/tidb/config"
	"github.com/pingcap/tidb/kv"

	"github.com/gorilla/mux"
	"github.com/pingcap/tidb"
	"github.com/pingcap/tidb/server"
)

type Layer struct {
	cfg      *config.Config
	handler  http.Handler
	server   *server.Server
	queryCtx server.QueryCtx
	driver   server.IDriver
	storage  kv.Storage
}

func NewLayer(svr *server.Server, cfg *config.Config) *Layer {
	layer := &Layer{}
	router := mux.NewRouter()
	router.HandleFunc("/query", query(layer)).Methods("POST")
	router.PathPrefix("/").HandlerFunc(root)

	var tlsStatePtr *tls.ConnectionState
	fullPath := fmt.Sprintf("%s://%s", cfg.Store, cfg.Path)
	layer.cfg = cfg
	layer.storage, _ = tidb.NewStore(fullPath)
	layer.driver = server.NewTiDBDriver(layer.storage)
	layer.handler = router
	layer.server = svr
	layer.queryCtx, _ = OpenCtx(
		layer.storage,
		uint32(1811077),
		uint8(33),
		"",
	)
	return layer
}

// OpenCtx implements IDriver.
func OpenCtx(store kv.Storage, capability uint32, collation uint8, dbname string) (server.QueryCtx, error) {
	session, err := tidb.CreateSession(store)
	if err != nil {
		return nil, errors.Trace(err)
	}
	err = session.SetCollation(int(collation))
	if err != nil {
		return nil, errors.Trace(err)
	}
	session.SetClientCapability(capability)
	tc := &server.TiDBContext{
		session:   session,
		currentDB: dbname,
		stmts:     make(map[int]*TiDBStatement),
	}
	return tc, nil
}

func (el *Layer) Run() {
	go func() {
		http.ListenAndServe(":5000", el.handler)
	}()
}

func query(layer *Layer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var query string
		if result := getPostParams(r)["query"]; len(result) > 0 {
			query = result[0]
		}
		result, err := layer.queryCtx.Execute(query)
		data, _ := json.Marshal(map[string]interface{}{
			"result": result,
			"err":    err,
		})
		w.Write(data)
	}
}

func root(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hola mundo\n")
}
