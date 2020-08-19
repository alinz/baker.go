package baker

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/alinz/baker.go/internal/response"
)

type BaseRouter struct {
	store         Store
	processorsMap map[string]func(config json.RawMessage) (Processor, error)
}

func (br *BaseRouter) HostPolicy(ctx context.Context, host string) error {
	return nil
}

func (br *BaseRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	domain := r.Host
	path := r.URL.Path

	service := br.store.Query(domain, path)
	if service == nil {
		response.AsJSON(w, http.StatusServiceUnavailable, fmt.Errorf("service is not available"))
		return
	}

	processors := make([]Processor, 0)

	if service.Rule.Recipes != nil {
		for _, recipe := range service.Rule.Recipes {
			if builder, ok := br.processorsMap[recipe.Name]; ok {
				processor, err := builder(recipe.Config)
				if err != nil {
					response.AsJSON(w, http.StatusInternalServerError, fmt.Errorf("failed to build recipe for %s: %s", recipe.Name, err))
					return
				}

				processors = append(processors, processor)
			}
		}
	}

	errorIndex := -1

	proxy := &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			r.URL.Scheme = "http"
			r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
			r.URL.Host = service.Container.RemoteAddr.String()

			if _, ok := r.Header["User-Agent"]; !ok {
				// explicitly disable User-Agent so it's not set to default value
				r.Header.Set("User-Agent", "")
			}

			for _, processor := range processors {
				processor.Request(r)
			}
		},
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 10 * time.Second,
			}).Dial,
		},
		ModifyResponse: func(r *http.Response) error {
			for i, processor := range processors {
				err := processor.Response(r)
				if err != nil {
					errorIndex = i
					return err
				}
			}
			return nil
		},
		ErrorHandler: func(rw http.ResponseWriter, r *http.Request, err error) {
			if errorIndex == -1 {
				response.AsJSON(w, http.StatusInternalServerError, err)
				return
			}

			processor := processors[errorIndex]
			processor.HandleError(rw, r, err)
		},
	}

	proxy.ServeHTTP(w, r)
}

func (r *BaseRouter) AddProcessor(name string, builder func(config json.RawMessage) (Processor, error)) *BaseRouter {
	r.processorsMap[name] = builder
	return r
}

func NewBaseRouter(store Store) *BaseRouter {
	return &BaseRouter{
		store:         store,
		processorsMap: make(map[string]func(config json.RawMessage) (Processor, error)),
	}
}
