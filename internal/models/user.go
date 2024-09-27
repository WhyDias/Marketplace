// internal/models/user.go

package models

import "time"

// User модель пользователя
type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         []string  `json:"role"` // Изменено с string на []string
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type SetPasswordResponse struct {
	Message string `json:"message"`
}
