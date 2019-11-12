// A plugin for retry HTTP requests.
package retry

import (
	"net/http"

	"github.com/cenkalti/backoff/v3"
	"github.com/izumin5210/hx"
)

// When creates an option that provides retry mechanism for your http client.
//  bo := backoff.NewExponentialBackOff()
//  bo.InitialInterval = 50 * time.Millisecond
//  bo.MaxInterval = 500 * time.Millisecond
//
//  err := hx.Post(ctx, "https://example.com/api/messages",
//  	retry.When(hx.Any(hx.IsServerError(), hx.IsTemporaryError()), bo),
//  	hx.JSON(&in),
//  	hx.WhenSuccess(hx.AsJSON(&out)),
//  	hx.WhenFailure(hx.AsError()),
//  )
func When(cond hx.ResponseHandlerCond, bo backoff.BackOff) hx.Option {
	return hx.TransportFrom(func(t http.RoundTripper) http.RoundTripper {
		return NewTransport(t, cond, bo)
	})
}
