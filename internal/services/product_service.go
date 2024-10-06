// internal/services/product_service.go

package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/WhyDias/Marketplace/internal/db"
	"github.com/WhyDias/Marketplace/internal/models"
	"github.com/WhyDias/Marketplace/internal/utils"
	"github.com/gin-gonic/gin"
	"log"
	"mime/multipart"
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

	// Сохраняем значения атрибутов для основного продукта
	for _, attribute := range req.Attributes {
		attributeID, err := db.GetAttributeIDByName(req.CategoryID, attribute.Name)
		if err != nil {
			return fmt.Errorf("Ошибка при получении ID атрибута '%s': %v", attribute.Name, err)
		}

		// Обработка значений разных типов (string, bool, int, float и т.д.)
		var attributeValueStr string
		switch v := attribute.Value.(type) {
		case string:
			attributeValueStr = v
		case bool:
			// Преобразование bool в строку
			if v {
				attributeValueStr = "true"
			} else {
				attributeValueStr = "false"
			}
		case float64:
			// Преобразование float64 в строку (учитывая точность до двух знаков после запятой)
			attributeValueStr = fmt.Sprintf("%.2f", v)
		case int:
			attributeValueStr = fmt.Sprintf("%d", v)
		default:
			return fmt.Errorf("неподдерживаемый тип значения атрибута '%s': %T", attribute.Name, v)
		}

		// Получаем ID значения атрибута или создаем новое значение
		attributeValueID, err := db.CreateOrUpdateAttributeValue(attributeID, json.RawMessage(`"`+attributeValueStr+`"`))
		if err != nil {
			return fmt.Errorf("Ошибка при создании или обновлении значения атрибута '%s': %v", attribute.Name, err)
		}

		// Создаем связь между продуктом и значением атрибута
		productAttributeValue := &models.ProductAttributeValue{
			ProductID:        product.ID,
			AttributeValueID: attributeValueID,
		}

		if err := db.CreateProductAttributeValue(productAttributeValue); err != nil {
			return fmt.Errorf("Ошибка при создании значения атрибута для продукта: %v", err)
		}
	}

	// Добавляем вариации
	if err := p.AddProductVariations(variations, product.ID, product.CategoryID, supplier.ID, c); err != nil {
		return fmt.Errorf("не удалось добавить вариации продукта: %v", err)
	}

	return nil
}

func (p *ProductService) GetProductAttributes(categoryID int) ([]models.Attribute, error) {
	attributes, err := db.GetCategoryAttributes(categoryID)
	if err != nil {
		log.Printf("GetProductAttributes: Ошибка при получении атрибутов для категории %d: %v", categoryID, err)
		return nil, fmt.Errorf("не удалось получить атрибуты категории: %v", err)
	}
	return attributes, nil
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
	for i, variationReq := range variations {
		log.Printf("AddProductVariations: Обработка вариации %d", i+1)

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

		// Сохранение изображений для вариации
		imageField := fmt.Sprintf("variation_images_%d", i+1)
		if images, ok := c.Request.MultipartForm.File[imageField]; ok {
			if err := p.SaveVariationImages(productVariation.ID, images, c); err != nil {
				return fmt.Errorf("не удалось сохранить изображения для вариации %d: %v", i+1, err)
			}
		}

		// Сохранение атрибутов для вариации
		for _, attribute := range variationReq.Attributes {
			log.Printf("SaveVariationAttributes: Сохранение атрибута '%s' для вариации ID: %d", attribute.Name, productVariation.ID)

			// Получаем ID атрибута по имени
			attributeID, err := db.GetAttributeIDByNameAndCategory(attribute.Name, categoryID)
			if err != nil {
				log.Printf("SaveVariationAttributes: Ошибка при получении ID атрибута '%s': %v", attribute.Name, err)
				return fmt.Errorf("ошибка при получении ID атрибута '%s': %v", attribute.Name, err)
			}

			// Проверяем или создаем значение атрибута в таблице attribute_value
			attributeValueStr, ok := attribute.Value.(string)
			if !ok {
				return fmt.Errorf("ошибка приведения значения атрибута '%s' к строке: %v", attribute.Name, attribute.Value)
			}
			attributeValueID, err := db.CreateOrUpdateAttributeValue(attributeID, json.RawMessage(attributeValueStr))
			if err != nil {
				log.Printf("SaveVariationAttributes: Ошибка при создании или обновлении значения атрибута '%s': %v", attribute.Name, err)
				return fmt.Errorf("ошибка при создании или обновлении значения атрибута '%s': %v", attribute.Name, err)
			}

			variationAttributeValue := models.VariationAttributeValue{
				ProductVariationID: productVariation.ID,
				AttributeValueID:   attributeValueID,
			}

			if err := db.CreateVariationAttributeValue(&variationAttributeValue); err != nil {
				log.Printf("AddProductVariations: Ошибка при создании записи в variation_attribute_values для атрибута '%s': %v", attribute.Name, err)
				return fmt.Errorf("ошибка при создании связи в variation_attribute_values для атрибута '%s': %v", attribute.Name, err)
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

func (p *ProductService) SaveVariationImages(variationID int, images []*multipart.FileHeader, c *gin.Context) error {
	// Создаем директорию для изображений этой вариации
	uploadDir := fmt.Sprintf("uploads/variations/%d", variationID)
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return fmt.Errorf("не удалось создать директорию для изображений вариации: %v", err)
	}

	// Проверяем, действительно ли директория создана
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		return fmt.Errorf("директория для изображений не была создана: %v", err)
	}

	log.Printf("Директория для загрузки изображений вариации: %s успешно создана", uploadDir)

	// Сохраняем каждое изображение в файловую систему и записываем его в базу данных
	for _, fileHeader := range images {
		if fileHeader != nil {
			// Генерируем имя файла и сохраняем
			fileName := filepath.Join(uploadDir, fileHeader.Filename)
			log.Printf("Попытка сохранить файл: %s", fileName)

			if err := utils.SaveUploadedFile(fileHeader, fileName); err != nil {
				return fmt.Errorf("не удалось сохранить изображение вариации: %v", err)
			}

			// Создаем запись изображения в базе данных
			productVariationImage := &models.ProductVariationImage{
				ProductVariationID: variationID,
				ImageURL:           fileName,
			}

			if err := db.CreateProductVariationImage(productVariationImage); err != nil {
				return fmt.Errorf("не удалось сохранить изображение в базе данных: %v", err)
			} else {
				log.Printf("Информация об изображении успешно сохранена для вариации ID: %d", variationID)
			}
		} else {
			log.Printf("Пропущен файл изображения для вариации ID: %d, так как он равен nil", variationID)
		}
	}

	return nil
}

func (p *ProductService) SaveVariationAttributes(variationID int, categoryID int, attributes []models.AttributeValueRequest) error {
	for _, attribute := range attributes {
		log.Printf("SaveVariationAttributes: Сохранение атрибута '%s' для вариации ID: %d", attribute.Name, variationID)

		// Получение ID атрибута по имени и категории
		attributeID, err := db.GetAttributeIDByNameAndCategory(attribute.Name, categoryID)
		if err != nil {
			return fmt.Errorf("не удалось получить ID атрибута '%s': %v", attribute.Name, err)
		}

		// Преобразование значения атрибута в JSON
		valueJSON, err := json.Marshal(attribute.Value)
		if err != nil {
			return fmt.Errorf("не удалось преобразовать значение атрибута '%s' в JSON: %v", attribute.Name, err)
		}

		// Обновление значения атрибута в базе данных
		if err := db.UpdateAttributeValue(attributeID, json.RawMessage(valueJSON)); err != nil {
			return fmt.Errorf("ошибка при обновлении значения атрибута '%s': %v", attribute.Name, err)
		}
	}

	return nil
}
