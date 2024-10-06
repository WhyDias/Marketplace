// internal/db/supplier.go

package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

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

func UpdateSupplierDetailsByUserID(userID int, marketID int, place string, rowName string, categoryIDs []int) error {
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("could not begin transaction: %v", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// Update supplier details
	query := `UPDATE supplier SET market_id = $1, place_name = $2, row_name = $3, updated_at = $4 WHERE user_id = $5`
	_, err = tx.Exec(query, marketID, place, rowName, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("could not update supplier details: %v", err)
	}

	// Get supplier ID
	var supplierID int
	err = tx.QueryRow(`SELECT id FROM supplier WHERE user_id = $1`, userID).Scan(&supplierID)
	if err != nil {
		return fmt.Errorf("could not get supplier ID: %v", err)
	}

	// Delete existing categories
	_, err = tx.Exec(`DELETE FROM supplier_categories WHERE supplier_id = $1`, supplierID)
	if err != nil {
		return fmt.Errorf("could not delete existing categories: %v", err)
	}

	// Insert new categories
	for _, categoryID := range categoryIDs {
		_, err = tx.Exec(`INSERT INTO supplier_categories (supplier_id, category_id) VALUES ($1, $2)`, supplierID, categoryID)
		if err != nil {
			return fmt.Errorf("could not insert category: %v", err)
		}
	}

	return nil
}

func GetMarketIDBySupplierID(supplierID int) (int, error) {
	var marketID int
	query := `SELECT market_id FROM supplier WHERE id = $1`
	err := DB.QueryRow(query, supplierID).Scan(&marketID)
	if err != nil {
		return 0, fmt.Errorf("Не удалось получить market_id для supplier_id %d: %v", supplierID, err)
	}
	return marketID, nil
}

func GetCategoriesBySupplierID(supplierID int) ([]models.Category, error) {
	query := `
        SELECT c.id, c.name 
        FROM categories c
        JOIN supplier_categories sc ON sc.category_id = c.id
        WHERE sc.supplier_id = $1
    `
	rows, err := DB.Query(query, supplierID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var category models.Category
		if err := rows.Scan(&category.ID, &category.Name); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}

func GetSupplierIDByUserID(userID int) (int, error) {
	var supplierID int
	query := `SELECT id FROM supplier WHERE user_id = $1`
	err := DB.QueryRow(query, userID).Scan(&supplierID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("Поставщик не найден для user_id %d", userID)
		}
		return 0, err
	}
	return supplierID, nil
}
