package project

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

func (h *Handler) Serve(projectRouter chi.Router) {
	projectRouter.Use(middlewares.AuthMiddleware)
	projectRouter.Route("/public", func(home chi.Router) {
		home.Get("/locations", h.getProjectLocations)
		home.Get("/", h.getAllProjects)
		home.Get("/status/{status}", h.getProjectsByStatus)
		home.Get("/id/{id}", h.getProjectByID)
	})
	projectRouter.Get("/", h.getAllProjects)
	projectRouter.Post("/", h.createProject)
	projectRouter.Get("/status/{status}", h.getProjectsByStatus)
	projectRouter.Get("/id/{id}", h.getProjectByID)
}

func (h *Handler) getAllProjects(w http.ResponseWriter, _ *http.Request) {
	projects, err := h.service.getAllProjects()
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "failed to get all projects")
		return
	}

	utils.RespondJSON(w, http.StatusOK, projects)
}

func (h *Handler) getProjectsByStatus(w http.ResponseWriter, r *http.Request) {
	status := chi.URLParam(r, "status")
	address := r.URL.Query().Get("location")

	details, err := h.service.getProjectsByStatus(status, address)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "failed to get all projects")
		return
	}

	utils.RespondJSON(w, http.StatusOK, details)
}

func (h *Handler) getProjectByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "failed to parse id")
		return
	}

	details, err := h.service.getProjectById(id)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "failed to get details")
		return
	}
	utils.RespondJSON(w, http.StatusOK, details)
}

func (h *Handler) getProjectLocations(w http.ResponseWriter, _r *http.Request) {
	locations, err := h.service.getProjectLocations()
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "failed to get all project locations")
		return
	}

	utils.RespondJSON(w, http.StatusOK, locations)
}

func (h *Handler) createProject(resp http.ResponseWriter, request *http.Request) {
	uc := middlewares.GetUserContext(request)
	var familyRequest createProjectRequest
	err := json.NewDecoder(request.Body).Decode(&familyRequest)
	if err != nil {
		utils.RespondError(resp, http.StatusBadRequest, err, "failed to parse the request")
		return
	}

	projectID, err := h.service.createProject(familyRequest, uc.ID)
	if err != nil {
		utils.RespondError(resp, http.StatusInternalServerError, err, "failed to create company")
		return
	}

	utils.RespondJSON(resp, http.StatusOK, map[string]interface{}{
		"id": projectID,
	})
}
