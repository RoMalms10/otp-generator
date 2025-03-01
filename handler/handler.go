package handler

import (
	"encoding/json"
	"github.com/RoMalms10/otp-generator/models"
	"github.com/RoMalms10/otp-generator/service"
	"net/http"

	"github.com/go-chi/render"
)

type Handler struct {
	OTPService *service.OTPService
}

type ErrResponse struct {
	HTTPStatusCode int    `json:"-"`
	StatusText     string `json:"status"`
	ErrorText      string `json:"error,omitempty"`
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func NewErrResponse(statusCode int, statusText, errorText string) *ErrResponse {
	return &ErrResponse{
		HTTPStatusCode: statusCode,
		StatusText:     statusText,
		ErrorText:      errorText,
	}
}

func NewHandler(otpService *service.OTPService) *Handler {
	return &Handler{OTPService: otpService}
}

func (h *Handler) GenerateOTPHandler(w http.ResponseWriter, r *http.Request) {
	var req models.GenerateRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Username == "" {
		render.Render(w, r, NewErrResponse(http.StatusBadRequest, "Bad Request", "Username is required"))
		return
	}

	if req.MessageType != "email" && req.MessageType != "sms" {
		render.Render(w, r, NewErrResponse(http.StatusBadRequest, "Bad Request", "Unsupported MessageType"))
		return
	}

	otp, err := h.OTPService.GenerateOTP(req.Username)
	if err != nil {
		render.Render(w, r, NewErrResponse(http.StatusInternalServerError, "Internal Server Error", err.Error()))
		return
	}

	render.JSON(w, r, map[string]string{"otp": otp})
}

func (h *Handler) ValidateOTPHandler(w http.ResponseWriter, r *http.Request) {
	var req models.ValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Render(w, r, NewErrResponse(http.StatusBadRequest, "Bad Request", "Invalid request payload"))
		return
	}

	// Validate OTP using the service
	success, err := h.OTPService.ValidateOTP(req.Username, req.OTP)
	if err != nil {
		render.Render(w, r, NewErrResponse(http.StatusUnauthorized, "Unauthorized", err.Error()))
		return
	}

	render.JSON(w, r, map[string]string{"status": success})
}
