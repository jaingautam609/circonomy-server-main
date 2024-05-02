package farmer

import (
	"circonomy-server/dbhelpers"
	"circonomy-server/middlewares"
	"circonomy-server/models"
	"circonomy-server/utils"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/volatiletech/null"
	"mime/multipart"
	"net/http"
)

const storedOTP = "6666"
const MaxMemorySize int64 = 32
const MaxMemoryLimit int64 = 20

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Serve(farmerRouter chi.Router) {
	farmerRouter.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		utils.RespondJSON(w, http.StatusOK, struct {
			Status string `json:"status"`
		}{Status: "server is running"})
	})
	farmerRouter.Post("/send-otp", h.sendOTP)
	farmerRouter.Post("/refresh-auth-token", h.refreshAuthToken)
	farmerRouter.Post("/verify-otp", h.verifyOTP)
	farmerRouter.Get("/crops", h.getCrops)
	farmerRouter.Get("/fertilizers", h.getFertilizers) // to get fertilizers
	farmerRouter.Get("/farmer-videos", h.getFarmingVideos)

	farmerRouter.Route("/", func(private chi.Router) {
		private.Use(middlewares.FarmerAuthMiddleware)
		private.Post("/upload", h.uploadFile)
		private.Get("/video-content", h.getVideos)
		private.Post("/preferred-crops", h.addPreferredCrops)

		private.Route("/profile", func(details chi.Router) {
			details.Get("/", h.fetchUserDetails)
			details.Get("/preferred-crops", h.fetchUserPreferredCrops) // new API
			details.Put("/", h.updateUserDetails)
		})

		private.Route("/farm-crops", func(crop chi.Router) {
			crop.Get("/", h.farmCrops)
			crop.Post("/", h.addFarmCrop)
			crop.Route("/{farmCropId}", func(farmCropRoute chi.Router) {
				farmCropRoute.Get("/", h.getFarmCropDetails)
				farmCropRoute.Put("/cropping", h.editCropping)
				farmCropRoute.Put("/harvesting", h.editHarvesting)
				farmCropRoute.Put("/sundrying", h.editSundrying)
				farmCropRoute.Put("/transportation", h.editTransportation)
				farmCropRoute.Put("/farmer-move-to-production", h.moveToProduction)
				farmCropRoute.Put("/change-status", h.cropChangeStatus)
				farmCropRoute.Put("/distribution", h.editDistribution)

			})
		})

		private.Route("/farm", func(farm chi.Router) {
			farm.Post("/", h.addFarmDetails)
			farm.Get("/", h.fetchFarmDetails)
			farm.Get("/for-add-crop", h.fetchFarmDetails)
			farm.Route("/{farmId}", func(farmID chi.Router) {
				farmID.Delete("/", h.deleteFarm)
			})
		})
	})
}

// sendOTP send OTP
func (h *Handler) sendOTP(resp http.ResponseWriter, request *http.Request) {
	var otp sendOTPRequest
	err := json.NewDecoder(request.Body).Decode(&otp)
	if err != nil {
		utils.RespondError(resp, http.StatusBadRequest, err, "failed to parse the request")
		return
	}
	if otp.PhoneNumber == "" || otp.CountryCode == "" {
		utils.RespondError(resp, http.StatusBadRequest, err, "phone number or country code were not send")
		return
	}
	err = h.service.sendOTP(otp)
	if err != nil {
		utils.RespondError(resp, http.StatusBadRequest, err, "failed to send OTP to user")
		return
	}
	utils.RespondJSON(resp, http.StatusOK, sendOTPResponse{
		Message: "OTP sent",
	})
}

// verifyOTP verify OTP and register unregister phone no
func (h *Handler) verifyOTP(resp http.ResponseWriter, request *http.Request) {
	var otp checkOTPRequest
	err := json.NewDecoder(request.Body).Decode(&otp)
	if err != nil {
		utils.RespondError(resp, http.StatusBadRequest, err, "failed to parse the request")
		return
	}
	res, err := h.service.verifyOTPAndGenerateToken(otp)
	if err != nil {
		utils.RespondError(resp, http.StatusBadRequest, err, "failed to check user OTP")
		return
	}
	utils.RespondJSON(resp, http.StatusOK, res)
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

// updateUserDetails add user details like name, age, gender, address, profile image id
func (h *Handler) updateUserDetails(w http.ResponseWriter, r *http.Request) {
	var farmerInfo updateFarmerDetails
	err := json.NewDecoder(r.Body).Decode(&farmerInfo)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, string(parseError))
		return
	}
	userCtx := r.Context().Value(middlewares.UserContextKey).(*models.FarmerClaims)
	err = h.service.updateAccountDetails(userCtx.FarmerID, farmerInfo)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "error updating farmer details")
		return
	}

	utils.RespondJSON(w, http.StatusOK, messageResponse{
		Message: "User details updated successfully",
	})
}

// addFarmDetails add farm details
func (h *Handler) addFarmDetails(w http.ResponseWriter, r *http.Request) {
	var farmInfo addFarmDetails
	err := json.NewDecoder(r.Body).Decode(&farmInfo)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, string(parseError))
		return
	}
	userCtx := r.Context().Value(middlewares.UserContextKey).(*models.FarmerClaims)
	farmID, err := h.service.addFarm(userCtx.FarmerID, farmInfo)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "error adding farm details")
		return
	}

	utils.RespondJSON(w, http.StatusOK, addFarmResponse{
		FarmID: farmID,
	})
}

// getCrops get crops
func (h *Handler) getCrops(resp http.ResponseWriter, request *http.Request) {
	filters := utils.NewFilters(request.URL.Query())
	cropsResponse, err := h.service.GetCrops(filters)
	if err != nil {
		utils.RespondError(resp, http.StatusBadRequest, err, "failed to fetch crops")
		return
	}
	utils.RespondJSON(resp, http.StatusOK, cropsResponse)
}

// fetchUserDetails fetch user details
func (h *Handler) fetchUserDetails(w http.ResponseWriter, r *http.Request) {
	userCtx := r.Context().Value(middlewares.UserContextKey).(*models.FarmerClaims)
	farmerInfo, err := h.service.fetchFarmerDetails(userCtx.FarmerID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "error fetching farm details")
		return
	}

	utils.RespondJSON(w, http.StatusOK, farmerInfo)
}

// fetchFarmDetails fetch farm details
func (h *Handler) fetchFarmDetails(w http.ResponseWriter, r *http.Request) {
	userCtx := r.Context().Value(middlewares.UserContextKey).(*models.FarmerClaims)
	farmInfo, err := h.service.fetchFarmDetails(userCtx.FarmerID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "error fetching farm details")
		return
	}

	utils.RespondJSON(w, http.StatusOK, fetchFarmResponse{
		Farms: farmInfo,
	})
}

// uploadFile upload image
func (h *Handler) uploadFile(w http.ResponseWriter, r *http.Request) {
	// Parse the multipart form, up to 32 MiB in memory
	if err := r.ParseMultipartForm(MaxMemorySize << MaxMemoryLimit); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "failed to parse multipartForm")
		return
	}

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

// deleteFarm deletes farm
func (h *Handler) deleteFarm(w http.ResponseWriter, r *http.Request) {
	farmID := chi.URLParam(r, "farmId")
	userCtx := r.Context().Value(middlewares.UserContextKey).(*models.FarmerClaims)
	rowsAffected, err := h.service.deleteFarm(farmID, userCtx.FarmerID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "error deleting farm")
		return
	}

	if rowsAffected == 0 {
		utils.RespondError(w, http.StatusForbidden, err, "user can't delete this resource")
		return
	}

	utils.RespondJSON(w, http.StatusOK, messageResponse{
		Message: "Farm deleted successfully",
	})
}

// addPreferredCrops add farmer preferred crops
func (h *Handler) addPreferredCrops(resp http.ResponseWriter, request *http.Request) {
	var addCropRequest addCropsRequest
	err := json.NewDecoder(request.Body).Decode(&addCropRequest)
	if err != nil || addCropRequest.IDs == nil {
		utils.RespondError(resp, http.StatusBadRequest, err, "failed to parse the request")
		return
	}

	userCtx := request.Context().Value(middlewares.UserContextKey).(*models.FarmerClaims)
	err = h.service.addPreferredCrops(userCtx.FarmerID, addCropRequest)
	if err != nil {
		utils.RespondError(resp, http.StatusBadRequest, err, "failed to add crops")
		return
	}

	utils.RespondJSON(resp, http.StatusOK, messageResponse{
		Message: "crops added to db",
	})
}

// farmCrops get farmer crops by filters
func (h *Handler) farmCrops(w http.ResponseWriter, r *http.Request) {
	userCtx := r.Context().Value(middlewares.UserContextKey).(*models.FarmerClaims)
	filters := utils.CropStageFilters(r.URL.Query())
	cropsResponse, err := h.service.getFarmCrops(userCtx.FarmerID, filters)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "failed to add crops")
		return
	}
	utils.RespondJSON(w, http.StatusOK, cropsResponse)
}

// addFarmCrop add farm crop
func (h *Handler) addFarmCrop(w http.ResponseWriter, r *http.Request) {
	var crop cropFormRequest
	err := json.NewDecoder(r.Body).Decode(&crop)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, string(parseError))
		return
	}
	cropID, err := h.service.addFarmCrop(crop)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "error adding farm details")
		return
	}
	utils.RespondJSON(w, http.StatusOK, addFarmCropResponse{
		ID: cropID,
	})
}

// addFarmCrop add farm crop
func (h *Handler) editCropping(w http.ResponseWriter, r *http.Request) {
	farmCropId := chi.URLParam(r, "farmCropId")
	var crop editCroppingRequest
	err := json.NewDecoder(r.Body).Decode(&crop)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, string(parseError))
		return
	}
	err = h.service.editCropping(farmCropId, crop)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "error editing crop details")
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) editHarvesting(w http.ResponseWriter, r *http.Request) {
	farmCropId := chi.URLParam(r, "farmCropId")
	var crop editHarvestingRequest
	err := json.NewDecoder(r.Body).Decode(&crop)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, string(parseError))
		return
	}
	err = h.service.editHarvesting(farmCropId, crop)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "error editing harvesting details")
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// editSundrying edit sundrying
func (h *Handler) editSundrying(w http.ResponseWriter, r *http.Request) {
	farmCropId := chi.URLParam(r, "farmCropId")
	var crop editSundryingRequest
	err := json.NewDecoder(r.Body).Decode(&crop)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, string(parseError))
		return
	}
	err = h.service.editSundrying(farmCropId, crop)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "error editing sun-drying details")
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// getFertilizers get fertilizer
func (h *Handler) getFertilizers(resp http.ResponseWriter, request *http.Request) {
	filters := utils.NewFilters(request.URL.Query())
	fertilizersResponse, err := h.service.GetFertilizers(filters)
	if err != nil {
		utils.RespondError(resp, http.StatusBadRequest, err, "failed to fetch fertilizers")
		return
	}
	utils.RespondJSON(resp, http.StatusOK, fertilizersResponse)
}

func (h *Handler) getVideos(resp http.ResponseWriter, request *http.Request) {
	filters := utils.VideoFilter(request.URL.Query())
	videoResponse, err := h.service.GetVideos(filters)
	if err != nil {
		utils.RespondError(resp, http.StatusBadRequest, err, "failed to fetch videos")
		return
	}
	utils.RespondJSON(resp, http.StatusOK, videoResponse)
}

func (h *Handler) getFarmingVideos(resp http.ResponseWriter, request *http.Request) {
	filters := utils.VideoFilter(request.URL.Query())
	filters.VideoType = null.StringFrom("farming")
	videoResponse, err := h.service.GetVideos(filters)
	if err != nil {
		utils.RespondError(resp, http.StatusBadRequest, err, "failed to fetch videos")
		return
	}
	utils.RespondJSON(resp, http.StatusOK, videoResponse)
}

// getFarmCropDetails get crop Details
func (h *Handler) getFarmCropDetails(w http.ResponseWriter, r *http.Request) {
	cropUrlID := chi.URLParam(r, "farmCropId")

	res, err := h.service.GetFarmCropDetails(cropUrlID)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "failed to get crop details")
		return
	}
	utils.RespondJSON(w, http.StatusOK, res)
}

// cropChangeStatus crop change status
func (h *Handler) cropChangeStatus(w http.ResponseWriter, r *http.Request) {
	cropUrlID := chi.URLParam(r, "farmCropId")
	var cropStage cropStageRequest
	err := json.NewDecoder(r.Body).Decode(&cropStage)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, string(parseError))
		return
	}
	err = h.service.cropChangeStatus(cropUrlID, cropStage.Status)
	if err != nil {
		if err.Error() == "can't change status, detail are not complete" {
			utils.RespondJSON(w, http.StatusBadRequest, changeStatusResponse{
				IsChangeStatusPossible: false,
			})
			return
		}
		utils.RespondError(w, http.StatusBadRequest, err, "failed to change crop status")
		return
	}
	utils.RespondJSON(w, http.StatusOK, messageResponse{
		Message: "crop status changed",
	})
}

// fetchUserPreferredCrops fetch user preferred crops
func (h *Handler) fetchUserPreferredCrops(w http.ResponseWriter, r *http.Request) {
	userCtx := r.Context().Value(middlewares.UserContextKey).(*models.FarmerClaims)
	cropsInfo, err := h.service.fetchUserPreferredCrops(userCtx.FarmerID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "error fetching farmer preferred crops")
		return
	}
	utils.RespondJSON(w, http.StatusOK, cropsInfo)
}

// editTransportation edit transportation
func (h *Handler) editTransportation(w http.ResponseWriter, r *http.Request) {
	farmCropId := chi.URLParam(r, "farmCropId")
	var crop editTransportationRequest
	err := json.NewDecoder(r.Body).Decode(&crop)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, string(parseError))
		return
	}
	err = h.service.editTransportation(farmCropId, crop)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "error editing transportation details")
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// moveToProductionRequest move to production - check transportation details, update production details, change status
func (h *Handler) moveToProduction(w http.ResponseWriter, r *http.Request) {
	farmCropId := chi.URLParam(r, "farmCropId")
	var crop moveToProductionRequest
	err := json.NewDecoder(r.Body).Decode(&crop)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, string(parseError))
		return
	}
	err = h.service.moveToProduction(farmCropId, crop)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "error moving crop to production")
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// editDistribution edit distribution
func (h *Handler) editDistribution(w http.ResponseWriter, r *http.Request) {
	farmCropId := chi.URLParam(r, "farmCropId")
	var crop editDistributionRequest
	err := json.NewDecoder(r.Body).Decode(&crop)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, string(parseError))
		return
	}
	err = h.service.editDistribution(farmCropId, crop)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "error editing transportation details")
		return
	}
	w.WriteHeader(http.StatusCreated)
}
