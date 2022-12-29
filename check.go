package healthz

import (
	"sync"
	"time"
)

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
