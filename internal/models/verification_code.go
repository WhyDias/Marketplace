// internal/models/verification_code.go

package models

import "time"

// VerificationCode модель для хранения кодов верификации
type VerificationCode struct {
	ID        int       `json:"id"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	UserID    int       `json:"user_id"`
}
