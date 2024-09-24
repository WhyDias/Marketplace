// internal/middlewares/auth.go

package middlewares

import (
	"fmt"
	"github.com/WhyDias/Marketplace/pkg/jwt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

// AuthMiddleware проверяет наличие и валидность JWT токена
func AuthMiddleware(jwtService jwt.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Отсутствует заголовок Authorization"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный формат заголовка Authorization"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		token, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			// Выводим подробную информацию об ошибке
			fmt.Println("Ошибка проверки токена:", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный токен", "details": err.Error()})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(*jwt.JWTCustomClaim); ok && token.Valid {
			c.Set("user_id", claims.UserID)
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверные данные токена"})
			c.Abort()
		}
	}
}
