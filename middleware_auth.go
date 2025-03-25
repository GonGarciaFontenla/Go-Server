package main

import (
	"fmt"
	"net/http"

	"github.com/GonGarciaFontenla/rssagg/internal/auth"
	"github.com/GonGarciaFontenla/rssagg/internal/database"
)

type authedHandler func(http.ResponseWriter, *http.Request, database.User)

func (cfg *apiConfig) middlewareAuth(handler authedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey, err := auth.GetAPIKey(r.Header)

		if err != nil {
			respondWithError(w, 403, fmt.Sprintf("Authetication error: %v", err))
			return
		}

		user, err := cfg.DB.GetUserByAPIKey(r.Context(), apiKey)

		if err != nil {
			respondWithError(w, 404, fmt.Sprintf("Error getting the user: %v", err))
			return
		}

		handler(w, r, user)
	}
}
