package utils

import (
	"fmt"
	"math/rand"
	"mime/multipart"
	"os"
	"time"
)

func SaveUploadedFile(file *multipart.FileHeader, fileName string) error {
	// Открываем файл
	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("Ошибка при открытии файла: %v", err)
	}
	defer src.Close()

	// Создаем файл на диске по указанному пути
	dst, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("Ошибка при создании файла: %v", err)
	}
	defer dst.Close()

	// Копируем содержимое
	_, err = dst.ReadFrom(src)
	if err != nil {
		return fmt.Errorf("Ошибка при копировании содержимого файла: %v", err)
	}

	return nil
}

func GenerateSKU() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%06d", rand.Intn(1000000)) // Генерация 6-значного случайного номера
}
