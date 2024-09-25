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
	// Проверка существования пользователя
	existingUser, err := db.GetUserByUsername(username)
	if err == nil && existingUser != nil {
		return nil, errors.New("Пользователь уже существует")
	}

	// Хеширование пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("Не удалось хешировать пароль")
	}

	// Создание пользователя
	user := &models.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
		Role:         "supplier",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = db.CreateUser(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// AuthenticateUser аутентифицирует пользователя по имени пользователя и паролю
func (s *UserService) AuthenticateUser(username, password string) (*models.User, error) {
	// Получаем пользователя из базы данных по имени пользователя
	user, err := db.GetUserByUsername(username)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Сравниваем хеш пароля из базы данных с введённым паролем
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid password")
	}

	return user, nil
}

// CheckPhoneExists проверяет, существует ли пользователь с заданным username (номером телефона)
func (s *UserService) CheckPhoneExists(username string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`
	var exists bool
	err := db.DB.QueryRow(query, username).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// CheckUserExists проверяет, существует ли пользователь с заданным username (номером телефона)
func (s *UserService) CheckUserExists(username string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`
	var exists bool
	err := db.DB.QueryRow(query, username).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// ResetPassword устанавливает новый пароль для пользователя
func (s *UserService) ResetPassword(username, newPassword string) error {
	// Хеширование нового пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("Не удалось хешировать пароль")
	}

	// Обновление пароля в базе данных
	query := `UPDATE users SET password_hash = $1, updated_at = $2 WHERE username = $3`
	result, err := db.DB.Exec(query, string(hashedPassword), time.Now(), username)
	if err != nil {
		return errors.New("Не удалось обновить пароль")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.New("Ошибка при обновлении пароля")
	}

	if rowsAffected == 0 {
		return errors.New("Пользователь не найден")
	}

	return nil
}
