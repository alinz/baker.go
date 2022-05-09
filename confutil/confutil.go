package confutil

import (
	"encoding/json"
	"net/http"
)

type Endpoints struct {
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

func (e *Endpoints) New(domain, path string, ready bool) *Endpoints {
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

func (e *Endpoints) WithRules(rules ...struct {
	Args any    `json:"args"`
	Type string `json:"type"`
}) *Endpoints {
	if len(e.collection) == 0 {
		return e
	}

	e.collection[len(e.collection)-1].Rule = rules

	return e
}

func (e *Endpoints) Done(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(e.collection)
}
