// internal/models/user.go

package models

import "time"

// User модель пользователя
type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Roles        []Role    `json:"roles,omitempty"`
}

type SetPasswordResponse struct {
	Message string `json:"message"`
}
