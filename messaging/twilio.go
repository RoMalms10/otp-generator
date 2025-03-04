package messaging

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// TwilioConfig holds configuration for Twilio API
type TwilioConfig struct {
	AccountSID  string
	AuthToken   string
	PhoneNumber string
	BaseURL     string
}

// TwilioService implements SMS and WhatsApp sending functionality
type TwilioService struct {
	Config TwilioConfig
}

// NewTwilioService creates a new TwilioService with the specified configuration
func NewTwilioService(config TwilioConfig) *TwilioService {
	// Set default API endpoint if not specified
	if config.BaseURL == "" {
		config.BaseURL = "https://api.twilio.com/2010-04-01"
	}

	return &TwilioService{
		Config: config,
	}
}

// SendSMS sends an SMS message through Twilio
func (s *TwilioService) SendSMS(to, message string) error {
	// Format the API URL
	apiURL := fmt.Sprintf("%s/Accounts/%s/Messages.json", s.Config.BaseURL, s.Config.AccountSID)

	// Create form data
	formData := url.Values{}
	formData.Set("To", to)
	formData.Set("From", s.Config.PhoneNumber)
	formData.Set("Body", message)

	// Create HTTP request
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.SetBasicAuth(s.Config.AccountSID, s.Config.AuthToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send SMS: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("twilio API returned non-success status code: %d", resp.StatusCode)
	}

	return nil
}

// SendWhatsApp sends a WhatsApp message through Twilio
func (s *TwilioService) SendWhatsApp(to, message string) error {
	// Format the API URL
	apiURL := fmt.Sprintf("%s/Accounts/%s/Messages.json", s.Config.BaseURL, s.Config.AccountSID)

	// Format WhatsApp number (prefixed with "whatsapp:" for Twilio)
	fromWhatsApp := fmt.Sprintf("whatsapp:%s", s.Config.PhoneNumber)
	toWhatsApp := fmt.Sprintf("whatsapp:%s", to)

	// Create form data
	formData := url.Values{}
	formData.Set("To", toWhatsApp)
	formData.Set("From", fromWhatsApp)
	formData.Set("Body", message)

	// Create HTTP request
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.SetBasicAuth(s.Config.AccountSID, s.Config.AuthToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send WhatsApp message: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("twilio API returned non-success status code: %d", resp.StatusCode)
	}

	return nil
}

// SendOTP sends a one-time password via SMS
func (s *TwilioService) SendOTP(to, otp string) error {
	message := fmt.Sprintf("Your verification code is: %s. It will expire in 10 minutes.", otp)
	return s.SendSMS(to, message)
}

// SendOTPWhatsApp sends a one-time password via WhatsApp
func (s *TwilioService) SendOTPWhatsApp(to, otp string) error {
	message := fmt.Sprintf("Your verification code is: %s. It will expire in 10 minutes.", otp)
	return s.SendWhatsApp(to, message)
}
