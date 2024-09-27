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

func CreateUser(user *models.User) error {
	query := `INSERT INTO users (username, password_hash, role, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5) RETURNING id`

	err := DB.QueryRow(query, user.Username, user.PasswordHash, pq.Array(user.Role), user.CreatedAt, user.UpdatedAt).Scan(&user.ID)
	if err != nil {
		return fmt.Errorf("не удалось создать пользователя: %v", err)
	}

	return nil
}

// GetUserByUsername получает пользователя из базы данных по имени пользователя
func GetUserByUsername(username string) (*models.User, error) {
	user := &models.User{}

	query := `SELECT id, username, email, password_hash, role, created_at, updated_at FROM users WHERE username = $1`
	err := DB.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, pq.Array(&user.Role), &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("не удалось получить пользователя по имени: %v", err)
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
