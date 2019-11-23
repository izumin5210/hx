package cache

import (
	"bufio"
	"bytes"
	"context"
	"net/http"
	"net/http/httputil"
	"time"
)

type Transport struct {
	parent http.RoundTripper
	store  Store
}

var _ http.RoundTripper = (*Transport)(nil)

func NewTransport(
	parent http.RoundTripper,
	store Store,
) *Transport {
	return &Transport{
		parent: parent,
		store:  store,
	}
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()

	cacheCfg, cacheEnabled := getConfig(ctx)
	if cacheEnabled {
		resp, err := t.getCache(ctx, req, cacheCfg.Key)
		if err == nil && resp != nil {
			return resp, nil
		}
	}

	next := t.parent
	if next == nil {
		next = http.DefaultTransport
	}

	resp, err := next.RoundTrip(req)

	if resp != nil && cacheEnabled && cacheCfg.shouldCache(resp, err) {
		_ = t.putCache(ctx, resp, cacheCfg.Key, cacheCfg.TTL)
	}

	return resp, err
}

func (t *Transport) getCache(ctx context.Context, req *http.Request, key string) (*http.Response, error) {
	data, err := t.store.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	return http.ReadResponse(bufio.NewReader(bytes.NewBuffer(data)), req)
}

func (t *Transport) putCache(ctx context.Context, resp *http.Response, key string, ttl time.Duration) error {
	data, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return err
	}

	return t.store.Put(ctx, key, data, ttl)
}
