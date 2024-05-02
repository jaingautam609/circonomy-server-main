package admin

import (
	"circonomy-server/models"
	"circonomy-server/utils"
	"golang.org/x/sync/errgroup"
)

type Service struct {
	repository *Repository
}

func NewService(jobsRepository *Repository) *Service {
	return &Service{
		repository: jobsRepository,
	}
}

// createCrop ADMIN add crop
func (s *Service) createCrop(crop AddCropRequest) (string, error) {
	return s.repository.createCrop(crop)
}

// deleteCrop ADMIN delete crop
func (s *Service) deleteCrop(cropID string) error {
	return s.repository.deleteCrop(cropID)
}

func (s *Service) createVideo(videoRequest AddVideoRequest) (string, error) {
	return s.repository.createVideo(videoRequest)
}

func (s *Service) deleteVideo(videoID string) error {
	return s.repository.deleteVideo(videoID)
}

// createFertilizer ADMIN add crop
func (s *Service) createFertilizer(fertilizer AddFertilizerRequest) (string, error) {
	fertilizerID, err := s.repository.createFertilizer(fertilizer)
	if err != nil {
		return fertilizerID, err
	}
	return fertilizerID, nil
}

// deleteFertilizer ADMIN delete crop
func (s *Service) deleteFertilizer(fertilizerID string) error {
	err := s.repository.deleteFertilizer(fertilizerID)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) login(credentials models.LoginCredentials) (*loginResponse, error) {
	userId, err := s.verifyLoginCredentials(credentials)
	if err != nil {
		return nil, err
	}
	return s.generateToken(userId)
}

func (s *Service) verifyLoginCredentials(credentials models.LoginCredentials) (string, error) {
	userID, storedPassword, err := s.repository.
		getUserPasswordByEmailAndAccountType(credentials.Email, models.UserAccountTypeAdmin)
	if err != nil {
		return userID, err
	}
	checkErr := utils.CheckPassword(credentials.Password, storedPassword)
	return userID, checkErr
}

func (s *Service) generateToken(userID string) (*loginResponse, error) {
	res := loginResponse{}

	token, err := utils.GenerateTokenPair(userID)
	if err != nil {
		return nil, err
	}

	res.IsValid = true
	res.Token = token["token"]
	res.RefreshToken = token["refresh_token"]
	return &res, nil
}

// createBiomassAggregator ADMIN create biomass aggregator
func (s *Service) createBiomassAggregator(biomassAggregator biomassAggregatorRequest) error {
	return s.repository.txx(func(txRepo *Repository) error {

		user := user{
			Name:        biomassAggregator.Name,
			PhoneNo:     biomassAggregator.Manager.PhoneNo,
			CountryCode: biomassAggregator.Manager.CountryCode,
			Email:       biomassAggregator.Manager.Email,
			Password:    biomassAggregator.Manager.Password,
		}
		userID, err := txRepo.createUser(user, models.UserAccountTypeFarmer)
		if err != nil {
			return err
		}
		biomassAggregatorID, err := txRepo.createBiomassAggregator(biomassAggregator)
		if err != nil {
			return err
		}
		err = txRepo.createBiomassAggregatorManager(biomassAggregatorID, userID)
		if err != nil {
			return err
		}
		return nil
	})
}

// getBiomassAggregator get biomass aggregator
func (s *Service) getBiomassAggregator(filters models.GenericFilters) (biomassAggregatorResponse, error) {
	var biomassAggregatorRes biomassAggregatorResponse
	erg := new(errgroup.Group)
	erg.Go(func() error {
		var cropsErr error
		biomassAggregatorRes.BiomassAggregators, cropsErr = s.repository.getBiomassAggregator(filters)
		return cropsErr
	})
	erg.Go(func() error {
		var countErr error
		biomassAggregatorRes.Count, countErr = s.repository.getBiomassAggregatorCount(filters)
		return countErr
	})
	if err := erg.Wait(); err != nil {
		return biomassAggregatorRes, err
	}
	biomassAggregatorRes.Page = filters.Page
	biomassAggregatorRes.Limit = filters.Limit
	return biomassAggregatorRes, nil
}

// getBiomassAggregatorById get biomass aggregator by id
func (s *Service) getBiomassAggregatorById(biomassAggregatorUrlID string) (biomassAggregator biomassAggregatorDetailsResponse, err error) {
	biomassAggregator, err = s.repository.getBiomassAggregatorById(biomassAggregatorUrlID)
	if err != nil {
		return biomassAggregator, err
	}
	return biomassAggregator, nil
}

// editBiomassAggregator edit biomass aggregator by id
func (s *Service) editBiomassAggregator(biomassAggregator biomassAggregatorDetailsRequest, biomassAggregatorUrlID string) error {
	return s.repository.txx(func(txRepo *Repository) error {
		return txRepo.editBiomassAggregator(biomassAggregatorUrlID, biomassAggregator)
	})
}

// deleteBiomassAggregator delete biomass aggregator by id
func (s *Service) deleteBiomassAggregator(biomassAggregatorUrlID string) error {
	return s.repository.txx(func(txRepo *Repository) error {
		return txRepo.deleteBiomassAggregator(biomassAggregatorUrlID)
	})
}

// getBANetworkList get BA network list
func (s *Service) getBANetworkList(filters models.GenericFilters, biomassAggregatorID string) (bANetworkResponse, error) {
	var bANetworkRes bANetworkResponse
	erg := new(errgroup.Group)
	erg.Go(func() error {
		var cropsErr error
		bANetworkRes.BiomassAggregators, cropsErr = s.repository.getBANetworkList(filters, biomassAggregatorID)
		return cropsErr
	})
	erg.Go(func() error {
		var countErr error
		bANetworkRes.Count, countErr = s.repository.getBANetworkListCount(filters, biomassAggregatorID)
		return countErr
	})
	if err := erg.Wait(); err != nil {
		return bANetworkRes, err
	}
	bANetworkRes.Page = filters.Page
	bANetworkRes.Limit = filters.Limit
	return bANetworkRes, nil
}

// createCSNetwork ADMIN create C S network
func (s *Service) createCSNetwork(cSNetwork networkRequest) error {
	return s.repository.txx(func(txRepo *Repository) error {
		return txRepo.createCSNetwork(cSNetwork)
	})
}

// getCSNetwork get C S network
func (s *Service) getCSNetwork(filters models.GenericFilters) (networkResponse, error) {
	var res networkResponse
	erg := new(errgroup.Group)
	erg.Go(func() error {
		var cropsErr error
		res.Network, cropsErr = s.repository.getCSNetwork(filters)
		return cropsErr
	})
	erg.Go(func() error {
		var countErr error
		res.Count, countErr = s.repository.getCSNetworkCount(filters)
		return countErr
	})
	if err := erg.Wait(); err != nil {
		return res, err
	}
	res.Page = filters.Page
	res.Limit = filters.Limit
	return res, nil
}

// editCSNetwork edit C S network by id
func (s *Service) editCSNetwork(network editNetworkRequest, id string) error {
	return s.repository.txx(func(txRepo *Repository) error {
		return txRepo.editCSNetwork(id, network)
	})
}

// deleteCSNetwork deleteCSNetwork by id
func (s *Service) deleteCSNetwork(id string) error {
	return s.repository.txx(func(txRepo *Repository) error {
		return txRepo.deleteCSNetwork(id)
	})
}

// createCSNetworkManager ADMIN createCSNetworkManager
func (s *Service) createCSNetworkManager(networkManager networkManagerRequest) error {
	return s.repository.txx(func(txRepo *Repository) error {

		user := user{
			Name:        networkManager.Name,
			PhoneNo:     networkManager.PhoneNo,
			CountryCode: networkManager.CountryCode,
			Email:       networkManager.Email,
			Password:    networkManager.Password,
		}
		userID, err := txRepo.createUser(user, models.UserAccountTypeCSinkManager)
		if err != nil {
			return err
		}
		return txRepo.createCSNetworkManager(networkManager.NetworkID.String, userID)
	})
}

// getBAApprovedFarmerList get B A Approved Farmer List
func (s *Service) getBAApprovedFarmerList(filters models.GenericFilters, biomassAggregatorID string) (bAFarmerResponse, error) {
	var bAFarmerRes bAFarmerResponse
	erg := new(errgroup.Group)
	erg.Go(func() error {
		var cropsErr error
		bAFarmerRes.Farmers, cropsErr = s.repository.getBAApprovedFarmerList(filters, biomassAggregatorID)
		return cropsErr
	})
	erg.Go(func() error {
		var countErr error
		bAFarmerRes.Count, countErr = s.repository.getBAApprovedFarmerListCount(filters, biomassAggregatorID)
		return countErr
	})
	if err := erg.Wait(); err != nil {
		return bAFarmerRes, err
	}
	bAFarmerRes.Page = filters.Page
	bAFarmerRes.Limit = filters.Limit
	return bAFarmerRes, nil
}

// getBAPendingFarmerList get B A Pending Farmer List
func (s *Service) getBAPendingFarmerList(filters models.GenericFilters, id string) (bAFarmerResponse, error) {
	var bANetworkRes bAFarmerResponse
	erg := new(errgroup.Group)
	erg.Go(func() error {
		var cropsErr error
		bANetworkRes.Farmers, cropsErr = s.repository.getBAPendingFarmerList(filters, id)
		return cropsErr
	})
	erg.Go(func() error {
		var countErr error
		bANetworkRes.Count, countErr = s.repository.getBAPendingFarmerListCount(filters, id)
		return countErr
	})
	if err := erg.Wait(); err != nil {
		return bANetworkRes, err
	}
	bANetworkRes.Page = filters.Page
	bANetworkRes.Limit = filters.Limit
	return bANetworkRes, nil
}

// createKiln create kiln
func (s *Service) createKiln(kiln kilnRequest) error {
	return s.repository.txx(func(txRepo *Repository) error {
		return txRepo.createKiln(kiln)
	})
}

// getKiln get Kiln
func (s *Service) getKiln(filters models.GenericFilters) (kilnResponse, error) {
	var res kilnResponse
	erg := new(errgroup.Group)
	erg.Go(func() error {
		var cropsErr error
		res.Kilns, cropsErr = s.repository.getKiln(filters)
		return cropsErr
	})
	erg.Go(func() error {
		var countErr error
		res.Count, countErr = s.repository.getKilnCount(filters)
		return countErr
	})
	if err := erg.Wait(); err != nil {
		return res, err
	}

	// kiln operators info from kiln operator ids
	allKilnOperatorIds := make([]string, 0)
	for _, kiln := range res.Kilns {
		for _, kOpId := range kiln.KilnOperatorIDs {
			allKilnOperatorIds = append(allKilnOperatorIds, kOpId)
		}
	}

	kilnOperators, err := s.repository.fetchKilnOperators(allKilnOperatorIds)
	if err != nil {
		return res, err
	}

	kilnOperatorMap := make(map[string]kilnOperator, 0)
	for _, kOperator := range kilnOperators {
		kilnOperatorMap[kOperator.ID] = kOperator
	}

	for idx := range res.Kilns {
		for _, kilnOperatorID := range res.Kilns[idx].KilnOperatorIDs {
			res.Kilns[idx].KilnOperators = append(res.Kilns[idx].KilnOperators, kilnOperatorMap[kilnOperatorID])
		}
	}
	//

	res.Page = filters.Page
	res.Limit = filters.Limit
	return res, nil
}

// editKiln edit Kiln by id
func (s *Service) editKiln(kiln editKilnRequest, id string) error {
	return s.repository.txx(func(txRepo *Repository) error {
		return txRepo.editKiln(id, kiln)
	})
}

// deleteKiln delete kiln by id
func (s *Service) deleteKiln(id string) error {
	return s.repository.txx(func(txRepo *Repository) error {
		return txRepo.deleteKiln(id)
	})
}

// createKilnOperator create kiln operator
func (s *Service) createKilnOperator(kilnOperator kilnOperatorRequest) error {
	return s.repository.txx(func(txRepo *Repository) error {
		return txRepo.createKilnOperator(kilnOperator)
	})
}

// assigningFarmerToCSNetwork assigning Farmer To C S Network
func (s *Service) assigningFarmerToCSNetwork(farmer farmerId, id string) error {
	return s.repository.txx(func(txRepo *Repository) error {
		return txRepo.assigningFarmerToCSNetwork(farmer, id)
	})
}

// getBARejectedFarmerList get B A Rejected Farmer List
func (s *Service) getBARejectedFarmerList(filters models.GenericFilters, id string) (bAFarmerResponse, error) {
	var bARejectedFarmerRes bAFarmerResponse
	erg := new(errgroup.Group)
	erg.Go(func() error {
		var cropsErr error
		bARejectedFarmerRes.Farmers, cropsErr = s.repository.getBARejectedFarmerList(filters, id)
		return cropsErr
	})
	erg.Go(func() error {
		var countErr error
		bARejectedFarmerRes.Count, countErr = s.repository.getBARejectedFarmerListCount(filters, id)
		return countErr
	})
	if err := erg.Wait(); err != nil {
		return bARejectedFarmerRes, err
	}
	bARejectedFarmerRes.Page = filters.Page
	bARejectedFarmerRes.Limit = filters.Limit
	return bARejectedFarmerRes, nil
}

// getCSNKilnList get C S N by id Kiln List
func (s *Service) getCSNKilnList(filters models.GenericFilters, csnID string) (kilnResponse, error) {
	var res kilnResponse
	erg := new(errgroup.Group)
	erg.Go(func() error {
		var cropsErr error
		res.Kilns, cropsErr = s.repository.getCSNKilnList(filters, csnID)
		return cropsErr
	})
	erg.Go(func() error {
		var countErr error
		res.Count, countErr = s.repository.getCSNKilnListCount(filters, csnID)
		return countErr
	})
	if err := erg.Wait(); err != nil {
		return res, err
	}

	// kiln operators info from kiln operator ids
	allKilnOperatorIds := make([]string, 0)
	for _, kiln := range res.Kilns {
		for _, kOpId := range kiln.KilnOperatorIDs {
			allKilnOperatorIds = append(allKilnOperatorIds, kOpId)
		}
	}

	kilnOperators, err := s.repository.fetchKilnOperators(allKilnOperatorIds)
	if err != nil {
		return res, err
	}

	kilnOperatorMap := make(map[string]kilnOperator, 0)
	for _, kOperator := range kilnOperators {
		kilnOperatorMap[kOperator.ID] = kOperator
	}

	for idx := range res.Kilns {
		for _, kilnOperatorID := range res.Kilns[idx].KilnOperatorIDs {
			res.Kilns[idx].KilnOperators = append(res.Kilns[idx].KilnOperators, kilnOperatorMap[kilnOperatorID])
		}
	}
	//

	res.Page = filters.Page
	res.Limit = filters.Limit
	return res, nil
}

// getCSNFarmerList get C S N by id Farmer List Count
func (s *Service) getCSNFarmerList(filters models.GenericFilters, csnID string) (bAFarmerResponse, error) {
	var cSNFarmerRes bAFarmerResponse
	erg := new(errgroup.Group)
	erg.Go(func() error {
		var cropsErr error
		cSNFarmerRes.Farmers, cropsErr = s.repository.getCSNFarmerList(filters, csnID)
		return cropsErr
	})
	erg.Go(func() error {
		var countErr error
		cSNFarmerRes.Count, countErr = s.repository.getCSNFarmerListCount(filters, csnID)
		return countErr
	})
	if err := erg.Wait(); err != nil {
		return cSNFarmerRes, err
	}
	cSNFarmerRes.Page = filters.Page
	cSNFarmerRes.Limit = filters.Limit
	return cSNFarmerRes, nil
}

// bARejectFarmer B A reject farmer
func (s *Service) bARejectFarmer(farmer farmerId, id string) error {
	return s.repository.txx(func(txRepo *Repository) error {
		return txRepo.bARejectFarmer(farmer, id)
	})
}

// getCSNetworkDetailsByID get C S Network Details by id
func (s *Service) getCSNetworkDetailsByID(id string) (network, error) {
	return s.repository.getCSNetworkDetailsByID(id)
}
