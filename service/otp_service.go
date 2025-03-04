package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/RoMalms10/otp-generator/messaging"
	"github.com/RoMalms10/otp-generator/models"
	"github.com/go-redis/redis/v8"
	"math/big"
	"time"
)

type OTPService struct {
	RedisClient   *redis.Client
	Context       context.Context
	OTPTTL        time.Duration // Configurable TTL
	TwilioService *messaging.TwilioService
}

func NewOTPService(redisClient *redis.Client, ctx context.Context, ttl time.Duration, twilioService *messaging.TwilioService) *OTPService {
	return &OTPService{
		RedisClient:   redisClient,
		Context:       ctx,
		OTPTTL:        ttl,
		TwilioService: twilioService,
	}
}

// GenerateOTP creates a new OTP for the specified username and stores it in Redis
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

// GetStoredOTP retrieves the current stored OTP for a user
// Returns empty string and redis.Nil error if not found
func (s *OTPService) GetStoredOTP(username string) (string, error) {
	otpKey := fmt.Sprintf("otp:%s", username)
	return s.RedisClient.Get(s.Context, otpKey).Result()
}

// SendOTP sends an OTP via the specified message type
func (s *OTPService) SendOTP(recipient, otp, messageType string) error {
	if s.TwilioService == nil {
		return fmt.Errorf("twilio service not configured")
	}

	switch messageType {
	case models.MessageTypeSMS:
		// Send OTP via SMS
		return s.TwilioService.SendOTP(recipient, otp)

	case models.MessageTypeWhatsApp:
		// Send OTP via WhatsApp
		return s.TwilioService.SendOTPWhatsApp(recipient, otp)

	case models.MessageTypeEmail:
		// Email sending would be implemented here
		return nil

	default:
		return fmt.Errorf("unsupported message type: %s", messageType)
	}
}

// ValidateOTP checks if the provided OTP matches the stored OTP for the user
func (s *OTPService) ValidateOTP(username, otp string) (string, error) {
	storedOTP, err := s.GetStoredOTP(username)
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
