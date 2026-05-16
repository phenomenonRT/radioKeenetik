.PHONY: help build run test clean deps install uninstall system-deps update

BINARY_NAME=radioKeenetik
MAIN_FILE=main.go
VERSION=$(shell git describe --tags --always 2>/dev/null || echo "1.0.0")
BUILD_TIME=$(shell date -u '+%Y-%m-%d %H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
INSTALL_DIR=/opt/radioKeenetik

help: ## 📖 Справка
	@echo "🎙️  radioKeenetik - Команды для Ubuntu 24+"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

system-deps: ## 🔧 Установить зависимости системы
	@echo "📦 Установка системных зависимостей для Ubuntu 24..."
	sudo apt update
	sudo apt install -y \
		curl \
		wget \
		git \
		build-essential \
		alsa-utils \
		pulseaudio \
		mpv \
		ffmpeg \
		sox \
		libatlas-base-dev \
		libjasper-dev \
		libtiff5-dev \
		libjasper1 \
		libharfbuzz0b \
		libwebp6 \
		libtiff5 \
		libwebp6 \
		libharfbuzz0b \
		libdc1394-25 \
		liblapack3 \
		libopenjp2-7 \
		libharfbuzz0b \
		libwebp6 \
		libtiff5 \
		fonts-dejavu-core
	@echo "✅ Зависимости установлены!"

deps: ## 📥 Скачать Go зависимости
	@echo "📥 Скачивание Go зависимостей..."
	go mod download
	go mod tidy
	@echo "✅ Зависимости скачаны!"

build: ## 🔨 Собрать приложение
	@echo "🔨 Сборка $(BINARY_NAME)..."
	go build $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_FILE)
	@echo "✅ Готово! Бинарник: ./$(BINARY_NAME)"
	@ls -lh $(BINARY_NAME)

build-arm: ## 🔨 Собрать для роутера ARM (Raspberry Pi, OpenWrt)
	@echo "🔨 Сборка для ARM (GOARCH=arm GOARM=7)..."
	GOOS=linux GOARCH=arm GOARM=7 go build $(LDFLAGS) -o $(BINARY_NAME)-arm $(MAIN_FILE)
	@echo "✅ Готово! Бинарник: ./$(BINARY_NAME)-arm"
	@ls -lh $(BINARY_NAME)-arm

build-arm64: ## 🔨 Собрать для ARM64 (Orange Pi, новые роутеры)
	@echo "🔨 Сборка для ARM64..."
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY_NAME)-arm64 $(MAIN_FILE)
	@echo "✅ Готово! Бинарник: ./$(BINARY_NAME)-arm64"
	@ls -lh $(BINARY_NAME)-arm64

build-x86: ## 🔨 Собрать для x86 роутеров
	@echo "🔨 Сборка для Linux x86..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-x86 $(MAIN_FILE)
	@echo "✅ Готово! Бинарник: ./$(BINARY_NAME)-x86"
	@ls -lh $(BINARY_NAME)-x86

build-all: build build-arm build-arm64 build-x86 ## 🔨 Собрать для всех платформ
	@echo "✅ Сборка для всех платформ завершена!"

run: build ## 🚀 Собрать и запустить локально
	@echo "🚀 Запуск radioKeenetik..."
	./$(BINARY_NAME)

test: ## 🧪 Запустить тесты
	@echo "🧪 Запуск тестов..."
	go test -v ./...

fmt: ## 🎨 Форматировать код
	@echo "🎨 Форматирование кода..."
	go fmt ./...
	goimports -w .
	@echo "✅ Готово!"

lint: ## 🔍 Проверить код
	@echo "🔍 Проверка кода..."
	go vet ./...
	@echo "✅ Проверка завершена!"

clean: ## 🧹 Очистить артефакты
	@echo "🧹 Очистка..."
	rm -f $(BINARY_NAME) $(BINARY_NAME)-* *.log
	go clean
	@echo "✅ Очищено!"

install: build ## 📦 Установить на эту систему
	@echo "📦 Установка в систему..."
	sudo mkdir -p $(INSTALL_DIR)
	sudo cp $(BINARY_NAME) $(INSTALL_DIR)/
	sudo chmod +x $(INSTALL_DIR)/$(BINARY_NAME)
	sudo ln -sf $(INSTALL_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@echo "✅ Установлено в $(INSTALL_DIR)"
	@echo ""
	@echo "Запустите:"
	@echo "  radioKeenetik"
	@echo ""
	@echo "Или создайте systemd сервис:"
	@echo "  sudo nano /etc/systemd/system/radioKeenetik.service"

uninstall: ## 🗑️  Удалить из системы
	@echo "🗑️  Удаление..."
	sudo rm -f /usr/local/bin/$(BINARY_NAME)
	sudo rm -rf $(INSTALL_DIR)
	@echo "✅ Удалено!"

version: ## 📌 Показать версию
	@echo "radioKeenetik $(VERSION)"
	@echo "Build: $(BUILD_TIME)"
	@echo "Commit: $(GIT_COMMIT)"

info: ## ℹ️  Информация о проекте
	@echo "🎙️  radioKeenetik"
	@echo "=================="
	@echo "Интернет-радио с USB аудиокартой на роутере"
	@echo ""
	@echo "Версия: $(VERSION)"
	@echo "Сборка: $(BUILD_TIME)"
	@echo "Commit: $(GIT_COMMIT)"
	@echo ""
	@echo "Поддерживаемые платформы:"
	@echo "  ✓ Linux AMD64"
	@echo "  ✓ Linux ARM (Raspberry Pi, роутеры)"
	@echo "  ✓ Linux ARM64"
	@echo ""
	@echo "Воспроизведение:"
	@echo "  ✓ MP3 потоки"
	@echo "  ✓ AAC/M4A потоки"
	@echo "  ✓ HLS потоки"
	@echo "  ✓ FLAC, OGG, и другие"
	@echo ""
	@echo "Аудиокарты:"
	@echo "  ✓ USB Audio Device"
	@echo "  ✓ Встроенная карта"
	@echo "  ✓ HDMI"
	@echo ""
	@echo "Плееры:"
	@echo "  ✓ mpv (рекомендуется)"
	@echo "  ✓ ffplay"
	@echo "  ✓ mplayer"
	@echo "  ✓ sox play"
	@echo ""
	@echo "Для справки: make help"

update: ## 🔄 Обновить зависимости
	@echo "🔄 Обновление зависимостей..."
	go get -u ./...
	go mod tidy
	@echo "✅ Обновлено!"

check-player: ## 🎵 Проверить доступные плееры
	@echo "🎵 Проверка доступных плееров..."
	@command -v mpv >/dev/null 2>&1 && echo "✓ mpv найден" || echo "✗ mpv не найден"
	@command -v ffplay >/dev/null 2>&1 && echo "✓ ffplay найден" || echo "✗ ffplay не найден"
	@command -v mplayer >/dev/null 2>&1 && echo "✓ mplayer найден" || echo "✗ mplayer не найден"
	@command -v play >/dev/null 2>&1 && echo "✓ sox play найден" || echo "✗ sox play не найден"

check-audio: ## 🔊 Проверить аудиокарты
	@echo "🔊 Доступные аудиокарты:"
	aplay -l 2>/dev/null || echo "Используйте: pactl list short sinks"

quick-install: system-deps deps build install ## ⚡ Быстрая установка (система + зависимости + сборка + установка)
	@echo "✅ Быстрая установка завершена!"
	@echo ""
	@echo "Запустите:"
	@echo "  radioKeenetik"
	@echo ""
	@echo "Затем откройте в браузере:"
	@echo "  http://localhost:8080"

.DEFAULT_GOAL := help
