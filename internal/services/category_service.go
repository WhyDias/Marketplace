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
	log.Printf("User ID %d добавляет атрибуты для категории ID %d", userID, req.CategoryID)

	for _, attrReq := range req.Attributes {
		var valueJSON json.RawMessage
		var dropdownValues []string // Объявление переменной для значений dropdown
		var rangeValues []int       // Объявление переменной для значений range
		var boolVal bool            // Объявление переменной для switcher
		var textVal string          // Объявление переменной для text
		var numericVal int          // Объявление переменной для numeric

		// Обработка значений атрибутов в зависимости от типа
		switch attrReq.TypeOfOption {
		case "dropdown":
			// Ожидаем слайс строк
			values, ok := attrReq.Value.([]interface{})
			if !ok {
				return fmt.Errorf("некорректный тип value для dropdown")
			}
			dropdownValues = make([]string, len(values))
			for i, v := range values {
				str, ok := v.(string)
				if !ok {
					return fmt.Errorf("некорректное значение в dropdown")
				}
				dropdownValues[i] = str
			}
			valueJSON, _ = json.Marshal(dropdownValues)

		case "range":
			// Ожидаем слайс из двух чисел
			values, ok := attrReq.Value.([]interface{})
			if !ok || len(values) != 2 {
				return fmt.Errorf("некорректный тип value для range")
			}
			rangeValues = make([]int, 2)
			for i, v := range values {
				num, ok := v.(float64) // JSON числа unmarshaled как float64
				if !ok {
					return fmt.Errorf("некорректное значение в range")
				}
				rangeValues[i] = int(num)
			}
			valueJSON, _ = json.Marshal(rangeValues)

		case "switcher":
			// Ожидаем bool или устанавливаем дефолтное значение false
			if attrReq.Value != nil {
				boolVal, ok := attrReq.Value.(bool)
				if !ok {
					return fmt.Errorf("некорректный тип value для switcher")
				}
				valueJSON, _ = json.Marshal(boolVal)
			} else {
				boolVal = false
				valueJSON, _ = json.Marshal(false)
			}

		case "text":
			// Ожидаем строку или устанавливаем дефолтное значение ""
			if attrReq.Value != nil {
				str, ok := attrReq.Value.(string)
				if !ok {
					return fmt.Errorf("некорректный тип value для text")
				}
				textVal = str
			} else {
				textVal = ""
			}
			valueJSON, _ = json.Marshal(textVal)

		case "numeric":
			// Ожидаем число или устанавливаем дефолтное значение 0
			if attrReq.Value != nil {
				num, ok := attrReq.Value.(float64) // JSON числа unmarshaled как float64
				if !ok {
					return fmt.Errorf("некорректный тип value для numeric")
				}
				numericVal = int(num)
			} else {
				numericVal = 0
			}
			valueJSON, _ = json.Marshal(numericVal)

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

		// Создаем атрибут категории в базе данных
		err := db.CreateCategoryAttribute(attribute)
		if err != nil {
			log.Printf("Не удалось создать атрибут для пользователя %d: %s. Ошибка: %v", userID, attrReq.Name, err)
			return fmt.Errorf("не удалось создать атрибут %s: %v", attrReq.Name, err)
		}

		// После создания атрибута, добавляем значения в таблицу attribute_value
		switch attrReq.TypeOfOption {
		case "dropdown":
			// Добавление значений в таблицу attribute_value
			for _, value := range dropdownValues {
				err := db.CreateAttributeValue(attribute.ID, value)
				if err != nil {
					return fmt.Errorf("не удалось создать значение атрибута dropdown %s: %v", value, err)
				}
			}

		case "range":
			// Добавление диапазона в attribute_value
			rangeStr := fmt.Sprintf("[%d, %d]", rangeValues[0], rangeValues[1])
			err := db.CreateAttributeValue(attribute.ID, rangeStr)
			if err != nil {
				return fmt.Errorf("не удалось создать значение атрибута range %s: %v", rangeStr, err)
			}

		case "switcher":
			// Добавление значения в attribute_value
			boolStr := fmt.Sprintf("%t", boolVal)
			err := db.CreateAttributeValue(attribute.ID, boolStr)
			if err != nil {
				return fmt.Errorf("не удалось создать значение атрибута switcher %s: %v", boolStr, err)
			}

		case "text":
			// Добавление значения в attribute_value
			err := db.CreateAttributeValue(attribute.ID, textVal)
			if err != nil {
				return fmt.Errorf("не удалось создать значение атрибута text %s: %v", textVal, err)
			}

		case "numeric":
			// Добавление значения в attribute_value
			numericStr := fmt.Sprintf("%d", numericVal)
			err := db.CreateAttributeValue(attribute.ID, numericStr)
			if err != nil {
				return fmt.Errorf("не удалось создать значение атрибута numeric %s: %v", numericStr, err)
			}
		}
	}

	return nil
}

// GetCategoryByID возвращает категорию по её ID
func (s *CategoryService) GetCategoryByID(categoryID int) (*models.Category, error) {
	return db.GetCategoryByID(categoryID)
}

func StringPtr(s string) *string {
	return &s
}

// GetCategoryAttributes возвращает атрибуты для заданной категории
func (s *CategoryService) GetCategoryAttributes(categoryID int) ([]models.CategoryAttribute, error) {
	// Получаем атрибуты категории из базы данных
	attributes, err := db.GetCategoryAttributes(categoryID)
	if err != nil {
		return nil, err
	}

	// Преобразуем []models.Attribute в []models.CategoryAttribute
	categoryAttributes := make([]models.CategoryAttribute, len(attributes))
	for i, attr := range attributes {
		categoryAttributes[i] = models.CategoryAttribute{
			ID:           attr.ID,    // Используем ID атрибута
			CategoryID:   categoryID, // Используем переданный в метод categoryID
			Name:         attr.Name,
			Description:  StringPtr(attr.Description),  // Преобразуем строку в *string, если требуется
			TypeOfOption: StringPtr(attr.TypeOfOption), // Преобразуем строку в *string, если требуется
			Value:        attr.Value,
		}
	}

	return categoryAttributes, nil
}

func (s *CategoryService) GetCategoryAttributesByCategoryID(userID int, categoryID int) ([]models.CategoryAttributeResponse, error) {
	// Логирование запроса
	log.Printf("User ID %d запрашивает атрибуты для категории ID %d", userID, categoryID)

	// Получаем атрибуты из базы данных
	attributes, err := db.GetCategoryAttributesByCategoryID(categoryID)
	if err != nil {
		log.Printf("Ошибка при получении атрибутов для категории %d: %v", categoryID, err)
		return nil, fmt.Errorf("ошибка при получении атрибутов: %v", err)
	}

	// Преобразуем атрибуты в DTO
	var response []models.CategoryAttributeResponse
	for _, attr := range attributes {
		var value interface{}

		// Десериализуем Value в зависимости от TypeOfOption
		switch *attr.TypeOfOption {
		case "dropdown":
			var dropdownValues []string
			if err := json.Unmarshal(attr.Value, &dropdownValues); err != nil {
				log.Printf("Ошибка десериализации value для dropdown: %v", err)
				return nil, fmt.Errorf("некорректное значение атрибута %s", attr.Name)
			}
			value = dropdownValues

		case "range":
			var rangeValues []int
			if err := json.Unmarshal(attr.Value, &rangeValues); err != nil {
				log.Printf("Ошибка десериализации value для range: %v", err)
				return nil, fmt.Errorf("некорректное значение атрибута %s", attr.Name)
			}
			value = rangeValues

		case "switcher":
			var switcherValue bool
			if err := json.Unmarshal(attr.Value, &switcherValue); err != nil {
				log.Printf("Ошибка десериализации value для switcher: %v", err)
				return nil, fmt.Errorf("некорректное значение атрибута %s", attr.Name)
			}
			value = switcherValue

		case "text":
			var textValue string
			if err := json.Unmarshal(attr.Value, &textValue); err != nil {
				log.Printf("Ошибка десериализации value для text: %v", err)
				return nil, fmt.Errorf("некорректное значение атрибута %s", attr.Name)
			}
			value = textValue

		case "numeric":
			var numericValue int
			if err := json.Unmarshal(attr.Value, &numericValue); err != nil {
				log.Printf("Ошибка десериализации value для numeric: %v", err)
				return nil, fmt.Errorf("некорректное значение атрибута %s", attr.Name)
			}
			value = numericValue

		default:
			log.Printf("Неизвестный type_of_option: %s", *attr.TypeOfOption)
			return nil, fmt.Errorf("неизвестный тип опции атрибута %s", attr.Name)
		}

		response = append(response, models.CategoryAttributeResponse{
			ID:           attr.ID,
			Name:         attr.Name,
			Description:  attr.Description,
			TypeOfOption: attr.TypeOfOption,
			Value:        value,
		})
	}

	return response, nil
}

func (s *CategoryService) GetCategoryAttributesAndValues(categoryID int) (map[string]map[string]int, error) {
	// Получаем атрибуты для категории
	attributes, err := db.GetCategoryAttributes(categoryID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении атрибутов категории: %v", err)
	}

	// Мапа для хранения атрибутов и их значений
	attributeMap := make(map[string]map[string]int)

	for _, attr := range attributes {
		// Получаем все значения для текущего атрибута
		values, err := db.GetAttributeValues(attr.ID)
		if err != nil {
			return nil, fmt.Errorf("ошибка при получении значений атрибута: %v", err)
		}

		// Создаем вложенную мапу для хранения значений атрибута
		valueMap := make(map[string]int)
		for _, value := range values {
			valueMap[value.Value] = value.ID
		}

		// Сохраняем атрибут и его значения в мапе
		attributeMap[attr.Name] = valueMap
	}

	return attributeMap, nil
}
