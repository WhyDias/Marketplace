// internal/models/product.go

package models

import (
	"encoding/json"
	"mime/multipart"
)

type AttributeValueImage struct {
	ID               int      `json:"id"`
	AttributeValueID int      `json:"attribute_value_id"`
	ImageURLs        []string `json:"image_urls"`
}

type AttributeValue struct {
	ID          int    `json:"id"`
	AttributeID int    `json:"attribute_id"`
	Name        string `json:"name"`
	Value       string `json:"value"`
}

type Color struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type UpdateProductRequest struct {
	Name        string                    `json:"name"`
	Description string                    `json:"description"`
	CategoryID  int                       `json:"category_id"`
	Images      []string                  `json:"images"`
	Variations  []ProductVariationRequest `json:"variations"`
}

type ProductCreationRequest struct {
	Name        string                     `json:"name" binding:"required" example:"Мужская футболка"`
	CategoryID  int                        `json:"category_id" binding:"required" example:"5"`
	SKU         string                     `json:"sku" binding:"required" example:"TSHIRT-001"`
	Description string                     `json:"description" binding:"required" example:"Высококачественная мужская футболка из хлопка."`
	Price       float64                    `json:"price" binding:"required" example:"29.99"`
	Images      []ImageUploadRequest       `json:"images" binding:"required,min=1" example:"['image1.jpg', 'image2.jpg']"`
	Variations  []ProductVariationCreation `json:"variations" binding:"required,dive,required"`
}

// ImageUploadRequest представляет структуру для загрузки изображений
type ImageUploadRequest struct {
	File string `json:"file" binding:"required" example:"base64encodedstring"`
}

// ProductVariationCreation представляет структуру для вариаций продукта
type ProductVariationCreation struct {
	SKU        string                      `json:"sku" binding:"required" example:"TSHIRT-001-BLACK-M"`
	Price      float64                     `json:"price" binding:"required" example:"29.99"`
	Stock      int                         `json:"stock" binding:"required" example:"100"`
	Attributes []VariationAttributeRequest `json:"attributes" binding:"required,dive,required"`
}

// VariationAttributeRequest представляет структуру для атрибутов вариации
type VariationAttributeRequest struct {
	Name  string `json:"name" binding:"required" example:"Цвет"`
	Value string `json:"value" binding:"required" example:"Черный"`
}

type ProductRequest struct {
	Name           string                  `form:"name" binding:"required"`
	CategoryID     int                     `form:"category_id" binding:"required"`
	Description    string                  `form:"description"`
	Price          float64                 `form:"price" binding:"required"`
	Stock          int                     `form:"stock" binding:"required"`
	Images         []*multipart.FileHeader `form:"images" binding:"required"`
	Attributes     []AttributeValueRequest `form:"attributes"`                    // Добавлено поле Attributes
	VariationsJSON string                  `form:"variations" binding:"required"` // JSON строка с данными о вариациях
	VariationFiles []*multipart.FileHeader `form:"variation_images"`              // Файлы изображений для вариаций
}

type ProductVariationReq struct {
	SKU        string                  `form:"sku"`
	Attributes []AttributeValueRequest `form:"attributes"`
	Images     []*multipart.FileHeader `form:"images"`
}

type ProductVariationRequest struct {
	SKU        string                  `json:"sku" binding:"required"`
	Price      float64                 `json:"price" binding:"required"`
	Stock      int                     `json:"stock" binding:"required"`
	Attributes []AttributeValueRequest `json:"attributes"`
	Images     []string                `json:"images"`
}

type AttributeValueRequest struct {
	Name  string      `json:"name" binding:"required"`
	Value interface{} `json:"value" binding:"required"`
}

type VariationAttributeValue struct {
	ID                 int `json:"id"`
	ProductVariationID int `json:"product_variation_id"`
	AttributeValueID   int `json:"attribute_value_id"`
}

type ProductVariationImage struct {
	ID                 int      `json:"id"`                   // Уникальный идентификатор изображения
	ProductVariationID int      `json:"product_variation_id"` // ID вариации продукта, к которой относится изображение
	ImageURLs          []string `json:"image_urls"`           // URL изображения
	ImagePath          string   `json:"image_path"`           // Путь к изображению на сервере
}

type AddProductRequest struct {
	Name        string  `form:"name" binding:"required"`
	CategoryID  int     `form:"category_id" binding:"required"`
	Description string  `form:"description"`
	Price       float64 `form:"price" binding:"required"`
	Stock       int     `form:"stock" binding:"required"`
}

type ProductVariation struct {
	ID         int                       `json:"id"`
	ProductID  int                       `json:"product_id"`
	SKU        string                    `json:"sku"`
	Price      float64                   `json:"price"`
	Stock      int                       `json:"stock"`
	Attributes []VariationAttributeValue `json:"attributes,omitempty"`
	Images     []ProductVariationImage   `json:"images,omitempty"`
}
type ProductStatus struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Attribute struct {
	ID           int             `json:"id"`
	Name         string          `json:"name"`
	Description  string          `json:"description"` // Добавляем описание
	TypeOfOption string          `json:"type_of_option"`
	Value        json.RawMessage `json:"value"`
}

type ProductAttributeValue struct {
	ProductID        int `json:"product_id"`
	AttributeValueID int `json:"attribute_value_id"`
}

type Product struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	CategoryID  int     `json:"category_id"`
	MarketID    int     `json:"market_id"`
	StatusID    int     `json:"status_id"`
	SupplierID  int     `json:"supplier_id"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
}

type ProductImage struct {
	ID        int      `json:"id"`
	ProductID int      `json:"product_id"`
	ImageURLs []string `json:"image_urls"`
	ImagePath string   `json:"image_path"`
}
