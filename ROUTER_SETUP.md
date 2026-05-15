# 🎙️ radioKeenetik - Интернет-радио на USB аудиокарте роутера

## 📡 Описание

**radioKeenetik** - это приложение для воспроизведения интернет-радио прямо на USB аудиокарте вашего роутера.

Поддерживает:
- ✅ MP3 потоки (http://, https://)
- ✅ AAC/M4A потоки (например, HLS)
- ✅ Любые форматы, поддерживаемые mpv/ffplay
- ✅ WebSocket реал-тайм обновления
- ✅ Веб-интерфейс в браузере
- ✅ REST API

## 🛠️ Требования

### Для сборки на Ubuntu 24+:
```
Go 1.21+
git
build-essential
```

### Для роутера:
```
USB аудиокарта
Linux ОС (OpenWrt, DD-WRT и совместимые)
ARM или x86 процессор
~10-20MB свободной памяти
```

## 📥 Установка на Ubuntu 24

### Шаг 1: Установить зависимости системы

```bash
make system-deps
```

Или вручную:
```bash
sudo apt update
sudo apt install -y mpv ffmpeg sox alsa-utils pulseaudio
```

### Шаг 2: Клонировать и собрать

```bash
git clone https://github.com/yourusername/radioKeenetik.git
cd radioKeenetik
go mod download
make build
```

### Шаг 3: Запустить локально

```bash
./radioKeenetik
```

Откройте в браузере: **http://localhost:8080**

## 🔥 Быстрая установка (одна команда)

```bash
make quick-install
```

Затем:
```bash
radioKeenetik
```

## 🚀 Сборка для роутера

### Определить архитектуру роутера

```bash
# Подключитесь к роутеру по SSH
ssh root@192.168.1.1

# Проверьте архитектуру
uname -m
```

**Результаты:**
- `armv7l` → **ARM (самое распространенное)** → используйте `make build-arm`
- `armv6l` → **ARM старый** → используйте `make build-arm`
- `aarch64` → **ARM64 новый** → используйте `make build-arm64`
- `x86_64` → **x86/x64** → используйте `make build-x86`
- `mips` → **MIPS** → нужна пользовательская сборка

### Собрать для роутера

**Для большинства роутеров (ARM):**
```bash
make build-arm
# Результат: radioKeenetik-arm
```

**Для новых роутеров (ARM64):**
```bash
make build-arm64
# Результат: radioKeenetik-arm64
```

**Для x86 роутеров:**
```bash
make build-x86
# Результат: radioKeenetik-x86
```

**Для всех платформ сразу:**
```bash
make build-all
```

## 📤 Загрузка на роутер

### Способ 1: SCP (SSH Copy)

```bash
# Замените IP адрес на ваш роутер
scp radioKeenetik-arm root@192.168.1.1:/opt/

# Подключитесь к роутеру
ssh root@192.168.1.1

# Сделайте исполняемым и запустите
chmod +x /opt/radioKeenetik-arm
/opt/radioKeenetik-arm
```

Затем откройте: **http://192.168.1.1:8080**

### Способ 2: FTP/SFTP

Используйте FileZilla или другой FTP клиент:
1. Подключитесь к роутеру
2. Загрузите бинарник в `/opt/`
3. Установите права: `chmod +x radioKeenetik-arm`

### Способ 3: WebUI роутера (если поддерживается)

Загрузите файл через веб-интерфейс управления роутером.

## 🎛️ Проверка USB аудиокарты

Перед запуском проверьте наличие USB аудиокарты:

### На Ubuntu:
```bash
# Список всех карт
aplay -l

# Вывод: 
# **** List of PLAYBACK Hardware Devices ****
# card 0: Analog [Analog], device 0: ALC1220 [ALC1220 Analog]
# card 1: Device [USB Audio Device], device 0: USB Audio [USB Audio]
#                   ↑ Ваша USB карта

# Или через PulseAudio
pactl list short sinks
```

### На роутере (SSH):
```bash
ssh root@192.168.1.1
aplay -l
# или
cat /proc/asound/cards
```

## ▶️ Запуск приложения

### Локальный запуск (Ubuntu):

```bash
./radioKeenetik
# Или если установлено
radioKeenetik
# Откройте: http://localhost:8080
```

### На роутере (SSH):

```bash
ssh root@192.168.1.1
/opt/radioKeenetik-arm
# Откройте: http://192.168.1.1:8080
```

### Фоновый запуск на роутере:

```bash
/opt/radioKeenetik-arm &
# или с перенаправлением логов
/opt/radioKeenetik-arm > /var/log/radioKeenetik.log 2>&1 &
```

## 🌐 Веб-интерфейс

После запуска откройте в браузере:
- **Локально**: http://localhost:8080
- **На роутере**: http://192.168.1.1:8080 (или IP роутера)

### Функции:
- 📻 Список радиостанций
- ▶️ Запуск/остановка воспроизведения
- 🔊 Регулировка громкости
- 🎧 Выбор USB аудиокарты
- ➕ Добавление новых станций

## 📡 Примеры потоков

Вы можете добавить эти потоки:

### MP3:
```
https://hls-01-radiorecord.hostingradio.ru/record-livedjsets/112/l0_6a054e73b4592b18390d2063.aac
```

### AAC:
```
https://stream.example.com/aac.aac
```

### HLS:
```
https://example.com/playlist.m3u8
```

## 🔧 Команды Makefile

```bash
make help                # Справка
make system-deps         # Установить системные зависимости
make deps                # Скачать Go зависимости
make build               # Собрать для текущей ОС
make build-arm           # Собрать для ARM роутеров
make build-arm64         # Собрать для ARM64
make build-x86           # Собрать для x86
make build-all           # Собрать все
make run                 # Запустить локально
make test                # Тесты
make install             # Установить локально
make quick-install       # Все + установка (одна команда)
make check-player        # Проверить плееры
make check-audio         # Проверить аудиокарты
make systemd-install     # Установить как systemd сервис
```

## 🔊 Управление громкостью

### Способ 1: Веб-интерфейс
Используйте ползунок "🔊 Громкость" в веб-интерфейсе.

### Способ 2: Командная строка
```bash
# Linux системы
amixer -c 1 set Master 80%

# PulseAudio
pactl set-sink-volume 0 80%
```

### Способ 3: Через alsamixer
```bash
alsamixer
# Навигация: стрелки, F5 = выбор карты, ESC = выход
```

## 🎵 REST API

### Получить список станций
```bash
curl http://localhost:8080/api/stations | jq .
```

### Получить текущую станцию
```bash
curl http://localhost:8080/api/current | jq .
```

### Запустить станцию (ID=1)
```bash
curl -X POST http://localhost:8080/api/play?id=1
```

### Остановить
```bash
curl -X POST http://localhost:8080/api/stop
```

### Получить состояние воспроизведения
```bash
curl http://localhost:8080/api/playback | jq .
```

### Получить список USB аудиокарт
```bash
curl http://localhost:8080/api/audio-devices | jq .
```

## 🔌 WebSocket API

Подключение: `ws://localhost:8080/ws`

**Запустить станцию:**
```json
{
  "action": "play",
  "stationId": "1"
}
```

**Остановить:**
```json
{
  "action": "stop"
}
```

**Установить громкость (0-100):**
```json
{
  "action": "setVolume",
  "volume": 80
}
```

**Выбрать USB карту:**
```json
{
  "action": "setDevice",
  "device": "hw:1,0"
}
```

**Добавить станцию:**
```json
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
```

## 🆘 Решение проблем

### Приложение не запускается

**Проверить наличие плеера:**
```bash
make check-player
```

Установить mpv:
```bash
# Ubuntu 24
sudo apt install -y mpv

# На роутере (если поддерживается)
opkg install mpv
```

### Нет звука

1. **Проверить USB карту:**
```bash
aplay -l
```

2. **Проверить громкость:**
- Веб-интерфейс: переместите ползунок громкости
- Командная строка: `alsamixer`

3. **Проверить, работает ли плеер:**
```bash
mpv --ao=alsa:device=hw:1,0 https://example.com/stream.mp3
```

### Ошибка "устройство не найдено"

Указано неправильное устройство ALSA. Проверьте:
```bash
aplay -l
# Найдите номер вашей карты (например, card 1)
# Используйте: hw:1,0
```

### Приложение требует CPU

Это нормально при потоковой передаче. Используйте более мощный плеер:
- ✅ mpv (рекомендуется)
- ✅ ffplay
- ⚠️ mplayer (медленнее)

### Проблемы с форматом потока

Убедитесь, что поток доступен:
```bash
# Проверить с помощью mpv
mpv https://example.com/stream.mp3 --no-audio-display

# Или ffmpeg
ffmpeg -i https://example.com/stream.mp3 -f null -
```

## 🔄 Обновление

Просто пересоберите и загрузите новый бинарник:
```bash
git pull
make build-arm
scp radioKeenetik-arm root@192.168.1.1:/opt/
```

## 📊 Файлы конфигурации

Приложение не использует конфиг файлы - всё управляется через веб-интерфейс.

Станции хранятся в памяти и восстанавливаются при следующем запуске (вернутся примеры станций).

## 🔐 Безопасность

- Приложение слушает на `0.0.0.0:8080` (доступно из сети)
- WebSocket без аутентификации
- Используйте firewall или reverse proxy для защиты:

```bash
# Блокировать прямой доступ
sudo iptables -A INPUT -p tcp --dport 8080 -j DROP

# Или разрешить только локально
sudo iptables -A INPUT -p tcp --dport 8080 -s 192.168.1.0/24 -j ACCEPT
```

## 🐳 Docker (опционально)

Хотя рекомендуется собирать нативно, можно использовать Docker:
```bash
docker build -t radioKeenetik .
docker run -d \
  -p 8080:8080 \
  --device /dev/snd \
  -v /etc/asound.conf:/etc/asound.conf:ro \
  radioKeenetik
```

## 📦 Установка как systemd сервис (Ubuntu)

```bash
make systemd-install

# Управление
sudo systemctl start radioKeenetik
sudo systemctl stop radioKeenetik
sudo systemctl restart radioKeenetik
sudo systemctl status radioKeenetik
sudo journalctl -u radioKeenetik -f  # логи
```

## 📚 Дополнительно

- **mpv документация**: https://mpv.io/manual/
- **ALSA гайд**: https://wiki.archlinux.org/title/ALSA
- **OpenWrt**: https://openwrt.org/

## 🎙️ Примеры реальных потоков

Попробуйте добавить эти станции:

```
Название: Radio Record - Live DJ Sets
URL: https://hls-01-radiorecord.hostingradio.ru/record-livedjsets/112/l0_6a054e73b4592b18390d2063.aac
Жанр: Electronic
Страна: Russia
```

## 💡 Советы

1. **Для лучшего качества**: используйте mpv
2. **Для малых систем**: используйте mplayer или ffplay
3. **Проверяйте логи**: смотрите консоль для диагностики
4. **Сохраняйте станции**: список сохраняется при добавлении

## 📞 Поддержка

- GitHub Issues
- GitHub Discussions
- Email: support@radioKeenetik.local

---

**Готовы начать? Запустите:**

```bash
make quick-install
radioKeenetik
```

Затем откройте: **http://localhost:8080** 🎵

