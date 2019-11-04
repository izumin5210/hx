package httpx

import (
	"context"
	"net/http"
)

func Get(ctx context.Context, url string, opts ...ClientOption) (*http.Response, error) {
	return NewClient().Get(ctx, url, opts...)
}

func Post(ctx context.Context, url string, opts ...ClientOption) (*http.Response, error) {
	return NewClient().Post(ctx, url, opts...)
}

func Put(ctx context.Context, url string, opts ...ClientOption) (*http.Response, error) {
	return NewClient().Put(ctx, url, opts...)
}

func Patch(ctx context.Context, url string, opts ...ClientOption) (*http.Response, error) {
	return NewClient().Patch(ctx, url, opts...)
}

func Delete(ctx context.Context, url string, opts ...ClientOption) (*http.Response, error) {
	return NewClient().Delete(ctx, url, opts...)
}

type Client struct {
	opts []ClientOption
}

// NewClient creates a new http client instance.
func NewClient(opts ...ClientOption) *Client {
	return &Client{
		opts: opts,
	}
}

func (c *Client) Get(ctx context.Context, url string, opts ...ClientOption) (*http.Response, error) {
	resp, err := c.request(ctx, http.MethodGet, url, opts...)
	return resp, err
}

func (c *Client) Post(ctx context.Context, url string, opts ...ClientOption) (*http.Response, error) {
	resp, err := c.request(ctx, http.MethodPost, url, opts...)
	return resp, err
}

func (c *Client) Put(ctx context.Context, url string, opts ...ClientOption) (*http.Response, error) {
	resp, err := c.request(ctx, http.MethodPut, url, opts...)
	return resp, err
}

func (c *Client) Patch(ctx context.Context, url string, opts ...ClientOption) (*http.Response, error) {
	resp, err := c.request(ctx, http.MethodPatch, url, opts...)
	return resp, err
}

func (c *Client) Delete(ctx context.Context, url string, opts ...ClientOption) (*http.Response, error) {
	resp, err := c.request(ctx, http.MethodDelete, url, opts...)
	return resp, err
}

func (c *Client) request(ctx context.Context, meth string, url string, opts ...ClientOption) (*http.Response, error) {
	var err error

	cfg := new(ClientConfig)
	err = cfg.apply(c.opts...)
	if err != nil {
		return nil, err
	}
	err = WithURL(url)(cfg)
	if err != nil {
		return nil, err
	}
	err = cfg.apply(opts...)
	if err != nil {
		return nil, err
	}

	resp, err := cfg.do(ctx, meth)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
