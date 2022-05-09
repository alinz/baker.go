package trie_test

import (
	"testing"

	"github.com/alinz/baker.go/pkg/trie"
	"github.com/stretchr/testify/assert"
)

func TestPaths(t *testing.T) {

	t.Run("testing adding new value to trie", func(t *testing.T) {
		trie := trie.New[int]()
		trie.Put([]rune("/"), 1)

		assert.Equal(t, 1, trie.Size())
		assert.Equal(t, 1, trie.Get([]rune("/")))

		trie.Del([]rune("/"))
		assert.Equal(t, 0, trie.Size())
		assert.Equal(t, 0, trie.Get([]rune("/")))
	})

	t.Run("testing children", func(t *testing.T) {
		trie := trie.New[int]()
		trie.Put([]rune("/a/b/c"), 1)
		trie.Put([]rune("/a/b"), 2)

		// assert.Equal(t, 2, trie.Size())

		assert.Equal(t, 1, trie.Get([]rune("/a/b/c")))
		assert.Equal(t, 2, trie.Get([]rune("/a/b")))

		trie.Del([]rune("/a/b/c"))
		// assert.Equal(t, 1, trie.Size())
		assert.Equal(t, 2, trie.Get([]rune("/a/b")))
		assert.Equal(t, 0, trie.Get([]rune("/a/b/c")))
	})
}
