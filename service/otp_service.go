package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/go-redis/redis/v8"
	"math/big"
	"time"
)

type OTPService struct {
	RedisClient *redis.Client
	Context     context.Context
	OTPTTL      time.Duration // Configurable TTL
}

func NewOTPService(redisClient *redis.Client, ctx context.Context, ttl time.Duration) *OTPService {
	return &OTPService{
		RedisClient: redisClient,
		Context:     ctx,
		OTPTTL:      ttl,
	}
}

func (s *OTPService) GenerateOTP(username string) (string, error) {
	// Set the maximum value for a 6-digit number (999999)
	const max = 1000000

	// Generate a random number in the range [0, 999999]
	n, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return "", err
	}

	otp := fmt.Sprintf("%06d", n.Int64())
	otpKey := fmt.Sprintf("otp:%s", username)

	err = s.RedisClient.Set(s.Context, otpKey, otp, s.OTPTTL).Err()
	if err != nil {
		return "", err
	}

	return otp, nil
}

func (s *OTPService) ValidateOTP(username, otp string) (string, error) {
	otpKey := fmt.Sprintf("otp:%s", username)
	storedOTP, err := s.RedisClient.Get(s.Context, otpKey).Result()
	if err == redis.Nil {
		return "invalid", fmt.Errorf("OTP has expired or does not exist")
	} else if err != nil {
		return "invalid", fmt.Errorf("Server error: %v", err)
	}

	if storedOTP != otp {
		return "invalid", fmt.Errorf("Incorrect OTP entered")
	}

	return "valid", nil
}
