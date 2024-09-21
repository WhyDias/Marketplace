// internal/controllers/user_controller.go

package controllers

import (
	"github.com/WhyDias/Marketplace/internal/services"
	_ "github.com/WhyDias/Marketplace/internal/utils"
	"github.com/WhyDias/Marketplace/pkg/jwt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
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
	Service *services.UserService
	JWT     jwt.JWTService
}

// NewUserController конструктор контроллера
func NewUserController(service *services.UserService, jwtService jwt.JWTService) *UserController {
	return &UserController{
		Service: service,
		JWT:     jwtService,
	}
}

// Validator экземпляр валидатора
var validate = validator.New()

// RegisterUser @Summary Register a new user
// @Description Registers a new user with username and password
// @Tags users
// @Accept json
// @Produce json
// @Param user body RegisterUserRequest true "User registration data"
// @Success 201 {object} RegisterUserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/users/register [post]
func (uc *UserController) RegisterUser(c *gin.Context) {
	var req RegisterUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Валидация данных
	if err := validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Регистрация пользователя
	user, err := uc.Service.RegisterUser(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, RegisterUserResponse{
		ID:        user.ID,
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
	})
}

// LoginUser @Summary Login a user
// @Description Authenticates a user and returns a JWT token
// @Tags users
// @Accept json
// @Produce json
// @Param credentials body LoginUserRequest true "User credentials"
// @Success 200 {object} LoginUserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/users/login [post]
func (uc *UserController) LoginUser(c *gin.Context) {
	var req LoginUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	user, err := uc.Service.AuthenticateUser(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid credentials"})
		return
	}

	// Генерация JWT токена
	token, err := uc.JWT.GenerateToken(user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, LoginUserResponse{
		AccessToken: token,
		ExpiresAt:   time.Now().Add(24 * time.Hour).Format(time.RFC3339),
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
}
