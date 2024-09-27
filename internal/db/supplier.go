// internal/db/supplier.go

package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/WhyDias/Marketplace/internal/models"
)

// FetchSupplierByUserID получает поставщика по user_id
func FetchSupplierByUserID(userID int) (*models.Supplier, error) {
	supplier := &models.Supplier{}

	query := `
        SELECT id, user_id, phone_number, is_verified, created_at, updated_at
        FROM supplier
        WHERE user_id = $1
        LIMIT 1
    `

	err := DB.QueryRow(query, userID).Scan(
		&supplier.ID,
		&supplier.UserID,
		&supplier.PhoneNumber,
		&supplier.IsVerified,
		&supplier.CreatedAt,
		&supplier.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("FetchSupplierByUserID: не найден поставщик для user_id %d", userID)
			return nil, fmt.Errorf("поставщик не найден для user_id %d", userID)
		}
		log.Printf("FetchSupplierByUserID: ошибка при выполнении запроса для user_id %d: %v", userID, err)
		return nil, fmt.Errorf("ошибка при выполнении запроса: %v", err)
	}

	log.Printf("FetchSupplierByUserID: получен поставщик id=%d для user_id %d", supplier.ID, userID)
	return supplier, nil
}

func GetAllCategories() ([]models.Category, error) {
	query := `SELECT id, name, path, image_url FROM categories`

	rows, err := DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка при выполнении запроса: %v", err)
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var category models.Category
		if err := rows.Scan(&category.ID, &category.Name, &category.Path, &category.ImageURL); err != nil {
			log.Printf("GetAllCategories: ошибка при сканировании строки: %v", err)
			return nil, fmt.Errorf("ошибка при сканировании строки: %v", err)
		}
		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		log.Printf("GetAllCategories: ошибка при итерации по строкам: %v", err)
		return nil, fmt.Errorf("ошибка при итерации по строкам: %v", err)
	}

	return categories, nil
}
