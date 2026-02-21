# Build stage
FROM golang:1.25-alpine AS builder

# Установка необходимых пакетов
RUN apk add --no-cache git

# Рабочая директория
WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код И docs
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

# Final stage
FROM alpine:latest

# Устанавливаем CA сертификаты и timezone data
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Копируем бинарник из builder
COPY --from=builder /app/main .

# Копируем миграции
COPY --from=builder /app/migrations ./migrations

# Копируем docs для Swagger
COPY --from=builder /app/docs ./docs

# Копируем .env файл (опционально)
COPY --from=builder /app/.env* ./

# Открываем порт
EXPOSE 8080

# Запускаем приложение
CMD ["./main"]