package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/alinz/baker.go"
)

type AppendPath struct {
	Begin string `json:"begin"`
	End   string `json:"end"`
}

var _ baker.Middleware = (*AppendPath)(nil)

func (p *AppendPath) Process(next http.Handler) http.Handler {
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
	registeredMiddlewares["AppendPath"] = func(raw json.RawMessage) (baker.Middleware, error) {
		AppendPath := &AppendPath{}
		err := json.Unmarshal(raw, AppendPath)
		if err != nil {
			return nil, err
		}
		return AppendPath, nil
	}
}
