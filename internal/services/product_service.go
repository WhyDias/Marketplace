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
	// Получаем поставщика по userID
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
		return fmt.Errorf("не удалось создать директорию для изображений продукта: %v", err)
	}

	var productImages []string
	for _, fileHeader := range req.Images {
		if fileHeader != nil {
			// Генерируем имя файла и сохраняем
			fileName := filepath.Join(uploadDir, fileHeader.Filename)
			if err := utils.SaveUploadedFile(fileHeader, fileName); err != nil {
				return fmt.Errorf("не удалось сохранить изображение продукта: %v", err)
			}
			productImages = append(productImages, fmt.Sprintf("http://195.49.215.120:8080/%s", fileName))
		}
	}

	// Сохраняем ссылки на изображения в базе данных
	if len(productImages) > 0 {
		imagesJSON, err := json.Marshal(productImages)
		if err != nil {
			return fmt.Errorf("не удалось преобразовать массив изображений в JSON: %v", err)
		}

		productImageRecord := &models.ProductImage{
			ProductID: product.ID,
			ImageURLs: string(imagesJSON), // Хранится как JSON строка в базе данных
		}

		if err := db.CreateProductImage(productImageRecord); err != nil {
			return fmt.Errorf("не удалось сохранить изображения продукта в базе данных: %v", err)
		}
	}

	// Сохраняем значения атрибутов для основного продукта
	for _, attribute := range req.Attributes {
		attributeID, err := db.GetAttributeIDByName(req.CategoryID, attribute.Name)
		if err != nil {
			return fmt.Errorf("Ошибка при получении ID атрибута '%s': %v", attribute.Name, err)
		}

		attributeValueStr, ok := attribute.Value.(string)
		if !ok {
			return fmt.Errorf("ошибка приведения значения атрибута '%s' к строке: %v", attribute.Name, attribute.Value)
		}

		attributeValueID, err := db.GetAttributeValueID(attributeID, attributeValueStr)
		if err != nil {
			if err == sql.ErrNoRows {
				attributeValueID, err = db.CreateAttributeValue(attributeID, json.RawMessage(attributeValueStr))
				if err != nil {
					return fmt.Errorf("Ошибка при создании значения атрибута '%s': %v", attribute.Name, err)
				}
			} else {
				return fmt.Errorf("Ошибка при получении ID значения атрибута: %v", err)
			}
		}

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
	multipartForm, err := c.MultipartForm()
	if err != nil {
		return fmt.Errorf("не удалось получить multipart form: %v", err)
	}

	// Извлекаем файлы для вариаций из формы
	files := multipartForm.File

	for i, variationReq := range variations {
		log.Printf("AddProductVariations: Обработка вариации %d", i)

		// Создаем запись для вариации
		productVariation := models.ProductVariation{
			ProductID: productID,
			SKU:       fmt.Sprintf("%d-%s", supplierID, utils.GenerateSKU()), // Автогенерация SKU
		}

		// Сохраняем вариацию в базе данных
		if err := db.CreateProductVariation(&productVariation); err != nil {
			log.Printf("AddProductVariations: Не удалось создать вариацию продукта, ошибка: %v", err)
			return fmt.Errorf("не удалось создать вариацию продукта: %v", err)
		} else {
			log.Printf("AddProductVariations: Вариация успешно создана с ID: %d", productVariation.ID)
		}

		// Сохранение изображений для вариации
		paramName := fmt.Sprintf("variation_images_%d", i+1)
		variationImages, ok := files[paramName]
		if !ok || len(variationImages) == 0 {
			log.Printf("Не удалось найти изображения для вариации %d", i+1)
		} else {
			if err := p.SaveVariationImages(productVariation.ID, variationImages, c); err != nil {
				return fmt.Errorf("не удалось сохранить изображения для вариации: %v", err)
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

			// Преобразуем значение атрибута в json.RawMessage
			valueJSON, err := json.Marshal(attribute.Value)
			if err != nil {
				log.Printf("SaveVariationAttributes: Ошибка при преобразовании значения атрибута '%s' в JSON: %v", attribute.Name, err)
				return fmt.Errorf("ошибка при преобразовании значения атрибута '%s': %v", attribute.Name, err)
			}

			// Создаем или обновляем значение атрибута в attribute_value и получаем его ID
			attributeValueID, err := db.CreateOrUpdateAttributeValue(attributeID, json.RawMessage(valueJSON))
			if err != nil {
				log.Printf("SaveVariationAttributes: Ошибка при создании или обновлении значения атрибута '%s': %v", attribute.Name, err)
				return fmt.Errorf("ошибка при создании или обновлении значения атрибута '%s': %v", attribute.Name, err)
			}

			// Создаем запись в таблице variation_attribute_values
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

	log.Printf("Директория для загрузки изображений вариации: %s успешно создана", uploadDir)

	var imagePaths []string
	for _, fileHeader := range images {
		if fileHeader != nil {
			fileName := filepath.Join(uploadDir, fileHeader.Filename)
			if err := utils.SaveUploadedFile(fileHeader, fileName); err != nil {
				return fmt.Errorf("не удалось сохранить изображение вариации: %v", err)
			}
			imagePaths = append(imagePaths, fmt.Sprintf("http://195.49.215.120:8080/%s", fileName))
		}
	}

	if len(imagePaths) > 0 {
		imagesJSON, err := json.Marshal(imagePaths)
		if err != nil {
			return fmt.Errorf("не удалось преобразовать массив изображений в JSON: %v", err)
		}

		productVariationImage := &models.ProductVariationImage{
			ProductVariationID: variationID,
			ImageURLs:          string(imagesJSON), // Хранится как JSON строка в базе данных
		}

		if err := db.CreateProductVariationImage(productVariationImage); err != nil {
			return fmt.Errorf("не удалось сохранить изображения в базе данных: %v", err)
		}

		log.Printf("Информация об изображениях успешно сохранена для вариации ID: %d", variationID)
	} else {
		log.Printf("Не было найдено изображений для сохранения для вариации ID: %d", variationID)
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
