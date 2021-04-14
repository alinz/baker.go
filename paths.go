package baker

const (
	wildChar rune = '*'
)

type Paths struct {
	parent   *Paths
	children map[rune]*Paths
	val      *Service
	key      rune
	wild     bool
}

func (p *Paths) Get(key []rune) *Service {
	var wild *Paths

	current := p
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
			return nil
		}
	}

	if current.val == nil {
		return nil
	}

	return current.val
}

func (p *Paths) Put(key []rune, val *Service) {
	current := p

	for i := 0; i < len(key); i++ {
		r := key[i]

		next, ok := current.children[r]
		if !ok {
			next = NewPaths()
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
}

func (p *Paths) Del(key []rune) {
	current := p

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
		shouldContinue := len(current.children) == 0 && (targetNode == current || current.val == nil)
		if !shouldContinue {
			return
		}

		if current.parent != nil {
			delete(current.parent.children, current.key)
		}

		current.val = nil
		current.wild = false

		current = current.parent
	}
}

func (p *Paths) Size() int {
	return len(p.children)
}

func NewPaths() *Paths {
	return &Paths{
		children: make(map[rune]*Paths),
	}
}
