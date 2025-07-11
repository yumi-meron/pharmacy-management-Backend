package infrastructure

import (
	"pharmacy-management-backend/config"

	"github.com/rs/zerolog"
	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
)

// TwilioService defines the interface for sending SMS
type TwilioService struct {
	client *twilio.RestClient
	from   string
	logger zerolog.Logger
	mock   bool
}

// NewTwilioService creates a new TwilioService
func NewTwilioService(cfg *config.Config, logger zerolog.Logger) *TwilioService {
	var client *twilio.RestClient
	if !cfg.MockTwilio {
		client = twilio.NewRestClientWithParams(twilio.ClientParams{
			Username: cfg.TwilioSID,
			Password: cfg.TwilioToken,
		})
	}
	return &TwilioService{
		client: client,
		from:   cfg.TwilioFrom,
		logger: logger,
		mock:   cfg.MockTwilio,
	}
}

// SendSMS sends an SMS message
func (s *TwilioService) SendSMS(to, body string) error {
	if s.mock {
		s.logger.Info().Str("to", to).Str("body", body).Msg("Mock SMS sent")
		return nil
	}

	params := &twilioApi.CreateMessageParams{}
	params.SetTo(to)
	params.SetFrom(s.from)
	params.SetBody(body)

	_, err := s.client.Api.CreateMessage(params)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to send SMS")
		return err
	}
	s.logger.Info().Str("to", to).Msg("SMS sent successfully")
	return nil
}
