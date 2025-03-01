package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/RoMalms10/otp-generator/models"
	"github.com/RoMalms10/otp-generator/server"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

func setupRouterWithTestServer(otpTTL time.Duration) (*mux.Router, *redis.Client, context.Context) {
	ctx := context.Background()
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1,
	})
	router := server.NewRouter(redisClient, ctx, otpTTL)
	return router, redisClient, ctx
}

func TestOTPService(t *testing.T) {
	otpTTL := 1 * time.Second // Default OTP TTL for tests
	router, redisClient, ctx := setupRouterWithTestServer(otpTTL)
	defer redisClient.FlushDB(ctx)

	t.Run("Generate OTP", func(t *testing.T) {
		t.Run("Valid Request", func(t *testing.T) {
			body, _ := json.Marshal(models.GenerateRequest{Username: "testuser", MessageType: "email"})
			req, _ := http.NewRequest("POST", "/otp/generate", bytes.NewBuffer(body))
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)

			var resp map[string]string
			_ = json.Unmarshal(rec.Body.Bytes(), &resp)

			otp, exists := resp["otp"]
			assert.True(t, exists, "Expected 'otp' key in response")
			assert.Len(t, otp, 6, "Expected OTP to be 6 digits")
		})

		t.Run("Missing Username", func(t *testing.T) {
			body, _ := json.Marshal(models.GenerateRequest{Username: "", MessageType: "email"})
			req, _ := http.NewRequest("POST", "/otp/generate", bytes.NewBuffer(body))
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusBadRequest, rec.Code, "Expected status 400 for missing username")
		})

		t.Run("Invalid MessageType", func(t *testing.T) {
			body, _ := json.Marshal(models.GenerateRequest{Username: "testuser", MessageType: "unsupported"})
			req, _ := http.NewRequest("POST", "/otp/generate", bytes.NewBuffer(body))
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusBadRequest, rec.Code, "Expected status 400 for unsupported message type")
		})
	})

	t.Run("Validate OTP", func(t *testing.T) {
		// Generate an OTP for validation tests
		body, _ := json.Marshal(models.GenerateRequest{Username: "testuser", MessageType: "email"})
		req, _ := http.NewRequest("POST", "/otp/generate", bytes.NewBuffer(body))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		var resp map[string]string
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
		otp := resp["otp"]

		t.Run("Valid OTP", func(t *testing.T) {
			validateBody, _ := json.Marshal(models.ValidationRequest{Username: "testuser", OTP: otp})
			req, _ := http.NewRequest("POST", "/otp/validate", bytes.NewBuffer(validateBody))
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code, "Expected status 200 for valid OTP")
		})

		t.Run("Invalid OTP", func(t *testing.T) {
			validateBody, _ := json.Marshal(models.ValidationRequest{Username: "testuser", OTP: "000000"})
			req, _ := http.NewRequest("POST", "/otp/validate", bytes.NewBuffer(validateBody))
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusUnauthorized, rec.Code, "Expected status 401 for invalid OTP")
		})

		t.Run("Expired OTP", func(t *testing.T) {
			time.Sleep(otpTTL + 1*time.Second) // Wait for OTP to expire

			validateBody, _ := json.Marshal(models.ValidationRequest{Username: "testuser", OTP: otp})
			req, _ := http.NewRequest("POST", "/otp/validate", bytes.NewBuffer(validateBody))
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusUnauthorized, rec.Code, "Expected status 401 for expired OTP")
		})
	})

	t.Run("Re-generate OTP Invalidates Previous", func(t *testing.T) {
		body, _ := json.Marshal(models.GenerateRequest{Username: "testuser", MessageType: "email"})

		// Generate the first OTP
		req, _ := http.NewRequest("POST", "/otp/generate", bytes.NewBuffer(body))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		var firstResp map[string]string
		_ = json.Unmarshal(rec.Body.Bytes(), &firstResp)
		firstOTP := firstResp["otp"]

		// Generate a new OTP for the same user
		req, _ = http.NewRequest("POST", "/otp/generate", bytes.NewBuffer(body))
		rec = httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		var secondResp map[string]string
		_ = json.Unmarshal(rec.Body.Bytes(), &secondResp)
		secondOTP := secondResp["otp"]

		assert.NotEqual(t, firstOTP, secondOTP, "New OTP should replace the previous one")

		// Validate first (old) OTP (should fail)
		validateBody, _ := json.Marshal(models.ValidationRequest{Username: "testuser", OTP: firstOTP})
		req, _ = http.NewRequest("POST", "/otp/validate", bytes.NewBuffer(validateBody))
		rec = httptest.NewRecorder()

		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "Old OTP should no longer be valid")

		// Validate second (new) OTP (should succeed)
		validateBody, _ = json.Marshal(models.ValidationRequest{Username: "testuser", OTP: secondOTP})
		req, _ = http.NewRequest("POST", "/otp/validate", bytes.NewBuffer(validateBody))
		rec = httptest.NewRecorder()

		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code, "New OTP should be valid")
	})
}
