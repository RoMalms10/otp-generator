package models

type GenerateRequest struct {
	Username    string `json:"username"`
	MessageType string `json:"messageType"`
}

type ValidationRequest struct {
	Username string `json:"username"`
	OTP      string `json:"otp"`
}

// Valid message types constant
const (
	MessageTypeEmail    = "email"
	MessageTypeSMS      = "sms"
	MessageTypeWhatsApp = "whatsapp"
)
