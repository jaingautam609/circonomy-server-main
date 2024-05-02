package family

import (
	"circonomy-server/models"
	"circonomy-server/providers"
	"circonomy-server/utils"
	"database/sql"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type Service struct {
	repository    *Repository
	emailProvider providers.EmailProvider
}

func NewService(jobsRepository *Repository, emailProvider providers.EmailProvider) *Service {
	return &Service{
		repository:    jobsRepository,
		emailProvider: emailProvider,
	}
}

func (s *Service) createFamily(request createFamilyRequest, user *models.UserSessionInfo) (*family, error) {
	if user.AccountType == models.UserAccountTypeCorporate {
		return nil, errors.New("Corporate user can not create a family")
	} else if user.AccountType == models.UserAccountTypeSME {
		return nil, errors.New("SME user can not create a family")
	}
	return s.repository.createFamily(request, user.ID)
}

func (s *Service) updateFamily(request updateFamilyRequest, familyID uuid.UUID, userID uuid.UUID) (*family, error) {
	return s.repository.updateFamily(request, familyID, userID)
}

func (s *Service) getFamily(familyID uuid.UUID, userID uuid.UUID) (*family, error) {
	return s.repository.getFamily(familyID, userID)
}

func (s *Service) invite(familyID uuid.UUID, email string) (string, error) {
	existingEmail, err := s.repository.checkAlreadyInvitedEmail(familyID, email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	//existing email case in already users of the family
	if existingEmail == email {
		return "This email has already been invited. Please ask the member to check the emails", nil
	}

	existingEmail, err = s.repository.checkAlreadyExistingMemberEmail(familyID, email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	//existing email case in already users of the family
	if existingEmail == email {
		return "The person with this email is already a member of the family", nil
	}

	invitationID, err := s.repository.invite(familyID, email)
	if err != nil {
		return "", err
	}

	familyName, ownerName := s.repository.getFamilyNameOwnerName(familyID)

	template, _ := s.emailProvider.GetEmailTemplate(providers.EmailTypeInviteFamilyMember)
	template.AddRecipient("", email)
	template.DynamicData["url"] = utils.GetInviteUrl(invitationID)
	template.DynamicData["ownerName"] = ownerName
	template.DynamicData["familyName"] = familyName
	s.emailProvider.Send(template)

	return "The user has been successfully invited", nil
}

func (s *Service) getOwnFamily(userID uuid.UUID) (*family, error) {
	return s.repository.getOwnFamily(userID)
}

func (s *Service) getInvitationDetails(invitationID, userID uuid.UUID) (*invitationDetails, error) {
	var details invitationDetails
	var err error
	details.InvitationOwnerDetails, err = s.repository.getInvitationDetails(invitationID)
	if err != nil {
		return nil, err
	}
	if details.InvitationOwnerDetails.ImagePath.Valid {
		url, urlErr := utils.GenerateSignedURL(details.InvitationOwnerDetails.ImagePath.String)
		if urlErr != nil {
			return nil, err
		}
		details.InvitationOwnerDetails.ImageUrl = url
	}

	details.CurrentOwnerDetails, err = s.repository.getCurrentFamilyDetails(userID)
	if details.CurrentOwnerDetails != nil && details.CurrentOwnerDetails.ImagePath.Valid {
		url, urlErr := utils.GenerateSignedURL(details.CurrentOwnerDetails.ImagePath.String)
		if urlErr != nil {
			return nil, err
		}
		details.CurrentOwnerDetails.ImageUrl = url
	}

	return &details, nil
}

func (s *Service) GetInvitations(userID uuid.UUID) ([]*OwnerDetails, error) {
	ownerDetails, err := s.repository.getCurrentInvitations(userID)
	if err != nil {
		return nil, err
	}
	for i := range ownerDetails {
		if ownerDetails[i].ImagePath.Valid {
			url, urlErr := utils.GenerateSignedURL(ownerDetails[i].ImagePath.String)
			if urlErr != nil {
				return nil, err
			}

			ownerDetails[i].ImageUrl = url
		}
	}
	return ownerDetails, nil
}

func (s *Service) joinFamily(invitationID, userID uuid.UUID) error {
	return s.repository.joinFamily(invitationID, userID)
}
