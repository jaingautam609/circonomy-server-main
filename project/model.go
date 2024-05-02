package project

import (
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/volatiletech/null"
)

type projectCreditOperation string

const (
	projectCreditOperationInitial   = "initial"
	projectCreditOperationAddition  = "addition"
	projectCreditOperationDeduction = "deduction"
)

type certificateStatus string

const (
	certificateStatusValid   = "valid"
	certificateStatusExpired = "expired"
)

type projectStatus string

const (
	projectStatusActive   = "active"
	projectStatusSoldOut  = "sold out"
	projectStatusUpcoming = "upcoming"
)

type projectQuality string

const (
	projectQualityHigh = "high quality"
	projectQualityLow  = "low quality"
)

type UUIDSlice []uuid.UUID

func (s *UUIDSlice) Scan(src interface{}) error {
	if src == nil {
		*s = nil
		return nil
	}

	var arr []string
	err := pq.Array(&arr).Scan(src)
	if err != nil {
		return err
	}

	uuids := make([]uuid.UUID, 0, len(arr))
	for _, str := range arr {
		id, err := uuid.Parse(str)
		if err != nil {
			return err
		}
		uuids = append(uuids, id)
	}

	*s = uuids
	return nil
}

type project struct {
	ID              uuid.UUID      `json:"id" db:"id"`
	Name            string         `json:"name" db:"name"`
	ProjectTime     string         `json:"projectTime" db:"project_time"`
	Capacity        int            `json:"capacity" db:"capacity"`
	ImagePath       null.String    `json:"imagePath,omitempty" db:"path"`
	Address         string         `json:"address" db:"address"`
	Lat             float64        `json:"lat" db:"lat"`
	Long            float64        `json:"long" db:"long"`
	Continent       null.String    `json:"continent" db:"continent"`
	Country         null.String    `json:"country" db:"country"`
	Available       int            `json:"available" db:"available"`
	Rate            int            `json:"rate" db:"rate"`
	Method          projectQuality `json:"method" db:"method"`
	Description     string         `json:"description" db:"description"`
	CertificatesIds UUIDSlice      `json:"certificatesIds" db:"certificates_ids"`
	ContactsIds     UUIDSlice      `json:"contactsIds" db:"contacts_ids"`
	ProjectStatus   projectStatus  `json:"projectStatus" db:"project_status"`
	Methodology     null.String    `json:"methodology" db:"methodology"`
}

type createProjectRequest struct {
	Name           string                  `json:"name" db:"name"`
	ProjectTime    string                  `json:"projectTime" db:"project_time"`
	Capacity       int                     `json:"capacity" db:"capacity"`
	Address        string                  `json:"address" db:"address"`
	Lat            float64                 `json:"lat" db:"lat"`
	Long           float64                 `json:"long" db:"long"`
	Continent      null.String             `json:"continent" db:"continent"`
	Country        null.String             `json:"country" db:"country"`
	Available      int                     `json:"available" db:"available"`
	Rate           int                     `json:"rate" db:"rate"`
	Method         projectQuality          `json:"method" db:"method"`
	Description    string                  `json:"description" db:"description"`
	Certificates   []certificateRequest    `json:"certificates" db:"certificates_ids"`
	ProjectDetails []projectDetailsRequest `json:"projectDetails" db:"-"`
	Contacts       []contactRequest        `json:"contacts" db:"contacts_ids"`
	ProjectStatus  projectStatus           `json:"projectStatus" db:"project_status"`
	ImageId        uuid.UUID               `json:"imageId"`
	Methodology    null.String             `json:"methodology" db:"methodology"`
}

type contactRequest struct {
	Name         string      `json:"name"`
	ImageId      uuid.UUID   `json:"imageId"`
	Description  string      `json:"description"`
	Email        string      `json:"email"`
	Designation  null.String `json:"designation"`
	Phone        null.String `json:"phone"`
	LinkedinLink null.String `json:"linkedinLink"`
}

type certificateRequest struct {
	Name    string            `json:"name" db:"name"`
	ImageId uuid.UUID         `json:"imageId"`
	Status  certificateStatus `json:"status"`
}

type projectDetailsRequest struct {
	Name    string    `json:"name" db:"name"`
	ImageId uuid.UUID `json:"imageId"`
}

type details struct {
	ID             uuid.UUID        `json:"id" db:"id"`
	Name           string           `json:"name" db:"name"`
	ProjectTime    string           `json:"projectTime" db:"project_time"`
	Capacity       int              `json:"capacity" db:"capacity"`
	ImagePath      null.String      `json:"imagePath,omitempty" db:"path"`
	ImageURL       string           `json:"imageURL,omitempty"`
	Address        string           `json:"address" db:"address"`
	Lat            float64          `json:"lat" db:"lat"`
	Long           float64          `json:"long" db:"long"`
	Continent      null.String      `json:"continent" db:"continent"`
	Country        null.String      `json:"country" db:"country"`
	Available      int              `json:"available" db:"available"`
	Rate           int              `json:"rate" db:"rate"`
	Method         projectQuality   `json:"method" db:"method"`
	Description    string           `json:"description" db:"description"`
	Certificates   []certificate    `json:"certificates" db:"certificates"`
	ProjectDetails []projectDetails `json:"projectDetails" db:"-"`
	Contacts       []contact        `json:"contacts" db:"contacts"`
	ProjectStatus  projectStatus    `json:"projectStatus" db:"project_status"`
	Methodology    null.String      `json:"methodology"`
}

type certificate struct {
	ID        uuid.UUID         `json:"id" db:"id"`
	Name      string            `json:"name" db:"name"`
	ImagePath null.String       `json:"imagePath,omitempty" db:"path"`
	ImageURL  string            `json:"imageURL,omitempty"`
	Status    certificateStatus `json:"status" db:"status"`
}

type projectDetails struct {
	ID        uuid.UUID   `json:"id" db:"id"`
	Name      string      `json:"name" db:"name"`
	ImagePath null.String `json:"imagePath,omitempty" db:"path"`
	ImageURL  string      `json:"imageURL,omitempty"`
}

type contact struct {
	ID           uuid.UUID   `json:"id" db:"id"`
	Name         string      `json:"name" db:"name"`
	ImagePath    null.String `json:"imagePath,omitempty" db:"path"`
	ImageURL     string      `json:"imageURL,omitempty"`
	Description  string      `json:"description" db:"description"`
	Email        string      `json:"email" db:"email"`
	Designation  null.String `json:"designation" db:"designation"`
	Phone        null.String `json:"phone" db:"phone"`
	LinkedinLink null.String `json:"linkedinLink" db:"linkedin_link"`
}

type locationFilter struct {
	Address    string `json:"address" db:"address"`
	IsFiltered bool   `json:"isFiltered"`
}

type locationDetails struct {
	Address   string      `json:"address" db:"address"`
	Lat       float64     `json:"lat" db:"lat"`
	Long      float64     `json:"long" db:"long"`
	Continent null.String `json:"continent" db:"continent"`
	Country   null.String `json:"country" db:"country"`
}
