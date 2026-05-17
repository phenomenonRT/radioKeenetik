# Этап сборки
FROM golang:1.21-alpine AS builder

WORKDIR /build

# Скопировать исходный код
COPY app.go server.go audio.go go.mod go.sum ./

# Скачать зависимости
RUN go mod download

# Собрать приложение
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o radioKeenetik app.go server.go audio.go

# Финальный образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates mpv

WORKDIR /app

# Скопировать собранный бинарник
COPY --from=builder /build/radioKeenetik .

# Экспортировать порт
EXPOSE 8080

# Запустить приложение
CMD ["./radioKeenetik"]
