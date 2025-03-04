package handler

import (
	"encoding/json"
	"github.com/RoMalms10/otp-generator/models"
	"github.com/RoMalms10/otp-generator/service"
	"github.com/go-chi/render"
	"net/http"
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

	// Validate message type
	if !isValidMessageType(req.MessageType) {
		render.Render(w, r, NewErrResponse(http.StatusBadRequest, "Bad Request",
			"Unsupported MessageType. Use 'sms', 'whatsapp', or 'email'"))
		return
	}

	// First, generate the OTP
	otp, err := h.OTPService.GenerateOTP(req.Username)
	if err != nil {
		render.Render(w, r, NewErrResponse(http.StatusInternalServerError, "Internal Server Error",
			"Failed to generate OTP: "+err.Error()))
		return
	}

	// Then, send the OTP
	err = h.OTPService.SendOTP(req.Username, otp, req.MessageType)
	if err != nil {
		// Note: OTP was generated but not sent
		render.Render(w, r, NewErrResponse(http.StatusInternalServerError, "Internal Server Error",
			"OTP generated but sending failed: "+err.Error()))
		return
	}

	// Return the OTP in the response (in production, you might just return a success message)
	render.JSON(w, r, map[string]string{
		"status":  "success",
		"message": "OTP generated and sent successfully via " + req.MessageType,
		"otp":     otp, // Note: In production, you might not want to include this
	})
}

// ResendOTPHandler handles requests to resend an existing OTP
func (h *Handler) ResendOTPHandler(w http.ResponseWriter, r *http.Request) {
	var req models.GenerateRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Username == "" {
		render.Render(w, r, NewErrResponse(http.StatusBadRequest, "Bad Request", "Username is required"))
		return
	}

	// Validate message type
	if !isValidMessageType(req.MessageType) {
		render.Render(w, r, NewErrResponse(http.StatusBadRequest, "Bad Request",
			"Unsupported MessageType. Use 'sms', 'whatsapp', or 'email'"))
		return
	}

	// Get the existing OTP
	otp, err := h.OTPService.GetStoredOTP(req.Username)
	if err != nil {
		render.Render(w, r, NewErrResponse(http.StatusNotFound, "Not Found", "No valid OTP exists for this user"))
		return
	}

	// Resend the OTP
	err = h.OTPService.SendOTP(req.Username, otp, req.MessageType)
	if err != nil {
		render.Render(w, r, NewErrResponse(http.StatusInternalServerError, "Internal Server Error",
			"Failed to send OTP: "+err.Error()))
		return
	}

	render.JSON(w, r, map[string]string{
		"status":  "success",
		"message": "OTP resent successfully via " + req.MessageType,
	})
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

// Helper function to validate message type
func isValidMessageType(messageType string) bool {
	return messageType == models.MessageTypeSMS ||
		messageType == models.MessageTypeWhatsApp ||
		messageType == models.MessageTypeEmail
}
