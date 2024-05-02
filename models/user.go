package models

import (
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/volatiletech/null"
	"time"
)

type UserAccountType string

const (
	UserAccountTypeIndividual   UserAccountType = "individual"
	UserAccountTypeCorporate    UserAccountType = "corporate"
	UserAccountTypeSME          UserAccountType = "small business"
	UserAccountTypeFarmer       UserAccountType = "farmer"
	UserAccountTypeAdmin        UserAccountType = "admin"
	UserAccountTypeCSinkManager UserAccountType = "c_sink_manager"
)

type VideoType string

const (
	videoTypeFarming = "farming"
	videoTypeBiochar = "biochar"
)

type RegisterUserRequest struct {
	Name        string          `json:"name" db:"name"`
	Email       string          `json:"email" db:"email"`
	Number      null.String     `json:"number" db:"number"`
	CountryCode null.String     `json:"countryCode" db:"country_code"`
	Address     string          `json:"address" db:"address"`
	Password    string          `json:"password" db:"password"`
	AccountType UserAccountType `json:"accountType" db:"account_type"`
	OrgDetails  string          `json:"orgDetails"`
}

type User struct {
	Name           string          `json:"name" db:"name"`
	Email          string          `json:"email" db:"email"`
	Number         null.String     `json:"number" db:"number"`
	CountryCode    null.String     `json:"countryCode" db:"country_code"`
	Address        string          `json:"address" db:"address"`
	Password       string          `json:"password" db:"password"`
	AccountType    UserAccountType `json:"accountType" db:"account_type"`
	OrganizationID uuid.UUID       `json:"orgID" db:"organization_id"`
	ImagePath      null.String     `json:"imagePath,omitempty" db:"path"`
	ImageURL       string          `json:"imageURL,omitempty"`
}

type LoginCredentials struct {
	Email    string `json:"email" db:"email"`
	Password string `json:"password" db:"password"`
}

type EnquiryRequest struct {
	Email       string `json:"email"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	QueryString string `json:"queryString"`
}

type UserInfo struct {
	UserID      uuid.UUID       `json:"userID" db:"id"`
	AccountType UserAccountType `json:"accountType" db:"account_type"`
	ImagePath   null.String     `json:"imagePath,omitempty" db:"path"`
	ImageURL    string          `json:"imageURL"`
}

type UserSessionInfo struct {
	ID          uuid.UUID       `json:"-" db:"id"`
	Email       string          `json:"-" db:"email"`
	Number      null.String     `json:"-" db:"number"`
	AccountType UserAccountType `json:"-" db:"account_type"`
}

type UserBasicInfo struct {
	ID   uuid.UUID `json:"id" db:"id"`
	Name string    `json:"name" db:"name"`
}

type Organization struct {
	ID   uuid.UUID `json:"id" db:"id"`
	Name string    `json:"name" db:"name"`
}

type Session struct {
	UserID uuid.UUID `json:"userID" db:"user_id"`
	Expiry time.Time `json:"expiry" db:"expiry_time"`
}

type OTP struct {
	Input         string    `json:"input" db:"input"`
	OTP           string    `json:"OTP" db:"otp"`
	Type          string    `json:"type" db:"type"`
	Expiry        time.Time `json:"expiry" db:"expiry"`
	ResetPassword bool      `json:"resetPassword"`
}

type FarmerClaims struct {
	FarmerID string `json:"farmerId"`
	jwt.StandardClaims
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type KilnOperatorClaims struct {
	KilnOperatorID string `json:"KilnOperatorId"`
	jwt.StandardClaims
}

type NumberRequest struct {
	PhoneNumber string `json:"phone"`
	CountryCode string `json:"countryCode"`
}

type SendOTP struct {
	PhoneNumber string    `json:"phoneNumber" db:"phone_no"`
	CountryCode string    `json:"countryCode" db:"country_code"`
	OTP         string    ` db:"otp"`
	Type        string    `db:"type"`
	Expiry      time.Time ` db:"expiry"`
}
