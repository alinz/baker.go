package baker_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alinz/baker.go"
	"github.com/alinz/baker.go/confutil"
	"github.com/alinz/baker.go/rule"
	"github.com/stretchr/testify/assert"
)

func TestDomains(t *testing.T) {
	domains := baker.NewDomains()

	endpoint1 := &baker.Endpoint{
		Domain: "example.com",
		Path:   "/*",
		Ready:  true,
	}

	endpoint2 := &baker.Endpoint{
		Domain: "example.com",
		Path:   "/rpc/*",
		Ready:  true,
	}

	container1 := &baker.Container{
		ID: "1",
	}

	container2 := &baker.Container{
		ID: "2",
	}

	domains.Paths(endpoint1.Domain, true).Service(endpoint1.Path, true).Add(container1, endpoint1)
	domains.Paths(endpoint2.Domain, true).Service(endpoint2.Path, true).Add(container2, endpoint2)

	for i := 0; i < 200; i++ {
		container, endpoint, ok := domains.Paths("example.com", false).Service("/manifest.json", false).Select()

		assert.True(t, ok)
		assert.Equal(t, container1, container)
		assert.Equal(t, endpoint1, endpoint)
	}
}

func TestBaker(t *testing.T) {
	containers := MockDriver(t, confutil.NewEndpoints().New("example.com", "/*", true))

	baker := baker.New(
		containers,
		baker.WithPingDuration(2*time.Second),
		baker.WithRules(
			rule.RegisterAppendPath(),
			rule.RegisterReplacePath(),
			rule.RegisterRateLimiter(),
		),
	)

	s := httptest.NewServer(baker)
	t.Cleanup(s.Close)

	time.Sleep(3 * time.Second)

	httpClient := http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/manifest.json", s.URL), nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Host = "example.com"

	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
