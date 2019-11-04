# hx
[![GoDoc](https://godoc.org/github.com/izumin5210/hx?status.svg)](https://godoc.org/github.com/izumin5210/hx)
[![License](https://img.shields.io/github/license/izumin5210/hx)](./LICENSE)

Developer-friendly, Production-ready and extensible HTTP client for Go

## Examples
### Simple GET

```go
type Content struct {
	Body string `json:"body"`
}

var cont Content

ctx := context.Background()
err := hx.Get(ctx, "https://api.example.com/contents/1",
	hx.WhenOK(httpx.AsJSON(&cont)),
	hx.WhenNotOK(httpx.AsError()),
)
```

### Real-world

```go
func init() {
	// Set global options
	hx.DefaultClientOptions = append(
		hx.DefaultClientOptions,
		hx.UserAgent(fmt.Sprintf("yourapp (%s)", hx.DefaultUserAgent)),
		hx.Transport(func(_ context.Context, rt http.RoundTripper) http.RoundTripper {
			return &ochttp.Transport{}
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

	err := a.client.Get(ctx, fmt.Sprintf("/api/contents/%d", id),
		hx.WhenOK(httpx.AsJSON(&cont)),
		hx.WhenNotOK(httpx.AsError()),
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
		hx.WhenOK(httpx.AsJSON(&out)),
		hx.WhenNotOK(httpx.AsError()),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create content: %w", err)
	}

	return &out, nil
}
```
