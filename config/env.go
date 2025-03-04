package config

import (
	"os"
	"time"
)

// Application settings (constants & variables for configuration)
const (
	// Redis configuration
	RedisHost = "localhost:6379"
	// HTTP server configuration
	ServerPort = "8080"
	// Default OTP TTL (e.g., 10 minutes)
	DefaultOTPTTL = 10 * time.Minute
)

var (
	// OTP TTL (can remain as a variable if you anticipate changes at runtime)
	OTPTTL = DefaultOTPTTL

	// Twilio configuration (loaded from environment variables)
	TwilioAccountSID  = getEnvOrDefault("TWILIO_ACCOUNT_SID", "")
	TwilioAuthToken   = getEnvOrDefault("TWILIO_AUTH_TOKEN", "")
	TwilioPhoneNumber = getEnvOrDefault("TWILIO_PHONE_NUMBER", "")
)

// getEnvOrDefault gets an environment variable or returns a default value if not set
func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
