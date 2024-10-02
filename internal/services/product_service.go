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

func (s *ProductService) AddProduct(product *models.Product) error {
	tx, err := db.DB.Begin()
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

	// Создаем продукт
	err = db.CreateProductTx(tx, product)
	if err != nil {
		return err
	}

	// Добавляем изображения продукта
	if len(product.Images) > 0 {
		productImage := &models.ProductImage{
			ProductID: product.ID,
			ImageURLs: product.Images[0].ImageURLs,
		}
		err = db.CreateProductImageTx(tx, productImage)
		if err != nil {
			return err
		}
	}

	// Добавляем вариации продукта
	for i, variation := range product.Variations {
		variation.ProductID = product.ID
		err = db.CreateProductVariationTx(tx, &variation)
		if err != nil {
			return err
		}
		product.Variations[i].ID = variation.ID

		// Добавляем изображения вариации
		if len(variation.Images) > 0 {
			variationImage := &models.ProductVariationImage{
				ProductVariationID: variation.ID,
				ImageURLs:          variation.Images[0].ImageURLs,
			}
			err = db.CreateProductVariationImageTx(tx, variationImage)
			if err != nil {
				return err
			}
		}

		// Обрабатываем атрибуты вариации
		for _, attr := range variation.Attributes {
			// Получаем или создаем атрибут
			attributeID, err := db.GetAttributeIDByName(tx, attr.Name)
			if err != nil {
				return err
			}
			attr.AttributeID = attributeID

			// Получаем или создаем значение атрибута
			attributeValueID, err := db.GetAttributeValueID(tx, attributeID, attr.Value)
			if err != nil {
				// Создаем новое значение атрибута
				attrValue := &models.AttributeValue{
					AttributeID: attributeID,
					Value:       attr.Value,
				}
				err = db.CreateAttributeValueTx(tx, attrValue)
				if err != nil {
					return err
				}
				attributeValueID = attrValue.ID
			}

			// Связываем вариацию с значением атрибута
			err = db.CreateVariationAttributeValueTx(tx, variation.ID, attributeValueID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
