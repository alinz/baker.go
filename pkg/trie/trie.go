package trie

const (
	wildChar rune = '*'
)

type Node[T any] struct {
	val      T
	parent   *Node[T]
	children map[rune]*Node[T]
	key      rune
	set      bool
	wild     bool
}

func (n *Node[T]) Get(key []rune) T {
	var wild *Node[T]
	var defaultValue T

	current := n
	i := 0

	for i < len(key) {
		r := key[i]

		if current.wild {
			wild = current
		}

		next, ok := current.children[r]
		if !ok {
			break
		}

		current = next
		i++
	}

	if i != len(key) {
		if wild != nil {
			current = wild
		} else {
			return defaultValue
		}
	}

	if !current.set {
		return defaultValue
	}

	return current.val
}

func (n *Node[T]) Put(key []rune, val T) {
	current := n

	for i := 0; i < len(key); i++ {
		r := key[i]

		next, ok := current.children[r]
		if !ok {
			next = New[T]()
			if r == wildChar {
				current.wild = true
				break
			} else {
				next.key = r
				next.parent = current
				current.children[r] = next
			}
		}

		current = next
	}

	current.val = val
	current.set = true
}

func (n *Node[T]) Del(key []rune) {
	current := n

	// first need to find the node
	for i := 0; i < len(key); i++ {
		r := key[i]

		if r == wildChar {
			if current.wild {
				continue
			}

			return
		}

		next, ok := current.children[r]
		if !ok {
			return
		}

		current = next
	}

	targetNode := current

	// backtrack and clean up node
	for current != nil {
		//
		// This condition is very important. (shouldContinue)
		// if for example we have
		//
		// /a/b/c -> node 1
		// /a/b -> node 2
		//
		// if /a/b/c is deleted, it should not delete /a/b because /a/b has a value node 2
		//
		shouldContinue := len(current.children) == 0 && (targetNode == current || !current.set)
		if !shouldContinue {
			return
		}

		if current.parent != nil {
			delete(current.parent.children, current.key)
		}

		var defaultValue T
		current.val = defaultValue
		current.wild = false

		current = current.parent
	}
}

func (n *Node[T]) Size() int {
	return len(n.children)
}

func New[T any]() *Node[T] {
	return &Node[T]{
		children: make(map[rune]*Node[T]),
	}
}
