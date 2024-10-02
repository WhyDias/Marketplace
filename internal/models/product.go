// internal/models/product.go

package models

type Product struct {
	ID          int                `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	CategoryID  int                `json:"category_id"`
	MarketID    int                `json:"market_id"`
	StatusID    int                `json:"status_id"`
	SupplierID  int                `json:"supplier_id"`
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

type ProductVariation struct {
	ID         int                     `json:"id"`
	ProductID  int                     `json:"product_id"`
	SKU        string                  `json:"sku"`
	Price      float64                 `json:"price"`
	Stock      int                     `json:"stock"`
	Images     []ProductVariationImage `json:"images,omitempty"`
	Attributes []AttributeValue        `json:"attributes,omitempty"`
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
