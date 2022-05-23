package collection_test

import (
	"testing"

	"github.com/alinz/baker.go/pkg/collection"
	"github.com/stretchr/testify/assert"
)

func TestPathTrie(t *testing.T) {

	type Query struct {
		path  string
		found bool
	}

	testCases := []struct {
		paths   []string
		queries []Query
	}{
		{
			paths: []string{
				"/a",
			},
			queries: []Query{
				{
					path:  "/a",
					found: true,
				},
				{
					path:  "/a/",
					found: false,
				},
				{
					path:  "/a/1",
					found: false,
				},
				{
					path:  "/a/aa",
					found: false,
				},
				{
					path:  "/aa/",
					found: false,
				},
			},
		},
		{
			paths: []string{
				"/a",
				"/a/*",
			},
			queries: []Query{
				{
					path:  "/a",
					found: true,
				},
				{
					path:  "/a/1",
					found: true,
				},
				{
					path:  "/a/1/",
					found: false,
				},
			},
		},
		{
			paths: []string{
				"/a",
				"/a/*",
				"/a/*/b",
			},
			queries: []Query{
				{
					path:  "/a/1/b",
					found: true,
				},
			},
		},
		{
			paths: []string{
				"/a/*",
				"/a/*/b",
			},
			queries: []Query{
				{
					path:  "/a/1",
					found: true,
				},
				{
					path:  "/a/",
					found: false,
				},
			},
		},
	}

	for _, tc := range testCases {
		pt := collection.NewTrie[bool]()
		for _, path := range tc.paths {
			pt.Put([]rune(path), true)
		}

		for _, query := range tc.queries {
			_, ok := pt.Get([]rune(query.path))
			assert.Equal(t, query.found, ok)
		}
	}
}
