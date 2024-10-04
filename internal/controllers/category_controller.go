// internal/controllers/category_controller.go

package controllers

import (
	"encoding/json"
	"github.com/WhyDias/Marketplace/internal/models"
	"github.com/WhyDias/Marketplace/internal/services"
	"github.com/WhyDias/Marketplace/internal/utils"
	"github.com/gin-gonic/gin"
	"log"
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
	CategoryID   int             `json:"category_id" binding:"required"`
	Name         string          `json:"name" binding:"required"`
	Description  string          `json:"description"`
	TypeOfOption string          `json:"type_of_option" binding:"required,oneof=dropdown range switcher text number"`
	Value        json.RawMessage `json:"value" binding:"required"` // Добавляем поле Value
}

type AddCategoryAttributesRequest struct {
	CategoryID int                        `json:"category_id" binding:"required"`
	Attributes []CategoryAttributePayload `json:"attributes" binding:"required,dive,required"`
}

type CategoryAttributePayload struct {
	Name         string      `json:"name" binding:"required"`
	Description  string      `json:"description"`
	TypeOfOption string      `json:"type_of_option" binding:"required"`
	Value        interface{} `json:"value" binding:"required"`
}

// AddCategoryAttributes добавляет несколько атрибутов к категории
// @Summary Добавить несколько атрибутов к категории
// @Description Добавляет массив атрибутов, связанных с определенной категорией
// @Tags Categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Param attributes body AddCategoryAttributesRequest true "Массив атрибутов для добавления"
// @Success 200 {object} map[string]string
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/categories/{id}/attributes [post]
func (cc *CategoryController) AddCategoryAttributes(c *gin.Context) {
	var req AddCategoryAttributesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("AddCategoryAttributes: ошибка при связывании JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Подготовка атрибутов для сервиса
	var attributes []models.CategoryAttribute
	for _, attrPayload := range req.Attributes {
		// Сериализуем значение в JSON
		valueBytes, err := json.Marshal(attrPayload.Value)
		if err != nil {
			log.Printf("AddCategoryAttributes: ошибка при маршалинге значения атрибута: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректное значение атрибута"})
			return
		}

		categoryAttribute := models.CategoryAttribute{
			CategoryID:   req.CategoryID,
			Name:         attrPayload.Name,
			Description:  attrPayload.Description,
			TypeOfOption: attrPayload.TypeOfOption,
			Value:        valueBytes,
		}

		attributes = append(attributes, categoryAttribute)
	}

	// Вызов сервиса для добавления атрибутов
	err := cc.Service.AddCategoryAttributes(attributes)
	if err != nil {
		log.Printf("AddCategoryAttributes: ошибка при добавлении атрибутов: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось добавить атрибуты категории"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Атрибуты успешно добавлены к категории",
	})
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

// GetCategoryAttributesByCategoryID возвращает список атрибутов для категории
// @Summary Get category attributes by category ID
// @Description Получает список атрибутов, связанных с категорией по её ID
// @Tags Categories
// @Accept json
// @Produce json
// @Param category_id query int true "Category ID"
// @Success 200 {array} models.CategoryAttribute
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/categories/attributes [get]
func (cc *CategoryController) GetCategoryAttributesByCategoryID(c *gin.Context) {
	categoryIDStr := c.Query("category_id")
	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil || categoryID <= 0 {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "Некорректный category_id"})
		return
	}

	attributes, err := cc.Service.GetCategoryAttributes(categoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "Не удалось получить атрибуты категории"})
		return
	}

	c.JSON(http.StatusOK, attributes)
}
