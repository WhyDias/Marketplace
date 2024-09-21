// internal/services/user_service.go

package services

import (
	"errors"
	"github.com/WhyDias/Marketplace/internal/db"
	"github.com/WhyDias/Marketplace/internal/models"

	"golang.org/x/crypto/bcrypt"
	"time"
)

// UserService структура сервиса пользователей
type UserService struct{}

// NewUserService конструктор сервиса пользователей
func NewUserService() *UserService {
	return &UserService{}
}

// RegisterUser регистрирует нового пользователя
func (s *UserService) RegisterUser(username, password string) (*models.User, error) {
	// Проверка, существует ли пользователь с таким именем
	existingUser, err := db.GetUserByUsername(username)
	if err == nil && existingUser != nil {
		return nil, errors.New("user already exists")
	}

	// Хеширование пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// Создание нового пользователя
	user := &models.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = db.CreateUser(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// AuthenticateUser аутентифицирует пользователя по имени и паролю
func (s *UserService) AuthenticateUser(username, password string) (*models.User, error) {
	user, err := db.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}

	// Сравнение хеша пароля
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid password")
	}

	return user, nil
}
