// internal/db/verification_code.go

package db

import (
	"fmt"
	"github.com/WhyDias/Marketplace/internal/models"
	"time"
)

func CreateVerificationCode(phoneNumber, code string, expiresAt time.Time) error {
	query := `INSERT INTO verification_codes (phone_number, code, created_at, expires_at)
              VALUES ($1, $2, $3, $4)`
	_, err := DB.Exec(query, phoneNumber, code, time.Now(), expiresAt)
	if err != nil {
		return fmt.Errorf("failed to create verification code: %v", err)
	}
	return nil
}

func GetLatestVerificationCode(phoneNumber string) (*models.VerificationCode, error) {
	query := `SELECT id, phone_number, code, created_at, expires_at
              FROM verification_codes
              WHERE phone_number = $1
              ORDER BY created_at DESC
              LIMIT 1`

	var vc models.VerificationCode
	err := DB.QueryRow(query, phoneNumber).Scan(
		&vc.ID,
		&vc.PhoneNumber,
		&vc.Code,
		&vc.CreatedAt,
		&vc.ExpiresAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get verification code: %v", err)
	}
	return &vc, nil
}

// internal/db/verification_code.go

func DeleteVerificationCodes(phoneNumber string) error {
	query := `DELETE FROM verification_codes WHERE phone_number = $1`
	_, err := DB.Exec(query, phoneNumber)
	if err != nil {
		return fmt.Errorf("failed to delete verification codes: %v", err)
	}
	return nil
}
