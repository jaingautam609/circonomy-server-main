package kilnOperator

import (
	"github.com/lib/pq"
	"github.com/volatiletech/null"
	"time"
)

type gender string
type otpType string
type errType string
type weightUnit string
type fileType string

const (
	OTPExpiryMinutes time.Duration = 10

	otpTypeNumber otpType = "number"
	otpTypeEmail  otpType = "email"

	parseError errType = "failed to parse the request"

	kg   weightUnit = "kg"
	gm   weightUnit = "gm"
	sack weightUnit = "sack"
	ton  weightUnit = "ton"

	male   gender = "male"
	female gender = "female"
	other  gender = "other"

	imageFileType fileType = "image"
	videoFileType fileType = "video"
)

type sendOTPRequest struct {
	PhoneNumber string `json:"phoneNumber" db:"phone_no"`
	CountryCode string `json:"countryCode" db:"country_code"`
}

type sendOTP struct {
	PhoneNumber string    `json:"phoneNumber" db:"phone_no"`
	CountryCode string    `json:"countryCode" db:"country_code"`
	OTP         string    ` db:"otp"`
	Type        string    `db:"type"`
	Expiry      time.Time ` db:"expiry"`
}

type sendOTPResponse struct {
	Message string `json:"message"`
}

type checkOTPRequest struct {
	PhoneNumber string `json:"phoneNumber" db:"phone_no"`
	CountryCode string `json:"countryCode" db:"country_code"`
	OTP         string `json:"otp" db:"otp"`
}

type checkOTPResponse struct {
	IsValid      bool   `json:"isValid"`
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
}

type image struct {
	ID   string `json:"id"`
	Path string `json:"path"`
	Type string `json:"type"`
}

type video struct {
	ID            string `db:"id"`
	KlinProcessId string `db:"klin_process_id"`
	Path          string `db:"path"`
}

type imageURL struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

type editFarmCropBiomassRequest struct {
	BiomassQuantity null.Float64 `json:"biomassQuantity"`
	Unit            *weightUnit  `json:"unit"`
	ImageIds        []string     `json:"imageIds"`
}

type farmCropBiomasses struct {
	Id              string         `json:"id" db:"id"`
	FarmerName      string         `json:"farmerName" db:"farmer_name"`
	CropID          string         `json:"cropId,omitempty" db:"crop_id"`
	CropName        string         `json:"cropName" db:"crop_name"`
	KilnID          string         `json:"kilnId,omitempty" db:"kiln_id"`
	PhoneNo         string         `json:"phoneNo" db:"number"`
	CountryCode     string         `json:"countryCode" db:"country_code"`
	BiomassQuantity null.Float64   `json:"biomassQuantity" db:"biomass_quantity"`
	Unit            *weightUnit    `json:"unit" db:"biomass_quantity_unit"`
	ImageIDs        pq.StringArray `json:"-" db:"farm_crop_images_ids"`
	ImageURLs       []imageURL     `json:"imageURLs"`
}

type farmCropBiomassDetails struct {
	Count         int                 `json:"count"`
	Page          int                 `json:"page"`
	Limit         int                 `json:"limit"`
	CropsFiltered []farmCropBiomasses `json:"crops"`
}

type kilnInfo struct {
	ID              string       `json:"id" db:"id"`
	Name            string       `json:"name" db:"name"`
	NetworkID       string       `json:"networkId" db:"network_id"`
	NetworkName     string       `json:"networkName" db:"network_name"`
	Address         string       `json:"address"`
	BiocharQuantity null.Float64 `json:"biocharQuantity" db:"biochar_quantity"`
}

type kilnOperatorDetails struct {
	ID              string      `json:"id" db:"id"`
	Name            string      `json:"name" db:"name"`
	Age             int         `json:"age" db:"age"`
	Gender          gender      `json:"gender" db:"gender"`
	Address         string      `json:"address" db:"address"`
	PhoneNo         string      `json:"phoneNumber" db:"number"`
	CountryCode     string      `json:"countryCode" db:"country_code"`
	ProfileImageID  null.String `json:"-" db:"profile_image_id"`
	ProfileImageURL imageURL    `json:"profileImageUrl" `
	AadhaarNo       null.String `json:"aadhaarNumber" db:"aadhaar_no"`
	AadhaarImageID  null.String `json:"-" db:"aadhaar_no_image_id"`
	AadhaarImageURL imageURL    `json:"aadhaarImageUrl"`
	KilnInfo        []kilnInfo  `json:"kilnInfo"`
}

type kilnBiomass struct {
	KilnID                 string       `json:"kilnID" db:"kiln_id"`
	CropID                 string       `json:"cropID" db:"crop_id"`
	CropName               string       `json:"cropName" db:"crop_name"`
	CurrentBiomassQuantity null.Float64 `json:"currentBiomassQuantity" db:"current_quantity"`
}

type kilnProcesses struct {
	Id              string         `json:"id" db:"id"`
	StartingDate    time.Time      `json:"startingDate" db:"starting_date"`
	StartTime       time.Time      `json:"startTime" db:"created_at"`
	EndTime         null.Time      `json:"endTime" db:"end_time"`
	BiomassQuantity null.Float64   `json:"biomassQuantity" db:"biomass_quantity"`
	BiCharQuantity  null.Float64   `json:"biCharQuantity" db:"biochar_quantity"`
	KilnID          string         `json:"kilnID" db:"kiln_id"`
	CropID          string         `json:"cropID" db:"crop_id"`
	CropName        string         `json:"cropName" db:"crop_name"`
	ImageIDs        pq.StringArray `json:"-" db:"kiln_process_images_ids"`
	ImageURLs       []imageURL     `json:"imageURLs"`
	VideoURLs       []imageURL     `json:"VideoURLs"`
}

type kilnProcessResponse struct {
	Page          int             `json:"page"`
	Limit         int             `json:"limit"`
	Count         int             `json:"count"`
	KilnProcesses []kilnProcesses `json:"kilnProcesses" db:""`
}

type kilnProcessRequest struct {
	CropID          string       `json:"cropId" db:"crop_Id"`
	BiomassQuantity null.Float64 `json:"biomassQuantity" db:"biomass_quantity"`
	ImageIds        []string     `json:"imageIds"`
	VideoIds        []string     `json:"VideoIds"`
}

type kilnProcessEditRequest struct {
	KilnProcessID string   `json:"kilnProcessId" db:"id"`
	ImageIds      []string `json:"imageIds"`
	VideoIds      []string `json:"VideoIds"`
}

type kilnProcessDoneRequest struct {
	KilnProcessID   string       `json:"kilnProcessId" db:"id"`
	BioCharQuantity null.Float64 `json:"bioCharQuantity" db:"biochar_quantity"`
	ImageIds        []string     `json:"imageIds"`
	VideoIds        []string     `json:"VideoIds"`
}

type kilnBioChar struct {
	BiocharQuantity null.Float64 `json:"biocharQuantity" db:"biochar_quantity"`
}

type kilnDistributionRequest struct {
	KilnID          string       `json:"kilnId" db:"id"`
	BioCharQuantity null.Float64 `json:"biocharQuantity" db:"biochar_quantity"`
	ImageIds        []string     `json:"imageIds"`
}

type distributedFarmCrop struct {
	ID                  string         `json:"id" db:"id"`
	FarmerName          null.String    `json:"farmerName" db:"farmer_name"`
	PhoneNo             null.String    `json:"phoneNumber" db:"number"`
	CountryCode         null.String    `json:"countryCode" db:"country_code"`
	DistributedDate     null.Time      `json:"distributedDate" db:"starting_time"`
	BiomassQuantity     null.Float64   `json:"biomassQuantity" db:"biomass_quantity"`
	BiomassQuantityUnit *weightUnit    `json:"biomassQuantityUnit" db:"biomass_quantity_unit"`
	BiocharQuantity     null.Float64   `json:"biocharQuantity,omitempty" db:"biochar_quantity"`
	BiocharQuantityUnit null.String    `json:"biocharQuantityUnit,omitempty" db:"biochar_quantity_unit"`
	ImageIDs            pq.StringArray `json:"-" db:"farm_images_ids"`
	ImageURLs           []imageURL     `json:"farmImageURLs"`
	FarmCropUpdatedAt   null.Time      `json:"-" db:"updated_at"`
}

type distributedFarmCropsResponse struct {
	Page                int                   `json:"page"`
	Limit               int                   `json:"limit"`
	Count               int                   `json:"count"`
	DistributedFarmCrop []distributedFarmCrop `json:"distributedFarmCrop"`
}
