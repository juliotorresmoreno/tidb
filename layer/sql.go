package layer

import (
	"encoding/json"
	"net/http"
)

func (el *Layer) sql(w http.ResponseWriter, r *http.Request) {
	query := getPostParams(r).Get("query")
	if len(query) == 0 {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
		})
		return
	}
	data, err := execute(el, query)
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
