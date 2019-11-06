package hx

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

type Config struct {
	URLOptions   []func(context.Context, *url.URL) error
	BodyOption   func(context.Context) (io.Reader, error)
	Interceptors []Interceptor
}

func (cfg *Config) Apply(opts ...Option) {
	for _, f := range opts {
		f.Apply(cfg)
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

	req, err := http.NewRequest(meth, url.String(), body)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)

	handler := wrapHandler(
		combineInterceptors(cfg.Interceptors),
		func(c *http.Client, r *http.Request) (*http.Response, error) { return c.Do(r) },
	)

	return handler(new(http.Client), req)
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
