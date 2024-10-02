// internal/controllers/attribute_controller.go

package controllers

import (
	"net/http"
	"strconv"

	"github.com/WhyDias/Marketplace/internal/models"
	"github.com/WhyDias/Marketplace/internal/services"
	"github.com/gin-gonic/gin"
)

type AttributeController struct {
	AttributeService *services.AttributeService
}

func NewAttributeController(attributeService *services.AttributeService) *AttributeController {
	return &AttributeController{
		AttributeService: attributeService,
	}
}

type AddAttributeValueImageRequest struct {
	AttributeValueID int      `json:"attribute_value_id" binding:"required"`
	ImageURLs        []string `json:"image_urls" binding:"required"`
}

func (ac *AttributeController) AddAttributeValueImage(c *gin.Context) {
	var req AddAttributeValueImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	image := &models.AttributeValueImage{
		AttributeValueID: req.AttributeValueID,
		ImageURLs:        req.ImageURLs,
	}

	err := ac.AttributeService.AddAttributeValueImage(image)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось добавить изображения атрибута"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Изображения атрибута успешно добавлены",
		"id":      image.ID,
	})
}

func (ac *AttributeController) GetAttributeValueImage(c *gin.Context) {
	attributeValueIDParam := c.Param("attribute_value_id")
	attributeValueID, err := strconv.Atoi(attributeValueIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID значения атрибута"})
		return
	}

	image, err := ac.AttributeService.GetAttributeValueImage(attributeValueID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить изображения атрибута"})
		return
	}

	c.JSON(http.StatusOK, image)
}
