package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paulhoffM/chirpy/internal/auth"
	"github.com/paulhoffM/chirpy/internal/database"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Request does not have a token", err)
		return
	}

	validatedID, err := auth.ValidateJWT(tokenString, cfg.JWTSecret)
	if err != nil {
		respondWithError(w, http.StatusForbidden, "Invalid User", err)
		return
	}

	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not parse chirpID to uuid", err)
		return
	}
	chirp, err := cfg.DB.GetChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Could not find the chirp", err)
		return
	}
	if chirp.UserID != validatedID {
		respondWithError(w, http.StatusForbidden, "Access Forbidden", nil)
		return
	}

	err = cfg.DB.DeleteChirp(r.Context(), database.DeleteChirpParams{
		ID:     chirpID,
		UserID: validatedID,
	})
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not delete Chirp", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, "Chirp Deleted")
}

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not parse chirpID to uuid", err)
		return
	}

	chirpDB, err := cfg.DB.GetChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Could not find the chirp", err)
		return
	}

	chirpResp := Chirp{
		ID:        chirpDB.ID,
		CreatedAt: chirpDB.CreatedAt,
		UpdatedAt: chirpDB.UpdatedAt,
		Body:      chirpDB.Body,
		UserID:    chirpDB.UserID,
	}

	respondWithJSON(w, http.StatusOK, chirpResp)
}

func (cfg *apiConfig) handlerChirpMsg(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could't decode parameters", err)
		return
	}
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Log in again", err)
		return
	}
	validatedID, err := auth.ValidateJWT(tokenString, cfg.JWTSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Log in again", err)
		return
	}
	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}
	params.Body = profaneFilter(params.Body)

	fmt.Println(params.UserID)
	chirpDB, err := cfg.DB.CreateChirp(r.Context(), database.CreateChirpParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Body:      params.Body,
		UserID:    validatedID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could't create record in database", err)
		return
	}
	chirpResp := Chirp{
		ID:        chirpDB.ID,
		CreatedAt: chirpDB.CreatedAt,
		UpdatedAt: chirpDB.UpdatedAt,
		Body:      chirpDB.Body,
		UserID:    chirpDB.UserID,
	}
	respondWithJSON(w, http.StatusCreated, chirpResp)
}

func (cfg *apiConfig) handlerChirpList(w http.ResponseWriter, r *http.Request) {
	sortDirection := r.URL.Query().Get("sort")
	authorID := r.URL.Query().Get("author_id")
	var chirpDBList []database.Chirp
	var err error
	if authorID == "" && sortDirection != "desc" {
		chirpDBList, err = cfg.DB.GetChirps(r.Context())
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Could not get list of chirps", err)
			return
		}
	} else if authorID == "" && sortDirection == "desc" {
		chirpDBList, err = cfg.DB.GetChirpsDesc(r.Context())
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Could not get list of chirps", err)
			return
		}
	} else if authorID != "" && sortDirection != "desc" {
		var ID uuid.UUID
		ID, err = uuid.Parse(authorID)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Coudl not parse ID", err)
			return
		}
		chirpDBList, err = cfg.DB.GetChirpsFromAuthor(r.Context(), ID)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Could not get list of chirps", err)
			return
		}
	} else {
		var ID uuid.UUID
		ID, err = uuid.Parse(authorID)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Coudl not parse ID", err)
			return
		}
		chirpDBList, err = cfg.DB.GetChirpsFromAuthorDesc(r.Context(), ID)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Could not get list of chirps", err)
			return
		}
	}

	var chirpRespList []Chirp
	for _, chirpDB := range chirpDBList {
		chirpResp := Chirp{
			ID:        chirpDB.ID,
			CreatedAt: chirpDB.CreatedAt,
			UpdatedAt: chirpDB.UpdatedAt,
			Body:      chirpDB.Body,
			UserID:    chirpDB.UserID,
		}
		chirpRespList = append(chirpRespList, chirpResp)

	}

	respondWithJSON(w, http.StatusOK, chirpRespList)
}

func profaneFilter(msg string) string {
	words := strings.Split(msg, " ")
	for i, word := range words {
		loweredWord := strings.ToLower(word)
		switch loweredWord {
		case "kerfuffle":
			words[i] = "****"
		case "sharbert":
			words[i] = "****"
		case "fornax":
			words[i] = "****"
		}
	}
	msg = strings.Join(words, " ")
	return msg
}
