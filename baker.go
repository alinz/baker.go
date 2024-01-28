package baker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/netip"
	"net/url"
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

func (e *Endpoint) getHashKey() string {
	var sb strings.Builder

	sb.WriteString(e.Domain)
	sb.WriteString(e.Path)

	return sb.String()
}

type Container struct {
	ID   string         `json:"id"`
	Addr netip.AddrPort `json:"addr"`
	Path string         `json:"path"`
}

var emptyPaths = NewPaths()
var emptyService = NewService()

type Domains struct {
	paths *collection.Map[*Paths]
}

func NewDomains() *Domains {
	return &Domains{
		paths: collection.NewMap[*Paths](),
	}
}

func (d *Domains) Paths(domain string, insert bool) *Paths {
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
	services       *collection.Trie[*Service]
	registeredPath *collection.Map[*Service]
}

func (p *Paths) Service(path string, insert bool) *Service {
	runePath := []rune(path)

	if insert {
		s, ok := p.registeredPath.Get(path)
		if ok {
			return s
		}

		s = NewService()
		p.services.Put(runePath, s)
		p.registeredPath.Put(path, s)

		return s
	}

	s, ok := p.services.Get(runePath)
	if !ok {
		return emptyService
	}

	return s
}

func NewPaths() *Paths {
	return &Paths{
		services:       collection.NewTrie[*Service](),
		registeredPath: collection.NewMap[*Service](),
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
	if !s.containers.Contains(container.ID) {
		log.Info().
			Str("id", container.ID).
			Str("domain", endpoint.Domain).
			Str("path", endpoint.Path).
			Msg("a new container is added")
	}
	s.containers.Put(container.ID, &value{
		container: container,
		endpoint:  endpoint,
	})
}

func (s *Service) Remove(container *Container) int {
	value, ok := s.containers.Get(container.ID)
	if ok {
		log.Info().
			Str("id", container.ID).
			Str("domain", value.endpoint.Domain).
			Str("path", value.endpoint.Path).
			Msg("an exisiting container is removed")
	}

	return s.containers.Remove(container.ID)
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
	domains            *Domains
	rules              map[string]rule.BuilderFunc
	pingDuration       time.Duration
	containers         *collection.Set[string, *Container]
	done               chan struct{}
	http               httpclient.GetterFunc
	refMap             *collection.Map[*value]
	middlewareCacheMap *collection.Map[rule.Middleware]
}

var _ http.Handler = &Server{}

func (s *Server) pinger() {
	for {
		select {
		case <-s.done:
			return
		case <-time.After(s.pingDuration):
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

	log.Debug().Str("domain", domain).Str("path", path).Msg("a request received")

	container, endpoint, ok := s.domains.Paths(domain, false).Service(path, false).Select()
	if !ok {
		log.Debug().Str("domain", domain).Str("path", path).Msg("not found")
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"error": "service is not available"}`))
		return
	}

	proxy := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			url := &url.URL{
				Scheme: "http",
				Host:   container.Addr.String(),
			}

			log.Debug().
				Str("recv_from", r.In.URL.String()).
				Str("send_to", url.String()).
				Msg("rewriting url")

			r.SetURL(url)     // Forward request to outboundURL.
			r.SetXForwarded() // Set X-Forwarded-* headers.
		},
	}

	rules, err := s.getMiddlewares(endpoint)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(struct {
			Error string `json:"error"`
		}{
			Error: err.Error(),
		})
		return
	}

	log.
		Debug().
		Str("domain", domain).
		Str("path", path).
		Str("container_id", container.ID).
		Msg("routing to the container")
	s.apply(proxy, rules...).ServeHTTP(w, r)
}

func (s *Server) getMiddlewares(endpoint *Endpoint) ([]rule.Middleware, error) {
	if len(endpoint.Rules) == 0 {
		return rule.Empty, nil
	}

	middlewares := make([]rule.Middleware, 0)

	for _, r := range endpoint.Rules {
		builder, ok := s.rules[r.Type]
		if !ok {
			return nil, fmt.Errorf("failed to find rule builder for %s", r.Type)
		}

		middleware, err := builder(r.Args)
		if err != nil {
			return nil, fmt.Errorf("failed to parse args for rule %s: %w", r.Type, err)
		}

		if middleware.IsCachable() {
			middleware = s.middlewareCacheMap.GetAndUpdate(endpoint.getHashKey(), func(old rule.Middleware, found bool) rule.Middleware {
				// NOTE: the reason we are doing this is because we want to update the middleware
				// and we don;t want to recreate some internal state of the middleware over and over
				// The responsibility of initializing the internal state of middleware is on the
				// UpdateMiddleware method.
				var current rule.Middleware
				if found {
					current = old
				} else {
					current = middleware
				}
				return current.UpdateMiddelware(middleware)
			})
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

type bakerOption struct {
	rules        map[string]rule.BuilderFunc
	pingDuration time.Duration
}

type bakerOptionFunc func(*bakerOption)

func WithPingDuration(d time.Duration) bakerOptionFunc {
	return func(o *bakerOption) {
		o.pingDuration = d
	}
}

func WithRules(rules ...rule.RegisterFunc) bakerOptionFunc {
	return func(o *bakerOption) {
		for _, rule := range rules {
			rule(o.rules)
		}
	}
}

func New(containers <-chan *Container, optFuncs ...bakerOptionFunc) *Server {
	opt := &bakerOption{
		rules:        make(map[string]rule.BuilderFunc),
		pingDuration: 10 * time.Second,
	}

	for _, optFunc := range optFuncs {
		optFunc(opt)
	}

	s := &Server{
		domains:            NewDomains(),
		rules:              opt.rules,
		pingDuration:       opt.pingDuration,
		containers:         collection.NewSet[string, *Container](),
		done:               make(chan struct{}, 1),
		http:               httpclient.New(),
		refMap:             collection.NewMap[*value](),
		middlewareCacheMap: collection.NewMap[rule.Middleware](),
	}

	go s.pinger()
	go func() {
		for {
			select {
			case <-s.done:
				return
			case container := <-containers:
				if container.Addr.IsValid() {
					log.Debug().Str("container_id", container.ID).Msg("adding to the container list")
					s.containers.Put(container.ID, container)
				} else {
					log.Debug().Str("container_id", container.ID).Msg("removing from the container list")
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

					remaining := s.domains.
						Paths(value.endpoint.Domain, false).
						Service(value.endpoint.Path, false).
						Remove(value.container)

					// NOTE: if there is no more containers for this endpoint
					// we can remove the middleware from the cache
					if remaining == 0 {
						s.middlewareCacheMap.Delete(value.endpoint.getHashKey())
					}
				}
			}
		}
	}()

	return s
}
