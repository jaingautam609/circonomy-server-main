package providers

import (
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/volatiletech/null"
)

type EmailType string

const (
	EmailTypeResetPassword      EmailType = "reset_password"
	EmailTypeVerifyEmail        EmailType = "verify_email"
	EmailTypeInviteFamilyMember EmailType = "invite_family_member"
	EmailTypeContactUs          EmailType = "contactUs"
	EmailTypeSubscribe          EmailType = "subscribe"
)

// EmailProvider provides the email service to send emails.
type EmailProvider interface {
	// Send the email to all recipients
	Send(dt *DynamicTemplate) error

	// GetEmailTemplate returns the dynamic template from the email provider(sendgrid) service.
	GetEmailTemplate(emailType EmailType) (*DynamicTemplate, error)
}

type DynamicTemplate struct {
	TemplateID  string
	DynamicData map[string]interface{}
	Categories  []string
	Recipients  []mail.Email
	FromEmail   null.String
	Attachments []*mail.Attachment
}

func (d *DynamicTemplate) AddRecipient(name, email string) {
	d.Recipients = append(d.Recipients, mail.Email{
		Name:    name,
		Address: email,
	})
}
