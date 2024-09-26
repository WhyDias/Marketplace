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

func (s *ProductService) GetProductsByStatus(statusID int) ([]models.Product, error) {
	products, err := db.FetchProductsByStatus(statusID)
	if err != nil {
		log.Printf("ProductService: ошибка при получении продуктов со статусом %d: %v", statusID, err)
		return nil, fmt.Errorf("не удалось получить продукты: %v", err)
	}
	return products, nil
}

func (s *ProductService) GetProductsBySupplierAndStatus(supplierID, statusID int) ([]models.Product, error) {
	products, err := db.FetchProductsBySupplierAndStatus(supplierID, statusID)
	if err != nil {
		log.Printf("ProductService: ошибка при получении продуктов для supplier_id %d и status_id %d: %v", supplierID, statusID, err)
		return nil, fmt.Errorf("не удалось получить продукты: %v", err)
	}
	return products, nil
}

func (s *ProductService) GetSupplierIDByUserID(userID int) (int, error) {
	supplier, err := db.FetchSupplierByUserID(userID)
	if err != nil {
		log.Printf("GetSupplierIDByUserID: ошибка при получении поставщика для user_id %d: %v", userID, err)
		return 0, fmt.Errorf("не удалось получить supplier_id: %v", err)
	}
	return supplier.ID, nil
}
