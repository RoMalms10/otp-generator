package tests

import (
	"bytes"
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
	"golang.org/x/net/context"
)

var defaultOTPTTL = time.Minute // Default OTP expiration time

// Setup reusable test server with configurable OTP TTL
func setupTestServerWithTTL(ttl time.Duration) (*mux.Router, *redis.Client, context.Context) {
	ctx := context.Background()
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1,
	})
	router := server.NewRouter(redisClient, ctx, ttl) // Pass the TTL to the server
	return router, redisClient, ctx
}

func TestGenerateOTPEndpoint(t *testing.T) {
	// Set a short OTP TTL for testing purposes
	testOTPTTL := 2 * time.Second
	r, redisClient, ctx := setupTestServerWithTTL(testOTPTTL)
	defer redisClient.FlushDB(ctx) // Clean up after test

	t.Run("Valid Request", func(t *testing.T) {
		reqBody, _ := json.Marshal(models.GenerateRequest{Username: "testuser", MessageType: "email"})
		req, err := http.NewRequest("POST", "/otp/generate", bytes.NewBuffer(reqBody))
		assert.NoError(t, err, "Error creating request")
		req.Header.Set("Content-Type", "application/json")

		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code, "Expected status 200")

		var generateRespBody map[string]string
		err = json.Unmarshal(resp.Body.Bytes(), &generateRespBody)
		assert.NoError(t, err, "Failed to parse response JSON")

		otp, exists := generateRespBody["otp"]
		assert.True(t, exists, "Expected 'otp' key in response")
		assert.Len(t, otp, 6, "Expected OTP to be 6 digits")
	})

	t.Run("Missing Username", func(t *testing.T) {
		reqBody, _ := json.Marshal(models.GenerateRequest{Username: "", MessageType: "email"})
		req, err := http.NewRequest("POST", "/otp/generate", bytes.NewBuffer(reqBody))
		assert.NoError(t, err, "Error creating request")
		req.Header.Set("Content-Type", "application/json")

		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code, "Expected status 400 for missing username")
	})

	t.Run("Unsupported MessageType", func(t *testing.T) {
		reqBody, _ := json.Marshal(models.GenerateRequest{Username: "testuser", MessageType: "unsupported"})
		req, err := http.NewRequest("POST", "/otp/generate", bytes.NewBuffer(reqBody))
		assert.NoError(t, err, "Error creating request")
		req.Header.Set("Content-Type", "application/json")

		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code, "Expected status 400 for unsupported message type")
	})
}

func TestValidateOTPEndpoint(t *testing.T) {
	// Set a short OTP TTL for testing purposes
	testOTPTTL := 2 * time.Second
	r, redisClient, ctx := setupTestServerWithTTL(testOTPTTL)
	defer redisClient.FlushDB(ctx) // Clean up after test

	// Generate valid OTP first
	generateBody, _ := json.Marshal(models.GenerateRequest{Username: "testuser", MessageType: "email"})
	generateReq, _ := http.NewRequest("POST", "/otp/generate", bytes.NewBuffer(generateBody))
	generateReq.Header.Set("Content-Type", "application/json")
	generateResp := httptest.NewRecorder()
	r.ServeHTTP(generateResp, generateReq)

	assert.Equal(t, http.StatusOK, generateResp.Code, "Expected status 200")

	var generateRespBody map[string]string
	err := json.Unmarshal(generateResp.Body.Bytes(), &generateRespBody)
	assert.NoError(t, err, "Failed to parse response JSON")

	otp := generateRespBody["otp"]

	t.Run("Valid OTP", func(t *testing.T) {
		validateBody, _ := json.Marshal(models.ValidationRequest{Username: "testuser", OTP: otp})
		validateReq, _ := http.NewRequest("POST", "/otp/validate", bytes.NewBuffer(validateBody))
		validateReq.Header.Set("Content-Type", "application/json")
		validateResp := httptest.NewRecorder()
		r.ServeHTTP(validateResp, validateReq)

		assert.Equal(t, http.StatusOK, validateResp.Code, "Expected status 200")
	})

	t.Run("Invalid OTP", func(t *testing.T) {
		validateBody, _ := json.Marshal(models.ValidationRequest{Username: "testuser", OTP: "999999"})
		validateReq, _ := http.NewRequest("POST", "/otp/validate", bytes.NewBuffer(validateBody))
		validateReq.Header.Set("Content-Type", "application/json")
		validateResp := httptest.NewRecorder()
		r.ServeHTTP(validateResp, validateReq)

		assert.Equal(t, http.StatusUnauthorized, validateResp.Code, "Expected status 401")
	})

	t.Run("Expired OTP", func(t *testing.T) {
		// Wait for OTP to expire
		time.Sleep(3 * time.Second)

		validateBody, _ := json.Marshal(models.ValidationRequest{Username: "testuser", OTP: otp})
		validateReq, _ := http.NewRequest("POST", "/otp/validate", bytes.NewBuffer(validateBody))
		validateReq.Header.Set("Content-Type", "application/json")
		validateResp := httptest.NewRecorder()
		r.ServeHTTP(validateResp, validateReq)

		assert.Equal(t, http.StatusUnauthorized, validateResp.Code, "Expected status 401 for expired OTP")
	})

}

func TestConcurrencyGenerateOTP(t *testing.T) {
	// Set a short OTP TTL for testing purposes
	testOTPTTL := 2 * time.Second
	r, redisClient, ctx := setupTestServerWithTTL(testOTPTTL)
	defer redisClient.FlushDB(ctx) // Clean up after test

	t.Run("Concurrent Requests", func(t *testing.T) {
		const parallelCount = 5
		errorChannel := make(chan error, parallelCount)

		for i := 0; i < parallelCount; i++ {
			go func(i int) {
				reqBody, _ := json.Marshal(models.GenerateRequest{Username: "testuser", MessageType: "email"})
				req, err := http.NewRequest("POST", "/otp/generate", bytes.NewBuffer(reqBody))
				if err != nil {
					errorChannel <- err
					return
				}
				req.Header.Set("Content-Type", "application/json")

				resp := httptest.NewRecorder()
				r.ServeHTTP(resp, req)

				if resp.Code != http.StatusOK {
					errorChannel <- err
					return
				}
				errorChannel <- nil
			}(i)
		}

		// Collect and assert results
		for i := 0; i < parallelCount; i++ {
			err := <-errorChannel
			assert.NoError(t, err, "Concurrent OTP generation failed")
		}
	})
}
