// internal/services/attribute_service.go

package services

import (
	"github.com/WhyDias/Marketplace/internal/db"
	"github.com/WhyDias/Marketplace/internal/models"
)

type AttributeService struct{}

func NewAttributeService() *AttributeService {
	return &AttributeService{}
}

func (s *AttributeService) AddAttributeValueImage(image *models.AttributeValueImage) error {
	return db.CreateAttributeValueImage(image)
}

func (s *AttributeService) GetAttributeValueImage(attributeValueID int) (*models.AttributeValueImage, error) {
	return db.GetAttributeValueImageByAttributeValueID(attributeValueID)
}
