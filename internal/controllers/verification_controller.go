// internal/controllers/verification_controller.go

package controllers

import (
	"github.com/WhyDias/Marketplace/internal/services"
	"github.com/gin-gonic/gin"
	"net/http"
)

// VerificationController структура контроллера верификации
type VerificationController struct {
	SupplierService *services.SupplierService
	UserService     *services.UserService
}

func NewVerificationController(supplierService *services.SupplierService, userService *services.UserService) *VerificationController {
	return &VerificationController{
		SupplierService: supplierService,
		UserService:     userService,
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

// SendVerificationCode отправляет код подтверждения на указанный номер телефона.
// @Summary      Отправка кода подтверждения
// @Description  Отправляет код подтверждения на указанный номер телефона.
// @Tags         Авторизация
// @Accept       json
// @Produce      json
// @Param        input  body      RegisterRequest  true  "Данные для регистрации"
// @Success      200    {object}  RegisterResponse
// @Failure      400    {object}  ErrorResponse
// @Failure      500    {object}  ErrorResponse
// @Router       /register [post]
func (vc *VerificationController) SendVerificationCode(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Проверяем, существует ли пользователь с указанным номером телефона
	user, err := vc.UserService.GetUserByUsername(req.PhoneNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Ошибка при проверке пользователя"})
		return
	}

	if user == nil {
		// Если пользователь не существует, создаем нового пользователя без пароля
		user, err = vc.UserService.RegisterUser(req.PhoneNumber, "", []string{"supplier"})
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Не удалось создать пользователя"})
			return
		}
	}

	// Отправляем код подтверждения
	err = vc.SupplierService.SendVerificationCode(user.ID, req.PhoneNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Не удалось отправить код подтверждения"})
		return
	}

	c.JSON(http.StatusOK, RegisterResponse{
		Message: "Код подтверждения отправлен",
	})
}

type VerifyCodeRequest struct {
	Phone_number string `json:"phone_number"`
	Code         string `json:"code"`
}
type VerifyCodeResponse struct {
	Message string `json:"message"`
}

// VerifyCode проверяет код подтверждения для указанного номера телефона.
// @Summary      Верификация кода подтверждения
// @Description  Проверяет код подтверждения, отправленный на номер телефона.
// @Tags         Авторизация
// @Accept       json
// @Produce      json
// @Param        input  body      VerifyCodeRequest  true  "Данные для верификации"
// @Success      200    {object}  VerifyCodeResponse
// @Failure      400    {object}  ErrorResponse
// @Failure      500    {object}  ErrorResponse
// @Router       /verify [post]
func (vc *VerificationController) VerifyCode(c *gin.Context) {
	var req VerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Получаем пользователя по номеру телефона
	user, err := vc.UserService.GetUserByUsername(req.PhoneNumber)
	if err != nil || user == nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Пользователь не найден"})
		return
	}

	// Проверяем код подтверждения
	isValid := vc.SupplierService.ValidateVerificationCode(user.ID, req.Code)
	if !isValid {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Неверный или истекший код подтверждения"})
		return
	}

	// Отмечаем пользователя как верифицированного
	err = vc.SupplierService.MarkPhoneNumberAsVerified(req.PhoneNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Не удалось обновить статус верификации"})
		return
	}

	c.JSON(http.StatusOK, VerifyResponse{
		Message: "Верификация успешна",
	})
}
