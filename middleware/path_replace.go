package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/alinz/baker.go"
)

type ReplacePath struct {
	Search  string `json:"search"`
	Replace string `json:"replace"`
	Times   int    `json:"times"`
}

var _ baker.Middleware = (*ReplacePath)(nil)

func (p *ReplacePath) Process(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.Replace(r.URL.Path, p.Search, p.Replace, p.Times)
		next.ServeHTTP(w, r)
	})
}

func init() {
	registeredMiddlewares["ReplacePath"] = func(raw json.RawMessage) (baker.Middleware, error) {
		ReplacePath := &ReplacePath{}
		err := json.Unmarshal(raw, ReplacePath)
		if err != nil {
			return nil, err
		}
		return ReplacePath, nil
	}
}
