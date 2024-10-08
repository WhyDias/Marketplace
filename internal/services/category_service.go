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
		var err error
		var stringValues []string
		var rangeValues []int

		// Преобразование значений в JSON в зависимости от типа атрибута
		switch attrReq.TypeOfOption {
		case "dropdown":
			values, ok := attrReq.Value.([]interface{})
			if !ok {
				return fmt.Errorf("некорректный тип value для dropdown")
			}
			stringValues = make([]string, len(values))
			for i, v := range values {
				str, ok := v.(string)
				if !ok {
					return fmt.Errorf("некорректное значение в dropdown")
				}
				stringValues[i] = str
			}
			valueJSON, _ = json.Marshal(stringValues)

		case "range":
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
			var boolVal bool
			if attrReq.Value != nil {
				value, ok := attrReq.Value.(bool)
				if !ok {
					return fmt.Errorf("некорректный тип value для switcher")
				}
				boolVal = value
			} else {
				boolVal = false
			}
			valueJSON, _ = json.Marshal(boolVal)

		case "text":
			var textVal string
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
			var numericVal int
			if attrReq.Value != nil {
				num, ok := attrReq.Value.(float64)
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

		// Проверка на наличие атрибута с таким именем для данной категории
		existingCategoryAttribute, err := db.GetCategoryAttributeByName(req.CategoryID, attrReq.Name)
		if err != nil {
			log.Printf("Ошибка при проверке существования атрибута: %v", err)
			return fmt.Errorf("не удалось проверить существование атрибута: %v", err)
		}

		if existingCategoryAttribute != nil {
			// Атрибут с таким именем существует - обновляем его
			existingCategoryAttribute.Description = &attrReq.Description
			existingCategoryAttribute.TypeOfOption = &attrReq.TypeOfOption
			existingCategoryAttribute.Value = valueJSON

			err = db.UpdateCategoryAttribute(existingCategoryAttribute)
			if err != nil {
				log.Printf("Ошибка при обновлении атрибута категории: %v", err)
				return fmt.Errorf("не удалось обновить атрибут категории: %v", err)
			}
		} else {
			// Атрибут с таким именем не найден - создаём новый
			categoryAttribute := models.CategoryAttribute{
				CategoryID:   req.CategoryID,
				Name:         attrReq.Name,
				Description:  &attrReq.Description,
				TypeOfOption: &attrReq.TypeOfOption,
				Value:        valueJSON,
			}

			_, err := db.CreateCategoryAttribute(&categoryAttribute)
			if err != nil {
				log.Printf("Ошибка при создании нового атрибута категории: %v", err)
				return fmt.Errorf("не удалось создать новый атрибут категории: %v", err)
			}
		}

		// Теперь добавляем или обновляем запись в таблице attributes
		existingAttribute, err := db.GetAttributeByNameAndCategoryID(attrReq.Name, req.CategoryID)
		if err != nil {
			log.Printf("Ошибка при проверке существования атрибута в attributes: %v", err)
			return fmt.Errorf("не удалось проверить существование атрибута в attributes: %v", err)
		}

		if existingAttribute != nil {
			// Обновляем атрибут
			existingAttribute.Description = attrReq.Description
			existingAttribute.TypeOfOption = attrReq.TypeOfOption
			existingAttribute.Value = valueJSON
			existingAttribute.IsLinked = attrReq.IsLinked

			err = db.UpdateAttribute(existingAttribute)
			if err != nil {
				log.Printf("Ошибка при обновлении атрибута в таблице attributes: %v", err)
				return fmt.Errorf("не удалось обновить атрибут в таблице attributes: %v", err)
			}
		} else {
			// Создаём новый атрибут
			attribute := models.Attribute{
				Name:         attrReq.Name,
				CategoryID:   req.CategoryID,
				Description:  attrReq.Description,
				TypeOfOption: attrReq.TypeOfOption,
				IsLinked:     attrReq.IsLinked,
				Value:        valueJSON,
			}

			_, err := db.CreateAttribute(&attribute)
			if err != nil {
				log.Printf("Ошибка при создании атрибута в таблице attributes: %v", err)
				return fmt.Errorf("не удалось создать атрибут в таблице attributes: %v", err)
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
	attributes, err := db.GetCategoryAttributes(categoryID)
	if err != nil {
		return nil, err
	}

	// Создаем карту для хранения атрибутов и их значений
	attributeMap := make(map[string]map[string]int)

	for _, attribute := range attributes {
		// Загружаем все значения для текущего атрибута
		attributeValues, err := db.GetAttributeValues(attribute.ID)
		if err != nil {
			return nil, fmt.Errorf("не удалось загрузить значения для атрибута '%s': %v", attribute.Name, err)
		}

		// Создаем мапу значений атрибута (name -> id)
		valueMap := make(map[string]int)
		for _, value := range attributeValues {
			valueMap[value.Value] = value.ID
		}

		// Добавляем атрибут и его значения в attributeMap
		attributeMap[attribute.Name] = valueMap
	}

	return attributeMap, nil
}

func (s *CategoryService) GetRootCategories() ([]models.Category, error) {
	rootCategories, err := db.GetRootCategories()
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении корневых категорий: %v", err)
	}
	return rootCategories, nil
}

func (s *CategoryService) DeleteCategoryAttributes(categoryID int) error {
	log.Printf("Удаление атрибутов для категории ID %d", categoryID)

	// Удаление атрибутов категории из таблицы category_attributes
	if err := db.DeleteCategoryAttributes(categoryID); err != nil {
		log.Printf("DeleteCategoryAttributes: ошибка при удалении атрибутов из category_attributes для категории %d: %v", categoryID, err)
		return err
	}

	//// Удаление атрибутов из таблицы attributes
	//if err := db.DeleteAttributesByCategoryID(categoryID); err != nil {
	//	log.Printf("DeleteAttributesByCategoryID: ошибка при удалении атрибутов из attributes для категории %d: %v", categoryID, err)
	//	return err
	//}

	return nil
}

// GetCategoryByPath возвращает категорию по её пути (path)
func (s *CategoryService) GetCategoryByPath(path string) (*models.Category, error) {
	// Обращаемся к базе данных для получения категории по path
	category, err := db.GetCategoryByPath(path)
	if err != nil {
		log.Printf("GetCategoryByPath: ошибка при получении категории по path %s: %v", path, err)
		return nil, fmt.Errorf("не удалось найти категорию по path %s", path)
	}
	return category, nil
}

func (s *CategoryService) GetAttributesByCategoryAndIsLinked(categoryID int, isLinked bool) ([]models.Attribute, error) {
	attributes, err := db.GetAttributesByCategoryAndIsLinked(categoryID, isLinked)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить атрибуты: %v", err)
	}
	return attributes, nil
}
