// internal/middlewares/auth.go

package middlewares

import (
	"github.com/WhyDias/Marketplace/pkg/jwt"
	"github.com/gin-gonic/gin"
	"net/http"
)

// AuthMiddleware проверяет наличие и валидность JWT токена
func AuthMiddleware(jwtService jwt.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Отсутствует заголовок Authorization"})
			c.Abort()
			return
		}

		token, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный токен"})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(*jwt.JWTCustomClaim); ok && token.Valid {
			// Устанавливаем информацию о пользователе в контекст
			c.Set("user_id", claims.UserID)
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверные клеймы токена"})
			c.Abort()
		}
	}
}
