package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/RoMalms10/otp-generator/models"
	"github.com/RoMalms10/otp-generator/server"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func setupRouterWithTestServer(otpTTL time.Duration) (*mux.Router, *redis.Client, context.Context) {
	ctx := context.Background()
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1,
	})
	return server.NewRouter(redisClient, ctx, otpTTL), redisClient, ctx
}

func TestExpiredOTP(t *testing.T) {
	otpTTL := 2 * time.Second // Custom TTL for faster testing
	router, redisClient, ctx := setupRouterWithTestServer(otpTTL)
	defer redisClient.FlushDB(ctx)

	// Generate OTP
	body, _ := json.Marshal(models.GenerateRequest{Username: "testuser", MessageType: "email"})
	req, _ := http.NewRequest("POST", "/otp/generate", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// Wait for TTL expiration
	time.Sleep(otpTTL + 1*time.Second)

	// Validate OTP after expiration
	body, _ = json.Marshal(models.ValidationRequest{Username: "testuser", OTP: "123456"})
	req, _ = http.NewRequest("POST", "/otp/validate", bytes.NewBuffer(body))
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
