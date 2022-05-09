package confutil_test

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alinz/baker.go/confutil"
	"github.com/alinz/baker.go/rule"
	"github.com/stretchr/testify/assert"
)

func TestBakerEndpoints(t *testing.T) {
	{
		var endpoints confutil.Endpoints
		rr := httptest.NewRecorder()
		endpoints.
			New("example.com", "/", true).
			Done(rr)
		assert.JSONEq(t, `[{"domain":"example.com","path":"/","rules":null,"ready":true}]`, strings.TrimSpace(rr.Body.String()))
	}

	{
		var endpoints confutil.Endpoints
		rr := httptest.NewRecorder()
		endpoints.
			New("example.com", "/", true).
			WithRules(rule.NewAppendPath("a", "b")).
			Done(rr)
		assert.JSONEq(t, `[{"domain":"example.com","path":"/","rules":[{"args":{"begin":"a","end":"b"},"type":"AppendPath"}],"ready":true}]`, strings.TrimSpace(rr.Body.String()))
	}

	{
		var endpoints confutil.Endpoints
		rr := httptest.NewRecorder()
		endpoints.
			New("example.com", "/", true).
			WithRules(
				rule.NewAppendPath("a", "b"),
				rule.NewReplacePath("/a", "/b", 1),
			).
			Done(rr)
		assert.JSONEq(t, `[{"domain":"example.com","path":"/","rules":[{"args":{"begin":"a","end":"b"},"type":"AppendPath"},{"args":{"search":"/a","replace":"/b","times":1},"type":"ReplacePath"}],"ready":true}]`, strings.TrimSpace(rr.Body.String()))
	}
}
