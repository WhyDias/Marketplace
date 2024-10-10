// internal/services/product_service.go

package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/WhyDias/Marketplace/internal/db"
	"github.com/WhyDias/Marketplace/internal/models"
	"github.com/WhyDias/Marketplace/internal/utils"
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

func (p *ProductService) AddProduct(req *models.ProductRequest, userID int, attributes []models.AttributeValueRequest, variations []models.ProductVariationRequest) error {
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
		Stock:       0, // Устанавливаем 0 по умолчанию
	}

	if err := db.CreateProduct(&product); err != nil {
		return fmt.Errorf("не удалось создать продукт: %v", err)
	}

	// Сохраняем общие атрибуты для продукта
	for _, attribute := range attributes {
		err := p.SaveProductAttribute(product.ID, req.CategoryID, attribute)
		if err != nil {
			return fmt.Errorf("Ошибка при сохранении атрибута '%s': %v", attribute.Name, err)
		}
	}

	// Обработка вариаций
	for i, variationReq := range variations {
		err := p.AddProductVariation(variationReq, product.ID, req.CategoryID, supplier.ID, i)
		if err != nil {
			return fmt.Errorf("Ошибка при добавлении вариации: %v", err)
		}
	}

	return nil
}

func (p *ProductService) SaveProductAttribute(productID int, categoryID int, attribute models.AttributeValueRequest) error {
	// Получаем ID атрибута по имени и категории
	attributeID, err := db.GetAttributeIDByNameAndCategory(attribute.Name, categoryID)
	if err != nil {
		return fmt.Errorf("не удалось получить ID атрибута '%s': %v", attribute.Name, err)
	}

	// Проверяем, что атрибут не является linked (is_linked = false)
	isLinked, err := db.IsAttributeLinked(attributeID)
	if err != nil {
		return fmt.Errorf("Ошибка при проверке is_linked для атрибута '%s': %v", attribute.Name, err)
	}
	if isLinked {
		return fmt.Errorf("Атрибут '%s' является linked и не может быть общим для продукта", attribute.Name)
	}

	// Создаем или получаем значение атрибута
	attributeValueID, err := db.CreateOrGetAttributeValue(attributeID, attribute.Value)
	if err != nil {
		return fmt.Errorf("Ошибка при создании или получении значения атрибута '%s': %v", attribute.Name, err)
	}

	// Связываем атрибут с продуктом
	productAttributeValue := &models.ProductAttributeValue{
		ProductID:        productID,
		AttributeValueID: attributeValueID,
	}

	if err := db.CreateProductAttributeValue(productAttributeValue); err != nil {
		return fmt.Errorf("Ошибка при сохранении атрибута для продукта: %v", err)
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

func (p *ProductService) AddProductVariation(variationReq models.ProductVariationRequest, productID int, categoryID int, supplierID int, index int) error {
	// Создаем запись для вариации
	productVariation := models.ProductVariation{
		ProductID: productID,
	}

	// Если SKU не указан, генерируем автоматически
	if variationReq.SKU == "" {
		productVariation.SKU = fmt.Sprintf("%d-%s", supplierID, utils.GenerateSKU())
	} else {
		productVariation.SKU = variationReq.SKU
	}

	// Устанавливаем цену вариации, если указана
	if variationReq.Price > 0 {
		productVariation.Price = variationReq.Price
	}

	// Сохраняем вариацию в базе данных
	if err := db.CreateProductVariation(&productVariation); err != nil {
		return fmt.Errorf("не удалось создать вариацию продукта: %v", err)
	}

	// Сохранение атрибутов для вариации
	for _, attribute := range variationReq.Attributes {
		err := p.SaveVariationAttribute(productVariation.ID, categoryID, attribute)
		if err != nil {
			return fmt.Errorf("Ошибка при сохранении атрибута вариации '%s': %v", attribute.Name, err)
		}
	}

	// Сохранение изображений для вариации
	if len(variationReq.Images) > 0 {
		err := p.SaveVariationImages(productVariation.ID, variationReq.Images)
		if err != nil {
			return fmt.Errorf("не удалось сохранить изображения для вариации: %v", err)
		}
	}

	return nil
}

func (p *ProductService) SaveVariationAttribute(variationID int, categoryID int, attribute models.AttributeValueRequest) error {
	// Получаем ID атрибута по имени и категории
	attributeID, err := db.GetAttributeIDByNameAndCategory(attribute.Name, categoryID)
	if err != nil {
		return fmt.Errorf("не удалось получить ID атрибута '%s': %v", attribute.Name, err)
	}

	// Проверяем, что атрибут является linked (is_linked = true)
	isLinked, err := db.IsAttributeLinked(attributeID)
	if err != nil {
		return fmt.Errorf("Ошибка при проверке is_linked для атрибута '%s': %v", attribute.Name, err)
	}
	if !isLinked {
		return fmt.Errorf("Атрибут '%s' не является linked и не может быть использован для вариации", attribute.Name)
	}

	// Создаем или получаем значение атрибута
	attributeValueID, err := db.CreateOrGetAttributeValue(attributeID, attribute.Value)
	if err != nil {
		return fmt.Errorf("Ошибка при создании или получении значения атрибута '%s': %v", attribute.Name, err)
	}

	// Связываем атрибут с вариацией
	variationAttributeValue := &models.VariationAttributeValue{
		ProductVariationID: variationID,
		AttributeValueID:   attributeValueID,
	}

	if err := db.CreateVariationAttributeValue(variationAttributeValue); err != nil {
		return fmt.Errorf("Ошибка при сохранении атрибута для вариации: %v", err)
	}

	return nil
}

func (p *ProductService) SaveVariationImages(variationID int, images []*multipart.FileHeader) error {
	// Создаем директорию для изображений этой вариации
	uploadDir := fmt.Sprintf("uploads/variations/%d", variationID)
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return fmt.Errorf("не удалось создать директорию для изображений вариации: %v", err)
	}

	var imagePaths []string
	for _, fileHeader := range images {
		if fileHeader != nil {
			fileName := filepath.Join(uploadDir, fileHeader.Filename)
			if err := utils.SaveUploadedFile(fileHeader, fileName); err != nil {
				return fmt.Errorf("не удалось сохранить изображение вариации: %v", err)
			}
			imagePaths = append(imagePaths, fmt.Sprintf("http://195.49.215.120/%s", fileName))
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
	}

	return nil
}

func (s *ProductService) GetSupplierByUserID(userID int) (*models.Supplier, error) {
	return db.GetSupplierByUserID(userID)
}

func (s *ProductService) AddProductImage(image *models.ProductImage) error {
	return db.CreateProductImage(image)
}
