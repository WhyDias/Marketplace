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

func (s *CategoryService) AddCategoryAttributes(userID int, req *models.AddCategoryAttributesRequest) error {
	// Здесь можно добавить проверку прав пользователя, если необходимо

	for _, attrReq := range req.Attributes {
		var valueJSON json.RawMessage

		switch attrReq.TypeOfOption {
		case "dropdown":
			// Ожидаем слайс строк
			values, ok := attrReq.Value.([]interface{})
			if !ok {
				return fmt.Errorf("некорректный тип value для dropdown")
			}
			stringValues := make([]string, len(values))
			for i, v := range values {
				str, ok := v.(string)
				if !ok {
					return fmt.Errorf("некорректное значение в dropdown")
				}
				stringValues[i] = str
			}
			valueJSON, _ = json.Marshal(stringValues)

		case "range":
			// Ожидаем слайс из двух чисел
			values, ok := attrReq.Value.([]interface{})
			if !ok || len(values) != 2 {
				return fmt.Errorf("некорректный тип value для range")
			}
			rangeValues := make([]int, 2)
			for i, v := range values {
				num, ok := v.(float64) // JSON числа unmarshaled as float64
				if !ok {
					return fmt.Errorf("некорректное значение в range")
				}
				rangeValues[i] = int(num)
			}
			valueJSON, _ = json.Marshal(rangeValues)

		case "switcher":
			// Ожидаем bool или пустое значение
			if attrReq.Value != nil {
				boolVal, ok := attrReq.Value.(bool)
				if !ok {
					return fmt.Errorf("некорректный тип value для switcher")
				}
				valueJSON, _ = json.Marshal(boolVal)
			} else {
				// Если значение отсутствует, устанавливаем false
				valueJSON, _ = json.Marshal(false)
			}

		case "text":
			// Ожидаем строку
			str, ok := attrReq.Value.(string)
			if !ok {
				return fmt.Errorf("некорректный тип value для text")
			}
			valueJSON, _ = json.Marshal(str)

		case "numeric":
			// Ожидаем число
			num, ok := attrReq.Value.(float64) // JSON числа unmarshaled as float64
			if !ok {
				return fmt.Errorf("некорректный тип value для numeric")
			}
			valueJSON, _ = json.Marshal(int(num))

		default:
			return fmt.Errorf("неподдерживаемый тип option: %s", attrReq.TypeOfOption)
		}

		// Создаем запись атрибута категории
		attribute := &models.CategoryAttribute{
			CategoryID:   req.CategoryID,
			Name:         attrReq.Name,
			Description:  &attrReq.Description,
			TypeOfOption: &attrReq.TypeOfOption,
			Value:        valueJSON,
		}

		err := db.CreateCategoryAttribute(attribute)
		if err != nil {
			log.Printf("Не удалось создать атрибут: %s. Ошибка: %v", attrReq.Name, err)
			return fmt.Errorf("не удалось создать атрибут %s: %v", attrReq.Name, err)
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
