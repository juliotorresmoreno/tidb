package layer

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/juju/errors"
	"github.com/pingcap/tidb"
	"github.com/pingcap/tidb/config"
	"github.com/pingcap/tidb/kv"
	"github.com/pingcap/tidb/server"
)

type Layer struct {
	cfg        *config.Config
	handler    http.Handler
	queryCtx   server.QueryCtx
	driver     server.IDriver
	storage    kv.Storage
	session    tidb.Session
	capability uint32
	collation  int
	dbname     string
}

func NewLayer(cfg *config.Config, store kv.Storage) (*Layer, error) {
	layer := &Layer{}
	router := mux.NewRouter()
	router.HandleFunc("/query", layer.handlerSQL).Methods("POST")
	router.HandleFunc("/{database}/{table}", layer.handlerSelect).Methods("GET")
	router.HandleFunc("/{database}/{table}", layer.handlerInsert).Methods("PUT")
	router.HandleFunc("/{database}/{table}/{id}", layer.handlerUpdate).Methods("PATCH")
	router.HandleFunc("/{database}/{table}/{id}", layer.handlerDelete).Methods("DELETE")
	router.PathPrefix("/").HandlerFunc(root).Methods("GET")

	layer.cfg = cfg
	layer.storage = store
	layer.driver = server.NewTiDBDriver(layer.storage)
	layer.handler = router

	layer.capability = 1811077
	layer.collation = 33
	layer.dbname = ""
	err := layer.OpenCtx()
	return layer, err
}

// OpenCtx implements IDriver.
func (el *Layer) OpenCtx() error {
	var err error
	el.session, err = tidb.CreateSession(el.storage)
	if err != nil {
		return errors.Trace(err)
	}
	err = el.session.SetCollation(el.collation)
	if err != nil {
		return errors.Trace(err)
	}
	el.session.SetClientCapability(el.capability)
	el.queryCtx = server.NewTiDBContext(el.session, el.dbname)
	return nil
}

//Run run server api
func (el *Layer) Run() {
	go func() {
		http.ListenAndServe(":5000", el.handler)
	}()
}

func root(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "OK\n")
}
