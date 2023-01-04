package healthz

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// RemoteOptions are options passed to RegisterRemote
type RemoteOptions struct {
	// Client allows you to override the default http.Client
	Client *http.Client // optional client override

	// Timeout allows you to override the default timeout of RemoteDefaultTimeout
	// used by RegisterRemote. If the Client is overridden, this does nothing.
	Timeout time.Duration // optional timeout, if the default is not OK

	// AsWarnings instructs RegisterRemote to downgrade any remote errors to
	// warnings for this check.
	AsWarnings bool

	// Warn404 make a remote 404 a warning instead of an error. This is useful
	// if you are not sure if the target url has a healthz endpoint.
	Warn404 bool
}

const (
	// RemoteDefaultTimeout is the default timeout for fetching a remote
	// healthz in RegisterRemote.
	RemoteDefaultTimeout = 10 * time.Second
)

// RegisterRemote registers a remote /healthz endpoint that needs to be monitored.
// The period parameter determines the poll interval.
// The url must contain the full url to the healthz endpoint.
//
// If the remote endpoint uses the same JSON structure as this instance does,
// the name is used to prefix all remote failures and warnings, separated
// by a slash ('/'). E.g. if the name is "foo", a remote "bar" error will end
// up as "foo/bar" in the status reported by this instance.
// It the remote endpoint is not served by this library and no compatible
// failures and warnings keys could be found, this check returns a single error
// for this endpoint with the requested name.
func (c *Checker) RegisterRemote(name string, period time.Duration, url string, opt *RemoteOptions) error {
	var client *http.Client
	if opt != nil && opt.Client != nil {
		client = opt.Client
	} else {
		timeout := RemoteDefaultTimeout
		if opt != nil && opt.Timeout > 0 {
			timeout = opt.Timeout
		}
		client = &http.Client{
			Timeout: timeout,
		}
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	asWarnings := false
	errorf := fmt.Errorf
	if opt != nil {
		asWarnings = opt.AsWarnings
		errorf = Warnf
	}

	c.Register(name, period, func() error {
		res, err := client.Do(req)
		if err != nil {
			return err
		}
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		// Accept error codes in the 2xx and 5xx range
		sc := res.StatusCode
		if sc < 200 || (sc >= 300 && sc < 500) || sc >= 600 {
			if sc == 404 && opt != nil && opt.Warn404 {
				return Warnf("remote healthz endpoint does not exist")
			}
			return errorf("unexpected healthz http status code: %d", sc)
		}
		remoteOK := sc < 300

		// Try to decode as json
		var st Status
		if err := json.Unmarshal(body, &st); err != nil {
			// Not JSON or not the expected format, base on http status code
			if remoteOK {
				return nil
			} else {
				return errorf("remote http code %d, contents:\n%s", sc, string(body))
			}
		}
		if !remoteOK && len(st.Failures) == 0 {
			// No failures listed, but we got an error status code, so the remote
			// json is not compatible with our json.
			return errorf("remote http code %d, contents:\n%s", sc, string(body))
		}

		// Extract failures and warnings from another instance that uses the
		// same reporting format.
		me := make(ScopedMultiError)
		for key, msg := range st.Failures {
			if asWarnings {
				// Downgrade to warning if requested in RemoteOptions
				me[key] = Warn(msg)
			} else {
				me[key] = errors.New(msg)
			}
		}
		for key, msg := range st.Warnings {
			me[key] = Warn(msg)
		}
		if len(me) == 0 {
			return nil
		}
		return me
	})

	return nil
}
