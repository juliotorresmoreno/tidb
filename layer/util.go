package layer

import (
	"encoding/json"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

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
