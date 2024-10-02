// internal/models/product.go

package models

type Product struct {
	ID          int                `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	CategoryID  int                `json:"category_id"`
	SupplierID  int                `json:"supplier_id"`
	MarketID    int                `json:"market_id"`
	StatusID    int                `json:"status_id"` // Добавляем это поле
	Price       float64            `json:"price"`
	Stock       int                `json:"stock"`
	Images      []ProductImage     `json:"images,omitempty"`
	Variations  []ProductVariation `json:"variations,omitempty"`
}

type AttributeValueImage struct {
	ID               int      `json:"id"`
	AttributeValueID int      `json:"attribute_value_id"`
	ImageURLs        []string `json:"image_urls"`
}

type ProductImage struct {
	ID        int      `json:"id"`
	ProductID int      `json:"product_id"`
	ImageURLs []string `json:"image_urls"`
}

type ProductVariationImage struct {
	ID                 int      `json:"id"`
	ProductVariationID int      `json:"product_variation_id"`
	ImageURLs          []string `json:"image_urls"`
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

type ProductVariation struct {
	ID         int                     `json:"id"`
	ProductID  int                     `json:"product_id"`
	SKU        string                  `json:"sku"`
	Price      float64                 `json:"price"`
	Stock      int                     `json:"stock"`
	Images     []ProductVariationImage `json:"images,omitempty"`
	Attributes []AttributeValue        `json:"attributes,omitempty"`
	Colors     []Color                 `json:"colors,omitempty"` // Добавили поле Colors
}

type UpdateProductRequest struct {
	Name        string                    `json:"name"`
	Description string                    `json:"description"`
	CategoryID  int                       `json:"category_id"`
	Images      []string                  `json:"images"`
	Variations  []ProductVariationRequest `json:"variations"`
}

type ProductVariationRequest struct {
	SKU        string                  `json:"sku" binding:"required"`
	Price      float64                 `json:"price" binding:"required"`
	Stock      int                     `json:"stock" binding:"required"`
	Images     []string                `json:"images"`
	Attributes []AttributeValueRequest `json:"attributes"`
	Colors     []Color                 `json:"colors"` // Добавили поле Colors
}

type AttributeValueRequest struct {
	Name   string   `json:"name" binding:"required"`
	Values []string `json:"values" binding:"required"`
}

type AddProductRequest struct {
	Name        string                  `json:"name" binding:"required"`
	Description string                  `json:"description"`
	CategoryID  int                     `json:"category_id" binding:"required"`
	Images      []string                `json:"images"`
	Attributes  []AttributeValueRequest `json:"attributes"`
	Price       float64                 `json:"price" binding:"required"`
	Stock       int                     `json:"stock" binding:"required"`
}
