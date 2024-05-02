package handlers

import (
	"circonomy-server/dbhelpers"
	"circonomy-server/utils"
	"net/http"
)

func GetOrgClientsList(w http.ResponseWriter, r *http.Request) {
	list, err := dbhelpers.GetAllOrgClients()
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to get all organisation clients")
		return
	}

	for i := 0; i < len(list); i++ {
		if list[i].ImagePath.String != "" {
			url, urlErr := utils.GenerateSignedURL(list[i].ImagePath.String)
			if urlErr != nil {
				utils.RespondError(w, http.StatusInternalServerError, urlErr, "failed to get url for image")
				return
			}
			list[i].ImageURL = url
		}
	}

	utils.RespondJSON(w, http.StatusOK, list)
}

func GetIndividualClientsList(w http.ResponseWriter, r *http.Request) {
	list, err := dbhelpers.GetAllIndividualClients()
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to get all organisation clients")
		return
	}

	for i := 0; i < len(list); i++ {
		if list[i].ImagePath.String != "" {
			url, urlErr := utils.GenerateSignedURL(list[i].ImagePath.String)
			if urlErr != nil {
				utils.RespondError(w, http.StatusInternalServerError, urlErr, "failed to get url for image")
				return
			}
			list[i].ImageURL = url
		}
	}

	utils.RespondJSON(w, http.StatusOK, list)
}
