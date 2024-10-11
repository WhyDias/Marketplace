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

func (p *ProductService) UpdateProduct(productID int, req *models.ProductRequest, userID int, attributes []models.AttributeValueRequest, variations []models.ProductVariationRequest) error {
	// Получаем поставщика по userID
	supplier, err := db.GetSupplierByUserID(userID)
	if err != nil {
		return fmt.Errorf("не удалось получить информацию о поставщике: %v", err)
	}

	// Получаем текущий продукт из базы данных
	existingProduct, err := db.GetProductByID(productID)
	if err != nil {
		return fmt.Errorf("не удалось получить продукт: %v", err)
	}

	// Проверяем, что продукт принадлежит поставщику
	if existingProduct.SupplierID != supplier.ID {
		return fmt.Errorf("вы не можете обновить этот продукт")
	}

	// Обновляем поля продукта
	updatedProduct := existingProduct
	if req.Name != "" {
		updatedProduct.Name = req.Name
	}
	if req.Description != "" {
		updatedProduct.Description = req.Description
	}
	if req.CategoryID != 0 {
		updatedProduct.CategoryID = req.CategoryID
	}
	if req.Price != 0 {
		updatedProduct.Price = req.Price
	}

	// Обновляем продукт в базе данных
	if err := db.UpdateProduct(&updatedProduct); err != nil {
		return fmt.Errorf("не удалось обновить продукт: %v", err)
	}

	// Обновление атрибутов продукта
	if len(attributes) > 0 {
		if err := p.UpdateProductAttributes(productID, updatedProduct.CategoryID, attributes); err != nil {
			return fmt.Errorf("не удалось обновить атрибуты продукта: %v", err)
		}
	}

	// Обновление вариаций продукта
	if len(variations) > 0 {
		if err := p.UpdateProductVariations(productID, updatedProduct.CategoryID, supplier.ID, variations); err != nil {
			return fmt.Errorf("не удалось обновить вариации продукта: %v", err)
		}
	}

	return nil
}

func (p *ProductService) UpdateProductAttributes(productID int, categoryID int, attributes []models.AttributeValueRequest) error {
	// Удаляем существующие атрибуты продукта
	if err := db.DeleteProductAttributes(productID); err != nil {
		return fmt.Errorf("не удалось удалить существующие атрибуты продукта: %v", err)
	}

	// Сохраняем новые атрибуты
	for _, attribute := range attributes {
		err := p.SaveProductAttribute(productID, categoryID, attribute)
		if err != nil {
			return fmt.Errorf("Ошибка при сохранении атрибута '%s': %v", attribute.Name, err)
		}
	}

	return nil
}

func (p *ProductService) UpdateProductVariations(productID int, categoryID int, supplierID int, variations []models.ProductVariationRequest) error {
	// Получаем существующие вариации продукта
	existingVariations, err := db.GetProductVariations(productID)
	if err != nil {
		return fmt.Errorf("не удалось получить существующие вариации продукта: %v", err)
	}

	// Создаем мапу для быстрого доступа по SKU
	existingVariationsMap := make(map[string]models.ProductVariation)
	for _, v := range existingVariations {
		existingVariationsMap[v.SKU] = v
	}

	for _, variationReq := range variations {
		if existingVariation, exists := existingVariationsMap[variationReq.SKU]; exists {
			// Обновляем существующую вариацию
			err := p.UpdateProductVariation(existingVariation.ID, variationReq, categoryID)
			if err != nil {
				return fmt.Errorf("не удалось обновить вариацию SKU '%s': %v", variationReq.SKU, err)
			}
			// Удаляем из мапы, чтобы в конце осталось то, что нужно удалить
			delete(existingVariationsMap, variationReq.SKU)
		} else {
			// Добавляем новую вариацию
			err := p.AddProductVariation(variationReq, productID, categoryID, supplierID, 0)
			if err != nil {
				return fmt.Errorf("не удалось добавить новую вариацию SKU '%s': %v", variationReq.SKU, err)
			}
		}
	}

	// Удаляем вариации, которые не были в новом списке
	for _, variationToDelete := range existingVariationsMap {
		err := db.DeleteProductVariation(variationToDelete.ID)
		if err != nil {
			return fmt.Errorf("не удалось удалить вариацию SKU '%s': %v", variationToDelete.SKU, err)
		}
	}

	return nil
}

func (p *ProductService) UpdateProductVariation(variationID int, variationReq models.ProductVariationRequest, categoryID int) error {
	// Обновляем поля вариации
	updatedVariation := models.ProductVariation{
		ID:    variationID,
		SKU:   variationReq.SKU,
		Price: variationReq.Price,
		Stock: variationReq.Stock,
	}

	// Обновляем вариацию в базе данных
	if err := db.UpdateProductVariation(&updatedVariation); err != nil {
		return fmt.Errorf("не удалось обновить вариацию продукта: %v", err)
	}

	// Обновляем атрибуты вариации
	if err := p.UpdateVariationAttributes(variationID, categoryID, variationReq.Attributes); err != nil {
		return fmt.Errorf("не удалось обновить атрибуты вариации: %v", err)
	}

	// Обновляем изображения вариации
	if len(variationReq.Images) > 0 {
		if err := p.UpdateVariationImages(variationID, variationReq.Images); err != nil {
			return fmt.Errorf("не удалось обновить изображения вариации: %v", err)
		}
	}

	return nil
}

func (p *ProductService) UpdateVariationAttributes(variationID int, categoryID int, attributes []models.AttributeValueRequest) error {
	// Удаляем существующие атрибуты вариации
	if err := db.DeleteVariationAttributes(variationID); err != nil {
		return fmt.Errorf("не удалось удалить существующие атрибуты вариации: %v", err)
	}

	// Сохраняем новые атрибуты
	for _, attribute := range attributes {
		err := p.SaveVariationAttribute(variationID, categoryID, attribute)
		if err != nil {
			return fmt.Errorf("Ошибка при сохранении атрибута вариации '%s': %v", attribute.Name, err)
		}
	}

	return nil
}

func (p *ProductService) UpdateVariationImages(variationID int, images []*multipart.FileHeader) error {
	// Удаляем существующие изображения вариации из базы данных
	if err := db.DeleteVariationImages(variationID); err != nil {
		return fmt.Errorf("не удалось удалить существующие изображения вариации: %v", err)
	}

	// Загружаем и сохраняем новые изображения
	return p.SaveVariationImages(variationID, images)
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
	var imageURLs []string
	for _, fileHeader := range images {
		if fileHeader != nil {
			// Указываем путь внутри бакета для сохранения изображений вариации
			folderPath := fmt.Sprintf("variations/%d", variationID)

			// Загружаем файл в Yandex Cloud Storage
			imageURL, err := utils.UploadFileToYandex(fileHeader, folderPath)
			if err != nil {
				return fmt.Errorf("не удалось загрузить изображение вариации: %v", err)
			}

			imageURLs = append(imageURLs, imageURL)
		}
	}

	if len(imageURLs) > 0 {
		// Преобразуем массив ссылок в JSON
		imagesJSON, err := json.Marshal(imageURLs)
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

// ApproveProduct подтверждает продукт
func (p *ProductService) ApproveProduct(productID int, userID int) error {
	// Начинаем транзакцию
	tx, err := db.DB.Begin()
	if err != nil {
		return fmt.Errorf("не удалось начать транзакцию: %v", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// Обновляем статус продукта на 4 (Подтвержденный)
	err = db.UpdateProductStatusTx(tx, productID, 4)
	if err != nil {
		return fmt.Errorf("не удалось обновить статус продукта: %v", err)
	}

	// Добавляем комментарий "Подтвержден"
	comment := models.Comment{
		UserID:    userID,
		ProductID: productID,
		Content:   "Подтвержден",
	}
	err = db.CreateCommentTx(tx, &comment)
	if err != nil {
		return fmt.Errorf("не удалось добавить комментарий: %v", err)
	}

	return nil
}

// RejectProduct отклоняет продукт с комментарием
func (p *ProductService) RejectProduct(productID int, userID int, content string) error {
	// Начинаем транзакцию
	tx, err := db.DB.Begin()
	if err != nil {
		return fmt.Errorf("не удалось начать транзакцию: %v", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// Обновляем статус продукта на 3 (Отклоненный)
	err = db.UpdateProductStatusTx(tx, productID, 3)
	if err != nil {
		return fmt.Errorf("не удалось обновить статус продукта: %v", err)
	}

	// Добавляем комментарий модератора
	comment := models.Comment{
		UserID:    userID,
		ProductID: productID,
		Content:   content,
	}
	err = db.CreateCommentTx(tx, &comment)
	if err != nil {
		return fmt.Errorf("не удалось добавить комментарий: %v", err)
	}

	return nil
}
