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

// GetModeratedProducts получает продукты со статусом 3 (с модерацией)
// @Summary Получить продукты с модерацией
// @Description Возвращает список продуктов, находящихся на модерации (status_id = 3)
// @Tags Products
// @Accept json
// @Produce json
// @Success 200 {array} models.Product
// @Failure 500 {object} ErrorResponse
// @Router /api/products/moderated [get]
func (pc *ProductController) GetModeratedProducts(c *gin.Context) {
	statusID := 3
	products, err := pc.Service.GetProductsByStatus(statusID)
	if err != nil {
		log.Printf("GetModeratedProducts: ошибка при получении продуктов со статусом %d: %v", statusID, err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Не удалось получить продукты с модерацией"})
		return
	}

	c.JSON(http.StatusOK, products)
}

// GetUnmoderatedProducts получает продукты со статусом 2 (без модерации)
// @Summary Получить продукты без модерации
// @Description Возвращает список продуктов, не находящихся на модерации (status_id = 2)
// @Tags Products
// @Accept json
// @Produce json
// @Success 200 {array} models.Product
// @Failure 500 {object} ErrorResponse
// @Router /api/products/unmoderated [get]
func (pc *ProductController) GetUnmoderatedProducts(c *gin.Context) {
	statusID := 2
	products, err := pc.Service.GetProductsByStatus(statusID)
	if err != nil {
		log.Printf("GetUnmoderatedProducts: ошибка при получении продуктов со статусом %d: %v", statusID, err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Не удалось получить продукты без модерации"})
		return
	}

	c.JSON(http.StatusOK, products)
}
