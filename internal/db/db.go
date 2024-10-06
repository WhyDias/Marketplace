package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/WhyDias/Marketplace/internal/models"
	"github.com/lib/pq"
	_ "github.com/lib/pq" // Импорт драйвера PostgreSQL
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type Config struct {
	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DBName   string `yaml:"dbname"`
		SSLMode  string `yaml:"sslmode"`
	} `yaml:"database"`
	JWT struct {
		Secret     string `yaml:"secret"`
		Expiration string `yaml:"expiration"`
	} `yaml:"jwt"`
}

var DB *sql.DB

func InitDB(configPath string) error {
	config := Config{}
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("error reading config file: %v", err)
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return fmt.Errorf("error unmarshalling config: %v", err)
	}

	// Формирование строки подключения
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Database.Host,
		config.Database.Port,
		config.Database.User,
		config.Database.Password,
		config.Database.DBName,
		config.Database.SSLMode,
	)

	// Открытие подключения к базе данных
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("error opening database: %v", err)
	}

	// Проверка соединения
	err = db.Ping()
	if err != nil {
		return fmt.Errorf("error connecting to database: %v", err)
	}

	DB = db
	fmt.Println("Successfully connected to database")
	return nil
}

// GetUserByUsername получает пользователя из базы данных по имени пользователя

func CreateUser(user *models.User, roleIDs []int) error {
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("Не удалось начать транзакцию: %v", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// Вставляем пользователя
	query := `INSERT INTO users (username, password_hash, created_at, updated_at)
              VALUES ($1, $2, $3, $4) RETURNING id`
	err = tx.QueryRow(query, user.Username, user.PasswordHash, user.CreatedAt, user.UpdatedAt).Scan(&user.ID)
	if err != nil {
		return fmt.Errorf("Не удалось создать пользователя: %v", err)
	}

	// Вставляем роли в таблицу user_roles
	for _, roleID := range roleIDs {
		_, err = tx.Exec(`INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)`, user.ID, roleID)
		if err != nil {
			return fmt.Errorf("Не удалось назначить роль пользователю: %v", err)
		}
	}

	return nil
}

// GetUserByUsername получает пользователя из базы данных по имени пользователя
func GetUserByUsername(username string) (*models.User, error) {
	user := &models.User{}

	query := `SELECT id, username, password_hash, created_at, updated_at FROM users WHERE username = $1`
	err := DB.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("GetUserByUsername: User %s not found", username)
			return nil, nil
		}
		return nil, fmt.Errorf("could not get user by username: %v", err)
	}

	// Get roles
	rolesQuery := `SELECT r.id, r.name FROM roles r INNER JOIN user_roles ur ON r.id = ur.role_id WHERE ur.user_id = $1`
	rows, err := DB.Query(rolesQuery, user.ID)
	if err != nil {
		return nil, fmt.Errorf("could not get user roles: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var role models.Role
		if err := rows.Scan(&role.ID, &role.Name); err != nil {
			return nil, fmt.Errorf("could not scan role: %v", err)
		}
		user.Roles = append(user.Roles, role)
	}

	return user, nil
}

func GetCategoryByPath(path string) (*models.Category, error) {
	category := &models.Category{}

	query := `SELECT id, name, path, image_url FROM categories WHERE path = $1 LIMIT 1`
	err := DB.QueryRow(query, path).Scan(
		&category.ID,
		&category.Name,
		&category.Path,
		&category.ImageURL,
	)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			log.Printf("GetCategoryByPath: категория не найдена для path %s", path)
			return nil, fmt.Errorf("категория не найдена для path %s", path)
		}
		log.Printf("GetCategoryByPath: ошибка при выполнении запроса для path %s: %v", path, err)
		return nil, fmt.Errorf("ошибка при выполнении запроса: %v", err)
	}

	log.Printf("GetCategoryByPath: получена категория id=%d для path %s", category.ID, path)
	return category, nil
}

func CreateCategory(category *models.Category) error {
	query := `INSERT INTO categories (name, path, image_url)
	          VALUES ($1, $2, $3) RETURNING id`
	err := DB.QueryRow(query, category.Name, category.Path, category.ImageURL).Scan(&category.ID)
	if err != nil {
		log.Printf("CreateCategory: ошибка при выполнении запроса: %v", err)
		return fmt.Errorf("ошибка при создании категории: %v", err)
	}

	log.Printf("CreateCategory: создана категория id=%d, name=%s", category.ID, category.Name)
	return nil
}

func GetRoleIDsByNames(roleNames []string) ([]int, error) {
	var roleIDs []int
	for _, roleName := range roleNames {
		var roleID int
		err := DB.QueryRow(`SELECT id FROM roles WHERE name = $1`, roleName).Scan(&roleID)
		if err != nil {
			return nil, fmt.Errorf("Роль %s не найдена: %v", roleName, err)
		}
		roleIDs = append(roleIDs, roleID)
	}
	return roleIDs, nil
}

func GetPhoneNumberByUserID(userID int) (string, error) {
	var phoneNumber string
	query := `SELECT username FROM users WHERE id = $1`
	err := DB.QueryRow(query, userID).Scan(&phoneNumber)
	if err != nil {
		return "", fmt.Errorf("Не удалось получить номер телефона пользователя: %v", err)
	}
	return phoneNumber, nil
}

func GetRolesByUserID(userID int) ([]models.Role, error) {
	query := `
        SELECT r.id, r.name
        FROM roles r
        INNER JOIN user_roles ur ON r.id = ur.role_id
        WHERE ur.user_id = $1
    `
	rows, err := DB.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при получении ролей по userID: %v", err)
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var role models.Role
		if err := rows.Scan(&role.ID, &role.Name); err != nil {
			return nil, fmt.Errorf("Ошибка при сканировании роли: %v", err)
		}
		roles = append(roles, role)
	}

	return roles, nil
}

func GetSubcategoriesByPath(path string) ([]models.Category, error) {
	query := `
        SELECT id, name, path, image_url
        FROM categories
        WHERE path <@ $1
    `

	rows, err := DB.Query(query, path)
	if err != nil {
		log.Printf("GetSubcategoriesByPath: ошибка при выполнении запроса: %v", err)
		return nil, fmt.Errorf("ошибка при выполнении запроса: %v", err)
	}
	defer rows.Close()

	var categories []models.Category

	for rows.Next() {
		var category models.Category
		if err := rows.Scan(&category.ID, &category.Name, &category.Path, &category.ImageURL); err != nil {
			log.Printf("GetSubcategoriesByPath: ошибка при сканировании строки: %v", err)
			return nil, fmt.Errorf("ошибка при сканировании строки: %v", err)
		}
		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		log.Printf("GetSubcategoriesByPath: ошибка при итерации по строкам: %v", err)
		return nil, fmt.Errorf("ошибка при итерации по строкам: %v", err)
	}

	return categories, nil
}

func CreateAttributeValueImage(image *models.AttributeValueImage) error {
	query := `
        INSERT INTO attribute_value_image (attribute_value_id, image_url)
        VALUES ($1, $2)
        RETURNING id
    `
	err := DB.QueryRow(query, image.AttributeValueID, pq.Array(image.ImageURLs)).Scan(&image.ID)
	if err != nil {
		return fmt.Errorf("Не удалось создать запись в attribute_value_image: %v", err)
	}
	return nil
}

func GetAttributeValueImageByAttributeValueID(attributeValueID int) (*models.AttributeValueImage, error) {
	query := `
        SELECT id, attribute_value_id, image_url
        FROM attribute_value_image
        WHERE attribute_value_id = $1
    `
	var image models.AttributeValueImage
	err := DB.QueryRow(query, attributeValueID).Scan(&image.ID, &image.AttributeValueID, pq.Array(&image.ImageURLs))
	if err != nil {
		return nil, fmt.Errorf("Не удалось получить AttributeValueImage: %v", err)
	}
	return &image, nil
}

func GetAllCategories() ([]models.Category, error) {
	query := `SELECT id, name, path, image_url, parent_id FROM categories ORDER BY path`

	rows, err := DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка при выполнении запроса: %v", err)
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var category models.Category
		if err := rows.Scan(&category.ID, &category.Name, &category.Path, &category.ImageURL, &category.ParentID); err != nil {
			return nil, fmt.Errorf("ошибка при сканировании строки: %v", err)
		}
		categories = append(categories, category)
	}

	return categories, nil
}

func CreateCategoryAttribute(attribute *models.CategoryAttribute) (int, error) {
	query := `
		INSERT INTO category_attributes (category_id, name, description, type_of_option, value)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	var createdAttributeID int
	err := DB.QueryRow(query, attribute.CategoryID, attribute.Name, attribute.Description, attribute.TypeOfOption, attribute.Value).Scan(&createdAttributeID)
	if err != nil {
		return 0, err
	}
	return createdAttributeID, nil
}

func GetCategoryByID(categoryID int) (*models.Category, error) {
	category := &models.Category{}

	query := `SELECT id, name, path, image_url, parent_id FROM categories WHERE id = $1`
	err := DB.QueryRow(query, categoryID).Scan(
		&category.ID,
		&category.Name,
		&category.Path,
		&category.ImageURL,
		&category.ParentID, // Теперь ParentID это sql.NullInt64
	)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("GetCategoryByID: категория с id=%d не найдена", categoryID)
			return nil, nil
		}
		log.Printf("GetCategoryByID: ошибка при выполнении запроса для id=%d: %v", categoryID, err)
		return nil, fmt.Errorf("ошибка при выполнении запроса: %v", err)
	}

	return category, nil
}

// GetCategoryAttributes получает атрибуты для заданной категории
func AddCategoryAttribute(attr models.CategoryAttribute) error {
	query := `
        INSERT INTO category_attributes (category_id, name, description, type_of_option, value)
        VALUES ($1, $2, $3, $4, $5)
    `
	_, err := DB.Exec(query, attr.CategoryID, attr.Name, attr.Description, attr.TypeOfOption, attr.Value)
	if err != nil {
		log.Printf("AddCategoryAttribute: ошибка при выполнении запроса: %v", err)
		return fmt.Errorf("ошибка при выполнении запроса: %v", err)
	}
	return nil
}

func CreateProduct(product *models.Product) error {
	query := `
		INSERT INTO product (name, category_id, market_id, status_id, supplier_id, description, price, stock)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`
	err := DB.QueryRow(query, product.Name, product.CategoryID, product.MarketID, product.StatusID, product.SupplierID, product.Description, product.Price, product.Stock).Scan(&product.ID)
	if err != nil {
		return err
	}
	return nil
}

func GetAttributeIDByName(categoryID int, name string) (int, error) {
	query := `
        SELECT id FROM category_attributes
        WHERE category_id = $1 AND name = $2
    `
	var id int
	err := DB.QueryRow(query, categoryID, name).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("атрибут с именем %s не найден", name)
		}
		log.Printf("GetAttributeIDByName: ошибка при выполнении запроса: %v", err)
		return 0, err
	}
	return id, nil
}

func GetOrCreateAttributeValue(attributeID int, value string) (int, error) {
	// Попробуем найти существующее значение
	query := `
        SELECT id FROM attribute_value
        WHERE attribute_id = $1 AND value = $2
    `
	var id int
	err := DB.QueryRow(query, attributeID, value).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			// Создаем новое значение
			insertQuery := `
                INSERT INTO attribute_value (attribute_id, value)
                VALUES ($1, $2)
                RETURNING id
            `
			err = DB.QueryRow(insertQuery, attributeID, value).Scan(&id)
			if err != nil {
				log.Printf("GetOrCreateAttributeValue: ошибка при создании значения атрибута: %v", err)
				return 0, fmt.Errorf("ошибка при создании значения атрибута: %v", err)
			}
			return id, nil
		}
		log.Printf("GetOrCreateAttributeValue: ошибка при выполнении запроса: %v", err)
		return 0, err
	}
	return id, nil
}

func GetCategoryAttributesByCategoryID(categoryID int) ([]models.CategoryAttribute, error) {
	query := `
		SELECT id, category_id, name, description, type_of_option, value
		FROM category_attributes
		WHERE category_id = $1
	`

	rows, err := DB.Query(query, categoryID)
	if err != nil {
		log.Printf("GetCategoryAttributesByCategoryID: ошибка при выполнении запроса: %v", err)
		return nil, fmt.Errorf("ошибка при выполнении запроса: %v", err)
	}
	defer rows.Close()

	var attributes []models.CategoryAttribute
	for rows.Next() {
		var attr models.CategoryAttribute
		err := rows.Scan(&attr.ID, &attr.CategoryID, &attr.Name, &attr.Description, &attr.TypeOfOption, &attr.Value)
		if err != nil {
			log.Printf("GetCategoryAttributesByCategoryID: ошибка при сканировании строки: %v", err)
			return nil, fmt.Errorf("ошибка при сканировании строки: %v", err)
		}
		attributes = append(attributes, attr)
	}

	if err = rows.Err(); err != nil {
		log.Printf("GetCategoryAttributesByCategoryID: ошибка при итерации по строкам: %v", err)
		return nil, fmt.Errorf("ошибка при итерации по строкам: %v", err)
	}

	return attributes, nil
}

func CreateProductVariationTx(tx *sql.Tx, variation *models.ProductVariation) error {
	query := `INSERT INTO product_variation (product_id, sku) VALUES ($1, $2) RETURNING id`
	err := tx.QueryRow(query, variation.ProductID, variation.SKU).Scan(&variation.ID)
	if err != nil {
		return fmt.Errorf("Ошибка при создании вариации продукта: %v", err)
	}
	return nil
}

func GetCategoryAttributes(categoryID int) ([]models.Attribute, error) {
	var attributes []models.Attribute
	query := `
        SELECT id, name, description, type_of_option, value
        FROM category_attributes
        WHERE category_id = $1
    `
	rows, err := DB.Query(query, categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var attribute models.Attribute
		var description sql.NullString

		// Сканирование данных
		if err := rows.Scan(&attribute.ID, &attribute.Name, &description, &attribute.TypeOfOption, &attribute.Value); err != nil {
			return nil, err
		}

		// Обработка sql.NullString для Description
		if description.Valid {
			attribute.Description = description.String
		}

		attributes = append(attributes, attribute)
	}
	return attributes, nil
}

func CreateProductAttributeValue(attributeValue *models.ProductAttributeValue) error {
	query := `
        INSERT INTO product_attribute_values (product_id, attribute_value_id)
        VALUES ($1, $2)
        RETURNING product_id, attribute_value_id`

	_, err := DB.Exec(query, attributeValue.ProductID, attributeValue.AttributeValueID)
	if err != nil {
		return fmt.Errorf("ошибка при создании значения атрибута продукта: %v", err)
	}
	return nil
}

func CreateProductImage(image *models.ProductImage) error {
	query := `
        INSERT INTO product_images (product_id, image_urls, image_path)
        VALUES ($1, $2, $3)
    `
	_, err := DB.Exec(query, image.ProductID, image.ImageURLs, image.ImagePath)
	if err != nil {
		return fmt.Errorf("ошибка при создании изображения продукта: %v", err)
	}
	return nil
}

func GetSupplierByUserID(userID int) (*models.Supplier, error) {
	var supplier models.Supplier
	query := `
        SELECT id, name, market_id
        FROM supplier
        WHERE user_id = $1
    `
	err := DB.QueryRow(query, userID).Scan(&supplier.ID, &supplier.Name, &supplier.MarketID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("поставщик с user_id %d не найден", userID)
		}
		return nil, fmt.Errorf("ошибка при получении поставщика: %v", err)
	}
	return &supplier, nil
}

func CreateProductVariation(variation *models.ProductVariation) error {
	query := `
        INSERT INTO product_variation (product_id, sku)
        VALUES ($1, $2)
        RETURNING id
    `
	err := DB.QueryRow(query, variation.ProductID, variation.SKU).Scan(&variation.ID)
	if err != nil {
		return fmt.Errorf("ошибка при создании вариации продукта: %v", err)
	}
	return nil
}

func CreateVariationAttributeValue(variationAttributeValue *models.VariationAttributeValue) error {
	query := `
        INSERT INTO variation_attribute_values (product_variation_id, attribute_value_id)
        VALUES ($1, $2)
    `
	_, err := DB.Exec(query, variationAttributeValue.ProductVariationID, variationAttributeValue.AttributeValueID)
	return err
}

func CreateProductVariationImage(image *models.ProductVariationImage) error {
	query := `
		INSERT INTO product_variation_images (product_variation_id, image_urls, image_path)
		VALUES ($1, $2, $3)
		RETURNING id`

	err := DB.QueryRow(query, image.ProductVariationID, image.ImageURLs, image.ImagePath).Scan(&image.ID)
	if err != nil {
		return fmt.Errorf("ошибка при создании изображения вариации продукта: %v", err)
	}
	return nil
}

func CreateAttributeValue(attributeID int, value json.RawMessage) (int, error) {
	var createdAttributeValueID int
	query := `
        INSERT INTO attribute_value (attribute_id, value_json)
        VALUES ($1, $2)
        RETURNING id
    `
	err := DB.QueryRow(query, attributeID, value).Scan(&createdAttributeValueID)
	if err != nil {
		return 0, fmt.Errorf("Ошибка при создании значения атрибута: %v", err)
	}
	return createdAttributeValueID, nil
}

// GetAttributeValueID возвращает ID значения атрибута по его значению
func GetAttributeValueID(attributeID int, value string) (int, error) {
	query := `SELECT id FROM attribute_value WHERE attribute_id = $1 AND value = $2`
	var id int
	err := DB.QueryRow(query, attributeID, value).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("значение атрибута '%s' не найдено для attribute_id %d", value, attributeID)
		}
		return 0, err
	}
	return id, nil
}

func GetAttributeValues(attributeID int) ([]models.AttributeValue, error) {
	var attributeValues []models.AttributeValue

	query := `
        SELECT id, value
        FROM attribute_value
        WHERE attribute_id = $1
    `
	rows, err := DB.Query(query, attributeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var attributeValue models.AttributeValue
		if err := rows.Scan(&attributeValue.ID, &attributeValue.Value); err != nil {
			return nil, err
		}
		attributeValues = append(attributeValues, attributeValue)
	}
	return attributeValues, nil
}

func UpdateAttributeValue(attributeID int, value json.RawMessage) error {
	query := `
        UPDATE attribute_value
        SET value_json = $2
        WHERE attribute_id = $1
    `
	_, err := DB.Exec(query, attributeID, value)
	return err
}

func GetAttributeIDByNameAndCategory(attributeName string, categoryID int) (int, error) {
	var attributeID int

	query := `
        SELECT id
        FROM attributes
        WHERE name = $1 AND category_id = $2
    `

	err := DB.QueryRow(query, attributeName, categoryID).Scan(&attributeID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("атрибут с именем '%s' для категории с ID %d не найден", attributeName, categoryID)
		}
		return 0, fmt.Errorf("ошибка при поиске атрибута '%s': %v", attributeName, err)
	}

	return attributeID, nil
}

func CreateOrUpdateAttributeValue(attributeID int, value json.RawMessage) (int, error) {
	var attributeValueID int

	// Проверяем, существует ли значение атрибута
	query := `
		SELECT id
		FROM attribute_value
		WHERE attribute_id = $1 AND value_json = $2
	`
	err := DB.QueryRow(query, attributeID, value).Scan(&attributeValueID)

	if err == sql.ErrNoRows {
		// Значение не существует, создаем новое
		insertQuery := `
			INSERT INTO attribute_value (attribute_id, value_json)
			VALUES ($1, $2)
			RETURNING id
		`
		err = DB.QueryRow(insertQuery, attributeID, value).Scan(&attributeValueID)
		if err != nil {
			log.Printf("CreateOrUpdateAttributeValue: Ошибка при создании значения атрибута с attribute_id %d и value %s: %v", attributeID, string(value), err)
			return 0, fmt.Errorf("ошибка при создании значения атрибута: %v", err)
		}
	} else if err != nil {
		// Ошибка при выполнении запроса
		log.Printf("CreateOrUpdateAttributeValue: Ошибка при проверке значения атрибута с attribute_id %d и value %s: %v", attributeID, string(value), err)
		return 0, fmt.Errorf("ошибка при проверке значения атрибута: %v", err)
	}

	// Возвращаем существующий или созданный ID значения атрибута
	return attributeValueID, nil
}

func GetRootCategories() ([]models.Category, error) {
	query := `
        SELECT id, name, path, image_url, parent_id
        FROM categories
        WHERE nlevel(path) = 1
    `
	rows, err := DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %v", err)
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var category models.Category
		if err := rows.Scan(&category.ID, &category.Name, &category.Path, &category.ImageURL, &category.ParentID); err != nil {
			return nil, fmt.Errorf("ошибка при чтении данных категории: %v", err)
		}
		categories = append(categories, category)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по строкам результата: %v", err)
	}

	return categories, nil
}
