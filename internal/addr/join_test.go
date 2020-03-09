package addr_test

import (
	"testing"

	"github.com/alinz/baker.go/internal/addr"
)

func TestJoin(t *testing.T) {
	testCases := []struct {
		remote   addr.Endpoint
		path     string
		expected string
	}{
		{
			remote:   addr.Remote("0.0.0.0", 1000),
			path:     "/hello/world",
			expected: "http://0.0.0.0:1000/hello/world",
		},
		{
			remote:   addr.Remote("0.0.0.0", 1000),
			path:     "",
			expected: "http://0.0.0.0:1000/",
		},
		{
			remote:   addr.Remote("0.0.0.0", 1000),
			path:     "/",
			expected: "http://0.0.0.0:1000/",
		},
		{
			remote:   addr.RemoteHTTP(addr.Remote("0.0.0.0", 1000), "/hello", false),
			path:     "/",
			expected: "http://0.0.0.0:1000/hello",
		},
		{
			remote:   addr.RemoteHTTP(addr.Remote("0.0.0.0", 1000), "/hello", false),
			path:     "/world",
			expected: "http://0.0.0.0:1000/hello/world",
		},
		{
			remote:   addr.RemoteHTTP(addr.Remote("0.0.0.0", 1000), "/hello", true),
			path:     "/world",
			expected: "https://0.0.0.0:1000/hello/world",
		},
	}

	for i, testCase := range testCases {
		url, err := addr.Join(testCase.remote, testCase.path)
		if err != nil {
			t.Fatalf("in testcase %d, failed with %s", i+1, err)
		}

		if url != testCase.expected {
			t.Fatalf("in testcase %d, expected '%s', but got '%s'", i+1, testCase.expected, url)
		}
	}

}
