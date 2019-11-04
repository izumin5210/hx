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
	BodyOption       func(context.Context) (io.Reader, error)
	ResponseHandlers []ResponseHandler
}

func (cfg *ClientConfig) Apply(opts ...ClientOption) {
	for _, f := range opts {
		f.Apply(cfg)
	}
}

func (cfg *ClientConfig) DoRequest(ctx context.Context, meth string) (*http.Response, error) {
	url, err := cfg.buildURL(ctx)
	if err != nil {
		return nil, err
	}

	body, err := cfg.buildBody(ctx)
	if err != nil {
		return nil, err
	}

	req, err := cfg.buildRequest(ctx, url, meth, body)
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

func (cfg *ClientConfig) buildBody(ctx context.Context) (io.Reader, error) {
	f := cfg.BodyOption
	if f == nil {
		return nil, nil
	}
	body, err := cfg.BodyOption(ctx)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (cfg *ClientConfig) buildRequest(ctx context.Context, url *url.URL, method string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url.String(), body)
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
