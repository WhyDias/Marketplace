// internal/models/supplier.go

package models

import "time"

// Supplier модель для хранения информации о поставщике
type Supplier struct {
	ID          int       `json:"id"`
	PhoneNumber string    `json:"phone_number"`
	IsVerified  bool      `json:"is_verified"`
	Name        string    `json:"name,omitempty"`
	MarketName  string    `json:"market_name,omitempty"`
	PlacesRows  string    `json:"places_rows,omitempty"`
	Category    string    `json:"category,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UpdateSupplierRequest struct {
	Name       string `json:"name" binding:"required"`
	MarketName string `json:"market_name" binding:"required"`
	PlacesRows string `json:"places_rows" binding:"required"`
	Category   string `json:"category" binding:"required"`
}
