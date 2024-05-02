package utils

import (
	"circonomy-server/models"
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/pkg/errors"
	"io"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"

	"golang.org/x/crypto/bcrypt"

	"github.com/sirupsen/logrus"
	"github.com/teris-io/shortid"
)

var generator *shortid.Shortid

const generatorSeed = 1000

type sortOrder string

const (
	defaultLimit  = 10
	defaultPageNo = 0

	ASC  sortOrder = "ASC"
	DESC sortOrder = "DESC"
)

type clientError struct {
	ID            string `json:"id"`
	MessageToUser string `json:"messageToUser"`
	DeveloperInfo string `json:"developerInfo"`
	Err           string `json:"error"`
	StatusCode    int    `json:"statusCode"`
	IsClientError bool   `json:"isClientError"`
}

func init() {
	n, err := rand.Int(rand.Reader, big.NewInt(generatorSeed))
	if err != nil {
		logrus.Panicf("failed to initialize utilities with random seed, %+v", err)
		return
	}

	g, err := shortid.New(1, shortid.DefaultABC, n.Uint64())

	if err != nil {
		logrus.Panicf("Failed to initialize utils package with error: %+v", err)
	}
	generator = g
}

// ParseBody parses the values from io reader to a given interface
func ParseBody(body io.Reader, out interface{}) error {
	err := json.NewDecoder(body).Decode(out)
	if err != nil {
		return err
	}

	return nil
}

// EncodeJSONBody writes the JSON body to response writer
func EncodeJSONBody(resp http.ResponseWriter, data interface{}) error {
	return json.NewEncoder(resp).Encode(data)
}

// RespondJSON sends the interface as a JSON
func RespondJSON(w http.ResponseWriter, statusCode int, body interface{}) {
	w.WriteHeader(statusCode)
	if body != nil {
		if err := EncodeJSONBody(w, body); err != nil {
			logrus.Errorf("Failed to respond JSON with error: %+v", err)
		}
	}
}

// newClientError creates structured client error response message
func newClientError(err error, statusCode int, messageToUser string, additionalInfoForDevs ...string) *clientError {
	additionalInfoJoined := strings.Join(additionalInfoForDevs, "\n")
	if additionalInfoJoined == "" {
		additionalInfoJoined = messageToUser
	}

	errorID, _ := generator.Generate()
	var errString string
	if err != nil {
		errString = err.Error()
	}
	return &clientError{
		ID:            errorID,
		MessageToUser: messageToUser,
		DeveloperInfo: additionalInfoJoined,
		Err:           errString,
		StatusCode:    statusCode,
		IsClientError: true,
	}
}

// RespondError sends an error message to the API caller and logs the error
func RespondError(w http.ResponseWriter, statusCode int, err error, messageToUser string, additionalInfoForDevs ...string) {
	logrus.Errorf("status: %d, message: %s, err: %+v ", statusCode, messageToUser, err)
	clientErr := newClientError(err, statusCode, messageToUser, additionalInfoForDevs...)
	w.WriteHeader(statusCode)
	if encodeErr := json.NewEncoder(w).Encode(clientErr); encodeErr != nil {
		logrus.Errorf("Failed to send error to caller with error: %+v", encodeErr)
	}
}

// HashString generates SHA256 for a given string
func HashString(toHash string) string {
	sha := sha512.New()
	sha.Write([]byte(toHash))
	return hex.EncodeToString(sha.Sum(nil))
}

// HashPassword returns the bcrypt hash of the password
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedPassword), nil
}

// CheckPassword checks if the provided password is correct or not
func CheckPassword(password, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// ContainsUUID checks if the provided UUID is present in UUID slice or not
func ContainsUUID(uuids []uuid.UUID, targetUUID uuid.UUID) bool {
	for _, UUID := range uuids {
		if UUID == targetUUID {
			return true
		}
	}
	return false
}

func SQLErrorLogger(err error, sql string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	return errors.Wrapf(err, "SQL: %s \n args %+v", sql, args)
}

func EncodeToString(max int) string {
	b := make([]byte, max)
	n, err := io.ReadAtLeast(rand.Reader, b, max)
	if n != max {
		panic(err)
	}
	for i := 0; i < len(b); i++ {
		b[i] = table[int(b[i])%len(table)]
	}
	return string(b)
}

var table = [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}

type Environment string

const (
	EnvironmentDev   = "dev"
	EnvironmentState = "stage"
	EnvironmentProd  = "prod"
)

func GetEnvironment() Environment {
	env := os.Getenv("ENV")
	if env == "" {
		return EnvironmentDev
	}
	return Environment(env)
}

func IsDevEnvironment() bool {
	return GetEnvironment() == EnvironmentDev || GetEnvironment() == ""
}

func getWebURL() string {
	switch GetEnvironment() {
	case EnvironmentDev:
		return "https://circonomy-dev.web.app/"
	}
	return "https://circonomy-dev.web.app/"
}

func GetInviteUrl(invitationID uuid.UUID) string {
	return fmt.Sprintf("%s/invite/%s", getWebURL(), invitationID.String())
}

func GenerateTokenPair(farmerID string) (map[string]string, error) {
	// Create token
	jwtExpirationTime := time.Now().Add(time.Minute * 60).Unix()
	claims := &models.FarmerClaims{
		FarmerID: farmerID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: jwtExpirationTime,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	// The signing string should be secret (a generated UUID works too)
	t, err := token.SignedString([]byte(os.Getenv("jwtSecret")))
	if err != nil {
		return nil, err
	}

	refreshToken := jwt.New(jwt.SigningMethodHS256)
	rtClaims := refreshToken.Claims.(jwt.MapClaims)
	rtClaims["farmerId"] = farmerID
	rtClaims["exp"] = time.Now().Add(time.Hour * 24 * 60).Unix()

	rt, err := refreshToken.SignedString([]byte(os.Getenv("jwtSecret")))
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"token":         t,
		"refresh_token": rt,
	}, nil
}

func RefreshAuthToken(refreshToken models.RefreshTokenRequest) (map[string]string, error) {
	// Parse takes the token string and a function for looking up the key.
	// The latter is especially useful if you use multiple keys for your application.
	// The standard is to use 'kid' in the head of the token to identify
	// which key to use, but the parsed token (head and claims) is provided
	// to the callback, providing flexibility.
	token, err := jwt.Parse(refreshToken.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(os.Getenv("jwtSecret")), nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Get the user record from database or
		// run through your business logic to verify if the user can log in
		newTokenPair, err := GenerateTokenPair(claims["farmerId"].(string))
		if err != nil {
			return nil, err
		}

		return newTokenPair, nil
	}

	return nil, err
}

func GenerateTokenPairKilnOperator(kilnOperatorID string) (map[string]string, error) {
	jwtExpirationTime := time.Now().Add(time.Minute * 60).Unix()
	claims := &models.KilnOperatorClaims{
		KilnOperatorID: kilnOperatorID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: jwtExpirationTime,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString([]byte(os.Getenv("jwtSecret")))
	if err != nil {
		return nil, err
	}

	refreshToken := jwt.New(jwt.SigningMethodHS256)
	rtClaims := refreshToken.Claims.(jwt.MapClaims)
	rtClaims["kilnOperatorId"] = kilnOperatorID
	rtClaims["exp"] = time.Now().Add(time.Hour * 24 * 30).Unix()

	rt, err := refreshToken.SignedString([]byte(os.Getenv("jwtSecret")))
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"token":         t,
		"refresh_token": rt,
	}, nil
}

func RefreshAuthTokenKilnOperator(refreshToken models.RefreshTokenRequest) (map[string]string, error) {
	token, err := jwt.Parse(refreshToken.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("jwtSecret")), nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		newTokenPair, err := GenerateTokenPair(claims["kilnOperatorId"].(string))
		if err != nil {
			return nil, err
		}
		return newTokenPair, nil
	}
	return nil, err
}

func BypassDevCheckSMSNumbers() []string {
	return []string{}
}
