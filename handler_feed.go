package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/GonGarciaFontenla/rssagg/internal/database"
	"github.com/google/uuid"
)

func (apiCfg *apiConfig) handlerCreateFeed(w http.ResponseWriter, r *http.Request, user database.User) {
	type parameters struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	}

	decoder := json.NewDecoder(r.Body)

	params := parameters{}
	err := decoder.Decode(&params)

	if err != nil {
		respondWithError(w, 400, "Invalid request payload")
		return
	}

	feed, err := apiCfg.DB.CreateFeeds(r.Context(), database.CreateFeedsParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      params.Name,
		Url:       params.URL,
		UserID:    user.ID,
	})

	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error creating new feed: %v", err))
		return
	}

	respondWithJSON(w, 201, dataBaseFeedToFeed(feed))
}

func (apiCfg *apiConfig) handlerGetFeeds(w http.ResponseWriter, r *http.Request) {
	feeds, err := apiCfg.DB.GetFeeds(r.Context())
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error getting feeds: %v", err))
		return
	}

	respondWithJSON(w, 200, dataBaseFeedsToFeed(feeds))
}
