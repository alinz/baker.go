package store

import (
	"github.com/alinz/baker.go/internal/errors"
)

const (
	ErrItemNotFound        = errors.Value("item not found")
	ErrItemAlreadyHasValue = errors.Value("item already has value")
)

type KeyValue interface {
	Get(key []rune) (interface{}, error)
	Put(key []rune, value interface{}) error
	Del(key []rune) error
}
