package admin

import (
	"circonomy-server/models"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/volatiletech/null"
)

type errType string

const (
	parseError errType = "failed to parse the request"
)

type messageResponse struct {
	Message string `json:"message"`
}

type AddCropRequest struct {
	Name    string    `json:"name" db:"crop_name"`
	Season  string    `json:"season" db:"season"`
	ImageID uuid.UUID `json:"imageId" db:"crop_image_id"`
}

type AddVideoRequest struct {
	Title       string           `json:"title"`
	Description string           `json:"description"`
	ImageID     uuid.UUID        `json:"imageId"`
	VideoURL    string           `json:"videoURL"`
	VideoType   models.VideoType `json:"videoType"`
}

type addCropResponse struct {
	ID string `json:"id"`
}

type AddFertilizerRequest struct {
	Name string `json:"name" db:"name"`
}

type addVideoResponse struct {
	ID string `json:"id"`
}

type loginResponse struct {
	IsValid      bool   `json:"isValid"`
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
}

type biomassAggregatorRequest struct {
	Name         null.String                     `json:"name" db:"name"`
	Latitude     float32                         `json:"latitude"`
	Longitude    float32                         `json:"longitude"`
	LocationName null.String                     `json:"locationName" db:"location_name"` // add column in table
	Manager      biomassAggregatorManagerRequest `json:"manager"`
}

type biomassAggregatorManagerRequest struct {
	Email       null.String `json:"email" db:"email"`
	Password    null.String `json:"password" db:"password"`
	PhoneNo     null.String `json:"phoneNumber" db:"number"`
	CountryCode null.String `json:"countryCode" db:"country_code"`
	Trained     bool        `json:"trained" db:"trained"`
}

type biomassAggregator struct {
	ID           string      `json:"id" db:"id"`
	Name         null.String `json:"name" db:"name"`
	Email        null.String `json:"email" db:"email"`
	CSinkCount   int         `json:"CSinkCount" db:"c_sink_count"`
	FarmersCount int         `json:"farmersCount" db:"farmers_count"`
	Trained      bool        `json:"trained" db:"trained"`
	LocationName null.String `json:"locationName" db:"location_name"`
}

type biomassAggregatorResponse struct {
	Count              int                 `json:"count"`
	Limit              int                 `json:"limit"`
	Page               int                 `json:"page"`
	BiomassAggregators []biomassAggregator `json:"biomassAggregators"`
}

type biomassAggregatorDetailsResponse struct {
	ID           string      `json:"id" db:"id"`
	Name         null.String `json:"name" db:"name"`
	Email        null.String `json:"email" db:"email"`
	Trained      bool        `json:"trained" db:"trained"`
	LocationName null.String `json:"locationName" db:"location_name"`
	Location     null.String `json:"location" db:"location"`
	PhoneNo      null.String `json:"phoneNo" db:"number"`
	CountryCode  null.String `json:"countryCode" db:"country_code"`
}

type biomassAggregatorDetailsRequest struct {
	Name         null.String `json:"name" db:"name"`
	Trained      bool        `json:"trained" db:"trained"`
	LocationName null.String `json:"locationName" db:"location_name"`
	Latitude     float32     `json:"latitude"`
	Longitude    float32     `json:"longitude"`
}

type bANetwork struct {
	ID           string      `json:"id" db:"id"`
	Name         null.String `json:"name" db:"name"`
	LocationName null.String `json:"locationName" db:"location_name"`
	FarmersCount int         `json:"farmersCount" db:"farmers_count"`
	ManagerName  null.String `json:"managerName" db:"manager_name"`
}

type bANetworkResponse struct {
	Count              int         `json:"count"`
	Limit              int         `json:"limit"`
	Page               int         `json:"page"`
	BiomassAggregators []bANetwork `json:"biomassAggregators"`
}

type networkRequest struct {
	Name          null.String `json:"name" db:"name"`
	LocationName  null.String `json:"locationName" db:"location_name"`
	BAggregatorID null.String `json:"biomassAggregator" db:"ba_id"`
}

type editNetworkRequest struct {
	Name          null.String `json:"name" db:"name"`
	LocationName  null.String `json:"locationName" db:"location_name"`
	BAggregatorID null.String `json:"biomassAggregator" db:"ba_id"`
	CSMangerID    null.String `json:"cSinkManager" db:"manager_id"`
}

type networkManagerRequest struct {
	Name        null.String `json:"name" db:"name"`
	PhoneNo     null.String `json:"phoneNo" db:"number"`
	CountryCode null.String `json:"countryCode" db:"country_code"`
	Email       null.String `json:"email" db:"email"`
	Password    null.String `json:"password" db:"password"`
	NetworkID   null.String `json:"networkId" db:"network_id"`
}

type network struct {
	ID                    string      `json:"id" db:"id"`
	Name                  null.String `json:"name" db:"name"`
	LocationName          null.String `json:"locationName" db:"location_name"`
	FarmersCount          int         `json:"farmersCount" db:"farmers_count"`
	ManagerID             null.String `json:"managerId" db:"manager_id"`
	ManagerName           null.String `json:"managerName" db:"manager_name"`
	ManagerEmail          null.String `json:"managerEmail" db:"manager_email"`
	ManagerPhoneNumber    null.String `json:"managerPhoneNumber" db:"manager_number"`
	ManagerCountryCode    null.String `json:"managerCountryCode" db:"manager_country_code"`
	NetworkKilnCount      int         `json:"networkKilnCount" db:"kiln_count"`
	BiomassAggregatorID   null.String `json:"biomassAggregatorId" db:"biomass_aggregator_id"`
	BiomassAggregatorName null.String `json:"biomassAggregatorName" db:"biomass_aggregator_name"`
}

type networkResponse struct {
	Count   int       `json:"count"`
	Limit   int       `json:"limit"`
	Page    int       `json:"page"`
	Network []network `json:"network"`
}

//type networkManager struct {
//	ID    string      `json:"id" db:"id"`
//	Name  null.String `json:"name" db:"name"`
//	Email null.String `json:"email" db:"email"`
//}

//type networkManagerResponse struct {
//	Count          int              `json:"count"`
//	Limit          int              `json:"limit"`
//	Page           int              `json:"page"`
//	NetworkManager []networkManager `json:"network"`
//}

type user struct {
	Name        null.String `json:"-" db:"name"`
	PhoneNo     null.String `json:"-" db:"number"`
	CountryCode null.String `json:"-" db:"country_code"`
	Email       null.String `json:"-" db:"email"`
	Password    null.String `json:"-" db:"password"`
}

type bAFarmer struct {
	ID               string      `json:"id" db:"id"`
	Name             null.String `json:"name" db:"name"`
	Address          null.String `json:"address" db:"address"`
	PhoneNo          null.String `json:"phoneNo" db:"number"`
	CountryCode      null.String `json:"countryCode" db:"country_code"`
	FarmsCount       int         `json:"farmsCount" db:"farms_count"`
	CSinkNetworkID   null.String `json:"cSinkNetworkId,omitempty" db:"network_id"`
	CSinkNetworkName null.String `json:"cSinkNetworkName,omitempty" db:"network_name"`
}

type bAFarmerResponse struct {
	Count   int        `json:"count"`
	Limit   int        `json:"limit"`
	Page    int        `json:"page"`
	Farmers []bAFarmer `json:"farmers"`
}

type kilnRequest struct {
	Name      null.String `json:"name" db:"name"`
	Address   null.String `json:"address" db:"address"`
	NetworkID null.String `json:"networkId" db:"network_id"`
}

type editKilnRequest struct {
	Name    null.String `json:"name" db:"name"`
	Address null.String `json:"address" db:"address"`
}

type kilnOperatorRequest struct {
	UserID null.String `json:"userId" db:"user_id"`
	KilnID null.String `json:"kilnId" db:"kiln_id"`
}

type kiln struct {
	ID              string         `json:"id" db:"id"`
	Name            null.String    `json:"name" db:"name"`
	Address         null.String    `json:"address" db:"address"`
	KilnOperatorIDs pq.StringArray `json:"-" db:"kiln_operator_ids"`
	KilnOperators   []kilnOperator `json:"kilnOperators"`
	NetworkID       null.String    `json:"networkId" db:"network_id"`
	NetworkName     null.String    `json:"networkName" db:"network_name"`
}

type kilnResponse struct {
	Count int    `json:"count"`
	Limit int    `json:"limit"`
	Page  int    `json:"page"`
	Kilns []kiln `json:"kilns"`
}

type kilnOperator struct {
	ID          string      `json:"id" db:"id"`
	Name        null.String `json:"name" db:"name"`
	PhoneNo     null.String `json:"phoneNo" db:"number"`
	CountryCode null.String `json:"countryCode" db:"country_code"`
	KilnID      null.String `json:"-" db:"klin_id"`
}

type kilnOperatorResponse struct {
	Count         int            `json:"count"`
	Limit         int            `json:"limit"`
	Page          int            `json:"page"`
	KilnOperators []kilnOperator `json:"kilnOperators"`
}

type farmerId struct {
	FarmerID string `json:"farmerId"`
}
