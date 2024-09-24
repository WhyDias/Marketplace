// pkg/jwt/jwt.go

package jwt

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

// JWTService интерфейс для генерации и валидации токенов
type JWTService interface {
	GenerateToken(userID int) (string, error)
	ValidateToken(tokenString string) (*jwt.Token, error)
}

// JWTCustomClaim экспортируемая структура для кастомных клеймов
type JWTCustomClaim struct {
	UserID int `json:"user_id"`
	jwt.StandardClaims
}

// jwtService реализация интерфейса JWTService
type jwtService struct {
	secretKey string
}

// NewJWTService создает новый JWT сервис
func NewJWTService(secretKey string) JWTService {
	return &jwtService{
		secretKey: secretKey,
	}
}

// ValidateToken валидирует JWT токен
func (j *jwtService) ValidateToken(tokenString string) (*jwt.Token, error) {
	return jwt.ParseWithClaims(tokenString, &JWTCustomClaim{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.secretKey), nil
	})
}

// GenerateToken генерирует JWT токен
func (j *jwtService) GenerateToken(userID int) (string, error) {
	claims := &JWTCustomClaim{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
			Subject:   "user_token",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secretKey))
}
