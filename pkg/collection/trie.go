package collection

const (
	plus rune = '+'
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
	isPlus := false

	for i < len(key) {
		r := key[i]
		if r == path {
			isPlus = false
		}

		if isPlus {
			i++
			continue
		}

		next, ok := current.children[r]
		if !ok {
			next, ok = current.children[wild]
			if ok {
				current = next
				break
			}

			next, ok = current.children[plus]
			if !ok {
				return found, false
			}
			isPlus = true
		}

		current = next
		i++
	}

	// need to check if the last node is a wildcard
	if next, ok := current.children[wild]; ok {
		current = next
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
