package middlewares

import (
	"circonomy-server/dbhelpers"
	"circonomy-server/models"
	"circonomy-server/utils"
	"context"
	"github.com/golang-jwt/jwt"
	"github.com/pkg/errors"
	"net/http"
	"os"
	"strings"
	"time"
)

type userContextType string

const (
	UserContextKey userContextType = "user_context"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "public") {
			next.ServeHTTP(w, r)
		} else {
			var validSession models.Session
			sessionToken := r.Header.Get("token")
			validSession, err := dbhelpers.GetSession(sessionToken)
			if err != nil {
				utils.RespondError(w, http.StatusUnauthorized, err, "failed to get session details")
				return
			}
			if validSession.Expiry.Before(time.Now()) {
				execErr := dbhelpers.DelSession(sessionToken)
				if execErr != nil {
					utils.RespondError(w, http.StatusInternalServerError, execErr, "failed to delete expired session")
					return
				}
				utils.RespondJSON(w, http.StatusUnauthorized, "session expired")
				return
			}

			userInfo, err := dbhelpers.GetUserByUserID(validSession.UserID)
			if err != nil {
				utils.RespondError(w, http.StatusInternalServerError, err, "failed to load user session")
				return
			}
			r = r.WithContext(context.WithValue(r.Context(), UserContextKey, &userInfo))

			next.ServeHTTP(w, r)
		}
	})
}

func GetUserContext(req *http.Request) *models.UserSessionInfo {
	return req.Context().Value(UserContextKey).(*models.UserSessionInfo)
}

func FarmerAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bypass := BypassMiddleware(r)
		if bypass {
			next.ServeHTTP(w, r)
		} else {
			token := r.Header.Get("authorization")
			claims := &models.FarmerClaims{}
			if token == "" {
				utils.RespondError(w, http.StatusUnauthorized, errors.New("token not sent in header"), "token not sent in header")
				return
			} else {
				parseToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
					return []byte(os.Getenv("jwtSecret")), nil
				})
				if err != nil {
					if err == jwt.ErrSignatureInvalid {
						utils.RespondError(w, http.StatusUnauthorized, err, "Invalid token signature")
						return
					}
					utils.RespondError(w, http.StatusUnauthorized, err, "Token is expired")
					return
				}
				if !parseToken.Valid {
					utils.RespondError(w, http.StatusUnauthorized, err, "Invalid token")
					return
				}
				if claims.FarmerID == "" {
					utils.RespondError(w, http.StatusUnauthorized, err, "Invalid token")
					return
				}
				ctx := context.WithValue(r.Context(), UserContextKey, claims)
				next.ServeHTTP(w, r.WithContext(ctx))
			}
		}
	})
}

func BypassMiddleware(r *http.Request) bool {
	switch r.URL.Path {
	case "/farmer/video-content":
		videoType := r.URL.Query().Get("videoType")
		if videoType == "biochar" {
			return false
		}
		return true
	}
	return false
}

// KilnOperatorAuthMiddleware authentication and setting kiln operator id in context
func KilnOperatorAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("authorization")
		claims := &models.KilnOperatorClaims{}
		if token == "" {
			utils.RespondError(w, http.StatusUnauthorized, errors.New("token not sent in header"), "token not sent in header")
			return
		} else {
			parseToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(os.Getenv("jwtSecret")), nil
			})
			if err != nil {
				if err == jwt.ErrSignatureInvalid {
					utils.RespondError(w, http.StatusUnauthorized, err, "Invalid token signature")
					return
				}
				utils.RespondError(w, http.StatusUnauthorized, err, "Token is expired")
				return
			}
			if !parseToken.Valid {
				utils.RespondError(w, http.StatusUnauthorized, err, "Invalid token")
				return
			}
			if claims.KilnOperatorID == "" {
				utils.RespondError(w, http.StatusUnauthorized, err, "Invalid token")
				return
			}
			ctx := context.WithValue(r.Context(), UserContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}
