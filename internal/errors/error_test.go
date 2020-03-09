package errors_test

import (
	"fmt"
	"testing"

	"github.com/alinz/baker.go/internal/errors"
)

func TestIs(t *testing.T) {
	base := errors.Value("base error")
	notFound := errors.Value("not found")

	testCases := []struct {
		base     error
		wrapped  error
		expected bool
	}{
		{
			base:     base,
			wrapped:  errors.Wrap(base, "not found"),
			expected: true,
		},
		{
			base:     base,
			wrapped:  errors.Wrap(notFound, "not found"),
			expected: false,
		},
		{
			base:     base,
			wrapped:  errors.Wrap(errors.Warps(base, notFound), "item not found"),
			expected: true,
		},
		{
			base:     notFound,
			wrapped:  errors.Wrap(errors.Warps(base, notFound), "item not found"),
			expected: true,
		},
	}

	for i, testCase := range testCases {
		result := errors.Is(testCase.wrapped, testCase.base)
		if result != testCase.expected {
			t.Fatalf("in testcase %d, expected Is returns %v but got %v", i, testCase.expected, result)
		}

		fmt.Println(testCase.wrapped)
	}
}
