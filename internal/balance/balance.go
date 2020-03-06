package balance

import (
	"github.com/alinz/baker.go/internal/errors"
)

const (
	ErrNotFound     = errors.Value("key not found")
	ErrDuplicateKey = errors.Value("duplicate key detected")
	ErrEmpty        = errors.Value("empty list")
)

type Balancer interface {
	Add(Key string) error
	Del(Key string) error
	Get() (string, error)
}
