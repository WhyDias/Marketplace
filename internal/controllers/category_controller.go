// internal/controllers/category_controller.go

package controllers

import (
	"github.com/WhyDias/Marketplace/internal/services"
	"github.com/gin-gonic/gin"
	"net/http"
)

type CategoryController struct {
	Service *services.CategoryService
}

func NewCategoryController(service *services.CategoryService) *CategoryController {
	return &CategoryController{
		Service: service,
	}
}

// GetSubcategoriesByPath возвращает подкатегории для заданного path.
// @Summary      Получение подкатегорий по path
// @Description  Возвращает список всех подкатегорий для заданного path категории.
// @Tags         Категории
// @Accept       json
// @Produce      json
// @Param        path  query     string  true  "Path категории"
// @Success      200   {array}   Category
// @Failure      400   {object}  ErrorResponse
// @Failure      500   {object}  ErrorResponse
// @Router       /api/categories/subcategories [get]
func (cc *CategoryController) GetSubcategoriesByPath(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Параметр path обязателен"})
		return
	}

	categories, err := cc.Service.GetSubcategoriesByPath(path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Не удалось получить подкатегории"})
		return
	}

	c.JSON(http.StatusOK, categories)
}

type Category struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	ImageURL string `json:"image_url"`
}
