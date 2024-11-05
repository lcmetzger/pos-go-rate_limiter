package repository

import "context"

type IRateLimiterRespository interface {
	Save(ctx context.Context, key, value string)
	Update(ctx context.Context, key, value string)
	Find(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) bool
}
