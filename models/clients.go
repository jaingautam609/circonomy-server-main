package models

import (
	"github.com/google/uuid"
	"github.com/volatiletech/null"
)

type Clients struct {
	UserID       uuid.UUID    `json:"userID" db:"id"`
	Name         string       `json:"name" db:"name"`
	ImagePath    null.String  `json:"imagePath,omitempty" db:"path"`
	TotalCredits null.Float64 `json:"totalCredits" db:"total_credits"`
	ImageURL     string       `json:"imageURL,omitempty"`
}
