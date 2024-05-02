package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/volatiletech/null"
)

type UserProfile struct {
	Name         string          `json:"name" db:"name"`
	Email        string          `json:"email" db:"email"`
	Number       null.String     `json:"number" db:"number"`
	CountryCode  null.String     `json:"countryCode" db:"country_code"`
	Address      string          `json:"address" db:"address"`
	AccountType  UserAccountType `json:"accountType" db:"account_type"`
	OrgName      string          `json:"orgName,omitempty"`
	OrgID        uuid.NullUUID   `json:"orgID"`
	OrgSize      null.String     `json:"orgSize,omitempty" db:"size"`
	ImagePath    null.String     `json:"imagePath,omitempty" db:"path"`
	ImageURL     string          `json:"imageURL,omitempty"`
	UploadID     uuid.NullUUID   `json:"uploadID" db:"upload_id"`
	TotalCredits null.Float64    `json:"totalCredits" db:"total_credits"`
}

type EditPersonProfileInput struct {
	Name     string        `json:"name" db:"name"`
	Address  string        `json:"address" db:"address"`
	OrgID    uuid.NullUUID `json:"orgID"`
	OrgName  string        `json:"orgName,omitempty"`
	UploadID string        `json:"uploadID" db:"upload_id"`
}

type EditPersonProfile struct {
	Name     string        `json:"name" db:"name"`
	Address  string        `json:"address" db:"address"`
	OrgName  string        `json:"orgName,omitempty"`
	OrgID    uuid.NullUUID `json:"orgID"`
	UploadID string        `json:"uploadID" db:"upload_id"`
}

type EditOrgProfile struct {
	Name     string        `json:"name" db:"name"`
	Address  string        `json:"address" db:"address"`
	UploadID uuid.NullUUID `json:"uploadID" db:"upload_id"`
	Size     string        `json:"size" db:"size"`
}

type EditOrgProfileInput struct {
	Name     string        `json:"name" db:"name"`
	Address  string        `json:"address" db:"address"`
	UploadID uuid.NullUUID `json:"uploadID" db:"upload_id"`
	Size     string        `json:"size" db:"size"`
}

type OrgPeople struct {
	ID            uuid.UUID    `json:"id" db:"id"`
	Name          string       `json:"name" db:"name"`
	OrgName       string       `json:"orgName,omitempty" db:"organization_name"`
	Address       string       `json:"address" db:"address"`
	ImagePath     null.String  `json:"imagePath,omitempty" db:"path"`
	ImageURL      string       `json:"imageURL,omitempty"`
	TotalCredits  null.Float64 `json:"totalCredits" db:"total_credits"`
	TotalProjects float64      `json:"totalProjects" db:"total_projects"`
}

type CreditsHistory struct {
	ProjectID   uuid.UUID   `json:"projectId" db:"id"`
	ProjectName string      `json:"projectName" db:"name"`
	BoughtBy    string      `json:"boughtBy,omitempty" db:"bought_by_name"`
	Quantity    float64     `json:"quantity" db:"credits"`
	Rate        float64     `json:"rate" db:"bought_at_rate"`
	TotalAmount float64     `json:"totalAmount" db:"bought_at_cost"`
	BoughtDate  time.Time   `json:"boughtDate" db:"created_at"`
	ImagePath   null.String `json:"imagePath,omitempty" db:"path"`
	ImageURL    string      `json:"imageURL,omitempty"`
}
