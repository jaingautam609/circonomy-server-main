package utils

import (
	"circonomy-server/models"
	"github.com/volatiletech/null"
	"net/url"
	"strconv"
	"strings"
)

func NewFilters(queryParams url.Values) models.GenericFilters {
	filter := models.GenericFilters{}

	strLimit := queryParams.Get("limit")
	if strLimit == "" {
		filter.Limit = defaultLimit
	} else {
		limit, err := strconv.Atoi(strLimit)
		if err != nil {
			return filter
		}
		filter.Limit = limit
	}

	strPage := queryParams.Get("page")
	if strPage == "" {
		filter.Page = defaultPageNo
	} else {
		pageNo, err := strconv.Atoi(strPage)
		if err != nil {
			return filter
		}
		filter.Page = pageNo
	}

	strString := queryParams.Get("search")
	if strString != "" {
		filter.SearchString = strString
	}

	sortingOrder := queryParams.Get("sortOrder")
	switch sortingOrder {
	case string(models.ASC):
		filter.SortOrder = models.ASC
	default:
		filter.SortOrder = models.DESC
	}

	return filter
}

func NewFarmingContentFilters(queryParams url.Values) models.FarmingContentFilters {
	var filters models.FarmingContentFilters
	genericFilters := NewFilters(queryParams)
	filters.GenericFilters = genericFilters

	videoType := queryParams.Get("videoType")
	filters.VideoType = models.Farming
	if videoType == string(models.BioChar) {
		filters.VideoType = models.BioChar
	}

	return filters
}

func CropStageFilters(queryParams url.Values) models.CropsFilters {
	var filters models.CropsFilters
	genericFilters := NewFilters(queryParams)
	filters.GenericFilters = genericFilters

	if len(queryParams.Get("cropStage")) > 0 {
		stages := strings.Split(queryParams.Get("cropStage"), ",")
		filters.CropStages = make([]models.StageCrop, len(stages))
		for i := range stages {
			filters.CropStages[i] = models.StageCrop(stages[i])
		}
	}

	return filters
}

func KilnFilter(queryParams url.Values) models.KilnFilters {
	var filters models.KilnFilters
	genericFilters := NewFilters(queryParams)
	filters.GenericFilters = genericFilters

	return filters
}

func VideoFilter(queryParams url.Values) models.VideoFilters {
	var filters models.VideoFilters
	genericFilters := NewFilters(queryParams)
	filters.GenericFilters = genericFilters
	if len(queryParams.Get("videoType")) > 0 {
		filters.VideoType = null.StringFrom(queryParams.Get("videoType"))
	}
	return filters
}
