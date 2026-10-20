package auth

import (

	// "github.com/google/uuid"

	"github.com/alexedwards/argon2id"
)

// func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error)

// func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error){
// 	claims := &jwt.RegisteredClaims{
// 		Issuer: "chirpy-access"
// 		IssuedAt: time.Now.UTC()
// 		ExipresAt: time.Now.UTC() + expiresIn
// 		Subject: userID.String()
// 	}
// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256(), claims)
// 	ss, err := token.SignedString(tokenSecret)
// 	if err != nil{
// 		respondWithError(w, http.StatusInternalServerError, "Could not sign the token", err)
// 	}
// 	return ss, nil
// }

func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}
	return hash, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}
	return match, nil
}
