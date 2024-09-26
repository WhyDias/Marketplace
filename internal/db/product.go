// internal/db/product.go

package db

import (
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
