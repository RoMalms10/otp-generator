package main

import (
	"context"
	"github.com/RoMalms10/otp-generator/config"
	"github.com/RoMalms10/otp-generator/messaging"
	"github.com/RoMalms10/otp-generator/server"
	"github.com/go-redis/redis/v8"
	"log"
	"net/http"
)

func main() {
	// Create a context
	ctx := context.Background()

	// Create Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: config.RedisHost,
	})

	// Test Redis connection
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Create Twilio service if credentials are provided
	var twilioService *messaging.TwilioService
	if config.TwilioAccountSID != "" && config.TwilioAuthToken != "" && config.TwilioPhoneNumber != "" {
		twilioConfig := messaging.TwilioConfig{
			AccountSID:  config.TwilioAccountSID,
			AuthToken:   config.TwilioAuthToken,
			PhoneNumber: config.TwilioPhoneNumber,
		}
		twilioService = messaging.NewTwilioService(twilioConfig)
		log.Println("Twilio service initialized")
	} else {
		log.Println("Warning: Twilio credentials not provided. SMS functionality will not work.")
	}

	// Create router with Redis and Twilio service
	router := server.NewRouter(redisClient, ctx, config.OTPTTL, twilioService)

	// Start the server
	log.Printf("Starting server on port %s", config.ServerPort)
	if err := http.ListenAndServe(":"+config.ServerPort, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
