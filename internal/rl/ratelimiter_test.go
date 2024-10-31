package rl_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/lcmetzger/rate_limiter/internal/middleware"
	"github.com/lcmetzger/rate_limiter/internal/rl"
)

func TestMain(m *testing.M) {
	os.Setenv("REDIS", "localhost:6379")
	os.Setenv("IP_RATE_LIMIT", "2")
	os.Setenv("TOKEN_RATE_LIMIT", "3")
	os.Setenv("BLOCK_DURATION", "3")
	os.Exit(m.Run())
}

func TestRateLimiter(t *testing.T) {
	redisAddr := os.Getenv("REDIS")
	ipRateLimit, _ := strconv.Atoi(os.Getenv("IP_RATE_LIMIT"))
	tokenRateLimit, _ := strconv.Atoi(os.Getenv("TOKEN_RATE_LIMIT"))
	blockDuration, _ := strconv.Atoi(os.Getenv("BLOCK_DURATION"))
	rateLimitType := "IP"

	rateLimiter := rl.NewRateLimiter(redisAddr, ipRateLimit, tokenRateLimit, blockDuration, rateLimitType)
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
		t.Errorf("esperado status %v, porém foi recebido status %v", http.StatusOK, recorder.Code)
	}

	recorder = httptest.NewRecorder()
	middlewareHandler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Errorf("esperado status %v, porém foi recebido status %v", http.StatusOK, recorder.Code)
	}

	recorder = httptest.NewRecorder()
	middlewareHandler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusTooManyRequests {
		t.Errorf("esperado status %v, porém foi recebido status %v", http.StatusTooManyRequests, recorder.Code)
	}

	// Aguarda 4 segundos para rodar novo teste
	time.Sleep(time.Second * 4)
	recorder = httptest.NewRecorder()
	middlewareHandler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Errorf("esperado status %v, porém foi recebido status %v", http.StatusOK, recorder.Code)
	}

	// Testes para Token
	redisAddr = os.Getenv("REDIS")
	ipRateLimit, _ = strconv.Atoi(os.Getenv("IP_RATE_LIMIT"))
	tokenRateLimit, _ = strconv.Atoi(os.Getenv("TOKEN_RATE_LIMIT"))
	blockDuration, _ = strconv.Atoi(os.Getenv("BLOCK_DURATION"))
	rateLimitType = "TOKEN"

	rateLimiter = rl.NewRateLimiter(redisAddr, ipRateLimit, tokenRateLimit, blockDuration, rateLimitType)
	helloHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, world!"))
	})
	middlewareHandler = middleware.RateLimiterMiddleware(rateLimiter, helloHandler)

	req.Header.Set("API_KEY", "test-token")
	recorder = httptest.NewRecorder()
	middlewareHandler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Errorf("esperado status %v, porém foi recebido status %v", http.StatusOK, recorder.Code)
	}

	recorder = httptest.NewRecorder()
	middlewareHandler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Errorf("esperado status %v, porém foi recebido status %v", http.StatusOK, recorder.Code)
	}

	recorder = httptest.NewRecorder()
	middlewareHandler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Errorf("esperado status %v, porém foi recebido status %v", http.StatusOK, recorder.Code)
	}

	recorder = httptest.NewRecorder()
	middlewareHandler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusTooManyRequests {
		t.Errorf("esperado status %v, porém foi recebido status %v", http.StatusTooManyRequests, recorder.Code)
	}

	// Aguarda 4 segundos para rodar novo teste
	time.Sleep(time.Second * 4)
	recorder = httptest.NewRecorder()
	middlewareHandler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Errorf("esperado status %v, porém foi recebido status %v", http.StatusOK, recorder.Code)
	}

}
