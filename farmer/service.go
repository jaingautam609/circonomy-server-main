package farmer

import (
	"circonomy-server/models"
	"circonomy-server/providers"
	"circonomy-server/utils"
	"database/sql"
	"errors"
	"github.com/lib/pq"
	"github.com/thoas/go-funk"
	"golang.org/x/sync/errgroup"
	"time"
)

type Service struct {
	repository  *Repository
	smsProvider providers.SMSProvider
}

func NewService(jobsRepository *Repository, smsProvider providers.SMSProvider) *Service {
	return &Service{
		repository:  jobsRepository,
		smsProvider: smsProvider,
	}
}

// sendOTP send OTP -> delete previous OTP and send OTP
func (s *Service) sendOTP(otpReq sendOTPRequest) error {
	var dynamicOTP string

	bypassOTPService := utils.IsDevEnvironment() && !funk.ContainsString(utils.BypassDevCheckSMSNumbers(), otpReq.PhoneNumber)
	if bypassOTPService {
		dynamicOTP = "6666"
	} else {
		dynamicOTP = utils.EncodeToString(4)
	}

	otpResponse := models.SendOTP{
		PhoneNumber: otpReq.PhoneNumber,
		CountryCode: otpReq.CountryCode,
		OTP:         dynamicOTP,
		Type:        string(otpTypeNumber),
		Expiry:      time.Now().Add(time.Minute * OTPExpiryMinutes),
	}

	if !bypassOTPService {
		err := s.smsProvider.Send(otpResponse)
		if err != nil {
			return err
		}
	}

	err := s.repository.updateOTP(otpResponse)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) verifyOTPAndGenerateToken(otpReq checkOTPRequest) (*checkOTPResponse, error) {
	err := s.verifyAndUpdateOTP(otpReq)
	if err != nil {
		return nil, err
	}

	return s.generateToken(otpReq.PhoneNumber, otpReq.CountryCode)
}

func (s *Service) generateToken(phoneNumber string, countryCode string) (*checkOTPResponse, error) {
	res := checkOTPResponse{}

	userID, err := s.repository.userPhoneExists(phoneNumber, countryCode)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	res.IsPhoneExists = true
	if errors.Is(err, sql.ErrNoRows) { // user not present
		res.IsPhoneExists = false
		userID, err = s.repository.insertFarmer(phoneNumber, countryCode)
		if err != nil {
			return nil, err
		}
	}

	token, err := utils.GenerateTokenPair(userID)
	if err != nil {
		return nil, err
	}

	res.IsValid = true
	res.Token = token["token"]
	res.RefreshToken = token["refresh_token"]
	return &res, nil
}

func (s *Service) verifyAndUpdateOTP(otpReq checkOTPRequest) error {
	isOTPValid, err := s.repository.verifyOTP(otpReq)
	if err != nil {
		return err
	}
	if !isOTPValid {
		return errors.New("OTP is not valid")
	}

	return s.repository.markVerifiedOTP(otpReq)
}

// updateAccountDetails update farmer details
func (s *Service) updateAccountDetails(farmerID string, farmerDetails updateFarmerDetails) error {
	var status registrationStatus
	farmerInfo, err := s.repository.fetchFarmerDetails(farmerID)
	if err != nil {
		return err
	}
	if farmerInfo.RegistrationStatus == verifyAccount {
		status = createAccount
	} else {
		status = farmerInfo.RegistrationStatus
	}

	erg := new(errgroup.Group)
	erg.Go(func() error {
		err = s.repository.updateAccountDetails(farmerDetails, farmerID)
		if err != nil {
			return err
		}
		return nil
	})
	erg.Go(func() error {
		err = s.repository.updateFarmerRegistrationStatus(farmerID, status)
		if err != nil {
			return err
		}
		return nil
	})
	if err := erg.Wait(); err != nil {
		return err
	}
	return nil
}

// addFarm add farm details
func (s *Service) addFarm(farmerID string, farmDetails addFarmDetails) (string, error) {
	status := addFarm
	croppingPattern := getCroppingPattern(farmDetails.CroppingPattern)
	return s.repository.addFarm(farmDetails, farmerID, status, croppingPattern)
}

// getCroppingPattern get cropping pattern
func getCroppingPattern(croppingPattern croppingPattern) string {
	var cropping string
	switch croppingPattern {
	case monoCropping:
		cropping = "mono_cropping"
	case interCropping:
		cropping = "inter_cropping"
	case cropRotation:
		cropping = "crop_rotation"
	case mixedCropping:
		cropping = "mixed_cropping"
	}
	return cropping
}

// GetCrops get crops
func (s *Service) GetCrops(filters models.GenericFilters) (cropResponse, error) {
	var cropsResponse cropResponse
	erg := new(errgroup.Group)
	erg.Go(func() error {
		var cropsErr error
		cropsResponse.Crops, cropsErr = s.repository.getCrops(filters)
		return cropsErr
	})
	erg.Go(func() error {
		var countErr error
		cropsResponse.Count, countErr = s.repository.fetchCropsCount(filters)
		return countErr
	})
	if err := erg.Wait(); err != nil {
		return cropsResponse, err
	}

	allImageIds := make([]string, 0)
	for _, crop := range cropsResponse.Crops {
		allImageIds = append(allImageIds, crop.ImageID)
	}

	images, err := s.repository.fetchAdminCropImagePath(allImageIds)
	if err != nil {
		return cropsResponse, err
	}

	imageMap := make(map[string]string, 0)
	for _, image := range images {
		imageURl, err := utils.GenerateSignedURL(image.Path)
		if err != nil {
			return cropsResponse, err
		}
		imageMap[image.ID] = imageURl
	}

	for idx := range cropsResponse.Crops {
		cropsResponse.Crops[idx].ImageURL = imageMap[cropsResponse.Crops[idx].ImageID]
	}

	return cropsResponse, nil
}

// fetchFarmerDetails fetch farmer details
func (s *Service) fetchFarmerDetails(farmerID string) (farmerInfo farmerDetails, err error) {

	farmerInfo, err = s.repository.fetchFarmerDetails(farmerID)
	if err != nil {
		return farmerInfo, err
	}

	allImageIds := make([]string, 0)
	if farmerInfo.ProfileImageID.Valid {
		allImageIds = append(allImageIds, farmerInfo.ProfileImageID.String)
	}
	if farmerInfo.AadhaarImageID.Valid {
		allImageIds = append(allImageIds, farmerInfo.AadhaarImageID.String)
	}
	if len(allImageIds) != 0 {
		images, err := s.repository.fetchProfileImagePath(allImageIds)
		if err != nil {
			return farmerInfo, err
		}
		imageMap := make(map[string]string, 0)
		for _, image := range images {
			imageURl, err := utils.GenerateSignedURL(image.Path)
			if err != nil {
				return farmerInfo, err
			}
			imageMap[image.ID] = imageURl
		}
		if farmerInfo.AadhaarImageID.Valid {
			farmerInfo.AadhaarImageURL = &imageURL{
				ID:  farmerInfo.AadhaarImageID.String,
				URL: imageMap[farmerInfo.AadhaarImageID.String],
			}
		}
		if farmerInfo.ProfileImageID.Valid {
			farmerInfo.ProfileImageURL = &imageURL{
				ID:  farmerInfo.ProfileImageID.String,
				URL: imageMap[farmerInfo.ProfileImageID.String],
			}
		}
	}

	return farmerInfo, nil
}

// fetchFarmDetails fetch farm details
func (s *Service) fetchFarmDetails(farmerID string) (farms []farmDetails, err error) {
	farms, err = s.repository.fetchFarmDetails(farmerID)
	if err != nil {
		return farms, err
	}

	allImageIds := make([]string, 0)
	for _, farm := range farms {
		allImageIds = append(allImageIds, farm.FarmImageIDs...)
	}

	images, err := s.repository.fetchImagePath(allImageIds)
	if err != nil {
		return farms, err
	}

	imageMap := make(map[string]string, 0)
	for _, image := range images {
		imageURl, err := utils.GenerateSignedURL(image.Path)
		if err != nil {
			return farms, err
		}
		imageMap[image.ID] = imageURl
	}

	for idx := range farms {
		for _, imageID := range farms[idx].FarmImageIDs {
			farms[idx].FarmImageURLs = append(farms[idx].FarmImageURLs, imageURL{
				ID:  imageID,
				URL: imageMap[imageID],
			})
		}
	}
	return farms, nil
}

// addPreferredCrops add preferred farmer crops
func (s *Service) addPreferredCrops(farmerID string, crops addCropsRequest) error {
	var status registrationStatus
	farmerInfo, err := s.repository.fetchFarmerDetails(farmerID)
	if err != nil {
		return err
	}
	if farmerInfo.RegistrationStatus == createAccount {
		status = addCrop
	} else {
		status = farmerInfo.RegistrationStatus
	}
	return s.repository.addPreferredCrops(farmerID, crops.IDs, status)
}

// deleteFarm delete farm
func (s *Service) deleteFarm(farmID, farmerID string) (int64, error) {
	return s.repository.deleteFarm(farmID, farmerID)
}

// getFarmCrops gets farmer crops filtered for card
func (s *Service) getFarmCrops(farmerID string, filters models.CropsFilters) (farmerCropsResponse, error) {
	cropsResponse := farmerCropsResponse{}
	farmIds, err := s.repository.fetchFarmerFarmIDs(farmerID)
	if err != nil {
		return cropsResponse, err
	}

	var farmIDs pq.StringArray
	for idx := range farmIds {
		farmIDs = append(farmIDs, farmIds[idx])
	}

	cropRes, err := s.repository.fetchFarmCrops(farmIDs, filters)
	if err != nil {
		return cropsResponse, err
	}

	//cropping stage arrays
	var cropIDs pq.StringArray
	for idx := range cropRes {
		cropIDs = append(cropIDs, cropRes[idx].Id)
	}

	cropStagesTimes, err := s.repository.fetchFarmerCropStages(cropIDs)
	if err != nil {
		return cropsResponse, err
	}

	cropStagesMap := make(map[string][]cropStageTime)
	for i := range cropStagesTimes {
		cropStagesMap[cropStagesTimes[i].CropID] = append(cropStagesMap[cropStagesTimes[i].CropID], cropStagesTimes[i])
	}

	for i := range cropRes {
		cropRes[i].Stages = cropStagesMap[cropRes[i].Id]
	}

	// image urls
	allImageIds := make([]string, 0)
	for _, crop := range cropRes {
		allImageIds = append(allImageIds, crop.ImageIDs...)
	}

	images, err := s.repository.fetchImagePath(allImageIds)
	if err != nil {
		return cropsResponse, err
	}

	imageStageMap := make(map[string]models.StageCrop, 0)
	imageMap := make(map[string]string, 0)
	for _, image := range images {
		imageURl, err := utils.GenerateSignedURL(image.Path)
		if err != nil {
			return cropsResponse, err
		}
		imageMap[image.ID] = imageURl
		imageStageMap[image.ID] = image.Stage
	}

	for idx := range cropRes {
		for _, imageID := range cropRes[idx].ImageIDs {
			cropRes[idx].ImageURLs = append(cropRes[idx].ImageURLs, imageURL{
				ID:        imageID,
				URL:       imageMap[imageID],
				CropStage: imageStageMap[imageID],
			})
		}
	}

	total, err := s.repository.fetchFarmerCropsCount(farmIDs, filters)
	if err != nil {
		return cropsResponse, err
	}

	cropsResponse.CropsFiltered = cropRes
	cropsResponse.Count = total
	cropsResponse.Page = filters.Page
	cropsResponse.Limit = filters.Limit
	return cropsResponse, nil
}

// addCrop add crop details
func (s *Service) addFarmCrop(cropDetails cropFormRequest) (string, error) {
	cropID, err := s.repository.addFarmCrop(cropDetails)
	if err != nil {
		return cropID, err
	}
	return cropID, nil
}

// GetFertilizers get fertilizer
func (s *Service) GetFertilizers(filters models.GenericFilters) (fertilizerResponse, error) {
	var fertilizerResponse fertilizerResponse
	erg := new(errgroup.Group)
	erg.Go(func() error {
		var cropsErr error
		fertilizerResponse.Fertilizers, cropsErr = s.repository.getFertilizers(filters)
		return cropsErr
	})
	erg.Go(func() error {
		var countErr error
		fertilizerResponse.Count, countErr = s.repository.fetchFertilizersCount(filters)
		return countErr
	})
	if err := erg.Wait(); err != nil {
		return fertilizerResponse, err
	}

	return fertilizerResponse, nil
}

func (s *Service) GetVideos(filters models.VideoFilters) (VideoResponse, error) {
	var videoResponse VideoResponse
	erg := new(errgroup.Group)
	erg.Go(func() error {
		var cropsErr error
		videoResponse.Videos, cropsErr = s.repository.getVideos(filters)
		return cropsErr
	})
	erg.Go(func() error {
		var countErr error
		videoResponse.Count, countErr = s.repository.fetchVideosCount(filters)
		return countErr
	})
	if err := erg.Wait(); err != nil {
		return videoResponse, err
	}

	for i := range videoResponse.Videos {
		if videoResponse.Videos[i].ThumbnailImagePath.Valid {
			videoResponse.Videos[i].ThumbnailImageURL, _ =
				utils.GenerateSignedURL(videoResponse.Videos[i].ThumbnailImagePath.String)
		}
	}

	return videoResponse, nil
}

// editCropping edit cropping
func (s *Service) editCropping(farmCropId string, cropDetails editCroppingRequest) error {
	return s.repository.txx(func(txRepo *Repository) error {
		return txRepo.editCropping(farmCropId, cropDetails)
	})
}

// editHarvesting edit harvesting
func (s *Service) editHarvesting(farmCropId string, cropDetails editHarvestingRequest) error {
	return s.repository.txx(func(txRepo *Repository) error {
		return txRepo.editHarvesting(farmCropId, cropDetails)
	})
}

// editSundrying edit sun drying
func (s *Service) editSundrying(farmCropId string, cropDetails editSundryingRequest) error {
	return s.repository.txx(func(txRepo *Repository) error {
		return txRepo.editSundrying(farmCropId, cropDetails)
	})
}

// GetFarmCropDetails get Crop details
func (s *Service) GetFarmCropDetails(cropID string) (farmCropDetails, error) {
	farmCropDetail, err := s.repository.getFarmCropDetails(cropID)
	if err != nil {
		return farmCropDetail, err
	}

	// crop image logic
	cropImage, err := s.repository.fetchSingleImagePath(farmCropDetail.CropImageID)
	if err != nil {
		return farmCropDetail, err
	}
	cropImageURl, err := utils.GenerateSignedURL(cropImage.Path)
	if err != nil {
		return farmCropDetail, err
	}
	farmCropDetail.CropImageURL = imageURL{
		ID:  farmCropDetail.CropImageID,
		URL: cropImageURl,
	}

	imageStates, err := s.repository.getCropImageDetails(farmCropDetail.ID)
	if err != nil {
		return farmCropDetail, err
	}
	farmCropDetail.ImageURLs = make([]imageURL, len(imageStates))
	for i, state := range imageStates {
		imageURl, err := utils.GenerateSignedURL(state.Path)
		if err != nil {
			return farmCropDetail, err
		}
		farmCropDetail.ImageURLs[i] = imageURL{
			ID:        state.ID,
			CropStage: state.Stage,
			URL:       imageURl,
		}
	}

	// cropping stage arrays
	cropStagesTimes, err := s.repository.fetchFarmCropStages(farmCropDetail.ID)
	if err != nil {
		return farmCropDetail, err
	}
	for i := range cropStagesTimes {
		farmCropDetail.Stages = append(farmCropDetail.Stages, cropStagesTimes[i])
	}

	// farm crop fertilizers
	fertilizers, err := s.repository.fetchFarmCropFertilizers(farmCropDetail.ID)
	if err != nil {
		return farmCropDetail, err
	}
	for i := range fertilizers {
		farmCropDetail.Fertilizers = append(farmCropDetail.Fertilizers, fertilizers[i])
	}
	return farmCropDetail, nil
}

// cropChangeStatus change crop status
func (s *Service) cropChangeStatus(farmCropId string, cropStatus models.StageCrop) error {
	farmCropDetail, err := s.repository.getFarmCropDetails(farmCropId)
	if err != nil {
		return err
	}
	imageStages, err := s.repository.getImageStates(farmCropDetail.ID)
	if err != nil {
		return err
	}
	changeStatus := checkCropStatus(farmCropDetail, imageStages)
	if changeStatus == true {
		return s.repository.txx(func(txRepo *Repository) error {
			return txRepo.cropChangeStatus(farmCropId, cropStatus)
		})
	}
	return errors.New("can't change status, detail are not complete")
}

// checkCropStatus check crop status
func checkCropStatus(crop farmCropDetails, imageStages cropStages) bool {
	switch crop.CropStage {
	case models.Cropping:
		return checkCroppingDetail(crop, imageStages)
	case models.Harvesting:
		return checkHarvestingDetail(crop, imageStages)
	case models.SunDrying:
		return checkSunDryingDetail(crop, imageStages)
	case models.Transportation:
		return checkTransportationDetail(crop, imageStages)
	case models.TransportFarmToKiln:
		return checkTransportationDetail(crop, imageStages)
	default:
		return false
	}
}

// checkCroppingDetail check cropping detail
func checkCroppingDetail(crop farmCropDetails, imageStages cropStages) bool {
	if !crop.SeedQuantity.Valid {
		return false
	}
	for _, stage := range imageStages.Stages {
		if stage == models.Cropping {
			return true
		}
	}
	return false
}

// checkHarvestingDetail check harvesting detail
func checkHarvestingDetail(crop farmCropDetails, imageStages cropStages) bool {
	if !crop.SeedQuantity.Valid || !crop.YieldQuantity.Valid {
		return false
	}
	for _, stage := range imageStages.Stages {
		if stage == models.Harvesting {
			return true
		}
	}
	return false
}

// checkSunDryingDetail check sun-drying detail
func checkSunDryingDetail(crop farmCropDetails, imageStages cropStages) bool {
	if !crop.SeedQuantity.Valid || !crop.YieldQuantity.Valid {
		return false
	}
	for _, stage := range imageStages.Stages {
		if stage == models.SunDrying {
			return true
		}
	}
	return false
}

// fetchUserPreferredCrops fetch user preferred crops
func (s *Service) fetchUserPreferredCrops(farmerID string) (cropsInfo []preferredCrops, err error) {

	cropsInfo, err = s.repository.fetchUserPreferredCrops(farmerID)
	if err != nil {
		return cropsInfo, err
	}

	// crop image logic
	allImageIds := make([]string, 0)
	for _, crop := range cropsInfo {
		allImageIds = append(allImageIds, crop.CropImageID)
	}

	images, err := s.repository.fetchImagePath(allImageIds)
	if err != nil {
		return cropsInfo, err
	}

	imageMap := make(map[string]string, 0)
	for _, image := range images {
		imageURl, err := utils.GenerateSignedURL(image.Path)
		if err != nil {
			return cropsInfo, err
		}
		imageMap[image.ID] = imageURl
	}

	for idx := range cropsInfo {
		cropsInfo[idx].CropImageURL = imageURL{
			ID:  cropsInfo[idx].CropImageID,
			URL: imageMap[cropsInfo[idx].CropImageID],
		}
	}

	return cropsInfo, nil
}

// editTransportation edit transportation
func (s *Service) editTransportation(farmCropId string, cropDetails editTransportationRequest) error {
	return s.repository.txx(func(txRepo *Repository) error {
		return txRepo.editTransportation(farmCropId, cropDetails)
	})
}

// moveToProductionRequest move to production
func (s *Service) moveToProduction(farmCropId string, crop moveToProductionRequest) error {

	farmCropDetail, err := s.repository.getFarmCropDetails(farmCropId)
	if err != nil {
		return err
	}
	imageStages, err := s.repository.getImageStates(farmCropDetail.ID)
	if err != nil {
		return err
	}

	canChangeStatus := checkCropStatus(farmCropDetail, imageStages)
	if !canChangeStatus {
		return errors.New("can't move to Production status, details are not complete")
	}

	return s.repository.txx(func(txRepo *Repository) error {
		// 1. crop status change
		var err error
		err = txRepo.cropChangeStatus(farmCropId, models.TransportFarmToKiln)
		if err != nil {
			return err
		}

		// 2. which kiln is associated with farm_crop - INSERT kilnID in farm_crop table
		err = txRepo.addMoveToProductionCropDetails(farmCropId, crop)
		if err != nil {
			return err
		}

		// 3. images archive and add image IDs
		err = txRepo.imagesAdd(models.TransportFarmToKiln, farmCropId, crop.ImageIds)
		if err != nil {
			return err
		}

		return nil
	})

}

// checkTransportationDetail check transportation detail
func checkTransportationDetail(crop farmCropDetails, imageStages cropStages) bool {
	if !crop.TransportationVehicle.Valid {
		return false
	}
	for _, stage := range imageStages.Stages {
		if stage == models.Transportation {
			return true
		}
	}
	return false
}

// editDistribution edit distribution
func (s *Service) editDistribution(farmCropId string, cropDetails editDistributionRequest) error {
	return s.repository.txx(func(txRepo *Repository) error {
		err := txRepo.editDistribution(farmCropId, cropDetails)
		if err != nil {
			return err
		}

		err = txRepo.cropChangeStatus(farmCropId, models.Distribution)
		if err != nil {
			return err
		}
		return nil

	})
}
