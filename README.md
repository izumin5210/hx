# hx
[![GoDoc](https://godoc.org/github.com/izumin5210/hx?status.svg)](https://godoc.org/github.com/izumin5210/hx)
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
	hx.WhenOK(hx.AsJSON(&cont)),
	hx.WhenNotOK(hx.AsError()),
)
```

### Real-world

```go
func init() {
	// https://github.com/golang/go/blob/go1.13.4/src/net/http/transport.go#L42-L54
	defaultTransport := &http.Transport{
		Proxy: ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	// Tweak keep-alive configuration
	defaultTransport.MaxIdleConns = 500
	defaultTransport.MaxIdleConnsPerHost = 100

	// Set global options
	hx.DefaultClientOptions = append(
		hx.DefaultClientOptions,
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
		hx.WhenOK(hx.AsJSON(&cont)),
		hx.WhenNotOK(hx.AsError()),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get content: %w", err)
	}

	return &cont, nil
}

func (a *ContentAPI) CreateContent(ctx context.Context, in *Content) (*Content, error) {
	var out Content

	err := a.client.Post(ctx, "/api/contents",
		hx.JSON(in),
		hx.WhenOK(hx.AsJSON(&out)),
		hx.WhenNotOK(hx.AsError()),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create content: %w", err)
	}

	return &out, nil
}
```
