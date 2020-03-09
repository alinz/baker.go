package rule

import (
	"net/http"
	"strings"
)

type PathReplace struct {
	Search  string `json:"search"`
	Replace string `json:"replace"`
	Times   int    `json:"times"`
}

var _ Handler = (*PathReplace)(nil)

func (p *PathReplace) Valid() error {
	return nil
}

func (p *PathReplace) ApplyRule(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.Replace(r.URL.Path, p.Search, p.Replace, p.Times)
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

type PathReplaceRegistry struct{}

var _ Registrar = (*PathReplaceRegistry)(nil)

func (p *PathReplaceRegistry) Name() string {
	return "path_replace"
}

func (p *PathReplaceRegistry) CreateHandler() Handler {
	return &PathReplace{}
}
