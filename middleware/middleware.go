package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/alinz/baker.go"
)

var emptyMiddlewares = []baker.Middleware{}
var registeredMiddlewares = map[string]func(raw json.RawMessage) (baker.Middleware, error){}

func Parse(rules []baker.Rule) ([]baker.Middleware, error) {
	if len(rules) == 0 {
		return emptyMiddlewares, nil
	}

	middlewares := make([]baker.Middleware, 0)

	for _, rule := range rules {
		builder, ok := registeredMiddlewares[rule.Type]
		if !ok {
			return nil, fmt.Errorf("failed to find middleware builder for %s", rule.Type)
		}

		middleware, err := builder(rule.Args)
		if err != nil {
			return nil, fmt.Errorf("failed to parse args for middleware %s: %w", rule.Type, err)
		}

		middlewares = append(middlewares, middleware)
	}

	return middlewares, nil
}

func Apply(next http.Handler, middlewares ...baker.Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		next = middlewares[i].Process(next)
	}

	return next
}
