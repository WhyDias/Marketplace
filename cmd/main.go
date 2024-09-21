// cmd/main.go

package main

import (
	_ "github.com/WhyDias/Marketplace/docs" // Для Swagger
	"github.com/WhyDias/Marketplace/internal/controllers"
	"github.com/WhyDias/Marketplace/internal/db"
	"github.com/WhyDias/Marketplace/internal/middlewares"
	"github.com/WhyDias/Marketplace/internal/services"
	"github.com/WhyDias/Marketplace/pkg/jwt"
	"log"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Marketplace API
// @version 1.0
// @description This is the API documentation for the Marketplace application.
// @termsOfService http://yourproject.com/terms/

// @contact.name API Support
// @contact.url http://yourproject.com/support
// @contact.email support@yourproject.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {

	// Инициализация базы данных
	err := db.InitDB("/app/configs/config.yaml")
	if err != nil {
		log.Fatalf("Could not initialize database: %v", err)
	}

	// Инициализация JWT сервиса
	jwtService := jwt.NewJWTService("asd")

	// Инициализация сервисов
	userService := services.NewUserService()
	supplierService := services.NewSupplierService()

	// Инициализация контроллеров
	userController := controllers.NewUserController(userService, jwtService)
	supplierController := controllers.NewSupplierController(supplierService)
	verificationController := controllers.NewVerificationController(supplierService)

	// Создание роутера Gin
	router := gin.Default()

	// Настройка маршрута для Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Публичные маршруты (доступны без токена)
	router.POST("/register", verificationController.Register)
	router.POST("/verify", verificationController.Verify)
	router.POST("/api/users/login", userController.LoginUser)

	// Группа защищённых маршрутов
	authorized := router.Group("/")
	authorized.Use(middlewares.AuthMiddleware(jwtService))
	{
		authorized.POST("/suppliers/register", supplierController.RegisterSupplier)
		authorized.GET("/suppliers", supplierController.GetSuppliers)
		authorized.GET("/suppliers/info", supplierController.GetSupplierInfo)
		authorized.POST("/api/users/register", userController.RegisterUser)
	}

	// Запуск сервера
	port := ":8080"
	log.Printf("Starting server on port %s", port)
	if err := router.Run(port); err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}
