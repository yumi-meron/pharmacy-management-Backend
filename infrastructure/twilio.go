package infrastructure

import (
	"context"
	"fmt"
	"os"

	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

type TwilioService struct {
	Client          *twilio.RestClient
	FromPhoneNumber string
}

func NewTwilioService() *TwilioService {
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: os.Getenv("TWILIO_ACCOUNT_SID"),
		Password: os.Getenv("TWILIO_AUTH_TOKEN"),
	})
	return &TwilioService{
		Client:          client,
		FromPhoneNumber: os.Getenv("TWILIO_PHONE_NUMBER"),
	}
}

func (s *TwilioService) SendOTP(ctx context.Context, toPhone string, otp string) error {
	params := &openapi.CreateMessageParams{}
	params.SetTo(toPhone)
	params.SetFrom(s.FromPhoneNumber)
	params.SetBody(fmt.Sprintf("Your OTP code is: %s", otp))

	_, err := s.Client.Api.CreateMessage(params)
	return err
}
