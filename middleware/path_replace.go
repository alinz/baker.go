package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/alinz/baker.go"
)

type PathReplace struct {
	Search  string `json:"search"`
	Replace string `json:"replace"`
	Times   int    `json:"times"`
}

var _ baker.Middleware = (*PathReplace)(nil)

func (p *PathReplace) Process(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.Replace(r.URL.Path, p.Search, p.Replace, p.Times)
		next.ServeHTTP(w, r)
	})
}

func init() {
	registeredMiddlewares["PathReplace"] = func(raw json.RawMessage) (baker.Middleware, error) {
		pathReplace := &PathReplace{}
		err := json.Unmarshal(raw, pathReplace)
		if err != nil {
			return nil, err
		}
		return pathReplace, nil
	}
}
