# retry

Backoff retry with `Idempotency-Key`

```go
var cont Content

bo := backoff.NewExponentialBackOff()
bo.InitialInterval = 50 * time.Millisecond
bo.MaxInterval = 500 * time.Millisecond

err := hx.Get(ctx, "https://api.example.com/contents/1",
	retry.When(hx.Any(hx.IsServerError(), hx.IsNetworkError()), bo),
	hx.WhenSuccess(hx.AsJSON(&cont)),
	hx.WhenFailure(hx.AsError()),
)
```
