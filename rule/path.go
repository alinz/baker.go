package rule

import (
	"encoding/json"
	"net/http"
	"strings"
)

type AppendPath struct {
	Begin string `json:"begin"`
	End   string `json:"end"`
}

var _ Middleware = (*AppendPath)(nil)

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

func NewAppendPath(begin, end string) struct {
	Args any    `json:"args"`
	Type string `json:"type"`
} {
	return struct {
		Args any    `json:"args"`
		Type string `json:"type"`
	}{
		Type: "AppendPath",
		Args: AppendPath{
			Begin: begin,
			End:   end,
		},
	}
}

func RegisterAppendPath() RegisterFunc {
	return func(m map[string]BuilderFunc) error {
		m["AppendPath"] = func(raw json.RawMessage) (Middleware, error) {
			AppendPath := &AppendPath{}
			err := json.Unmarshal(raw, AppendPath)
			if err != nil {
				return nil, err
			}
			return AppendPath, nil
		}

		return nil
	}
}

type ReplacePath struct {
	Search  string `json:"search"`
	Replace string `json:"replace"`
	Times   int    `json:"times"`
}

var _ Middleware = (*ReplacePath)(nil)

func (p *ReplacePath) Process(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.Replace(r.URL.Path, p.Search, p.Replace, p.Times)
		next.ServeHTTP(w, r)
	})
}

func NewReplacePath(search string, replace string, times int) struct {
	Args any    `json:"args"`
	Type string `json:"type"`
} {
	return struct {
		Args any    `json:"args"`
		Type string `json:"type"`
	}{
		Type: "ReplacePath",
		Args: ReplacePath{
			Search:  search,
			Replace: replace,
			Times:   times,
		},
	}
}

func RegisterReplacePath() RegisterFunc {
	return func(m map[string]BuilderFunc) error {
		m["ReplacePath"] = func(raw json.RawMessage) (Middleware, error) {
			ReplacePath := &ReplacePath{}
			err := json.Unmarshal(raw, ReplacePath)
			if err != nil {
				return nil, err
			}
			return ReplacePath, nil
		}

		return nil
	}
}
