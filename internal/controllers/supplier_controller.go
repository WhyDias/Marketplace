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
type RegisterSupplierRequest struct {
	Name        string `json:"name" binding:"required"`
	PhoneNumber string `json:"phone_number" binding:"required,e164"`
	MarketName  string `json:"market_name" binding:"required"`
	PlacesRows  string `json:"places_rows" binding:"required"`
	Category    string `json:"category" binding:"required"`
}

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

// RegisterSupplier @Summary Register or Update a supplier
// @Description Updates supplier details if verified, or creates a new verified supplier.
// @Tags suppliers
// @Accept json
// @Produce json
// @Param supplier body RegisterSupplierRequest true "Supplier registration data"
// @Success 200 {object} VerifyResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /suppliers/register [post]
func (sc *SupplierController) RegisterSupplier(c *gin.Context) {
	var req RegisterSupplierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Проверка, верифицирован ли номер телефона
	isVerified, err := sc.Service.IsPhoneNumberVerified(req.PhoneNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to verify phone number"})
		return
	}

	if !isVerified {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Phone number not verified"})
		return
	}

	// Обновление данных поставщика
	updateReq := models.UpdateSupplierRequest{
		Name:       req.Name,
		MarketName: req.MarketName,
		PlacesRows: req.PlacesRows,
		Category:   req.Category,
	}

	err = sc.Service.UpdateSupplierDetails(req.PhoneNumber, updateReq)
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
	query := `UPDATE suppliers 
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
