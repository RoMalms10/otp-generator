package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/go-redis/redis/v8"
)

type OTPService struct {
	RedisClient *redis.Client
	Ctx         context.Context
}

func NewOTPService(redisClient *redis.Client, ctx context.Context) *OTPService {
	return &OTPService{RedisClient: redisClient, Ctx: ctx}
}

func (s *OTPService) GenerateOTP(username string) (string, error) {
	otp := fmt.Sprintf("%06d", rand.Intn(1000000))
	redisKey := fmt.Sprintf("otp:%s", username)

	err := s.RedisClient.Set(s.Ctx, redisKey, otp, 10*time.Minute).Err()
	if err != nil {
		return "", err
	}

	return otp, nil
}

func (s *OTPService) ValidateOTP(username, otp string) (string, error) {
	redisKey := fmt.Sprintf("otp:%s", username)
	storedOTP, err := s.RedisClient.Get(s.Ctx, redisKey).Result()

	if err == redis.Nil {
		return "", fmt.Errorf("OTP expired or not found")
	} else if err != nil {
		return "", fmt.Errorf("Internal server error")
	}

	if storedOTP != otp {
		return "", fmt.Errorf("Invalid OTP")
	}

	s.RedisClient.Del(s.Ctx, redisKey)
	return "success", nil
}
