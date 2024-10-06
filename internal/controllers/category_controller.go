// internal/controllers/category_controller.go

package controllers

import (
	"encoding/json"
	"github.com/WhyDias/Marketplace/internal/models"
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

// GetImmediateSubcategories
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

// GetAllCategories
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
	Description  *string     `json:"description"` // *string для поддержки NULL
	TypeOfOption string      `json:"type_of_option" binding:"required"`
	Value        interface{} `json:"value" binding:"required"`
}

// Helper function to get pointer to string
func StringPtr(s string) *string {
	return &s
}

// AddCategoryAttributes
// @Summary Добавить атрибуты к категории
// @Description Добавляет один или несколько атрибутов к заданной категории
// @Tags Categories
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer <token>"
// @Param attributes body AddCategoryAttributesRequest true "Данные атрибутов"
// @Success 201 {object} utils.ErrorResponse "Атрибуты успешно добавлены"
// @Failure 400 {object} utils.ErrorResponse "Неверный формат данных или ошибки валидации"
// @Failure 401 {object} utils.ErrorResponse "Необходима авторизация"
// @Failure 500 {object} utils.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/categories/attributes [post]
func (cc *CategoryController) AddCategoryAttributes(c *gin.Context) {
	var req models.AddCategoryAttributesRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "Неверный формат данных: " + err.Error()})
		return
	}

	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{Error: "Необходима авторизация"})
		return
	}

	// Преобразуем userID напрямую к int
	userID, ok := userIDInterface.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "Неверный формат user_id"})
		return
	}

	err := cc.Service.AddCategoryAttributes(userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "Не удалось добавить атрибуты: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, utils.ErrorResponse{Error: "Атрибуты успешно добавлены"})
}

type CategoryAttribute struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// GetCategoryAttributes
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
// @Summary Получить атрибуты категории по ID
// @Description Возвращает список атрибутов для заданной категории
// @Tags Categories
// @Accept  json
// @Produce  json
// @Param id path int true "ID категории"
// @Success 200 {array} models.CategoryAttributeResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /categories/{id}/attributes [get]
func (cc *CategoryController) GetCategoryAttributesByCategoryID(c *gin.Context) {
	// Извлекаем category_id из URL
	categoryIDStr := c.Param("id")
	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "Некорректный ID категории"})
		return
	}

	// Получаем user_id из контекста (если необходимо для прав доступа)
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{Error: "Необходима авторизация"})
		return
	}

	// Преобразуем userID напрямую к int
	userID, ok := userIDInterface.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "Неверный формат user_id"})
		return
	}

	// Получаем атрибуты категории через сервис
	attributes, err := cc.Service.GetCategoryAttributesByCategoryID(userID, categoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "Не удалось получить атрибуты: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, attributes)
}

// GetRootCategories
// @Summary Get root categories
// @Description Получает список корневых категорий
// @Tags Categories
// @Accept json
// @Produce json
// @Success 200 {array} Category
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/categories/root [get]
func (ctrl *CategoryController) GetRootCategories(c *gin.Context) {
	categories, err := ctrl.Service.GetRootCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, categories)
}
