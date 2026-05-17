# 📂 Структура проекта radioKeenetik v2.0.0

## 🎯 Краткое описание файлов

### 🔵 Go исходные файлы (Обязательные)

#### `app.go` (50 строк) - ГЛАВНЫЙ ФАЙЛ
```
Содержит: main() функция
Ответственен за: Инициализация сервера, graceful shutdown
Изменяемость: Редко меняется
```

#### `server.go` (400 строк) - БИЗНЕС ЛОГИКА
```
Содержит: RadioServer, все HTTP handlers, WebSocket
Ответственен за: Управление станциями, маршруты
Изменяемость: Часто (при добавлении новых функций)
```

#### `audio.go` (300 строк) - РАБОТА С АУДИО
```
Содержит: AudioPlayer, управление плеером
Ответственен за: Запуск/остановка плеера, управление громкостью
Изменяемость: Редко
```

#### `main.go` (800+ строк) - АЛЬТЕРНАТИВНЫЙ МОНОЛИТ
```
Содержит: ВСЁ в одном файле (для совместимости)
Используйте если: Не нравится модульная структура
Примечание: Содержит встроенный HTML с исправленным JS
```

---

### 📦 Зависимости Go

#### `go.mod` (3 строки) - Файл зависимостей
```go
module radioKeenetik
go 1.21
require (
    github.com/gorilla/mux v1.8.0
    github.com/gorilla/websocket v1.5.0
)
```

#### `go.sum` (4 строки) - Хеши зависимостей
```
Автоматически генерируется
Требуется для проверки целостности
```

---

### 🛠️ Конфигурация и сборка

#### `Makefile` (200+ строк) - Команды сборки
```makefile
make help              # Справка
make quick-install     # Всё в одной команде
make build             # Собрать для Linux
make build-arm         # Собрать для ARM роутеров
make run               # Запустить
make install           # Установить в систему
```

#### `Dockerfile` (20 строк) - Контейнеризация
```
Используйте для: Docker/Kubernetes
Команда: docker build -t radioKeenetik .
```

#### `docker-compose.yml` (20 строк) - Docker Compose
```
Используйте для: Быстрый Docker запуск
Команда: docker-compose up -d
```

---

### 📚 Документация

#### `README.md` (300 строк) ⭐ ГЛАВНАЯ ДОКУМЕНТАЦИЯ
```
Содержит: Всё о проекте
Включает: Установка, использование, API, решение проблем
Читайте первым!
```

#### `QUICKSTART.md` (100 строк) - БЫСТРЫЙ СТАРТ
```
Содержит: За 5 минут запуск
Для кого: Нетерпеливых пользователей
Читайте если: Спешите
```

#### `CHANGES.md` (400 строк) - ЧТО ИЗМЕНИЛОСЬ
```
Содержит: Подробные изменения в v2.0.0
Важно: Почитайте что было исправлено!
```

#### `ROUTER_SETUP.md` (500 строк) - УСТАНОВКА НА РОУТЕР
```
Содержит: Подробное руководство для роутеров
Для кого: Пользователей OpenWrt/DD-WRT
Читайте если: Хотите запустить на роутере
```

---

### 🔧 Утилиты и примеры

#### `API_EXAMPLES.sh` (400 строк) - ПРИМЕРЫ API
```
Содержит: Примеры REST/WebSocket/Python/JavaScript
Используйте для: Тестирования и изучения API
Запуск: bash API_EXAMPLES.sh
```

#### `install.sh` (100 строк) - УСТАНОВКА НА OPENWRT
```
Используйте для: Автоматическая установка на OpenWrt
Запуск: bash install.sh
```

#### `push-to-github.sh` (150 строк) - ЗАГРУЗКА НА GITHUB
```
Используйте для: Загрузить проект на GitHub
Запуск: bash push-to-github.sh username
```

---

## 🚀 Как начать

### Способ 1️⃣ - Модульный (рекомендуется)
```bash
# Используйте файлы:
# ✓ app.go
# ✓ server.go
# ✓ audio.go
# ✓ go.mod
# ✓ go.sum

go build -o radioKeenetik app.go server.go audio.go
./radioKeenetik
```

### Способ 2️⃣ - Монолит (один файл)
```bash
# Используйте файл:
# ✓ main.go
# ✓ go.mod
# ✓ go.sum

go build -o radioKeenetik main.go
./radioKeenetik
```

### Способ 3️⃣ - С Makefile
```bash
# Используйте:
# ✓ Makefile
# + любые Go файлы

make build
./radioKeenetik
```

---

## 📋 Контрольный список перед запуском

### Обязательные файлы
- [x] `app.go` или `main.go`
- [x] `server.go` (если используете app.go)
- [x] `audio.go` (если используете app.go)
- [x] `go.mod`
- [x] `go.sum`

### Опциональные файлы
- [ ] `README.md` - для справки
- [ ] `QUICKSTART.md` - для быстрого старта
- [ ] `Makefile` - для удобства
- [ ] `API_EXAMPLES.sh` - для тестирования

### Зависимости системы
- [ ] Go 1.21+
- [ ] mpv (или ffplay/mplayer)
- [ ] git (опционально)

---

## 🔄 Типовые сценарии использования

### Сценарий 1: Быстрый запуск на Ubuntu
```bash
# Скопируйте эти файлы:
# app.go, server.go, audio.go, go.mod, go.sum, Makefile

make quick-install
radioKeenetik
# Откройте: http://localhost:8080
```

### Сценарий 2: Запуск на роутере (ARM)
```bash
# Скопируйте:
# app.go, server.go, audio.go, go.mod, go.sum, Makefile

make build-arm
# Загрузите radioKeenetik-arm на роутер
scp radioKeenetik-arm root@192.168.1.1:/opt/
```

### Сценарий 3: Docker контейнер
```bash
# Скопируйте:
# ВСЕ Go файлы, Dockerfile, docker-compose.yml

docker-compose up -d
# Откройте: http://localhost:8080
```

### Сценарий 4: Модификация для себя
```bash
# Скопируйте:
# app.go, server.go, audio.go (разделенный код легче менять)
# Редактируйте нужные модули
# Проверьте в README.md как что работает
```

---

## 🔍 Где найти что

### Хочу понять как работает
→ Читайте `README.md` + комментарии в коде

### Хочу быстро запустить
→ Читайте `QUICKSTART.md`

### Хочу запустить на роутере
→ Читайте `ROUTER_SETUP.md`

### Хочу использовать API
→ Запустите `bash API_EXAMPLES.sh`

### Хочу выложить на GitHub
→ Используйте `push-to-github.sh`

### Хочу установить на OpenWrt
→ Используйте `install.sh`

### Хочу что-то изменить
→ Разбирайтесь с модулями в `server.go` и `audio.go`

---

## 📊 Размеры файлов

```
app.go              50 строк      (инициализация)
server.go          400 строк      (бизнес-логика) ⭐
audio.go           300 строк      (аудио)
main.go            800+ строк     (альтернативный монолит)
go.mod              3 строки       (зависимости)
go.sum              4 строки       (хеши)
README.md          300 строк      (документация) ⭐
Makefile           200 строк      (сборка)
Dockerfile         20 строк       (контейнер)
docker-compose.yml 20 строк       (docker)
QUICKSTART.md      100 строк      (быстрый старт)
CHANGES.md         400 строк      (что изменилось)
ROUTER_SETUP.md    500 строк      (роутер)
API_EXAMPLES.sh    400 строк      (примеры)
```

---

## 🎯 Минимальный набор для запуска

### Минимум (монолит):
```
main.go
go.mod
go.sum
```

### Минимум (модули):
```
app.go
server.go
audio.go
go.mod
go.sum
```

### Рекомендуемый набор:
```
app.go
server.go
audio.go
go.mod
go.sum
Makefile
README.md
QUICKSTART.md
```

### Полный набор:
```
(Всё вышеперечисленное) +
CHANGES.md
API_EXAMPLES.sh
Dockerfile
docker-compose.yml
install.sh
push-to-github.sh
ROUTER_SETUP.md
```

---

## ✅ Проверка перед запуском

```bash
# 1. Проверить наличие необходимых файлов
ls -la app.go server.go audio.go go.mod go.sum

# 2. Проверить версию Go
go version  # Должна быть 1.21+

# 3. Скачать зависимости
go mod download

# 4. Собрать
go build -o radioKeenetik app.go server.go audio.go

# 5. Проверить результат
ls -lh radioKeenetik

# 6. Запустить
./radioKeenetik

# 7. Проверить в браузере
# http://localhost:8080
```

---

## 🆘 Если что-то не так

### Ошибка "No such file"
```bash
# Проверьте что все файлы на месте
ls -la *.go
```

### Ошибка "cannot find module"
```bash
# Скачайте зависимости
go mod download
go mod tidy
```

### Приложение не запускается
```bash
# Проверьте плеер
which mpv
sudo apt install -y mpv
```

### Порт занят
```bash
# Найдите что занимает порт 8080
lsof -i :8080

# Или используйте другой порт (редактируйте app.go)
```

---

## 📞 Поддержка

- 📖 Прочитайте README.md
- 🚀 Попробуйте QUICKSTART.md
- 🔧 Смотрите API_EXAMPLES.sh
- 📝 Напишите GitHub Issue

---

**Версия:** 2.0.0  
**Состояние:** ✅ Production Ready  
**Дата:** 2024

Удачи! 🎵
