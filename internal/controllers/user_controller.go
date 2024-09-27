// internal/controllers/user_controller.go

package controllers

import (
	"github.com/WhyDias/Marketplace/internal/services"
	_ "github.com/WhyDias/Marketplace/internal/utils"
	"github.com/WhyDias/Marketplace/pkg/jwt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ErrorResponse структура для ошибок
type ErrorResponse struct {
	Error string `json:"error"`
}

// RegisterUserRequest структура запроса для регистрации пользователя
type RegisterUserRequest struct {
	Username string `json:"username" binding:"required" validate:"min=3,max=50"`
	Password string `json:"password" binding:"required" validate:"min=6"`
}

// RegisterUserResponse структура ответа при регистрации пользователя
type RegisterUserResponse struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

// UserController структура контроллера пользователей
type UserController struct {
	Service         *services.UserService
	SupplierService *services.SupplierService
	JWT             jwt.JWTService
}

func NewUserController(service *services.UserService, supplierService *services.SupplierService, jwtService jwt.JWTService) *UserController {
	return &UserController{
		Service:         service,
		SupplierService: supplierService,
		JWT:             jwtService,
	}
}

// LoginUser выполняет аутентификацию пользователя и возвращает JWT-токен.
// @Summary      Авторизация пользователя
// @Description  Аутентифицирует пользователя и возвращает JWT-токен.
// @Tags         Авторизация
// @Accept       json
// @Produce      json
// @Param        input  body      LoginUserRequest  true  "Данные для входа"
// @Success      200    {object}  LoginUserResponse
// @Failure      400    {object}  ErrorResponse
// @Failure      401    {object}  ErrorResponse
// @Failure      500    {object}  ErrorResponse
// @Router       /api/users/login [post]
func (uc *UserController) LoginUser(c *gin.Context) {
	var req LoginUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := uc.Service.AuthenticateUser(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверное имя пользователя или пароль"})
		return
	}

	token, err := uc.JWT.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать токен"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": token,
		"expires_at":   time.Now().Add(72 * time.Hour).Format(time.RFC3339),
		"role":         user.Role,
	})
}

// LoginUserRequest структура запроса для входа пользователя
type LoginUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginUserResponse структура ответа при входе пользователя
type LoginUserResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresAt   string `json:"expires_at"`
	Role        string `json:"role"`
}

// internal/controllers/user_controller.go

type SetPasswordRequest struct {
	PhoneNumber     string `json:"phone_number" binding:"required,e164"`
	Password        string `json:"password" binding:"required,min=6"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=Password"`
}

type SetPasswordResponse struct {
	Message     string `json:"message"`
	AccessToken string `json:"access_token"`
	ExpiresAt   string `json:"expires_at"`
}

// SetPassword устанавливает пароль для пользователя после верификации номера телефона.
// @Summary      Установка пароля
// @Description  Устанавливает пароль для пользователя.
// @Tags         Авторизация
// @Accept       json
// @Produce      json
// @Param        input  body      SetPasswordRequest      true  "Данные для установки пароля"
// @Success      200    {object}  SetPasswordResponse
// @Failure      400    {object}  ErrorResponse
// @Failure      401    {object}  ErrorResponse
// @Failure      500    {object}  ErrorResponse
// @Router       /set_password [post]
func (uc *UserController) SetPassword(c *gin.Context) {
	var req SetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Проверяем, что номер телефона верифицирован
	isVerified, err := uc.SupplierService.IsPhoneNumberVerified(req.PhoneNumber)
	if err != nil || !isVerified {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Номер телефона не верифицирован"})
		return
	}

	// Создаем учетную запись пользователя
	user, err := uc.Service.RegisterUser(req.PhoneNumber, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Не удалось создать учетную запись пользователя"})
		return
	}

	// Связываем пользователя с поставщиком
	err = uc.SupplierService.LinkUserToSupplier(req.PhoneNumber, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Не удалось связать пользователя с поставщиком"})
		return
	}

	// Генерируем JWT-токен
	token, err := uc.JWT.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Не удалось создать токен"})
		return
	}

	expiresAt := time.Now().Add(72 * time.Hour).Format(time.RFC3339)

	// Возвращаем ответ с токеном
	c.JSON(http.StatusOK, SetPasswordResponse{
		Message:     "Пароль установлен, учетная запись создана",
		AccessToken: token,
		ExpiresAt:   expiresAt,
	})
}

// CheckPhoneRequest структура запроса для проверки номера телефона
type CheckPhoneRequest struct {
	Username string `json:"username" binding:"required"` // Username используется как phone_number
}

// CheckPhoneResponse структура ответа после проверки номера телефона
type CheckPhoneResponse struct {
	Exists bool `json:"exists"`
}

// CheckPhone проверяет, существует ли номер телефона в таблице users
// @Summary      Проверка существования номера телефона
// @Description  Проверяет, существует ли пользователь с указанным номером телефона
// @Tags         Авторизация
// @Accept       json
// @Produce      json
// @Param        input  body      CheckPhoneRequest         true  "Номер телефона для проверки"
// @Success      200    {object}  CheckPhoneResponse
// @Failure      400    {object}  ErrorResponse
// @Failure      500    {object}  ErrorResponse
// @Router       /api/users/check_phone [post]
func (uc *UserController) CheckPhone(c *gin.Context) {
	var req CheckPhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Проверяем существование пользователя по phone_number (username)
	exists, err := uc.Service.CheckPhoneExists(req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Ошибка при проверке номера телефона"})
		return
	}

	c.JSON(http.StatusOK, CheckPhoneResponse{Exists: exists})
}

// RequestPasswordResetRequest структура запроса для инициации сброса пароля
type RequestPasswordResetRequest struct {
	Username string `json:"username" binding:"required"` // Username используется как phone_number
}

// RequestPasswordResetResponse структура ответа после инициации сброса пароля
type RequestPasswordResetResponse struct {
	Message string `json:"message"`
}

// ConfirmPasswordResetRequest структура запроса для подтверждения сброса пароля
type ConfirmPasswordResetRequest struct {
	Username        string `json:"username" binding:"required"`
	Code            string `json:"code" binding:"required,len=6"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=NewPassword"`
}

// ConfirmPasswordResetResponse структура ответа после подтверждения сброса пароля
type ConfirmPasswordResetResponse struct {
	Message string `json:"message"`
}

// SetNewPasswordRequest структура запроса для установки нового пароля
type SetNewPasswordRequest struct {
	Username        string `json:"username" binding:"required,e164"` // Username используется как phone_number
	NewPassword     string `json:"new_password" binding:"required,min=6"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=NewPassword"`
}

// SetNewPasswordResponse структура ответа после установки нового пароля
type SetNewPasswordResponse struct {
	Message string `json:"message"`
}

// internal/controllers/user_controller.go

// RequestPasswordReset инициирует процесс сброса пароля, отправляя код подтверждения
// @Summary      Запрос на сброс пароля
// @Description  Отправляет код подтверждения на указанный номер телефона для сброса пароля
// @Tags         Восстановление доступа
// @Accept       json
// @Produce      json
// @Param        input  body      RequestPasswordResetRequest  true  "Номер телефона для сброса пароля"
// @Success      200    {object}  RequestPasswordResetResponse
// @Failure      400    {object}  ErrorResponse
// @Failure      500    {object}  ErrorResponse
// @Router       /api/users/request_password_reset [post]
func (uc *UserController) RequestPasswordReset(c *gin.Context) {
	var req RequestPasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Ошибка при связывании JSON для RequestPasswordReset: %v", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	log.Printf("Инициация сброса пароля для пользователя %s", req.Username)
	// Проверяем, существует ли пользователь с таким username (phone_number)
	exists, err := uc.Service.CheckUserExists(req.Username)
	if err != nil {
		log.Printf("Ошибка при проверке существования пользователя %s: %v", req.Username, err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Ошибка при проверке пользователя"})
		return
	}

	if !exists {
		log.Printf("Пользователь с номером телефона %s не найден", req.Username)
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Пользователь с таким номером телефона не найден"})
		return
	}

	// Генерируем и отправляем код подтверждения
	log.Printf("Отправка кода подтверждения для пользователя %s", req.Username)
	err = uc.SupplierService.SendVerificationCode(req.Username)
	if err != nil {
		log.Printf("Не удалось отправить код подтверждения пользователю %s: %v", req.Username, err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Не удалось отправить код подтверждения"})
		return
	}

	log.Printf("Код подтверждения успешно отправлен пользователю %s", req.Username)
	c.JSON(http.StatusOK, RequestPasswordResetResponse{Message: "Код подтверждения отправлен"})
}

// internal/controllers/user_controller.go

// VerifyCode проверяет код подтверждения для указанного username.
// @Summary      Верификация кода подтверждения
// @Description  Проверяет код подтверждения, отправленный на номер телефона.
// @Tags         Восстановление доступа
// @Accept       json
// @Produce      json
// @Param        input  body      VerifyCodeRequest  true  "Данные для верификации"
// @Success      200    {object}  VerifyCodeResponse
// @Failure      400    {object}  ErrorResponse
// @Failure      401    {object}  ErrorResponse
// @Failure      500    {object}  ErrorResponse
// @Router       /api/users/verify_code [post]
func (uc *UserController) VerifyCode(c *gin.Context) {
	var req VerifyCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Проверяем корректность кода подтверждения
	isValid := uc.SupplierService.ValidateVerificationCode(req.Phone_number, req.Code)
	if !isValid {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Неверный или истекший код подтверждения"})
		return
	}

	c.JSON(http.StatusOK, VerifyCodeResponse{Message: "Код подтвержден"})
}

// SetNewPassword устанавливает новый пароль для пользователя после верификации кода сброса пароля.
// @Summary      Установка нового пароля
// @Description  Устанавливает новый пароль для пользователя после верификации кода сброса пароля
// @Tags         Восстановление доступа
// @Accept       json
// @Produce      json
// @Param        input  body      SetNewPasswordRequest  true  "Данные для установки нового пароля"
// @Success      200    {object}  SetNewPasswordResponse
// @Failure      400    {object}  ErrorResponse
// @Failure      500    {object}  ErrorResponse
// @Router       /api/users/set_new_password [post]
func (uc *UserController) SetNewPassword(c *gin.Context) {
	var req SetNewPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Устанавливаем новый пароль
	err := uc.Service.ResetPassword(req.Username, req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Не удалось сбросить пароль"})
		return
	}

	c.JSON(http.StatusOK, SetNewPasswordResponse{Message: "Пароль успешно установлен"})
}
