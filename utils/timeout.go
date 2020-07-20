package utils

import (
	"time"
)

type TimeoutChan struct {
	CH      chan []byte
	timeout time.Duration
	done    <-chan time.Time
}

func NewTimeoutChan(ch chan []byte) *TimeoutChan {
	return &TimeoutChan{
		CH: ch,
	}
}

func (tc *TimeoutChan) SetTimeout(duration time.Duration) {
	tc.timeout = duration
	tc.done = time.Tick(tc.timeout)
}

func (tc *TimeoutChan) Wait() {
	<-tc.done
}

func (tc *TimeoutChan) ResetTicker() {
	tc.done = time.Tick(tc.timeout)
}

func (tc *TimeoutChan) ReadWithTimeout() ([]byte, bool) {
	select {
	case data := <-tc.CH:
		return data, false
	case <-tc.done:
		return nil, true
	}
}

func (tc *TimeoutChan) WriteWithTimeout(data []byte) bool {
	select {
	case tc.CH <- data:
		return false
	case <-tc.done:
		return true
	}
}
