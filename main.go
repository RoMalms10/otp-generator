package main

import (
	"github.com/RoMalms10/otp-generator/config"
	"github.com/RoMalms10/otp-generator/server"
	"github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
	"log"
	"net/http"
)

func main() {
	// Application context
	ctx := context.Background()

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: config.RedisHost, // Use RedisHost from config package
		DB:   0,
	})
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Printf("Error closing Redis connection: %v", err)
		}
	}()

	// Test Redis connection
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis")

	// Initialize the router with the OTP TTL
	r := server.NewRouter(redisClient, ctx, config.OTPTTL)

	// Start the HTTP server
	log.Printf("OTP service running on port %s with OTP TTL: %v", config.ServerPort, config.OTPTTL)
	log.Fatal(http.ListenAndServe(":"+config.ServerPort, r))
}
