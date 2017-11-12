package layer

import (
	"encoding/json"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	goctx "golang.org/x/net/context"
)

func execute(layer *Layer, sql string) ([]map[string]interface{}, error) {
	data := make([]map[string]interface{}, 0)
	result, err := layer.session.Execute(goctx.Background(), sql)
	if err != nil {
		return data, err
	}

	if len(result) > 0 {
		fields := make([]string, 0)
		_fields, _ := result[0].Fields()
		for _, _field := range _fields {
			fields = append(fields, _field.ColumnAsName.L)
		}
		rows := result[0]
		_row, _ := rows.Next()
		for _row != nil {
			row := map[string]interface{}{}
			for index, element := range _row.Data {
				value := element.GetValue()
				switch value.(type) {
				case []byte:
					row[fields[index]] = string(value.([]byte))
				default:
					row[fields[index]] = value
				}
			}
			data = append(data, row)
			_row, _ = rows.Next()
		}
		rows.Close()
	}
	return data, nil
}

//getPostParams Get the parameters sent by the post method in an http request
func getPostParams(r *http.Request) url.Values {
	switch {
	case r.Header.Get("Content-Type") == "application/json":
		params := map[string]interface{}{}
		result := url.Values{}
		decoder := json.NewDecoder(r.Body)
		decoder.Decode(&params)
		for k, v := range params {
			if reflect.ValueOf(v).Kind().String() == "string" {
				result.Set(k, v.(string))
			}
		}
		return result
	case r.Header.Get("Content-Type") == "application/x-www-form-urlencoded":
		r.ParseForm()
		return r.Form
	case strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data"):
		r.ParseMultipartForm(int64(10 * 1000))
		return r.Form
	}
	return url.Values{}
}
