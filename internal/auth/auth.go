package auth

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	pHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(pHash), nil
}

func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func MakeJWT(userID uuid.UUID, tokenSecret []byte, expiresIn time.Duration) (string, error) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": jwt.NewNumericDate(time.Now().Add(expiresIn).UTC()),
		"iss": "chirpy",
		"iat": jwt.NewNumericDate(time.Now().UTC()),
		"sub": userID.String(),
	})
	return token.SignedString(tokenSecret)
}

func ValidateJWT(tokenString string, tokenSecret []byte) (uuid.UUID, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (any, error) {
		return tokenSecret, nil
	})
	if err != nil {
		return uuid.Nil, err
	}
	userID, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}
	return uuid.Parse(userID)
}

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	fields := strings.Split(authHeader, ",")
	for _, field := range fields {
		keyValue := strings.Fields(field)
		if keyValue[0] == "Bearer" {
			return keyValue[1], nil
		}
	}
	return "", fmt.Errorf("bearer token not found")
}
