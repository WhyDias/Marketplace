// internal/controllers/product_controller.go

package controllers

import (
	"github.com/WhyDias/Marketplace/internal/services"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

// ProductController структура контроллера продуктов
type ProductController struct {
	Service *services.ProductService
}

// NewProductController конструктор контроллера продуктов
func NewProductController(service *services.ProductService) *ProductController {
	return &ProductController{
		Service: service,
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
	supplierID, err := pc.Service.GetSupplierIDByUserID(userIDInt)
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
