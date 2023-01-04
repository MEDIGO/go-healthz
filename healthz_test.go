package healthz_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/wojas/go-healthz"
)

func TestHealthz(t *testing.T) {
	ch := healthz.NewChecker(nil)

	s := httptest.NewServer(ch.Handler())
	defer s.Close()

	// it should get the health status
	status, code, err := get(s.URL)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, code)
	require.True(t, status.OK)
	require.False(t, status.HasWarnings)
	require.Len(t, status.Metadata, 0)
	require.Len(t, status.Failures, 0)
	require.Len(t, status.Warnings, 0)

	// it should set a metadata value
	ch.SetMeta("version", "1.0.0")

	status, code, err = get(s.URL)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "1.0.0", status.Metadata["version"])

	// it should delete a metadata value
	ch.DeleteMeta("version")

	status, code, err = get(s.URL)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, code)
	require.Nil(t, status.Metadata["version"])

	// register a check that succeeds until an error is set through
	// this channel.
	checkErrorChan := make(chan error)
	ch.Register("a_check", 10*time.Millisecond, func() healthz.CheckFunc {
		var curErr error
		return func() error {
			select {
			case err, ok := <-checkErrorChan:
				if ok {
					curErr = err
				}
			default:
			}
			return curErr
		}
	}())
	defer ch.Deregister("a_check")

	// TODO: Also test static status

	status, code, err = get(s.URL)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, code)
	require.True(t, status.OK)
	require.Empty(t, status.Failures["a_check"])

	ch.Register("a_warning", time.Second, func() error {
		return healthz.Warn("test warning")
	})
	defer ch.Deregister("a_warning")

	status, code, err = get(s.URL)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, code)
	require.True(t, status.OK)
	require.True(t, status.HasWarnings)
	require.Empty(t, status.Failures["a_warning"])
	require.Equal(t, "test warning", status.Warnings["a_warning"])

	// SetMeta an error
	checkErrorChan <- errors.New("bad thing happened")

	status, code, err = get(s.URL)
	require.NoError(t, err)
	require.Equal(t, http.StatusServiceUnavailable, code)
	require.False(t, status.OK)
	require.Equal(t, "bad thing happened", status.Failures["a_check"])

	// Reset the error
	ch.Deregister("a_check")
	status, code, err = get(s.URL)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, code)
	require.True(t, status.OK)
	require.Empty(t, status.Failures["a_check"])

	// SetMeta a static error
	ch.Set("static", errors.New("static error"), 200*time.Millisecond)
	defer ch.Deregister("static")

	status, code, err = get(s.URL)
	require.NoError(t, err)
	require.Equal(t, http.StatusServiceUnavailable, code)
	require.False(t, status.OK)
	require.Equal(t, "static error", status.Failures["static"])

	// SetMeta to replace a static error
	ch.Set("static", errors.New("new static error"), 200*time.Millisecond)
	defer ch.Deregister("static")

	status, code, err = get(s.URL)
	require.NoError(t, err)
	require.Equal(t, http.StatusServiceUnavailable, code)
	require.False(t, status.OK)
	require.Equal(t, "new static error", status.Failures["static"])

	time.Sleep(200 * time.Millisecond)

	status, code, err = get(s.URL)
	require.NoError(t, err)
	require.Equal(t, http.StatusServiceUnavailable, code)
	require.False(t, status.OK)
	require.Equal(t, "status expired after 200ms", status.Failures["static"])
}

func get(url string) (*healthz.Status, int, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	status := new(healthz.Status)
	return status, res.StatusCode, json.NewDecoder(res.Body).Decode(status)
}
