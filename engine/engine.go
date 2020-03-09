package engine

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"

	"github.com/alinz/baker.go"
	"github.com/alinz/baker.go/driver"
	"github.com/alinz/baker.go/internal/addr"
	"github.com/alinz/baker.go/internal/response"
)

func normalizeHost(host string) (string, bool) {
	www := false
	if strings.HasPrefix(host, "www.") {
		www = true
		host = strings.Replace(host, "www.", "", 1)
	}
	return host, www
}

type Engine struct {
	store   *baker.Store
	watcher driver.Watcher
	mux     sync.RWMutex
}

var _ http.Handler = (*Engine)(nil)

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	path := r.URL.Path

	e.mux.RLock()
	domain, err := e.store.Get(host)
	e.mux.RUnlock()
	if err != nil {
		response.AsJSON(w, http.StatusServiceUnavailable, err)
		return
	}

	e.mux.RLock()
	service, err := domain.Get(path)
	e.mux.RUnlock()
	if err != nil {
		response.AsJSON(w, http.StatusServiceUnavailable, err)
		return
	}

	e.mux.RLock()
	target, err := service.Get()
	e.mux.RUnlock()
	if err != nil {
		response.AsJSON(w, http.StatusServiceUnavailable, err)
		return
	}

	remoteAddr, err := url.Parse(addr.RemoteHTTP(target.Container.RemoteAddr, target.Container.ConfigPath, false).String())
	if err != nil {
		response.AsJSON(w, http.StatusServiceUnavailable, err)
		return
	}

	var handler http.Handler = httputil.NewSingleHostReverseProxy(remoteAddr)
	for i := len(target.Config.RuleHandlers); i > 0; i-- {
		handler = target.Config.RuleHandlers[i-1].ApplyRule(handler)
	}

	r.URL.Path = ""

	handler.ServeHTTP(w, r)
}

// New creates a new Reverse Proxy Engine based on given driver
func New(watcher driver.Watcher) *Engine {
	return &Engine{
		store:   baker.NewStore(),
		watcher: watcher,
	}
}
