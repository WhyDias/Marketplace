// internal/services/supplier_service.go

package services

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
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
	query := `INSERT INTO supplier (phone_number, is_verified, user_id, created_at, updated_at)
              VALUES ($1, $2, $3, $4, $5) RETURNING id`
	// Включаем user_id в запрос
	err := db.DB.QueryRow(query, supplier.PhoneNumber, supplier.IsVerified, supplier.UserID, supplier.CreatedAt, supplier.UpdatedAt).Scan(&supplier.ID)
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
func (s *SupplierService) MarkPhoneNumberAsVerified(phoneNumber string, userID int) error {
	supplier, err := s.GetSupplierByPhoneNumber(phoneNumber)
	if err != nil {
		if err.Error() == "supplier not found" {
			// Если поставщик не найден, создаём новую запись
			supplier = &models.Supplier{
				PhoneNumber: phoneNumber,
				IsVerified:  true,
				UserID:      userID, // Устанавливаем userID
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			return s.CreateSupplier(supplier)
		} else {
			return fmt.Errorf("Ошибка при получении поставщика: %v", err)
		}
	}

	// Если поставщик найден, обновляем его статус и userID
	query := `UPDATE supplier SET is_verified = $1, user_id = $2, updated_at = $3 WHERE phone_number = $4`
	_, err = db.DB.Exec(query, true, userID, time.Now(), phoneNumber)
	if err != nil {
		return fmt.Errorf("Не удалось обновить статус поставщика: %v", err)
	}

	return nil
}

func (s *SupplierService) SendVerificationCode(userID int, phoneNumber string) error {
	code := utils.GenerateSixDigitCode()
	message := fmt.Sprintf("Ваш код подтверждения: %s", code)

	// Отправляем сообщение через WhatsApp
	err := utils.SendTextMessage(message, phoneNumber)
	if err != nil {
		return fmt.Errorf("Ошибка при отправке сообщения: %v", err)
	}

	// Удаляем старые коды подтверждения для пользователя
	err = db.DeleteVerificationCodes(userID)
	if err != nil {
		return fmt.Errorf("Ошибка при удалении старых кодов подтверждения: %v", err)
	}

	// Сохраняем новый код подтверждения в базе данных
	expiresAt := time.Now().Add(10 * time.Minute)
	err = db.CreateVerificationCode(userID, code, expiresAt)
	if err != nil {
		return fmt.Errorf("Ошибка при сохранении кода подтверждения: %v", err)
	}

	return nil
}

func (s *SupplierService) ValidateVerificationCode(userID int, code string) bool {
	verificationCode, err := db.GetLatestVerificationCode(userID)
	if err != nil {
		return false
	}

	if time.Now().After(verificationCode.ExpiresAt) {
		return false
	}

	return code == verificationCode.Code
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

func (s *SupplierService) UpdateSupplierDetailsByUserID(userID int, marketID int, place string, rowName string, categoryIDs []int) error {
	err := db.UpdateSupplierDetailsByUserID(userID, marketID, place, rowName, categoryIDs)
	if err != nil {
		return fmt.Errorf("could not update supplier details: %v", err)
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

	// Добавляем обязательный атрибут "Цвет"
	mandatoryAttribute := models.CategoryAttribute{
		CategoryID:   category.ID,
		Name:         "Цвет",
		Description:  StringPtr("Выберите цвет товара"),
		TypeOfOption: StringPtr("dropdown"),
	}

	// Значения атрибута "Цвет"
	colorValues := []string{"Красный", "Синий", "Зеленый"}
	valueJSON, err := json.Marshal(colorValues)
	if err != nil {
		log.Printf("AddCategory: ошибка при сериализации значений цвета: %v", err)
		return nil, fmt.Errorf("ошибка при сериализации значений цвета: %v", err)
	}
	mandatoryAttribute.Value = valueJSON

	// Создание обязательного атрибута для новой категории
	_, err = db.CreateCategoryAttribute(&mandatoryAttribute)
	if err != nil {
		log.Printf("AddCategory: ошибка при создании обязательного атрибута 'Цвет' для категории %d: %v", category.ID, err)
		return nil, fmt.Errorf("ошибка при создании обязательного атрибута 'Цвет': %v", err)
	}

	return category, nil
}

// GetSupplierIDByUserID возвращает supplier_id по user_id
func (s *SupplierService) GetSupplierIDByUserID(userID int) (int, error) {
	supplierID, err := db.GetSupplierIDByUserID(userID)
	if err != nil {
		return 0, fmt.Errorf("Не удалось получить supplier_id для user_id %d: %v", userID, err)
	}
	return supplierID, nil
}

// GetCategoriesBySupplierID возвращает категории, связанные с поставщиком
func (s *SupplierService) GetCategoriesBySupplierID(supplierID int) ([]models.Category, error) {
	categories, err := db.GetCategoriesBySupplierID(supplierID)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при получении категорий для поставщика %d: %v", supplierID, err)
	}
	return categories, nil
}
