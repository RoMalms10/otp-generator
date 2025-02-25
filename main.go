package main

import (
	"github.com/RoMalms10/otp-generator/server"
	"log"
	"net/http"

	"github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
)

var ctx = context.Background()

func main() {
	ctx := context.Background()
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

	r := server.NewRouter(redisClient, ctx)
	log.Println("OTP service running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
