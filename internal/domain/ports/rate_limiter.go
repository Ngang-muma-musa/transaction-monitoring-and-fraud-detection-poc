package ports

import "context"

type RateLimiterPort interface {
	Allow(ctx context.Context, key string) (bool, error)
}
