package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/lcmetzger/rate_limiter/internal/middleware"
	"github.com/lcmetzger/rate_limiter/internal/repository"
	"github.com/lcmetzger/rate_limiter/internal/rl"
)

func main() {
	// Carregar variáveis de ambiente
	addr := os.Getenv("ADDR")
	ipRateLimit, _ := strconv.ParseInt(os.Getenv("IP_RATE_LIMIT"), 10, 64)
	tokenRateLimit, _ := strconv.ParseInt(os.Getenv("TOKEN_RATE_LIMIT"), 10, 64)
	blockDuration, _ := strconv.Atoi(os.Getenv("BLOCK_DURATION"))
	rateLimitType := os.Getenv("RATE_LIMIT_TYPE")
	repositoryType := os.Getenv("REPO_TYPE")

	var repo repository.IRateLimiterRespository = nil
	if repositoryType == "REDIS" {
		repo = repository.NewRedisRepository(addr)
	}
	if repositoryType == "PGSQL" {
		repo = repository.NewPgSqlRepository(addr)
	}
	if repo == nil {
		panic("Definir o tipo de repositorio a ser utilizado")
	}

	rateLimiter := rl.NewRateLimiter(repo, ipRateLimit, tokenRateLimit, blockDuration, rateLimitType)

	log.Printf("IpRateLimit: %v", ipRateLimit)
	log.Printf("TokenRateLimit: %v", tokenRateLimit)
	log.Printf("BlockDuration: %v", blockDuration)
	log.Printf("RateLimitType: %v", rateLimitType)
	log.Printf("RedisAddr: %v", addr)

	log.Println("Server is running...")

	mux := http.NewServeMux()
	mux.Handle("/", middleware.RateLimiterMiddleware(rateLimiter, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, world!"))
	})))

	http.ListenAndServe(":8080", mux)
}
