// internal/services/category_service.go

package services

import (
	"encoding/json"
	"fmt"
	"github.com/WhyDias/Marketplace/internal/db"
	"github.com/WhyDias/Marketplace/internal/models"
	"log"
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
		if !category.ParentID.Valid {
			roots = append(roots, *categoryMap[category.ID])
		} else {
			parentID := int(category.ParentID.Int64)
			parent, exists := categoryMap[parentID]
			if exists {
				parent.Children = append(parent.Children, *categoryMap[category.ID])
			}
		}
	}

	return roots
}

func (s *CategoryService) AddCategoryAttributes(attributes []models.CategoryAttribute) error {
	for _, attr := range attributes {
		if attr.TypeOfOption == nil {
			log.Printf("AddCategoryAttributes: TypeOfOption is NULL for attribute ID=%d", attr.ID)
			return fmt.Errorf("TypeOfOption не может быть NULL для атрибута ID=%d", attr.ID)
		}

		switch *attr.TypeOfOption {
		case "dropdown":
			var dropdown []string
			if err := json.Unmarshal(attr.Value, &dropdown); err != nil {
				log.Printf("AddCategoryAttributes: ошибка при маршалинге dropdown для атрибута ID=%d: %v", attr.ID, err)
				return fmt.Errorf("Некорректное значение dropdown для атрибута ID=%d", attr.ID)
			}
			// Дополнительная логика для dropdown
		case "range":
			var rng struct {
				From string `json:"from"`
				To   string `json:"to"`
			}
			if err := json.Unmarshal(attr.Value, &rng); err != nil {
				log.Printf("AddCategoryAttributes: ошибка при маршалинге range для атрибута ID=%d: %v", attr.ID, err)
				return fmt.Errorf("Некорректное значение range для атрибута ID=%d", attr.ID)
			}
			// Дополнительная логика для range
		case "switcher":
			var sw bool
			if err := json.Unmarshal(attr.Value, &sw); err != nil {
				log.Printf("AddCategoryAttributes: ошибка при маршалинге switcher для атрибута ID=%d: %v", attr.ID, err)
				return fmt.Errorf("Некорректное значение switcher для атрибута ID=%d", attr.ID)
			}
			// Дополнительная логика для switcher
		case "text":
			var txt string
			if err := json.Unmarshal(attr.Value, &txt); err != nil {
				log.Printf("AddCategoryAttributes: ошибка при маршалинге text для атрибута ID=%d: %v", attr.ID, err)
				return fmt.Errorf("Некорректное значение text для атрибута ID=%d", attr.ID)
			}
			// Дополнительная логика для text
		case "number":
			var num int
			if err := json.Unmarshal(attr.Value, &num); err != nil {
				log.Printf("AddCategoryAttributes: ошибка при маршалинге number для атрибута ID=%d: %v", attr.ID, err)
				return fmt.Errorf("Некорректное значение number для атрибута ID=%d", attr.ID)
			}
			// Дополнительная логика для number
		default:
			log.Printf("AddCategoryAttributes: Неизвестный type_of_option: %s для атрибута ID=%d", *attr.TypeOfOption, attr.ID)
			return fmt.Errorf("Неизвестный type_of_option: %s для атрибута ID=%d", *attr.TypeOfOption, attr.ID)
		}

		// Добавление атрибута в базу данных
		err := db.AddCategoryAttribute(attr)
		if err != nil {
			log.Printf("AddCategoryAttributes: ошибка при добавлении атрибута ID=%d: %v", attr.ID, err)
			return fmt.Errorf("Не удалось добавить атрибут ID=%d: %v", attr.ID, err)
		}
	}

	return nil
}

// GetCategoryByID возвращает категорию по её ID
func (s *CategoryService) GetCategoryByID(categoryID int) (*models.Category, error) {
	return db.GetCategoryByID(categoryID)
}

// GetCategoryAttributes возвращает атрибуты для заданной категории
func (s *CategoryService) GetCategoryAttributes(categoryID int) ([]models.CategoryAttribute, error) {
	return db.GetCategoryAttributes(categoryID)
}
