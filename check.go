package healthz

import (
	"fmt"
	"sync"
	"time"
)

// CheckFunc is an application health check function.
type CheckFunc func() error

// Expired is the error status set after a status set by SetMeta has
// expired.
type Expired struct {
	expiry time.Duration
}

func (e Expired) Error() string {
	return fmt.Sprintf("status expired after %s", e.expiry)
}

type check struct {
	mutex  sync.Mutex
	period time.Duration
	expiry time.Duration
	static bool
	fn     CheckFunc
	err    error
	stopch chan bool
}

func (c *check) Do() {
	if c.static {
		c.doStatic()
	} else {
		c.doCallbacks()
	}
}

func (c *check) doStatic() {
	// Static value set by SetMeta, with optional expiry
	if c.expiry == 0 {
		return
	}
	t := time.NewTimer(c.expiry)
	defer t.Stop()
	select {
	case <-t.C:
		c.mutex.Lock()
		c.err = Expired{expiry: c.expiry}
		c.mutex.Unlock()
	case <-c.stopch:
	}
}

func (c *check) doCallbacks() {
	// Periodic callbacks
	t := time.NewTicker(c.period)
	defer t.Stop()

	c.doOnce()
	for {
		select {
		case <-t.C:
			c.doOnce()
		case <-c.stopch:
			return
		}
	}
}

func (c *check) doOnce() {
	// TODO: Perhaps log transitions?
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.err = c.fn()
}

func (c *check) Close() {
	select {
	case c.stopch <- true:
	default: // do not block if called twice
	}
}

func (c *check) Status() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.err
}
