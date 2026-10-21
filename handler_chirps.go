package main

import (
	"fmt"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paulhoffM/chirpy/internal/database"
	"github.com/paulhoffM/chirpy/internal/auth"
)

type Chirp struct{
	ID uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body string `json:"body"`
	UserID uuid.UUID `json:"user_id"`
}
func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request){
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not parse chirpID to uuid", err)
		return
	}
	fmt.Println(chirpID)

	chirpDB, err := cfg.DB.GetChirp(r.Context(),chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Could not find the chirp", err)
		return
	}

	chirpResp := Chirp{
		ID: chirpDB.ID,
		CreatedAt: chirpDB.CreatedAt,
		UpdatedAt: chirpDB.UpdatedAt,
		Body: chirpDB.Body,
		UserID: chirpDB.UserID,
	}

	respondWithJSON(w, http.StatusOK, chirpResp)
}

func (cfg *apiConfig) handlerChirpMsg(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
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
	if err != nil{
		respondWithError(w,http.StatusUnauthorized, "Log in again",err)
	}
	validatedID, err := auth.ValidateJWT(tokenString, cfg.JWTSecret)
	if err != nil{
		respondWithError(w,http.StatusUnauthorized, "Log in again",err)
	}
	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}
	params.Body = profaneFilter(params.Body)
	
	fmt.Println(params.UserID)
	chirpDB, err := cfg.DB.CreateChirp(r.Context(), database.CreateChirpParams{
		ID: uuid.New(),
		CreatedAt:time.Now().UTC(),
		UpdatedAt:time.Now().UTC(),
		Body: params.Body,
		UserID: validatedID,
	})
	if err != nil{
		respondWithError(w, http.StatusInternalServerError, "Could't create record in database", err)
		return
	}
	chirpResp := Chirp{
		ID: chirpDB.ID,
		CreatedAt: chirpDB.CreatedAt,
		UpdatedAt: chirpDB.UpdatedAt,
		Body: chirpDB.Body,
		UserID: chirpDB.UserID,
	}
	respondWithJSON(w, http.StatusCreated, chirpResp)
}

func (cfg *apiConfig) handlerChirpList(w http.ResponseWriter, r *http.Request){
	chirpDBList, err := cfg.DB.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not get list of chirps", err)
		return
	}
	var chirpRespList []Chirp
	for _, chirpDB := range chirpDBList{
		chirpResp := Chirp{
			ID: chirpDB.ID,
			CreatedAt: chirpDB.CreatedAt,
			UpdatedAt: chirpDB.UpdatedAt,
			Body: chirpDB.Body,
			UserID: chirpDB.UserID,
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


