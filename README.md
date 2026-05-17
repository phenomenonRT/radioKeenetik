# 🎙️ radioKeenetik - Интернет-радио на USB аудиокарте

## ✨ Что было исправлено

### 🐛 Критические баги
1. **JavaScript синтаксис в Go коде** ❌ → ✅ Исправлено
   ```javascript
   // БЫЛО (неправильно):
   div.Set("innerHTML", '...') // Go синтаксис в JS!
   
   // СТАЛО (правильно):
   div.innerHTML = '...' // Корректный JavaScript
   ```

2. **Неуведомление клиентов при добавлении станций** ❌ → ✅ Исправлено
   ```go
   // БЫЛО:
   // rs.notifyClients()  // Закомментирована!
   
   // СТАЛО:
   rs.notifyClients()  // Вызывается правильно
   ```

3. **Отсутствие модульности** ❌ → ✅ Разделено на модули
   - `app.go` - главный файл с инициализацией
   - `server.go` - серверная логика и маршруты
   - `audio.go` - работа с плеером и аудиокартами

### 🎨 Улучшения интерфейса
- ✅ Полностью адаптивный дизайн (мобильный, планшет, ПК)
- ✅ Темная тема по умолчанию
- ✅ Реал-тайм обновления через WebSocket
- ✅ Поиск по названию, жанру, стране
- ✅ Система избранного (⭐)
- ✅ Удаление станций
- ✅ Валидация URL потоков
- ✅ Лучший обработчик ошибок
- ✅ Индикаторы загрузки

### 🏗️ Архитектурные улучшения
- ✅ Разделение на логические модули
- ✅ Убран дублирующийся код
- ✅ Добавлены комментарии
- ✅ Правильное управление ошибками
- ✅ Безопасные операции с concurrency

## 📋 Структура проекта

```
radioKeenetik/
├── app.go              # Главный файл, инициализация сервера
├── server.go           # Серверная логика, маршруты, WebSocket
├── audio.go            # Работа с плеером и аудиокартами
├── main.go             # Альтернативный монолитный вариант (для совместимости)
├── go.mod              # Файл зависимостей Go
├── go.sum              # Хеши зависимостей
├── Makefile            # Команды для сборки
├── Dockerfile          # Контейнеризация
├── docker-compose.yml  # Docker Compose конфиг
├── install.sh          # Установка на OpenWrt
├── push-to-github.sh   # Загрузка на GitHub
├── API_EXAMPLES.sh     # Примеры использования API
└── ROUTER_SETUP.md     # Подробное руководство
```

## 🚀 Быстрый старт

### 1️⃣ На Ubuntu 24+

```bash
# Установить зависимости
sudo apt update
sudo apt install -y go-lang git build-essential mpv

# Клонировать проект (или скачать файлы)
git clone https://github.com/yourusername/radioKeenetik.git
cd radioKeenetik

# Скачать зависимости Go
go mod download

# Собрать приложение
go build -o radioKeenetik *.go

# Запустить
./radioKeenetik
```

Затем откройте: **http://localhost:8080**

### 2️⃣ С использованием Makefile

```bash
# Установить все зависимости и собрать
make quick-install

# Запустить
radioKeenetik

# Открыть в браузере
# http://localhost:8080
```

### 3️⃣ Собрать для роутера (ARM)

```bash
# Для большинства роутеров
make build-arm
# Результат: radioKeenetik-arm

# Загрузить на роутер
scp radioKeenetik-arm root@192.168.1.1:/opt/

# Подключиться к роутеру
ssh root@192.168.1.1

# Запустить
/opt/radioKeenetik-arm
```

## 📊 Основные изменения

### Файл `app.go` - Главный файл
- Чистая инициализация приложения
- Graceful shutdown
- Логирование

### Файл `server.go` - Серверная логика
```go
// Создание сервера
rs := NewRadioServer()

// Регистрация маршрутов
r := mux.NewRouter()
rs.RegisterRoutes(r)

// Запуск
http.ListenAndServe(":8080", r)
```

### Файл `audio.go` - Работа с аудио
- Управление плеером (mpv, ffplay, mplayer)
- Управление громкостью
- Работа с USB аудиокартами

## 🎵 Использование

### Веб-интерфейс

1. **Откройте в браузере**: `http://localhost:8080`
2. **Добавьте станции** через форму "➕ Добавить станцию"
3. **Нажмите ▶️** для воспроизведения
4. **Отрегулируйте громкость** 🔊
5. **Выберите USB карту** 🎧

### REST API

```bash
# Получить список станций
curl http://localhost:8080/api/stations | jq .

# Получить текущую станцию
curl http://localhost:8080/api/current | jq .

# Запустить станцию (ID=1)
curl -X POST http://localhost:8080/api/play?id=1

# Остановить
curl -X POST http://localhost:8080/api/stop

# Получить состояние
curl http://localhost:8080/api/playback | jq .

# Получить USB карты
curl http://localhost:8080/api/audio-devices | jq .
```

### WebSocket API

Подключитесь к `ws://localhost:8080/ws` и отправляйте команды:

```json
// Запустить станцию
{
  "action": "play",
  "stationId": "1"
}

// Остановить
{
  "action": "stop"
}

// Установить громкость
{
  "action": "setVolume",
  "volume": 80
}

// Выбрать USB карту
{
  "action": "setDevice",
  "device": "hw:1,0"
}

// Добавить станцию
{
  "action": "addStation",
  "station": {
    "id": "4",
    "name": "My Radio",
    "url": "https://example.com/stream.mp3",
    "genre": "Pop",
    "country": "Russia"
  }
}

// Добавить в избранное
{
  "action": "toggleFavorite",
  "stationId": "1"
}

// Удалить станцию
{
  "action": "deleteStation",
  "stationId": "1"
}
```

## 🔧 Команды Makefile

```bash
make help              # Справка
make system-deps       # Установить системные зависимости
make deps              # Скачать Go зависимости
make build             # Собрать для Linux
make build-arm         # Собрать для ARM роутеров
make build-arm64       # Собрать для ARM64
make build-x86         # Собрать для x86
make build-all         # Собрать для всех платформ
make run               # Запустить локально
make test              # Тесты
make install           # Установить в систему
make quick-install     # Все в одной команде
make check-player      # Проверить доступные плееры
make check-audio       # Проверить аудиокарты
```

## 🎧 Проверка USB аудиокарты

```bash
# Список всех аудиокарт
aplay -l

# Вывод:
# **** List of PLAYBACK Hardware Devices ****
# card 0: Analog [Analog], device 0: ALC1220 Analog
# card 1: Device [USB Audio Device], device 0: USB Audio
#                   ↑ Ваша USB карта
```

## 🔊 Управление громкостью

### Веб-интерфейс
Используйте ползунок в приложении (0-100%)

### Командная строка
```bash
# Через amixer
amixer -c 1 set Master 80%

# Через pactl (PulseAudio)
pactl set-sink-volume 0 80%
```

## 🌐 Примеры потоков

### MP3
```
https://hls-01-radiorecord.hostingradio.ru/record-livedjsets/112/l0_6a054e73b4592b18390d2063.aac
```

### AAC
```
https://stream.example.com/aac.aac
```

### HLS
```
https://example.com/playlist.m3u8
```

## 🐛 Решение проблем

### Приложение не запускается
```bash
# Проверить плеер
which mpv
# или
make check-player
```

Установить mpv:
```bash
sudo apt install -y mpv
```

### Нет звука

1. **Проверить USB карту:**
```bash
aplay -l
```

2. **Проверить громкость:**
```bash
alsamixer
# Стрелки для навигации, ESC для выхода
```

3. **Проверить устройство в приложении:**
- Откройте http://localhost:8080
- Выберите USB карту в "🎧 Аудиокарта"

### Ошибка "Устройство не найдено"

Указано неправильное устройство ALSA. Используйте:
```bash
# Найти правильный номер карты
aplay -l | grep USB
# Затем используйте hw:X,0 (где X - номер карты)
```

## 🐳 Docker

### Собрать образ
```bash
docker build -t radioKeenetik .
```

### Запустить контейнер
```bash
docker run -d \
  -p 8080:8080 \
  --device /dev/snd \
  --name radioKeenetik \
  radioKeenetik
```

### Docker Compose
```bash
docker-compose up -d
```

## 📦 Установка на OpenWrt

### Способ 1: Скрипт установки
```bash
bash install.sh
```

### Способ 2: Вручную

1. **Загрузить бинарник:**
```bash
scp radioKeenetik-arm root@192.168.1.1:/opt/
ssh root@192.168.1.1
chmod +x /opt/radioKeenetik-arm
```

2. **Создать systemd сервис:**
```bash
# На роутере
cat > /etc/init.d/radioKeenetik << 'EOF'
#!/bin/sh /etc/rc.common
START=99
STOP=01

start() {
  /opt/radioKeenetik-arm > /var/log/radioKeenetik.log 2>&1 &
  echo $! > /var/run/radioKeenetik.pid
}

stop() {
  kill $(cat /var/run/radioKeenetik.pid) 2>/dev/null
}
EOF
chmod +x /etc/init.d/radioKeenetik
/etc/init.d/radioKeenetik enable
/etc/init.d/radioKeenetik start
```

## 🔐 Безопасность

Приложение слушает на `0.0.0.0:8080`. Для защиты используйте firewall:

```bash
# Разрешить только локально
sudo iptables -A INPUT -p tcp --dport 8080 -s 127.0.0.1 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 8080 -j DROP

# Или разрешить только из сети
sudo iptables -A INPUT -p tcp --dport 8080 -s 192.168.1.0/24 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 8080 -j DROP
```

## 📊 Логирование

Логи выводятся в консоль. Для сохранения:

```bash
# На Ubuntu
./radioKeenetik > /var/log/radioKeenetik.log 2>&1 &

# На роутере
/opt/radioKeenetik-arm > /var/log/radioKeenetik.log 2>&1 &
tail -f /var/log/radioKeenetik.log
```

## 🎯 Поддерживаемые плееры

| Плеер | Поддержка | Качество |
|-------|-----------|----------|
| **mpv** | ✅ MP3, AAC, HLS | Высокое |
| **ffplay** | ✅ MP3, AAC, HLS | Среднее |
| **mplayer** | ✅ MP3, AAC, HLS | Среднее |
| **sox play** | ⚠️ Простые потоки | Низкое |

**Рекомендуется:** mpv

## 📱 Совместимость браузеров

| Браузер | Веб | WebSocket | Совместимость |
|---------|-----|-----------|---------------|
| Chrome/Chromium | ✅ | ✅ | Полная |
| Firefox | ✅ | ✅ | Полная |
| Safari | ✅ | ✅ | Полная |
| Edge | ✅ | ✅ | Полная |
| Opera | ✅ | ✅ | Полная |
| Lynx | ⚠️ | ❌ | Частичная |

## 🌐 Совместимость роутеров

- ✅ OpenWrt
- ✅ DD-WRT
- ✅ Keenetic (KeeneticOS)
- ✅ Tomato
- ✅ LEDE
- ✅ Любой Linux ARM/x86

## 📚 Дополнительные ресурсы

- [mpv документация](https://mpv.io/manual/)
- [ALSA гайд](https://wiki.archlinux.org/title/ALSA)
- [OpenWrt вики](https://openwrt.org/)
- [Go документация](https://golang.org/doc/)

## 🤝 Благодарности

- [Gorilla Mux](https://github.com/gorilla/mux) - маршрутизация
- [Gorilla WebSocket](https://github.com/gorilla/websocket) - WebSocket
- [mpv](https://mpv.io/) - плеер

## 📄 Лицензия

MIT License - см. LICENSE.txt

## 📞 Поддержка

- GitHub Issues: [Сообщить о проблеме](https://github.com/yourusername/radioKeenetik/issues)
- GitHub Discussions: [Обсудить](https://github.com/yourusername/radioKeenetik/discussions)
- Email: support@radioKeenetik.local

---

**Версия:** 2.0.0 (Переработана архитектура)  
**Дата обновления:** 2024  
**Состояние:** ✅ Production Ready

## 🎉 Готовы к началу?

```bash
# Для Ubuntu
make quick-install
radioKeenetik

# Откройте http://localhost:8080

# Для роутера
make build-arm
scp radioKeenetik-arm root@192.168.1.1:/opt/
ssh root@192.168.1.1
/opt/radioKeenetik-arm
```

Наслаждайтесь радио! 🎵
