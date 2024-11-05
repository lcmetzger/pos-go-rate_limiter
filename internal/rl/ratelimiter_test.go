package rl_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/lcmetzger/rate_limiter/internal/middleware"
	"github.com/lcmetzger/rate_limiter/internal/repository"
	"github.com/lcmetzger/rate_limiter/internal/rl"
)

func TestMain(m *testing.M) {
	os.Setenv("ADDR", "localhost:6379")
	os.Setenv("IP_RATE_LIMIT", "2")
	os.Setenv("TOKEN_RATE_LIMIT", "2")
	os.Setenv("BLOCK_DURATION", "3")
	os.Setenv("REPO_TYPE", "REDIS")
	os.Exit(m.Run())
}

func TestRedisRateLimiter(t *testing.T) {
	redisAddr := os.Getenv("ADDR")
	ipRateLimit, _ := strconv.ParseInt(os.Getenv("IP_RATE_LIMIT"), 10, 64)
	tokenRateLimit, _ := strconv.ParseInt(os.Getenv("TOKEN_RATE_LIMIT"), 10, 64)
	blockDuration, _ := strconv.Atoi(os.Getenv("BLOCK_DURATION"))
	rateLimitType := "IP"
	repository_type := os.Getenv("REPO_TYPE")

	var repo repository.IRateLimiterRespository = nil
	if repository_type == "REDIS" {
		repo = repository.NewRedisRepository(redisAddr)
	}
	if repository_type == "PGSQL" {
		repo = repository.NewRedisRepository(redisAddr)
	}
	if repo == nil {
		panic("Definir o tipo de repositorio a ser utilizado")
	}

	rateLimiter := rl.NewRateLimiter(repo, ipRateLimit, tokenRateLimit, blockDuration, rateLimitType)
	helloHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, world!"))
	})
	middlewareHandler := middleware.RateLimiterMiddleware(rateLimiter, helloHandler)

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.178.0.1"
	recorder := httptest.NewRecorder()

	// Testes para IP
	middlewareHandler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Errorf("IP: esperado status %v, porém foi recebido status %v", http.StatusOK, recorder.Code)
	}

	recorder = httptest.NewRecorder()
	middlewareHandler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Errorf("IP: esperado status %v, porém foi recebido status %v", http.StatusOK, recorder.Code)
	}

	recorder = httptest.NewRecorder()
	middlewareHandler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusTooManyRequests {
		t.Errorf("IP: esperado status %v, porém foi recebido status %v", http.StatusTooManyRequests, recorder.Code)
	}

	// Aguarda 4 segundos para rodar novo teste
	time.Sleep(time.Second * 4)
	recorder = httptest.NewRecorder()
	middlewareHandler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Errorf("IP: esperado status %v, porém foi recebido status %v", http.StatusOK, recorder.Code)
	}

	// Testes para Token
	redisAddr = os.Getenv("ADDR")
	ipRateLimit, _ = strconv.ParseInt(os.Getenv("IP_RATE_LIMIT"), 10, 64)
	tokenRateLimit, _ = strconv.ParseInt(os.Getenv("TOKEN_RATE_LIMIT"), 10, 64)
	blockDuration, _ = strconv.Atoi(os.Getenv("BLOCK_DURATION"))
	rateLimitType = "TOKEN"
	repository_type = os.Getenv("REPO_TYPE")

	repo = nil
	if repository_type == "REDIS" {
		repo = repository.NewRedisRepository(redisAddr)
	}
	if repository_type == "PGSQL" {
		repo = repository.NewPgSqlRepository(redisAddr)
	}
	if repo == nil {
		panic("TOKEN: Definir o tipo de repositorio a ser utilizado")
	}

	rateLimiter = rl.NewRateLimiter(repo, ipRateLimit, tokenRateLimit, blockDuration, rateLimitType)
	helloHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, world!"))
	})
	middlewareHandler = middleware.RateLimiterMiddleware(rateLimiter, helloHandler)

	req.Header.Set("API_KEY", "test-token")
	recorder = httptest.NewRecorder()
	middlewareHandler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Errorf("TOKEN: esperado status %v, porém foi recebido status %v", http.StatusOK, recorder.Code)
	}

	recorder = httptest.NewRecorder()
	middlewareHandler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Errorf("TOKEN: esperado status %v, porém foi recebido status %v", http.StatusOK, recorder.Code)
	}

	recorder = httptest.NewRecorder()
	middlewareHandler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusTooManyRequests {
		t.Errorf("TOKEN: esperado status %v, porém foi recebido status %v", http.StatusTooManyRequests, recorder.Code)
	}

}
