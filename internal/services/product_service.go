// internal/services/product_service.go

package services

import (
	"database/sql"
	"fmt"
	"github.com/WhyDias/Marketplace/internal/db"
	"github.com/WhyDias/Marketplace/internal/models"
	"log"
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

func (s *ProductService) AddProduct(userID int, req *models.ProductRequest) error {
	// Получаем supplier_id и market_id для текущего пользователя
	supplier, err := db.GetSupplierByUserID(userID)
	if err != nil {
		return fmt.Errorf("ошибка при получении поставщика: %v", err)
	}
	if supplier == nil {
		return fmt.Errorf("поставщик не найден для user_id=%d", userID)
	}

	// Создаем продукт
	product := &models.Product{
		Name:        req.Name,
		CategoryID:  req.CategoryID,
		StatusID:    2, // Присваиваем статус 2 (например, "Active")
		SupplierID:  supplier.ID,
		MarketID:    supplier.MarketID,
		Description: req.Description,
	}

	err = db.CreateProduct(product)
	if err != nil {
		return fmt.Errorf("ошибка при создании продукта: %v", err)
	}

	// Добавляем изображения продукта
	for _, imgURL := range req.Images {
		productImage := &models.ProductImage{
			ProductID: product.ID,
			ImageURL:  imgURL,
		}
		if err := db.CreateProductImage(productImage); err != nil {
			log.Printf("Не удалось добавить изображение продукта: %s. Ошибка: %v", imgURL, err)
			// Можно решить, продолжать ли добавление остальных изображений или прерывать
			// Для примера продолжаем
		}
	}

	// Добавляем вариации продукта
	for _, varReq := range req.Variations {
		variation := &models.ProductVariation{
			ProductID: product.ID,
			SKU:       varReq.SKU,
			Price:     varReq.Price,
			Stock:     varReq.Stock,
		}

		err = db.CreateProductVariation(variation)
		if err != nil {
			return fmt.Errorf("ошибка при создании вариации продукта: %v", err)
		}

		// Добавляем атрибуты для вариации
		for _, attrReq := range varReq.Attributes {
			attrID, err := db.GetAttributeIDByName(attrReq.Name)
			if err != nil {
				return fmt.Errorf("ошибка при получении атрибута %s: %v", attrReq.Name, err)
			}

			// Получаем или создаем значение атрибута
			attrValueID, err := db.GetOrCreateAttributeValue(attrID, attrReq.Value)
			if err != nil {
				return fmt.Errorf("ошибка при получении или создании значения атрибута: %v", err)
			}

			err = db.CreateVariationAttributeValue(variation.ID, attrValueID)
			if err != nil {
				return fmt.Errorf("ошибка при связывании атрибута вариации: %v", err)
			}
		}

		// Добавляем изображения вариации продукта
		for _, imgURL := range varReq.Images {
			variationImage := &models.ProductVariationImage{
				ProductVariationID: variation.ID,
				ImageURL:           imgURL,
			}
			if err := db.CreateProductVariationImage(variationImage); err != nil {
				log.Printf("Не удалось добавить изображение вариации продукта: %s. Ошибка: %v", imgURL, err)
				// Можно решить, продолжать ли добавление остальных изображений или прерывать
				// Для примера продолжаем
			}
		}
	}

	return nil
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

func (s *ProductService) UpdateProduct(userID, productID int, req *models.UpdateProductRequest) error {
	// Получаем supplier_id по user_id
	supplierID, err := s.GetSupplierIDByUserID(userID)
	if err != nil {
		return fmt.Errorf("Не удалось получить supplier_id: %v", err)
	}

	// Проверяем, что продукт принадлежит поставщику
	product, err := db.GetProductByID(productID)
	if err != nil {
		return fmt.Errorf("Не удалось получить продукт: %v", err)
	}

	if product.SupplierID != supplierID {
		return fmt.Errorf("У вас нет прав для обновления этого продукта")
	}

	// Обновляем продукт в базе данных
	return db.UpdateProduct(productID, req)
}

func (s *ProductService) addVariationColorsTx(tx *sql.Tx, variationID int, colors []models.Color) error {
	for _, color := range colors {
		// Проверяем, существует ли цвет
		colorID, err := db.GetColorIDByNameAndCode(tx, color.Name, color.Code)
		if err != nil {
			if err == sql.ErrNoRows {
				// Создаём новый цвет
				colorID, err = db.CreateColorTx(tx, color)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}

		// Создаём связь между вариацией и цветом
		err = db.CreateVariationColorLinkTx(tx, variationID, colorID)
		if err != nil {
			return err
		}
	}
	return nil
}
