package hx

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
)

func Get(ctx context.Context, url string, opts ...Option) error {
	return NewClient().Get(ctx, url, opts...)
}

func Post(ctx context.Context, url string, opts ...Option) error {
	return NewClient().Post(ctx, url, opts...)
}

func Put(ctx context.Context, url string, opts ...Option) error {
	return NewClient().Put(ctx, url, opts...)
}

func Patch(ctx context.Context, url string, opts ...Option) error {
	return NewClient().Patch(ctx, url, opts...)
}

func Delete(ctx context.Context, url string, opts ...Option) error {
	return NewClient().Delete(ctx, url, opts...)
}

type Client struct {
	opts []Option
}

// NewClient creates a new http client instance.
func NewClient(opts ...Option) *Client {
	return &Client{
		opts: opts,
	}
}

func (c *Client) Get(ctx context.Context, url string, opts ...Option) error {
	return c.request(ctx, http.MethodGet, url, opts...)
}

func (c *Client) Post(ctx context.Context, url string, opts ...Option) error {
	return c.request(ctx, http.MethodPost, url, opts...)
}

func (c *Client) Put(ctx context.Context, url string, opts ...Option) error {
	return c.request(ctx, http.MethodPut, url, opts...)
}

func (c *Client) Patch(ctx context.Context, url string, opts ...Option) error {
	return c.request(ctx, http.MethodPatch, url, opts...)
}

func (c *Client) Delete(ctx context.Context, url string, opts ...Option) error {
	return c.request(ctx, http.MethodDelete, url, opts...)
}

// With clones the current client and applies the given options.
func (c *Client) With(opts ...Option) *Client {
	newOpts := make([]Option, 0, len(c.opts)+len(opts))
	newOpts = append(newOpts, c.opts...)
	newOpts = append(newOpts, opts...)
	return NewClient(newOpts...)
}

func (c *Client) request(ctx context.Context, meth string, url string, opts ...Option) error {
	cfg, err := NewConfig()
	if err != nil {
		return err
	}
	err = cfg.Apply(c.opts...)
	if err != nil {
		return err
	}
	err = cfg.Apply(URL(url))
	if err != nil {
		return err
	}
	err = cfg.Apply(opts...)
	if err != nil {
		return err
	}

	resp, err := cfg.DoRequest(ctx, meth)
	if err != nil {
		return err
	}

	io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()

	return nil
}
