package providers

import (
	"circonomy-server/models"
)

type SMSProvider interface {
	Send(models.SendOTP) error
}
