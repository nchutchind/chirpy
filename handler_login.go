package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/nchutchind/chirpy/internal/auth"
	"github.com/nchutchind/chirpy/internal/database"
)

func (cfg *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	user, err := cfg.db.GetUser(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get user", err)
		return
	}

	passwordMatch, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to check password", err)
		return
	}
	if !passwordMatch {
		respondWithError(w, http.StatusUnauthorized, "Invalid credentials", nil)
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.jwtSecret, sessionExpireDuration)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create JWT", err)
		return
	}
	refreshTokenStr, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create refresh token", err)
		return
	}

	refreshToken, err :=cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		UserID: user.ID,
		Token:  refreshTokenStr,
		ExpiresAt: time.Now().UTC().Add(sessionExpireDuration),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to store refresh token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, User{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		IsChirpyRed:  user.IsChirpyRed,
		Token:        token,
		RefreshToken: refreshToken.Token,
	})
}