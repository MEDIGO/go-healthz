package healthz

import (
	"errors"
	"runtime"
	"sync"
	"time"
)

// Config parameterizes a Checker.
type Config struct {
	// RuntimeTTL is the time between checking runtime stats like memory usage.
	// It defaults to DefaultRuntimeTTL.
	RuntimeTTL time.Duration
}

// Checker is a health status checker responsible for evaluating the registered
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
// Using a custom Checker instead of the global DefaultChecker is not recommended.
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

// SetMeta sets a name value pair as a metadata entry to be returned with each response.
// This can be used to store useful debug information like version numbers
// or git commit hashes.
func (c *Checker) SetMeta(name string, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.metadata[name] = value
}

// DeleteMeta deletes a named entry from the configured metadata.
func (c *Checker) DeleteMeta(name string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.metadata, name)
}

// Register registers a check to be evaluated each given period.
func (c *Checker) Register(name string, period time.Duration, fn CheckFunc) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if ch := c.checks[name]; ch != nil {
		// TODO: error or log warning when registering the same twice?
		ch.Close()
	}

	ch := &check{
		static: false,
		period: period,
		fn:     fn,
		err:    errors.New("pending"),
		stopch: make(chan bool, 1),
	}

	go ch.Do()

	c.checks[name] = ch
}

// SetMeta sets a static status value without a periodic checker function.
// This can be useful if your application has an event loop that can directly
// update the status for real-time information, instead of relying on a
// checker function to run periodically.
// If the expiry duration is not 0, the status will be reset to Expired
// after this duration, if no new value is set in the meantime.
func (c *Checker) Set(name string, err error, expiry time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if ch := c.checks[name]; ch != nil {
		// TODO: error or log warning when registering the same twice?
		ch.Close()
	}

	ch := &check{
		static: true,
		expiry: expiry,
		err:    err,
		stopch: make(chan bool, 1),
	}

	go ch.Do()

	c.checks[name] = ch
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

// Status returns the current service status.
func (c *Checker) Status() Status {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()

	if c.runtime.CollectedAt.Add(c.runtimeTTL).Before(now) {
		c.runtime = collect()
	}

	failures := make(map[string]string)
	warnings := make(map[string]string)
	for name, check := range c.checks {
		if err := check.Status(); err != nil {
			if IsWarning(err) {
				warnings[name] = err.Error()
			} else {
				failures[name] = err.Error()
			}
		}
	}

	status := StatusOK
	if len(failures) > 0 {
		status = StatusUnavailable
	} else if len(warnings) > 0 {
		status = StatusWarning
	}

	return Status{
		OK:          len(failures) == 0,
		HasWarnings: len(warnings) > 0,
		Status:      status,
		Time:        now,
		Since:       c.since,
		Runtime:     c.runtime,
		Metadata:    c.metadata,
		Failures:    failures,
		Warnings:    warnings,
	}
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
