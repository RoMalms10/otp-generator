package config

import "time"

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
)
