package server

import (
	"github.com/RoMalms10/otp-generator/handler"
	"github.com/RoMalms10/otp-generator/service"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"
	"time"
)

func NewRouter(redisClient *redis.Client, ctx context.Context, otpTTL time.Duration) *mux.Router {
	otpService := service.NewOTPService(redisClient, ctx, otpTTL)
	otpHandler := handler.NewHandler(otpService)

	r := mux.NewRouter()
	r.HandleFunc("/otp/generate", otpHandler.GenerateOTPHandler).Methods("POST")
	r.HandleFunc("/otp/validate", otpHandler.ValidateOTPHandler).Methods("POST")

	return r
}
