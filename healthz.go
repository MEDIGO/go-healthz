// Package healthz provides an HTTP handler that returns information about the health status of the application.
package healthz

import (
	"encoding/json"
	"errors"
	"net/http"
	"runtime"
	"sync"
	"time"
)

var (
	// DefaultRuntimeTTL is the default TTL of the collected runtime stats.
	DefaultRuntimeTTL = 15 * time.Second
	// DefaultChecker is the default golbal checker referenced by the shortcut
	// functions in this package. Change this variable with caution.
	DefaultChecker = NewChecker(&Config{DefaultRuntimeTTL})
)

const (
	// StatusOK is returned when all the registered checks pass.
	StatusOK = "OK"
	// StatusUnavailable is returned when any of the registered checks fail.
	StatusUnavailable = "Unavailable"
)

// Status represents the service health status.
type Status struct {
	Status   string                 `json:"status"`
	Time     time.Time              `json:"time"`
	Since    time.Time              `json:"since"`
	Runtime  Runtime                `json:"runtime"`
	Failures map[string]string      `json:"failures,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
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

// CheckFunc is an application health check function.
type CheckFunc func() error

type check struct {
	mutex  sync.Mutex
	period time.Duration
	fn     CheckFunc
	err    error
	stopch chan bool
}

func (c *check) Do() {
	t := time.NewTicker(c.period)
	defer t.Stop()

	c.err = c.fn()

	for {
		select {
		case <-t.C:
			c.err = c.fn()
		case <-c.stopch:
			return
		}
	}
}

func (c *check) Close() {
	c.stopch <- true
}

func (c *check) Status() error {
	return c.err
}

// Config parameterizes a Checker.
type Config struct {
	RuntimeTTL time.Duration
}

// Checker is a health status checker responsible of evaluating the registered
// checks as well as of collecting useful runtime information about the Go Process.
// It provides an HTTP handler that returns the current health status.
type Checker struct {
	mutex      sync.Mutex
	since      time.Time
	metadata   map[string]interface{}
	checks     map[string]*check
	runtime    Runtime
	runtimeTTL time.Duration
}

// NewChecker creates a new Checker.
func NewChecker(config *Config) *Checker {
	if config == nil {
		config = &Config{}
	}

	if config.RuntimeTTL == 0 {
		config.RuntimeTTL = DefaultRuntimeTTL
	}

	return &Checker{
		since:      time.Now(),
		metadata:   make(map[string]interface{}),
		checks:     make(map[string]*check),
		runtimeTTL: config.RuntimeTTL,
	}
}

// Set sets a name value pair as a metadata entry to be returned with earch response.
// This can be used to store useful debug informaton like version numbers
// or git commit shas.
func (c *Checker) Set(name string, value interface{}) {
	c.metadata[name] = value
}

// Delete deletes a named entry from the configured metadata.
func (c *Checker) Delete(name string) {
	delete(c.metadata, name)
}

// Register registers a check to be evaluated each given period.
func (c *Checker) Register(name string, period time.Duration, fn CheckFunc) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if check := c.checks[name]; check != nil {
		check.Close()
	}

	check := &check{
		period: period,
		fn:     fn,
		err:    errors.New("pending"),
		stopch: make(chan bool, 1),
	}

	go check.Do()

	c.checks[name] = check
}

// Deregister deregisters a check.
func (c *Checker) Deregister(name string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	check := c.checks[name]
	if check == nil {
		return
	}

	check.Close()

	delete(c.checks, name)
}

// Handler returns an HTTP handler to be used as a health check endpoint. If the
// application is healthy and all the registered check pass, it returns a `200 OK`
// HTTP status code, otherwise, it fails with a `503 Service Unavailable` code.
// All responses contain a JSON encoded payload with information about the
// runtime system, current checks statuses and some configurable metadata.
func (c *Checker) Handler() http.Handler {
	return http.HandlerFunc(c.handle)
}

// Status returns the current service status.
func (c *Checker) Status() Status {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()

	if c.runtime.CollectedAt.Add(c.runtimeTTL).Before(now) {
		c.runtime = collect()
	}

	failures := make(map[string]string)
	for name, check := range c.checks {
		if err := check.Status(); err != nil {
			failures[name] = err.Error()
		}
	}

	status := StatusOK
	if len(failures) > 0 {
		status = StatusUnavailable
	}

	return Status{
		Status:   status,
		Time:     now,
		Since:    c.since,
		Runtime:  c.runtime,
		Metadata: c.metadata,
		Failures: failures,
	}
}

func (c *Checker) handle(w http.ResponseWriter, r *http.Request) {
	status := c.Status()

	code := http.StatusOK
	if status.Status != StatusOK {
		code = http.StatusServiceUnavailable
	}

	write(w, code, status)
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

func collect() Runtime {
	ms := runtime.MemStats{}
	runtime.ReadMemStats(&ms)
	return Runtime{
		CollectedAt:      time.Now(),
		Arch:             runtime.GOARCH,
		OS:               runtime.GOOS,
		Version:          runtime.Version(),
		GoroutinesCount:  runtime.NumGoroutine(),
		HeapObjectsCount: int(ms.HeapObjects),
		AllocBytes:       int(ms.Alloc),
		TotalAllocBytes:  int(ms.TotalAlloc),
	}
}

func write(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	data, err := json.Marshal(payload)
	if err != nil {
		return
	}

	w.Write(data)
}
