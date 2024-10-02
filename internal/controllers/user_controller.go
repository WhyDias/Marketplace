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

// LoginUserRequest структура запроса для логина
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse структура ответа при успешном логине
type LoginResponse struct {
	Message     string `json:"message"`
	AccessToken string `json:"access_token"`
	ExpiresAt   int64  `json:"expires_at"` // Unix timestamp
}

// LoginUser обрабатывает логин пользователя
// @Summary      Логин пользователя
// @Description  Аутентифицирует пользователя и возвращает JWT-токен с ролями.
// @Tags         Авторизация
// @Accept       json
// @Produce      json
// @Param        input  body      LoginRequest      true  "Данные для логина"
// @Success      200    {object}  LoginResponse
// @Failure      400    {object}  ErrorResponse
// @Failure      401    {object}  ErrorResponse
// @Failure      500    {object}  ErrorResponse
// @Router       /api/users/login [post]
func (uc *UserController) LoginUser(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("LoginUser: ошибка при связывании JSON: %v", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	user, err := uc.Service.AuthenticateUser(req.Username, req.Password)
	if err != nil {
		log.Printf("LoginUser: аутентификация не удалась для пользователя %s: %v", req.Username, err)
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Неверное имя пользователя или пароль"})
		return
	}

	var roleNames []string
	for _, role := range user.Roles {
		roleNames = append(roleNames, role.Name)
	}

	// Generate token with roles
	token, err := uc.JWT.GenerateTokenWithRoles(user.ID, roleNames)
	if err != nil {
		log.Printf("LoginUser: ошибка при генерации токена для пользователя %d: %v", user.ID, err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Не удалось создать токен"})
		return
	}

	expiresAt := time.Now().Add(72 * time.Hour).Unix()

	response := LoginResponse{
		Message:     "Успешный вход",
		AccessToken: token,
		ExpiresAt:   expiresAt,
	}

	c.JSON(http.StatusOK, response)
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
// @Description  Устанавливает пароль для пользователя и возвращает JWT-токен с ролями.
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

	// Создаем учетную запись пользователя с ролью 'supplier'
	user, err := uc.Service.RegisterUser(req.PhoneNumber, req.Password, []string{"supplier"})
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

	// Генерируем JWT-токен с ролями
	var roleNames []string
	for _, role := range user.Roles {
		roleNames = append(roleNames, role.Name)
	}

	token, err := uc.JWT.GenerateTokenWithRoles(user.ID, roleNames)
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

	// Получаем пользователя по номеру телефона (username)
	user, err := uc.Service.GetUserByUsername(req.Username)
	if err != nil || user == nil {
		log.Printf("Пользователь с номером телефона %s не найден", req.Username)
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Пользователь не найден"})
		return
	}

	// Отправляем код подтверждения
	log.Printf("Отправка кода подтверждения для пользователя %s", req.Username)
	err = uc.SupplierService.SendVerificationCode(user.ID)
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

	// Получаем пользователя по номеру телефона (username)
	user, err := uc.Service.GetUserByUsername(req.Phone_number)
	if err != nil || user == nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Пользователь не найден"})
		return
	}

	// Проверяем код подтверждения
	isValid := uc.SupplierService.ValidateVerificationCode(user.ID, req.Code)
	if !isValid {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Неверный или истекший код подтверждения"})
		return
	}

	c.JSON(http.StatusOK, VerifyCodeResponse{Message: "Код подтверждён"})
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

// RegisterUserRequest структура запроса для регистрации пользователя
type RegisterUserRequest struct {
	Username string   `json:"username" binding:"required"`
	Email    string   `json:"email" binding:"required,email"`
	Password string   `json:"password" binding:"required,min=6"`
	Roles    []string `json:"roles" binding:"required"` // Добавлено поле для ролей
}

// RegisterUserResponse структура ответа при регистрации пользователя
type RegisterUserResponse struct {
	Message     string `json:"message"`
	AccessToken string `json:"access_token"`
	ExpiresAt   string `json:"expires_at"`
}

// RegisterUser регистрирует нового пользователя и возвращает токен
// @Summary      Регистрация пользователя
// @Description  Регистрация нового пользователя с указанными ролями.
// @Tags         Пользователи
// @Accept       json
// @Produce      json
// @Param        input  body      RegisterUserRequest  true  "Данные для регистрации пользователя"
// @Success      201    {object}  RegisterUserResponse
// @Failure      400    {object}  ErrorResponse
// @Failure      500    {object}  ErrorResponse
// @Router       /api/users/register [post]
func (uc *UserController) RegisterUser(c *gin.Context) {
	var req RegisterUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("RegisterUser: ошибка при связывании JSON: %v", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Регистрация пользователя через сервисный слой
	user, err := uc.Service.RegisterUser(req.Username, req.Password, req.Roles)
	if err != nil {
		log.Printf("RegisterUser: ошибка при регистрации пользователя: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	// Извлекаем имена ролей
	var roleNames []string
	for _, role := range user.Roles {
		roleNames = append(roleNames, role.Name)
	}

	// Генерация JWT-токена
	token, err := uc.JWT.GenerateTokenWithRoles(user.ID, roleNames)
	if err != nil {
		log.Printf("RegisterUser: ошибка при генерации токена: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Не удалось создать токен"})
		return
	}

	expiresAt := time.Now().Add(72 * time.Hour).Format(time.RFC3339)

	// Возвращаем ответ с токеном
	c.JSON(http.StatusCreated, RegisterUserResponse{
		Message:     "Пользователь успешно зарегистрирован",
		AccessToken: token,
		ExpiresAt:   expiresAt,
	})
}
