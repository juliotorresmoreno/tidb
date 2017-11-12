package layer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/pingcap/tidb/ast"
	goctx "golang.org/x/net/context"
)

func (layer *Layer) Execute(sql string) ([]map[string]interface{}, error) {
	result, err := layer.session.Execute(goctx.Background(), sql)
	if err != nil {
		return make([]map[string]interface{}, 0), err
	}
	data := proccesRecordSet(result)
	return data, nil
}

func (layer *Layer) ExecuteStmt(sql string, param ...interface{}) ([]map[string]interface{}, error) {
	sql = parseStmt(sql, param...)
	result, err := layer.session.Execute(goctx.Background(), sql)
	if err != nil {
		return make([]map[string]interface{}, 0), err
	}
	data := proccesRecordSet(result)
	return data, nil
}

func parseStmt(sql string, param ...interface{}) string {
	for _, val := range param {
		switch val.(type) {
		case int:
			sql = strings.Replace(sql, "?", fmt.Sprintf(`%v`, val), 1)
		case int32:
			sql = strings.Replace(sql, "?", fmt.Sprintf(`%v`, val), 1)
		case int64:
			sql = strings.Replace(sql, "?", fmt.Sprintf(`%v`, val), 1)
		case float32:
			sql = strings.Replace(sql, "?", fmt.Sprintf(`%v`, val), 1)
		case float64:
			sql = strings.Replace(sql, "?", fmt.Sprintf(`%v`, val), 1)
		case []byte:
			sql = strings.Replace(sql, "?", fmt.Sprintf(`"%v"`, string(val.([]byte))), 1)
		default:
			sql = strings.Replace(sql, "?", fmt.Sprintf(`"%v"`, val), 1)
		}
	}
	return sql
}

func proccesRecordSet(result []ast.RecordSet) []map[string]interface{} {
	data := make([]map[string]interface{}, 0)
	if len(result) > 0 {
		fields := make([]string, 0)
		_fields := result[0].Fields()
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
	return data
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
			fmt.Println(reflect.ValueOf(v).Kind().String())
			switch v.(type) {
			case string:
				result.Set(k, v.(string))
			default:
				result.Set(k, fmt.Sprintf("%v", v))
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
