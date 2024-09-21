// internal/services/supplier_service.go

package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/WhyDias/Marketplace/internal/db"
	"github.com/WhyDias/Marketplace/internal/models"
)

// SupplierService структура сервиса поставщиков
type SupplierService struct{}

// NewSupplierService конструктор сервиса поставщиков
func NewSupplierService() *SupplierService {
	return &SupplierService{}
}

// CreateSupplier создает нового поставщика
func (s *SupplierService) CreateSupplier(supplier *models.Supplier) error {
	query := `INSERT INTO suppliers (name, phone_number, market_name, places_rows, category, created_at, updated_at, is_verified) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, created_at, updated_at`

	// Если некоторые поля пусты, устанавливаем значения по умолчанию
	if supplier.Name == "" {
		supplier.Name = "Unnamed Supplier"
	}
	if supplier.MarketName == "" {
		supplier.MarketName = "Unnamed Market"
	}
	if supplier.PlacesRows == "" {
		supplier.PlacesRows = "Not specified"
	}
	if supplier.Category == "" {
		supplier.Category = "Uncategorized"
	}

	err := db.DB.QueryRow(query, supplier.Name, supplier.PhoneNumber, supplier.MarketName, supplier.PlacesRows, supplier.Category, time.Now(), time.Now(), supplier.IsVerified).Scan(&supplier.ID, &supplier.CreatedAt, &supplier.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create supplier: %v", err)
	}

	return nil
}

// GetSupplierInfo получает информацию о поставщике по номеру телефона
func (s *SupplierService) GetSupplierInfo(phoneNumber string) (*models.Supplier, error) {
	supplier := &models.Supplier{}

	query := `SELECT id, name, phone_number, market_name, places_rows, category, created_at, updated_at, is_verified 
	          FROM suppliers WHERE phone_number = $1`

	err := db.DB.QueryRow(query, phoneNumber).Scan(
		&supplier.ID,
		&supplier.Name,
		&supplier.PhoneNumber,
		&supplier.MarketName,
		&supplier.PlacesRows,
		&supplier.Category,
		&supplier.CreatedAt,
		&supplier.UpdatedAt,
		&supplier.IsVerified,
	)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, errors.New("supplier not found")
		}
		return nil, fmt.Errorf("error fetching supplier: %v", err)
	}

	return supplier, nil
}

// GetAllSuppliers получает список всех поставщиков
func (s *SupplierService) GetAllSuppliers() ([]models.Supplier, error) {
	query := `SELECT id, name, phone_number, market_name, places_rows, category, created_at, updated_at, is_verified FROM suppliers`

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error fetching suppliers: %v", err)
	}
	defer rows.Close()

	var suppliers []models.Supplier
	for rows.Next() {
		var supplier models.Supplier
		if err := rows.Scan(&supplier.ID, &supplier.Name, &supplier.PhoneNumber, &supplier.MarketName, &supplier.PlacesRows, &supplier.Category, &supplier.CreatedAt, &supplier.UpdatedAt, &supplier.IsVerified); err != nil {
			return nil, fmt.Errorf("error scanning supplier: %v", err)
		}
		suppliers = append(suppliers, supplier)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %v", err)
	}

	return suppliers, nil
}

// IsPhoneNumberVerified проверяет, верифицирован ли номер телефона
func (s *SupplierService) IsPhoneNumberVerified(phoneNumber string) (bool, error) {
	query := `SELECT is_verified FROM suppliers WHERE phone_number = $1`

	var isVerified bool
	err := db.DB.QueryRow(query, phoneNumber).Scan(&isVerified)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return false, errors.New("supplier not found")
		}
		return false, fmt.Errorf("error checking verification status: %v", err)
	}

	return isVerified, nil
}

type UpdateSupplierRequest struct {
	Name       string `json:"name" binding:"required"`
	MarketName string `json:"market_name" binding:"required"`
	PlacesRows string `json:"places_rows" binding:"required"`
	Category   string `json:"category" binding:"required"`
}

// UpdateSupplierDetails обновляет детали поставщика
func (s *SupplierService) UpdateSupplierDetails(phoneNumber string, req models.UpdateSupplierRequest) error {
	query := `UPDATE suppliers 
	          SET name = COALESCE($1, name), 
	              market_name = COALESCE($2, market_name), 
	              places_rows = COALESCE($3, places_rows), 
	              category = COALESCE($4, category), 
	              updated_at = $5 
	          WHERE phone_number = $6`

	result, err := db.DB.Exec(query, req.Name, req.MarketName, req.PlacesRows, req.Category, time.Now(), phoneNumber)
	if err != nil {
		return fmt.Errorf("failed to update supplier details: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error fetching rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("supplier with phone number %s not found", phoneNumber)
	}

	return nil
}

// MarkPhoneNumberAsVerified обновляет поле is_verified для поставщика
func (s *SupplierService) MarkPhoneNumberAsVerified(phoneNumber string) error {
	query := `UPDATE suppliers SET is_verified = TRUE WHERE phone_number = $1`
	result, err := db.DB.Exec(query, phoneNumber)
	if err != nil {
		return fmt.Errorf("failed to mark phone number as verified: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error fetching rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("supplier with phone number %s not found", phoneNumber)
	}

	return nil
}
