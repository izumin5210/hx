package cache

import (
	"context"
	"net/http"
	"time"

	"github.com/izumin5210/hx"
)

func With(key string, opts ...Option) hx.Option {
	return hx.HandleRequest(func(req *http.Request) (*http.Request, error) {
		cfg := &Config{Key: key}
		for _, f := range opts {
			f(cfg)
		}

		return req.WithContext(SetConfig(req.Context(), cfg)), nil
	})
}

type Option func(*Config)

type Config struct {
	Key  string
	TTL  time.Duration
	Cond hx.ResponseHandlerCond
}

func (c *Config) shouldCache(resp *http.Response, err error) bool {
	if c == nil {
		return false
	}

	cond := c.Cond
	if cond == nil {
		cond = hx.IsSuccess
	}
	return cond(resp, err)
}

type ctxConfig struct{}

func getConfig(ctx context.Context) (*Config, bool) {
	v, ok := ctx.Value(ctxConfig{}).(*Config)
	return v, ok
}

func SetConfig(ctx context.Context, cfg *Config) context.Context {
	return context.WithValue(ctx, ctxConfig{}, cfg)
}

func TTL(ttl time.Duration) Option {
	return func(c *Config) { c.TTL = ttl }
}

func When(cond hx.ResponseHandlerCond) Option {
	return func(c *Config) { c.Cond = cond }
}
