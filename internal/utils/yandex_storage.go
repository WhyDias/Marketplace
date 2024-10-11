package utils

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
)

var (
	yandexBucketName = "dias"
	yandexRegion     = "ru-central1"
	yandexEndpoint   = "https://storage.yandexcloud.net"
	yandexAccessKey  = "YCAJE3coFn84R2S5dfwCpbQ2B"
	yandexSecretKey  = "YCOkgDHLVUWcPPpW-KgpoVzuyx2S9C0O2GU87d63"
)

func UploadFileToYandex(file *multipart.FileHeader, folderPath string) (string, error) {
	// Открываем файл
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("Не удалось открыть файл: %v", err)
	}
	defer src.Close()

	// Читаем файл в буфер
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, src); err != nil {
		return "", fmt.Errorf("Ошибка при чтении файла: %v", err)
	}

	// Генерируем уникальное имя файла
	uniqueFileName := uuid.New().String() + filepath.Ext(file.Filename)

	// Создаем полный путь внутри бакета
	// Если folderPath пустой, файл сохраняется в корне бакета
	var objectKey string
	if folderPath != "" {
		objectKey = fmt.Sprintf("%s/%s", folderPath, uniqueFileName)
	} else {
		objectKey = uniqueFileName
	}

	// Создаем сессию AWS
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(yandexRegion),
		Endpoint:         aws.String(yandexEndpoint),
		Credentials:      credentials.NewStaticCredentials(yandexAccessKey, yandexSecretKey, ""),
		S3ForcePathStyle: aws.Bool(true),
	})
	if err != nil {
		return "", fmt.Errorf("Не удалось создать сессию: %v", err)
	}

	// Создаем клиент S3
	svc := s3.New(sess)

	// Загружаем файл
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket:        aws.String(yandexBucketName),
		Key:           aws.String(objectKey),
		Body:          bytes.NewReader(buf.Bytes()),
		ContentLength: aws.Int64(file.Size),
		ContentType:   aws.String(file.Header.Get("Content-Type")),
		ACL:           aws.String("public-read"),
	})
	if err != nil {
		return "", fmt.Errorf("Ошибка при загрузке файла в Yandex Cloud Storage: %v", err)
	}

	// Формируем URL загруженного файла
	fileURL := fmt.Sprintf("%s/%s/%s", yandexEndpoint, yandexBucketName, objectKey)

	return fileURL, nil
}
