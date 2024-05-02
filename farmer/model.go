package farmer

import (
	"circonomy-server/models"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/volatiletech/null"
	"time"
)

type gender string
type croppingPattern string
type errType string
type otpType string
type registrationStatus string
type areaUnit string
type weightUnit string
type transportationVehicle string

const OTPExpiryMinutes time.Duration = 10

const (
	diesel       transportationVehicle = "diesel"
	petrol       transportationVehicle = "petrol"
	nonMotorised transportationVehicle = "non_motorised"

	kg   weightUnit = "kg"
	gm   weightUnit = "gm"
	sack weightUnit = "sack"
	ton  weightUnit = "ton"

	bigha   areaUnit = "bigha"
	hectare areaUnit = "hectare"

	male   gender = "male"
	female gender = "female"
	other  gender = "other"

	monoCropping  croppingPattern = "monoCropping"
	interCropping croppingPattern = "interCropping"
	mixedCropping croppingPattern = "mixedCropping"
	cropRotation  croppingPattern = "cropRotation"

	otpTypeNumber otpType = "number"
	otpTypeEmail  otpType = "email"

	parseError errType = "failed to parse the request"

	verifyAccount registrationStatus = "verify_account"
	createAccount registrationStatus = "create_account"
	addCrop       registrationStatus = "add_crop"
	addFarm       registrationStatus = "add_farm"
)

type sendOTPRequest struct {
	PhoneNumber string `json:"phoneNumber" db:"phone_no"`
	CountryCode string `json:"countryCode" db:"country_code"`
}

type SendOTP struct {
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
	IsValid       bool   `json:"isValid"`
	IsPhoneExists bool   `json:"isPhoneExists"`
	Token         string `json:"token"`
	RefreshToken  string `json:"refreshToken"`
}

type addFarmDetails struct {
	Landmark        string          `json:"landmark" db:"landmark"`
	FieldSize       float32         `json:"fieldSize" db:"size"`
	FieldSizeUnit   string          `json:"fieldSizeUnit" db:"size_unit"`
	CroppingPattern croppingPattern `json:"croppingPattern" db:"size_unit"`
	FarmLatitude    float32         `json:"farmLatitude"`
	FarmLongitude   float32         `json:"farmLongitude"`
	FarmImageIDs    []string        `json:"farmImageIds" db:"farm_images_ids"`
}

type updateFarmerDetails struct {
	Name           string      `json:"fullName" db:"name"`
	Age            int         `json:"age" db:"age"`
	Gender         null.String `json:"gender" db:"gender"`
	Address        string      `json:"address" db:"address"`
	ImageID        null.String `json:"profileImageId" db:"profile_image_id"`
	AadhaarNo      null.String `json:"aadhaarNumber" db:"aadhaar_no"`
	AadhaarImageID null.String `json:"aadhaarImageID" db:"aadhaar_no_image_id"`
}

type addFarmResponse struct {
	FarmID string `json:"farmId"`
}

type messageResponse struct {
	Message string `json:"message"`
}

type changeStatusResponse struct {
	IsChangeStatusPossible bool `json:"isChangeStatusPossible"`
}

type farmerDetails struct {
	ID                 string             `json:"id" db:"id"`
	Name               null.String        `json:"name" db:"name"`
	Age                null.Int           `json:"age" db:"age"`
	Gender             null.String        `json:"gender" db:"gender"`
	Address            null.String        `json:"address" db:"address"`
	PhoneNo            null.String        `json:"phoneNumber" db:"number"`
	CountryCode        null.String        `json:"countryCode" db:"country_code"`
	ProfileImageID     null.String        `json:"-" db:"profile_image_id"`
	ProfileImageURL    *imageURL          `json:"profileImageUrl" `
	AadhaarNo          null.String        `json:"aadhaarNumber" db:"aadhaar_no"`
	AadhaarImageID     null.String        `json:"-" db:"aadhaar_no_image_id"`
	AadhaarImageURL    *imageURL          `json:"aadhaarImageUrl"`
	RegistrationStatus registrationStatus `json:"registrationStatus" db:"registration_status"`
}

type farmDetails struct {
	ID              string         `json:"id" db:"id"`
	Landmark        *string        `json:"landmark" db:"landmark"`
	Size            float32        `json:"size" db:"size"`
	SizeUnit        string         `json:"sizeUnit" db:"size_unit"`
	CroppingPattern string         `json:"croppingPattern" db:"cropping_pattern"`
	FarmLocation    string         `json:"farmLocation" db:"farm_location"`
	FarmImageIDs    pq.StringArray `json:"-" db:"farm_images_ids"`
	FarmImageURLs   []imageURL     `json:"farmImageURLs"`
}

type fetchFarmResponse struct {
	Farms []farmDetails `json:"farms"`
}

type imageURL struct {
	ID        string           `json:"id"`
	CropStage models.StageCrop `json:"cropStage,omitempty"`
	URL       string           `json:"url"`
}

type crop struct {
	ID       uuid.UUID `json:"id" db:"id"`
	Name     string    `json:"name" db:"crop_name"`
	Season   string    `json:"season" db:"season"`
	ImageURL string    `json:"imageUrl,omitempty"`
	ImageID  string    `json:"-" db:"crop_image_id"`
}

type video struct {
	ID                 uuid.UUID        `json:"id" db:"id"`
	Title              string           `json:"title" db:"title"`
	Description        string           `json:"description" db:"description"`
	ThumbnailImageID   uuid.UUID        `json:"-" db:"thumbnail_image_id"`
	ThumbnailImagePath null.String      `json:"-" db:"thumbnail_image_path"`
	ThumbnailImageURL  string           `json:"thumbnailImageURL"`
	VideoURL           string           `json:"videoURL" db:"video_url"`
	VideoType          models.VideoType `json:"videoType" db:"video_tag"`
}

type cropResponse struct {
	Count int    `json:"count"`
	Crops []crop `json:"crops"`
}

type image struct {
	ID    string           `json:"id" db:"id"`
	Path  string           `json:"path" db:"path"`
	Type  string           `json:"type" db:"type"`
	Stage models.StageCrop `json:"-" db:"crop_status"`
}

type videoContent struct {
	ID          string               `json:"id" db:"id"`
	Title       string               `json:"title" db:"title"`
	Description string               `json:"description" db:"description"`
	ContentType models.FarmVideoType `json:"contentType" db:"content_type"`
	URL         string               `json:"url" db:"url"`
}

type fetchVideoResponse struct {
	Count  int            `json:"count"`
	Videos []videoContent `json:"videos"`
}

type addCropsRequest struct {
	IDs []string `json:"ids"`
}

type cropStageTime struct {
	Id           string           `json:"id" db:"id"`
	Stage        models.StageCrop `json:"stage" db:"stage"`
	StartingTime time.Time        `json:"startingTime" db:"starting_time"`
	CropID       string           `json:"cropID,omitempty" db:"farm_crop_id"`
}

type farmCrops struct {
	Id           string          `json:"id" db:"id"`
	FarmID       string          `json:"-" db:"farm_id"`
	Name         string          `json:"name" db:"crop_name"`
	Landmark     null.String     `json:"landmark" db:"landmark"`
	FarmLocation string          `json:"farmLocation" db:"farm_location"`
	Area         *float32        `json:"area" db:"crop_area"`
	AreaUnit     *areaUnit       `json:"areaUnit" db:"crop_area_unit"`
	ImageIDs     pq.StringArray  `json:"-" db:"farm_crop_images_ids"`
	ImageURLs    []imageURL      `json:"imageURLs"`
	StageIDs     pq.StringArray  `json:"-" db:"stage_ids"`
	Stages       []cropStageTime `json:"auditHistory"`
}

type farmerCropsResponse struct {
	Count         int         `json:"count"`
	Page          int         `json:"page"`
	Limit         int         `json:"limit"`
	CropsFiltered []farmCrops `json:"crops"`
}

type cropFormRequest struct {
	FarmID            string                `json:"farmID"`
	CropID            string                `json:"cropID"`
	Stage             models.StageCrop      `json:"stage"`
	CroppingDetails   editCroppingRequest   `json:"croppingDetails"`
	HarvestingDetails editHarvestingRequest `json:"harvestingDetails"`
	SundryingDetails  editSundryingRequest  `json:"sundryingDetails"`
}

type editCroppingRequest struct {
	SeedQuantity null.Float64     `json:"seedQuantity"`
	Unit         weightUnit       `json:"unit"`
	ImageIds     []string         `json:"imageIds"`
	Fertilizers  []fertilizerInfo `json:"fertilizers"`
}

type fertilizerInfo struct {
	Id     string     `json:"fertilizerId" db:"fertilizer_id"`
	Name   string     `json:"name" db:"name"`
	Weight float32    `json:"weight" db:"fertilizer_quantity"`
	Unit   weightUnit `json:"unit" db:"fertilizer_quantity_unit"`
}

type editHarvestingRequest struct {
	YieldQuantity null.Float64 `json:"yieldQuantity"`
	Unit          weightUnit   `json:"unit"`
	ImageIds      []string     `json:"imageIds"`
}

type editSundryingRequest struct {
	ImageIds []string `json:"imageIds"`
}

type addFarmCropResponse struct {
	ID string `json:"id"`
}

type fertilizer struct {
	ID   uuid.UUID `json:"id" db:"id"`
	Name string    `json:"name" db:"name"`
}

type fertilizerResponse struct {
	Count       int          `json:"count"`
	Fertilizers []fertilizer `json:"fertilizers"`
}

type VideoResponse struct {
	Count  int     `json:"count"`
	Videos []video `json:"videos"`
}

type farmCropDetails struct {
	ID                    string           `json:"id" db:"id"`
	FarmerID              string           `json:"farmerId" db:"farmer_id"`
	FarmerName            string           `json:"farmerName" db:"farmer_name"`
	PhoneNo               string           `json:"phoneNo" db:"number"`
	CountryCode           string           `json:"countryCode" db:"country_code"`
	Landmark              null.String      `json:"landmark" db:"landmark"`
	FarmID                null.String      `json:"farmId" db:"farm_id"`
	Name                  null.String      `json:"cropName" db:"crop_name"`
	Season                null.String      `json:"season" db:"season"`
	CropImageID           string           `json:"-" db:"crop_image_id"`
	CropImageURL          imageURL         `json:"cropImageURL"`
	CropArea              null.Float64     `json:"cropArea" db:"crop_area"`
	CropAreaUnit          *areaUnit        `json:"cropAreaUnit" db:"crop_area_unit"`
	CropStage             models.StageCrop `json:"cropStage" db:"crop_stage"`
	YieldQuantity         null.Float64     `json:"yieldQuantity" db:"yield_quantity"`
	YieldQuantityUnit     *weightUnit      `json:"yieldQuantityUnit" db:"yield_quantity_unit"`
	BiomassQuantity       null.Float64     `json:"biomassQuantity" db:"biomass_quantity"`
	BiomassQuantityUnit   *weightUnit      `json:"biomassQuantityUnit" db:"biomass_quantity_unit"`
	SeedQuantity          null.Float64     `json:"seedQuantity,omitempty" db:"seed_quantity"`
	SeedQuantityUnit      *weightUnit      `json:"seedQuantityUnit,omitempty" db:"seed_quantity_unit"`
	TransportationVehicle null.String      `json:"vehicleType" db:"biomass_transportation_vehicle_type"`
	KilnID                null.String      `json:"kilnId,omitempty" db:"kiln_id"`
	KilnName              null.String      `json:"kilnName,omitempty" db:"kiln_name"`
	BiocharQuantity       null.Float64     `json:"biocharQuantity,omitempty" db:"biochar_quantity"`
	BiocharQuantityUnit   null.String      `json:"biocharQuantityUnit,omitempty" db:"biochar_quantity_unit"`
	Fertilizers           []fertilizerInfo `json:"fertilizers"`
	Stages                []cropStageTime  `json:"auditHistory"`
	ImageURLs             []imageURL       `json:"farmImageURLs"`
}

type cropStageRequest struct {
	Status models.StageCrop `json:"status" db:"crop_stage"`
}

type cropStages struct {
	Stages []models.StageCrop `json:"stages"`
}

type imageDetails struct {
	ID    string           `json:"-" db:"id"`
	Path  string           `json:"-" db:"path"`
	Stage models.StageCrop `json:"-" db:"crop_status"`
}

type preferredCrops struct {
	ID           string   `json:"cropId" db:"id"`
	Name         string   `json:"name" db:"crop_name"`
	CropImageID  string   `json:"-" db:"crop_image_id"`
	CropImageURL imageURL `json:"cropImageURL"`
}

type editTransportationRequest struct {
	VehicleType transportationVehicle `json:"vehicleType" db:"biomass_transportation_vehicle_type"`
	ImageIds    []string              `json:"imageIds"`
}

type moveToProductionRequest struct {
	BiomassQuantity     null.Float64 `json:"biomassQuantity" db:"biomass_quantity"`
	BiomassQuantityUnit weightUnit   `json:"biomassQuantityUnit" db:"biomass_quantity_unit"`
	KilnID              string       `json:"kilnId" db:"id"`
	ImageIds            []string     `json:"imageIds"`
}

type editDistributionRequest struct {
	ImageIds []string `json:"imageIds"`
}
