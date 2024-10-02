// internal/services/category_service.go

package services

import (
	"github.com/WhyDias/Marketplace/internal/db"
	"github.com/WhyDias/Marketplace/internal/models"
)

type CategoryService struct{}

func NewCategoryService() *CategoryService {
	return &CategoryService{}
}

func (s *CategoryService) GetImmediateSubcategoriesByPath(path string) ([]models.Category, error) {
	categories, err := db.GetImmediateSubcategoriesByPath(path)
	if err != nil {
		return nil, err
	}
	return categories, nil
}

// GetAllCategoriesTree возвращает все категории в виде иерархического дерева
func (s *CategoryService) GetAllCategoriesTree() ([]models.CategoryNode, error) {
	categories, err := db.GetAllCategories()
	if err != nil {
		return nil, err
	}

	categoryTree := buildCategoryTree(categories)
	return categoryTree, nil
}

// buildCategoryTree строит дерево категорий
func buildCategoryTree(categories []models.Category) []models.CategoryNode {
	categoryMap := make(map[int]*models.CategoryNode)
	var roots []models.CategoryNode

	// Инициализируем узлы
	for _, category := range categories {
		categoryMap[category.ID] = &models.CategoryNode{
			ID:       category.ID,
			Name:     category.Name,
			Path:     category.Path,
			ImageURL: category.ImageURL,
			Children: []models.CategoryNode{},
		}
	}

	// Строим дерево
	for _, category := range categories {
		if category.ParentID == 0 {
			roots = append(roots, *categoryMap[category.ID])
		} else {
			parent, exists := categoryMap[category.ParentID]
			if exists {
				parent.Children = append(parent.Children, *categoryMap[category.ID])
			}
		}
	}

	return roots
}

func (s *CategoryService) AddCategoryAttribute(attribute *models.CategoryAttribute) error {
	return db.CreateCategoryAttribute(attribute)
}

func (s *CategoryService) GetCategoryAttributes(categoryID int) ([]models.CategoryAttribute, error) {
	return db.GetCategoryAttributes(categoryID)
}

func (s *CategoryService) GetCategoryByID(categoryID int) (*models.Category, error) {
	return db.GetCategoryByID(categoryID)
}
