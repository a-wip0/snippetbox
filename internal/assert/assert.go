package assert

import (
	"strings"
	"testing"
)

func Equal[T comparable](t *testing.T, actual, expected T) {
	// Asserts that `actual` is equal to `expected`.
	t.Helper()

	if actual != expected {
		// Error if `actual` is not equal to `expected`.
		t.Errorf("got: %v; want: %v", actual, expected)
	}
}

func StringContains(t *testing.T, actual, expectedSubstring string) {
	// Asserts that `actual` contains `expectedSubstring`.
	t.Helper()

	if !strings.Contains(actual, expectedSubstring) {
		// Error if `actual` does not contain `expectedSubstring`.
		t.Errorf("got: %q; expected to contain: %q", actual, expectedSubstring)
	}
}

func NilError(t *testing.T, actual error) {
	// Asserts that `actual` is a nil error.
	t.Helper()

	if actual != nil {
		// Error if `actual` is not nil.
		t.Errorf("got: %v; expected: nil", actual)
	}
}
