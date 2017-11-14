package layer

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/juju/errors"
	"github.com/justinas/alice"
	"github.com/pingcap/tidb"
	"github.com/pingcap/tidb/config"
	"github.com/pingcap/tidb/kv"
	"github.com/pingcap/tidb/server"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
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
	hub        *Hub
}

func NewLayer(cfg *config.Config, store kv.Storage) (*Layer, error) {
	layer := &Layer{}
	layer.hub = &Hub{clients: make(map[string]*user)}
	c := alice.New()
	log := zerolog.New(os.Stdout).With().
		Timestamp().
		Str("role", "TiDB").
		Logger()
	// Install the logger handler with default output on the console
	c = c.Append(hlog.NewHandler(log))

	// Install some provided extra handler to set some request's context fields.
	// Thanks to those handler, all our logs will come with some pre-populated fields.
	c = c.Append(hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		hlog.FromRequest(r).Info().
			Str("method", r.Method).
			Str("url", r.URL.String()).
			Int("status", status).
			Int("size", size).
			Dur("duration", duration).
			Msg("")
	}))
	c = c.Append(hlog.RemoteAddrHandler("ip"))
	c = c.Append(hlog.UserAgentHandler("user_agent"))
	c = c.Append(hlog.RefererHandler("referer"))
	c = c.Append(hlog.RequestIDHandler("req_id", "Request-Id"))
	c = c.Append(cors.New(cors.Options{OptionsPassthrough: true}).Handler)

	router := mux.NewRouter()
	router.HandleFunc("/query", layer.handlerSQL).Methods("POST")
	router.HandleFunc("/{database}/{table}", layer.handlerSelect).Methods("GET")
	router.HandleFunc("/{database}/{table}", layer.handlerInsert).Methods("PUT")
	router.HandleFunc("/{database}/{table}/{id}", layer.handlerUpdate).Methods("PATCH")
	router.HandleFunc("/{database}/{table}/{id}", layer.handlerDelete).Methods("DELETE")
	router.HandleFunc("/{database}/{table}/subscribe", layer.handlerSubscribe).Methods("GET")
	// websocket
	router.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		layer.hub.ServeWs(w, r, "session")
	}).Methods("GET")

	router.PathPrefix("/").HandlerFunc(root).Methods("GET")

	layer.cfg = cfg
	layer.storage = store
	layer.driver = server.NewTiDBDriver(layer.storage)
	layer.handler = c.Then(router)

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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"server": "TiDB 5.7",
	})
}
