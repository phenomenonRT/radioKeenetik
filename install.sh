#!/bin/bash

# Установочный скрипт radioKeenetik для OpenWrt/LEDE
# Использование: bash install.sh

set -e

echo "🎙️  radioKeenetik - Установщик для OpenWrt"
echo "=========================================="
echo ""

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Функции для вывода
info() {
    echo -e "${GREEN}ℹ${NC} $1"
}

warn() {
    echo -e "${YELLOW}⚠${NC} $1"
}

error() {
    echo -e "${RED}✗${NC} $1"
    exit 1
}

success() {
    echo -e "${GREEN}✓${NC} $1"
}

# Проверка root
if [ "$EUID" -ne 0 ]; then 
    error "Этот скрипт должен запускаться с правами root (используйте sudo)"
fi

# Проверка доступа в интернет
info "Проверка доступа в интернет..."
if ! ping -c 1 8.8.8.8 > /dev/null 2>&1; then
    warn "Нет доступа в интернет. Продолжу с локальными файлами."
fi

# Переменные
INSTALL_DIR="/opt/radioKeenetik"
BINARY_NAME="radioKeenetik"
DOWNLOAD_URL="https://github.com/yourusername/radioKeenetik/releases/download/v1.0"
ARCH=$(uname -m)

# Определить архитектуру
case $ARCH in
    armv7l)
        BINARY_URL="${DOWNLOAD_URL}/radioKeenetik-linux-arm"
        BINARY_FILE="radioKeenetik-linux-arm"
        ;;
    armv6l)
        BINARY_URL="${DOWNLOAD_URL}/radioKeenetik-linux-arm"
        BINARY_FILE="radioKeenetik-linux-arm"
        ;;
    aarch64)
        BINARY_URL="${DOWNLOAD_URL}/radioKeenetik-linux-arm64"
        BINARY_FILE="radioKeenetik-linux-arm64"
        ;;
    x86_64)
        BINARY_URL="${DOWNLOAD_URL}/radioKeenetik-linux-amd64"
        BINARY_FILE="radioKeenetik-linux-amd64"
        ;;
    *)
        error "Неподдерживаемая архитектура: $ARCH"
        ;;
esac

success "Архитектура: $ARCH"
success "Будет загружен: $BINARY_FILE"

# Создать директорию
info "Создание директории установки..."
mkdir -p "$INSTALL_DIR"
success "Директория создана: $INSTALL_DIR"

# Скачать бинарник
info "Скачивание бинарника..."
cd "$INSTALL_DIR"

if command -v curl &> /dev/null; then
    curl -L -o "$BINARY_FILE" "$BINARY_URL" 2>/dev/null || {
        warn "Не удалось скачать. Проверьте URL и интернет-соединение."
    }
elif command -v wget &> /dev/null; then
    wget -O "$BINARY_FILE" "$BINARY_URL" 2>/dev/null || {
        warn "Не удалось скачать. Проверьте URL и интернет-соединение."
    }
else
    warn "curl и wget не найдены. Скопируйте бинарник вручную в $INSTALL_DIR"
fi

# Сделать исполняемым
if [ -f "$BINARY_FILE" ]; then
    chmod +x "$BINARY_FILE"
    success "Бинарник загружен и сделан исполняемым"
else
    warn "Бинарник не найден. Пожалуйста, скопируйте его вручную."
fi

# Создать init скрипт
info "Создание init скрипта..."
cat > "/etc/init.d/radioKeenetik" << 'EOF'
#!/bin/sh /etc/rc.common

START=99
STOP=01

SERVICE_PATH="/opt/radioKeenetik/radioKeenetik-linux-arm"
LISTEN_ADDR=":8080"

start() {
    if [ -x "$SERVICE_PATH" ]; then
        echo "Starting radioKeenetik..."
        $SERVICE_PATH > /var/log/radioKeenetik.log 2>&1 &
        echo $! > /var/run/radioKeenetik.pid
        echo "radioKeenetik started (PID: $(cat /var/run/radioKeenetik.pid))"
    else
        echo "Error: radioKeenetik binary not found at $SERVICE_PATH"
        return 1
    fi
}

stop() {
    echo "Stopping radioKeenetik..."
    if [ -f /var/run/radioKeenetik.pid ]; then
        kill $(cat /var/run/radioKeenetik.pid) 2>/dev/null
        rm /var/run/radioKeenetik.pid
    else
        killall radioKeenetik 2>/dev/null
    fi
    echo "radioKeenetik stopped"
}

restart() {
    stop
    sleep 1
    start
}

status() {
    if [ -f /var/run/radioKeenetik.pid ]; then
        if kill -0 $(cat /var/run/radioKeenetik.pid) 2>/dev/null; then
            echo "radioKeenetik is running (PID: $(cat /var/run/radioKeenetik.pid))"
            return 0
        fi
    fi
    echo "radioKeenetik is not running"
    return 1
}
EOF

chmod +x "/etc/init.d/radioKeenetik"
success "Init скрипт создан"

# Включить автозагрузку
info "Включение автозагрузки..."
/etc/init.d/radioKeenetik enable 2>/dev/null || {
    warn "Не удалось включить автозагрузку. Выполните вручную:"
    echo "  /etc/init.d/radioKeenetik enable"
}

# Создать логи
mkdir -p /var/log
touch /var/log/radioKeenetik.log
chmod 666 /var/log/radioKeenetik.log

# Открыть порт в firewall
info "Конфигурирование firewall..."
if grep -q "radioKeenetik" /etc/config/firewall 2>/dev/null; then
    success "Правило firewall уже существует"
else
    if [ -f /etc/config/firewall ]; then
        cat >> /etc/config/firewall << 'EOF'

config rule
    option name 'Allow-radioKeenetik'
    option src 'lan'
    option dest_port '8080'
    option proto 'tcp'
    option target 'ACCEPT'
EOF
        /etc/init.d/firewall restart 2>/dev/null || true
        success "Правило firewall добавлено"
    else
        warn "Файл firewall не найден. Добавьте правило вручную."
    fi
fi

# Вывод информации
echo ""
echo "======================================"
success "Установка завершена!"
echo "======================================"
echo ""
echo "📍 Путь установки: $INSTALL_DIR"
echo "🎵 Бинарник: $BINARY_FILE"
echo ""
echo "Следующие шаги:"
echo ""
echo "1️⃣  Запустить сервис:"
echo "   /etc/init.d/radioKeenetik start"
echo ""
echo "2️⃣  Проверить статус:"
echo "   /etc/init.d/radioKeenetik status"
echo ""
echo "3️⃣  Просмотреть логи:"
echo "   tail -f /var/log/radioKeenetik.log"
echo ""
echo "4️⃣  Открыть в браузере:"
echo "   http://192.168.1.1:8080"
echo ""
echo "5️⃣  Отключить/включить автозагрузку:"
echo "   /etc/init.d/radioKeenetik disable"
echo "   /etc/init.d/radioKeenetik enable"
echo ""
echo "🆘 Помощь:"
echo "   /etc/init.d/radioKeenetik restart  # Перезагрузить"
echo "   /etc/init.d/radioKeenetik stop     # Остановить"
echo "   ps aux | grep radioKeenetik        # Проверить процесс"
echo "   netstat -tlnp | grep 8080          # Проверить порт"
echo ""
