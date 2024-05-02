package kilnOperator

import (
	"circonomy-server/models"
	"circonomy-server/providers"
	"circonomy-server/utils"
	"database/sql"
	"github.com/pkg/errors"
	"github.com/thoas/go-funk"
	"time"
)

type Service struct {
	repository  *Repository
	smsProvider providers.SMSProvider
}

func NewService(klinRepository *Repository, smsProvider providers.SMSProvider) *Service {
	return &Service{
		repository:  klinRepository,
		smsProvider: smsProvider,
	}
}

// sendOTP send OTP -> delete previous OTP and send OTP
func (s *Service) sendOTP(otpReq sendOTPRequest) error {

	err := s.repository.checkKilnOperator(otpReq)
	if err != nil {
		return errors.New("entered contact details are not of kiln operator")
	}

	var dynamicOTP string
	bypassOTPService := utils.IsDevEnvironment() && !funk.ContainsString(utils.BypassDevCheckSMSNumbers(), otpReq.PhoneNumber)
	if bypassOTPService {
		dynamicOTP = storedOTP
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

	err = s.repository.updateOTP(otpResponse)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) verifyOTPAndGenerateToken(otpReq checkOTPRequest) (*checkOTPResponse, error) {

	check := sendOTPRequest{
		PhoneNumber: otpReq.PhoneNumber,
		CountryCode: otpReq.CountryCode,
	}
	err := s.repository.checkKilnOperator(check)
	if err != nil {
		return nil, errors.New("entered contact details are not of kiln operator")
	}

	err = s.verifyAndUpdateOTP(otpReq)
	if err != nil {
		return nil, err
	}
	return s.generateToken(otpReq.PhoneNumber, otpReq.CountryCode)
}

func (s *Service) generateToken(phoneNumber string, countryCode string) (*checkOTPResponse, error) {
	res := checkOTPResponse{}

	userID, err := s.repository.fetchUserID(phoneNumber, countryCode)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, errors.Wrap(err, "user not found in database")
	}

	token, err := utils.GenerateTokenPairKilnOperator(userID)
	if err != nil {
		return nil, err
	}

	res.IsValid = true
	res.Token = token["token"]
	res.RefreshToken = token["refresh_token"]
	return &res, nil
}

// verifyAndUpdateOTP verify OTP present and mark verify OTP
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

// getFarmCropsBiomasses get farm crops biomasses for card
func (s *Service) getFarmCropsBiomasses(kilnOperatorID string, filters models.KilnFilters) (farmCropBiomassDetails, error) {
	farmCropBiomassResponse := farmCropBiomassDetails{}
	farmCropBiomasses, err := s.repository.getFarmCropsBiomasses(kilnOperatorID, filters)
	if err != nil {
		return farmCropBiomassResponse, err
	}

	// image urls
	allImageIds := make([]string, 0)
	for _, crop := range farmCropBiomasses {
		allImageIds = append(allImageIds, crop.ImageIDs...)
	}

	images, err := s.repository.fetchImagePath(allImageIds)
	if err != nil {
		return farmCropBiomassResponse, err
	}

	imageMap := make(map[string]string, 0)
	for _, image := range images {
		imageURl, err := utils.GenerateSignedURL(image.Path)
		if err != nil {
			return farmCropBiomassResponse, err
		}
		imageMap[image.ID] = imageURl
	}

	for idx := range farmCropBiomasses {
		for _, imageID := range farmCropBiomasses[idx].ImageIDs {
			farmCropBiomasses[idx].ImageURLs = append(farmCropBiomasses[idx].ImageURLs, imageURL{
				ID:  imageID,
				URL: imageMap[imageID],
			})
		}
	}

	total, err := s.repository.getFarmCropsBiomassesCount(kilnOperatorID)
	if err != nil {
		return farmCropBiomassResponse, err
	}

	farmCropBiomassResponse.CropsFiltered = farmCropBiomasses
	farmCropBiomassResponse.Count = total
	farmCropBiomassResponse.Page = filters.Page
	farmCropBiomassResponse.Limit = filters.Limit
	return farmCropBiomassResponse, nil
}

// editFarmCropAndAddKilnBiomass edit farm crop and add kiln biomass
func (s *Service) editFarmCropAndAddKilnBiomass(farmCropId, kilnID string, cropDetails editFarmCropBiomassRequest) error {
	return s.repository.txx(func(txRepo *Repository) error {
		return txRepo.editFarmCropAndAddKilnBiomass(farmCropId, kilnID, cropDetails)
	})
}

// fetchKilnOperatorInfo fetch Kiln Operator Info
func (s *Service) fetchKilnOperatorInfo(kilnOperatorID string) (kilnOperatorInfo kilnOperatorDetails, err error) {
	kilnOperatorInfo, err = s.repository.fetchKilnOperatorInfo(kilnOperatorID)
	if err != nil {
		return kilnOperatorInfo, err
	}

	// image urls
	allImageIds := make([]string, 0)
	if kilnOperatorInfo.ProfileImageID.Valid {
		allImageIds = append(allImageIds, kilnOperatorInfo.ProfileImageID.String)
	}
	if kilnOperatorInfo.AadhaarImageID.Valid {
		allImageIds = append(allImageIds, kilnOperatorInfo.AadhaarImageID.String)
	}
	if len(allImageIds) != 0 {
		images, err := s.repository.fetchImagePath(allImageIds)
		if err != nil {
			return kilnOperatorInfo, err
		}
		imageMap := make(map[string]string, 0)
		for _, image := range images {
			imageURl, err := utils.GenerateSignedURL(image.Path)
			if err != nil {
				return kilnOperatorInfo, err
			}
			imageMap[image.ID] = imageURl
		}
		if kilnOperatorInfo.AadhaarImageID.Valid {
			kilnOperatorInfo.AadhaarImageURL = imageURL{
				ID:  kilnOperatorInfo.AadhaarImageID.String,
				URL: imageMap[kilnOperatorInfo.AadhaarImageID.String],
			}
		}
		if kilnOperatorInfo.ProfileImageID.Valid {
			kilnOperatorInfo.ProfileImageURL = imageURL{
				ID:  kilnOperatorInfo.ProfileImageID.String,
				URL: imageMap[kilnOperatorInfo.ProfileImageID.String],
			}
		}
	}

	// image logic ends
	kilnInfo, err := s.repository.fetchKilnInfo(kilnOperatorInfo.ID)
	if err != nil {
		return kilnOperatorInfo, err
	}
	kilnOperatorInfo.KilnInfo = kilnInfo

	return kilnOperatorInfo, nil
}

// getKilnCropBiomasses get Kiln Crop Biomasses
func (s *Service) getKilnCropBiomasses(kilnId string) (kilnBiomasses []kilnBiomass, err error) {
	kilnBiomasses, err = s.repository.getKilnCropBiomasses(kilnId)
	if err != nil {
		return kilnBiomasses, err
	}
	return kilnBiomasses, nil
}

// getKilnProcessDetails get Kiln Process Details
func (s *Service) getKilnProcessDetails(kilnId string, filters models.GenericFilters) (kilnProcessResponse, error) {
	var kilnProcessesResponse kilnProcessResponse
	var err error
	kilnProcessesResponse, err = s.repository.getKilnProcessDetails(kilnId, filters)
	if err != nil {
		return kilnProcessesResponse, err
	}

	kilnProcessesResponse.Count, err = s.repository.getKilnProcessDetailsCount(kilnId)
	if err != nil {
		return kilnProcessesResponse, err
	}

	// image logic
	allImageIds := make([]string, 0)
	for idx := range kilnProcessesResponse.KilnProcesses {
		for _, kilnP := range kilnProcessesResponse.KilnProcesses[idx].ImageIDs {
			allImageIds = append(allImageIds, kilnP)
		}
	}
	images, err := s.repository.fetchImagePath(allImageIds)
	if err != nil {
		return kilnProcessesResponse, err
	}

	allProcessIds := make([]string, len(kilnProcessesResponse.KilnProcesses))
	for id, detail := range kilnProcessesResponse.KilnProcesses {
		allProcessIds[id] = detail.Id
	}
	videos, err := s.repository.fetchKlinVideoPath(allProcessIds)
	if err != nil {
		return kilnProcessesResponse, err
	}

	videoProcessMap := make(map[string][]imageURL)
	for i := range videos {
		if _, ok := videoProcessMap[videos[i].KlinProcessId]; !ok {
			videoProcessMap[videos[i].KlinProcessId] = make([]imageURL, 0)
		}
		url, err := utils.GenerateSignedURL(videos[i].Path)
		if err != nil {
			return kilnProcessesResponse, err
		}
		videoProcessMap[videos[i].KlinProcessId] = append(videoProcessMap[videos[i].KlinProcessId], imageURL{
			ID:  videos[i].ID,
			URL: url,
		})
	}

	imageMap := make(map[string]string, 0)
	for _, image := range images {
		imageURl, err := utils.GenerateSignedURL(image.Path)
		if err != nil {
			return kilnProcessesResponse, err
		}
		imageMap[image.ID] = imageURl
	}

	for idx := range kilnProcessesResponse.KilnProcesses {
		for _, imageID := range kilnProcessesResponse.KilnProcesses[idx].ImageIDs {
			kilnProcessesResponse.KilnProcesses[idx].ImageURLs = append(kilnProcessesResponse.KilnProcesses[idx].ImageURLs, imageURL{
				ID:  imageID,
				URL: imageMap[imageID],
			})
		}
		kilnProcessesResponse.KilnProcesses[idx].VideoURLs = videoProcessMap[kilnProcessesResponse.KilnProcesses[idx].Id]
	}

	return kilnProcessesResponse, nil
}

// addKilnProcessBatch add kiln process batch
func (s *Service) addKilnProcessBatch(kilnId string, request kilnProcessRequest) error {
	return s.repository.txx(func(txRepo *Repository) error {
		return txRepo.addKilnProcessBatch(kilnId, request)
	})
}

// editKilnProcessDetails edit kiln process details
func (s *Service) editKilnProcessDetails(kilnId string, request kilnProcessEditRequest) error {
	return s.repository.txx(func(txRepo *Repository) error {
		return txRepo.editKilnProcessDetails(kilnId, request)
	})
}

// doneKilnProcess done kiln process
func (s *Service) doneKilnProcess(kilnId string, request kilnProcessDoneRequest) error {
	return s.repository.txx(func(txRepo *Repository) error {
		return txRepo.doneKilnProcess(kilnId, request)
	})
}

// fetchKilnBiochar fetch kiln biochar
func (s *Service) fetchKilnBiochar(kilnId string) (kilnBioChar, error) {
	return s.repository.fetchKilnBiochar(kilnId)
}

// addFarmCropBioCharDetails add farm crop biochar details
func (s *Service) addFarmCropBioCharDetails(farmCropId string, request kilnDistributionRequest) error {
	return s.repository.txx(func(txRepo *Repository) error {
		return txRepo.addFarmCropBioCharDetails(farmCropId, request)
	})
}

// fetchDistributedFarmCrops fetch distributed farm crops
func (s *Service) fetchDistributedFarmCrops(kilnID string, filters models.GenericFilters) (distributedFarmCropsResponse, error) {
	var cropsResponse distributedFarmCropsResponse
	distributedFarmCrop, err := s.repository.fetchDistributedFarmCrops(kilnID, filters)
	if err != nil {
		return cropsResponse, err
	}

	cropsResponse.Count, err = s.repository.fetchDistributedFarmCropsCount(kilnID)
	if err != nil {
		return cropsResponse, err
	}

	allImageIds := make([]string, 0)
	for _, crop := range distributedFarmCrop {
		allImageIds = append(allImageIds, crop.ImageIDs...)
	}

	images, err := s.repository.fetchImagePath(allImageIds)
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

	for idx := range distributedFarmCrop {
		for _, imageID := range distributedFarmCrop[idx].ImageIDs {
			distributedFarmCrop[idx].ImageURLs = append(distributedFarmCrop[idx].ImageURLs, imageURL{
				ID:  imageID,
				URL: imageMap[imageID],
			})
		}
	}

	cropsResponse.DistributedFarmCrop = distributedFarmCrop
	cropsResponse.Page = filters.Page
	cropsResponse.Limit = filters.Limit
	return cropsResponse, nil
}
