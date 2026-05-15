#!/bin/bash

# radioKeenetik - Примеры использования API
# 
# Эти команды показывают, как взаимодействовать с API radioKeenetik
# через cURL и WebSocket

echo "🎙️  radioKeenetik - Примеры API"
echo "=================================="
echo ""

# Переменные
RADIO_URL="http://localhost:8080"

# Цвета
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

# Функция для вывода примера
example() {
    echo -e "${BLUE}$1${NC}"
    echo "---"
}

# ==========================================
# REST API ПРИМЕРЫ
# ==========================================

echo "REST API"
echo "========"
echo ""

example "1. Получить список всех станций"
cat << 'EOF'
curl -X GET http://localhost:8080/api/stations | jq .
EOF
echo ""

example "2. Получить информацию о текущей станции"
cat << 'EOF'
curl -X GET http://localhost:8080/api/current | jq .
EOF
echo ""

example "3. Запустить воспроизведение станции"
cat << 'EOF'
curl -X POST "http://localhost:8080/api/play?id=1" | jq .
EOF
echo ""

example "4. Остановить воспроизведение"
cat << 'EOF'
curl -X POST "http://localhost:8080/api/stop" | jq .
EOF
echo ""

# ==========================================
# PYTHON ПРИМЕРЫ
# ==========================================

echo ""
echo "Python примеры"
echo "=============="
echo ""

example "1. Получить список станций (Python)"
cat << 'EOF'
import requests
import json

response = requests.get('http://localhost:8080/api/stations')
stations = response.json()

for station in stations:
    print(f"{station['name']} - {station['genre']}")
EOF
echo ""

example "2. Управление плеером (Python)"
cat << 'EOF'
import requests
import json

# Получить текущую станцию
current = requests.get('http://localhost:8080/api/current').json()
print(f"Currently playing: {current.get('name', 'Nothing')}")

# Запустить станцию с ID=1
response = requests.post('http://localhost:8080/api/play?id=1')
print(response.json())

# Остановить
response = requests.post('http://localhost:8080/api/stop')
print(response.json())
EOF
echo ""

example "3. WebSocket (Python)"
cat << 'EOF'
import websocket
import json

def on_message(ws, message):
    data = json.loads(message)
    print(f"Update: {data['type']}")
    print(f"Current: {data.get('current', {}).get('name', 'None')}")

def on_open(ws):
    print("Connected!")
    # Отправить команду
    ws.send(json.dumps({
        "action": "play",
        "stationId": "1"
    }))

ws = websocket.WebSocketApp(
    "ws://localhost:8080/ws",
    on_message=on_message,
    on_open=on_open
)

ws.run_forever()
EOF
echo ""

# ==========================================
# JAVASCRIPT/Node.js ПРИМЕРЫ
# ==========================================

echo ""
echo "JavaScript/Node.js примеры"
echo "==========================="
echo ""

example "1. Получить список станций (JavaScript)"
cat << 'EOF'
fetch('http://localhost:8080/api/stations')
  .then(response => response.json())
  .then(stations => {
    stations.forEach(station => {
      console.log(`${station.name} - ${station.genre}`);
    });
  });
EOF
echo ""

example "2. WebSocket (JavaScript)"
cat << 'EOF'
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onopen = () => {
  console.log('Connected!');
  
  // Отправить команду
  ws.send(JSON.stringify({
    action: 'play',
    stationId: '1'
  }));
};

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Update:', data);
  console.log('Now playing:', data.current?.name || 'None');
};

ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};
EOF
echo ""

example "3. Управление плеером (Node.js + axios)"
cat << 'EOF'
const axios = require('axios');

const api = axios.create({
  baseURL: 'http://localhost:8080/api'
});

// Получить текущую станцию
async function getCurrent() {
  const response = await api.get('/current');
  return response.data;
}

// Запустить станцию
async function play(stationId) {
  const response = await api.post(`/play?id=${stationId}`);
  return response.data;
}

// Остановить
async function stop() {
  const response = await api.post('/stop');
  return response.data;
}

(async () => {
  console.log('Current:', await getCurrent());
  console.log('Playing:', await play('1'));
  await new Promise(r => setTimeout(r, 5000));
  console.log('Stopped:', await stop());
})();
EOF
echo ""

# ==========================================
# BASH/cURL ПРИМЕРЫ
# ==========================================

echo ""
echo "Bash/cURL примеры"
echo "================="
echo ""

example "1. Прямые примеры с cURL"
cat << 'EOF'
# Получить JSON красиво
curl -s http://localhost:8080/api/stations | jq .

# Получить только названия станций
curl -s http://localhost:8080/api/stations | jq -r '.[] | .name'

# Получить жанры
curl -s http://localhost:8080/api/stations | jq '.[] | {name, genre}'

# Запустить станцию и вывести ответ
curl -s -X POST http://localhost:8080/api/play?id=1 | jq .
EOF
echo ""

example "2. Скрипт для переключения станций"
cat << 'EOF'
#!/bin/bash

# Функция для переключения станций
switch_radio() {
  local station_id=$1
  echo "🎙️  Переключение на станцию: $station_id"
  curl -s -X POST "http://localhost:8080/api/play?id=$station_id" | jq .
  
  # Получить текущую станцию
  echo ""
  echo "Сейчас играет:"
  curl -s http://localhost:8080/api/current | jq '.name'
}

# Использование
switch_radio "1"
EOF
echo ""

example "3. Мониторинг в реальном времени"
cat << 'EOF'
#!/bin/bash

echo "Мониторинг radioKeenetik..."
while true; do
  clear
  echo "=== radioKeenetik Monitor ==="
  echo "Время: $(date)"
  echo ""
  echo "Текущая станция:"
  curl -s http://localhost:8080/api/current | jq '.name, .genre, .country'
  echo ""
  echo "Все станции:"
  curl -s http://localhost:8080/api/stations | jq '.[] | "\(.name) - \(.genre)"'
  echo ""
  sleep 5
done
EOF
echo ""

# ==========================================
# cURL ПРИМЕРЫ (прямые команды)
# ==========================================

echo ""
echo "Прямые cURL команды для копирования"
echo "===================================="
echo ""

cat << 'EOF'
# Получить все станции (красиво)
curl -s http://localhost:8080/api/stations | jq .

# Получить текущую станцию
curl -s http://localhost:8080/api/current | jq .

# Запустить станцию ID 1
curl -X POST "http://localhost:8080/api/play?id=1"

# Остановить
curl -X POST "http://localhost:8080/api/stop"

# Добавить новую станцию через WebSocket
# (используйте клиент WebSocket или JavaScript в браузере)

# Проверить доступность
curl -I http://localhost:8080/

# Получить HTML интерфейс
curl -s http://localhost:8080/ | head -50
EOF
echo ""

# ==========================================
# ТЕСТИРОВАНИЕ
# ==========================================

echo ""
echo "Тестирование"
echo "============"
echo ""

example "Проверить, запущен ли сервер"
if curl -s http://localhost:8080/ > /dev/null; then
    echo -e "${GREEN}✓ Сервер запущен${NC}"
else
    echo "✗ Сервер не запущен"
fi

echo ""
example "Получить количество станций"
STATION_COUNT=$(curl -s http://localhost:8080/api/stations | jq 'length')
echo "Количество станций: $STATION_COUNT"

echo ""
echo ""
echo "💡 Для полной документации см. README_RU.md"
echo ""
