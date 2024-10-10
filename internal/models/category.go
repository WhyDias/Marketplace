// internal/models/category.go

package models

import (
	"database/sql"
	"encoding/json"
)

type SupplierCategory struct {
	SupplierID int `json:"supplier_id"`
	CategoryID int `json:"category_id"`
}

type Category struct {
	ID       int           `json:"id"`
	Name     string        `json:"name"`
	Path     string        `json:"path"`
	ImageURL string        `json:"image_url"`
	ParentID sql.NullInt64 `json:"parent_id"` // Используем sql.NullInt64
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
	Description  *string         `json:"description"`
	TypeOfOption *string         `json:"type_of_option"`
	Value        json.RawMessage `json:"value"`
	IsLinked     bool            `json:"is_linked"`
}

type CategoryAttributeResponse struct {
	ID           int         `json:"id"`
	Name         string      `json:"name"`
	Description  *string     `json:"description,omitempty"`    // *string с omitempty
	TypeOfOption *string     `json:"type_of_option,omitempty"` // *string с omitempty
	Value        interface{} `json:"value"`
	IsLinked     bool        `json:"is_linked"`
}

type AddCategoryAttributesRequest struct {
	CategoryID int                `json:"category_id" binding:"required"`
	Attributes []AttributeRequest `json:"attributes" binding:"required,dive,required"`
}

type AttributeRequest struct {
	Name         string      `json:"name" binding:"required"`
	Description  string      `json:"description" binding:"required"`
	TypeOfOption string      `json:"type_of_option" binding:"required,oneof=dropdown range switcher text numeric"`
	Value        interface{} `json:"value,omitempty"`
	IsLinked     bool        `json:"is_linked"`
}
