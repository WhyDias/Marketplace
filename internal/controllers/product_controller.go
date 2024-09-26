// internal/controllers/product_controller.go

package controllers

import (
	"net/http"
	"strconv"

	"github.com/WhyDias/Marketplace/internal/services"
	"github.com/gin-gonic/gin"
	"log"
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
// @Param page query int false "Номер страницы" default(1)
// @Param page_size query int false "Количество продуктов на странице" default(10)
// @Success 200 {array} models.Product
// @Failure 500 {object} ErrorResponse
// @Router /api/products/moderated [get]
func (pc *ProductController) GetModeratedProducts(c *gin.Context) {
	statusID := 3

	// Получение параметров пагинации
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		log.Printf("GetModeratedProducts: некорректный параметр page: %v", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Некорректный параметр page"})
		return
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil || pageSize < 1 || pageSize > 100 {
		log.Printf("GetModeratedProducts: некорректный параметр page_size: %v", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Некорректный параметр page_size"})
		return
	}

	offset := (page - 1) * pageSize

	products, err := pc.Service.GetProductsByStatusWithPagination(statusID, pageSize, offset)
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
// @Param page query int false "Номер страницы" default(1)
// @Param page_size query int false "Количество продуктов на странице" default(10)
// @Success 200 {array} models.Product
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/products/unmoderated [get]
func (pc *ProductController) GetUnmoderatedProducts(c *gin.Context) {
	statusID := 2

	// Получение параметров пагинации
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		log.Printf("GetUnmoderatedProducts: некорректный параметр page: %v", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Некорректный параметр page"})
		return
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil || pageSize < 1 || pageSize > 100 {
		log.Printf("GetUnmoderatedProducts: некорректный параметр page_size: %v", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Некорректный параметр page_size"})
		return
	}

	offset := (page - 1) * pageSize

	products, err := pc.Service.GetProductsByStatusWithPagination(statusID, pageSize, offset)
	if err != nil {
		log.Printf("GetUnmoderatedProducts: ошибка при получении продуктов со статусом %d: %v", statusID, err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Не удалось получить продукты без модерации"})
		return
	}

	c.JSON(http.StatusOK, products)
}
