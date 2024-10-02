// internal/controllers/category_controller.go

package controllers

import (
	"github.com/WhyDias/Marketplace/internal/services"
	"github.com/WhyDias/Marketplace/internal/utils"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type CategoryController struct {
	Service *services.CategoryService
}

func NewCategoryController(service *services.CategoryService) *CategoryController {
	return &CategoryController{
		Service: service,
	}
}

type Category struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	ImageURL string `json:"image_url"`
	ParentID int    `json:"parent_id"`
}

// GetImmediateSubcategories возвращает непосредственные подкатегории для заданного пути
// @Summary Get immediate subcategories
// @Description Получает подкатегории первого уровня для заданного пути категории
// @Tags Categories
// @Accept json
// @Produce json
// @Param path query string true "Category path"
// @Success 200 {array} Category
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/categories/subcategories [get]
func (cc *CategoryController) GetImmediateSubcategories(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Параметр 'path' обязателен"})
		return
	}

	categories, err := cc.Service.GetImmediateSubcategoriesByPath(path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Не удалось получить подкатегории"})
		return
	}

	c.JSON(http.StatusOK, categories)
}

type CategoryNode struct {
	ID       int            `json:"id"`
	Name     string         `json:"name"`
	Path     string         `json:"path"`
	ImageURL string         `json:"image_url"`
	Children []CategoryNode `json:"children,omitempty"`
}

// GetAllCategories возвращает все категории в виде дерева
// @Summary Get all categories
// @Description Получает все категории и подкатегории в иерархическом формате
// @Tags Categories
// @Accept json
// @Produce json
// @Success 200 {array} CategoryNode
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/categories [get]
func (cc *CategoryController) GetAllCategories(c *gin.Context) {
	categories, err := cc.Service.GetAllCategoriesTree()
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "Не удалось получить категории"})
		return
	}

	c.JSON(http.StatusOK, categories)
}

type AddCategoryAttributeRequest struct {
	CategoryID int    `json:"category_id" binding:"required"`
	Name       string `json:"name" binding:"required"`
}

// AddCategoryAttribute добавляет новый атрибут к категории
// @Summary Add a new attribute to a category
// @Description Добавляет новый атрибут, связанный с определенной категорией
// @Tags Categories
// @Accept json
// @Produce json
// @Param attribute body AddCategoryAttributeRequest true "Attribute to add"
// @Success 200 {object} map[string]string
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/categories/attributes [post]
func (cc *CategoryController) AddCategoryAttribute(c *gin.Context) {
	var req AddCategoryAttributeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: err.Error()})
		return
	}

	err := cc.Service.AddCategoryAttribute(req.CategoryID, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "Не удалось добавить атрибут категории"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Атрибут категории успешно добавлен"})
}

type CategoryAttribute struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// GetCategoryAttributes возвращает список атрибутов для категории
// @Summary Get category attributes
// @Description Получает список атрибутов, связанных с категорией по её ID
// @Tags Categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {array} CategoryAttribute
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/categories/{id}/attributes [get]
func (cc *CategoryController) GetCategoryAttributes(c *gin.Context) {
	categoryIDStr := c.Param("id")
	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "Некорректный ID категории"})
		return
	}

	attributes, err := cc.Service.GetCategoryAttributes(categoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "Не удалось получить атрибуты категории"})
		return
	}

	c.JSON(http.StatusOK, attributes)
}

// GetCategoryByID возвращает информацию о категории по её ID
// @Summary Get category by ID
// @Description Получает информацию о категории по её ID
// @Tags Categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} Category
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/categories/{id} [get]
func (cc *CategoryController) GetCategoryByID(c *gin.Context) {
	categoryIDStr := c.Param("id")
	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "Некорректный ID категории"})
		return
	}

	category, err := cc.Service.GetCategoryByID(categoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "Не удалось получить категорию"})
		return
	}

	c.JSON(http.StatusOK, category)
}
