package token

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTMaker struct {
	SecretKey string
}

func NewJWTMaker(secret string) *JWTMaker {
	return &JWTMaker{SecretKey: secret}
}

type CustomClaims struct {
	UserId int64  `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func (j *JWTMaker) CreateToken(userID int64, email string, duration time.Duration) (string, error) {
	fmt.Println("Creating token for userID:", userID, "with email:", email, "and duration:", duration)
	claims := CustomClaims{
		UserId: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return token.SignedString([]byte(j.SecretKey))
}
