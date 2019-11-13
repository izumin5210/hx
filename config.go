package hx

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

type Config struct {
	URL              *url.URL
	Body             io.Reader
	HTTPClient       *http.Client
	QueryParams      url.Values
	RequestHandlers  []RequestHandler
	ResponseHandlers []ResponseHandler
}

func NewConfig() (*Config, error) {
	cfg := &Config{URL: new(url.URL), HTTPClient: new(http.Client), QueryParams: url.Values{}}
	err := cfg.Apply(DefaultOptions...)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (cfg *Config) Apply(opts ...Option) error {
	for _, f := range opts {
		err := f.ApplyOption(cfg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cfg *Config) DoRequest(ctx context.Context, meth string) (*http.Response, error) {
	q, err := url.ParseQuery(cfg.URL.RawQuery)
	if err != nil {
		return nil, err
	}
	for k, values := range cfg.QueryParams {
		for _, v := range values {
			q.Add(k, v)
		}
	}
	cfg.URL.RawQuery = q.Encode()

	req, err := newRequest(ctx, meth, cfg.URL, cfg.Body)
	if err != nil {
		return nil, err
	}

	for _, h := range cfg.RequestHandlers {
		req, err = h(req)
		if err != nil {
			return nil, err
		}
	}

	resp, err := cfg.HTTPClient.Do(req)

	for _, h := range cfg.ResponseHandlers {
		resp, err = h(resp, err)
		if err != nil {
			return nil, err
		}
	}

	return resp, err
}
