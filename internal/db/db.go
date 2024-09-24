package db

import (
	"database/sql"
	"fmt"
	"github.com/WhyDias/Marketplace/internal/models"
	_ "github.com/lib/pq" // Импорт драйвера PostgreSQL
	"gopkg.in/yaml.v2"
	"io/ioutil"
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

func CreateSupplier(supplier *models.Supplier) error {
	query := `INSERT INTO suppliers (name, phone_number, market_id, place_id, row_id, categories, created_at, updated_at, is_verified) 
              VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`

	err := DB.QueryRow(query, supplier.Name, supplier.PhoneNumber, supplier.MarketID, supplier.PlaceID, supplier.Categories, supplier.CreatedAt, supplier.UpdatedAt).Scan(&supplier.ID)
	if err != nil {
		return fmt.Errorf("failed to create supplier: %v", err)
	}

	return nil
}

// GetSupplierInfo получает информацию о поставщике по номеру телефона
func GetSupplierInfo(phoneNumber string) (*models.Supplier, error) {
	supplier := &models.Supplier{}

	query := `SELECT name, phone_number, market_id, place_id, row_id, categories, created_at, updated_at 
              FROM suppliers WHERE phone_number = $1`

	err := DB.QueryRow(query, phoneNumber).Scan(
		&supplier.ID,
		&supplier.Name,
		&supplier.PhoneNumber,
		&supplier.MarketID,
		&supplier.PlaceID,
		&supplier.Categories,
		&supplier.CreatedAt,
		&supplier.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("supplier not found")
		}
		return nil, fmt.Errorf("error fetching supplier: %v", err)
	}

	return supplier, nil
}

// GetAllSuppliers получает список всех поставщиков
func GetAllSuppliers() ([]models.Supplier, error) {
	query := `SELECT id, name, phone_number, market_id, place_id, row_id, categories, created_at, updated_at FROM suppliers`

	rows, err := DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error fetching suppliers: %v", err)
	}
	defer rows.Close()

	var suppliers []models.Supplier
	for rows.Next() {
		var supplier models.Supplier
		if err := rows.Scan(&supplier.ID, &supplier.Name, &supplier.PhoneNumber, &supplier.MarketID, &supplier.PlaceID, &supplier.Categories, &supplier.CreatedAt, &supplier.UpdatedAt); err != nil {
			return nil, fmt.Errorf("error scanning supplier: %v", err)
		}
		suppliers = append(suppliers, supplier)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %v", err)
	}

	return suppliers, nil
}

// GetLatestVerificationCode получает последний созданный код для данного номера телефона

// MarkPhoneNumberAsVerified обновляет поле is_verified для поставщика
func MarkPhoneNumberAsVerified(phoneNumber string) error {
	query := `UPDATE suppliers SET is_verified = TRUE WHERE phone_number = $1`
	result, err := DB.Exec(query, phoneNumber)
	if err != nil {
		return fmt.Errorf("failed to mark phone number as verified: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error fetching rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("supplier with phone number %s not found", phoneNumber)
	}

	return nil
}

func CreateUser(user *models.User) error {
	query := `INSERT INTO users (username, password_hash, role, created_at, updated_at)
              VALUES ($1, $2, $3, $4, $5) RETURNING id`

	err := DB.QueryRow(query, user.Username, user.PasswordHash, user.Role, user.CreatedAt, user.UpdatedAt).Scan(&user.ID)
	if err != nil {
		return fmt.Errorf("Не удалось создать пользователя: %v", err)
	}

	return nil
}

func GetUserByUsername(username string) (*models.User, error) {
	user := &models.User{}

	query := `SELECT id, username, password_hash, role, created_at, updated_at FROM users WHERE username = $1`
	err := DB.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("Не удалось получить пользователя по имени: %v", err)
	}

	return user, nil
}
