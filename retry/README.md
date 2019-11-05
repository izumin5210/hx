# retry

Backoff retry with `Idempotency-Key`

```go
var cont Content

newBackOff := func() backoff.BackOff {
	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = 50 * time.Millisecond
	bo.MaxInterval = 500 * time.Millisecond
	return bo
}

err := hx.Get(ctx, "https://api.example.com/contents/1",
	retry.When(hx.Any(hx.IsServerError(), hx.IsNetworkError()), newBackOff),
	hx.WhenOK(hx.AsJSON(&cont)),
	hx.WhenStatus(hx.AsError()),
)
```
