package handler

import (
	"encoding/json"
	"fmt"
	"github.com/RoMalms10/otp-generator/models"
	"github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
	"math/rand"
	"net/http"
	"time"
)

type Handler struct {
	RedisClient *redis.Client
	Ctx         context.Context
}

func NewHandler(redisClient *redis.Client, ctx context.Context) *Handler {
	return &Handler{
		RedisClient: redisClient,
		Ctx:         ctx}
}

func (h *Handler) GenerateOTPHandler(w http.ResponseWriter, r *http.Request) {
	var req models.GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	otp := fmt.Sprintf("%06d", rand.Intn(1000000))
	redisKey := fmt.Sprintf("otp:%s", req.Username)

	err := h.RedisClient.Set(h.Ctx, redisKey, otp, 10*time.Minute).Err()
	if err != nil {
		http.Error(w, "Failed to store OTP", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"otp": otp})
}

func (h *Handler) ValidateOTPHandler(w http.ResponseWriter, r *http.Request) {
	var req models.ValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	redisKey := fmt.Sprintf("otp:%s", req.Username)
	storedOTP, err := h.RedisClient.Get(h.Ctx, redisKey).Result()

	if err == redis.Nil {
		http.Error(w, "OTP expired or not found", http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if storedOTP != req.OTP {
		http.Error(w, "Invalid OTP", http.StatusUnauthorized)
		return
	}

	h.RedisClient.Del(h.Ctx, redisKey)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
