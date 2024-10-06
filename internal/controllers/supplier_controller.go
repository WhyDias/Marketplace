// internal/controllers/supplier_controller.go

package controllers

import (
	"fmt"
	"github.com/WhyDias/Marketplace/internal/models"
	"github.com/WhyDias/Marketplace/internal/services"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
)

// RegisterSupplierRequest структура запроса для регистрации поставщика

// RegisterSupplierResponse структура ответа при регистрации поставщика
type RegisterSupplierResponse struct {
	Supplier models.Supplier `json:"supplier"`
}

// SupplierController структура контроллера поставщиков
type SupplierController struct {
	Service *services.SupplierService
}

// NewSupplierController конструктор контроллера поставщиков
func NewSupplierController(service *services.SupplierService) *SupplierController {
	return &SupplierController{
		Service: service,
	}
}

// SupplierService структура сервиса поставщиков
type SupplierService struct{}

// internal/controllers/supplier_controller.go

type UpdateSupplierDetailsRequest struct {
	MarketID   int    `json:"market_id" binding:"required"`
	Place      string `json:"place" binding:"required"`
	RowName    string `json:"row_name" binding:"required"`
	Categories []int  `json:"categories" binding:"required"`
}

// UpdateSupplierDetails обновляет данные поставщика.
// @Summary      Обновление данных поставщика
// @Description  Обновляет информацию о поставщике.
// @Tags         Поставщик
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        input  body      UpdateSupplierDetailsRequest  true  "Данные для обновления"
// @Success      200    {object}  models.UpdateSupplierDetailsResponse
// @Failure      400    {object}  ErrorResponse
// @Failure      401    {object}  ErrorResponse
// @Failure      500    {object}  ErrorResponse
// @Router       /supplier/update_details [post]
func (sc *SupplierController) UpdateSupplierDetails(c *gin.Context) {
	var req UpdateSupplierDetailsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем user_id из контекста
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не аутентифицирован"})
		return
	}

	userID, ok := userIDInterface.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Неверный формат user_id"})
		return
	}

	// Обновляем детали поставщика
	err := sc.Service.UpdateSupplierDetailsByUserID(userID, req.MarketID, req.Place, req.RowName, req.Categories)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось обновить данные поставщика"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Данные поставщика успешно обновлены",
	})
}

// GetMarkets возвращает список доступных рынков.
// @Summary      Получение списка рынков
// @Description  Возвращает список всех доступных рынков.
// @Tags         Справочники
// @Accept       json
// @Produce      json
// @Success      200  {array}   models.Market
// @Failure      500  {object}  ErrorResponse
// @Router       /markets [get]
func (sc *SupplierController) GetMarkets(c *gin.Context) {
	markets, err := sc.Service.GetAllMarkets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Не удалось получить список рынков"})
		return
	}

	c.JSON(http.StatusOK, markets)
}

// GetCategories возвращает список доступных категорий товаров.
// @Summary      Получение списка категорий
// @Description  Возвращает список всех доступных категорий товаров вместе с URL изображений.
// @Tags         Справочники
// @Accept       json
// @Produce      json
// @Success      200  {array}   models.Category
// @Failure      500  {object}  ErrorResponse
// @Router       /categories [get]
func (sc *SupplierController) GetCategories(c *gin.Context) {
	categories, err := sc.Service.GetAllCategories()
	if err != nil {
		log.Printf("GetCategories: ошибка при получении категорий: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Не удалось получить список категорий"})
		return
	}

	c.JSON(http.StatusOK, categories)
}

// GetCategoryByPath возвращает категорию по заданному path.
// @Summary      Получение категории по path
// @Description  Возвращает информацию о категории на основе переданного path.
// @Tags         Категории
// @Accept       json
// @Produce      json
// @Param        path  query     string  true  "Path категории"
// @Success      200   {object}  models.Category
// @Failure      400   {object}  ErrorResponse
// @Failure      404   {object}  ErrorResponse
// @Failure      500   {object}  ErrorResponse
// @Router       /api/categories/search [get]
func (sc *SupplierController) GetCategoryByPath(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Параметр path обязателен"})
		return
	}

	category, err := sc.Service.GetCategoryByPath(path)
	if err != nil {
		if err.Error() == fmt.Sprintf("категория не найдена для path %s", path) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Категория не найдена"})
		} else {
			log.Printf("GetCategoryByPath: ошибка при получении категории по path %s: %v", path, err)
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Не удалось получить категорию"})
		}
		return
	}

	c.JSON(http.StatusOK, category)
}

// AddCategoryRequest структура запроса для добавления категории
type AddCategoryRequest struct {
	Name     string `json:"name" binding:"required"`
	Path     string `json:"path" binding:"required"`
	ImageURL string `json:"image_url" binding:"required,url"`
}

// AddCategoryResponse структура ответа при добавлении категории
type AddCategoryResponse struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	ImageURL string `json:"image_url"`
}

// AddCategory добавляет новую категорию с изображением
// @Summary Добавить новую категорию
// @Description Добавляет новую категорию с изображением
// @Tags Categories
// @Accept multipart/form-data
// @Produce json
// @Param Authorization header string true "Bearer <token>"
// @Param name formData string true "Название категории"
// @Param path formData string true "Путь категории"
// @Param image formData file true "Изображение категории"
// @Success 201 {object} AddCategoryResponse "Категория успешно добавлена"
// @Failure 400 {object} utils.ErrorResponse "Неверный формат данных или ошибки валидации"
// @Failure 401 {object} utils.ErrorResponse "Необходима авторизация"
// @Failure 500 {object} utils.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/categories [post]
func (sc *SupplierController) AddCategory(c *gin.Context) {
	name := c.PostForm("name")
	path := c.PostForm("path")

	file, err := c.FormFile("image")
	if err != nil {
		log.Printf("AddCategory: ошибка при получении изображения: %v", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Необходимо предоставить изображение категории"})
		return
	}

	// Создаем директорию для изображений категорий, если её нет
	uploadDir := "uploads/categories"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		log.Printf("AddCategory: ошибка при создании директории для изображений: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Не удалось создать директорию для изображений"})
		return
	}

	// Генерируем путь к файлу
	filePath := fmt.Sprintf("%s/%s", uploadDir, file.Filename)
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		log.Printf("AddCategory: ошибка при сохранении изображения: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Не удалось сохранить изображение"})
		return
	}

	// Добавление категории через сервисный слой
	category, err := sc.Service.AddCategory(name, path, filePath)
	if err != nil {
		log.Printf("AddCategory: ошибка при добавлении категории: %v", err)
		if err.Error() == "категория с path '"+path+"' уже существует" {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Не удалось добавить категорию"})
		}
		return
	}

	// Формирование ответа
	response := AddCategoryResponse{
		ID:       category.ID,
		Name:     category.Name,
		Path:     category.Path,
		ImageURL: category.ImageURL,
	}

	c.JSON(http.StatusCreated, response)
}
