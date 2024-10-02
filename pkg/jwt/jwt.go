// pkg/jwt/jwt.go

package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// JWTService интерфейс для генерации и валидации токенов
type JWTService interface {
	GenerateToken(userID int) (string, error)
	GenerateTokenWithRoles(userID int, roles []string) (string, error) // Новый метод
	ValidateToken(tokenString string) (*jwt.Token, error)
}

// JWTCustomClaim структура для пользовательских claims
type JWTCustomClaim struct {
	UserID int      `json:"user_id"`
	Roles  []string `json:"roles"` // Новое поле для ролей
	jwt.RegisteredClaims
}

// jwtService реализация интерфейса JWTService
type jwtService struct {
	secretKey string
	issuer    string
}

// NewJWTService создает новый JWT сервис
func NewJWTService(secretKey string) JWTService {
	return &jwtService{
		secretKey: secretKey,
		issuer:    "MarketplaceApp",
	}
}

// ValidateToken валидирует JWT токен
func (j *jwtService) ValidateToken(tokenString string) (*jwt.Token, error) {
	// Разбираем токен с пользовательскими данными
	token, err := jwt.ParseWithClaims(tokenString, &JWTCustomClaim{}, func(token *jwt.Token) (interface{}, error) {
		// Проверяем метод подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// Возвращаем секретный ключ
		return []byte(j.secretKey), nil
	})

	if err != nil {
		// Проверяем, истёк ли токен
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, errors.New("token is expired")
			}
		}
		return nil, err
	}

	// Проверяем валидность токена и его claims
	if _, ok := token.Claims.(*JWTCustomClaim); ok && token.Valid {
		return token, nil
	} else {
		return nil, errors.New("invalid token")
	}
}

// GenerateToken генерирует JWT токен
func (j *jwtService) GenerateToken(userID int) (string, error) {
	claims := &JWTCustomClaim{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   "user_token",
			Issuer:    j.issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secretKey))
}

// GenerateTokenWithRoles генерирует JWT токен с ролями
func (j *jwtService) GenerateTokenWithRoles(userID int, roles []string) (string, error) {
	claims := &JWTCustomClaim{
		UserID: userID,
		Roles:  roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(72 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   "user_token",
			Issuer:    j.issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secretKey))
}
