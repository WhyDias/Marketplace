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

func (s *CategoryService) AddCategoryAttributes(attributes []models.CategoryAttribute) error {
	// Начинаем транзакцию
	tx, err := db.DB.Begin()
	if err != nil {
		log.Printf("Не удалось начать транзакцию: %v", err)
		return fmt.Errorf("не удалось начать транзакцию: %v", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			log.Printf("Транзакция откатилась: %v", err)
		} else {
			tx.Commit()
			log.Println("Транзакция успешно завершена")
		}
	}()

	for _, attr := range attributes {
		// Валидация значения в зависимости от типа опции
		switch attr.TypeOfOption {
		case "dropdown":
			var dropdown []string
			if err := json.Unmarshal(attr.Value, &dropdown); err != nil {
				return fmt.Errorf("некорректное значение для dropdown: %v", err)
			}
		case "range":
			var rng struct {
				From string `json:"from"`
				To   string `json:"to"`
			}
			if err := json.Unmarshal(attr.Value, &rng); err != nil {
				return fmt.Errorf("некорректное значение для range: %v", err)
			}
		case "switcher":
			var sw bool
			if err := json.Unmarshal(attr.Value, &sw); err != nil {
				return fmt.Errorf("некорректное значение для switcher: %v", err)
			}
		case "text":
			var txt string
			if err := json.Unmarshal(attr.Value, &txt); err != nil {
				return fmt.Errorf("некорректное значение для text: %v", err)
			}
		case "number":
			var num int
			if err := json.Unmarshal(attr.Value, &num); err != nil {
				return fmt.Errorf("некорректное значение для number: %v", err)
			}
		default:
			return fmt.Errorf("неподдерживаемый type_of_option: %s", attr.TypeOfOption)
		}

		// Добавляем атрибут в базу данных внутри транзакции
		err = db.CreateCategoryAttributeTx(tx, &attr)
		if err != nil {
			log.Printf("Ошибка при добавлении атрибута: %v", err)
			return err
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
