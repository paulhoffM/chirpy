package auth

import (
	"testing"
	"time"
	"github.com/google/uuid"
) 

func TestMakeJWT(t *testing.T){
userID := uuid.New()
tokenSecret := "hello"
expiresIn, _ := time.ParseDuration("10s")

token, err := MakeJWT(userID, tokenSecret, expiresIn)
if err != nil {
	t.Fatalf("MakeJWT returned and error: %v", err)
}
if token == "" {
	t.Fatal("MakeJWT returned empty")
}

}
func TestValidateJWT(t *testing.T){
	userID := uuid.New()
	tokenSecret := "hello"
	expiresIn, _ := time.ParseDuration("10s")
	tokenString, err := MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT returned and error: %v", err)
	}
	userValidateID, err := ValidateJWT(tokenString, tokenSecret)
	if err != nil {
		t.Fatalf("ValidateJWT returned and error: %v", err)
	}

	if userID != userValidateID {
		t.Fatal("Could not validate user ID")
	}


}




// func TestHashPassword(t *testing.T) {
// 	password := "correcthorsebatterystaple"
// 	hash, err := HashPassword(password)
// 	if err != nil {
// 		t.Fatalf("HashPassword returned an error: %v", err)
// 	}
// 	if hash == "" {
// 		t.Fatal("HashPassword returned an empty hash")
// 	}
// }

// func TestCheckPasswordHash(t *testing.T) {
// 	password := "correcthorsebatterystaple"
// 	hash, err := HashPassword(password)
// 	if err != nil {
// 		t.Fatalf("HashPassword returned an error: %v", err)
// 	}

// 	matches, err := CheckPasswordHash(password, hash)
// 	if err != nil {
// 		t.Fatalf("CheckPasswordHash returned an error: %v", err)
// 	}
// 	if !matches {
// 		t.Fatal("expected password to match hash")
// 	}

// 	matches, err = CheckPasswordHash("wrongpassword", hash)
// 	if err != nil {
// 		t.Fatalf("CheckPasswordHash returned an error: %v", err)
// 	}
// 	if matches {
// 		t.Fatal("expected password to not match hash")
// 	}
// }