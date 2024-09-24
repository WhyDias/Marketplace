// internal/controllers/supplier_controller.go

package controllers

import (
	"fmt"
	"github.com/WhyDias/Marketplace/internal/db"
	"net/http"
	"time"

	"github.com/WhyDias/Marketplace/internal/models"
	"github.com/WhyDias/Marketplace/internal/services"
	"github.com/gin-gonic/gin"
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

// RegisterSupplier @Summary Register a supplier
// @Description Registers a supplier and sends a verification code.
// @Tags suppliers
// @Accept json
// @Produce json
// @Param supplier body RegisterSupplierRequest true "Supplier registration data"
// @Success 200 {object} RegisterResponse
// @Failure 400 {object} ErrorResponse
// @Router /suppliers/register [post]
func (sc *SupplierController) RegisterSupplier(c *gin.Context) {
	var req models.RegisterSupplierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Логика верификации и обновления
	err := sc.Service.UpdateSupplierDetails(req.PhoneNumber, req.BazaarID, req.PlaceName, req.RowName, req.Categories)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update supplier details"})
		return
	}

	c.JSON(http.StatusOK, VerifyResponse{
		Message: "Supplier details updated successfully",
	})
}

// GetSupplierInfo @Summary Get supplier information
// @Description Retrieves information about the authenticated supplier
// @Tags suppliers
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.Supplier
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /suppliers/info [get]
func (sc *SupplierController) GetSupplierInfo(c *gin.Context) {
	phoneNumber, exists := c.Get("phone_number")
	if !exists {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Phone number not found in context"})
		return
	}

	supplier, err := sc.Service.GetSupplierInfo(phoneNumber.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Supplier not found"})
		return
	}

	c.JSON(http.StatusOK, supplier)
}

// GetSuppliers @Summary Get all suppliers
// @Description Retrieves a list of all suppliers
// @Tags suppliers
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Supplier
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /suppliers [get]
func (sc *SupplierController) GetSuppliers(c *gin.Context) {
	suppliers, err := sc.Service.GetAllSuppliers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch suppliers"})
		return
	}

	c.JSON(http.StatusOK, suppliers)
}

// SupplierService структура сервиса поставщиков
type SupplierService struct{}

// UpdateSupplierDetails обновляет детали поставщика
func (s *SupplierService) UpdateSupplierDetails(phoneNumber string, req models.UpdateSupplierRequest) error {
	query := `UPDATE supplier 
	          SET name = $1, 
	              market_name = $2, 
	              places_rows = $3, 
	              category = $4, 
	              updated_at = $5 
	          WHERE phone_number = $6`

	result, err := db.DB.Exec(query, req.Name, req.MarketName, req.PlacesRows, req.Category, time.Now(), phoneNumber)
	if err != nil {
		return fmt.Errorf("failed to update supplier details: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error fetching rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("supplier with phone number %s not found", phoneNumber)
	}

	return nil
}

// GetBazaarList @Summary Get a list of bazaars
// @Description Retrieves a list of all bazaars
// @Tags suppliers
// @Produce json
// @Success 200 {array} models.Bazaar
// @Failure 500 {object} ErrorResponse
// @Router /suppliers/bazaars [get]
func (sc *SupplierController) GetBazaarList(c *gin.Context) {
	bazaars, err := sc.Service.GetAllBazaars()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch bazaars"})
		return
	}

	c.JSON(http.StatusOK, bazaars)
}

// CreatePlace @Summary Create a new place
// @Description Creates a new place for a bazaar
// @Tags suppliers
// @Accept json
// @Produce json
// @Param place body models.Place true "Place details"
// @Success 201 {object} models.Place
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /suppliers/places [post]
func (sc *SupplierController) CreatePlace(c *gin.Context) {
	var req models.Place
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if err := sc.Service.CreatePlace(&req); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create place"})
		return
	}

	c.JSON(http.StatusCreated, req)
}

// CreateRow @Summary Create a new row
// @Description Creates a new row for a place
// @Tags suppliers
// @Accept json
// @Produce json
// @Param row body models.Row true "Row details"
// @Success 201 {object} models.Row
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /suppliers/rows [post]
func (sc *SupplierController) CreateRow(c *gin.Context) {
	var req models.Row
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if err := sc.Service.CreateRow(&req); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create row"})
		return
	}

	c.JSON(http.StatusCreated, req)
}

// UpdateSupplier @Summary Update supplier information
// @Description Updates supplier details including market, places, rows, and categories
// @Tags suppliers
// @Accept json
// @Produce json
// @Param supplier body RegisterSupplierRequest true "Supplier registration data"
// @Success 200 {object} VerifyResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /suppliers/update [post]
func (sc *SupplierController) UpdateSupplier(c *gin.Context) {
	var req models.RegisterSupplierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	err := sc.Service.UpdateSupplierDetails(req.PhoneNumber, req.BazaarID, req.PlaceName, req.RowName, req.Categories)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update supplier data"})
		return
	}

	c.JSON(http.StatusOK, VerifyResponse{
		Message: "Supplier details updated successfully",
	})
}

// internal/controllers/supplier_controller.go

type UpdateSupplierDetailsRequest struct {
	MarketID   int   `json:"market_id" binding:"required"`
	PlaceID    int   `json:"place_id" binding:"required"`
	RowID      int   `json:"row_id" binding:"required"`
	Categories []int `json:"categories" binding:"required"`
}

func (sc *SupplierController) UpdateSupplierDetails(c *gin.Context) {
	var req UpdateSupplierDetailsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Получаем user_id из контекста
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Пользователь не аутентифицирован"})
		return
	}

	// Обновляем детали поставщика
	err := sc.Service.UpdateSupplierDetailsByUserID(userID.(int), req.MarketID, req.PlaceID, req.RowID, req.Categories)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Не удалось обновить данные поставщика"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Данные поставщика успешно обновлены",
	})
}

func (sc *SupplierController) GetMarkets(c *gin.Context) {
	markets, err := sc.Service.GetAllMarkets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Не удалось получить список рынков"})
		return
	}

	c.JSON(http.StatusOK, markets)
}

func (sc *SupplierController) GetCategories(c *gin.Context) {
	categories, err := sc.Service.GetAllCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Не удалось получить список категорий"})
		return
	}

	c.JSON(http.StatusOK, categories)
}
