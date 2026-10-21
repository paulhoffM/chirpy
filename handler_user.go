package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/paulhoffM/chirpy/internal/database"
	"github.com/paulhoffM/chirpy/internal/auth"
)


type Parameters struct {
	Password string `json:"password"`
	Email string `json:"email"`
	TokenExpire int `json:"expires_in_seconds"`
	}	
	
	type User struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Password string `json:"password"`
		Email     string    `json:"email"`
		Token    string  `json:"token"`
	}
	
	
func (cfg *apiConfig) handlerLoginUser(w http.ResponseWriter, r *http.Request){
	decoder := json.NewDecoder(r.Body)
	params := Parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could't decode parameters", err)
		return
		}	
	
	userDB, err := cfg.DB.GetUser(r.Context(),params.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not fetch user", err)
		return
	}
	match, err := auth.CheckPasswordHash(params.Password, userDB.HashedPassword)
	if err != nil{
		respondWithError(w, http.StatusInternalServerError, "Couldn't compare hash password with password", err)
		return
	}	
	if match != true {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		
	}
	var expiresIn time.Duration
	if params.TokenExpire == 0 || params.TokenExpire > 3600{
		expiresIn, _ = time.ParseDuration("1h")
	}else {
		expiresIn = time.Duration(params.TokenExpire)*time.Second
	}
	tokenJWT, err := auth.MakeJWT(userDB.ID, cfg.JWTSecret, expiresIn)
	if err != nil{
		respondWithError(w, http.StatusInternalServerError, "Could not create Token", err)
		return
	}
	
	userRes := User{
		ID:        userDB.ID,
		CreatedAt: userDB.CreatedAt,
		UpdatedAt: userDB.UpdatedAt,
		Email:     userDB.Email,
		Token: tokenJWT,
	}	
	respondWithJSON(w, http.StatusOK, userRes)
}	

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	params := Parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could't decode parameters", err)
		return
	}
	hashedPassword, err := auth.HashPassword(params.Password)
		if err != nil{
		respondWithError(w, http.StatusInternalServerError, "Couldn't create a hash password", err)
		return
	}


	userDb, err := cfg.DB.CreateUser(r.Context(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		HashedPassword: hashedPassword,
		Email:     params.Email,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could't create record in database", err)
		return
	}
	
	userRes := User{
		ID:        userDb.ID,
		CreatedAt: userDb.CreatedAt,
		UpdatedAt: userDb.UpdatedAt,
		Email:     userDb.Email,
	}

	respondWithJSON(w, http.StatusCreated, userRes)

}
