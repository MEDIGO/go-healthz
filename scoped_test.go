package healthz

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsScopedMultiError(t *testing.T) {
	me := ScopedMultiError{
		"foo": errors.New("foo-error"),
		"bar": errors.New("bar-error"),
	}
	require.Equal(t, "multiple errors:\nbar: bar-error\nfoo: foo-error", me.Error())
	require.True(t, IsScopedMultiError(me))
}

func TestIsScopedMultiError_empty(t *testing.T) {
	me := ScopedMultiError{}
	require.Equal(t, "multiple errors:", me.Error())
	require.True(t, IsScopedMultiError(me))
}
