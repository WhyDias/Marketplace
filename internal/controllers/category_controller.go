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
	Description  *string     `json:"description"` // *string для поддержки NULL
	TypeOfOption string      `json:"type_of_option" binding:"required"`
	Value        interface{} `json:"value" binding:"required"`
}

// Helper function to get pointer to string
func StringPtr(s string) *string {
	return &s
}

// AddCategoryAttributes обрабатывает запрос на добавление атрибутов к категории
// @Summary Добавить атрибуты к категории
// @Description Добавляет один или несколько атрибутов к заданной категории
// @Tags Categories
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer <token>"
// @Param attributes body models.AddCategoryAttributesRequest true "Данные атрибутов"
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
	categoryIDStr := c.Param("id")
	log.Printf("Получен category_id: %s", categoryIDStr)

	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil {
		log.Printf("Ошибка конвертации category_id: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный category_id"})
		return
	}

	// Проверяем существование категории
	category, err := cc.Service.GetCategoryByID(categoryID)
	if err != nil {
		log.Printf("Ошибка при получении категории: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении категории"})
		return
	}
	if category == nil {
		log.Printf("Категория с id=%d не найдена", categoryID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Категория не найдена"})
		return
	}

	// Получаем атрибуты категории
	attributes, err := cc.Service.GetCategoryAttributes(categoryID)
	if err != nil {
		log.Printf("Ошибка при получении атрибутов: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить атрибуты категории"})
		return
	}

	// Преобразуем атрибуты в структуру ответа без category_id
	var response []models.CategoryAttributeResponse
	for _, attr := range attributes {
		var value interface{}
		if attr.TypeOfOption != nil {
			switch *attr.TypeOfOption {
			case "dropdown":
				var dropdown []string
				if err := json.Unmarshal(attr.Value, &dropdown); err != nil {
					log.Printf("Ошибка при маршалинге dropdown: %v", err)
					c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректное значение атрибута"})
					return
				}
				value = dropdown
			case "range":
				var rng struct {
					From string `json:"from"`
					To   string `json:"to"`
				}
				if err := json.Unmarshal(attr.Value, &rng); err != nil {
					log.Printf("Ошибка при маршалинге range: %v", err)
					c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректное значение атрибута"})
					return
				}
				value = rng
			case "switcher":
				var sw bool
				if err := json.Unmarshal(attr.Value, &sw); err != nil {
					log.Printf("Ошибка при маршалинге switcher: %v", err)
					c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректное значение атрибута"})
					return
				}
				value = sw
			case "text":
				var txt string
				if err := json.Unmarshal(attr.Value, &txt); err != nil {
					log.Printf("Ошибка при маршалинге text: %v", err)
					c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректное значение атрибута"})
					return
				}
				value = txt
			case "number":
				var num int
				if err := json.Unmarshal(attr.Value, &num); err != nil {
					log.Printf("Ошибка при маршалинге number: %v", err)
					c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректное значение атрибута"})
					return
				}
				value = num
			default:
				log.Printf("Неизвестный type_of_option: %s", *attr.TypeOfOption)
				c.JSON(http.StatusBadRequest, gin.H{"error": "Неизвестный type_of_option"})
				return
			}
		} else {
			log.Printf("TypeOfOption is NULL for attribute id=%d", attr.ID)
			c.JSON(http.StatusBadRequest, gin.H{"error": "TypeOfOption не может быть NULL"})
			return
		}

		response = append(response, models.CategoryAttributeResponse{
			ID:           attr.ID,
			Name:         attr.Name,
			Description:  attr.Description,  // *string
			TypeOfOption: attr.TypeOfOption, // *string
			Value:        value,
		})
	}

	c.JSON(http.StatusOK, response)
}
