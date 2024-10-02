// internal/services/category_service.go

package services

import (
	"github.com/WhyDias/Marketplace/internal/db"
	"github.com/WhyDias/Marketplace/internal/models"
	"log"
)

type CategoryService struct{}

func NewCategoryService() *CategoryService {
	return &CategoryService{}
}

func (s *CategoryService) GetSubcategoriesByPath(path string) ([]models.Category, error) {
	categories, err := db.GetSubcategoriesByPath(path)
	if err != nil {
		log.Printf("GetSubcategoriesByPath: ошибка при получении подкатегорий для path %s: %v", path, err)
		return nil, err
	}
	return categories, nil
}
