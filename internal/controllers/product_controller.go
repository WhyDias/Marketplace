// internal/controllers/product_controller.go

package controllers

import (
	"github.com/WhyDias/Marketplace/internal/models"
	"github.com/WhyDias/Marketplace/internal/services"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

// ProductController структура контроллера продуктов
type ProductController struct {
	Service         *services.ProductService
	SupplierService *services.SupplierService
}

func NewProductController(productService *services.ProductService, supplierService *services.SupplierService) *ProductController {
	return &ProductController{
		Service:         productService,
		SupplierService: supplierService,
	}
}

// GetModeratedProducts возвращает список продуктов со статусом модерации (status_id = 3).
// @Summary      Получение продуктов с модерацией
// @Description  Возвращает список продуктов, находящихся на модерации (status_id = 3).
// @Tags         Продукты
// @Security     BearerAuth
// @Produce      json
// @Success      200  {array}   models.Product
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/products/moderated [get]
func (pc *ProductController) GetModeratedProducts(c *gin.Context) {
	statusID := 3

	// Извлекаем user_id из контекста
	userID, exists := c.Get("user_id")
	if !exists {
		log.Printf("GetModeratedProducts: user_id не найден в контексте")
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Неавторизованный доступ"})
		return
	}

	userIDInt, ok := userID.(int)
	if !ok {
		log.Printf("GetModeratedProducts: user_id имеет неверный тип")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Внутренняя ошибка сервера"})
		return
	}

	// Получаем supplier_id по user_id
	supplierID, err := pc.SupplierService.GetSupplierIDByUserID(userIDInt)
	if err != nil {
		log.Printf("GetModeratedProducts: ошибка при получении supplier_id для user_id %d: %v", userIDInt, err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Не удалось получить данные поставщика"})
		return
	}

	// Получаем продукты по supplier_id и status_id
	products, err := pc.Service.GetProductsBySupplierAndStatus(supplierID, statusID)
	if err != nil {
		log.Printf("GetModeratedProducts: ошибка при получении продуктов для supplier_id %d и status_id %d: %v", supplierID, statusID, err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Не удалось получить продукты с модерацией"})
		return
	}

	c.JSON(http.StatusOK, products)
}

// GetUnmoderatedProducts возвращает список продуктов без модерации (status_id = 2).
// @Summary      Получение продуктов без модерации
// @Description  Возвращает список продуктов, прошедших модерацию (status_id = 2).
// @Tags         Продукты
// @Security     BearerAuth
// @Produce      json
// @Success      200  {array}   models.Product
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/products/unmoderated [get]
func (pc *ProductController) GetUnmoderatedProducts(c *gin.Context) {
	statusID := 2

	// Извлекаем user_id из контекста
	userID, exists := c.Get("user_id")
	if !exists {
		log.Printf("GetModeratedProducts: user_id не найден в контексте")
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Неавторизованный доступ"})
		return
	}

	userIDInt, ok := userID.(int)
	if !ok {
		log.Printf("GetModeratedProducts: user_id имеет неверный тип")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Внутренняя ошибка сервера"})
		return
	}

	// Получаем supplier_id по user_id
	supplierID, err := pc.SupplierService.GetSupplierIDByUserID(userIDInt)
	if err != nil {
		log.Printf("GetModeratedProducts: ошибка при получении supplier_id для user_id %d: %v", userIDInt, err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Не удалось получить данные поставщика"})
		return
	}

	// Получаем продукты по supplier_id и status_id
	products, err := pc.Service.GetProductsBySupplierAndStatus(supplierID, statusID)
	if err != nil {
		log.Printf("GetModeratedProducts: ошибка при получении продуктов для supplier_id %d и status_id %d: %v", supplierID, statusID, err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Не удалось получить продукты с модерацией"})
		return
	}

	c.JSON(http.StatusOK, products)
}

type AddProductRequest struct {
	Name        string                         `json:"name" binding:"required"`
	Description string                         `json:"description"`
	CategoryID  int                            `json:"category_id" binding:"required"`
	Images      []string                       `json:"images"`
	Attributes  []models.AttributeValueRequest `json:"attributes"`
	Price       float64                        `json:"price" binding:"required"`
	Stock       int                            `json:"stock" binding:"required"`
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
	Name   string   `json:"name" binding:"required"`
	Values []string `json:"values" binding:"required"`
}

//type AddProductForm struct {
//	Name        string            `form:"name" binding:"required"`
//	Description string            `form:"description"`
//	CategoryID  int               `form:"category_id" binding:"required"`
//	SKU         string            `form:"sku" binding:"required"`
//	Price       float64           `form:"price" binding:"required"`
//	Stock       int               `form:"stock" binding:"required"`
//	Attributes  string            `form:"attributes" binding:"required"` // JSON-строка
//	Images      []*gin.FileHeader `form:"images" binding:"omitempty,dive,required"`
//}
//
//type ProductVariationForm struct {
//	SKU        string                         `form:"sku" binding:"required"`
//	Price      float64                        `form:"price" binding:"required"`
//	Stock      int                            `form:"stock" binding:"required"`
//	Attributes []models.AttributeValueRequest `form:"attributes" binding:"required,dive,required"`
//	Images     []*gin.FileHeader              `form:"images" binding:"omitempty,dive,required"`
//}

// AddProduct добавляет новый продукт
// @Summary Добавить новый продукт
// @Description Добавляет новый продукт с изображениями и вариациями
// @Tags Products
// @Accept multipart/form-data
// @Produce json
// @Param product body AddProductForm true "Продукт для добавления"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/products [post]
//func (pc *ProductController) AddProduct(c *gin.Context) {
//	var form AddProductForm
//	if err := c.ShouldBind(&form); err != nil {
//		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: err.Error()})
//		return
//	}
//
//	userIDInterface, exists := c.Get("user_id")
//	if !exists {
//		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{Error: "Пользователь не авторизован"})
//		return
//	}
//	userID := userIDInterface.(int)
//
//	// Получаем supplier_id и market_id
//	supplierID, err := pc.SupplierService.GetSupplierIDByUserID(userID)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "Не удалось получить supplier_id"})
//		return
//	}
//	marketID, err := pc.SupplierService.GetMarketIDBySupplierID(supplierID)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "Не удалось получить market_id"})
//		return
//	}
//
//	// Парсинг JSON-строки с атрибутами
//	var attributes []models.AttributeValueRequest
//	if err := json.Unmarshal([]byte(form.Attributes), &attributes); err != nil {
//		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "Некорректный формат атрибутов"})
//		return
//	}
//
//	product := &models.Product{
//		Name:        form.Name,
//		Description: form.Description,
//		CategoryID:  form.CategoryID,
//		SKU:         form.SKU,
//		Price:       form.Price,
//		Stock:       form.Stock,
//		Images:      []models.ProductImage{},
//		Variations:  []models.ProductVariation{},
//	}
//
//	// Обработка загрузки изображений
//	formImages := form.Images
//	for _, fileHeader := range formImages {
//		// Сохраняем файл локально
//		filename := filepath.Base(fileHeader.Filename)
//		localPath := filepath.Join("uploads", filename)
//		if err := c.SaveUploadedFile(fileHeader, localPath); err != nil {
//			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "Не удалось сохранить изображение"})
//			return
//		}
//
//		// Загружаем изображение на Яндекс.Диск и получаем публичный URL
//		yandexURL, err := pc.Service.UploadImageToYandexDisk(localPath)
//		if err != nil {
//			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "Не удалось загрузить изображение на Яндекс.Диск"})
//			return
//		}
//
//		// Добавляем URL в массив изображений продукта
//		product.Images = append(product.Images, models.ProductImage{
//			ImageURLs: []string{yandexURL},
//		})
//	}
//
//	// Добавляем продукт и вариации
//	err = pc.Service.AddProduct(product, attributes)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "Не удалось добавить продукт: " + err.Error()})
//		return
//	}
//
//	c.JSON(http.StatusCreated, gin.H{"message": "Продукт успешно добавлен", "product_id": product.ID})
//}

// UpdateProduct обновляет существующий продукт
// @Summary Обновить продукт
// @Description Обновляет существующий продукт с возможностью добавления новых изображений
// @Tags Products
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "ID продукта"
// @Param product body AddProductForm true "Продукт для обновления"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/products/{id} [put]
//func (pc *ProductController) UpdateProduct(c *gin.Context) {
//	productIDStr := c.Param("id")
//	productID, err := strconv.Atoi(productIDStr)
//	if err != nil {
//		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "Некорректный ID продукта"})
//		return
//	}
//
//	var form AddProductForm
//	if err := c.ShouldBind(&form); err != nil {
//		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: err.Error()})
//		return
//	}
//
//	userIDInterface, exists := c.Get("user_id")
//	if !exists {
//		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{Error: "Пользователь не авторизован"})
//		return
//	}
//	userID := userIDInterface.(int)
//
//	// Получаем supplier_id и market_id
//	supplierID, err := pc.SupplierService.GetSupplierIDByUserID(userID)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "Не удалось получить supplier_id"})
//		return
//	}
//	marketID, err := pc.SupplierService.GetMarketIDBySupplierID(supplierID)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "Не удалось получить market_id"})
//		return
//	}
//
//	// Парсинг JSON-строки с атрибутами
//	var attributes []models.AttributeValueRequest
//	if err := json.Unmarshal([]byte(form.Attributes), &attributes); err != nil {
//		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "Некорректный формат атрибутов"})
//		return
//	}
//
//	product := &models.Product{
//		ID:          productID,
//		Name:        form.Name,
//		Description: form.Description,
//		CategoryID:  form.CategoryID,
//		SKU:         form.SKU,
//		Price:       form.Price,
//		Stock:       form.Stock,
//		Images:      []models.ProductImage{},
//		Variations:  []models.ProductVariation{},
//	}
//
//	// Обработка загрузки изображений
//	formImages := form.Images
//	for _, fileHeader := range formImages {
//		// Сохраняем файл локально
//		filename := filepath.Base(fileHeader.Filename)
//		localPath := filepath.Join("uploads", filename)
//		if err := c.SaveUploadedFile(fileHeader, localPath); err != nil {
//			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "Не удалось сохранить изображение"})
//			return
//		}
//
//		// Загружаем изображение на Яндекс.Диск и получаем публичный URL
//		yandexURL, err := pc.Service.UploadImageToYandexDisk(localPath)
//		if err != nil {
//			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "Не удалось загрузить изображение на Яндекс.Диск"})
//			return
//		}
//
//		// Добавляем URL в массив изображений продукта
//		product.Images = append(product.Images, models.ProductImage{
//			ImageURLs: []string{yandexURL},
//		})
//	}
//
//	// Обновляем продукт и вариации
//	err = pc.Service.UpdateProduct(userID, productID, product, attributes)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "Не удалось обновить продукт: " + err.Error()})
//		return
//	}
//
//	c.JSON(http.StatusOK, gin.H{"message": "Продукт успешно обновлен"})
//}
