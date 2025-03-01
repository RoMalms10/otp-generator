package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RoMalms10/otp-generator/models"
	"github.com/RoMalms10/otp-generator/server"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"
)

func setupTestServer() *mux.Router {
	ctx := context.Background()
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1,
	})
	return server.NewRouter(redisClient, ctx)
}

func TestGenerateOTPEndpoint(t *testing.T) {
	r := setupTestServer()
	reqBody, _ := json.Marshal(models.GenerateRequest{Username: "testuser", MessageType: "email"})
	req, _ := http.NewRequest("POST", "/otp/generate", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.Code)
	}
}

func TestValidateOTPEndpoint(t *testing.T) {
	r := setupTestServer()

	// Generate OTP first
	generateBody, _ := json.Marshal(models.GenerateRequest{Username: "testuser", MessageType: "email"})
	generateReq, _ := http.NewRequest("POST", "/otp/generate", bytes.NewBuffer(generateBody))
	generateReq.Header.Set("Content-Type", "application/json")
	generateResp := httptest.NewRecorder()
	r.ServeHTTP(generateResp, generateReq)

	var generateRespBody map[string]string
	json.Unmarshal(generateResp.Body.Bytes(), &generateRespBody)
	otp := generateRespBody["otp"]

	// Validate OTP
	validateBody, _ := json.Marshal(models.ValidationRequest{Username: "testuser", OTP: otp})
	validateReq, _ := http.NewRequest("POST", "/otp/validate", bytes.NewBuffer(validateBody))
	validateReq.Header.Set("Content-Type", "application/json")
	validateResp := httptest.NewRecorder()
	r.ServeHTTP(validateResp, validateReq)

	if validateResp.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", validateResp.Code)
	}
}
