package healthz

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsWarning(t *testing.T) {
	assert.False(t, IsWarning(errors.New("foo")))
	assert.True(t, IsWarning(Warn("foo")))
	assert.True(t, IsWarning(Warnf("foo %d", 42)))
	assert.True(t, IsWarning(fmt.Errorf("wrapped: %w", Warn("foo"))))
	assert.False(t, IsWarning(fmt.Errorf("not wrapped: %v", Warn("foo"))))
}
