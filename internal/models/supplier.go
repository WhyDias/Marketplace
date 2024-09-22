// internal/models/supplier.go

package models

import "time"

// Supplier модель для хранения информации о поставщике
type Supplier struct {
	ID          int       `json:"id"`
	PhoneNumber string    `json:"phone_number"`
	IsVerified  bool      `json:"is_verified"`
	Name        string    `json:"name,omitempty"`
	MarketID    int       `json:"market_id,omitempty"`
	PlaceID     int       `json:"place_id,omitempty"`
	RowID       int       `json:"row_id,omitempty"`
	Categories  []int     `json:"categories,omitempty"` // Массив ID категорий
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UpdateSupplierRequest struct {
	Name       string `json:"name" binding:"required"`
	MarketName string `json:"market_name" binding:"required"`
	PlacesRows string `json:"places_rows" binding:"required"`
	Category   string `json:"category" binding:"required"`
}

type RegisterSupplierRequest struct {
	Name        string `json:"name" binding:"required"`
	PhoneNumber string `json:"phone_number" binding:"required,e164"`
	MarketName  string `json:"market_name" binding:"required"`
	PlacesRows  string `json:"places_rows" binding:"required"`
	Category    string `json:"category" binding:"required"`
}
