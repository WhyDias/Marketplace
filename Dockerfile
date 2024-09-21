# Используем официальный образ Golang в качестве базового
FROM golang:1.22-alpine

# Установка необходимых зависимостей (например, git)
RUN apk update && apk add --no-cache git

# Установка рабочего каталога внутри контейнера
WORKDIR /app

# Копирование файлов go.mod и go.sum
COPY go.mod go.sum ./

# Загрузка зависимостей
RUN go mod download

# Копирование всех файлов проекта
COPY . .

# Переход в директорию cmd, где находится main.go
WORKDIR /app/cmd

# Сборка приложения
RUN go build -o /app/main .

# Открытие порта 8080
EXPOSE 8080

# Команда для запуска приложения
CMD ["/app/main"]