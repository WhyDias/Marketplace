// internal/db/verification_code.go

package db

import (
	"database/sql"
	"fmt"
	"github.com/WhyDias/Marketplace/internal/models"
	"time"
)

func CreateVerificationCode(userID int, code string, expiresAt time.Time) error {
	query := `INSERT INTO verification_codes (user_id, code, created_at, expires_at)
              VALUES ($1, $2, $3, $4)`
	_, err := DB.Exec(query, userID, code, time.Now(), expiresAt)
	if err != nil {
		return fmt.Errorf("failed to create verification code: %v", err)
	}
	return nil
}

func GetLatestVerificationCode(userID int) (*models.VerificationCode, error) {
	verificationCode := &models.VerificationCode{}

	query := `
        SELECT id, user_id, code, created_at, expires_at
        FROM verification_codes
        WHERE user_id = $1
        ORDER BY created_at DESC
        LIMIT 1
    `

	err := DB.QueryRow(query, userID).Scan(
		&verificationCode.ID,
		&verificationCode.UserID,
		&verificationCode.Code,
		&verificationCode.CreatedAt,
		&verificationCode.ExpiresAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no verification code found for user ID %d", userID)
		}
		return nil, fmt.Errorf("failed to get verification code: %v", err)
	}

	return verificationCode, nil
}

func DeleteVerificationCodes(userID int) error {
	query := `DELETE FROM verification_codes WHERE user_id = $1`
	_, err := DB.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete verification codes: %v", err)
	}
	return nil
}
