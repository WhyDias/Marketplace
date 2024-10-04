// internal/middlewares/auth.go

package middlewares

import (
	"github.com/WhyDias/Marketplace/pkg/jwt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

// AuthMiddleware проверяет наличие и валидность JWT токена
func AuthMiddleware(jwtService jwt.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Извлекаем токен из заголовка Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Необходима авторизация"})
			c.Abort()
			return
		}

		// Проверяем формат заголовка (должен начинаться с "Bearer")
		splitToken := strings.Split(authHeader, " ")
		if len(splitToken) != 2 || strings.ToLower(splitToken[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный формат токена"})
			c.Abort()
			return
		}

		// Проверяем токен
		tokenString := splitToken[1]
		token, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный токен"})
			c.Abort()
			return
		}

		// Извлекаем userID из claims
		claims, ok := token.Claims.(*jwt.JWTCustomClaim)
		if !ok || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Невалидный токен"})
			c.Abort()
			return
		}

		// Сохраняем userID в контексте для последующего использования
		c.Set("user_id", claims.UserID)

		// Переходим к следующему обработчику
		c.Next()
	}
}
