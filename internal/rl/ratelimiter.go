package rl

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/lcmetzger/rate_limiter/internal/repository"
)

type RateLimiter struct {
	repo           repository.IRateLimiterRespository
	IpRateLimit    int64
	TokenRateLimit int64
	BlockDuration  int
	RateLimitType  string
}

func NewRateLimiter(repository repository.IRateLimiterRespository, ipRateLimit, tokenRateLimit int64, blockDuration int, rateLimitType string) *RateLimiter {
	if strings.ToUpper(rateLimitType) != "IP" && strings.ToUpper(rateLimitType) != "TOKEN" {
		panic("O tipo de rate limiter deve ser definido através da variável de ambiente RATE_LIMIT_TYPE")
	}

	return &RateLimiter{
		repo:           repository,
		IpRateLimit:    ipRateLimit,
		TokenRateLimit: tokenRateLimit,
		BlockDuration:  blockDuration,
		RateLimitType:  rateLimitType,
	}
}

func (rl *RateLimiter) Allow(key string, limit int64) bool {
	ctx := context.Background()
	var count int64 = 1

	res, err := rl.repo.Find(ctx, key)
	if err != nil {
		return false
	}
	if res == "" {
		rl.repo.Save(ctx, key, strconv.FormatInt(count, 10))
	} else {
		count, err = strconv.ParseInt(res, 10, 64)
		if err != nil {
			panic("erro de conversão")
		}
		count++
		rl.repo.Update(ctx, key, strconv.FormatInt(count, 10))
	}

	if count == 1 {
		go func(k string) {
			time.Sleep(time.Duration(rl.BlockDuration) * time.Second)
			rl.repo.Delete(ctx, k)
		}(key)
	}

	if count > limit {
		res, err := rl.repo.Find(ctx, key+":blocked")
		if err != nil {
			log.Println(err)
		}
		if res != "true" {
			rl.repo.Save(ctx, key+":blocked", "true")
			go func(k string) {
				time.Sleep(time.Duration(rl.BlockDuration) * time.Second)
				rl.repo.Delete(ctx, k)
			}(key + ":blocked")
		}
		return false
	}
	return true
}

func (rl *RateLimiter) IsBlocked(key string) bool {
	ctx := context.Background()
	blocked, err := rl.repo.Find(ctx, key+":blocked")
	if err == nil {
		return false
	}
	return blocked == "true"
}
