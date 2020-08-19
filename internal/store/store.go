package store

import "errors"

var (
	ErrItemNotFound        = errors.New("item not found")
	ErrItemAlreadyHasValue = errors.New("item already has value")
)

type KeyValue interface {
	Get(key []rune) (interface{}, error)
	Put(key []rune, value interface{}) error
	Del(key []rune) error
}
