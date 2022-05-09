package baker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/netip"
	"strings"
	"time"

	"github.com/alinz/baker.go/pkg/collection"
	"github.com/alinz/baker.go/pkg/httpclient"
	"github.com/alinz/baker.go/pkg/log"
	"github.com/alinz/baker.go/rule"
)

type Rule struct {
	Type string          `json:"type"`
	Args json.RawMessage `json:"args"`
}

type Endpoint struct {
	Domain string `json:"domain"`
	Path   string `json:"path"`
	Rules  []Rule `json:"rules"`
	Ready  bool   `json:"ready"`
}

type Container struct {
	ID   string         `json:"id"`
	Addr netip.AddrPort `json:"addr"`
	Path string         `json:"path"`
}

var emptyPaths = NewPaths()
var emptyService = NewService()

type domains struct {
	paths *collection.Map[*Paths]
}

func (d *domains) Paths(domain string, insert bool) *Paths {
	if p, ok := d.paths.Get(domain); ok {
		return p
	} else if !insert {
		return emptyPaths
	}

	p := NewPaths()
	d.paths.Put(domain, p)

	return p
}

type Paths struct {
	services *collection.Trie[*Service]
}

func (p *Paths) Service(path string, insert bool) *Service {
	runePath := []rune(path)

	if s, ok := p.services.Get(runePath); ok {
		return s
	} else if !insert {
		return emptyService
	}

	s := NewService()
	p.services.Put(runePath, s)

	return s
}

func NewPaths() *Paths {
	return &Paths{
		services: collection.NewTrie[*Service](),
	}
}

type value struct {
	container *Container
	endpoint  *Endpoint
}

type Service struct {
	containers *collection.Set[string, *value]
}

func (s *Service) Add(container *Container, endpoint *Endpoint) {
	s.containers.Put(container.ID, &value{
		container: container,
		endpoint:  endpoint,
	})
}

func (s *Service) Remove(container *Container) {
	s.containers.Remove(container.ID)
}

func (s *Service) Select() (*Container, *Endpoint, bool) {
	value, ok := s.containers.Random()
	if !ok {
		return nil, nil, false
	}
	return value.container, value.endpoint, true
}

func NewService() *Service {
	return &Service{
		containers: collection.NewSet[string, *value](),
	}
}

type Server struct {
	domains    *domains
	rules      map[string]rule.BuilderFunc
	containers *collection.Set[string, *Container]
	done       chan struct{}
	http       httpclient.GetterFunc
	refMap     *collection.Map[*value]
}

var _ http.Handler = &Server{}

func (s *Server) pinger() {
	for {
		select {
		case <-s.done:
			return
		case <-time.After(10 * time.Second):
			s.containers.Iterate(func(id string, container *Container) bool {
				configPath := fmt.Sprintf("http://%s%s", container.Addr, container.Path)
				body, err := s.http(configPath)
				if err != nil {
					log.Error().
						Err(err).
						Str("id", container.ID).
						Str("path", configPath).
						Msg("failed to get config")
					return true
				}

				var endpoints []*Endpoint

				err = json.NewDecoder(body).Decode(&endpoints)
				if err != nil {
					log.Error().
						Err(err).
						Str("id", container.ID).
						Str("path", configPath).
						Msg("failed to decode config")
					return true
				}

				for _, endpoint := range endpoints {
					log.
						Debug().
						Str("domain", endpoint.Domain).
						Str("path", endpoint.Path).
						Str("container_id", container.ID).
						Msg("added/updated endpoint")

					s.refMap.Put(container.ID, &value{
						container: container,
						endpoint:  endpoint,
					})

					s.domains.
						Paths(endpoint.Domain, true).
						Service(endpoint.Path, true).
						Add(container, endpoint)
				}

				return true
			})
		}
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	domain := r.Host
	path := r.URL.Path

	container, endpoint, ok := s.domains.Paths(domain, false).Service(path, false).Select()
	if !ok {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"error": "service is not available"}`))
		return
	}

	proxy := &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			r.URL.Scheme = "http"
			r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
			r.URL.Host = container.Addr.String()

			if _, ok := r.Header["User-Agent"]; !ok {
				// explicitly disable User-Agent so it's not set to default value
				r.Header.Set("User-Agent", "")
			}
		},
	}

	rules, err := s.getMiddlewares(endpoint.Rules)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(struct {
			Error string `json:"error"`
		}{
			Error: err.Error(),
		})
		return
	}

	s.apply(proxy, rules...).ServeHTTP(w, r)
}

func (s *Server) getMiddlewares(rules []Rule) ([]rule.Middleware, error) {
	if len(rules) == 0 {
		return rule.Empty, nil
	}

	middlewares := make([]rule.Middleware, 0)

	for _, rule := range rules {
		builder, ok := s.rules[rule.Type]
		if !ok {
			return nil, fmt.Errorf("failed to find rule builder for %s", rule.Type)
		}

		middleware, err := builder(rule.Args)
		if err != nil {
			return nil, fmt.Errorf("failed to parse args for rule %s: %w", rule.Type, err)
		}

		middlewares = append(middlewares, middleware)
	}

	return middlewares, nil
}

func (s *Server) apply(next http.Handler, rules ...rule.Middleware) http.Handler {
	for i := len(rules) - 1; i >= 0; i-- {
		next = rules[i].Process(next)
	}

	return next
}

func New(containers <-chan *Container, rules ...rule.RegisterFunc) *Server {
	s := &Server{
		domains: &domains{
			paths: collection.NewMap[*Paths](),
		},
		rules:      make(map[string]rule.BuilderFunc),
		containers: collection.NewSet[string, *Container](),
		done:       make(chan struct{}, 1),
		http:       httpclient.New(),
		refMap:     collection.NewMap[*value](),
	}

	for _, rule := range rules {
		rule(s.rules)
	}

	go s.pinger()
	go func() {
		for {
			select {
			case <-s.done:
				return
			case container := <-containers:
				if container.Addr.IsValid() {
					s.containers.Put(container.ID, container)
				} else {
					s.containers.Remove(container.ID)

					value, ok := s.refMap.Get(container.ID)
					if !ok {
						log.
							Error().
							Str("container_id", container.ID).
							Msg("failed to find container in refMap")
						return
					}

					s.refMap.Delete(container.ID)

					s.domains.
						Paths(value.endpoint.Domain, false).
						Service(value.endpoint.Path, false).
						Remove(value.container)
				}
			}
		}
	}()

	return s
}

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

func (e *Endpoints) WithRule(rules ...struct {
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
	json.NewEncoder(w).Encode(e)
}
