package layer

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/juju/errors"
	"github.com/pingcap/tidb/ast"
	"github.com/pingcap/tidb/config"
	"github.com/pingcap/tidb/kv"
	"github.com/pingcap/tidb/server"
	"github.com/rs/zerolog/log"

	"github.com/gorilla/mux"
	"github.com/pingcap/tidb"
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
	router.HandleFunc("/query", query(layer)).Methods("POST")
	router.PathPrefix("/").HandlerFunc(root)

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
		result, err := layer.session.Execute(query)
		if err == nil {
			err = fmt.Errorf("")
		}
		//data := make([][]interface{}, 0)
		for _, col := range result {
			//tmp := make([]interface{}, 0)
			fmt.Println(col.Fields())
			row, _ := col.Next()
			for _, element := range row.Data {
				log.Print(element.GetValue())
				fmt.Println(element)
			}
			/*for _, col := range row {
				tmp = append(tmp, col.GetInterface())
			}
			data = append(data, tmp)*/
		}
		var row []*ast.ResultField
		//var rowType string
		if len(result) > 0 {
			row, _ = result[len(result)-1].Fields()
			//rowType = fmt.Sprintf("%T", )
		}

		response, _ := json.MarshalIndent(map[string]interface{}{
			"type":   fmt.Sprintf("%T", result),
			"result": row,
			"err":    err.Error(),
			"query":  query,
		}, "", "\t")
		w.Write(response)
	}
}

func root(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hola mundo\n")
}
