package handlers

import (
	"circonomy-server/dbhelpers"
	"circonomy-server/family"
	"circonomy-server/models"
	"circonomy-server/providers"
	"circonomy-server/utils"
	"database/sql"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

var ExpiryTime time.Duration = 24
var OTPExpiryMinutes time.Duration = 10
var MaxMemorySize int64 = 32
var MaxMemoryLimit int64 = 20
var storedOTP = "666666"

var FamilyService *family.Service
var EmailProvider providers.EmailProvider

const enquiryEmail = "hello@circonomy.co"
const enquiryName = "Kul Kovid"

func SignUp(w http.ResponseWriter, r *http.Request) {
	var body models.RegisterUserRequest
	if err := utils.ParseBody(r.Body, &body); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "Failed to parse body")
		return
	}

	// hash password
	passwordHash, err := utils.HashPassword(body.Password)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to secure password")
		return
	}
	var uID uuid.UUID
	body.Password = passwordHash
	if body.AccountType == models.UserAccountTypeIndividual {
		user := models.User{
			Name:        body.Name,
			Email:       body.Email,
			Address:     body.Address,
			Password:    body.Password,
			AccountType: body.AccountType,
			Number:      body.Number,
			CountryCode: body.CountryCode,
		}
		userID, createErr := dbhelpers.CreateIndividual(&user)
		if createErr != nil {
			utils.RespondError(w, http.StatusInternalServerError, createErr, "failed to register individual user")
			return
		}
		uID = userID
	} else if body.AccountType == models.UserAccountTypeCorporate || body.AccountType == models.UserAccountTypeSME {
		_, userID, createErr := dbhelpers.CreateBusiness(&body)
		if createErr != nil {
			utils.RespondError(w, http.StatusInternalServerError, createErr, "failed to register organization")
			return
		}
		uID = userID
	}
	sessionToken := utils.HashString(body.Email + time.Now().String())
	Expires := time.Now().Add(time.Hour * ExpiryTime)
	sessionErr := dbhelpers.CreateUserSession(uID, sessionToken, Expires)
	if sessionErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "failed to create session")
		return
	}
	user := models.UserInfo{
		UserID:      uID,
		AccountType: body.AccountType,
	}
	invitations, err := FamilyService.GetInvitations(uID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "failed to load invitations")
		return
	}
	utils.RespondJSON(w, http.StatusOK, struct {
		Token       string                 `json:"token"`
		UserInfo    models.UserInfo        `json:"userInfo"`
		Invitations []*family.OwnerDetails `json:"invitations"`
	}{
		Token:       sessionToken,
		UserInfo:    user,
		Invitations: invitations,
	})
}

func Login(w http.ResponseWriter, r *http.Request) {
	var credentials models.LoginCredentials
	if err := utils.ParseBody(r.Body, &credentials); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "Failed to parse login credentials")
		return
	}

	storedPassword, err := dbhelpers.GetPasswordByEmail(credentials.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.RespondError(w, http.StatusBadRequest, err, "email not registered")
			return
		}
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch password")
		return
	}
	checkErr := utils.CheckPassword(credentials.Password, storedPassword)
	if checkErr != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "incorrect password")
		return
	}
	user, err := dbhelpers.GetUserByEmailID(credentials.Email)
	if user.ImagePath.Valid {
		url, urlErr := utils.GenerateSignedURL(user.ImagePath.String)
		if urlErr != nil {
			utils.RespondError(w, http.StatusInternalServerError, urlErr, "failed to get url for image")
			return
		}
		user.ImageURL = url
	}
	// Create user session
	sessionToken := utils.HashString(credentials.Email + time.Now().String())
	Expires := time.Now().Add(time.Hour * ExpiryTime)
	sessionErr := dbhelpers.CreateUserSession(user.UserID, sessionToken, Expires)
	if sessionErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "failed to create session")
		return
	}
	utils.RespondJSON(w, http.StatusOK, struct {
		Token    string          `json:"token"`
		UserInfo models.UserInfo `json:"userInfo"`
	}{
		Token:    sessionToken,
		UserInfo: user,
	})
}

func ContactUs(w http.ResponseWriter, r *http.Request) {
	var request models.EnquiryRequest
	if err := utils.ParseBody(r.Body, &request); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "Failed to parse request")
		return
	}

	template, _ := EmailProvider.GetEmailTemplate(providers.EmailTypeContactUs)
	template.AddRecipient(enquiryName, enquiryEmail)
	template.DynamicData["firstName"] = request.FirstName
	template.DynamicData["lastName"] = request.LastName
	template.DynamicData["email"] = request.Email
	template.DynamicData["queryString"] = request.QueryString
	EmailProvider.Send(template)

	err := dbhelpers.InsertEnquiry(request, providers.EmailTypeContactUs)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to store enquiry")
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func Subscribe(w http.ResponseWriter, r *http.Request) {
	var request models.EnquiryRequest
	if err := utils.ParseBody(r.Body, &request); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "Failed to parse request")
		return
	}

	template, _ := EmailProvider.GetEmailTemplate(providers.EmailTypeSubscribe)
	template.AddRecipient(enquiryName, enquiryEmail)
	template.DynamicData["firstName"] = request.FirstName
	template.DynamicData["lastName"] = request.LastName
	template.DynamicData["email"] = request.Email
	EmailProvider.Send(template)

	err := dbhelpers.InsertEnquiry(request, providers.EmailTypeSubscribe)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to store enquiry")
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func DoesEmailExist(w http.ResponseWriter, r *http.Request) {
	email := chi.URLParam(r, "email")
	exist, err := dbhelpers.DoesEmailExist(email)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to check if email exists")
		return
	}
	if exist {
		utils.RespondJSON(w, http.StatusConflict, "given email is already registered")
		return
	}

	utils.RespondJSON(w, http.StatusOK, "email is not already registered")
}

func DoesNumberExist(w http.ResponseWriter, r *http.Request) {
	var request models.NumberRequest
	if err := utils.ParseBody(r.Body, &request); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "Failed to parse request")
		return
	}
	exist, err := dbhelpers.DoesNumberExist(request.PhoneNumber, request.CountryCode)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to check if number exists")
		return
	}
	if exist {
		utils.RespondJSON(w, http.StatusConflict, "given number is already registered")
		return
	}

	utils.RespondJSON(w, http.StatusOK, "number is not already registered")
}

func GetOrgList(w http.ResponseWriter, _ *http.Request) {
	list, err := dbhelpers.GetAllOrgNames()
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to get all organisation names")
		return
	}

	utils.RespondJSON(w, http.StatusOK, list)
}

func GetIndividualList(w http.ResponseWriter, _ *http.Request) {
	list, err := dbhelpers.GetAllIndividualNames()
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to get all individual user names")
		return
	}

	utils.RespondJSON(w, http.StatusOK, list)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	sessionToken := r.Header.Get("token")
	execErr := dbhelpers.DelSession(sessionToken)
	if execErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, execErr, "failed to logout session")
		return
	}
	utils.RespondJSON(w, http.StatusOK, "logged out successfully")
}

func CheckPassword(w http.ResponseWriter, r *http.Request) {
	var credentials models.LoginCredentials
	if err := utils.ParseBody(r.Body, &credentials); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "Failed to parse login credentials")
		return
	}

	storedPassword, err := dbhelpers.GetPasswordByEmail(credentials.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.RespondError(w, http.StatusBadRequest, err, "email not registered")
			return
		}
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch password")
		return
	}
	checkErr := utils.CheckPassword(credentials.Password, storedPassword)
	if checkErr != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "incorrect password")
		return
	}

	utils.RespondJSON(w, http.StatusOK, "password checked successfully")
}

func ResetPassword(w http.ResponseWriter, r *http.Request) {
	var body models.LoginCredentials
	parseErr := utils.ParseBody(r.Body, &body)
	if parseErr != nil {
		utils.RespondError(w, http.StatusBadRequest, parseErr, "failed to parse body")
		return
	}

	// hash password
	passwordHash, err := utils.HashPassword(body.Password)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to secure password")
		return
	}
	body.Password = passwordHash

	err = dbhelpers.UpdatePassword(body)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to update password")
		return
	}

	utils.RespondJSON(w, http.StatusOK, "password changed successfully")
}

func SendOTP(w http.ResponseWriter, r *http.Request) {
	otp := models.OTP{}
	parseErr := utils.ParseBody(r.Body, &otp)
	if parseErr != nil {
		utils.RespondError(w, http.StatusBadRequest, parseErr, "failed to parse body")
		return
	}

	otp.Expiry = time.Now().Add(time.Minute * OTPExpiryMinutes)
	if utils.IsDevEnvironment() {
		otp.OTP = "666666"
	} else {
		otp.OTP = utils.EncodeToString(6) // replace with generatedOTP
	}
	err := dbhelpers.StoreOTP(otp)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "failed to store otp")
		return
	}

	if !utils.IsDevEnvironment() {
		if !otp.ResetPassword {
			template, _ := EmailProvider.GetEmailTemplate(providers.EmailTypeVerifyEmail)
			template.AddRecipient("", otp.Input)
			template.DynamicData["otp"] = otp.OTP
			EmailProvider.Send(template)
		} else {
			template, _ := EmailProvider.GetEmailTemplate(providers.EmailTypeResetPassword)
			template.AddRecipient("", otp.Input)
			template.DynamicData["otp"] = otp.OTP
			EmailProvider.Send(template)
		}
	}

	utils.RespondJSON(w, http.StatusOK, "otp sent")
}

func CheckOTP(w http.ResponseWriter, r *http.Request) {
	otp := models.OTP{}
	parseErr := utils.ParseBody(r.Body, &otp)
	if parseErr != nil {
		utils.RespondError(w, http.StatusBadRequest, parseErr, "failed to parse body")
		return
	}

	// fetch stored otp corresponding to the given input (otp.Input)
	storedOTP, getErr := dbhelpers.GetOTP(otp)
	if getErr != nil {
		if getErr == sql.ErrNoRows {
			utils.RespondJSON(w, http.StatusBadRequest, "incorrect otp")
			return
		}
		utils.RespondError(w, http.StatusInternalServerError, getErr, "failed to fetch stored otp")
		return
	}

	if otp.OTP != storedOTP {
		utils.RespondJSON(w, http.StatusBadRequest, "incorrect otp")
		return
	}

	utils.RespondJSON(w, http.StatusOK, "otp verified")
}

func UploadImage(w http.ResponseWriter, r *http.Request) {
	// Parse the multipart form, up to 32 MiB in memory
	if err := r.ParseMultipartForm(MaxMemorySize << MaxMemoryLimit); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "failed to parse multipartForm")
		return
	}

	// Get the file from the form
	file, handler, err := r.FormFile("image")
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "failed to get image from request")
		return
	}
	defer func(file multipart.File) {
		defErr := file.Close()
		if defErr != nil {
			return
		}
	}(file)
	imgType := r.FormValue("type")
	err = utils.UploadImageToBucket(file, handler)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "failed to upload image to bucket")
	}

	imageID, err := dbhelpers.AddImage(handler.Filename, imgType)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "failed to add image to db")
		return
	}
	url, urlErr := utils.GenerateSignedURL(handler.Filename)
	if urlErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, urlErr, "failed to get image url")
		return
	}
	// Return a success response
	utils.RespondJSON(w, http.StatusOK, struct {
		ImageID  uuid.UUID `json:"imageID"`
		ImageURL string    `json:"imageURL"`
	}{
		ImageID:  imageID,
		ImageURL: url,
	})
}
