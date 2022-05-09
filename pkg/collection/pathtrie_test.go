package collection_test

import (
	"testing"

	"github.com/alinz/baker.go/pkg/collection"
	"github.com/stretchr/testify/assert"
)

func TestPathTrie(t *testing.T) {

	testCases := []struct {
		paths  [][]rune
		search []rune
		ok     bool
	}{
		{
			paths: [][]rune{
				[]rune("/a"),
			},
			search: []rune("/a"),
			ok:     true,
		},
		{
			paths: [][]rune{
				[]rune("/a"),
				[]rune("/a/*"),
			},
			search: []rune("/a"),
			ok:     true,
		},
		{
			paths: [][]rune{
				[]rune("/a"),
				[]rune("/a/*"),
			},
			search: []rune("/a/1"),
			ok:     true,
		},
		{
			paths: [][]rune{
				[]rune("/a"),
				[]rune("/a/*"),
			},
			search: []rune("/a/1/"),
			ok:     false,
		},
		{
			paths: [][]rune{
				[]rune("/a"),
				[]rune("/a/*"),
				[]rune("/a/*/b"),
			},
			search: []rune("/a/1/b"),
			ok:     true,
		},
		{
			paths: [][]rune{
				[]rune("/a/*"),
				[]rune("/a/*/b"),
			},
			search: []rune("/a/1"),
			ok:     true,
		},
		{
			paths: [][]rune{
				[]rune("/a/*"),
				[]rune("/a/*/b"),
			},
			search: []rune("/a/"),
			ok:     false,
		},
	}

	for _, tc := range testCases {
		pt := collection.NewTrie[bool]()
		for _, path := range tc.paths {
			pt.Put(path, true)
		}

		_, ok := pt.Get(tc.search)
		assert.Equal(t, tc.ok, ok)
	}
}
