package admin

import (
	"circonomy-server/dbhelpers"
	"circonomy-server/farmer"
	"circonomy-server/models"
	"circonomy-server/utils"
	"database/sql"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"mime/multipart"
	"net/http"
)

const MaxMemorySize int64 = 32
const MaxMemoryLimit int64 = 20

type Handler struct {
	service       *Service
	farmerService *farmer.Service
}

func NewHandler(service *Service, farmerService *farmer.Service) *Handler {
	return &Handler{
		service:       service,
		farmerService: farmerService,
	}
}

func (h *Handler) Serve(adminRouter chi.Router) {
	adminRouter.Post("/login", h.login)
	adminRouter.Post("/refresh-auth-token", h.refreshAuthToken)

	adminRouter.Post("/upload", h.uploadFile)

	adminRouter.Route("/crops", func(cropsRouter chi.Router) {
		cropsRouter.Post("/", h.createCrop)
		cropsRouter.Get("/", h.getCrops)              // admin API
		cropsRouter.Delete("/{cropId}", h.deleteCrop) // admin API
	})

	adminRouter.Route("/video", func(cropsRouter chi.Router) {
		cropsRouter.Post("/", h.createVideo)
		cropsRouter.Get("/", h.getVideos)               // admin API
		cropsRouter.Delete("/{videoId}", h.deleteVideo) // admin API
	})

	adminRouter.Route("/fertilizers", func(fertilizerRouter chi.Router) {
		fertilizerRouter.Post("/", h.createFertilizer)                 // admin API
		fertilizerRouter.Get("/", h.getFertilizers)                    // admin API
		fertilizerRouter.Delete("/{fertilizerId}", h.deleteFertilizer) // admin API
	})

	adminRouter.Route("/biomass-aggregator", func(biomassAggregatorRouter chi.Router) {
		biomassAggregatorRouter.Post("/", h.createBiomassAggregator)
		biomassAggregatorRouter.Get("/", h.getBiomassAggregator)

		biomassAggregatorRouter.Get("/{biomassAggregatorId}", h.getBiomassAggregatorById)
		biomassAggregatorRouter.Put("/{biomassAggregatorId}", h.editBiomassAggregator)

		// check with UI to not use it for now
		biomassAggregatorRouter.Delete("/{biomassAggregatorId}", h.deleteBiomassAggregator)

		biomassAggregatorRouter.Get("/{biomassAggregatorId}/network", h.getBANetworkList)
		biomassAggregatorRouter.Get("/{biomassAggregatorId}/farmer-approved", h.getBAApprovedFarmerList)

		biomassAggregatorRouter.Get("/{biomassAggregatorId}/farmer-pending", h.getBAPendingFarmerList)

		biomassAggregatorRouter.Get("/{biomassAggregatorId}/farmer-rejected", h.getBARejectedFarmerList)
		biomassAggregatorRouter.Put("/{biomassAggregatorId}/farmer-reject", h.bARejectFarmer)

	})

	adminRouter.Route("/cs-network", func(csNetwork chi.Router) {
		csNetwork.Get("/", h.getCSNetwork)
		csNetwork.Post("/", h.createCSNetwork)
		csNetwork.Get("/{cs-networkId}", h.getCSNetworkDetailsByID)
		csNetwork.Put("/{cs-networkId}", h.editCSNetwork)
		csNetwork.Delete("/{cs-networkId}", h.deleteCSNetwork)

		csNetwork.Put("/{cs-networkId}/assigning-farmer", h.assigningFarmerToCSNetwork)

		csNetwork.Get("/{cs-networkId}/kilns", h.getCSNKilnList)
		csNetwork.Get("/{cs-networkId}/farmers", h.getCSNFarmerList)

	})

	adminRouter.Route("/cs-network-manager", func(csNetworkManager chi.Router) {
		csNetworkManager.Post("/", h.createCSNetworkManager)
	})

	adminRouter.Route("/kiln", func(kiln chi.Router) {
		kiln.Get("/", h.getKiln)
		kiln.Post("/", h.createKiln)
		kiln.Put("/{kilnId}", h.editKiln)
		kiln.Delete("/{kilnId}", h.deleteKiln)

	})

	adminRouter.Route("/kiln-operator", func(kilnOperator chi.Router) {
		kilnOperator.Post("/", h.createKilnOperator)
	})

}

// getCrops get crops
func (h *Handler) getCrops(resp http.ResponseWriter, request *http.Request) {
	filters := utils.NewFilters(request.URL.Query())
	cropsResponse, err := h.farmerService.GetCrops(filters)
	if err != nil {
		utils.RespondError(resp, http.StatusBadRequest, err, "failed to fetch crops")
		return
	}
	utils.RespondJSON(resp, http.StatusOK, cropsResponse)
}

// getCrops get crops
func (h *Handler) getVideo(resp http.ResponseWriter, request *http.Request) {
	filters := utils.NewFilters(request.URL.Query())
	cropsResponse, err := h.farmerService.GetCrops(filters)
	if err != nil {
		utils.RespondError(resp, http.StatusBadRequest, err, "failed to fetch crops")
		return
	}
	utils.RespondJSON(resp, http.StatusOK, cropsResponse)
}

// uploadFile upload image
func (h *Handler) uploadFile(w http.ResponseWriter, r *http.Request) {
	// Parse the multipart form, up to 32 MiB in memory
	if err := r.ParseMultipartForm(MaxMemorySize << MaxMemoryLimit); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "failed to parse multipartForm")
		return
	}

	// Get the file from the form
	file, handler, err := r.FormFile("file")
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "failed to get file from request")
		return
	}
	defer func(file multipart.File) {
		defErr := file.Close()
		if defErr != nil {
			return
		}
	}(file)
	fileType := r.FormValue("type")
	err = utils.UploadImageToBucket(file, handler)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "failed to upload file to bucket")
	}

	ID, err := dbhelpers.AddImage(handler.Filename, fileType)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "failed to add file to db")
		return
	}
	url, urlErr := utils.GenerateSignedURL(handler.Filename)
	if urlErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, urlErr, "failed to get file url")
		return
	}
	// Return a success response
	utils.RespondJSON(w, http.StatusOK, struct {
		ID  uuid.UUID `json:"id"`
		URL string    `json:"url"`
	}{
		ID:  ID,
		URL: url,
	})
}

// createCrop ADMIN add crop
func (h *Handler) createCrop(w http.ResponseWriter, r *http.Request) {
	var cropInfo AddCropRequest
	err := json.NewDecoder(r.Body).Decode(&cropInfo)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, string(parseError))
		return
	}
	cropID, err := h.service.createCrop(cropInfo)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "error adding admin crop")
		return
	}

	utils.RespondJSON(w, http.StatusCreated, addCropResponse{
		ID: cropID,
	})
}

// deleteCrop ADMIN delete crop
func (h *Handler) deleteCrop(w http.ResponseWriter, r *http.Request) {
	cropUrlID := chi.URLParam(r, "cropId")
	err := h.service.deleteCrop(cropUrlID)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "failed to delete admin crop")
		return
	}
	utils.RespondJSON(w, http.StatusOK, messageResponse{
		Message: "crop deleted",
	})
}

func (h *Handler) createVideo(w http.ResponseWriter, r *http.Request) {
	var videoInfo AddVideoRequest
	err := json.NewDecoder(r.Body).Decode(&videoInfo)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, string(parseError))
		return
	}
	cropID, err := h.service.createVideo(videoInfo)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "error adding admin video")
		return
	}

	utils.RespondJSON(w, http.StatusOK, addVideoResponse{
		ID: cropID,
	})
}

func (h *Handler) deleteVideo(w http.ResponseWriter, r *http.Request) {
	cropUrlID := chi.URLParam(r, "videoId")
	err := h.service.deleteVideo(cropUrlID)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "failed to delete admin video")
		return
	}
	utils.RespondJSON(w, http.StatusOK, messageResponse{
		Message: "video deleted",
	})
}

// createFertilizer ADMIN add fertilizer
func (h *Handler) createFertilizer(w http.ResponseWriter, r *http.Request) {
	var fertilizerInfo AddFertilizerRequest
	err := json.NewDecoder(r.Body).Decode(&fertilizerInfo)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, string(parseError))
		return
	}
	cropID, err := h.service.createFertilizer(fertilizerInfo)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "error adding admin fertilizer")
		return
	}

	utils.RespondJSON(w, http.StatusOK, addCropResponse{
		ID: cropID,
	})
}

// getFertilizers get fertilizer
func (h *Handler) getFertilizers(resp http.ResponseWriter, request *http.Request) {
	filters := utils.NewFilters(request.URL.Query())
	fertilizersResponse, err := h.farmerService.GetFertilizers(filters)
	if err != nil {
		utils.RespondError(resp, http.StatusBadRequest, err, "failed to fetch fertilizers")
		return
	}
	utils.RespondJSON(resp, http.StatusOK, fertilizersResponse)
}

// deleteFertilizer ADMIN delete fertilizer
func (h *Handler) deleteFertilizer(w http.ResponseWriter, r *http.Request) {
	fertilizerUrlID := chi.URLParam(r, "fertilizerId")
	err := h.service.deleteFertilizer(fertilizerUrlID)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "failed to delete admin fertilizer")
		return
	}
	utils.RespondJSON(w, http.StatusOK, messageResponse{
		Message: "fertilizer deleted",
	})
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var credentials models.LoginCredentials
	if err := utils.ParseBody(r.Body, &credentials); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "Failed to parse login credentials")
		return
	}

	loginResponse, err := h.service.login(credentials)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "Wrong email and password")
		return
	}
	utils.RespondJSON(w, http.StatusOK, loginResponse)
}

// refreshAuthToken refresh auth token
func (h *Handler) refreshAuthToken(resp http.ResponseWriter, request *http.Request) {
	tokenReq := models.RefreshTokenRequest{}
	err := json.NewDecoder(request.Body).Decode(&tokenReq)
	if err != nil {
		utils.RespondError(resp, http.StatusBadRequest, err, "failed to parse the request")
		return
	}

	token, err := utils.RefreshAuthToken(tokenReq)
	if err != nil {
		utils.RespondError(resp, http.StatusUnauthorized, err, "Error generating access token")
		return
	}

	utils.RespondJSON(resp, http.StatusOK, token)
}

func (h *Handler) getVideos(resp http.ResponseWriter, request *http.Request) {
	filters := utils.VideoFilter(request.URL.Query())
	videoResponse, err := h.farmerService.GetVideos(filters)
	if err != nil {
		utils.RespondError(resp, http.StatusBadRequest, err, "failed to fetch videos")
		return
	}
	utils.RespondJSON(resp, http.StatusOK, videoResponse)
}

// createBiomassAggregator ADMIN create biomass aggregator
func (h *Handler) createBiomassAggregator(w http.ResponseWriter, r *http.Request) {
	var biomassAggregator biomassAggregatorRequest
	err := json.NewDecoder(r.Body).Decode(&biomassAggregator)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, string(parseError))
		return
	}
	err = h.service.createBiomassAggregator(biomassAggregator)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "error creating biomass aggregator")
		return
	}

	utils.RespondJSON(w, http.StatusCreated, messageResponse{
		Message: "biomass aggregator created",
	})
}

// getBiomassAggregator get all biomass aggregator
func (h *Handler) getBiomassAggregator(resp http.ResponseWriter, request *http.Request) {
	filters := utils.NewFilters(request.URL.Query())
	biomassAggregatorResponse, err := h.service.getBiomassAggregator(filters)
	if err != nil {
		utils.RespondError(resp, http.StatusBadRequest, err, "failed to fetch biomass aggregators")
		return
	}
	utils.RespondJSON(resp, http.StatusOK, biomassAggregatorResponse)
}

// getBiomassAggregatorById get biomass aggregator by id
func (h *Handler) getBiomassAggregatorById(resp http.ResponseWriter, request *http.Request) {
	biomassAggregatorUrlID := chi.URLParam(request, "biomassAggregatorId")
	biomassAggregator, err := h.service.getBiomassAggregatorById(biomassAggregatorUrlID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		utils.RespondError(resp, http.StatusInternalServerError, err, "failed to fetch biomass aggregator details")
		return
	}
	if err != nil {
		utils.RespondError(resp, http.StatusBadRequest, err, "no biomass aggregator with that id in DB")
		return
	}
	utils.RespondJSON(resp, http.StatusOK, biomassAggregator)
}

// editBiomassAggregator ADMIN edit biomass aggregator
func (h *Handler) editBiomassAggregator(w http.ResponseWriter, r *http.Request) {
	var biomassAggregator biomassAggregatorDetailsRequest
	biomassAggregatorUrlID := chi.URLParam(r, "biomassAggregatorId")
	err := json.NewDecoder(r.Body).Decode(&biomassAggregator)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, string(parseError))
		return
	}
	err = h.service.editBiomassAggregator(biomassAggregator, biomassAggregatorUrlID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		utils.RespondError(w, http.StatusInternalServerError, err, "error editing c-sink network details")
		return
	}
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "no c-sink network with that id in DB")
		return
	}

	utils.RespondJSON(w, http.StatusCreated, messageResponse{
		Message: "biomass aggregator details edited",
	})
}

// deleteBiomassAggregator delete biomass aggregator by id
func (h *Handler) deleteBiomassAggregator(w http.ResponseWriter, r *http.Request) {
	biomassAggregatorUrlID := chi.URLParam(r, "biomassAggregatorId")
	err := h.service.deleteBiomassAggregator(biomassAggregatorUrlID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		utils.RespondError(w, http.StatusInternalServerError, err, "error deleting biomass aggregator details")
		return
	}
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "no biomass aggregator with that id in DB")
		return
	}
	utils.RespondJSON(w, http.StatusOK, messageResponse{
		Message: "biomass aggregator deleted",
	})
}

// getBANetworkList get all biomass aggregator
func (h *Handler) getBANetworkList(resp http.ResponseWriter, request *http.Request) {
	filters := utils.NewFilters(request.URL.Query())
	biomassAggregatorUrlID := chi.URLParam(request, "biomassAggregatorId")
	bANetworksResponse, err := h.service.getBANetworkList(filters, biomassAggregatorUrlID)
	if err != nil {
		utils.RespondError(resp, http.StatusBadRequest, err, "failed to fetch biomass aggregators c-sink networks")
		return
	}
	utils.RespondJSON(resp, http.StatusOK, bANetworksResponse)
}

// createCSNetwork ADMIN create C S Network
func (h *Handler) createCSNetwork(w http.ResponseWriter, r *http.Request) {
	var network networkRequest
	err := json.NewDecoder(r.Body).Decode(&network)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, string(parseError))
		return
	}
	err = h.service.createCSNetwork(network)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "error creating c-sink network")
		return
	}
	utils.RespondJSON(w, http.StatusCreated, messageResponse{
		Message: "c s network created",
	})
}

// getCSNetwork get all C S Network
func (h *Handler) getCSNetwork(resp http.ResponseWriter, request *http.Request) {
	filters := utils.NewFilters(request.URL.Query())
	networkResponse, err := h.service.getCSNetwork(filters)
	if err != nil {
		utils.RespondError(resp, http.StatusBadRequest, err, "failed to fetch c-sink network")
		return
	}
	utils.RespondJSON(resp, http.StatusOK, networkResponse)
}

// editCSNetwork ADMIN editCSNetwork
func (h *Handler) editCSNetwork(w http.ResponseWriter, r *http.Request) {
	var request editNetworkRequest
	id := chi.URLParam(r, "cs-networkId")
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, string(parseError))
		return
	}
	err = h.service.editCSNetwork(request, id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		utils.RespondError(w, http.StatusInternalServerError, err, "error editing c-sink network  details")
		return
	}
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "no c-sink network with that id in DB")
		return
	}

	utils.RespondJSON(w, http.StatusCreated, messageResponse{
		Message: "biomass aggregator details edited",
	})
}

// deleteCSNetwork deleteCSNetwork by id
func (h *Handler) deleteCSNetwork(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "cs-networkId")
	err := h.service.deleteCSNetwork(id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		utils.RespondError(w, http.StatusInternalServerError, err, "error deleting c-sink network details")
		return
	}
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "no c-sink network with that id in DB")
		return
	}
	utils.RespondJSON(w, http.StatusOK, messageResponse{
		Message: "biomass aggregator deleted",
	})
}

// createCSNetworkManager ADMIN create C S Network Manager
func (h *Handler) createCSNetworkManager(w http.ResponseWriter, r *http.Request) {
	var networkManger networkManagerRequest
	err := json.NewDecoder(r.Body).Decode(&networkManger)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, string(parseError))
		return
	}
	err = h.service.createCSNetworkManager(networkManger)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "error creating c-sink network manager")
		return
	}
	utils.RespondJSON(w, http.StatusCreated, messageResponse{
		Message: "c-sink network manager created",
	})
}

// getBAApprovedFarmerList get B A Approved Farmer List
func (h *Handler) getBAApprovedFarmerList(resp http.ResponseWriter, request *http.Request) {
	filters := utils.NewFilters(request.URL.Query())
	biomassAggregatorUrlID := chi.URLParam(request, "biomassAggregatorId")
	bANetworksResponse, err := h.service.getBAApprovedFarmerList(filters, biomassAggregatorUrlID)
	if err != nil {
		utils.RespondError(resp, http.StatusBadRequest, err, "failed to fetch biomass aggregators approved farmers")
		return
	}
	utils.RespondJSON(resp, http.StatusOK, bANetworksResponse)
}

// getBAPendingFarmerList get B A Pending Farmer List
func (h *Handler) getBAPendingFarmerList(resp http.ResponseWriter, request *http.Request) {
	filters := utils.NewFilters(request.URL.Query())
	biomassAggregatorUrlID := chi.URLParam(request, "biomassAggregatorId")
	bANetworksResponse, err := h.service.getBAPendingFarmerList(filters, biomassAggregatorUrlID)
	if err != nil {
		utils.RespondError(resp, http.StatusBadRequest, err, "failed to fetch biomass aggregators pending farmers")
		return
	}
	utils.RespondJSON(resp, http.StatusOK, bANetworksResponse)
}

// createKiln create Kiln
func (h *Handler) createKiln(w http.ResponseWriter, r *http.Request) {
	var kiln kilnRequest
	err := json.NewDecoder(r.Body).Decode(&kiln)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, string(parseError))
		return
	}
	err = h.service.createKiln(kiln)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "error creating kiln")
		return
	}
	utils.RespondJSON(w, http.StatusCreated, messageResponse{
		Message: "kiln created",
	})
}

// getKiln get kiln
func (h *Handler) getKiln(resp http.ResponseWriter, request *http.Request) {
	filters := utils.NewFilters(request.URL.Query())
	kilnResponse, err := h.service.getKiln(filters)
	if err != nil {
		utils.RespondError(resp, http.StatusBadRequest, err, "failed to fetch kilns")
		return
	}
	utils.RespondJSON(resp, http.StatusOK, kilnResponse)
}

// editKiln edit kiln
func (h *Handler) editKiln(w http.ResponseWriter, r *http.Request) {
	var request editKilnRequest
	id := chi.URLParam(r, "kilnId")
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, string(parseError))
		return
	}
	err = h.service.editKiln(request, id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		utils.RespondError(w, http.StatusInternalServerError, err, "error editing kiln details")
		return
	}
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "no kiln with that id in DB")
		return
	}

	utils.RespondJSON(w, http.StatusCreated, messageResponse{
		Message: "kiln details edited",
	})
}

// deleteKiln delete kiln by id
func (h *Handler) deleteKiln(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "kilnId")
	err := h.service.deleteKiln(id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		utils.RespondError(w, http.StatusInternalServerError, err, "error deleting kiln")
		return
	}
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "no kiln with that id in DB")
		return
	}
	utils.RespondJSON(w, http.StatusOK, messageResponse{
		Message: "kiln deleted",
	})
}

// createKilnOperator create kiln operator
func (h *Handler) createKilnOperator(w http.ResponseWriter, r *http.Request) {
	var kilnOperator kilnOperatorRequest
	err := json.NewDecoder(r.Body).Decode(&kilnOperator)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, string(parseError))
		return
	}
	err = h.service.createKilnOperator(kilnOperator)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "error creating kiln operator")
		return
	}
	utils.RespondJSON(w, http.StatusCreated, messageResponse{
		Message: "kiln operator created",
	})
}

// assigningFarmerToCSNetwork assigning Farmer To C S network
func (h *Handler) assigningFarmerToCSNetwork(w http.ResponseWriter, r *http.Request) {
	var request farmerId
	id := chi.URLParam(r, "cs-networkId")
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, string(parseError))
		return
	}
	err = h.service.assigningFarmerToCSNetwork(request, id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		utils.RespondError(w, http.StatusInternalServerError, err, "error assigning farmer to c s network")
		return
	}
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "no c-sink network with that id in DB")
		return
	}

	utils.RespondJSON(w, http.StatusCreated, messageResponse{
		Message: "assigning farmer to cs network successful",
	})
}

// getBARejectedFarmerList get B A Rejected Farmer List
func (h *Handler) getBARejectedFarmerList(resp http.ResponseWriter, request *http.Request) {
	filters := utils.NewFilters(request.URL.Query())
	biomassAggregatorUrlID := chi.URLParam(request, "biomassAggregatorId")
	bANetworksResponse, err := h.service.getBARejectedFarmerList(filters, biomassAggregatorUrlID)
	if err != nil {
		utils.RespondError(resp, http.StatusBadRequest, err, "failed to fetch biomass aggregators approved farmers")
		return
	}
	utils.RespondJSON(resp, http.StatusOK, bANetworksResponse)
}

// getCSNKilnList get C S N by id Kiln List
func (h *Handler) getCSNKilnList(w http.ResponseWriter, r *http.Request) {
	filters := utils.NewFilters(r.URL.Query())
	id := chi.URLParam(r, "cs-networkId")
	kilnResponse, err := h.service.getCSNKilnList(filters, id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		utils.RespondError(w, http.StatusInternalServerError, err, "error fetching c-sink network kilns list")
		return
	}
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "no c-sink network with that id in DB")
		return
	}
	utils.RespondJSON(w, http.StatusOK, kilnResponse)
}

// getCSNFarmerList get C S N by id Farmer List
func (h *Handler) getCSNFarmerList(w http.ResponseWriter, r *http.Request) {
	filters := utils.NewFilters(r.URL.Query())
	id := chi.URLParam(r, "cs-networkId")
	farmerResponse, err := h.service.getCSNFarmerList(filters, id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		utils.RespondError(w, http.StatusInternalServerError, err, "error fetching c-sink network farmer list")
		return
	}
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "no c-sink network with that id in DB")
		return
	}
	utils.RespondJSON(w, http.StatusOK, farmerResponse)
}

// bARejectFarmer B A reject farmer
func (h *Handler) bARejectFarmer(w http.ResponseWriter, r *http.Request) {
	var request farmerId
	id := chi.URLParam(r, "biomassAggregatorId")
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, string(parseError))
		return
	}
	err = h.service.bARejectFarmer(request, id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		utils.RespondError(w, http.StatusInternalServerError, err, "error rejecting farmer")
		return
	}
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "no biomass aggregator with that id in DB")
		return
	}

	utils.RespondJSON(w, http.StatusCreated, messageResponse{
		Message: "farmer is successfully rejected by biomass aggregator",
	})
}

// getCSNetworkDetailsByID get all C S Network
func (h *Handler) getCSNetworkDetailsByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "cs-networkId")
	biomassAggregatorResponse, err := h.service.getCSNetworkDetailsByID(id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		utils.RespondError(w, http.StatusInternalServerError, err, "error fetching c-sink network details")
		return
	}
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "no c-sink network with that id in DB")
		return
	}
	utils.RespondJSON(w, http.StatusOK, biomassAggregatorResponse)
}
