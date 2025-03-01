package handler

import (
	"encoding/json"
	"net/http"

	"github.com/RoMalms10/otp-generator/models"
	"github.com/RoMalms10/otp-generator/service"
)

type Handler struct {
	OTPService *service.OTPService
}

func NewHandler(otpService *service.OTPService) *Handler {
	return &Handler{OTPService: otpService}
}

func (h *Handler) GenerateOTPHandler(w http.ResponseWriter, r *http.Request) {
	var req models.GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

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

	success, err := h.OTPService.ValidateOTP(req.Username, req.OTP)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": success})
}
