// internal/models/supplier.go

package models

import "time"

// Supplier модель для хранения информации о поставщике
type Supplier struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	PhoneNumber string    `json:"phone_number"`
	Categories  []int     `json:"categories"`
	Place       string    `json:"place"`
	MarketID    int       `json:"market_id"`
	RowName     string    `json:"row_name"`
	UserID      int       `json:"user_id"`
	IsVerified  bool      `json:"is_verified"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type RegisterSupplierRequest struct {
	Name        string `json:"name" binding:"required"`
	PhoneNumber string `json:"phone_number" binding:"required,e164"`
	BazaarID    int    `json:"bazaar_id" binding:"required"`
	PlaceName   string `json:"place_name" binding:"required"`
	RowName     string `json:"row_name" binding:"required"`
	Categories  []int  `json:"categories" binding:"required"`
}

type UpdateSupplierRequest struct {
	Name       string `json:"name" binding:"required"`
	MarketName string `json:"market_name" binding:"required"`
	PlacesRows string `json:"places_rows" binding:"required"`
	Category   string `json:"category" binding:"required"`
}

type UpdateSupplierDetailsResponse struct {
	Message string `json:"message"`
}
