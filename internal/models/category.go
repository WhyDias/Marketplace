// internal/models/category.go

package models

// Category модель категории
type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
}
