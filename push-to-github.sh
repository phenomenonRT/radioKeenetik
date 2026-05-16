#!/bin/bash

# 🎙️ radioKeenetik - Автоматическая загрузка на GitHub
# Этот скрипт автоматически подготавливает и загружает проект на GitHub

set -e  # Выход при ошибке

# Цвета
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'  # No Color

# Функции
info() { echo -e "${BLUE}ℹ${NC} $1"; }
success() { echo -e "${GREEN}✓${NC} $1"; }
warn() { echo -e "${YELLOW}⚠${NC} $1"; }
error() { echo -e "${RED}✗${NC} $1"; exit 1; }

# Заголовок
clear
echo "╔═════════════════════════════════════════════════════════════════╗"
echo "║                                                                 ║"
echo "║  🎙️  radioKeenetik - Загрузка на GitHub                       ║"
echo "║                                                                 ║"
echo "╚═════════════════════════════════════════════════════════════════╝"
echo ""

# Проверка параметров
if [ $# -lt 1 ]; then
    error "Использование: $0 <github_username> [email@example.com]"
fi

USERNAME=$1
EMAIL=${2:-"$USERNAME@github.com"}

info "GitHub username: $USERNAME"
info "Email: $EMAIL"
echo ""

# Шаг 1: Проверить что Git установлен
info "Проверка Git..."
if ! command -v git &> /dev/null; then
    error "Git не установлен! Установите: sudo apt install git"
fi
success "Git найден: $(git --version)"
echo ""

# Шаг 2: Скопировать файлы
info "Копирование файлов проекта..."
PROJECT_DIR="$HOME/radioKeenetik"
if [ -d "$PROJECT_DIR" ]; then
    warn "Директория $PROJECT_DIR уже существует"
    read -p "Перезаписать? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm -rf "$PROJECT_DIR"
    else
        error "Отмено"
    fi
fi

cp -r "$(dirname "${BASH_SOURCE[0]}")" "$PROJECT_DIR"
success "Файлы скопированы в $PROJECT_DIR"
cd "$PROJECT_DIR"
echo ""

# Шаг 3: Проверить README
info "Проверка README.md..."
if [ -f "GITHUB_README.md" ] && [ ! -f "README.md" ]; then
    mv GITHUB_README.md README.md
    success "Переименован GITHUB_README.md → README.md"
elif [ -f "README.md" ]; then
    success "README.md уже существует"
else
    warn "README.md не найден"
fi
echo ""

# Шаг 4: Инициализировать Git
info "Инициализация Git..."
git init
git config user.name "$USERNAME"
git config user.email "$EMAIL"
success "Git инициализирован"
echo ""

# Шаг 5: Проверить файлы
info "Подготовка файлов..."
FILE_COUNT=$(find . -type f ! -path './.git/*' | wc -l)
success "Найдено $FILE_COUNT файлов"
echo ""

# Шаг 6: Добавить файлы
info "Добавление файлов в Git..."
git add .
STAGED=$(git status --short | wc -l)
success "Добавлено $STAGED файлов"
echo ""

# Шаг 7: Создать коммит
info "Создание коммита..."
git commit -m "Initial commit: radioKeenetik v1.1.0

🎙️ Internet Radio with USB Audio Card Support

Features:
✓ Full audio playback on USB audio cards
✓ Support for MP3, AAC, HLS streams
✓ Modern web interface with real-time updates
✓ REST API and WebSocket support
✓ Cross-platform (ARM, x86, ARM64)
✓ Complete documentation
✓ Production ready

Platforms:
- Ubuntu 24+ (and compatible)
- OpenWrt, DD-WRT routers
- Raspberry Pi
- Any Linux ARM/x86/ARM64 system

License: MIT"

success "Коммит создан"
echo ""

# Шаг 8: Добавить remote
info "Добавление GitHub репозитория..."
REPO_URL="https://github.com/${USERNAME}/radioKeenetik.git"
git remote add origin "$REPO_URL"
success "Remote добавлен: $REPO_URL"
echo ""

# Шаг 9: Переименовать ветку
info "Переименование ветки..."
git branch -M main
success "Ветка переименована в main"
echo ""

# Шаг 10: Показать инструкции перед push
echo "╔════════════════════════════════════════════════════════════════╗"
echo "║                  ⚠️  ВАЖНО!                                    ║"
echo "╚════════════════════════════════════════════════════════════════╝"
echo ""
echo "1️⃣  СОЗДАЙТЕ репозиторий на GitHub:"
echo "   https://github.com/new"
echo ""
echo "2️⃣  Заполните:"
echo "   - Repository name: radioKeenetik"
echo "   - Description: Internet Radio with USB Audio Card Support"
echo "   - Public (для публичного доступа)"
echo "   - НЕ инициализируйте README (у нас уже есть)"
echo ""
echo "3️⃣  Нажмите Create repository"
echo ""
echo "4️⃣  Если вам нужен Personal Access Token:"
echo "   https://github.com/settings/tokens"
echo "   - Generate new token (classic)"
echo "   - Выберите права: repo, workflow"
echo "   - Скопируйте token (будет видно только один раз!)"
echo ""
echo "────────────────────────────────────────────────────────────────"
read -p "Нажмите Enter когда репозиторий создан..." -r
echo ""

# Шаг 11: Push на GitHub
info "Отправка на GitHub (git push)..."
echo "Вас может попросить ввести пароль или token..."
echo ""

if git push -u origin main; then
    success "Успешно отправлено на GitHub!"
    echo ""
    echo "╔════════════════════════════════════════════════════════════════╗"
    echo "║                  ✅ ГОТОВО!                                    ║"
    echo "╚════════════════════════════════════════════════════════════════╝"
    echo ""
    echo "🌐 Откройте ваш репозиторий:"
    echo "   https://github.com/${USERNAME}/radioKeenetik"
    echo ""
    echo "📋 Дополнительные действия:"
    echo "   1. Добавьте описание репо (About)"
    echo "   2. Добавьте tags: go, radio, audio, router"
    echo "   3. Включите Discussions в Settings"
    echo "   4. Поделитесь проектом!"
    echo ""
    echo "🚀 Для дальнейших обновлений используйте:"
    echo "   cd $PROJECT_DIR"
    echo "   git add ."
    echo "   git commit -m 'Ваше сообщение'"
    echo "   git push origin main"
    echo ""
else
    error "Ошибка при загрузке на GitHub"
fi

# Шаг 12: Проверка
echo "⏳ Проверка загрузки..."
sleep 2

info "Загрузка завершена! 🎉"
echo ""
echo "Проект доступен по адресу:"
echo "  https://github.com/${USERNAME}/radioKeenetik"
echo ""
