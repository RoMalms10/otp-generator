package models

type GenerateRequest struct {
	Username    string `json:"username"`
	MessageType string `json:"messageType"`
}

type ValidationRequest struct {
	Username string `json:"username"`
	OTP      string `json:"otp"`
}
