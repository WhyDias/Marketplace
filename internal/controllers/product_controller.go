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

	// Извлекаем user_id из контекста (предполагается, что JWT middleware установил его)
	userID, exists := c.Get("user_id")
	if !exists {
		log.Printf("GetModeratedProducts: user_id не найден в контексте")
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Неавторизованный доступ"})
		return
	}

	// Приводим userID к типу int
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

	// Извлекаем user_id из контекста (предполагается, что JWT middleware установил его)
	userID, exists := c.Get("user_id")
	if !exists {
		log.Printf("GetUnmoderatedProducts: user_id не найден в контексте")
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Неавторизованный доступ"})
		return
	}

	// Приводим userID к типу int
	userIDInt, ok := userID.(int)
	if !ok {
		log.Printf("GetUnmoderatedProducts: user_id имеет неверный тип")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Внутренняя ошибка сервера"})
		return
	}

	// Получаем supplier_id по user_id
	supplierID, err := pc.Service.GetSupplierIDByUserID(userIDInt)
	if err != nil {
		log.Printf("GetUnmoderatedProducts: ошибка при получении supplier_id для user_id %d: %v", userIDInt, err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Не удалось получить данные поставщика"})
		return
	}

	// Получаем продукты по supplier_id и status_id
	products, err := pc.Service.GetProductsBySupplierAndStatus(supplierID, statusID)
	if err != nil {
		log.Printf("GetUnmoderatedProducts: ошибка при получении продуктов для supplier_id %d и status_id %d: %v", supplierID, statusID, err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Не удалось получить продукты без модерации"})
		return
	}

	c.JSON(http.StatusOK, products)
}

type AddProductRequest struct {
	Name        string                    `json:"name" binding:"required"`
	Description string                    `json:"description"`
	CategoryID  int                       `json:"category_id" binding:"required"`
	Images      []string                  `json:"images"`
	Variations  []ProductVariationRequest `json:"variations"`
}

type ProductVariationRequest struct {
	SKU        string                  `json:"sku" binding:"required"`
	Price      float64                 `json:"price" binding:"required"`
	Stock      int                     `json:"stock" binding:"required"`
	Images     []string                `json:"images"`
	Attributes []AttributeValueRequest `json:"attributes"`
}

type AttributeValueRequest struct {
	Name  string `json:"name" binding:"required"`
	Value string `json:"value" binding:"required"`
}

// AddProduct добавляет новый продукт
// @Summary Добавление нового продукта
// @Description Добавляет новый продукт с изображениями и вариациями
// @Tags Продукты
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param product body AddProductRequest true "Данные продукта"
// @Success 201 {object} map[string]interface{} "Продукт успешно добавлен"
// @Failure 400 {object} map[string]string "Некорректный запрос"
// @Failure 401 {object} map[string]string "Неавторизован"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/products [post]
func (pc *ProductController) AddProduct(c *gin.Context) {
	var req AddProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем userID из контекста
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

	// Получаем supplierID через SupplierService
	supplierID, err := pc.SupplierService.GetSupplierIDByUserID(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить ID поставщика"})
		return
	}

	product := models.Product{
		Name:        req.Name,
		Description: req.Description,
		CategoryID:  req.CategoryID,
		SupplierID:  supplierID,
	}

	// Добавляем изображения продукта
	if len(req.Images) > 0 {
		product.Images = []models.ProductImage{
			{
				ImageURLs: req.Images,
			},
		}
	}

	// Обрабатываем вариации
	for _, varReq := range req.Variations {
		variation := models.ProductVariation{
			SKU:   varReq.SKU,
			Price: varReq.Price,
			Stock: varReq.Stock,
		}

		// Добавляем изображения вариации
		if len(varReq.Images) > 0 {
			variation.Images = []models.ProductVariationImage{
				{
					ImageURLs: varReq.Images,
				},
			}
		}

		// Обрабатываем атрибуты вариации
		for _, attrReq := range varReq.Attributes {
			variation.Attributes = append(variation.Attributes, models.AttributeValue{
				Name:  attrReq.Name,
				Value: attrReq.Value,
			})
		}

		product.Variations = append(product.Variations, variation)
	}

	// Добавляем продукт через сервис
	err = pc.Service.AddProduct(&product)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось добавить продукт"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Продукт успешно добавлен",
		"product_id": product.ID,
	})
}
