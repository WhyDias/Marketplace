// internal/services/supplier_service.go

package services

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"log"
	"strconv"
	"strings"
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

// GetSupplierInfo получает информацию о поставщике по номеру телефона
func (s *SupplierService) GetSupplierInfo(phoneNumber string) (*models.Supplier, error) {
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
		if err.Error() == "sql: no rows in result set" {
			return nil, errors.New("supplier not found")
		}
		return nil, fmt.Errorf("error fetching supplier: %v", err)
	}

	return supplier, nil
}

// GetAllSuppliers получает список всех поставщиков
func (s *SupplierService) GetAllSuppliers() ([]models.Supplier, error) {
	query := `SELECT id, name, phone_number, market_id, place_name, row_name, categories, created_at, updated_at, is_verified FROM supplier`

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error fetching suppliers: %v", err)
	}
	defer rows.Close()

	var suppliers []models.Supplier
	for rows.Next() {
		var supplier models.Supplier
		if err := rows.Scan(&supplier.ID, &supplier.Name, &supplier.PhoneNumber, &supplier.MarketID, &supplier.Place, &supplier.RowName, &supplier.Categories, &supplier.CreatedAt, &supplier.UpdatedAt, &supplier.IsVerified); err != nil {
			return nil, fmt.Errorf("error scanning supplier: %v", err)
		}
		suppliers = append(suppliers, supplier)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %v", err)
	}

	return suppliers, nil
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

// GetAllBazaars получает список всех базаров
func (s *SupplierService) GetAllBazaars() ([]models.Bazaar, error) {
	query := `SELECT id, name FROM bazaar`

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error fetching bazaars: %v", err)
	}
	defer rows.Close()

	var bazaars []models.Bazaar
	for rows.Next() {
		var bazaar models.Bazaar
		if err := rows.Scan(&bazaar.ID, &bazaar.Name); err != nil {
			return nil, fmt.Errorf("error scanning bazaar: %v", err)
		}
		bazaars = append(bazaars, bazaar)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %v", err)
	}

	return bazaars, nil
}

// CreatePlace создает новое место
func (s *SupplierService) CreatePlace(place *models.Place) error {
	query := `INSERT INTO places (name) VALUES ($1) RETURNING id`
	err := db.DB.QueryRow(query, place.Name).Scan(&place.ID)
	if err != nil {
		return fmt.Errorf("failed to create place: %v", err)
	}
	return nil
}

// CreateRow создает новый ряд
func (s *SupplierService) CreateRow(row *models.Row) error {
	query := `INSERT INTO rows (name, place_id) VALUES ($1, $2) RETURNING id`
	err := db.DB.QueryRow(query, row.Name, row.PlaceID).Scan(&row.ID)
	if err != nil {
		return fmt.Errorf("failed to create row: %v", err)
	}
	return nil
}

// Функция для преобразования []int в []string
func intSliceToStringSlice(ints []int) []string {
	strs := make([]string, len(ints))
	for i, v := range ints {
		strs[i] = strconv.Itoa(v)
	}
	return strs
}

func (s *SupplierService) CreateUser(phoneNumber string) error {
	query := `INSERT INTO users (username, password_hash, created_at, updated_at, role) 
              VALUES ($1, $2, $3, $4, 'supplier')` // Измените на подходящие значения

	// Пример значений:
	username := phoneNumber           // Или любое другое значение, которое вы хотите использовать
	passwordHash := "hashed_password" // Замените на фактический хэш пароля

	_, err := db.DB.Exec(query, username, passwordHash, time.Now(), time.Now())
	return err
}

// UpdateSupplierDetails обновляет данные поставщика
func (s *SupplierService) UpdateSupplierDetails(phoneNumber string, marketID int, placeName string, rowName string, categories []int) error {
	query := `UPDATE supplier 
	          SET market_id = $1, 
	              place_name = $2, 
	              row_name = $3, 
	              category_ids = $4, 
	              updated_at = $5 
	          WHERE phone_number = $6`

	// Преобразуем массив категорий в строку (например, через запятую)
	categoryIDs := strings.Join(intSliceToStringSlice(categories), ",")

	_, err := db.DB.Exec(query, marketID, placeName, rowName, categoryIDs, time.Now(), phoneNumber)
	if err != nil {
		return fmt.Errorf("failed to update supplier details: %v", err)
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

func (s *SupplierService) ValidateVerificationCode(username, code string) bool {
	log.Printf("Валидация кода подтверждения %s для пользователя %s", code, username)
	verificationCode, err := db.GetLatestVerificationCode(username)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Код подтверждения не найден для пользователя %s", username)
		} else {
			log.Printf("Ошибка при получении кода подтверждения для %s: %v", username, err)
		}
		return false
	}

	// Проверяем, не истёк ли код
	if time.Now().After(verificationCode.ExpiresAt) {
		log.Printf("Код подтверждения %s для пользователя %s истёк в %s", code, username, verificationCode.ExpiresAt)
		return false
	}

	// Сравниваем введённый код с сохранённым
	if code == verificationCode.Code {
		log.Printf("Код подтверждения %s для пользователя %s верен", code, username)
		return true
	}

	log.Printf("Код подтверждения %s для пользователя %s неверен", code, username)
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
	query := `SELECT id, name, path FROM categories`

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при получении категорий: %v", err)
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var category models.Category
		if err := rows.Scan(&category.ID, &category.Name, &category.Path); err != nil {
			return nil, fmt.Errorf("Ошибка при сканировании категории: %v", err)
		}
		categories = append(categories, category)
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
