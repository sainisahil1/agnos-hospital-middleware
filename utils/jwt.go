package utils

import (
	"time"
	"github.com/golang-jwt/jwt/v5"
	"agnos-hospital-middleware/models"
)

var jwtKey = []byte("secret_key") // change to env var in prod

type Claims struct {
	Username string
	HospitalID uint
	HospitalName string
	jwt.RegisteredClaims
}

func GenerateJWT(username string, hospital models.Hospital) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: username,
		HospitalID: hospital.ID,
		HospitalName: hospital.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func ValidateJWT(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	return claims, nil
}
