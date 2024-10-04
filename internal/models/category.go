// internal/models/category.go

package models

import (
	"encoding/json"
)

type SupplierCategory struct {
	SupplierID int `json:"supplier_id"`
	CategoryID int `json:"category_id"`
}

type Category struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	ImageURL string `json:"image_url"`
	ParentID *int   `json:"parent_id"` // Изменено с int на *int
}

type CategoryNode struct {
	ID       int            `json:"id"`
	Name     string         `json:"name"`
	Path     string         `json:"path"`
	ImageURL string         `json:"image_url"`
	Children []CategoryNode `json:"children,omitempty"`
}

type CategoryAttribute struct {
	ID           int             `json:"id"`
	CategoryID   int             `json:"category_id"`
	Name         string          `json:"name"`
	Description  string          `json:"description,omitempty"`
	TypeOfOption string          `json:"type_of_option"`
	Value        json.RawMessage `json:"value"` // Используем json.RawMessage для гибкости
}

type CategoryAttributeResponse struct {
	ID           int         `json:"id"`
	Name         string      `json:"name"`
	Description  string      `json:"description,omitempty"`
	TypeOfOption string      `json:"type_of_option"`
	Value        interface{} `json:"value"`
}
