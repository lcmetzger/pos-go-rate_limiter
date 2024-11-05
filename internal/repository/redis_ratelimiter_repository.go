package repository

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type RedidRepository struct {
	RedisClient *redis.Client
}

func NewRedisRepository(addr string) *RedidRepository {
	return &RedidRepository{
		RedisClient: redis.NewClient(&redis.Options{
			Addr: addr,
		}),
	}
}

func (repo *RedidRepository) Save(ctx context.Context, key, value string) {
	err := repo.RedisClient.Set(ctx, key, value, 0).Err()
	if err != nil {
		panic(err)
	}
}

func (repo *RedidRepository) Update(ctx context.Context, key, value string) {
	err := repo.RedisClient.Set(ctx, key, value, 0).Err()
	if err != nil {
		panic(err)
	}

}

func (repo *RedidRepository) Find(ctx context.Context, key string) (string, error) {
	res, err := repo.RedisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return res, nil
}

func (repo *RedidRepository) Delete(ctx context.Context, key string) bool {
	err := repo.RedisClient.Del(ctx, key).Err()
	return err == nil
}
