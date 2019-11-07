# hx
[![CI](https://github.com/izumin5210/hx/workflows/CI/badge.svg)](https://github.com/izumin5210/hx/actions?workflow=CI)
[![GoDoc](https://godoc.org/github.com/izumin5210/hx?status.svg)](https://godoc.org/github.com/izumin5210/hx)
[![codecov](https://codecov.io/gh/izumin5210/hx/branch/master/graph/badge.svg)](https://codecov.io/gh/izumin5210/hx)
[![License](https://img.shields.io/github/license/izumin5210/hx)](./LICENSE)

Developer-friendly, Production-ready and extensible HTTP client for Go

## Features

...


### Plugins

- [retry](./retry)

## Examples
### Simple GET

```go
type Content struct {
	Body string `json:"body"`
}

var cont Content

ctx := context.Background()
err := hx.Get(ctx, "https://api.example.com/contents/1",
	hx.WhenSuccess(hx.AsJSON(&cont)),
	hx.WhenFailure(hx.AsError()),
)
```

### Real-world

```go
func init() {
	defaultTransport := hx.CloneTransport(http.DefaultTransport)

	// Tweak keep-alive configuration
	defaultTransport.MaxIdleConns = 500
	defaultTransport.MaxIdleConnsPerHost = 100

	// Set global options
	hx.DefaultOptions = append(
		hx.DefaultOptions,
		hx.UserAgent(fmt.Sprintf("yourapp (%s)", hx.DefaultUserAgent)),
		hx.Transport(defaultTransport),
		hx.TransportFrom(func(rt http.RoundTripper) http.RoundTripper {
			return &ochttp.Transport{Base: rt}
		}),
	)
}

func NewContentAPI() *hx.Client {
	// Set common options for API ciient
	return &ContentAPI{
		client: hx.NewClient(
			hx.BaseURL("https://api.example.com"),
		),
	}
}

type ContentAPI struct {
	client *hx.Client
}

func (a *ContentAPI) GetContent(ctx context.Context, id int) (*Content, error) {
	var cont Content

	err := a.client.Get(ctx, hx.Path("api", "contents", id),
		hx.WhenSuccess(hx.AsJSON(&cont)),
		hx.WhenFailure(hx.AsError()),
	)

	if err != nil {
		// ...
	}

	return &cont, nil
}

func (a *ContentAPI) CreateContent(ctx context.Context, in *Content) (*Content, error) {
	var out Content

	err := a.client.Post(ctx, "/api/contents",
		hx.JSON(in),
		hx.WhenSuccess(hx.AsJSON(&out)),
		hx.WhenStatus(hx.AsErrorOf(&InvalidArgument{}), http.StatusBadRequest),
		hx.WhenFailure(hx.AsError()),
	)

	if err != nil {
		var (
			invalidArgErr *InvalidArgument
			respErr       *hx.ResponseError
		)
		if errors.As(err, &invalidArgErr) {
			// handle known error
		} else if errors.As(err, &respErr) {
			// handle unknown response error
		} else {
			err := errors.Unwrap(err)
			// handle unknown error
		}
	}

	return &out, nil
}
```
