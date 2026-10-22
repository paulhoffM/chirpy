package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/paulhoffM/chirpy/internal/auth"
	"github.com/paulhoffM/chirpy/internal/database"
)

type Parameters struct {
	Password    string `json:"password"`
	Email       string `json:"email"`
	TokenExpire int    `json:"expires_in_seconds"`
	Event       string `json:"event`
	Data        Data   `json:"data"`
}
type Data struct {
	UserID string `json:"user_id"`
}

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Password     string    `json:"password"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
}

func (cfg *apiConfig) handlerWebhook(w http.ResponseWriter, r *http.Request) {
	apiString, _ := auth.GetAPIKEY(r.Header)
	if apiString == cfg.PolkaKey {
		respondWithError(w, http.StatusUnauthorized, "No access", nil)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := Parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Could't decode request Body", err)
		return
	}
	if params.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	parsedID, err := uuid.Parse(params.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could't parse ID", err)
		return

	}
	err = cfg.DB.UpdateUserToRed(r.Context(), parsedID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "User not found", err)
		return

	}
	respondWithJSON(w, http.StatusNoContent, "")
}

func (cfg *apiConfig) handlerUpdate(w http.ResponseWriter, r *http.Request) {
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Request does not have a token", err)
		return
	}

	validatedID, err := auth.ValidateJWT(tokenString, cfg.JWTSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "No valid token", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := Parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could't decode parameters", err)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create a hash password", err)
		return
	}

	timeUpdate := time.Now().UTC()
	user, err := cfg.DB.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:             validatedID,
		Email:          params.Email,
		HashedPassword: hashedPassword,
		UpdatedAt:      timeUpdate,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update User in database", err)
		return
	}

	respondWithJSON(w, http.StatusOK, User{
		ID:        user.ID,
		UpdatedAt: timeUpdate,
		Email:     params.Email,
	})

}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Request does not have a token", err)
		return
	}

	err = cfg.DB.RevokeRefreshToken(r.Context(), database.RevokeRefreshTokenParams{
		Token:     tokenString,
		UpdatedAt: time.Now().UTC(),
		RevokedAt: sql.NullTime{Time: time.Now().UTC(), Valid: true},
	})
	respondWithJSON(w, http.StatusNoContent, "")
}

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Request does not have a token", err)
		return
	}
	refreshToken, err := cfg.DB.GetUserFromRefreshToken(r.Context(), tokenString)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Token not valid", err)
		return
	}
	if refreshToken.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "Token revoked", err)
		return
	}
	if time.Now().After(refreshToken.ExpiresAt) {
		respondWithError(w, http.StatusUnauthorized, "Token expired", err)
		return
	}
	expiresIn, _ := time.ParseDuration("1h")
	tokenJWT, err := auth.MakeJWT(refreshToken.UserID, cfg.JWTSecret, expiresIn)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create Token", err)
		return
	}
	type responseBody struct {
		Token string `json:"token"`
	}
	responseToken := responseBody{
		Token: tokenJWT,
	}
	respondWithJSON(w, http.StatusOK, responseToken)
}

func (cfg *apiConfig) handlerLoginUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	params := Parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could't decode parameters", err)
		return
	}

	userDB, err := cfg.DB.GetUser(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not fetch user", err)
		return
	}
	match, err := auth.CheckPasswordHash(params.Password, userDB.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't compare hash password with password", err)
		return
	}
	if match != true {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)

	}
	var expiresIn time.Duration
	if params.TokenExpire == 0 || params.TokenExpire > 3600 {
		expiresIn, _ = time.ParseDuration("1h")
	} else {
		expiresIn = time.Duration(params.TokenExpire) * time.Second
	}
	tokenJWT, err := auth.MakeJWT(userDB.ID, cfg.JWTSecret, expiresIn)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create Token", err)
		return
	}

	refreshToken, err := cfg.DB.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     auth.MakeRefreshToken(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    userDB.ID,
		ExpiresAt: time.Now().UTC().Add(time.Hour * time.Duration(1440)),
	})

	userRes := User{
		ID:           userDB.ID,
		CreatedAt:    userDB.CreatedAt,
		UpdatedAt:    userDB.UpdatedAt,
		Email:        userDB.Email,
		Token:        tokenJWT,
		RefreshToken: refreshToken.Token,
		IsChirpyRed:  userDB.IsChirpyRed.Valid,
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
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create a hash password", err)
		return
	}

	userDb, err := cfg.DB.CreateUser(r.Context(), database.CreateUserParams{
		ID:             uuid.New(),
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
		HashedPassword: hashedPassword,
		Email:          params.Email,
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
