package main

import (
	"fmt"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paulhoffM/chirpy/internal/database"
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
	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}
	params.Body = profaneFilter(params.Body)
	
	chirpDB, err := cfg.DB.CreateChirp(r.Context(), database.CreateChirpParams{
		ID: uuid.New(),
		CreatedAt:time.Now().UTC(),
		UpdatedAt:time.Now().UTC(),
		Body: params.Body,
		UserID: params.UserID,
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

func respondWithError(w http.ResponseWriter, code int, msg string, err error) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if err != nil {
		log.Println(err)
	}
	if code > 499 {
		log.Printf("Responding with 5XX error: %s", msg)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, errorResponse{
		Error: msg,
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
	}
	w.WriteHeader(code)
	w.Write(dat)
}

/*
curl -i -X POST http://localhost:8080/api/validate_chirp \
  -H "Content-Type: application/json" \
  -d '{"body": "This is a kerfuffle opinion"}'
*/
