package family

import (
	"circonomy-server/models"
	"github.com/google/uuid"
	"github.com/volatiletech/null"
)

type createFamilyRequest struct {
	Name string `json:"name"`
}

type updateFamilyRequest struct {
	Name string `json:"name"`
}

type inviteFamilyRequest struct {
	Email    string    `json:"email"`
	FamilyID uuid.UUID `json:"familyId"`
}

type family struct {
	Id        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	CreatedBy uuid.UUID `json:"createdBy" db:"created_by"`
	CreatedAt string    `json:"createdAt" db:"created_at"`

	CanSendInvitation bool                `json:"canSendInvitation"`
	Members           []*familyMember     `json:"members"`
	Invitations       []*familyInvitation `json:"invitations"`
}

type familyMember struct {
	Id           uuid.UUID    `json:"id" db:"id"`
	Name         string       `json:"name" db:"name"`
	Email        string       `json:"email" db:"email"`
	Number       null.String  `json:"number" db:"number"`
	ImagePath    null.String  `json:"imagePath,omitempty" db:"path"`
	ImageUrl     string       `json:"imageUrl"`
	ProjectCount null.Int     `json:"projectCount" db:"project_count"`
	TotalCredits null.Float64 `json:"totalCredits" db:"total_credits"`

	CreditHistory []models.CreditsHistory
}

type familyInvitation struct {
	Email     string `json:"email" db:"email"`
	CreatedAt string `json:"createdAt" db:"created_at"`
}

type OwnerDetails struct {
	FamilyID     string      `json:"familyId" db:"family_id"`
	FamilyName   string      `json:"familyName" db:"family_name"`
	UserID       uuid.UUID   `json:"userId" db:"user_id"`
	Name         string      `json:"name" db:"name"`
	Email        string      `json:"email" db:"email"`
	Number       null.String `json:"number" db:"number"`
	ImagePath    null.String `json:"imagePath,omitempty" db:"path"`
	ImageUrl     string      `json:"imageUrl,omitempty"`
	InvitationId uuid.UUID   `json:"invitationId" db:"invitation_id"`
}

type invitationDetails struct {
	InvitationOwnerDetails *OwnerDetails `json:"invitationOwnerDetails"`
	CurrentOwnerDetails    *OwnerDetails `json:"currentOwnerDetails"`
}
