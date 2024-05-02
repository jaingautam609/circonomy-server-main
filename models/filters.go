package models

import "github.com/volatiletech/null"

type sortOrder string
type FarmVideoType string
type StageCrop string

const (
	defaultLimit  = 10
	defaultPageNo = 0

	ASC  sortOrder = "ASC"
	DESC sortOrder = "DESC"

	Farming FarmVideoType = "farming"
	BioChar FarmVideoType = "biochar"

	Cropping       StageCrop = "cropping"
	Harvesting     StageCrop = "harvesting"
	SunDrying      StageCrop = "sun_drying"
	Transportation StageCrop = "transportation"

	TransportFarmToKiln StageCrop = "transport_farm_to_kiln"

	Production StageCrop = "production"

	TransportKilnToFarm StageCrop = "transport_kiln_to_farm"

	Distribution StageCrop = "distribution"
)

type GenericFilters struct {
	Limit        int
	Page         int
	SortBy       string
	SortOrder    sortOrder
	SearchString string
}

type FarmingContentFilters struct {
	GenericFilters
	VideoType FarmVideoType
}

type CropsFilters struct {
	GenericFilters
	CropStages []StageCrop
}

type KilnFilters struct {
	GenericFilters
	KilnID null.String
}

type VideoFilters struct {
	GenericFilters
	VideoType null.String
}
