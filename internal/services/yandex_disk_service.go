package services

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
)

type YandexDiskService struct {
	AccessToken string
}

func NewYandexDiskService(accessToken string) *YandexDiskService {
	return &YandexDiskService{AccessToken: accessToken}
}

// Создание папки на Яндекс.Диске
func (yd *YandexDiskService) CreateFolder(folderPath string) error {
	url := fmt.Sprintf("https://cloud-api.yandex.net/v1/disk/resources?path=%s", folderPath)

	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "OAuth "+yd.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
		return fmt.Errorf("failed to create folder: %v", resp.Status)
	}
	return nil
}

// Загрузка изображения на Яндекс.Диск
func (yd *YandexDiskService) UploadFile(folderPath string, file multipart.File, filename string) (string, error) {
	url := fmt.Sprintf("https://cloud-api.yandex.net/v1/disk/resources/upload?path=%s/%s&overwrite=true", folderPath, filename)

	// Получение ссылки для загрузки
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "OAuth "+yd.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var uploadInfo struct {
		href string `json:"href"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&uploadInfo); err != nil {
		return "", err
	}

	// Загрузка файла
	uploadReq, err := http.NewRequest("PUT", uploadInfo.href, file)
	if err != nil {
		return "", err
	}
	uploadReq.Header.Set("Content-Type", "application/octet-stream")

	uploadResp, err := http.DefaultClient.Do(uploadReq)
	if err != nil {
		return "", err
	}
	defer uploadResp.Body.Close()

	if uploadResp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("failed to upload file: %v", uploadResp.Status)
	}

	return fmt.Sprintf("https://disk.yandex.ru/disk/%s/%s", folderPath, filename), nil
}
