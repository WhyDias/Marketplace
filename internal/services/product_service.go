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

func (p *ProductService) AddProduct(req *models.ProductRequest, userID int, c *gin.Context) error {
	// Получаем информацию о поставщике и рынке по userID
	supplier, err := db.GetSupplierByUserID(userID)
	if err != nil {
		return fmt.Errorf("не удалось получить информацию о поставщике: %v", err)
	}

	// Создаем основной продукт
	product := models.Product{
		Name:        req.Name,
		CategoryID:  req.CategoryID,
		MarketID:    supplier.MarketID,
		StatusID:    2, // По умолчанию статус продукта
		SupplierID:  supplier.ID,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
	}

	if err := db.CreateProduct(&product); err != nil {
		return fmt.Errorf("не удалось создать продукт: %v", err)
	}

	// Сохраняем изображения продукта
	// Сохраняем изображения продукта
	uploadDir := fmt.Sprintf("uploads/products/%d", product.ID)

	// Создаем директорию, если она не существует
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return fmt.Errorf("не удалось создать директорию для изображений: %v", err)
	}

	// Проверяем, действительно ли директория создана
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		return fmt.Errorf("директория для изображений не была создана: %v", err)
	}

	log.Printf("Директория для загрузки изображений: %s успешно создана", uploadDir)

	for _, fileHeader := range req.Images {
		if fileHeader != nil {
			// Генерируем имя файла и сохраняем
			fileName := filepath.Join(uploadDir, fileHeader.Filename)
			log.Printf("Попытка сохранить файл: %s", fileName)

			// Убираем дополнительный "uploads" префикс
			if err := utils.SaveUploadedFile(fileHeader, fileName); err != nil {
				return fmt.Errorf("не удалось сохранить изображение продукта: %v", err)
			}

			// Создаем запись изображения в базе данных
			productImage := &models.ProductImage{
				ProductID: product.ID,
				ImageURL:  fileName,
			}

			if err := db.CreateProductImage(productImage); err != nil {
				return fmt.Errorf("не удалось сохранить изображение в базе данных: %v", err)
			}
		}
	}

	// Добавляем вариации
	if err := p.AddProductVariations(req.Variations, product.ID, supplier.ID, c); err != nil {
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

func (p *ProductService) AddProductVariations(variations []models.ProductVariationReq, productID int, supplierID int, c *gin.Context) error {
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

	for _, variationReq := range variations {
		// Создаем запись для вариации
		productVariation := models.ProductVariation{
			ProductID: productID,
			SKU:       fmt.Sprintf("%d-%s", supplierID, utils.GenerateSKU()),
		}

		// Сохраняем вариацию в базе данных
		if err := db.CreateProductVariation(&productVariation); err != nil {
			return fmt.Errorf("не удалось создать вариацию продукта: %v", err)
		}

		// Сохранение атрибутов вариации
		for _, attribute := range variationReq.Attributes {
			attributeID, err := db.GetAttributeIDByName(attribute.Name)
			if err != nil {
				return fmt.Errorf("Ошибка при получении ID атрибута: %v", err)
			}

			attributeValueID, err := db.GetAttributeValueID(attributeID, attribute.Value)
			if err != nil {
				return fmt.Errorf("Ошибка при получении ID значения атрибута: %v", err)
			}

			variationAttributeValue := models.VariationAttributeValue{
				ProductVariationID: productVariation.ID,
				AttributeValueID:   attributeValueID,
			}

			if err := db.CreateVariationAttributeValue(&variationAttributeValue); err != nil {
				return fmt.Errorf("Ошибка при создании значения атрибута для вариации: %v", err)
			}
		}

		// Сохранение изображений для вариации
		for _, file := range variationReq.Images {
			// Генерация имени файла для сохранения
			fileName := fmt.Sprintf("uploads/variations/%d/%s", productVariation.ID, file.Filename)

			// Сохранение изображения на диск
			if err := utils.SaveUploadedFile(file, fileName); err != nil {
				return fmt.Errorf("не удалось сохранить изображение вариации: %v", err)
			}

			// Сохранение информации об изображении в базу данных
			productVariationImage := &models.ProductVariationImage{
				ProductVariationID: productVariation.ID,
				ImageURL:           fileName,
			}

			if err := db.CreateProductVariationImage(productVariationImage); err != nil {
				return fmt.Errorf("не удалось сохранить изображение в базу данных: %v", err)
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
