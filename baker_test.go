package baker_test

import (
	"fmt"
	"io"
	"net/http"
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

func TestBasicBaker(t *testing.T) {
	configs := []interface {
		WriteResponse(w http.ResponseWriter)
	}{
		confutil.NewEndpoints().New("example.com", "/*", true).WithRules(
			rule.NewRateLimiter(1, time.Second),
		),
	}
	containers := MockDriver(t, configs...)

	url := StartBakerServer(t, containers, len(configs))

	t.Run("check if the pattern exists", func(t *testing.T) {
		httpClient := http.Client{}
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/manifest.json", url), nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Host = "example.com"

		resp, err := httpClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("return an error because od domain not found", func(t *testing.T) {
		httpClient := http.Client{}
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/manifest.json", url), nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Host = "example2.com"

		resp, err := httpClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		result, _ := io.ReadAll(resp.Body)

		assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
		assert.Equal(t, `{"error": "service is not available"}`, string(result))
	})

	t.Run("testing RateLimiting", func(t *testing.T) {

		httpClient := http.Client{}
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/manifest.json", url), nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Host = "example.com"

		time.Sleep(2 * time.Second)

		resp, err := httpClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		resp, err = httpClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)

		time.Sleep(2 * time.Second)

		resp, err = httpClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
