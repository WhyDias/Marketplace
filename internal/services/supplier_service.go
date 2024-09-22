// internal/services/supplier_service.go

package services

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
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
	query := `INSERT INTO suppliers (name, phone_number, market_id, place_id, row_id, categories, created_at, updated_at, is_verified)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, created_at, updated_at`

	// Параметры могут включать поле categories, если это массив
	err := db.DB.QueryRow(query, supplier.Name, supplier.PhoneNumber, supplier.MarketID, supplier.PlaceID, supplier.RowID, supplier.Categories, time.Now(), time.Now(), supplier.IsVerified).Scan(&supplier.ID, &supplier.CreatedAt, &supplier.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create supplier: %v", err)
	}

	return nil
}

// GetSupplierInfo получает информацию о поставщике по номеру телефона
func (s *SupplierService) GetSupplierInfo(phoneNumber string) (*models.Supplier, error) {
	supplier := &models.Supplier{}

	query := `SELECT id, name, phone_number, market_id, place_id, row_id, categories, created_at, updated_at, is_verified
	          FROM suppliers WHERE phone_number = $1`

	err := db.DB.QueryRow(query, phoneNumber).Scan(
		&supplier.ID,
		&supplier.Name,
		&supplier.PhoneNumber,
		&supplier.MarketID,
		&supplier.PlaceID,
		&supplier.Categories,
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
	query := `SELECT id, name, phone_number, market_id, place_id, row_id, categories, created_at, updated_at, is_verified FROM suppliers`

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error fetching suppliers: %v", err)
	}
	defer rows.Close()

	var suppliers []models.Supplier
	for rows.Next() {
		var supplier models.Supplier
		if err := rows.Scan(&supplier.ID, &supplier.Name, &supplier.PhoneNumber, &supplier.MarketID, &supplier.PlaceID, &supplier.Categories, &supplier.CreatedAt, &supplier.UpdatedAt, &supplier.IsVerified); err != nil {
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

// GetAllBazaars получает список всех базаров
func (s *SupplierService) GetAllBazaars() ([]models.Bazaar, error) {
	query := `SELECT id, name FROM bazaar`

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error fetching bazaars: %v", err)
	}
	defer rows.Close()

	var bazaars []models.Bazaar
	for rows.Next() {
		var bazaar models.Bazaar
		if err := rows.Scan(&bazaar.ID, &bazaar.Name); err != nil {
			return nil, fmt.Errorf("error scanning bazaar: %v", err)
		}
		bazaars = append(bazaars, bazaar)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %v", err)
	}

	return bazaars, nil
}

// CreatePlace создает новое место
func (s *SupplierService) CreatePlace(place *models.Place) error {
	query := `INSERT INTO places (name) VALUES ($1) RETURNING id`
	err := db.DB.QueryRow(query, place.Name).Scan(&place.ID)
	if err != nil {
		return fmt.Errorf("failed to create place: %v", err)
	}
	return nil
}

// CreateRow создает новый ряд
func (s *SupplierService) CreateRow(row *models.Row) error {
	query := `INSERT INTO rows (name, place_id) VALUES ($1, $2) RETURNING id`
	err := db.DB.QueryRow(query, row.Name, row.PlaceID).Scan(&row.ID)
	if err != nil {
		return fmt.Errorf("failed to create row: %v", err)
	}
	return nil
}

// UpdateSupplierDetails обновляет данные поставщика
func (s *SupplierService) UpdateSupplierDetails(phoneNumber, marketName string, placeID, rowID int, categories []int) error {
	query := `UPDATE suppliers 
	          SET market_name = $1, 
	              place_id = $2, 
	              row_id = $3, 
	              category_ids = $4, 
	              updated_at = $5 
	          WHERE phone_number = $6`

	// Преобразуем массив категорий в строку (например, через запятую)
	categoryIDs := strings.Join(intSliceToStringSlice(categories), ",")

	_, err := db.DB.Exec(query, marketName, placeID, rowID, categoryIDs, time.Now(), phoneNumber)
	if err != nil {
		return fmt.Errorf("failed to update supplier details: %v", err)
	}

	return nil
}

// Функция для преобразования []int в []string
func intSliceToStringSlice(ints []int) []string {
	strs := make([]string, len(ints))
	for i, v := range ints {
		strs[i] = strconv.Itoa(v)
	}
	return strs
}
