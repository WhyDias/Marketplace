package db

import (
	"database/sql"
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

func CreateCategoryAttribute(categoryID int, name string) error {
	query := `INSERT INTO category_attributes (category_id, name) VALUES ($1, $2)`
	_, err := DB.Exec(query, categoryID, name)
	if err != nil {
		return fmt.Errorf("Не удалось создать атрибут категории: %v", err)
	}
	return nil
}

func GetCategoryAttributes(categoryID int) ([]models.CategoryAttribute, error) {
	query := `SELECT id, name FROM category_attributes WHERE category_id = $1`

	rows, err := DB.Query(query, categoryID)
	if err != nil {
		return nil, fmt.Errorf("Не удалось получить атрибуты категории: %v", err)
	}
	defer rows.Close()

	var attributes []models.CategoryAttribute
	for rows.Next() {
		var attr models.CategoryAttribute
		if err := rows.Scan(&attr.ID, &attr.Name); err != nil {
			return nil, fmt.Errorf("Ошибка при чтении атрибута: %v", err)
		}
		attributes = append(attributes, attr)
	}

	return attributes, nil
}

func GetCategoryByID(categoryID int) (*models.Category, error) {
	query := `SELECT id, name, path, image_url FROM categories WHERE id = $1`
	category := &models.Category{}
	err := DB.QueryRow(query, categoryID).Scan(&category.ID, &category.Name, &category.Path, &category.ImageURL)
	if err != nil {
		return nil, fmt.Errorf("Не удалось получить категорию: %v", err)
	}
	return category, nil
}
