package rule

import (
	"encoding/json"
	"net/http"
	"strings"
)

const AppendPathName = "AppendPath"

type AppendPath struct {
	Begin string `json:"begin"`
	End   string `json:"end"`
}

var _ Middleware = (*AppendPath)(nil)

func (a *AppendPath) IsCachable() bool {
	return false
}

func (a *AppendPath) UpdateMiddelware(newImpl Middleware) Middleware {
	return nil
}

func (a *AppendPath) Process(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var bs strings.Builder

		bs.WriteString(a.Begin)
		bs.WriteString(r.URL.Path)
		bs.WriteString(a.End)

		r.URL.Path = bs.String()

		next.ServeHTTP(w, r)
	})
}

func NewAppendPath(begin, end string) struct {
	Type string `json:"type"`
	Args any    `json:"args"`
} {
	return struct {
		Type string `json:"type"`
		Args any    `json:"args"`
	}{
		Type: AppendPathName,
		Args: AppendPath{
			Begin: begin,
			End:   end,
		},
	}
}

func RegisterAppendPath() RegisterFunc {
	return func(m map[string]BuilderFunc) error {
		m[AppendPathName] = func(raw json.RawMessage) (Middleware, error) {
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

const ReplacePathName = "ReplacePath"

type ReplacePath struct {
	Search  string `json:"search"`
	Replace string `json:"replace"`
	Times   int    `json:"times"`
}

var _ Middleware = (*ReplacePath)(nil)

func (p *ReplacePath) IsCachable() bool {
	return false
}

func (p *ReplacePath) UpdateMiddelware(newImpl Middleware) Middleware {
	return nil
}

func (p *ReplacePath) Process(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.Replace(r.URL.Path, p.Search, p.Replace, p.Times)
		next.ServeHTTP(w, r)
	})
}

func NewReplacePath(search string, replace string, times int) struct {
	Type string `json:"type"`
	Args any    `json:"args"`
} {
	return struct {
		Type string `json:"type"`
		Args any    `json:"args"`
	}{
		Type: ReplacePathName,
		Args: ReplacePath{
			Search:  search,
			Replace: replace,
			Times:   times,
		},
	}
}

func RegisterReplacePath() RegisterFunc {
	return func(m map[string]BuilderFunc) error {
		m[ReplacePathName] = func(raw json.RawMessage) (Middleware, error) {
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
