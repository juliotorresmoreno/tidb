package layer

import (
	"encoding/json"
	"net/http"

	"github.com/go-xorm/builder"
	"github.com/gorilla/mux"
)

type Row map[string]interface{}

func (el *Layer) handlerSelect(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	database, table := vars["database"], vars["table"]
	data, err := el.Select(database, table)
	if err != nil {
		herror(w, err.Error())
		return
	}
	hsuccess(w, data)
}

func (el *Layer) handlerInsert(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	database := vars["database"]
	table := vars["table"]
	params := builder.Eq{}
	data := getPostParams(r)
	for key := range data {
		params[key] = data.Get(key)
	}
	err := el.Insert(database, table, params)
	if err != nil {
		herror(w, err.Error())
		return
	}
	json.NewEncoder(w).Encode(Row{
		"success": true,
	})
}

func (el *Layer) handlerUpdate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	database := vars["database"]
	table := vars["table"]
	id, _ := vars["id"]
	params := builder.Eq{}
	data := getPostParams(r)
	for key := range data {
		params[key] = data.Get(key)
	}
	err := el.Update(database, table, id, params)
	if err != nil {
		herror(w, err.Error())
		return
	}
	json.NewEncoder(w).Encode(Row{
		"success": true,
	})
}

func (el *Layer) handlerDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	database := vars["database"]
	table := vars["table"]
	id := vars["id"]
	err := el.Delete(database, table, id)
	if err != nil {
		herror(w, err.Error())
		return
	}
	json.NewEncoder(w).Encode(Row{
		"success": true,
	})
}

func hsuccess(w http.ResponseWriter, data []map[string]interface{}) {
	json.NewEncoder(w).Encode(Row{
		"success": true,
		"data":    data,
	})
}

func herror(w http.ResponseWriter, message string) {
	json.NewEncoder(w).Encode(Row{
		"success": false,
		"message": message,
	})
}
