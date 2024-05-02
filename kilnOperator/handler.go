package kilnOperator

import (
	"circonomy-server/dbhelpers"
	"circonomy-server/farmer"
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
	service       *Service
	farmerService *farmer.Service
}

func NewHandler(service *Service, farmerService *farmer.Service) *Handler {
	return &Handler{
		service:       service,
		farmerService: farmerService,
	}
}

// Serve router
func (h *Handler) Serve(kilnOperatorRouter chi.Router) {

	kilnOperatorRouter.Post("/send-otp", h.sendOTP)
	kilnOperatorRouter.Post("/refresh-auth-token", h.refreshAuthToken)
	kilnOperatorRouter.Post("/verify-otp", h.verifyOTP)

	kilnOperatorRouter.Route("/", func(private chi.Router) {
		private.Use(middlewares.KilnOperatorAuthMiddleware)
		private.Post("/upload", h.uploadFile)
		private.Get("/profile", h.fetchKilnOperatorInfo)

		private.Get("/farm-crop/{farmCropId}", h.getFarmCropBiomassDetails) // farmer service API

		private.Route("/{kilnId}", func(klinRoute chi.Router) {

			klinRoute.Route("/biomass", func(biomass chi.Router) {
				biomass.Get("/", h.getFarmCropsBiomasses)
				biomass.Put("/{farmCropId}/move-to-production", h.editFarmCropAndAddKilnBiomass)
			})

			klinRoute.Route("/biochar-production", func(kilnProcess chi.Router) {
				kilnProcess.Get("/biomasses", h.getKilnCropBiomasses)
				kilnProcess.Get("/kiln-process", h.getKilnProcessDetails)

				kilnProcess.Route("/batch", func(kilnBatch chi.Router) {
					kilnBatch.Post("/", h.addKilnProcessBatch)
					kilnBatch.Put("/edit", h.editKilnProcessDetails)
					kilnBatch.Post("/done", h.doneKilnProcess)
				})
				kilnProcess.Get("/biochar", h.fetchKilnBiochar)
			})

			klinRoute.Route("/distribution", func(distribution chi.Router) {
				distribution.Post("/{farmCropId}", h.addFarmCropBioCharDetails)
				distribution.Get("/", h.fetchDistributedFarmCrops)

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

// refreshAuthToken refresh auth token
func (h *Handler) refreshAuthToken(resp http.ResponseWriter, request *http.Request) {
	tokenReq := models.RefreshTokenRequest{}
	err := json.NewDecoder(request.Body).Decode(&tokenReq)
	if err != nil {
		utils.RespondError(resp, http.StatusBadRequest, err, "failed to parse the request")
		return
	}

	token, err := utils.RefreshAuthTokenKilnOperator(tokenReq)
	if err != nil {
		utils.RespondError(resp, http.StatusUnauthorized, err, "Error generating access token")
		return
	}

	utils.RespondJSON(resp, http.StatusOK, token)
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

// getFarmCropsBiomasses get farmer crops biomass by filters
func (h *Handler) getFarmCropsBiomasses(w http.ResponseWriter, r *http.Request) {
	userCtx := r.Context().Value(middlewares.UserContextKey).(*models.KilnOperatorClaims)
	kilnID := chi.URLParam(r, "kilnId")
	filters := utils.KilnFilter(r.URL.Query())
	filters.KilnID = null.StringFrom(kilnID)
	cropsBiomassResponse, err := h.service.getFarmCropsBiomasses(userCtx.KilnOperatorID, filters)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "failed to find farm crops biomass")
		return
	}
	utils.RespondJSON(w, http.StatusOK, cropsBiomassResponse)
}

// getFarmCropBiomassDetails get farm crop biomass details
func (h *Handler) getFarmCropBiomassDetails(w http.ResponseWriter, r *http.Request) {
	cropUrlID := chi.URLParam(r, "farmCropId")
	res, err := h.farmerService.GetFarmCropDetails(cropUrlID)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "failed to get crop biomass details")
		return
	}
	utils.RespondJSON(w, http.StatusOK, res)
}

// editFarmCropAndAddKilnBiomass edit farm crop and add kiln biomass
func (h *Handler) editFarmCropAndAddKilnBiomass(w http.ResponseWriter, r *http.Request) {
	farmCropId := chi.URLParam(r, "farmCropId")
	kilnID := chi.URLParam(r, "kilnId")
	var crop editFarmCropBiomassRequest
	err := json.NewDecoder(r.Body).Decode(&crop)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, string(parseError))
		return
	}
	err = h.service.editFarmCropAndAddKilnBiomass(farmCropId, kilnID, crop)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "error editing crop biomass details and update kiln biomass")
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// fetchKilnOperatorInfo fetch Kiln Operator Info
func (h *Handler) fetchKilnOperatorInfo(w http.ResponseWriter, r *http.Request) {
	userCtx := r.Context().Value(middlewares.UserContextKey).(*models.KilnOperatorClaims)
	kilnOperatorInfo, err := h.service.fetchKilnOperatorInfo(userCtx.KilnOperatorID)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "failed to fetch kiln operator Info")
		return
	}
	utils.RespondJSON(w, http.StatusOK, kilnOperatorInfo)
}

// getKilnCropBiomasses get Kiln Crop Biomasses
func (h *Handler) getKilnCropBiomasses(w http.ResponseWriter, r *http.Request) {
	kilnId := chi.URLParam(r, "kilnId")
	res, err := h.service.getKilnCropBiomasses(kilnId)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "failed to kiln crop biomasses")
		return
	}
	utils.RespondJSON(w, http.StatusOK, res)
}

// getKilnProcessDetails get Kiln Ongoing Process Details
func (h *Handler) getKilnProcessDetails(w http.ResponseWriter, r *http.Request) {
	kilnId := chi.URLParam(r, "kilnId")
	filters := utils.NewFilters(r.URL.Query())
	res, err := h.service.getKilnProcessDetails(kilnId, filters)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "failed to get kiln ongoing process details")
		return
	}
	utils.RespondJSON(w, http.StatusOK, res)
}

// addKilnProcessBatch add Kiln Process Batch
func (h *Handler) addKilnProcessBatch(w http.ResponseWriter, r *http.Request) {
	kilnId := chi.URLParam(r, "kilnId")
	var kilnProcess kilnProcessRequest
	err := json.NewDecoder(r.Body).Decode(&kilnProcess)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, string(parseError))
		return
	}
	err = h.service.addKilnProcessBatch(kilnId, kilnProcess)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "failed to add kiln batch details")
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// editKilnProcessDetails edit kiln process batch details
func (h *Handler) editKilnProcessDetails(w http.ResponseWriter, r *http.Request) {
	kilnId := chi.URLParam(r, "kilnId")
	var kilnProcess kilnProcessEditRequest
	err := json.NewDecoder(r.Body).Decode(&kilnProcess)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, string(parseError))
		return
	}
	err = h.service.editKilnProcessDetails(kilnId, kilnProcess)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "failed to add kiln batch image details")
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// doneKilnProcess done kiln process
func (h *Handler) doneKilnProcess(w http.ResponseWriter, r *http.Request) {
	kilnId := chi.URLParam(r, "kilnId")
	var kilnProcess kilnProcessDoneRequest
	err := json.NewDecoder(r.Body).Decode(&kilnProcess)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, string(parseError))
		return
	}
	err = h.service.doneKilnProcess(kilnId, kilnProcess)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "failed to complete kiln batch process")
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// fetchKilnBiochar fetch kiln biochar
func (h *Handler) fetchKilnBiochar(w http.ResponseWriter, r *http.Request) {
	kilnId := chi.URLParam(r, "kilnId")
	kilnBioChar, err := h.service.fetchKilnBiochar(kilnId)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "failed to complete kiln batch process")
		return
	}
	utils.RespondJSON(w, http.StatusOK, kilnBioChar)
}

// addFarmCropBioCharDetails add farm crop biochar details
func (h *Handler) addFarmCropBioCharDetails(w http.ResponseWriter, r *http.Request) {
	farmCropiD := chi.URLParam(r, "farmCropId")
	kilnID := chi.URLParam(r, "kilnId")
	var farmCropBiocharRequest kilnDistributionRequest
	err := json.NewDecoder(r.Body).Decode(&farmCropBiocharRequest)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, string(parseError))
		return
	}
	farmCropBiocharRequest.KilnID = kilnID
	err = h.service.addFarmCropBioCharDetails(farmCropiD, farmCropBiocharRequest)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "failed to add farm crop distribution biochar details")
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// fetchDistributedFarmCrops fetch distributed farm crops
func (h *Handler) fetchDistributedFarmCrops(w http.ResponseWriter, r *http.Request) {
	kilnId := chi.URLParam(r, "kilnId")
	filters := utils.NewFilters(r.URL.Query())
	res, err := h.service.fetchDistributedFarmCrops(kilnId, filters)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "failed to complete kiln batch process")
		return
	}
	utils.RespondJSON(w, http.StatusOK, res)
}
