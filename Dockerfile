# Этап сборки
FROM golang:1.21-alpine AS builder

WORKDIR /build

# Скопировать исходный код
COPY . .

# Скачать зависимости
RUN go mod download

# Собрать приложение
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o radioKeenetik main.go

# Финальный образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Скопировать собранный бинарник
COPY --from=builder /build/radioKeenetik .

# Создать директорию для статических файлов
RUN mkdir -p static

# Экспортировать порт
EXPOSE 8080

# Переменные окружения для роутера
ENV LISTEN_ADDR=0.0.0.0:8080

# Запустить приложение
CMD ["./radioKeenetik"]
