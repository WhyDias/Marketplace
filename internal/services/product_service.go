// internal/services/product_service.go

package services

import (
	"database/sql"
	"fmt"
	"github.com/WhyDias/Marketplace/internal/db"
	"github.com/WhyDias/Marketplace/internal/models"
	"github.com/WhyDias/Marketplace/internal/utils"
	"github.com/gin-gonic/gin"
	"log"
	"os"
	"path/filepath"
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

func (p *ProductService) AddProduct(req *models.ProductRequest, userID int, variations []models.ProductVariationReq, c *gin.Context) error {
	// Получаем информацию о поставщике по userID
	supplier, err := db.GetSupplierByUserID(userID)
	if err != nil {
		return fmt.Errorf("не удалось получить информацию о поставщике: %v", err)
	}

	// Создаем основной продукт
	product := models.Product{
		Name:        req.Name,
		CategoryID:  req.CategoryID,
		MarketID:    supplier.MarketID,
		StatusID:    2,
		SupplierID:  supplier.ID,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
	}

	if err := db.CreateProduct(&product); err != nil {
		return fmt.Errorf("не удалось создать продукт: %v", err)
	}

	// Сохраняем изображения продукта
	uploadDir := fmt.Sprintf("uploads/products/%d", product.ID)
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return fmt.Errorf("не удалось создать директорию для изображений: %v", err)
	}

	for _, fileHeader := range req.Images {
		fileName := filepath.Join(uploadDir, fileHeader.Filename)
		log.Printf("Попытка сохранить файл: %s", fileName)

		if err := utils.SaveUploadedFile(fileHeader, fileName); err != nil {
			return fmt.Errorf("не удалось сохранить изображение продукта: %v", err)
		}

		productImage := &models.ProductImage{
			ProductID: product.ID,
			ImageURL:  fileName,
		}

		if err := db.CreateProductImage(productImage); err != nil {
			return fmt.Errorf("не удалось сохранить изображение в базе данных: %v", err)
		}
	}

	// Добавляем вариации
	if err := p.AddProductVariations(variations, product.ID, product.CategoryID, supplier.ID, c); err != nil {
		return fmt.Errorf("не удалось добавить вариации продукта: %v", err)
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

func (p *ProductService) AddProductVariations(variations []models.ProductVariationReq, productID int, categoryID int, supplierID int, c *gin.Context) error {
	for _, variationReq := range variations {
		log.Printf("AddProductVariations: Обработка вариации")

		// Создаем запись для вариации вне транзакции
		productVariation := models.ProductVariation{
			ProductID: productID,
			SKU:       fmt.Sprintf("%d-%s", supplierID, utils.GenerateSKU()), // Автогенерация SKU
		}

		// Сохраняем вариацию в базе данных
		if err := db.CreateProductVariation(&productVariation); err != nil {
			log.Printf("AddProductVariations: Не удалось создать вариацию продукта, ошибка: %v", err)
			return fmt.Errorf("Не удалось создать вариацию продукта: %v", err)
		} else {
			log.Printf("AddProductVariations: Вариация успешно создана с ID: %d", productVariation.ID)
		}

		// Теперь, когда у нас есть ID вариации, можно добавить атрибуты и изображения
		for _, attribute := range variationReq.Attributes {
			log.Printf("AddProductVariations: Сохранение атрибута '%s' для вариации SKU: %s", attribute.Name, variationReq.SKU)

			attributeID, err := db.GetAttributeIDByName(categoryID, attribute.Name)
			if err != nil {
				log.Printf("AddProductVariations: Ошибка при получении ID атрибута '%s': %v", attribute.Name, err)
				return fmt.Errorf("Ошибка при получении ID атрибута: %v", err)
			}

			attributeValueStr, ok := attribute.Value.(string)
			if !ok {
				log.Printf("AddProductVariations: Ошибка приведения значения атрибута '%s' к строке: %v", attribute.Name, attribute.Value)
				return fmt.Errorf("Некорректный тип значения атрибута '%s'", attribute.Name)
			}

			attributeValueID, err := db.GetAttributeValueID(attributeID, attributeValueStr)
			if err != nil {
				log.Printf("AddProductVariations: Ошибка при получении ID значения атрибута '%s': %v", attributeValueStr, err)
				return fmt.Errorf("Ошибка при получении ID значения атрибута: %v", err)
			}

			variationAttributeValue := models.VariationAttributeValue{
				ProductVariationID: productVariation.ID,
				AttributeValueID:   attributeValueID,
			}

			if err := db.CreateVariationAttributeValue(&variationAttributeValue); err != nil {
				log.Printf("AddProductVariations: Ошибка при создании значения атрибута '%s' для вариации SKU: %s: %v", attribute.Name, variationReq.SKU, err)
				return fmt.Errorf("Ошибка при создании значения атрибута для вариации: %v", err)
			} else {
				log.Printf("AddProductVariations: Значение атрибута '%s' успешно сохранено для вариации SKU: %s", attribute.Name, variationReq.SKU)
			}
		}

		// Сохранение изображений для вариации
		for _, file := range variationReq.Images {
			if file != nil {
				log.Printf("AddProductVariations: Сохранение изображения '%s' для вариации SKU: %s", file.Filename, variationReq.SKU)

				// Генерация имени файла для сохранения
				fileName := fmt.Sprintf("uploads/variations/%d/%s", productVariation.ID, file.Filename)

				// Сохранение изображения на диск
				if err := utils.SaveUploadedFile(file, fileName); err != nil {
					log.Printf("AddProductVariations: Не удалось сохранить изображение '%s' для вариации SKU: %s: %v", file.Filename, variationReq.SKU, err)
					return fmt.Errorf("Не удалось сохранить изображение вариации: %v", err)
				} else {
					log.Printf("AddProductVariations: Изображение '%s' успешно сохранено для вариации SKU: %s", file.Filename, variationReq.SKU)
				}

				// Сохранение информации об изображении в базу данных
				productVariationImage := &models.ProductVariationImage{
					ProductVariationID: productVariation.ID,
					ImageURL:           fileName,
				}

				if err := db.CreateProductVariationImage(productVariationImage); err != nil {
					log.Printf("AddProductVariations: Не удалось сохранить информацию об изображении в базу данных для SKU: %s: %v", variationReq.SKU, err)
					return fmt.Errorf("Не удалось сохранить изображение в базу данных: %v", err)
				} else {
					log.Printf("AddProductVariations: Информация об изображении успешно сохранена для SKU: %s", variationReq.SKU)
				}
			} else {
				log.Printf("AddProductVariations: Пропущен файл изображения для SKU: %s, так как он равен nil", variationReq.SKU)
			}
		}
	}

	return nil
}

func (s *ProductService) GetSupplierByUserID(userID int) (*models.Supplier, error) {
	return db.GetSupplierByUserID(userID)
}

func (s *ProductService) AddProductImage(image *models.ProductImage) error {
	return db.CreateProductImage(image)
}
