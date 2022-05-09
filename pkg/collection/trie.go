package collection

const (
	wild rune = '*'
	path rune = '/'
)

type value[T any] struct {
	content T
}

type Trie[T any] struct {
	val      *value[T]
	children map[rune]*Trie[T]
}

func (t *Trie[T]) Put(key []rune, val T) {
	current := t
	i := 0

	for i < len(key) {
		r := key[i]

		next, ok := current.children[r]
		if !ok {
			next = NewTrie[T]()
			current.children[r] = next
		}

		current = next
		i++
	}

	if i != len(key) {
		return
	}

	current.val = &value[T]{
		content: val,
	}
}

func (t *Trie[T]) Get(key []rune) (found T, ok bool) {
	current := t
	i := 0
	isWild := false

	for i < len(key) {
		r := key[i]
		if r == path {
			isWild = false
		}

		if isWild {
			i++
			continue
		}

		next, ok := current.children[r]
		if !ok {
			next, ok = current.children[wild]
			if !ok {
				return found, false
			}
			isWild = true
		}

		current = next
		i++
	}

	if current.val == nil {
		return found, false
	}

	return current.val.content, true
}

func NewTrie[T any]() *Trie[T] {
	return &Trie[T]{
		children: make(map[rune]*Trie[T]),
	}
}
