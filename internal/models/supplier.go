// internal/models/supplier.go

package models

import "time"

// Supplier модель для хранения информации о поставщике
type Supplier struct {
	ID          int        `json:"id"`
	Name        string     `json:"name"`
	UserID      int        `json:"user_id"`
	IsVerified  bool       `json:"is_verified"`
	PlaceName   string     `json:"place_name"`
	RowName     string     `json:"row_name"`
	PhoneNumber string     `json:"phone_number"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	MarketID    int        `json:"market_id"`
	Categories  []Category `json:"categories,omitempty"`
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

type Attribute struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	IsLinked   bool   `json:"is_linked"`
	CategoryID int    `json:"category_id"`
	Type       string `json:"type"`
}
