package tests

import (
	"context"
	"testing"

	"github.com/RoMalms10/otp-generator/service"
	"github.com/go-redis/redis/v8"
)

func setupTestRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1,
	})
}

func TestGenerateOTP(t *testing.T) {
	ctx := context.Background()
	redisClient := setupTestRedis()
	defer redisClient.FlushDB(ctx)

	service := service.NewOTPService(redisClient, ctx)
	otp, err := service.GenerateOTP("testuser")

	if err != nil {
		t.Fatalf("Failed to generate OTP: %v", err)
	}

	if len(otp) != 6 {
		t.Errorf("Expected OTP length 6, got %d", len(otp))
	}
}

func TestValidateOTP(t *testing.T) {
	ctx := context.Background()
	redisClient := setupTestRedis()
	defer redisClient.FlushDB(ctx)

	service := service.NewOTPService(redisClient, ctx)
	otp, _ := service.GenerateOTP("testuser")

	status, err := service.ValidateOTP("testuser", otp)
	if err != nil || status != "success" {
		t.Errorf("Expected success, got error: %v", err)
	}
}
