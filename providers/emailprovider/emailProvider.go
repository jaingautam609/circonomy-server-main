package emailprovider

import (
	"circonomy-server/providers"
	"errors"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/sirupsen/logrus"
)

type emailProvider struct {
	sendgridClient *sendgrid.Client
}

func NewSendGridEmailProvider(apiKey string) providers.EmailProvider {
	return &emailProvider{
		sendgridClient: sendgrid.NewSendClient(apiKey),
	}
}

func (e emailProvider) Send(dt *providers.DynamicTemplate) error {
	//if env.InKubeCluster() {
	//	if !env.IsMain() {
	//		dt.DynamicData["env"] = "[DEV] "
	//	}
	//} else {
	//	dt.DynamicData["env"] = "[LOCAL] "
	//}

	for _, recipient := range dt.Recipients {
		newMail := mail.NewV3Mail()
		newMail.TemplateID = dt.TemplateID
		fromEmail := "noreply@circonomy.co"
		if dt.FromEmail.Valid {
			fromEmail = dt.FromEmail.String
		}
		newMail.From = &mail.Email{
			Name:    "Team Circonomy",
			Address: fromEmail,
		}
		newMail.Personalizations = []*mail.Personalization{}

		personalization := mail.NewPersonalization()
		personalization.To = []*mail.Email{
			{
				Name:    recipient.Name,
				Address: recipient.Address,
			},
		}
		personalization.Categories = []string{"Circonomy"}
		for i := range dt.Categories {
			personalization.Categories = append(personalization.Categories, dt.Categories[i])
		}

		for key, value := range dt.DynamicData {
			personalization.DynamicTemplateData[key] = value
		}

		newMail.AddPersonalizations(personalization)

		// if there is attachment, add it
		if len(dt.Attachments) > 0 {
			newMail.Attachments = dt.Attachments
		}

		resp, err := e.sendgridClient.Send(newMail)
		if resp.StatusCode > 300 {
			logrus.Errorf("error sending email: %v", resp.Body)
		}
		if err != nil {
			logrus.Errorf("error sending email: %v", err)
		}
	}

	return nil
}

func (e emailProvider) GetEmailTemplate(emailType providers.EmailType) (*providers.DynamicTemplate, error) {
	switch emailType {
	case providers.EmailTypeVerifyEmail:
		return verifyEmailTemplate()
	case providers.EmailTypeResetPassword:
		return forgotPasswordTemplate()
	case providers.EmailTypeInviteFamilyMember:
		return inviteMemberTemplate()
	case providers.EmailTypeContactUs:
		return contactUsTemplate()
	case providers.EmailTypeSubscribe:
		return subscribeTemplate()
	default:
		return nil, errors.New("email type invalid")
	}
}
