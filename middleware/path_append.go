package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/alinz/baker.go"
)

type PathAppend struct {
	Begin string `json:"begin"`
	End   string `json:"end"`
}

var _ baker.Middleware = (*PathAppend)(nil)

func (p *PathAppend) Process(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var bs strings.Builder

		bs.WriteString(p.Begin)
		bs.WriteString(r.URL.Path)
		bs.WriteString(p.End)

		r.URL.Path = bs.String()

		next.ServeHTTP(w, r)
	})
}

func init() {
	registeredMiddlewares["PathAppend"] = func(raw json.RawMessage) (baker.Middleware, error) {
		pathAppend := &PathAppend{}
		err := json.Unmarshal(raw, pathAppend)
		if err != nil {
			return nil, err
		}
		return pathAppend, nil
	}
}
