// internal/models/product.go

package models

type Product struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	CategoryID int    `json:"category_id"`
	MarketID   int    `json:"market_id"`
	StatusID   int    `json:"status_id"`
	SupplierID int    `json:"supplier_id"`
}
