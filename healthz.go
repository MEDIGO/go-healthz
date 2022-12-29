// Package healthz provides an HTTP handler that returns information about the health status of the application.
package healthz

import (
	"net/http"
	"time"
)

var (
	// DefaultRuntimeTTL is the default TTL of the collected runtime stats.
	DefaultRuntimeTTL = 15 * time.Second
	// DefaultChecker is the default global checker referenced by the shortcut
	// functions in this package. Change this variable with caution.
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
	Failures    map[string]string      `json:"failures,omitempty"`
	Warnings    map[string]string      `json:"warnings,omitempty"`
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

// Set is a shortcut for DefaultChecker.Set. See there for more information.
func Set(name string, value interface{}) {
	DefaultChecker.Set(name, value)
}

// Delete is a shortcut for DefaultChecker.Delete. See there for more information.
func Delete(name string) {
	DefaultChecker.Delete(name)
}

// Handler is a shortcut for DefaultChecker.Handler. See there for more information.
func Handler() http.Handler {
	return DefaultChecker.Handler()
}

// Register is a shortcut for DefaultChecker.Register. See there for more information.
func Register(name string, period time.Duration, fn CheckFunc) {
	DefaultChecker.Register(name, period, fn)
}

// Deregister is a shortcut for DefaultChecker.Deregister. See there for more information.
func Deregister(name string) {
	DefaultChecker.Deregister(name)
}
