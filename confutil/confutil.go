package confutil

import (
	"encoding/json"
	"net/http"
)

type endpoints struct {
	collection []struct {
		Domain string `json:"domain"`
		Path   string `json:"path"`
		Rule   []struct {
			Args any    `json:"args"`
			Type string `json:"type"`
		} `json:"rules"`
		Ready bool `json:"ready"`
	}
}

func (e *endpoints) New(domain, path string, ready bool) *endpoints {
	e.collection = append(e.collection, struct {
		Domain string `json:"domain"`
		Path   string `json:"path"`
		Rule   []struct {
			Args any    `json:"args"`
			Type string `json:"type"`
		} `json:"rules"`
		Ready bool `json:"ready"`
	}{
		Domain: domain,
		Path:   path,
		Ready:  ready,
	})
	return e
}

func (e *endpoints) WithRules(rules ...struct {
	Args any    `json:"args"`
	Type string `json:"type"`
}) *endpoints {
	if len(e.collection) == 0 {
		return e
	}

	e.collection[len(e.collection)-1].Rule = rules

	return e
}

func (e *endpoints) Done(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(e.collection)
}

func NewEndpoints() *endpoints {
	return &endpoints{}
}
