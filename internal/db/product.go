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
