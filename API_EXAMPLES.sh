#!/bin/bash

# 🎙️ radioKeenetik - API Examples
# Примеры для тестирования всех функций

set -e

# Цвета
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

API_URL="http://localhost:8080/api"
WS_URL="ws://localhost:8080/ws"

# Функции
info() { echo -e "${BLUE}ℹ${NC} $1"; }
success() { echo -e "${GREEN}✓${NC} $1"; }
error() { echo -e "${RED}✗${NC} $1"; }
warn() { echo -e "${YELLOW}⚠${NC} $1"; }

clear
echo "╔════════════════════════════════════════════════════════╗"
echo "║  🎙️  radioKeenetik - Примеры API                      ║"
echo "╚════════════════════════════════════════════════════════╝"
echo ""

# ============================================================
# 1. ПРОВЕРКА ПОДКЛЮЧЕНИЯ
# ============================================================
info "1. Проверка подключения к серверу..."
if curl -s -f http://localhost:8080/ > /dev/null; then
    success "Сервер доступен на http://localhost:8080"
else
    error "Сервер не запущен! Запустите: ./radioKeenetik"
    exit 1
fi
echo ""

# ============================================================
# 2. REST API ПРИМЕРЫ
# ============================================================
echo "╔════════════════════════════════════════════════════════╗"
echo "║  REST API Примеры                                      ║"
echo "╚════════════════════════════════════════════════════════╝"
echo ""

info "2a. Получить список всех станций"
info "Команда: curl -s $API_URL/stations | jq ."
echo ""
curl -s $API_URL/stations | jq . | head -20
echo ""

info "2b. Получить текущую станцию"
info "Команда: curl -s $API_URL/current | jq ."
echo ""
curl -s $API_URL/current | jq .
echo ""

info "2c. Получить состояние воспроизведения"
info "Команда: curl -s $API_URL/playback | jq ."
echo ""
curl -s $API_URL/playback | jq .
echo ""

info "2d. Получить доступные USB карты"
info "Команда: curl -s $API_URL/audio-devices | jq ."
echo ""
curl -s $API_URL/audio-devices | jq .
echo ""

# ============================================================
# 3. УПРАВЛЕНИЕ ВОСПРОИЗВЕДЕНИЕМ
# ============================================================
echo "╔════════════════════════════════════════════════════════╗"
echo "║  Управление Воспроизведением                           ║"
echo "╚════════════════════════════════════════════════════════╝"
echo ""

info "3a. Запустить станцию ID=1"
info "Команда: curl -X POST $API_URL/play?id=1 | jq ."
echo ""
curl -s -X POST "$API_URL/play?id=1" | jq .
echo ""

sleep 2

info "3b. Получить текущую станцию (должна быть запущена)"
echo ""
curl -s $API_URL/current | jq '.name'
echo ""

info "3c. Остановить воспроизведение"
info "Команда: curl -X POST $API_URL/stop | jq ."
echo ""
curl -s -X POST "$API_URL/stop" | jq .
echo ""

# ============================================================
# 4. ПРИМЕРЫ КОМАНД
# ============================================================
echo "╔════════════════════════════════════════════════════════╗"
echo "║  Полезные Команды                                      ║"
echo "╚════════════════════════════════════════════════════════╝"
echo ""

echo "📌 Копируйте эти команды для использования:"
echo ""

echo "▶️ Запустить станцию:"
echo "  ${YELLOW}curl -X POST http://localhost:8080/api/play?id=1${NC}"
echo ""

echo "⏹️ Остановить:"
echo "  ${YELLOW}curl -X POST http://localhost:8080/api/stop${NC}"
echo ""

echo "📻 Все станции (красиво):"
echo "  ${YELLOW}curl -s http://localhost:8080/api/stations | jq '.[].name'${NC}"
echo ""

echo "🎵 Только названия и жанры:"
echo "  ${YELLOW}curl -s http://localhost:8080/api/stations | jq '.[] | {name, genre}'${NC}"
echo ""

echo "🔊 Громкость (требует WebSocket):"
echo "  WebSocket приложение отправит:"
echo "  ${YELLOW}{ \"action\": \"setVolume\", \"volume\": 80 }${NC}"
echo ""

echo "🎧 Выбрать USB карту:"
echo "  ${YELLOW}{ \"action\": \"setDevice\", \"device\": \"hw:1,0\" }${NC}"
echo ""

echo "➕ Добавить станцию (через WebSocket):"
echo "  ${YELLOW}{ \"action\": \"addStation\", \"station\": { \"id\": \"1\", \"name\": \"My Radio\", \"url\": \"https://example.com/stream.mp3\", \"genre\": \"Pop\", \"country\": \"Russia\" } }${NC}"
echo ""

# ============================================================
# 5. PYTHON ПРИМЕРЫ
# ============================================================
echo ""
echo "╔════════════════════════════════════════════════════════╗"
echo "║  Python Примеры                                        ║"
echo "╚════════════════════════════════════════════════════════╝"
echo ""

cat << 'PYTHON_EOF'
# Python скрипт для управления radioKeenetik
import requests
import json

API_URL = 'http://localhost:8080/api'

# Получить все станции
def get_stations():
    response = requests.get(f'{API_URL}/stations')
    return response.json()

# Запустить станцию
def play_station(station_id):
    response = requests.post(f'{API_URL}/play?id={station_id}')
    return response.json()

# Остановить
def stop():
    response = requests.post(f'{API_URL}/stop')
    return response.json()

# Получить текущую станцию
def get_current():
    response = requests.get(f'{API_URL}/current')
    return response.json()

# Использование
if __name__ == '__main__':
    stations = get_stations()
    print(f"Всего станций: {len(stations)}")
    
    for station in stations:
        print(f"  {station['name']} - {station['genre']}")
    
    if stations:
        play_station(stations[0]['id'])
        print(f"Запущена: {get_current()['name']}")
PYTHON_EOF
echo ""

# ============================================================
# 6. JAVASCRIPT ПРИМЕРЫ
# ============================================================
echo "╔════════════════════════════════════════════════════════╗"
echo "║  JavaScript Примеры (Browser Console)                  ║"
echo "╚════════════════════════════════════════════════════════╝"
echo ""

cat << 'JS_EOF'
// Откройте DevTools (F12) и скопируйте эти команды

// Получить все станции
fetch('http://localhost:8080/api/stations')
  .then(r => r.json())
  .then(stations => console.log(stations));

// Запустить станцию
fetch('http://localhost:8080/api/play?id=1', { method: 'POST' })
  .then(r => r.json())
  .then(data => console.log(data));

// Остановить
fetch('http://localhost:8080/api/stop', { method: 'POST' })
  .then(r => r.json())
  .then(data => console.log(data));

// WebSocket пример
const ws = new WebSocket('ws://localhost:8080/ws');
ws.onopen = () => {
  ws.send(JSON.stringify({
    action: 'play',
    stationId: '1'
  }));
};
ws.onmessage = (event) => {
  console.log('Обновление:', JSON.parse(event.data));
};
EOF
echo ""

# ============================================================
# 7. CURL ПРИМЕРЫ (СКОПИРОВАТЬ-ВСТАВИТЬ)
# ============================================================
echo "╔════════════════════════════════════════════════════════╗"
echo "║  cURL Примеры для Копирования                          ║"
echo "╚════════════════════════════════════════════════════════╝"
echo ""

cat << 'CURL_EOF'
# Базовые операции
curl -s http://localhost:8080/api/stations | jq .
curl -s http://localhost:8080/api/current | jq .
curl -s http://localhost:8080/api/playback | jq .

# Управление
curl -X POST http://localhost:8080/api/play?id=1
curl -X POST http://localhost:8080/api/stop

# Фильтрация
curl -s http://localhost:8080/api/stations | jq '.[] | select(.genre == "Electronic")'

# Вывод только имён
curl -s http://localhost:8080/api/stations | jq -r '.[] | .name'

# Красивая таблица
curl -s http://localhost:8080/api/stations | jq '.[] | "\(.name) - \(.genre) (\(.country))"'

# Проверка доступности
curl -I http://localhost:8080/
CURL_EOF
echo ""

# ============================================================
# 8. МОНИТОРИНГ
# ============================================================
echo "╔════════════════════════════════════════════════════════╗"
echo "║  Мониторинг                                            ║"
echo "╚════════════════════════════════════════════════════════╝"
echo ""

info "Мониторинг в реальном времени (каждую секунду):"
echo ""
echo "  ${YELLOW}watch -n 1 'curl -s http://localhost:8080/api/current | jq .name'${NC}"
echo ""

info "Проверить процесс:"
echo "  ${YELLOW}ps aux | grep radioKeenetik${NC}"
echo ""

info "Проверить порт:"
echo "  ${YELLOW}netstat -tlnp | grep 8080${NC}"
echo ""

# ============================================================
# 9. СТАТИСТИКА
# ============================================================
echo "╔════════════════════════════════════════════════════════╗"
echo "║  Статистика                                            ║"
echo "╚════════════════════════════════════════════════════════╝"
echo ""

STATION_COUNT=$(curl -s $API_URL/stations | jq 'length')
success "Количество станций: $STATION_COUNT"

CURRENT=$(curl -s $API_URL/current | jq -r '.name // "Ничего"')
success "Текущая станция: $CURRENT"

STATE=$(curl -s $API_URL/playback | jq -r '.status')
success "Состояние: $STATE"

VOLUME=$(curl -s $API_URL/playback | jq -r '.volume')
success "Громкость: $VOLUME%"

echo ""

# ============================================================
# 10. ЗАКЛЮЧЕНИЕ
# ============================================================
echo "╔════════════════════════════════════════════════════════╗"
echo "║  ✅ Все примеры готовы к использованию               ║"
echo "╚════════════════════════════════════════════════════════╝"
echo ""

echo "📖 Документация:"
echo "  • Веб-интерфейс: http://localhost:8080"
echo "  • REST API: ${BLUE}$API_URL${NC}"
echo "  • WebSocket: ${BLUE}$WS_URL${NC}"
echo ""

echo "🔗 Полезные ссылки:"
echo "  • Документация: README.md"
echo "  • Быстрый старт: QUICKSTART.md"
echo "  • Изменения: CHANGES.md"
echo ""

echo "💡 Подсказка:"
echo "  Используйте 'jq' для красивого вывода JSON:"
echo "  ${YELLOW}sudo apt install jq${NC}"
echo ""

echo "Готово! 🎵"
echo ""
