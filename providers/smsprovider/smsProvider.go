package smsprovider

import (
	"circonomy-server/models"
	"circonomy-server/providers"
	"fmt"
	"github.com/pkg/errors"
	"net/http"
	"os"
)

type SMSProvider struct {
}

func NewSMSProvider() providers.SMSProvider {
	return &SMSProvider{}
}

func (s SMSProvider) Send(payload models.SendOTP) error {
	apiKey := os.Getenv("SMS_API_KEY")
	url := fmt.Sprintf("https://2factor.in/API/V1/%s/SMS/%s%s/%s/test", apiKey, payload.CountryCode, payload.PhoneNumber, payload.OTP)
	req, _ := http.NewRequest("GET", url, nil)
	res, _ := http.DefaultClient.Do(req)
	if res.StatusCode != http.StatusOK {
		return errors.New("error while sending otp")
	}
	defer res.Body.Close()
	return nil
}
