package handlers

import (
	"circonomy-server/dbhelpers"
	"circonomy-server/models"
	"circonomy-server/utils"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func GetProfileByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "failed to parse id")
		return
	}

	user, getErr := dbhelpers.GetProfileByID(id)
	if getErr != nil {
		utils.RespondError(w, http.StatusBadRequest, getErr, "failed to get individual profile")
		return
	}
	if user.ImagePath.Valid {
		pfpURL, urlErr := utils.GenerateSignedURL(user.ImagePath.String)
		if urlErr != nil {
			utils.RespondError(w, http.StatusInternalServerError, urlErr, "failed to get profile picture url")
			return
		}
		user.ImageURL = pfpURL
	}
	if user.AccountType == models.UserAccountTypeIndividual {
		orgName, orgID, nameErr := dbhelpers.GetUserOrgName(id)
		if nameErr != nil {
			logrus.Info("could not find organisation details of the individual user")
		} else {
			user.OrgName = orgName
			user.OrgID = orgID
		}
	}

	utils.RespondJSON(w, http.StatusOK, user)
}

func GetPeopleInOrg(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "failed to parse id")
		return
	}
	people, err := dbhelpers.GetPeopleInOrg(id)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "failed to get all people in org")
		return
	}
	for i := 0; i < len(people); i++ {
		if people[i].ImagePath.String != "" {
			url, urlErr := utils.GenerateSignedURL(people[i].ImagePath.String)
			if urlErr != nil {
				utils.RespondError(w, http.StatusInternalServerError, urlErr, "failed to get image url for people")
				return
			}
			people[i].ImageURL = url
		}
	}

	utils.RespondJSON(w, http.StatusOK, people)
}

func GetPeopleCreditsHistory(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "failed to parse id")
		return
	}

	history, err := dbhelpers.GetPersonCreditsHistory(id)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "failed to get person credits history")
		return
	}
	for i := 0; i < len(history); i++ {
		if history[i].ImagePath.String != "" {
			url, urlErr := utils.GenerateSignedURL(history[i].ImagePath.String)
			if urlErr != nil {
				utils.RespondError(w, http.StatusInternalServerError, urlErr, "failed to get image url for project")
				return
			}
			history[i].ImageURL = url
		}
	}

	utils.RespondJSON(w, http.StatusOK, history)
}

func GetOrgCreditsHistory(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "failed to parse id")
		return
	}

	history, err := dbhelpers.GetOrgCreditsHistory(id)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "failed to get person credits history")
		return
	}
	for i := 0; i < len(history); i++ {
		if history[i].ImagePath.String != "" {
			url, urlErr := utils.GenerateSignedURL(history[i].ImagePath.String)
			if urlErr != nil {
				utils.RespondError(w, http.StatusInternalServerError, urlErr, "failed to get image url for project")
				return
			}
			history[i].ImageURL = url
		}
	}

	utils.RespondJSON(w, http.StatusOK, history)
}

func EditOrgProfile(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "failed to parse id")
		return
	}

	var body models.EditOrgProfileInput
	parseErr := utils.ParseBody(r.Body, &body)
	if parseErr != nil {
		utils.RespondError(w, http.StatusBadRequest, parseErr, "failed to parse body")
		return
	}
	input := models.EditOrgProfile{
		Name:     body.Name,
		Address:  body.Address,
		UploadID: body.UploadID,
		Size:     body.Size,
	}

	updErr := dbhelpers.UpdateOrgProfileByID(id, &input)
	if updErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, updErr, "failed to edit profile")
		return
	}

	utils.RespondJSON(w, http.StatusOK, "profile edited successfully")
}

func EditPersonProfile(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "failed to parse id")
		return
	}

	var body models.EditPersonProfileInput
	parseErr := utils.ParseBody(r.Body, &body)
	if parseErr != nil {
		utils.RespondError(w, http.StatusBadRequest, parseErr, "failed to parse body")
		return
	}
	input := models.EditPersonProfile{
		Name:     body.Name,
		Address:  body.Address,
		UploadID: body.UploadID,
		OrgName:  body.OrgName,
		OrgID:    body.OrgID,
	}

	var oid uuid.NullUUID
	if body.OrgID.Valid {
		oid.UUID = body.OrgID.UUID
		oid.Valid = true
	} else if body.OrgName != "" {
		oid.UUID, err = dbhelpers.CreateOrganization(body.OrgName)
		oid.Valid = true
		if err != nil {
			utils.RespondError(w, http.StatusInternalServerError, err, "failed to get organization id")
			return
		}
	}

	updErr := dbhelpers.UpdatePersonProfileByID(id, &input)
	if updErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, updErr, "failed to edit profile")
		return
	}
	linkErr := dbhelpers.CreateUserOrganizationLink(id, oid)
	if linkErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, linkErr, "failed to edit user organisation")
		return
	}

	utils.RespondJSON(w, http.StatusOK, "profile edited successfully")
}
