package healthz

import (
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func setTestValues(remote *Checker) {
	remote.Set("error1", errors.New("error1 value"), 0)
	remote.Set("error2", errors.New("error2 value"), 0)
	remote.Set("warning1", Warn("warning1 value"), 0)
	remote.Set("multi", ScopedMultiError{
		"e1": errors.New("e1"),
		"e2": errors.New("e2"),
		"w1": Warn("w1"),
	}, 0)
}

func TestRegisterRemote(t *testing.T) {
	remote := NewChecker(nil)
	defer remote.Close()
	setTestValues(remote)

	server := httptest.NewServer(remote.Handler())
	defer server.Close()

	local := NewChecker(nil)
	defer local.Close()
	err := local.RegisterRemote("remote", time.Second, server.URL, nil)
	assert.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	st := local.Status()
	assert.Len(t, st.Failures, 4)
	assert.Len(t, st.Warnings, 2)
	assert.Equal(t, "error1 value", st.Failures["remote/error1"])
	assert.Equal(t, "error2 value", st.Failures["remote/error2"])
	assert.Equal(t, "warning1 value", st.Warnings["remote/warning1"])
	assert.Equal(t, "e1", st.Failures["remote/multi/e1"])
	assert.Equal(t, "e2", st.Failures["remote/multi/e2"])
	assert.Equal(t, "w1", st.Warnings["remote/multi/w1"])
}

func TestRegisterRemote_asWarnings(t *testing.T) {
	remote := NewChecker(nil)
	defer remote.Close()
	setTestValues(remote)

	server := httptest.NewServer(remote.Handler())
	defer server.Close()

	local := NewChecker(nil)
	defer local.Close()
	err := local.RegisterRemote("remote", time.Second, server.URL, &RemoteOptions{
		AsWarnings: true,
	})
	assert.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	st := local.Status()
	assert.Len(t, st.Failures, 0)
	assert.Len(t, st.Warnings, 6)
	assert.Equal(t, "error1 value", st.Warnings["remote/error1"])
	assert.Equal(t, "error2 value", st.Warnings["remote/error2"])
	assert.Equal(t, "warning1 value", st.Warnings["remote/warning1"])
	assert.Equal(t, "e1", st.Warnings["remote/multi/e1"])
	assert.Equal(t, "e2", st.Warnings["remote/multi/e2"])
	assert.Equal(t, "w1", st.Warnings["remote/multi/w1"])
}
