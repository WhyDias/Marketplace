// internal/services/product_service.go

package services

import (
	"fmt"
	"log"

	"github.com/WhyDias/Marketplace/internal/db"
	"github.com/WhyDias/Marketplace/internal/models"
)

type ProductService struct{}

func NewProductService() *ProductService {
	return &ProductService{}
}

// GetProductsByStatus получает список продуктов по заданному status_id
func (s *ProductService) GetProductsByStatusWithPagination(statusID, limit, offset int) ([]models.Product, error) {
	products, err := db.FetchProductsByStatus(statusID, limit, offset)
	if err != nil {
		log.Printf("ProductService: ошибка при получении продуктов со статусом %d: %v", statusID, err)
		return nil, fmt.Errorf("не удалось получить продукты: %v", err)
	}
	return products, nil
}
