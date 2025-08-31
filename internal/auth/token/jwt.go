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

func (j *JWTMaker) CreateToken(userID int64, duration time.Duration) (string, error) {
	fmt.Println("Creating token for userID:", userID, "and duration:", duration)
	claims := CustomClaims{
		UserId: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return token.SignedString([]byte(j.SecretKey))
}

func (j *JWTMaker) VerifyToken(tokenStr string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
		// Optional: make sure token uses HMAC SHA512
		// if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
		// 	return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		// }
		if t.Method != jwt.SigningMethodHS512 {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return []byte(j.SecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
