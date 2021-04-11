package trie_test

import (
	"testing"

	"github.com/alinz/baker.go/internal/store/trie"
)

func BenchmarkTrie(b *testing.B) {
	node := trie.New(true)
	key := []rune("134")
	value := 1

	for n := 0; n < b.N; n++ {
		node.Put(key, value)
		node.Get(key)
		node.Del(key)
	}
}
