package rl

import (
	"context"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	RedisClient    *redis.Client
	IpRateLimit    int
	TokenRateLimit int
	BlockDuration  int
	RateLimitType  string
}

func NewRateLimiter(redisAddr string, ipRateLimit, tokenRateLimit int, blockDuration int, rateLimitType string) *RateLimiter {
	if strings.ToUpper(rateLimitType) != "IP" && strings.ToUpper(rateLimitType) != "TOKEN" {
		panic("Tipo de rate limiter deve ser definido")
	}

	return &RateLimiter{
		RedisClient: redis.NewClient(&redis.Options{
			Addr: redisAddr,
		}),
		IpRateLimit:    ipRateLimit,
		TokenRateLimit: tokenRateLimit,
		BlockDuration:  blockDuration,
		RateLimitType:  rateLimitType,
	}
}

func (rl *RateLimiter) Allow(key string, limit int) bool {
	ctx := context.Background()

	// Incrementa o contador de requisições
	count, err := rl.RedisClient.Incr(ctx, key).Result()
	if err != nil {
		return false
	}

	// Define o tempo de expiração do contador
	if count == 1 {
		rl.RedisClient.Expire(ctx, key, time.Duration(rl.BlockDuration)*time.Second)
	}

	// Verifica se o limite foi excedido
	if int(count) > limit {
		rl.RedisClient.Set(ctx, key+":blocked", "true", time.Duration(rl.BlockDuration)*time.Second)
		return false
	}
	return true
}

func (rl *RateLimiter) IsBlocked(key string) bool {
	ctx := context.Background()
	blocked, err := rl.RedisClient.Get(ctx, key+":blocked").Result()
	if err == redis.Nil {
		return false
	}
	return blocked == "true"

}
