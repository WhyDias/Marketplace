// internal/db/product.go

package db

import (
	"database/sql"
	"fmt"
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
        WHERE path ~ $1 AND nlevel(path) = nlevel($2::ltree) + 1
    `

	// Параметр для сопоставления пути (lquery)
	lqueryPattern := fmt.Sprintf(`%s.*{1}`, path)

	// Параметр для nlevel (ltree)
	ltreePath := path

	rows, err := DB.Query(query, lqueryPattern, ltreePath)
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

func GetProductsBySupplierAndStatus(supplierID int, statusID int) ([]models.Product, error) {
	query := `
        SELECT 
            p.id, 
            p.name, 
            p.description, 
            p.category_id, 
            c.name AS category_name, 
            p.supplier_id, 
            p.market_id, 
            p.status_id, 
            p.price, 
            p.stock
        FROM product p
        JOIN categories c ON p.category_id = c.id
        WHERE p.supplier_id = $1 AND p.status_id = $2
    `

	rows, err := DB.Query(query, supplierID, statusID)
	if err != nil {
		return nil, fmt.Errorf("Не удалось получить продукты: %v", err)
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var product models.Product
		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Description,
			&product.CategoryID,
			&product.SupplierID,
			&product.MarketID,
			&product.StatusID,
			&product.Price,
			&product.Stock,
		)
		if err != nil {
			return nil, fmt.Errorf("Ошибка при чтении продукта: %v", err)
		}
		products = append(products, product)
	}

	return products, nil
}
