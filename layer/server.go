package layer

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type Layer struct {
	handler http.Handler
}

func NewLayer() *Layer {
	router := mux.NewRouter()
	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hola mundo")
	})
	return &Layer{router}
}

func (el *Layer) Run() {
	go func() {
		http.ListenAndServe(":5000", el.handler)
	}()
}
