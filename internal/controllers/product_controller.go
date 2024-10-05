// internal/controllers/product_controller.go

package controllers

import (
	"fmt"
	"github.com/WhyDias/Marketplace/internal/models"
	"github.com/WhyDias/Marketplace/internal/services"
	"github.com/WhyDias/Marketplace/internal/utils"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
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

// AddProduct обрабатывает запрос на добавление нового продукта
// @Summary Добавить новый продукт
// @Description Добавляет новый продукт вместе с его вариациями и изображениями
// @Tags Products
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer <token>"
// @Param product body dto.ProductRequest true "Данные нового продукта"
// @Success 201 {object} utils.ErrorResponse "Продукт успешно добавлен"
// @Failure 400 {object} utils.ErrorResponse "Неверный формат данных или ошибки валидации"
// @Failure 401 {object} utils.ErrorResponse "Необходима авторизация"
// @Failure 500 {object} utils.ErrorResponse "Внутренняя ошибка сервера"
// @Router /products [post]
func (p *ProductController) AddProduct(c *gin.Context) {
	// Сначала проверим авторизацию и получим userID из контекста
	//userIDInterface, exists := c.Get("userID")
	//if !exists {
	//	c.JSON(http.StatusUnauthorized, utils.ErrorResponse{Error: "Необходима авторизация"})
	//	return
	//}
	userID := 8
	//if !ok {
	//	c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "Ошибка при получении идентификатора пользователя"})
	//	return
	//}

	// Прочитаем параметры запроса и распарсим JSON-параметры продукта
	var req models.ProductRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "Некорректный запрос: неверный формат данных"})
		return
	}

	// Создаем новый продукт с помощью сервиса
	if err := p.Service.AddProduct(&req, userID, c); err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: fmt.Sprintf("Ошибка при добавлении продукта: %v", err)})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Продукт успешно добавлен"})
}

// UpdateProduct обновляет существующий продукт
// @Summary Update a product
// @Description Обновляет информацию о продукте и его вариациях
// @Tags Products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Param product body models.UpdateProductRequest true "Product data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /api/products/{id} [put]
func (pc *ProductController) UpdateProduct(c *gin.Context) {
	productIDStr := c.Param("id")
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Некорректный ID продукта"})
		return
	}

	var req models.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Получаем user_id из контекста (из JWT)
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Пользователь не авторизован"})
		return
	}
	userID := userIDInterface.(int)

	// Обновляем продукт через сервис
	err = pc.Service.UpdateProduct(userID, productID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Не удалось обновить продукт"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Продукт успешно обновлен"})
}
