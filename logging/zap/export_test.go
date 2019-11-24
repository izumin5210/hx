package zap

import "time"

func SetNow(f func() time.Time) func() {
	tmp := now
	now = f
	return func() { now = tmp }
}

func SetSince(f func(time.Time) time.Duration) func() {
	tmp := since
	since = f
	return func() { since = tmp }
}
