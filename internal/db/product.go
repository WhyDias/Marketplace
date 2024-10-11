// internal/db/product.go

package db

import (
	"database/sql"
	"encoding/json"
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

func IsAttributeLinked(attributeID int) (bool, error) {
	var isLinked bool
	query := "SELECT is_linked FROM attributes WHERE id = $1"
	err := DB.QueryRow(query, attributeID).Scan(&isLinked)
	if err != nil {
		return false, err
	}
	return isLinked, nil
}

func CreateOrGetAttributeValue(attributeID int, value interface{}) (int, error) {
	// Преобразуем значение в json.RawMessage
	valueJSON, err := json.Marshal(value)
	if err != nil {
		return 0, fmt.Errorf("Ошибка при преобразовании значения атрибута в JSON: %v", err)
	}

	// Проверяем, есть ли уже такое значение
	var attributeValueID int
	query := "SELECT id FROM attribute_value WHERE attribute_id = $1 AND value_json = $2"
	err = DB.QueryRow(query, attributeID, valueJSON).Scan(&attributeValueID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Значение не найдено, создаем новое
			query = "INSERT INTO attribute_value (attribute_id, value_json) VALUES ($1, $2) RETURNING id"
			err = DB.QueryRow(query, attributeID, valueJSON).Scan(&attributeValueID)
			if err != nil {
				return 0, fmt.Errorf("Ошибка при создании значения атрибута: %v", err)
			}
		} else {
			return 0, fmt.Errorf("Ошибка при получении значения атрибута: %v", err)
		}
	}

	return attributeValueID, nil
}

func UpdateProduct(product *models.Product) error {
	query := `
        UPDATE product
        SET name = $1, description = $2, category_id = $3, price = $4
        WHERE id = $5
    `
	_, err := DB.Exec(query, product.Name, product.Description, product.CategoryID, product.Price, product.ID)
	if err != nil {
		return fmt.Errorf("не удалось обновить продукт: %v", err)
	}
	return nil
}

func UpdateProductVariation(variation *models.ProductVariation) error {
	query := `
        UPDATE product_variation
        SET sku = $1, price = $2, stock = $3
        WHERE id = $4
    `
	_, err := DB.Exec(query, variation.SKU, variation.Price, variation.Stock, variation.ID)
	if err != nil {
		return fmt.Errorf("не удалось обновить вариацию продукта: %v", err)
	}
	return nil
}

func DeleteProductAttributes(productID int) error {
	query := `
        DELETE FROM product_attribute_values
        WHERE product_id = $1
    `
	_, err := DB.Exec(query, productID)
	if err != nil {
		return fmt.Errorf("не удалось удалить атрибуты продукта: %v", err)
	}
	return nil
}

func DeleteVariationAttributes(variationID int) error {
	query := `
        DELETE FROM variation_attribute_values
        WHERE product_variation_id = $1
    `
	_, err := DB.Exec(query, variationID)
	if err != nil {
		return fmt.Errorf("не удалось удалить атрибуты вариации: %v", err)
	}
	return nil
}

func DeleteProductVariation(variationID int) error {
	// Начинаем транзакцию, чтобы удалить связанные данные
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("не удалось начать транзакцию: %v", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// Удаляем атрибуты вариации
	_, err = tx.Exec(`DELETE FROM variation_attribute_values WHERE product_variation_id = $1`, variationID)
	if err != nil {
		return fmt.Errorf("не удалось удалить атрибуты вариации: %v", err)
	}

	// Удаляем изображения вариации
	_, err = tx.Exec(`DELETE FROM product_variation_images WHERE product_variation_id = $1`, variationID)
	if err != nil {
		return fmt.Errorf("не удалось удалить изображения вариации: %v", err)
	}

	// Удаляем саму вариацию
	_, err = tx.Exec(`DELETE FROM product_variation WHERE id = $1`, variationID)
	if err != nil {
		return fmt.Errorf("не удалось удалить вариацию: %v", err)
	}

	return nil
}

func DeleteVariationImages(variationID int) error {
	// Получаем ссылки на изображения для удаления из хранилища, если необходимо
	query := `
        SELECT image_urls
        FROM product_variation_images
        WHERE product_variation_id = $1
    `
	var imagesJSON string
	err := DB.QueryRow(query, variationID).Scan(&imagesJSON)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("не удалось получить изображения вариации: %v", err)
	}

	// Удаляем записи из базы данных
	_, err = DB.Exec(`DELETE FROM product_variation_images WHERE product_variation_id = $1`, variationID)
	if err != nil {
		return fmt.Errorf("не удалось удалить изображения вариации из базы данных: %v", err)
	}

	// Если необходимо, можно добавить код для удаления файлов из Yandex Cloud Storage

	return nil
}

func GetProductVariations(productID int) ([]models.ProductVariation, error) {
	query := `
        SELECT id, product_id, sku, price, stock
        FROM product_variation
        WHERE product_id = $1
    `
	rows, err := DB.Query(query, productID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить вариации продукта: %v", err)
	}
	defer rows.Close()

	var variations []models.ProductVariation
	for rows.Next() {
		var variation models.ProductVariation
		err := rows.Scan(&variation.ID, &variation.ProductID, &variation.SKU, &variation.Price, &variation.Stock)
		if err != nil {
			return nil, fmt.Errorf("ошибка при чтении вариации: %v", err)
		}
		variations = append(variations, variation)
	}

	return variations, nil
}

// UpdateProductStatusTx обновляет статус продукта в транзакции
func UpdateProductStatusTx(tx *sql.Tx, productID int, statusID int) error {
	query := `
        UPDATE product
        SET status_id = $1
        WHERE id = $2
    `
	_, err := tx.Exec(query, statusID, productID)
	if err != nil {
		return fmt.Errorf("не удалось обновить статус продукта: %v", err)
	}
	return nil
}

// CreateCommentTx создает комментарий в транзакции
func CreateCommentTx(tx *sql.Tx, comment *models.Comment) error {
	query := `
        INSERT INTO comments (user_id, product_id, content)
        VALUES ($1, $2, $3)
        RETURNING id, created_at, updated_at
    `
	err := tx.QueryRow(query, comment.UserID, comment.ProductID, comment.Content).Scan(&comment.ID, &comment.CreatedAt, &comment.UpdatedAt)
	if err != nil {
		return fmt.Errorf("не удалось создать комментарий: %v", err)
	}
	return nil
}
