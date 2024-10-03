// internal/services/product_service.go

package services

import (
	"fmt"
	"log"
	"strings"

	"github.com/WhyDias/Marketplace/internal/db"
	"github.com/WhyDias/Marketplace/internal/models"
)

type ProductService struct {
}

// NewProductService конструктор сервиса продуктов
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

func (s *ProductService) GetProductsBySupplierAndStatus(supplierID int, statusID int) ([]models.Product, error) {
	return db.GetProductsBySupplierAndStatus(supplierID, statusID)
}

func (s *ProductService) GetSupplierIDByUserID(userID int) (int, error) {
	supplier, err := db.FetchSupplierByUserID(userID)
	if err != nil {
		log.Printf("GetSupplierIDByUserID: ошибка при получении поставщика для user_id %d: %v", userID, err)
		return 0, fmt.Errorf("не удалось получить supplier_id: %v", err)
	}
	return supplier.ID, nil
}

//func (s *ProductService) AddProduct(product *models.Product, attributes []models.AttributeValueRequest) error {
//	tx, err := db.DB.Begin()
//	if err != nil {
//		return fmt.Errorf("не удалось начать транзакцию: %v", err)
//	}
//
//	defer func() {
//		if err != nil {
//			tx.Rollback()
//		} else {
//			tx.Commit()
//		}
//	}()
//
//	// Создаём продукт
//	err = db.CreateProductTx(tx, product)
//	if err != nil {
//		return err
//	}
//
//	// Добавляем изображения продукта
//	for _, img := range product.Images {
//		err = db.CreateProductImageTx(tx, &img)
//		if err != nil {
//			return err
//		}
//	}
//
//	// Генерация и добавление вариаций продукта
//	variations := s.GenerateVariations(attributes)
//	for _, variationAttributes := range variations {
//		variation := &models.ProductVariation{
//			ProductID:  product.ID,
//			SKU:        generateSKU(product.ID, variationAttributes),
//			Price:      product.Price,
//			Stock:      product.Stock,
//			Attributes: []models.AttributeValue{},
//			Images:     []models.ProductVariationImage{},
//		}
//
//		// Создаём вариацию
//		err = db.CreateProductVariationTx(tx, variation)
//		if err != nil {
//			return err
//		}
//
//		// Связываем вариацию с атрибутами
//		for name, value := range variationAttributes {
//			// Получаем или создаём атрибут и его значение
//			attributeID, err := db.GetAttributeIDByName(tx, name)
//			if err != nil {
//				return err
//			}
//			attributeValueID, err := db.GetAttributeValueID(tx, attributeID, value)
//			if err != nil {
//				// Создаём новое значение атрибута
//				attrValue := &models.AttributeValue{
//					AttributeID: attributeID,
//					Value:       value,
//				}
//				err = db.CreateAttributeValueTx(tx, attrValue)
//				if err != nil {
//					return err
//				}
//				attributeValueID = attrValue.ID
//			}
//
//			// Связываем вариацию с значением атрибута
//			err = db.CreateVariationAttributeValueTx(tx, variation.ID, attributeValueID)
//			if err != nil {
//				return err
//			}
//		}
//
//		// Добавляем изображения вариации, если они есть
//		for _, img := range variation.Images {
//			err = db.CreateProductVariationImageTx(tx, &img)
//			if err != nil {
//				return err
//			}
//		}
//	}
//
//	return nil
//}

func generateSKU(productID int, attributes map[string]string) string {
	// Реализуйте генерацию уникального SKU на основе атрибутов
	// Например: "PROD123-SIZE-M-COLOR-RED"
	sku := fmt.Sprintf("PROD%d", productID)
	for name, value := range attributes {
		sku += fmt.Sprintf("-%s-%s", strings.ToUpper(name), strings.ToUpper(value))
	}
	return sku
}

func (s *ProductService) GetMarketIDBySupplierID(supplierID int) (int, error) {
	return db.GetMarketIDBySupplierID(supplierID)
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

//func (s *ProductService) UpdateProduct(productID int, updatedProduct *models.Product) error {
//	tx, err := db.DB.Begin()
//	if err != nil {
//		return fmt.Errorf("не удалось начать транзакцию: %v", err)
//	}
//
//	defer func() {
//		if err != nil {
//			tx.Rollback()
//		} else {
//			tx.Commit()
//		}
//	}()
//
//	// Обновляем продукт
//	err = db.UpdateProductTx(tx, productID, updatedProduct)
//	if err != nil {
//		return err
//	}
//
//	// Обновляем изображения продукта
//	if len(updatedProduct.Images) > 0 {
//		for _, img := range updatedProduct.Images {
//			err = db.CreateProductImageTx(tx, &img)
//			if err != nil {
//				return err
//			}
//		}
//	}
//
//	return nil
//}
//
//func (s *ProductService) addVariationColorsTx(tx *sql.Tx, variationID int, colors []models.Color) error {
//	for _, color := range colors {
//		// Проверяем, существует ли цвет
//		colorID, err := db.GetColorIDByNameAndCode(tx, color.Name, color.Code)
//		if err != nil {
//			if err == sql.ErrNoRows {
//				// Создаём новый цвет
//				colorID, err = db.CreateColorTx(tx, color)
//				if err != nil {
//					return err
//				}
//			} else {
//				return err
//			}
//		}
//
//		// Создаём связь между вариацией и цветом
//		err = db.CreateVariationColorLinkTx(tx, variationID, colorID)
//		if err != nil {
//			return err
//		}
//	}
//	return nil
//}
//
//// GenerateVariations генерирует вариации на основе атрибутов
//func (s *ProductService) GenerateVariations(attributes []models.AttributeValueRequest) []map[string]string {
//	var results []map[string]string
//	s.generateVariationsHelper(attributes, 0, map[string]string{}, &results)
//	return results
//}
//
//func (s *ProductService) generateVariationsHelper(attributes []models.AttributeValueRequest, index int, current map[string]string, results *[]map[string]string) {
//	if index == len(attributes) {
//		temp := make(map[string]string)
//		for k, v := range current {
//			temp[k] = v
//		}
//		*results = append(*results, temp)
//		return
//	}
//
//	attr := attributes[index]
//	for _, value := range attr.Values {
//		current[attr.Name] = value
//		s.generateVariationsHelper(attributes, index+1, current, results)
//	}
//}
//
//func (s *ProductService) UploadImageToYandexDisk(localPath string) (string, error) {
//	yandexPath := "Marketplace/Products/" + filepath.Base(localPath)
//	publicURL, err := s.YandexDisk.UploadFile(localPath, yandexPath)
//	if err != nil {
//		return "", err
//	}
//
//	// Удаляем локальный файл после успешной загрузки
//	err = os.Remove(localPath)
//	if err != nil {
//		log.Printf("Не удалось удалить локальный файл %s: %v", localPath, err)
//		// Не возвращаем ошибку, так как основная операция уже выполнена
//	}
//
//	return publicURL, nil
//}
