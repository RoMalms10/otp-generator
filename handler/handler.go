package handler

import (
	"encoding/json"
	"github.com/RoMalms10/otp-generator/models"
	"github.com/RoMalms10/otp-generator/service"
	"net/http"
)

type Handler struct {
	OTPService *service.OTPService
}

func NewHandler(otpService *service.OTPService) *Handler {
	return &Handler{OTPService: otpService}
}

func (h *Handler) GenerateOTPHandler(w http.ResponseWriter, r *http.Request) {
	var req models.GenerateRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	// Validate supported MessageType
	if req.MessageType != "email" && req.MessageType != "sms" {
		http.Error(w, "Unsupported MessageType", http.StatusBadRequest)
		return
	}

	// Pass TTL setting to the service
	otp, err := h.OTPService.GenerateOTP(req.Username)
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

	// Validate OTP using the service
	success, err := h.OTPService.ValidateOTP(req.Username, req.OTP)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": success})
}
