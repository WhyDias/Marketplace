// internal/middlewares/auth.go

package middlewares

import (
	"github.com/WhyDias/Marketplace/internal/utils"
	"github.com/WhyDias/Marketplace/pkg/jwt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware проверяет наличие и валидность JWT токена
func AuthMiddleware(jwtService jwt.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{Error: "Authorization header required"})
			c.Abort()
			return
		}

		// Ожидается формат "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{Error: "Authorization header format must be Bearer {token}"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		token, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{Error: "Invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*jwt.JWTCustomClaim)
		if !ok || !token.Valid {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{Error: "Invalid token claims"})
			c.Abort()
			return
		}

		// Устанавливаем информацию о пользователе в контекст
		c.Set("phone_number", claims.PhoneNumber)
		c.Next()
	}
}
