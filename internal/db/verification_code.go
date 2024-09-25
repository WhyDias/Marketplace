// internal/db/verification_code.go

package db

import (
	"database/sql"
	"fmt"
	"github.com/WhyDias/Marketplace/internal/models"
	"log"
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

func GetLatestVerificationCode(username string) (*models.VerificationCode, error) {
	verificationCode := &models.VerificationCode{}

	query := `
		SELECT id, phone_number, code, created_at, expires_at
		FROM verification_codes
		WHERE phone_number = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	err := DB.QueryRow(query, username).Scan(
		&verificationCode.ID,
		&verificationCode.PhoneNumber,
		&verificationCode.Code,
		&verificationCode.CreatedAt,
		&verificationCode.ExpiresAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("GetLatestVerificationCode: не найдены коды подтверждения для пользователя %s", username)
			return nil, fmt.Errorf("failed to get verification code: %v", err)
		}
		log.Printf("GetLatestVerificationCode: ошибка при выполнении запроса для пользователя %s: %v", username, err)
		return nil, fmt.Errorf("failed to get verification code: %v", err)
	}

	log.Printf("GetLatestVerificationCode: получен код подтверждения %s для пользователя %s", verificationCode.Code, username)
	return verificationCode, nil
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
