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

// Инициализируем генератор случайных чисел
var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

// Инициализируем лимитер (например, 1 запрос в секунду)
var limiter = rate.NewLimiter(1, 3)

// GenerateSixDigitCode генерирует случайный 6-значный код
func GenerateSixDigitCode() string {
	return fmt.Sprintf("%06d", seededRand.Intn(1000000))
}

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

// sendTextMessage отправляет текстовое сообщение через WhatsApp API
func sendTextMessage(messageBody string, recipient string) {
	// Ожидаем разрешения от лимитера
	err := limiter.Wait(context.Background())
	if err != nil {
		fmt.Println("Ошибка при ожидании разрешения от лимитера:", err)
		return
	}

	id := "b0fbe69d-e68e" // Замените на ваш реальный profile_id
	url := fmt.Sprintf("https://wappi.pro/api/sync/message/send?profile_id=%s", id)
	apiKey := "234f33d4c7af8a62baeecdf1432fc9b5ffe911a1" // Замените на ваш реальный apiKey

	// Создание данных для отправки
	data := map[string]string{
		"body":      messageBody,
		"recipient": recipient,
	}

	// Преобразование данных в JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error marshaling data:", err)
		return
	}

	// Создание HTTP-запроса
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// Установка заголовков
	req.Header.Set("Authorization", apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Отправка HTTP-запроса
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	// Обработка ответа
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		fmt.Println("Response:", result)
	} else {
		fmt.Printf("Error sending message to %s, status code: %d\n", recipient, resp.StatusCode)
	}
}

// SendVerificationCode генерирует и отправляет код верификации через WhatsApp
func SendVerificationCode(phoneNumber string) error {
	// Генерация кода
	code := GenerateSixDigitCode()

	// Формирование сообщения
	message := fmt.Sprintf("Ваш код подтверждения: %s", code)

	// Отправка сообщения
	sendTextMessage(message, phoneNumber)

	// Удаление старых кодов для данного номера телефона
	err := db.DeleteVerificationCodes(phoneNumber)
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
