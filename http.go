package healthz

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Handler returns an HTTP handler to be used as a health check endpoint. If the
// application is healthy and all the registered check pass, it returns a `200 OK`
// HTTP status code, otherwise, it fails with a `503 Service Unavailable` code.
// All responses contain a JSON encoded payload with information about the
// runtime system, current checks statuses and some configurable metadata.
func (c *Checker) Handler() http.Handler {
	return http.HandlerFunc(c.handle)
}

func (c *Checker) handle(w http.ResponseWriter, r *http.Request) {
	// TODO: param to convert warning into errors?

	status := c.Status()

	code := http.StatusOK
	if !status.OK {
		code = http.StatusServiceUnavailable
	}

	data, err := json.Marshal(status)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(w, "internal healthz error: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(data)
}
