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

	"golang.org/x/time/rate"
)

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
