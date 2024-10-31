package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/lcmetzger/rate_limiter/internal/middleware"
	"github.com/lcmetzger/rate_limiter/internal/rl"
)

func main() {
	// Carregar variáveis de ambiente
	redisAddr := os.Getenv("REDIS")
	ipRateLimit, _ := strconv.Atoi(os.Getenv("IP_RATE_LIMIT"))
	tokenRateLimit, _ := strconv.Atoi(os.Getenv("TOKEN_RATE_LIMIT"))
	blockDuration, _ := strconv.Atoi(os.Getenv("BLOCK_DURATION"))
	rateLimitType := os.Getenv("RATE_LIMIT_TYPE")

	rateLimiter := rl.NewRateLimiter(redisAddr, ipRateLimit, tokenRateLimit, blockDuration, rateLimitType)

	log.Printf("IpRateLimit: %v", ipRateLimit)
	log.Printf("TokenRateLimit: %v", tokenRateLimit)
	log.Printf("BlockDuration: %v", blockDuration)
	log.Printf("RateLimitType: %v", rateLimitType)
	log.Printf("RedisAddr: %v", redisAddr)

	log.Println("Server is running...")

	mux := http.NewServeMux()
	mux.Handle("/", middleware.RateLimiterMiddleware(rateLimiter, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, world!"))
	})))

	http.ListenAndServe(":8080", mux)
}
