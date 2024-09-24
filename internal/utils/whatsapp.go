// internal/utils/whatsapp.go

package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/WhyDias/Marketplace/internal/db"
	"golang.org/x/time/rate"
)

// ValidateWhatsAppCode валидирует код из WhatsApp
func ValidateWhatsAppCode(phoneNumber, code string) bool {
	// Получаем последний код для данного номера телефона
	verificationCode, err := db.GetLatestVerificationCode(phoneNumber)
	if err != nil {
		return false
	}

	// Проверяем, не истёк ли код
	if time.Now().After(verificationCode.ExpiresAt) {
		return false
	}

	// Сравниваем введённый код с сохранённым
	return code == verificationCode.Code
}

func SendTextMessage(messageBody string, recipient string) error {
	// Ожидаем разрешения от лимитера (ограничиваем количество запросов)
	err := limiter.Wait(context.Background())
	if err != nil {
		fmt.Println("Ошибка при ожидании разрешения от лимитера:", err)
		return err
	}

	// Параметры API
	profileID := "b0fbe69d-e68e"                         // Замените на ваш реальный profile_id
	apiKey := "234f33d4c7af8a62baeecdf1432fc9b5ffe911a1" // Замените на ваш реальный apiKey
	url := fmt.Sprintf("https://wappi.pro/api/sync/message/send?profile_id=%s", profileID)

	// Создание данных для отправки
	data := map[string]string{
		"body":      messageBody,
		"recipient": recipient,
	}

	// Преобразование данных в JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error marshaling data:", err)
		return err
	}

	// Создание HTTP-запроса
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return err
	}

	// Установка заголовков
	req.Header.Set("Authorization", apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Отправка HTTP-запроса
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return err
	}
	defer resp.Body.Close()

	// Обработка ответа
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		fmt.Println("Response:", result)
	} else {
		fmt.Printf("Error sending message to %s, status code: %d\n", recipient, resp.StatusCode)
		return fmt.Errorf("failed to send message, status code: %d", resp.StatusCode)
	}

	return nil
}

// Инициализируем генератор случайных чисел
var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

// Инициализируем лимитер (например, 1 запрос в секунду)
var limiter = rate.NewLimiter(1, 3)

// GenerateSixDigitCode генерирует случайный 6-значный код
func GenerateSixDigitCode() string {
	return fmt.Sprintf("%06d", seededRand.Intn(1000000))
}

// SendVerificationCode генерирует и отправляет код верификации через WhatsApp
func SendVerificationCode(phoneNumber string) error {
	// Генерация кода
	code := GenerateSixDigitCode()

	// Формирование сообщения
	message := fmt.Sprintf("Ваш код подтверждения: %s", code)

	// Отправка сообщения
	err := SendTextMessage(message, phoneNumber)
	if err != nil {
		return err
	}

	// Удаление старых кодов для данного номера телефона
	err = db.DeleteVerificationCodes(phoneNumber)
	if err != nil {
		return err
	}

	// Сохранение нового кода в базу данных с истечением через 10 минут
	expiresAt := time.Now().Add(10 * time.Minute)
	err = db.CreateVerificationCode(phoneNumber, code, expiresAt)
	if err != nil {
		return err
	}

	return nil
}
