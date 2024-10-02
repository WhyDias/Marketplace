// internal/db/product.go

package db

import (
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"log"

	"github.com/WhyDias/Marketplace/internal/models"
)

// FetchProductsByStatus получает продукты по status_id
func FetchProductsByStatus(statusID int) ([]models.Product, error) {
	query := `
		SELECT id, name, category_id, market_id, status_id, supplier_id
		FROM product
		WHERE status_id = $1
	`

	rows, err := DB.Query(query, statusID)
	if err != nil {
		log.Printf("FetchProductsByStatus: ошибка при выполнении запроса: %v", err)
		return nil, fmt.Errorf("ошибка при выполнении запроса: %v", err)
	}
	defer rows.Close()

	var products []models.Product

	for rows.Next() {
		var p models.Product
		if err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.CategoryID,
			&p.MarketID,
			&p.StatusID,
			&p.SupplierID,
		); err != nil {
			log.Printf("FetchProductsByStatus: ошибка при сканировании строки: %v", err)
			return nil, fmt.Errorf("ошибка при сканировании строки: %v", err)
		}
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		log.Printf("FetchProductsByStatus: ошибка при итерации по строкам: %v", err)
		return nil, fmt.Errorf("ошибка при итерации по строкам: %v", err)
	}

	log.Printf("FetchProductsByStatus: найдено %d продуктов со статусом %d", len(products), statusID)
	return products, nil
}

func FetchProductsBySupplierAndStatus(supplierID, statusID int) ([]models.Product, error) {
	query := `
		SELECT id, name, category_id, market_id, status_id, supplier_id
		FROM product
		WHERE supplier_id = $1 AND status_id = $2
	`

	rows, err := DB.Query(query, supplierID, statusID)
	if err != nil {
		log.Printf("FetchProductsBySupplierAndStatus: ошибка при выполнении запроса: %v", err)
		return nil, fmt.Errorf("ошибка при выполнении запроса: %v", err)
	}
	defer rows.Close()

	var products []models.Product

	for rows.Next() {
		var p models.Product
		if err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.CategoryID,
			&p.MarketID,
			&p.StatusID,
			&p.SupplierID,
		); err != nil {
			log.Printf("FetchProductsBySupplierAndStatus: ошибка при сканировании строки: %v", err)
			return nil, fmt.Errorf("ошибка при сканировании строки: %v", err)
		}
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		log.Printf("FetchProductsBySupplierAndStatus: ошибка при итерации по строкам: %v", err)
		return nil, fmt.Errorf("ошибка при итерации по строкам: %v", err)
	}

	log.Printf("FetchProductsBySupplierAndStatus: найдено %d продуктов для supplier_id %d и status_id %d", len(products), supplierID, statusID)
	return products, nil
}

func CreateProductTx(tx *sql.Tx, product *models.Product) error {
	query := `
        INSERT INTO product (name, description, category_id, supplier_id)
        VALUES ($1, $2, $3, $4)
        RETURNING id
    `
	err := tx.QueryRow(query, product.Name, product.Description, product.CategoryID, product.SupplierID).Scan(&product.ID)
	if err != nil {
		return fmt.Errorf("Не удалось создать продукт: %v", err)
	}
	return nil
}

func CreateProductImageTx(tx *sql.Tx, image *models.ProductImage) error {
	query := `INSERT INTO product_images (product_id, image_url) VALUES ($1, $2)`
	_, err := tx.Exec(query, image.ProductID, pq.Array(image.ImageURLs))
	if err != nil {
		return fmt.Errorf("Не удалось добавить изображения продукта: %v", err)
	}
	return nil
}

func CreateProductVariationTx(tx *sql.Tx, variation *models.ProductVariation) error {
	query := `
        INSERT INTO product_variation (product_id, sku, price, stock)
        VALUES ($1, $2, $3, $4)
        RETURNING id
    `
	err := tx.QueryRow(query, variation.ProductID, variation.SKU, variation.Price, variation.Stock).Scan(&variation.ID)
	if err != nil {
		return fmt.Errorf("Не удалось создать вариацию продукта: %v", err)
	}
	return nil
}

func CreateProductVariationImageTx(tx *sql.Tx, image *models.ProductVariationImage) error {
	query := `INSERT INTO product_variation_images (product_variation_id, image_url) VALUES ($1, $2)`
	_, err := tx.Exec(query, image.ProductVariationID, pq.Array(image.ImageURLs))
	if err != nil {
		return fmt.Errorf("Не удалось добавить изображения вариации продукта: %v", err)
	}
	return nil
}

func GetAttributeIDByName(tx *sql.Tx, name string) (int, error) {
	var id int
	query := `SELECT id FROM attributes WHERE name = $1`
	err := tx.QueryRow(query, name).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("Атрибут '%s' не найден: %v", name, err)
	}
	return id, nil
}

func GetAttributeValueID(tx *sql.Tx, attributeID int, value string) (int, error) {
	var id int
	query := `SELECT id FROM attribute_value WHERE attribute_id = $1 AND value = $2`
	err := tx.QueryRow(query, attributeID, value).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("Значение атрибута '%s' не найдено: %v", value, err)
	}
	return id, nil
}

func CreateAttributeValueTx(tx *sql.Tx, attrValue *models.AttributeValue) error {
	query := `
        INSERT INTO attribute_value (attribute_id, value)
        VALUES ($1, $2)
        RETURNING id
    `
	err := tx.QueryRow(query, attrValue.AttributeID, attrValue.Value).Scan(&attrValue.ID)
	if err != nil {
		return fmt.Errorf("Не удалось создать значение атрибута: %v", err)
	}
	return nil
}

func CreateVariationAttributeValueTx(tx *sql.Tx, variationID int, attributeValueID int) error {
	query := `
        INSERT INTO variation_attribute_values (product_variation_id, attribute_value_id)
        VALUES ($1, $2)
    `
	_, err := tx.Exec(query, variationID, attributeValueID)
	if err != nil {
		return fmt.Errorf("Не удалось связать вариацию с атрибутом: %v", err)
	}
	return nil
}

func GetSupplierIDByUserID(userID int) (int, error) {
	var supplierID int
	query := `SELECT id FROM supplier WHERE user_id = $1`
	err := DB.QueryRow(query, userID).Scan(&supplierID)
	if err != nil {
		return 0, fmt.Errorf("Не удалось получить supplier_id для user_id %d: %v", userID, err)
	}
	return supplierID, nil
}

func GetProductByID(productID int) (*models.Product, error) {
	query := `
        SELECT id, name, description, category_id, market_id, status_id, supplier_id
        FROM product
        WHERE id = $1
    `

	var product models.Product
	err := DB.QueryRow(query, productID).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.CategoryID,
		&product.MarketID,
		&product.StatusID,
		&product.SupplierID,
	)
	if err != nil {
		return nil, fmt.Errorf("Не удалось получить продукт: %v", err)
	}

	return &product, nil
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

type Color struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type AttributeValueRequest struct {
	Name  string `json:"name" binding:"required"`
	Value string `json:"value" binding:"required"`
}

func UpdateProduct(productID int, req *models.UpdateProductRequest) error {
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("Не удалось начать транзакцию: %v", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// Обновляем информацию о продукте
	query := `UPDATE product SET name = $1, description = $2, category_id = $3 WHERE id = $4`
	_, err = tx.Exec(query, req.Name, req.Description, req.CategoryID, productID)
	if err != nil {
		return fmt.Errorf("Не удалось обновить продукт: %v", err)
	}

	// Обновляем изображения продукта
	// Сначала удаляем старые
	_, err = tx.Exec(`DELETE FROM product_images WHERE product_id = $1`, productID)
	if err != nil {
		return fmt.Errorf("Не удалось удалить старые изображения продукта: %v", err)
	}

	// Затем добавляем новые
	for _, imageURL := range req.Images {
		_, err = tx.Exec(`INSERT INTO product_images (product_id, image_url) VALUES ($1, $2)`, productID, imageURL)
		if err != nil {
			return fmt.Errorf("Не удалось добавить новое изображение продукта: %v", err)
		}
	}

	// Обновляем вариации продукта
	// Здесь можно реализовать логику обновления, удаления и добавления вариаций

	return nil
}

func GetColorIDByNameAndCode(tx *sql.Tx, name, code string) (int, error) {
	var id int
	query := `SELECT id FROM colors WHERE name = $1 AND code = $2`
	err := tx.QueryRow(query, name, code).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func CreateColorTx(tx *sql.Tx, color models.Color) (int, error) {
	var id int
	query := `INSERT INTO colors (name, code) VALUES ($1, $2) RETURNING id`
	err := tx.QueryRow(query, color.Name, color.Code).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("Не удалось создать цвет: %v", err)
	}
	return id, nil
}

func CreateVariationColorLinkTx(tx *sql.Tx, variationID, colorID int) error {
	query := `INSERT INTO variation_colors (variation_id, color_id) VALUES ($1, $2)`
	_, err := tx.Exec(query, variationID, colorID)
	if err != nil {
		return fmt.Errorf("Не удалось создать связь вариации и цвета: %v", err)
	}
	return nil
}

func GetImmediateSubcategoriesByPath(path string) ([]models.Category, error) {
	query := `
        SELECT id, name, path, image_url
        FROM categories
        WHERE path ~ $1 AND nlevel(path) = nlevel($1) + 1
    `

	pattern := fmt.Sprintf(`^%s\.[^.]+$`, path)

	rows, err := DB.Query(query, pattern)
	if err != nil {
		return nil, fmt.Errorf("ошибка при выполнении запроса: %v", err)
	}
	defer rows.Close()

	var categories []models.Category

	for rows.Next() {
		var category models.Category
		if err := rows.Scan(&category.ID, &category.Name, &category.Path, &category.ImageURL); err != nil {
			return nil, fmt.Errorf("ошибка при сканировании строки: %v", err)
		}
		categories = append(categories, category)
	}

	return categories, nil
}
