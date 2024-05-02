package middlewares

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/rs/cors"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

// corsOptions setting up routes for cors
func corsOptions() *cors.Cors {
	return cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "Access-Token", "importDate", "X-Client-Version", "Cache-Control", "Pragma", "x-started-at", "x-api-key", "token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
	})
}

// CommonMiddlewares middleware common for all routes
func CommonMiddlewares() chi.Middlewares {
	return chi.Chain(
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("Content-Type", "application/json")
				next.ServeHTTP(w, r)
			})
		},
		corsOptions().Handler,
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer func() {
					err := recover()
					if err != nil {
						logrus.Errorf("Request Panic err: %v", err)
						jsonBody, _ := json.Marshal(map[string]string{
							"error": "There was an internal server error",
							"trace": fmt.Sprintf("%+v", err),
							"stack": string(debug.Stack()),
						})
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusInternalServerError)
						_, err := w.Write(jsonBody)
						if err != nil {
							logrus.Errorf("Failed to send response from middleware with error: %+v", err)
						}
					}
				}()
				next.ServeHTTP(w, r)
			})
		},
	)
}
