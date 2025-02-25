package server

import (
	"github.com/RoMalms10/otp-generator/handler"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"
)

func NewRouter(redisClient *redis.Client, ctx context.Context) *mux.Router {
	r := mux.NewRouter()
	h := handler.NewHandler(redisClient, ctx)

	r.HandleFunc("/otp/generate", h.GenerateOTPHandler).Methods("POST")
	r.HandleFunc("/otp/validate", h.ValidateOTPHandler).Methods("POST")
	return r
}
