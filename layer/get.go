package layer

import (
	"encoding/json"
	"net/http"

	"github.com/go-xorm/builder"
	"github.com/gorilla/mux"
)

func (el *Layer) get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	table := vars["table"]
	sql, _, _ := builder.Select("*").From(table).ToSQL()
	data, err := execute(el, sql)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	response, _ := json.MarshalIndent(map[string]interface{}{
		"success": true,
		"data":    data,
	}, "", "\t")
	w.Write(response)
}
