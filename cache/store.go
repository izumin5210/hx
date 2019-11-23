package cache

import (
	"context"
	"time"
)

type Store interface {
	Put(ctx context.Context, key string, data []byte, ttl time.Duration) error
	Get(ctx context.Context, key string) ([]byte, error)
}
