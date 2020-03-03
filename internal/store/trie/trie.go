package trie

import (
	"github.com/alinz/baker.go/internal/store"
)

const (
	wildChar rune = '*'
)

type value struct {
	content interface{}
}

type Node struct {
	key      rune
	parent   *Node // need parent node for deletion
	children map[rune]*Node
	val      *value
	wild     bool
}

var _ store.KeyValue = (*Node)(nil)

func (n *Node) Get(key []rune) (interface{}, error) {
	var wild *Node

	curr := n
	i := 0

	for i < len(key) {
		r := key[i]

		if curr.wild {
			wild = curr
		}

		next, ok := curr.children[r]
		if !ok {
			break
		}

		curr = next
		i++
	}

	if i != len(key) {
		if wild != nil {
			curr = wild
		} else {
			return nil, store.ErrItemNotFound
		}
	}

	if curr.val == nil {
		return nil, store.ErrItemNotFound
	}

	return curr.val.content, nil
}

func (n *Node) Put(key []rune, val interface{}) error {
	curr := n

	for i := 0; i < len(key); i++ {
		r := key[i]

		next, ok := curr.children[r]
		if !ok {
			next = New()
			if r == wildChar {
				curr.wild = true
				break
			} else {
				next.key = r
				next.parent = curr
				curr.children[r] = next
			}
		}

		curr = next
	}

	if curr.val != nil {
		return store.ErrItemAlreadyHasValue
	}

	curr.val = &value{content: val}

	return nil
}

func (n *Node) Del(key []rune) error {
	curr := n

	// first need to find the node
	for i := 0; i < len(key); i++ {
		r := key[i]

		if r == wildChar {
			if curr.wild {
				continue
			}

			return store.ErrItemNotFound
		}

		next, ok := curr.children[r]
		if !ok {
			return store.ErrItemNotFound
		}

		curr = next
	}

	// backtrack and clean up node
	for curr != nil {
		curr.val = nil
		curr.wild = false

		if len(curr.children) == 0 {
			if curr.parent != nil {
				delete(curr.parent.children, curr.key)
			}
		}
		curr = curr.parent
	}

	return nil
}

func New() *Node {
	return &Node{
		children: make(map[rune]*Node),
	}
}
