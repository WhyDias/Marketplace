// internal/controllers/verification_controller.go

package controllers

import (
	"github.com/WhyDias/Marketplace/internal/models"
	"net/http"

	"github.com/WhyDias/Marketplace/internal/services"
	"github.com/WhyDias/Marketplace/internal/utils"
	"github.com/gin-gonic/gin"
)

// VerificationController структура контроллера верификации
type VerificationController struct {
	SupplierService *services.SupplierService
}

// NewVerificationController конструктор контроллера верификации
func NewVerificationController(supplierService *services.SupplierService) *VerificationController {
	return &VerificationController{
		SupplierService: supplierService,
	}
}

// RegisterRequest структура запроса для регистрации номера телефона
type RegisterRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required,e164"`
}

// RegisterResponse структура ответа при регистрации номера телефона
type RegisterResponse struct {
	Message string `json:"message"`
}

// VerifyRequest структура запроса для верификации кода
// internal/controllers/verification_controller.go

type VerifyRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required,e164"`
	Code        string `json:"code" binding:"required,len=6"`
}

// VerifyResponse структура ответа при успешной верификации
type VerifyResponse struct {
	Message string `json:"message"`
}

// Register @Summary Register phone number
// @Description Registers a phone number and sends a verification code via WhatsApp
// @Tags verification
// @Accept json
// @Produce json
// @Param phone body RegisterRequest true "Phone number"
// @Success 200 {object} RegisterResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /register [post]
func (vc *VerificationController) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Отправка кода верификации через WhatsApp
	err := utils.SendVerificationCode(req.PhoneNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to send verification code"})
		return
	}

	c.JSON(http.StatusOK, RegisterResponse{
		Message: "Код подтверждения отправлен",
	})
}

// Verify @Summary Verify phone number
// @Description Verifies a phone number with the provided code. If the phone number does not exist, a new supplier is created and marked as verified.
// @Tags verification
// @Accept json
// @Produce json
// @Param verification body VerifyRequest true "Phone number and verification code"
// @Success 200 {object} VerifyResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /verify [post]
func (vc *VerificationController) Verify(c *gin.Context) {
	var req VerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Проверка кода верификации
	isValid := utils.ValidateWhatsAppCode(req.PhoneNumber, req.Code)
	if !isValid {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid or expired verification code"})
		return
	}

	// Проверка существования номера телефона в таблице suppliers
	isVerified, err := vc.SupplierService.IsPhoneNumberVerified(req.PhoneNumber)
	if err != nil {
		if err.Error() == "supplier not found" {
			// Если поставщик не найден, создаём нового с пометкой, что он верифицирован
			newSupplier := &models.Supplier{
				PhoneNumber: req.PhoneNumber,
				IsVerified:  true,
			}

			err = vc.SupplierService.CreateSupplier(newSupplier)
			if err != nil {
				c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create supplier"})
				return
			}

			c.JSON(http.StatusOK, VerifyResponse{
				Message: "Верификация успешна и поставщик создан",
			})
			return
		}

		// Другие ошибки
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to verify phone number"})
		return
	}

	if !isVerified {
		// Обновляем статус верификации
		err := vc.SupplierService.MarkPhoneNumberAsVerified(req.PhoneNumber)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to mark phone number as verified"})
			return
		}
	}

	c.JSON(http.StatusOK, VerifyResponse{
		Message: "Верификация успешна",
	})
}
