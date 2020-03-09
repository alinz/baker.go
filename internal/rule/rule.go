package rule

import (
	"encoding/json"
	"net/http"
)

type Registrar interface {
	Name() string
	CreateHandler() Handler
}

var registrars = make(map[string]Registrar)

func Register(registrar Registrar) error {
	if _, ok := registrars[registrar.Name()]; ok {
		return nil
	}
	registrars[registrar.Name()] = registrar
	return nil
}

// Handler is an interface to define the logic
// to ApplyRule to incoming request
type Handler interface {
	Valid() error
	ApplyRule(next http.Handler) http.Handler
}

// Handlers is a type to implement UnmarshalJSON method
// every time a new rule is imeplemented, it needs to be added to
// append method
type Handlers []Handler

// append converts and valid the rule before adding it to the array
func (h *Handlers) append(name string, raw json.RawMessage) error {
	registrar, ok := registrars[name]
	if !ok {
		return nil
	}

	handler := registrar.CreateHandler()
	err := json.Unmarshal(raw, handler)
	if err != nil {
		return err
	}

	// valid if loaded rule config is valid
	err = handler.Valid()
	if err != nil {
		return err
	}

	*h = append(*h, handler)
	return nil
}

// Rule
//

// UnmarshalJSON overrides and decode rule handler configuration based on name
func (h *Handlers) UnmarshalJSON(p []byte) error {
	var rawMessages []json.RawMessage

	err := json.Unmarshal(p, &rawMessages)
	if err != nil {
		return err
	}

	// example of how a rule is configured in json payload
	// { "name": "path_replace", "config": { "search": "/api", "repalce": "", "times": 1 } }
	//
	rule := struct {
		Name   string          `json:"name"`
		Config json.RawMessage `json:"config"`
	}{}

	for _, rawMessage := range rawMessages {
		err = json.Unmarshal(rawMessage, &rule)
		if err != nil {
			return err
		}

		err = h.append(rule.Name, rule.Config)
		if err != nil {
			return err
		}
	}

	return nil
}
