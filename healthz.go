// Package healthz provides an HTTP handler that returns information about the health status of the application.
package healthz

import (
	"net/http"
	"time"
)

const (
	// DefaultRuntimeTTL is the default TTL of the collected runtime stats.
	DefaultRuntimeTTL = 15 * time.Second

	// DefaultCheckPeriod is the default check period if 0 is passed to Register
	DefaultCheckPeriod = time.Second
)

var (
	// DefaultChecker is the default global checker referenced by the shortcut
	// functions in this package. Change this variable with caution, because
	// you will lose any checkers that have already been registered to the
	// old one.
	DefaultChecker = NewChecker(&Config{DefaultRuntimeTTL})
)

const (
	// StatusOK is returned when all the registered checks pass.
	StatusOK = "OK"
	// StatusUnavailable is returned when any of the registered checks fail.
	StatusUnavailable = "Unavailable"
	// StatusWarning is returned when one or more registered check returns
	// a warning and none returns a fatal error.
	StatusWarning = "Warning"
)

// Status represents the service health status.
type Status struct {
	OK          bool                   `json:"ok"` // May have warnings
	HasWarnings bool                   `json:"has_warnings"`
	Status      string                 `json:"status"`
	Time        time.Time              `json:"time"`
	Since       time.Time              `json:"since"`
	Runtime     Runtime                `json:"runtime"`
	Failures    map[string]string      `json:"failures"`
	Warnings    map[string]string      `json:"warnings"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Runtime contains statistics about the Go's process.
type Runtime struct {
	CollectedAt      time.Time `json:"-"`
	Arch             string    `json:"arch"`
	OS               string    `json:"os"`
	Version          string    `json:"version"`
	GoroutinesCount  int       `json:"goroutines_count"`
	HeapObjectsCount int       `json:"heap_objects_count"`
	AllocBytes       int       `json:"alloc_bytes"`
	TotalAllocBytes  int       `json:"total_alloc_bytes"`
}

// SetMeta is a shortcut for DefaultChecker.SetMeta. See there for more information.
func SetMeta(name string, value interface{}) {
	DefaultChecker.SetMeta(name, value)
}

// DeleteMeta is a shortcut for DefaultChecker.DeleteMeta. See there for more information.
func DeleteMeta(name string) {
	DefaultChecker.DeleteMeta(name)
}

// Handler is a shortcut for DefaultChecker.Handler. See there for more information.
func Handler() http.Handler {
	return DefaultChecker.Handler()
}

// Register is a shortcut for DefaultChecker.Register. See there for more information.
func Register(name string, period time.Duration, fn CheckFunc) {
	DefaultChecker.Register(name, period, fn)
}

// Set is a shortcut for DefaultChecker.Set. See there for more information.
func Set(name string, err error, timeout time.Duration) {
	DefaultChecker.Set(name, err, timeout)
}

// RegisterRemote registers a remote /healthz endpoint that needs to be monitored.
// See Checker.RegisterRemote for details.
func RegisterRemote(name string, period time.Duration, url string, opt *RemoteOptions) error {
	return DefaultChecker.RegisterRemote(name, period, url, opt)
}

// Deregister is a shortcut for DefaultChecker.Deregister. See there for more information.
func Deregister(name string) {
	DefaultChecker.Deregister(name)
}

// AddBuildInfo adds some build info like VCS from debug.ReadBuildInfo to the
// metadata.
func AddBuildInfo() {
	DefaultChecker.AddBuildInfo()
}
