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
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Marketplace API
// @version         1.0
// @description     API для Marketplace приложения.
// @termsOfService  http://your-website.com/terms/

// @contact.name   Support Team
// @contact.url    http://your-website.com/support
// @contact.email  support@your-website.com

// @license.name  MIT License
// @license.url   https://opensource.org/licenses/MIT

// @host      195.49.215.120:8080
// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {

	// Инициализация базы данных
	time.Sleep(10 * time.Second)
	err := db.InitDB("/app/configs/config.yaml")
	if err != nil {
		log.Fatalf("Could not initialize database: %v", err)
	}

	// Инициализация JWT сервиса
	jwtService := jwt.NewJWTService("asd")

	// Инициализация сервисов
	userService := services.NewUserService()
	supplierService := services.NewSupplierService()
	productService := services.NewProductService()
	categoryService := services.NewCategoryService()
	attributeService := services.NewAttributeService()

	// Инициализация контроллеров
	attributeController := controllers.NewAttributeController(attributeService)
	productController := controllers.NewProductController(productService, supplierService)
	userController := controllers.NewUserController(userService, supplierService, jwtService)
	supplierController := controllers.NewSupplierController(supplierService)
	verificationController := controllers.NewVerificationController(supplierService, userService)
	categoryController := controllers.NewCategoryController(categoryService)

	// Создание роутера Gin
	router := gin.Default()

	router.Use(CORSMiddleware())

	// Настройка маршрута для Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Публичные маршруты (доступны без токена)
	// Публичные маршруты (доступны без токена)
	router.POST("/register", verificationController.SendVerificationCode)
	router.POST("/verify", verificationController.VerifyCode)
	router.POST("/set_password", userController.SetPassword)
	router.POST("/api/users/login", userController.LoginUser)
	router.POST("/api/users/register", userController.RegisterUser)
	router.POST("/api/users/check_phone", userController.CheckPhone)
	router.POST("/api/users/request_password_reset", userController.RequestPasswordReset)
	router.POST("/api/users/verify_code", userController.VerifyCode)          // Новый маршрут
	router.POST("/api/users/set_new_password", userController.SetNewPassword) // Новый маршрут
	router.GET("/api/categories/subcategories", categoryController.GetImmediateSubcategories)
	router.POST("/api/attributes/values/images", attributeController.AddAttributeValueImage)
	router.GET("/api/attributes/values/:attribute_value_id/images", attributeController.GetAttributeValueImage)
	router.GET("/api/categories/attributes", categoryController.GetCategoryAttributes)

	// Новый публичный маршрут для поиска категории по path
	router.GET("/api/categories/search", supplierController.GetCategoryByPath) // Перемещен в публичные маршруты
	router.GET("/api/categories", categoryController.GetAllCategories)
	//router.POST("/api/categories/attributes", categoryController.AddCategoryAttributes)
	//router.GET("/api/categories/:id/attributes", categoryController.GetCategoryAttributesByCategoryID)
	router.GET("/api/categories/:id", categoryController.GetCategoryByID)
	//router.POST("/api/products", productController.AddProduct)
	router.Static("/uploads", "./uploads")
	router.GET("/api/categories/root", categoryController.GetRootCategories)

	// Защищенные маршруты
	authorized := router.Group("/")
	authorized.Use(middlewares.AuthMiddleware(jwtService))
	{
		authorized.POST("/supplier/update_details", supplierController.UpdateSupplierDetails)
		authorized.GET("/api/products/moderated", productController.GetModeratedProducts)
		authorized.GET("/api/products/unmoderated", productController.GetUnmoderatedProducts)
		authorized.POST("/api/categories", supplierController.AddCategory)
		authorized.POST("/api/products", productController.AddProduct)
		authorized.POST("/api/categories/attributes", categoryController.AddCategoryAttributes)
		authorized.GET("/api/categories/:id/attributes", categoryController.GetCategoryAttributesByCategoryID)
		authorized.GET("/supplier/categories", supplierController.GetSupplierCategoriesHandler)
	}

	// Маршруты для получения рынков и категорий
	router.GET("/markets", supplierController.GetMarkets)
	router.GET("/categories", supplierController.GetCategories)

	// Запуск сервера
	port := ":8080"
	log.Printf("Starting server on port %s", port)
	if err := router.Run(port); err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == "http://localhost:3000" {
			c.Header("Access-Control-Allow-Origin", origin) // Укажите конкретный источник
		} else {
			c.Header("Access-Control-Allow-Origin", "*") // Для всех остальных
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
