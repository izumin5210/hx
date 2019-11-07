package retry

import (
	"net/http"

	"github.com/cenkalti/backoff/v3"
	"github.com/izumin5210/hx"
)

func When(cond hx.ResponseHandlerCond, bo backoff.BackOff) hx.Option {
	return hx.TransportFrom(func(t http.RoundTripper) http.RoundTripper {
		return NewTransport(t, cond, bo)
	})
}
