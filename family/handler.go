package family

import (
	"circonomy-server/middlewares"
	"circonomy-server/utils"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"net/http"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Serve(familyRouter chi.Router) {
	familyRouter.Use(middlewares.AuthMiddleware)

	// create family
	familyRouter.Post("/", h.createFamily)
	// get specific family details
	familyRouter.Get("/own", h.getOwnFamily)
	// update family
	familyRouter.Put("/{familyId}", h.updateFamily)
	// get specific family details
	familyRouter.Get("/{familyId}", h.getFamily)
	// invite to a family
	familyRouter.Post("/invite", h.invite)
	// get invitation details
	familyRouter.Get("/invite/{inviteId}", h.invitationDetails)
	// join family
	familyRouter.Post("/invite/{inviteId}/join", h.join)
}

func (h *Handler) createFamily(resp http.ResponseWriter, request *http.Request) {
	uc := middlewares.GetUserContext(request)
	var familyRequest createFamilyRequest
	err := json.NewDecoder(request.Body).Decode(&familyRequest)
	if err != nil {
		utils.RespondError(resp, http.StatusBadRequest, err, "failed to parse the request")
		return
	}

	family, err := h.service.createFamily(familyRequest, uc)
	if err != nil {
		utils.RespondError(resp, http.StatusInternalServerError, err, "failed to create company")
		return
	}

	utils.RespondJSON(resp, http.StatusOK, family)
}

func (h *Handler) updateFamily(resp http.ResponseWriter, request *http.Request) {
	uc := middlewares.GetUserContext(request)

	idStr := chi.URLParam(request, "familyId")
	familyID, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondError(resp, http.StatusBadRequest, err, "failed to parse id")
		return
	}

	var familyRequest updateFamilyRequest
	err = json.NewDecoder(request.Body).Decode(&familyRequest)
	if err != nil {
		utils.RespondError(resp, http.StatusBadRequest, err, "failed to parse the request")
		return
	}

	family, err := h.service.updateFamily(familyRequest, familyID, uc.ID)
	if err != nil {
		utils.RespondError(resp, http.StatusInternalServerError, err, "failed to update company")
		return
	}

	utils.RespondJSON(resp, http.StatusOK, family)
}

func (h *Handler) getFamily(resp http.ResponseWriter, request *http.Request) {
	uc := middlewares.GetUserContext(request)

	idStr := chi.URLParam(request, "familyId")
	familyID, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondError(resp, http.StatusBadRequest, err, "failed to parse id")
		return
	}
	family, err := h.service.getFamily(familyID, uc.ID)
	if err != nil {
		utils.RespondError(resp, http.StatusInternalServerError, err, "failed to get company")
		return
	}

	utils.RespondJSON(resp, http.StatusOK, family)
}

func (h *Handler) getOwnFamily(resp http.ResponseWriter, request *http.Request) {
	uc := middlewares.GetUserContext(request)

	family, err := h.service.getOwnFamily(uc.ID)
	if err != nil {
		utils.RespondError(resp, http.StatusInternalServerError, err, "failed to get company")
		return
	}

	utils.RespondJSON(resp, http.StatusOK, family)
}

func (h *Handler) invite(resp http.ResponseWriter, request *http.Request) {
	//uc := middlewares.GetUserContext(request)
	var inviteRequest inviteFamilyRequest
	err := json.NewDecoder(request.Body).Decode(&inviteRequest)
	if err != nil {
		utils.RespondError(resp, http.StatusBadRequest, err, "failed to parse the request")
		return
	}
	message, err := h.service.invite(inviteRequest.FamilyID, inviteRequest.Email)
	if err != nil {
		utils.RespondError(resp, http.StatusInternalServerError, err, "failed to send invite")
		return
	}

	utils.RespondJSON(resp, http.StatusOK, map[string]interface{}{
		"message": message,
	})
}

func (h *Handler) invitationDetails(resp http.ResponseWriter, request *http.Request) {
	uc := middlewares.GetUserContext(request)
	idStr := chi.URLParam(request, "inviteId")
	inviteID, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondError(resp, http.StatusBadRequest, err, "failed to parse id")
		return
	}

	details, err := h.service.getInvitationDetails(inviteID, uc.ID)
	if err != nil {
		utils.RespondError(resp, http.StatusInternalServerError, err, "failed to send invite")
		return
	}

	utils.RespondJSON(resp, http.StatusOK, details)
}

func (h *Handler) join(resp http.ResponseWriter, request *http.Request) {
	uc := middlewares.GetUserContext(request)
	idStr := chi.URLParam(request, "inviteId")
	inviteID, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondError(resp, http.StatusBadRequest, err, "failed to parse id")
		return
	}

	err = h.service.joinFamily(inviteID, uc.ID)
	if err != nil {
		utils.RespondError(resp, http.StatusInternalServerError, err, "failed to join invite")
		return
	}

	resp.WriteHeader(http.StatusCreated)
}
