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

func (ch *check) Do() {
	if ch.static {
		ch.doStatic()
	} else {
		ch.doCallbacks()
	}
}

func (ch *check) doStatic() {
	// Static value set by SetMeta, with optional expiry
	if ch.expiry == 0 {
		return
	}
	t := time.NewTimer(ch.expiry)
	defer t.Stop()
	select {
	case <-t.C:
		ch.mutex.Lock()
		ch.err = Expired{expiry: ch.expiry}
		ch.mutex.Unlock()
	case <-ch.stopch:
	}
}

func (ch *check) doCallbacks() {
	// Periodic callbacks
	t := time.NewTicker(ch.period)
	defer t.Stop()

	ch.doOnce()
	for {
		select {
		case <-t.C:
			ch.doOnce()
		case <-ch.stopch:
			return
		}
	}
}

func (ch *check) doOnce() {
	// TODO: Perhaps log transitions?
	ch.mutex.Lock()
	defer ch.mutex.Unlock()
	ch.err = ch.fn()
}

func (ch *check) Close() {
	select {
	case ch.stopch <- true:
	default: // do not block if called twice
	}
}

func (ch *check) Status() error {
	ch.mutex.Lock()
	defer ch.mutex.Unlock()
	return ch.err
}
