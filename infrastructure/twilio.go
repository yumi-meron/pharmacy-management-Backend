package infrastructure

import (
	"fmt"

	"pharmacist-backend/config"

	"github.com/rs/zerolog"
	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
)

// TwilioService handles SMS sending via Twilio
type TwilioService struct {
	client *twilio.RestClient
	cfg    *config.Config
	logger zerolog.Logger
}

// NewTwilioService creates a new Twilio service
func NewTwilioService(cfg *config.Config, logger zerolog.Logger) *TwilioService {
	var client *twilio.RestClient
	if !cfg.TwilioMockMode {
		client = twilio.NewRestClientWithParams(twilio.ClientParams{
			Username: cfg.TwilioAccountSID, // Use Username instead of AccountSid
			Password: cfg.TwilioAuthToken,  // Use Password instead of AuthToken
		})
	}
	return &TwilioService{
		client: client,
		cfg:    cfg,
		logger: logger,
	}
}

// SendSMS sends an SMS message to the specified phone number
func (s *TwilioService) SendSMS(to, body string) error {
	if s.cfg.TwilioMockMode {
		s.logger.Info().Str("to", to).Str("body", body).Msg("Mock SMS sent")
		return nil
	}

	params := &twilioApi.CreateMessageParams{}
	params.SetTo(to)
	params.SetFrom(s.cfg.TwilioPhoneNumber)
	params.SetBody(body)

	_, err := s.client.Api.CreateMessage(params)
	if err != nil {
		s.logger.Error().Err(err).Str("to", to).Msg("Failed to send SMS")
		return fmt.Errorf("failed to send SMS: %w", err)
	}

	s.logger.Info().Str("to", to).Msg("SMS sent successfully")
	return nil
}
