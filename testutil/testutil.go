package testutil

import (
	"reflect"
	"testing"
)

// AssertEq fails the test if got and want are not deeply equal.
func AssertEq[T any](t *testing.T, got, want T) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("\ngot:  %v\nwant: %v", got, want)
	}
}
