package layer

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pingcap/tidb/server"
)

type Layer struct {
	handler http.Handler
	server  *server.Server
}

func NewLayer(svr *server.Server) *Layer {
	layer := &Layer{}
	router := mux.NewRouter()
	router.HandleFunc("/query", query(layer)).Methods("POST")
	router.PathPrefix("/").HandlerFunc(root)
	layer.handler = router
	layer.server = svr
	return layer
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
		//layer.server.
		fmt.Fprint(w, query)
	}
}

func root(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hola mundo\n")
}
