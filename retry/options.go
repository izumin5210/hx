package retry

import (
	"context"
	"net/http"

	"github.com/cenkalti/backoff/v3"
	"github.com/google/uuid"
	"github.com/izumin5210/hx"
)

func When(cond hx.ResponseHandlerCond, newBackOff NewBackOff) hx.ClientOption {
	return hx.CombineClientOptions(
		IdempotencyKey(),
		hx.Transport(func(_ context.Context, t http.RoundTripper) http.RoundTripper {
			return NewTransport(t, cond, newBackOff)
		}),
	)
}

type NewBackOff func() backoff.BackOff

func IdempotencyKey() hx.ClientOption {
	return hx.ClientOptionFunc(func(c *hx.ClientConfig) {
		c.RequestOptions = append(c.RequestOptions, setIdempotencyKey)
	})
}

func setIdempotencyKey(_ context.Context, r *http.Request) error {
	if r.Header.Get("Idempotency-Key") != "" {
		return nil
	}
	r.Header.Set("Idempotency-Key", uuid.New().String())
	return nil
}
