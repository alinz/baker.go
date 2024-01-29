package baker_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"
	"time"

	"github.com/alinz/baker.go"
	"github.com/alinz/baker.go/pkg/collection"
	"github.com/alinz/baker.go/rule"
)

func MockDriver(t *testing.T, confs ...interface{ WriteResponse(w http.ResponseWriter) }) <-chan *baker.Container {
	containers := make(chan *baker.Container, len(confs))

	httpServers := make([]*httptest.Server, 0, len(confs))

	for i, conf := range confs {
		server := func(conf interface{ WriteResponse(w http.ResponseWriter) }) *httptest.Server {
			return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/config" {
					conf.WriteResponse(w)
					return
				}

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{}`))
			}))
		}(conf)

		addr, err := netip.ParseAddrPort(server.Listener.Addr().String())
		if err != nil {
			t.Fatal(err)
		}

		containers <- &baker.Container{
			ID:   fmt.Sprintf("container-%d", i),
			Addr: addr,
			Path: "/config",
		}

		httpServers = append(httpServers, server)
	}

	t.Cleanup(func() {
		for _, server := range httpServers {
			server.Close()
		}
	})

	return containers
}

func StartBakerServer(t *testing.T, containers <-chan *baker.Container, size int) (url string) {
	done := make(chan struct{}, 1)

	baker := baker.New(
		containers,
		baker.WithPingDuration(1*time.Second),
		baker.WithRules(
			rule.RegisterAppendPath(),
			rule.RegisterReplacePath(),
			rule.RegisterRateLimiter(),
		),
		baker.WithOnAfterPinger(func(containerSet *collection.Set[string, *baker.Container]) {
			if containerSet.Len() == size && done != nil {
				close(done)
				done = nil
			}
		}),
	)

	s := httptest.NewServer(baker)
	t.Cleanup(s.Close)

	<-done

	return s.URL
}
