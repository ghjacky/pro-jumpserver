package utils

import (
	"time"
)

type TimeoutChan struct {
	CH      chan []byte
	timeout time.Duration
	timer   *time.Timer
}

func NewTimeoutChan(ch chan []byte) *TimeoutChan {
	return &TimeoutChan{
		CH: ch,
	}
}

func (tc *TimeoutChan) SetTimeout(duration time.Duration) {
	tc.timeout = duration
	tc.timer = time.NewTimer(tc.timeout)
}

func (tc *TimeoutChan) ReadWithTimeout() ([]byte, bool) {
	select {
	case data := <-tc.CH:
		return data, false
	case <-tc.timer.C:
		tc.timer.Stop()
		return nil, true
	}
}

func (tc *TimeoutChan) WriteWithTimeout(data []byte) bool {
	select {
	case tc.CH <- data:
		return false
	case <-tc.timer.C:
		tc.timer.Stop()
		return true
	}
}
