package healthz

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHealthz(t *testing.T) {
	s := httptest.NewServer(Handler())
	defer s.Close()

	// it should get the health status
	status, code, err := get(s.URL)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, code)
	require.Len(t, status.Metadata, 0)
	require.Len(t, status.Failures, 0)

	// it should set a metadata value
	Set("version", "1.0.0")

	status, code, err = get(s.URL)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "1.0.0", status.Metadata["version"])

	// it should delete a metadata value
	Delete("version")

	status, code, err = get(s.URL)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, code)
	require.Nil(t, status.Metadata["version"])

	// it register a check that suceedes the first time is check but fails afterwards.
	Register("a_check", 3*time.Second, func() CheckFunc {
		var err error
		return func() error {
			if err == nil {
				err = errors.New("bad thing happened")
				return nil
			}
			return err
		}
	}())

	status, code, err = get(s.URL)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, code)
	require.Empty(t, status.Failures["a_check"])

	time.Sleep(3 * time.Second)

	status, code, err = get(s.URL)
	require.NoError(t, err)

	require.Equal(t, http.StatusServiceUnavailable, code)
	require.Equal(t, "bad thing happened", status.Failures["a_check"])
}

func get(url string) (*Status, int, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	status := new(Status)
	return status, res.StatusCode, json.NewDecoder(res.Body).Decode(status)
}
