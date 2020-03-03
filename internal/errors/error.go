package errors

type Value string

func (v Value) Error() string {
	return string(v)
}
