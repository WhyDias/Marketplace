// internal/models/category.go

package models

// Category модель категории
type Category struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	ImageURL string `json:"image_url"`
}

type SupplierCategory struct {
	SupplierID int `json:"supplier_id"`
	CategoryID int `json:"category_id"`
}
