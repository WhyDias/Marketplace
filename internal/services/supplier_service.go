// internal/services/supplier_service.go

package services

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"log"
	"time"

	"github.com/WhyDias/Marketplace/internal/db"
	"github.com/WhyDias/Marketplace/internal/models"
	"github.com/WhyDias/Marketplace/internal/utils"
)

// SupplierService структура сервиса поставщиков
type SupplierService struct{}

// NewSupplierService конструктор сервиса поставщиков
func NewSupplierService() *SupplierService {
	return &SupplierService{}
}

// CreateSupplier создает нового поставщика
func (s *SupplierService) CreateSupplier(supplier *models.Supplier) error {
	query := `INSERT INTO supplier (phone_number, is_verified, created_at, updated_at)
              VALUES ($1, $2, $3, $4) RETURNING id`
	err := db.DB.QueryRow(query, supplier.PhoneNumber, supplier.IsVerified, supplier.CreatedAt, supplier.UpdatedAt).Scan(&supplier.ID)
	if err != nil {
		return fmt.Errorf("failed to create supplier: %v", err)
	}
	return nil
}

// IsPhoneNumberVerified проверяет, верифицирован ли номер телефона
func (s *SupplierService) IsPhoneNumberVerified(phoneNumber string) (bool, error) {
	query := `SELECT is_verified FROM supplier WHERE phone_number = $1`

	var isVerified bool
	err := db.DB.QueryRow(query, phoneNumber).Scan(&isVerified)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return false, errors.New("supplier not found")
		}
		return false, fmt.Errorf("error checking verification status: %v", err)
	}

	return isVerified, nil
}

type UpdateSupplierRequest struct {
	Name       string `json:"name" binding:"required"`
	MarketName string `json:"market_name" binding:"required"`
	PlacesRows string `json:"places_rows" binding:"required"`
	Category   string `json:"category" binding:"required"`
}

// MarkPhoneNumberAsVerified обновляет поле is_verified для поставщика
func (s *SupplierService) MarkPhoneNumberAsVerified(phoneNumber string) error {
	supplier, err := s.GetSupplierByPhoneNumber(phoneNumber)
	if err != nil {
		if err.Error() == "supplier not found" {
			// Если поставщик не найден, создаём новую запись
			supplier = &models.Supplier{
				PhoneNumber: phoneNumber,
				IsVerified:  true,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			return s.CreateSupplier(supplier)
		} else {
			return fmt.Errorf("Ошибка при получении поставщика: %v", err)
		}
	}

	// Если поставщик найден, обновляем его статус
	query := `UPDATE supplier SET is_verified = $1, updated_at = $2 WHERE phone_number = $3`
	_, err = db.DB.Exec(query, true, time.Now(), phoneNumber)
	if err != nil {
		return fmt.Errorf("Не удалось обновить статус поставщика: %v", err)
	}

	return nil
}

func (s *SupplierService) SendVerificationCode(phoneNumber string) error {
	code := utils.GenerateSixDigitCode()

	// Шаг 2: Формирование сообщения
	message := fmt.Sprintf("Ваш код подтверждения: %s", code)

	// Шаг 3: Отправка сообщения пользователю
	err := utils.SendTextMessage(message, phoneNumber)
	if err != nil {
		fmt.Println("Error sending message:", err)
		return err
	}

	// Шаг 4: Удаление старых кодов для данного номера телефона
	err = db.DeleteVerificationCodes(phoneNumber)
	if err != nil {
		fmt.Println("Error deleting old verification codes:", err)
		return err
	}

	// Шаг 5: Сохранение нового кода в базе данных с временем истечения
	expiresAt := time.Now().Add(10 * time.Minute)
	err = db.CreateVerificationCode(phoneNumber, code, expiresAt)
	if err != nil {
		fmt.Println("Error saving verification code:", err)
		return err
	}

	return nil
}

func (s *SupplierService) ValidateVerificationCode(phone_number, code string) bool {
	log.Printf("Валидация кода подтверждения %s для пользователя %s", code, phone_number)
	verificationCode, err := db.GetLatestVerificationCode(phone_number)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Код подтверждения не найден для пользователя %s", phone_number)
		} else {
			log.Printf("Ошибка при получении кода подтверждения для %s: %v", phone_number, err)
		}
		return false
	}

	// Проверяем, не истёк ли код
	if time.Now().After(verificationCode.ExpiresAt) {
		log.Printf("Код подтверждения %s для пользователя %s истёк в %s", code, phone_number, verificationCode.ExpiresAt)
		return false
	}

	// Сравниваем введённый код с сохранённым
	if code == verificationCode.Code {
		log.Printf("Код подтверждения %s для пользователя %s верен", code, phone_number)
		return true
	}

	log.Printf("Код подтверждения %s для пользователя %s неверен", code, phone_number)
	return false
}

func (s *SupplierService) LinkUserToSupplier(phoneNumber string, userID int) error {
	query := `UPDATE supplier SET user_id = $1 WHERE phone_number = $2`
	result, err := db.DB.Exec(query, userID, phoneNumber)
	if err != nil {
		return fmt.Errorf("Не удалось связать пользователя с поставщиком: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Ошибка при получении количества затронутых строк: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("Поставщик с номером телефона %s не найден", phoneNumber)
	}

	return nil
}

func (s *SupplierService) UpdateSupplierDetailsByUserID(userID int, marketID int, place string, row_name string, categories []int) error {
	query := `UPDATE supplier 
              SET market_id = $1, place_name = $2, row_name = $3, categories = $4, updated_at = $5
              WHERE user_id = $6`

	_, err := db.DB.Exec(query, marketID, place, row_name, pq.Array(categories), time.Now(), userID)
	if err != nil {
		return fmt.Errorf("Не удалось обновить данные поставщика: %v", err)
	}

	return nil
}

func (s *SupplierService) GetAllMarkets() ([]models.Market, error) {
	query := `SELECT id, name FROM market`

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при получении рынков: %v", err)
	}
	defer rows.Close()

	var markets []models.Market
	for rows.Next() {
		var market models.Market
		if err := rows.Scan(&market.ID, &market.Name); err != nil {
			return nil, fmt.Errorf("Ошибка при сканировании рынка: %v", err)
		}
		markets = append(markets, market)
	}

	return markets, nil
}

func (s *SupplierService) GetAllCategories() ([]models.Category, error) {
	categories, err := db.GetAllCategories()
	if err != nil {
		log.Printf("GetAllCategories: ошибка при получении категорий: %v", err)
		return nil, fmt.Errorf("не удалось получить категории: %v", err)
	}
	return categories, nil
}

func (s *SupplierService) GetSupplierByPhoneNumber(phoneNumber string) (*models.Supplier, error) {
	supplier := &models.Supplier{}

	query := `SELECT id, phone_number, is_verified, created_at, updated_at FROM supplier WHERE phone_number = $1`

	err := db.DB.QueryRow(query, phoneNumber).Scan(
		&supplier.ID,
		&supplier.PhoneNumber,
		&supplier.IsVerified,
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

func (s *SupplierService) GetCategoryByPath(path string) (*models.Category, error) {
	category, err := db.GetCategoryByPath(path)
	if err != nil {
		log.Printf("GetCategoryByPath: ошибка при получении категории по path %s: %v", path, err)
		return nil, fmt.Errorf("не удалось получить категорию по path: %v", err)
	}
	return category, nil
}

func (s *SupplierService) AddCategory(name, path, imageURL string) (*models.Category, error) {
	// Проверка уникальности path
	existingCategory, err := db.GetCategoryByPath(path)
	if err == nil && existingCategory != nil {
		return nil, fmt.Errorf("категория с path '%s' уже существует", path)
	} else if err != nil && err.Error() != "категория не найдена для path "+path {
		log.Printf("AddCategory: ошибка при проверке существования категории: %v", err)
		return nil, fmt.Errorf("ошибка при проверке существования категории: %v", err)
	}

	// Создание новой категории
	category := &models.Category{
		Name:     name,
		Path:     path,
		ImageURL: imageURL,
	}

	err = db.CreateCategory(category)
	if err != nil {
		log.Printf("AddCategory: ошибка при создании категории: %v", err)
		return nil, fmt.Errorf("ошибка при создании категории: %v", err)
	}

	return category, nil
}
