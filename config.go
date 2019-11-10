package hx

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

type Config struct {
	URLOptions       []func(context.Context, *url.URL) error
	BodyOption       func(context.Context) (io.Reader, error)
	ClientOptions    []func(context.Context, *http.Client) error
	RequestHandlers  []RequestHandler
	ResponseHandlers []ResponseHandler
}

func NewConfig() *Config {
	cfg := new(Config)
	cfg.Apply(DefaultOptions...)
	return cfg
}

func (cfg *Config) Apply(opts ...Option) {
	for _, f := range opts {
		f.ApplyOption(cfg)
	}
}

func (cfg *Config) DoRequest(ctx context.Context, meth string) (*http.Response, error) {
	url, err := cfg.buildURL(ctx)
	if err != nil {
		return nil, err
	}

	body, err := cfg.buildBody(ctx)
	if err != nil {
		return nil, err
	}

	cli, err := cfg.buildClient(ctx)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(meth, url.String(), body)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)

	for _, h := range cfg.RequestHandlers {
		req, err = h(req)
		if err != nil {
			return nil, err
		}
	}

	resp, err := cli.Do(req)

	for _, h := range cfg.ResponseHandlers {
		resp, err = h(resp, err)
		if err != nil {
			return nil, err
		}
	}

	return resp, err
}

func (cfg *Config) buildURL(ctx context.Context) (*url.URL, error) {
	u := new(url.URL)

	for _, f := range cfg.URLOptions {
		err := f(ctx, u)
		if err != nil {
			return nil, err
		}
	}

	return u, nil
}

func (cfg *Config) buildBody(ctx context.Context) (io.Reader, error) {
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

func (cfg *Config) buildClient(ctx context.Context) (*http.Client, error) {
	c := new(http.Client)

	for _, f := range cfg.ClientOptions {
		err := f(ctx, c)
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}
