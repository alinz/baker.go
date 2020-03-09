package errors

import (
	"bytes"
	"fmt"
)

type Value string

func (v Value) Error() string {
	return string(v)
}

type wrapper struct {
	base error
	msg  string
}

type wrapped struct {
	errors []error
}

func (w *wrapped) Error() string {
	var buffer bytes.Buffer

	for i, err := range w.errors {
		if i != 0 {
			buffer.WriteString(": ")
		}
		buffer.WriteString(err.Error())
	}

	return buffer.String()
}

func (w *wrapper) Error() string {
	return fmt.Sprintf("%s: %s", w.base, w.msg)
}

func is(err, base error) bool {
	if err == base {
		return true
	}

	wrapped, ok := err.(*wrapped)
	if !ok {
		return false
	}

	for _, err = range wrapped.errors {
		if is(err, base) {
			return true
		}
	}

	return false
}

func Is(err, base error) bool {
	if err == base {
		return true
	}

	if is(err, base) {
		return true
	}

	wrapper, ok := err.(*wrapper)
	if !ok {
		return false
	}

	return Is(wrapper.base, base)
}

func Warps(errors ...error) error {
	return &wrapped{
		errors: errors,
	}
}

func Wrap(base error, msg string) error {
	return &wrapper{
		base: base,
		msg:  msg,
	}
}

func Wrapf(base error, msg string, args ...interface{}) error {
	return &wrapper{
		base: base,
		msg:  fmt.Sprintf(msg, args...),
	}
}
