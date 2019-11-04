package httpx

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

type ClientConfig struct {
	RequestOptions   []func(context.Context, *http.Request) error
	ClientOptions    []func(context.Context, *http.Client) error
	URLOptions       []func(context.Context, *url.URL) error
	Body             io.Reader
	ResponseHandlers []func(*http.Client, *http.Response, error) (*http.Response, error)
}

func (cfg *ClientConfig) apply(opts ...ClientOption) error {
	for _, f := range opts {
		err := f(cfg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cfg *ClientConfig) do(ctx context.Context, meth string) (*http.Response, error) {
	url, err := cfg.buildURL(ctx)
	if err != nil {
		return nil, err
	}

	req, err := cfg.buildRequest(ctx, url, meth)
	if err != nil {
		return nil, err
	}

	cli, err := cfg.buildClient(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := cli.Do(req)

	for _, h := range cfg.ResponseHandlers {
		resp, err = h(cli, resp, err)
	}

	return resp, err
}

func (cfg *ClientConfig) buildURL(ctx context.Context) (*url.URL, error) {
	u := new(url.URL)

	for _, f := range cfg.URLOptions {
		err := f(ctx, u)
		if err != nil {
			return nil, err
		}
	}

	return u, nil
}

func (cfg *ClientConfig) buildRequest(ctx context.Context, url *url.URL, method string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url.String(), cfg.Body)
	if err != nil {
		return nil, err
	}

	for _, f := range cfg.RequestOptions {
		err = f(ctx, req)
		if err != nil {
			return nil, err
		}
	}

	return req, nil
}

func (cfg *ClientConfig) buildClient(ctx context.Context) (*http.Client, error) {
	cli := new(http.Client)

	for _, f := range cfg.ClientOptions {
		err := f(ctx, cli)
		if err != nil {
			return nil, err
		}
	}

	return cli, nil
}
