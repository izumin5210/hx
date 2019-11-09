package hxutil

import "net/http"

type RoundTripperFunc func(*http.Request, http.RoundTripper) (*http.Response, error)

func (f RoundTripperFunc) Wrap(rt http.RoundTripper) http.RoundTripper {
	return &RoundTripperWrapper{Next: rt, Func: f}
}

type RoundTripperWrapper struct {
	Next http.RoundTripper
	Func func(*http.Request, http.RoundTripper) (*http.Response, error)
}

func (w *RoundTripperWrapper) RoundTrip(r *http.Request) (*http.Response, error) {
	next := w.Next
	if next == nil {
		next = http.DefaultTransport
	}
	return w.Func(r, next)
}
