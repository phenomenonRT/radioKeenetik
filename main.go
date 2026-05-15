package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// Station структура для хранения информации о радиостанции
type Station struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	URL       string `json:"url"`
	Genre     string `json:"genre"`
	Country   string `json:"country"`
	LogoURL   string `json:"logoUrl,omitempty"`
	IsPlaying bool   `json:"isPlaying"`
}

// PlaybackState состояние воспроизведения
type PlaybackState struct {
	Status   string    `json:"status"`   // playing, stopped, buffering, error
	Station  *Station  `json:"station,omitempty"`
	Volume   int       `json:"volume"`   // 0-100
	Device   string    `json:"device"`   // Текущее устройство
	ErrorMsg string    `json:"errorMsg,omitempty"`
	Bitrate  string    `json:"bitrate,omitempty"`
}

// AudioDevice информация об аудиокарте
type AudioDevice struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// AudioPlayer плеер для воспроизведения аудио
type AudioPlayer struct {
	cmd       *exec.Cmd
	process   *os.Process
	ctx       context.Context
	cancel    context.CancelFunc
	station   *Station
	state     PlaybackState
	mu        sync.RWMutex
	device    string // Текущее устройство (hw:1,0 или default)
	volume    int    // 0-100
	player    string // mpv, ffmpeg, mplayer, ffplay
}

// RadioServer основной сервер радио
type RadioServer struct {
	stations   map[string]*Station
	mu         sync.RWMutex
	current    *Station
	wsClients  map[*websocket.Conn]bool
	clientsMu  sync.RWMutex
	broadcast  chan interface{}
	player     *AudioPlayer
}

// NewAudioPlayer создает новый плеер
func NewAudioPlayer() *AudioPlayer {
	ctx, cancel := context.WithCancel(context.Background())
	
	// Определяем доступный плеер
	playerBinary := ""
	for _, p := range []string{"mpv", "ffplay", "mplayer", "play"} {
		if path, err := exec.LookPath(p); err == nil {
			playerBinary = path
			log.Printf("✓ Найден плеер: %s", p)
			break
		}
	}

	if playerBinary == "" {
		log.Println("⚠️  ВНИМАНИЕ: Не найден ни один плеер!")
		log.Println("   Установите один из: mpv, ffplay, mplayer, sox")
		log.Println("   Ubuntu: sudo apt install -y mpv")
	}

	return &AudioPlayer{
		device: "default",
		volume: 100,
		player: playerBinary,
		ctx:    ctx,
		cancel: cancel,
		state: PlaybackState{
			Status: "stopped",
			Volume: 100,
			Device: "default",
		},
	}
}

// Play запускает воспроизведение потока
func (ap *AudioPlayer) Play(url string, station *Station) error {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	// Останавливаем предыдущее воспроизведение
	if ap.cmd != nil && ap.cmd.Process != nil {
		ap.cmd.Process.Kill()
		ap.cmd.Wait()
	}

	ap.state.Status = "buffering"
	ap.state.Station = station
	ap.station = station

	// Получаем имя плеера из пути
	parts := strings.Split(ap.player, "/")
	playerName := parts[len(parts)-1]

	var cmd *exec.Cmd

	switch playerName {
	case "mpv":
		// mpv для всех форматов (AAC, MP3, HLS)
		args := []string{
			"--no-terminal",
			"--really-quiet",
			fmt.Sprintf("--volume=%d", ap.volume),
		}

		// Если устройство не default, используем ALSA
		if ap.device != "default" {
			args = append(args, fmt.Sprintf("--ao=alsa:device=%s", ap.device))
		}

		args = append(args, url)
		cmd = exec.CommandContext(ap.ctx, ap.player, args...)
		log.Printf("▶ Запуск mpv: %s", url)

	case "ffplay":
		// ffplay для MP3, AAC
		args := []string{
			"-nodisp",
			"-autoexit",
			"-loglevel", "quiet",
			url,
		}
		cmd = exec.CommandContext(ap.ctx, ap.player, args...)
		log.Printf("▶ Запуск ffplay: %s", url)

	case "mplayer":
		// mplayer для MP3, AAC, HLS
		args := []string{
			"-really-quiet",
			"-msglevel", "all=0",
			fmt.Sprintf("-volume=%d", ap.volume),
		}

		if ap.device != "default" {
			args = append(args, "-ao", fmt.Sprintf("alsa:device=%s", ap.device))
		}

		args = append(args, url)
		cmd = exec.CommandContext(ap.ctx, ap.player, args...)
		log.Printf("▶ Запуск mplayer: %s", url)

	case "play":
		// sox play для простых потоков
		args := []string{
			"-q",
			"-d",
			url,
		}
		cmd = exec.CommandContext(ap.ctx, ap.player, args...)
		log.Printf("▶ Запуск sox play: %s", url)

	default:
		return fmt.Errorf("неизвестный плеер: %s", playerName)
	}

	// Перенаправляем вывод в логи
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	ap.cmd = cmd
	ap.state.Status = "playing"

	// Запускаем в отдельной goroutine
	go func() {
		if err := cmd.Run(); err != nil {
			log.Printf("Ошибка воспроизведения: %v", err)
			ap.mu.Lock()
			ap.state.Status = "stopped"
			ap.state.ErrorMsg = err.Error()
			ap.mu.Unlock()
		} else {
			ap.mu.Lock()
			ap.state.Status = "stopped"
			ap.state.Station = nil
			ap.mu.Unlock()
		}
	}()

	return nil
}

// Stop останавливает воспроизведение
func (ap *AudioPlayer) Stop() error {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	if ap.cmd != nil && ap.cmd.Process != nil {
		ap.cmd.Process.Kill()
		ap.cmd.Wait()
	}

	ap.state.Status = "stopped"
	ap.state.Station = nil
	ap.station = nil
	ap.state.ErrorMsg = ""

	return nil
}

// SetVolume устанавливает громкость
func (ap *AudioPlayer) SetVolume(vol int) {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	if vol < 0 {
		vol = 0
	}
	if vol > 100 {
		vol = 100
	}

	ap.volume = vol
	ap.state.Volume = vol

	// Применяем громкость через amixer если возможно
	if ap.device != "default" {
		// Извлекаем номер карты из hw:X,Y
		parts := strings.Split(ap.device, ":")
		if len(parts) > 1 {
			cardNum := strings.Split(parts[1], ",")[0]
			exec.Command("amixer", "-c", cardNum, "set", "Master", fmt.Sprintf("%d%%", vol)).Run()
		}
	} else {
		// Для default используем pactl
		exec.Command("pactl", "set-sink-volume", "0", fmt.Sprintf("%d%%", vol)).Run()
	}
}

// GetState возвращает состояние плеера
func (ap *AudioPlayer) GetState() PlaybackState {
	ap.mu.RLock()
	defer ap.mu.RUnlock()
	return ap.state
}

// SetDevice устанавливает устройство воспроизведения
func (ap *AudioPlayer) SetDevice(device string) {
	ap.mu.Lock()
	ap.device = device
	ap.state.Device = device
	ap.mu.Unlock()
}

// GetAudioDevices возвращает список USB аудиокарт
func GetAudioDevices() []AudioDevice {
	devices := []AudioDevice{
		{ID: "default", Name: "По умолчанию"},
	}

	// Пытаемся найти USB устройства через aplay
	cmd := exec.Command("aplay", "-l")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "card") {
				// Парсим строку: card 1: USB [USB Audio Device], device 0: USB Audio [USB Audio]
				if parts := strings.Split(line, "card"); len(parts) > 1 {
					cardInfo := strings.Split(parts[1], ":")[0]
					cardNum := strings.TrimSpace(cardInfo)
					
					// Если это USB устройство, добавляем его
					if strings.Contains(line, "USB") {
						// Извлекаем полное имя
						nameParts := strings.Split(line, "[")
						if len(nameParts) > 1 {
							name := strings.TrimRight(nameParts[len(nameParts)-1], "]")
							devices = append(devices, AudioDevice{
								ID:   fmt.Sprintf("hw:%s,0", cardNum),
								Name: fmt.Sprintf("USB - %s", name),
							})
						}
					}
				}
			}
		}
	}

	return devices
}

// NewRadioServer создает новый сервер
func NewRadioServer() *RadioServer {
	return &RadioServer{
		stations:  make(map[string]*Station),
		wsClients: make(map[*websocket.Conn]bool),
		broadcast: make(chan interface{}, 10),
		player:    NewAudioPlayer(),
	}
}

// AddStation добавляет станцию
func (rs *RadioServer) AddStation(station *Station) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.stations[station.ID] = station
	rs.notifyClients()
}

// GetStations возвращает все станции
func (rs *RadioServer) GetStations() []*Station {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	stations := make([]*Station, 0, len(rs.stations))
	for _, s := range rs.stations {
		stations = append(stations, s)
	}
	return stations
}

// PlayStation запускает воспроизведение станции
func (rs *RadioServer) PlayStation(stationID string) error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	// Остановить текущее воспроизведение
	if rs.current != nil {
		rs.current.IsPlaying = false
		rs.player.Stop()
	}

	// Найти станцию
	if station, exists := rs.stations[stationID]; exists {
		station.IsPlaying = true
		rs.current = station

		// Запустить воспроизведение на аудиокарте
		if err := rs.player.Play(station.URL, station); err != nil {
			station.IsPlaying = false
			rs.current = nil
			rs.notifyClients()
			return err
		}

		rs.notifyClients()
		return nil
	}

	return fmt.Errorf("станция не найдена: %s", stationID)
}

// StopPlayback останавливает воспроизведение
func (rs *RadioServer) StopPlayback() {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if rs.current != nil {
		rs.current.IsPlaying = false
		rs.current = nil
	}

	rs.player.Stop()
	rs.notifyClients()
}

// GetCurrentStation возвращает текущую станцию
func (rs *RadioServer) GetCurrentStation() *Station {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.current
}

// notifyClients уведомляет клиентов об обновлении
func (rs *RadioServer) notifyClients() {
	rs.clientsMu.RLock()
	defer rs.clientsMu.RUnlock()

	data := map[string]interface{}{
		"type":        "update",
		"current":     rs.current,
		"stations":    rs.GetStations(),
		"playback":    rs.player.GetState(),
	}

	select {
	case rs.broadcast <- data:
	default:
	}
}

// HandleWebSocket обрабатывает WebSocket соединения
func (rs *RadioServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket ошибка: %v", err)
		return
	}
	defer conn.Close()

	rs.clientsMu.Lock()
	rs.wsClients[conn] = true
	rs.clientsMu.Unlock()

	// Отправить начальное состояние
	rs.mu.RLock()
	initialData := map[string]interface{}{
		"type":     "init",
		"current":  rs.current,
		"stations": rs.GetStations(),
		"playback": rs.player.GetState(),
	}
	rs.mu.RUnlock()
	conn.WriteJSON(initialData)

	// Слушать сообщения от клиента
	go func() {
		defer func() {
			rs.clientsMu.Lock()
			delete(rs.wsClients, conn)
			rs.clientsMu.Unlock()
		}()

		for {
			var msg map[string]interface{}
			if err := conn.ReadJSON(&msg); err != nil {
				return
			}
			rs.handleClientMessage(msg)
		}
	}()

	// Отправлять обновления клиенту
	for update := range rs.broadcast {
		if err := conn.WriteJSON(update); err != nil {
			return
		}
	}
}

// handleClientMessage обрабатывает сообщения от клиента
func (rs *RadioServer) handleClientMessage(msg map[string]interface{}) {
	action, ok := msg["action"].(string)
	if !ok {
		return
	}

	switch action {
	case "play":
		if stationID, ok := msg["stationId"].(string); ok {
			if err := rs.PlayStation(stationID); err != nil {
				log.Printf("Ошибка воспроизведения: %v", err)
			}
		}
	case "stop":
		rs.StopPlayback()
	case "setVolume":
		if vol, ok := msg["volume"].(float64); ok {
			rs.player.SetVolume(int(vol))
			rs.notifyClients()
		}
	case "setDevice":
		if device, ok := msg["device"].(string); ok {
			rs.player.SetDevice(device)
			rs.notifyClients()
		}
	case "addStation":
		if data, ok := msg["station"].(map[string]interface{}); ok {
			station := &Station{
				ID:      fmt.Sprintf("%v", data["id"]),
				Name:    fmt.Sprintf("%v", data["name"]),
				URL:     fmt.Sprintf("%v", data["url"]),
				Genre:   fmt.Sprintf("%v", data["genre"]),
				Country: fmt.Sprintf("%v", data["country"]),
			}
			rs.AddStation(station)
		}
	}
}

// HTTP Handlers

func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, getHTMLInterface())
}

func (rs *RadioServer) handleStations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rs.GetStations())
}

func (rs *RadioServer) handleCurrent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rs.GetCurrentStation())
}

func (rs *RadioServer) handlePlayback(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rs.player.GetState())
}

func (rs *RadioServer) handlePlay(w http.ResponseWriter, r *http.Request) {
	stationID := r.URL.Query().Get("id")
	if stationID == "" {
		http.Error(w, "ID станции требуется", http.StatusBadRequest)
		return
	}

	if err := rs.PlayStation(stationID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "playing"})
}

func (rs *RadioServer) handleStop(w http.ResponseWriter, r *http.Request) {
	rs.StopPlayback()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
}

func handleAudioDevices(w http.ResponseWriter, r *http.Request) {
	devices := GetAudioDevices()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"devices": devices,
	})
}

func getHTMLInterface() string {
	return `<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>radioKeenetik - USB Radio</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            justify-content: center;
            align-items: center;
            padding: 10px;
        }

        .container {
            background: white;
            border-radius: 20px;
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
            width: 100%;
            max-width: 700px;
            padding: 40px;
        }

        .header {
            text-align: center;
            margin-bottom: 30px;
        }

        .header h1 {
            font-size: 32px;
            margin-bottom: 10px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
        }

        .header p {
            color: #999;
            font-size: 14px;
        }

        .player {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            border-radius: 15px;
            padding: 30px;
            color: white;
            text-align: center;
            margin-bottom: 30px;
        }

        .now-playing {
            font-size: 14px;
            opacity: 0.8;
            margin-bottom: 10px;
        }

        .station-name {
            font-size: 24px;
            font-weight: bold;
            margin-bottom: 20px;
            min-height: 30px;
            word-break: break-word;
        }

        .station-info {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 15px;
            margin-bottom: 20px;
            font-size: 12px;
        }

        .info-item {
            background: rgba(255, 255, 255, 0.1);
            padding: 10px;
            border-radius: 8px;
        }

        .volume-control {
            margin: 20px 0;
        }

        .volume-label {
            display: flex;
            justify-content: space-between;
            font-size: 12px;
            margin-bottom: 8px;
        }

        .volume-slider {
            width: 100%;
            height: 6px;
            border-radius: 3px;
            background: rgba(255, 255, 255, 0.2);
            outline: none;
            -webkit-appearance: none;
            appearance: none;
        }

        .volume-slider::-webkit-slider-thumb {
            -webkit-appearance: none;
            appearance: none;
            width: 16px;
            height: 16px;
            border-radius: 50%;
            background: white;
            cursor: pointer;
        }

        .volume-slider::-moz-range-thumb {
            width: 16px;
            height: 16px;
            border-radius: 50%;
            background: white;
            cursor: pointer;
            border: none;
        }

        .devices-section {
            background: rgba(255, 255, 255, 0.2);
            border-radius: 10px;
            padding: 12px;
            margin-bottom: 20px;
            font-size: 12px;
        }

        .device-select {
            width: 100%;
            padding: 8px;
            border: none;
            border-radius: 5px;
            font-size: 12px;
            margin-top: 5px;
            background: white;
            color: #333;
        }

        .controls {
            display: flex;
            gap: 10px;
            justify-content: center;
            flex-wrap: wrap;
        }

        button {
            background: white;
            color: #667eea;
            border: none;
            padding: 12px 24px;
            border-radius: 25px;
            font-weight: bold;
            cursor: pointer;
            transition: all 0.3s ease;
            font-size: 14px;
        }

        button:hover {
            transform: translateY(-2px);
            box-shadow: 0 5px 15px rgba(0, 0, 0, 0.2);
        }

        button:active {
            transform: translateY(0);
        }

        .stations-section {
            margin-bottom: 30px;
        }

        .stations-section h2 {
            font-size: 18px;
            margin-bottom: 15px;
            color: #333;
        }

        .stations-list {
            display: grid;
            gap: 10px;
            max-height: 400px;
            overflow-y: auto;
        }

        .station-item {
            background: #f5f5f5;
            border: 2px solid #e0e0e0;
            border-radius: 10px;
            padding: 15px;
            cursor: pointer;
            transition: all 0.3s ease;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }

        .station-item:hover {
            background: #efefef;
            border-color: #667eea;
            box-shadow: 0 3px 10px rgba(102, 126, 234, 0.2);
        }

        .station-item.active {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border-color: transparent;
        }

        .station-details {
            flex: 1;
        }

        .station-name-item {
            font-weight: bold;
            font-size: 14px;
            margin-bottom: 3px;
        }

        .station-genre {
            font-size: 12px;
            opacity: 0.7;
        }

        .play-btn {
            background: white;
            color: #667eea;
            border: none;
            padding: 8px 16px;
            border-radius: 20px;
            cursor: pointer;
            font-weight: bold;
            font-size: 12px;
        }

        .add-station-form {
            background: #f9f9f9;
            padding: 20px;
            border-radius: 10px;
        }

        .form-group {
            margin-bottom: 12px;
        }

        .form-group label {
            display: block;
            font-size: 12px;
            font-weight: bold;
            margin-bottom: 4px;
            color: #333;
        }

        .form-group input {
            width: 100%;
            padding: 8px 12px;
            border: 1px solid #ddd;
            border-radius: 5px;
            font-size: 12px;
        }

        .form-group input:focus {
            outline: none;
            border-color: #667eea;
            box-shadow: 0 0 5px rgba(102, 126, 234, 0.2);
        }

        .add-btn {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            width: 100%;
            padding: 10px;
        }

        .status {
            padding: 12px;
            border-radius: 5px;
            margin-bottom: 15px;
            font-size: 12px;
            text-align: center;
            display: none;
        }

        .status.show {
            display: block;
        }

        .status.success {
            background: #d4edda;
            color: #155724;
        }

        .status.error {
            background: #f8d7da;
            color: #721c24;
        }

        .status.info {
            background: #d1ecf1;
            color: #0c5460;
        }

        @media (max-width: 600px) {
            .container {
                padding: 20px;
            }

            .header h1 {
                font-size: 24px;
            }

            .player {
                padding: 20px;
            }

            .station-info {
                grid-template-columns: 1fr;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎙️ radioKeenetik</h1>
            <p>Интернет-радио на USB аудиокарте роутера</p>
        </div>

        <div id="status" class="status"></div>

        <div class="player">
            <div class="now-playing">▶ Сейчас играет</div>
            <div class="station-name" id="currentStationName">—</div>
            <div class="station-info">
                <div class="info-item">
                    <div>🎵 Жанр: <span id="currentGenre">—</span></div>
                </div>
                <div class="info-item">
                    <div>🌍 Страна: <span id="currentCountry">—</span></div>
                </div>
            </div>

            <div class="volume-control">
                <div class="volume-label">
                    <span>🔊 Громкость</span>
                    <span id="volumeValue">100%</span>
                </div>
                <input type="range" id="volumeSlider" class="volume-slider" min="0" max="100" value="100">
            </div>

            <div class="devices-section">
                <div>🎧 USB Аудиокарта</div>
                <select id="deviceSelect" class="device-select">
                    <option value="default">По умолчанию</option>
                </select>
            </div>

            <div class="controls">
                <button onclick="stopPlayback()">⏹ Стоп</button>
            </div>
        </div>

        <div class="stations-section">
            <h2>📻 Радиостанции</h2>
            <div class="stations-list" id="stationsList"></div>
        </div>

        <div class="add-station-form">
            <h3 style="font-size: 16px; margin-bottom: 15px; color: #333;">➕ Добавить станцию</h3>
            <div class="form-group">
                <label>Название</label>
                <input type="text" id="stationName" placeholder="Название">
            </div>
            <div class="form-group">
                <label>URL потока (MP3, AAC, HLS)</label>
                <input type="text" id="stationUrl" placeholder="https://...">
            </div>
            <div class="form-group">
                <label>Жанр</label>
                <input type="text" id="stationGenre" placeholder="Поп, Рок, Электро...">
            </div>
            <div class="form-group">
                <label>Страна</label>
                <input type="text" id="stationCountry" placeholder="Россия, США...">
            </div>
            <button class="add-btn" onclick="addStation()">Добавить</button>
        </div>
    </div>

    <script>
        const ws = new WebSocket('ws://' + window.location.host + '/ws');
        let stations = [];
        let currentStation = null;

        ws.onopen = function() {
            console.log('✓ Подключено к серверу');
            loadAudioDevices();
        };

        ws.onmessage = function(event) {
            const data = JSON.parse(event.data);
            
            if (data.type === 'update' || data.type === 'init') {
                stations = data.stations || [];
                currentStation = data.current;
                updateUI();
            }
        };

        ws.onerror = function(error) {
            showStatus('Ошибка подключения к серверу', 'error');
            console.error('WebSocket ошибка:', error);
        };

        ws.onclose = function() {
            showStatus('Соединение потеряно', 'error');
        };

        function loadAudioDevices() {
            fetch('/api/audio-devices')
                .then(r => r.json())
                .then(data => {
                    const select = document.getElementById('deviceSelect');
                    if (data.devices && data.devices.length > 1) {
                        data.devices.forEach((device, idx) => {
                            if (idx > 0) {
                                const option = document.createElement('option');
                                option.value = device.id;
                                option.text = device.name;
                                select.appendChild(option);
                            }
                        });
                        showStatus('Найдено USB карт: ' + (data.devices.length - 1), 'info');
                    }
                })
                .catch(e => {
                    console.error('Ошибка загрузки устройств:', e);
                    showStatus('Ошибка загрузки устройств', 'error');
                });
        }

        function updateUI() {
            const nameEl = document.getElementById('currentStationName');
            const genreEl = document.getElementById('currentGenre');
            const countryEl = document.getElementById('currentCountry');

            if (currentStation) {
                nameEl.textContent = currentStation.name;
                genreEl.textContent = currentStation.genre || '—';
                countryEl.textContent = currentStation.country || '—';
            } else {
                nameEl.textContent = '—';
                genreEl.textContent = '—';
                countryEl.textContent = '—';
            }

            const listEl = document.getElementById('stationsList');
            listEl.innerHTML = '';

            stations.forEach(station => {
                const isActive = currentStation && currentStation.id === station.id;
                const div = document.createElement('div');
                div.className = 'station-item' + (isActive ? ' active' : '');
                div.innerHTML = \`
                    <div class="station-details">
                        <div class="station-name-item">\${station.name}</div>
                        <div class="station-genre">\${station.genre} • \${station.country}</div>
                    </div>
                    <button class="play-btn" onclick="playStation('\${station.id}')">
                        \${isActive ? '▶ Играет' : '▶ Играть'}
                    </button>
                \`;
                listEl.appendChild(div);
            });
        }

        document.getElementById('volumeSlider').addEventListener('input', function() {
            const volume = parseInt(this.value);
            document.getElementById('volumeValue').textContent = volume + '%';
            
            ws.send(JSON.stringify({
                action: 'setVolume',
                volume: volume
            }));
        });

        document.getElementById('deviceSelect').addEventListener('change', function() {
            ws.send(JSON.stringify({
                action: 'setDevice',
                device: this.value
            }));
            showStatus('Устройство: ' + this.options[this.selectedIndex].text, 'info');
        });

        function playStation(stationId) {
            ws.send(JSON.stringify({
                action: 'play',
                stationId: stationId
            }));
        }

        function stopPlayback() {
            ws.send(JSON.stringify({
                action: 'stop'
            }));
        }

        function addStation() {
            const name = document.getElementById('stationName').value.trim();
            const url = document.getElementById('stationUrl').value.trim();
            const genre = document.getElementById('stationGenre').value.trim();
            const country = document.getElementById('stationCountry').value.trim();

            if (!name || !url || !genre || !country) {
                showStatus('⚠️ Заполните все поля', 'error');
                return;
            }

            if (!url.startsWith('http')) {
                showStatus('⚠️ URL должен начинаться с http://', 'error');
                return;
            }

            ws.send(JSON.stringify({
                action: 'addStation',
                station: {
                    id: Date.now().toString(),
                    name: name,
                    url: url,
                    genre: genre,
                    country: country
                }
            }));

            document.getElementById('stationName').value = '';
            document.getElementById('stationUrl').value = '';
            document.getElementById('stationGenre').value = '';
            document.getElementById('stationCountry').value = '';
            
            showStatus('✓ Станция добавлена', 'success');
        }

        function showStatus(message, type) {
            const statusEl = document.getElementById('status');
            statusEl.textContent = message;
            statusEl.className = 'status show ' + type;
            
            setTimeout(() => {
                statusEl.classList.remove('show');
            }, 4000);
        }

        updateUI();
    </script>
</body>
</html>`
}

func main() {
	log.Println("🎙️  radioKeenetik - USB Audio Radio")
	log.Println("====================================")

	// Создать сервер
	radioServer := NewRadioServer()

	log.Printf("✓ Плеер: %s", radioServer.player.player)
	devices := GetAudioDevices()
	log.Printf("✓ Найдено USB устройств: %d", len(devices)-1)

	// Добавить примеры станций
	radioServer.AddStation(&Station{
		ID:      "1",
		Name:    "Radio Record - Live DJ Sets",
		URL:     "https://hls-01-radiorecord.hostingradio.ru/record-livedjsets/112/l0_6a054e73b4592b18390d2063.aac",
		Genre:   "Electronic",
		Country: "Russia",
	})

	radioServer.AddStation(&Station{
		ID:      "2",
		Name:    "Rock Forever",
		URL:     "https://stream.example.com/rock.mp3",
		Genre:   "Rock",
		Country: "USA",
	})

	radioServer.AddStation(&Station{
		ID:      "3",
		Name:    "Jazz Paradise",
		URL:     "https://stream.example.com/jazz.aac",
		Genre:   "Jazz",
		Country: "France",
	})

	// Настроить маршруты
	r := mux.NewRouter()

	// Веб-интерфейс
	r.HandleFunc("/", handleIndex).Methods("GET")

	// REST API
	r.HandleFunc("/api/stations", radioServer.handleStations).Methods("GET")
	r.HandleFunc("/api/current", radioServer.handleCurrent).Methods("GET")
	r.HandleFunc("/api/playback", radioServer.handlePlayback).Methods("GET")
	r.HandleFunc("/api/audio-devices", handleAudioDevices).Methods("GET")
	r.HandleFunc("/api/play", radioServer.handlePlay).Methods("POST")
	r.HandleFunc("/api/stop", radioServer.handleStop).Methods("POST")

	// WebSocket
	r.HandleFunc("/ws", radioServer.HandleWebSocket)

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("⏹ Выключение...")
		radioServer.StopPlayback()
		os.Exit(0)
	}()

	// Запустить сервер
	addr := ":8080"
	log.Printf("\n✓ Сервер запущен на http://0.0.0.0%s", addr)
	log.Printf("✓ Откройте в браузере: http://localhost%s\n", addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Ошибка сервера: %v", err)
	}
}
