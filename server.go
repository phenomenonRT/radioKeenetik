package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// ============================================================================
// STRUCTURES
// ============================================================================

// Station структура для хранения информации о радиостанции
type Station struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	URL       string `json:"url"`
	Genre     string `json:"genre"`
	Country   string `json:"country"`
	LogoURL   string `json:"logoUrl,omitempty"`
	IsPlaying bool   `json:"isPlaying"`
	IsFavorite bool   `json:"isFavorite"`
	AddedAt   int64  `json:"addedAt"`
}

// PlaybackState состояние воспроизведения
type PlaybackState struct {
	Status   string   `json:"status"` // playing, stopped, buffering, error
	Station  *Station `json:"station,omitempty"`
	Volume   int      `json:"volume"` // 0-100
	Device   string   `json:"device"` // Текущее устройство
	ErrorMsg string   `json:"errorMsg,omitempty"`
	Bitrate  string   `json:"bitrate,omitempty"`
}

// ============================================================================
// RADIO SERVER
// ============================================================================

// RadioServer основной сервер радио
type RadioServer struct {
	stations  map[string]*Station
	mu        sync.RWMutex
	current   *Station
	wsClients map[*websocket.Conn]bool
	clientsMu sync.RWMutex
	broadcast chan interface{}
	player    *AudioPlayer
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

// AddStation добавляет новую станцию в список
func (rs *RadioServer) AddStation(station *Station) {
	rs.mu.Lock()
	station.AddedAt = time.Now().Unix()
	rs.stations[station.ID] = station
	rs.mu.Unlock()
	
	// Уведомляем всех подключенных клиентов
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

		// Запустить воспроизведение
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
		"type":     "update",
		"current":  rs.current,
		"stations": rs.GetStations(),
		"playback": rs.player.GetState(),
	}

	select {
	case rs.broadcast <- data:
	default:
		// Канал переполнен, пропускаем обновление
	}
}

// ============================================================================
// HTTP HANDLERS
// ============================================================================

// handleIndex возвращает HTML интерфейс
func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := getHTMLInterface()
	fmt.Fprint(w, html)
}

// handleStations возвращает список станций
func (rs *RadioServer) handleStations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	json.NewEncoder(w).Encode(rs.GetStations())
}

// handleCurrent возвращает текущую станцию
func (rs *RadioServer) handleCurrent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	json.NewEncoder(w).Encode(rs.GetCurrentStation())
}

// handlePlayback возвращает состояние воспроизведения
func (rs *RadioServer) handlePlayback(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	json.NewEncoder(w).Encode(rs.player.GetState())
}

// handlePlay запускает воспроизведение станции
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

// handleStop останавливает воспроизведение
func (rs *RadioServer) handleStop(w http.ResponseWriter, r *http.Request) {
	rs.StopPlayback()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
}

// handleAudioDevices возвращает список аудиокарт
func handleAudioDevices(w http.ResponseWriter, r *http.Request) {
	devices := GetAudioDevices()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"devices": devices,
	})
}

// ============================================================================
// WEBSOCKET HANDLING
// ============================================================================

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Разрешаем все источники (для роутеров)
	},
}

// HandleWebSocket обрабатывает WebSocket соединения
func (rs *RadioServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
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
		"devices":  GetAudioDevices(),
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
			log.Printf("✓ Добавлена станция: %s", station.Name)
		}

	case "toggleFavorite":
		if stationID, ok := msg["stationId"].(string); ok {
			rs.mu.Lock()
			if station, exists := rs.stations[stationID]; exists {
				station.IsFavorite = !station.IsFavorite
			}
			rs.mu.Unlock()
			rs.notifyClients()
		}

	case "deleteStation":
		if stationID, ok := msg["stationId"].(string); ok {
			rs.mu.Lock()
			delete(rs.stations, stationID)
			rs.mu.Unlock()
			rs.notifyClients()
			log.Printf("✓ Удалена станция: %s", stationID)
		}
	}
}

// ============================================================================
// ROUTE REGISTRATION
// ============================================================================

// RegisterRoutes регистрирует все маршруты сервера
func (rs *RadioServer) RegisterRoutes(r *mux.Router) {
	// Веб-интерфейс
	r.HandleFunc("/", handleIndex).Methods("GET")

	// REST API
	r.HandleFunc("/api/stations", rs.handleStations).Methods("GET")
	r.HandleFunc("/api/current", rs.handleCurrent).Methods("GET")
	r.HandleFunc("/api/playback", rs.handlePlayback).Methods("GET")
	r.HandleFunc("/api/audio-devices", handleAudioDevices).Methods("GET")
	r.HandleFunc("/api/play", rs.handlePlay).Methods("POST")
	r.HandleFunc("/api/stop", rs.handleStop).Methods("POST")

	// WebSocket
	r.HandleFunc("/ws", rs.HandleWebSocket)
}

// ============================================================================
// HTML INTERFACE
// ============================================================================

// getHTMLInterface возвращает HTML интерфейс приложения
func getHTMLInterface() string {
	return `<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="theme-color" content="#1a1a2e">
    <title>radioKeenetik - Интернет-радио</title>
    <style>
        :root {
            --primary: #667eea;
            --primary-dark: #764ba2;
            --bg-dark: #0f0f1e;
            --bg-lighter: #1a1a2e;
            --text-primary: #ffffff;
            --text-secondary: #999999;
            --border-color: #2a2a3e;
            --success: #4caf50;
            --error: #f44336;
            --warning: #ff9800;
            --info: #2196f3;
        }

        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        html, body {
            width: 100%;
            height: 100%;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: var(--bg-dark);
            color: var(--text-primary);
            overflow: hidden;
        }

        .container {
            display: flex;
            flex-direction: column;
            height: 100vh;
            overflow: hidden;
        }

        .header {
            background: linear-gradient(135deg, var(--primary) 0%, var(--primary-dark) 100%);
            padding: 16px 20px;
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
            flex-shrink: 0;
        }

        .header-title {
            font-size: 24px;
            font-weight: 700;
            margin: 0;
            display: flex;
            align-items: center;
            gap: 10px;
        }

        .main-content {
            flex: 1;
            overflow-y: auto;
            padding: 20px;
            display: flex;
            flex-direction: column;
            gap: 20px;
        }

        .player-card {
            background: var(--bg-lighter);
            border: 1px solid var(--border-color);
            border-radius: 12px;
            padding: 20px;
            box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
        }

        .now-playing {
            text-align: center;
            margin-bottom: 16px;
        }

        .station-name {
            font-size: 20px;
            font-weight: 600;
            margin-bottom: 8px;
            min-height: 28px;
            word-break: break-word;
        }

        .station-meta {
            display: flex;
            justify-content: center;
            gap: 16px;
            font-size: 12px;
            color: var(--text-secondary);
            flex-wrap: wrap;
        }

        .meta-item {
            display: flex;
            align-items: center;
            gap: 6px;
        }

        .player-controls {
            display: flex;
            justify-content: center;
            gap: 12px;
            margin: 16px 0;
            flex-wrap: wrap;
        }

        .btn {
            background: var(--primary);
            color: white;
            border: none;
            padding: 10px 20px;
            border-radius: 8px;
            cursor: pointer;
            font-weight: 600;
            font-size: 14px;
            transition: all 0.3s ease;
            display: inline-flex;
            align-items: center;
            gap: 6px;
            white-space: nowrap;
        }

        .btn:hover {
            background: var(--primary-dark);
            transform: translateY(-2px);
            box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
        }

        .btn:active {
            transform: translateY(0);
        }

        .btn-secondary {
            background: var(--bg-lighter);
            color: var(--text-primary);
            border: 1px solid var(--border-color);
        }

        .btn-secondary:hover {
            background: #2a2a3e;
        }

        .btn-danger {
            background: var(--error);
        }

        .btn-danger:hover {
            background: #e53935;
        }

        .slider-container {
            margin: 16px 0;
        }

        .slider-label {
            display: flex;
            justify-content: space-between;
            font-size: 12px;
            margin-bottom: 8px;
            color: var(--text-secondary);
        }

        .slider {
            width: 100%;
            height: 6px;
            border-radius: 3px;
            background: var(--border-color);
            outline: none;
            -webkit-appearance: none;
            appearance: none;
        }

        .slider::-webkit-slider-thumb {
            -webkit-appearance: none;
            appearance: none;
            width: 16px;
            height: 16px;
            border-radius: 50%;
            background: var(--primary);
            cursor: pointer;
            box-shadow: 0 0 8px rgba(102, 126, 234, 0.3);
        }

        .slider::-moz-range-thumb {
            width: 16px;
            height: 16px;
            border-radius: 50%;
            background: var(--primary);
            cursor: pointer;
            border: none;
            box-shadow: 0 0 8px rgba(102, 126, 234, 0.3);
        }

        .device-selector {
            margin: 16px 0;
        }

        .device-selector select {
            width: 100%;
            padding: 8px 12px;
            background: var(--bg-dark);
            border: 1px solid var(--border-color);
            color: var(--text-primary);
            border-radius: 8px;
            cursor: pointer;
            font-size: 14px;
        }

        .device-selector select:focus {
            outline: none;
            border-color: var(--primary);
            box-shadow: 0 0 8px rgba(102, 126, 234, 0.2);
        }

        .section {
            display: flex;
            flex-direction: column;
            gap: 12px;
        }

        .section-title {
            font-size: 14px;
            font-weight: 600;
            color: var(--text-secondary);
            text-transform: uppercase;
            letter-spacing: 1px;
            margin-top: 8px;
        }

        .search-box {
            position: relative;
        }

        .search-box input {
            width: 100%;
            padding: 10px 16px;
            background: var(--bg-lighter);
            border: 1px solid var(--border-color);
            color: var(--text-primary);
            border-radius: 8px;
            font-size: 14px;
        }

        .search-box input::placeholder {
            color: var(--text-secondary);
        }

        .search-box input:focus {
            outline: none;
            border-color: var(--primary);
            box-shadow: 0 0 8px rgba(102, 126, 234, 0.2);
        }

        .stations-list {
            display: flex;
            flex-direction: column;
            gap: 8px;
            max-height: 400px;
            overflow-y: auto;
        }

        .station-item {
            background: var(--bg-lighter);
            border: 1px solid var(--border-color);
            border-radius: 8px;
            padding: 12px;
            cursor: pointer;
            transition: all 0.3s ease;
            display: flex;
            justify-content: space-between;
            align-items: center;
            gap: 12px;
        }

        .station-item:hover {
            background: #2a2a3e;
            border-color: var(--primary);
            box-shadow: 0 2px 8px rgba(102, 126, 234, 0.2);
        }

        .station-item.active {
            background: linear-gradient(135deg, var(--primary) 0%, var(--primary-dark) 100%);
            border-color: transparent;
        }

        .station-info {
            flex: 1;
            min-width: 0;
        }

        .station-name-item {
            font-weight: 600;
            font-size: 14px;
            margin-bottom: 4px;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        .station-genre {
            font-size: 12px;
            color: var(--text-secondary);
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        .station-actions {
            display: flex;
            gap: 8px;
            flex-shrink: 0;
        }

        .icon-btn {
            background: transparent;
            border: none;
            color: var(--text-secondary);
            cursor: pointer;
            font-size: 18px;
            padding: 4px 8px;
            transition: all 0.2s ease;
            display: inline-flex;
            align-items: center;
            justify-content: center;
        }

        .icon-btn:hover {
            color: var(--primary);
            transform: scale(1.2);
        }

        .icon-btn.favorite.active {
            color: #ff6b6b;
        }

        .add-station-form {
            background: var(--bg-lighter);
            border: 1px solid var(--border-color);
            border-radius: 8px;
            padding: 16px;
            margin-top: auto;
            flex-shrink: 0;
        }

        .form-title {
            font-size: 14px;
            font-weight: 600;
            margin-bottom: 12px;
            color: var(--text-secondary);
            text-transform: uppercase;
            letter-spacing: 1px;
        }

        .form-group {
            margin-bottom: 10px;
        }

        .form-group label {
            display: block;
            font-size: 12px;
            font-weight: 500;
            margin-bottom: 4px;
            color: var(--text-secondary);
        }

        .form-group input {
            width: 100%;
            padding: 8px 12px;
            background: var(--bg-dark);
            border: 1px solid var(--border-color);
            color: var(--text-primary);
            border-radius: 6px;
            font-size: 13px;
        }

        .form-group input:focus {
            outline: none;
            border-color: var(--primary);
            box-shadow: 0 0 6px rgba(102, 126, 234, 0.2);
        }

        .form-buttons {
            display: flex;
            gap: 8px;
            margin-top: 12px;
        }

        .form-buttons .btn {
            flex: 1;
            font-size: 13px;
            padding: 8px 12px;
        }

        .status {
            padding: 10px 12px;
            border-radius: 6px;
            font-size: 12px;
            display: none;
            animation: slideDown 0.3s ease;
        }

        .status.show {
            display: block;
        }

        .status.success {
            background: rgba(76, 175, 80, 0.2);
            color: #4caf50;
            border: 1px solid #4caf50;
        }

        .status.error {
            background: rgba(244, 67, 54, 0.2);
            color: #f44336;
            border: 1px solid #f44336;
        }

        .status.info {
            background: rgba(33, 150, 243, 0.2);
            color: #2196f3;
            border: 1px solid #2196f3;
        }

        @keyframes slideDown {
            from {
                opacity: 0;
                transform: translateY(-10px);
            }
            to {
                opacity: 1;
                transform: translateY(0);
            }
        }

        .loading {
            display: inline-block;
            width: 16px;
            height: 16px;
            border: 2px solid var(--border-color);
            border-top-color: var(--primary);
            border-radius: 50%;
            animation: spin 0.8s linear infinite;
        }

        @keyframes spin {
            to { transform: rotate(360deg); }
        }

        .tabs {
            display: flex;
            gap: 8px;
            margin-bottom: 12px;
            border-bottom: 1px solid var(--border-color);
        }

        .tab {
            background: transparent;
            border: none;
            color: var(--text-secondary);
            padding: 8px 12px;
            cursor: pointer;
            font-size: 12px;
            font-weight: 600;
            border-bottom: 2px solid transparent;
            transition: all 0.3s ease;
        }

        .tab.active {
            color: var(--primary);
            border-bottom-color: var(--primary);
        }

        .tab:hover {
            color: var(--text-primary);
        }

        .empty-state {
            text-align: center;
            padding: 20px;
            color: var(--text-secondary);
            font-size: 14px;
        }

        .empty-state-icon {
            font-size: 32px;
            margin-bottom: 8px;
        }

        @media (max-width: 768px) {
            .main-content {
                padding: 12px;
                gap: 12px;
            }

            .player-card {
                padding: 16px;
            }

            .header {
                padding: 12px 16px;
            }

            .header-title {
                font-size: 20px;
            }

            .station-item {
                flex-direction: column;
                align-items: flex-start;
            }

            .station-actions {
                width: 100%;
                justify-content: flex-end;
            }

            .form-group input {
                font-size: 16px;
            }
        }

        @media (max-width: 480px) {
            body {
                font-size: 14px;
            }

            .header-title {
                font-size: 18px;
            }

            .station-name {
                font-size: 16px;
            }

            .player-controls {
                gap: 8px;
            }

            .btn {
                padding: 8px 16px;
                font-size: 12px;
            }

            .station-meta {
                gap: 8px;
            }

            .add-station-form {
                max-height: 200px;
                overflow-y: auto;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1 class="header-title">🎙️ radioKeenetik</h1>
        </div>

        <div class="main-content">
            <div id="status" class="status"></div>

            <!-- PLAYER CARD -->
            <div class="player-card">
                <div class="now-playing">▶️ Сейчас играет</div>
                <div class="station-name" id="currentStationName">—</div>
                <div class="station-meta">
                    <div class="meta-item">
                        <span>🎵</span>
                        <span id="currentGenre">—</span>
                    </div>
                    <div class="meta-item">
                        <span>🌍</span>
                        <span id="currentCountry">—</span>
                    </div>
                </div>

                <div class="player-controls">
                    <button class="btn" id="playBtn" onclick="playCurrentStation()">▶️ Играть</button>
                    <button class="btn btn-secondary" onclick="stopPlayback()">⏹️ Стоп</button>
                </div>

                <div class="slider-container">
                    <div class="slider-label">
                        <span>🔊 Громкость</span>
                        <span id="volumeValue">100%</span>
                    </div>
                    <input type="range" id="volumeSlider" class="slider" min="0" max="100" value="100">
                </div>

                <div class="device-selector">
                    <label for="deviceSelect" style="display: block; margin-bottom: 6px; font-size: 12px; color: var(--text-secondary);">🎧 Аудиокарта</label>
                    <select id="deviceSelect">
                        <option value="default">По умолчанию</option>
                    </select>
                </div>
            </div>

            <!-- STATIONS SECTION -->
            <div class="section">
                <div class="tabs">
                    <button class="tab active" data-tab="all" onclick="switchTab('all')">Все</button>
                    <button class="tab" data-tab="favorites" onclick="switchTab('favorites')">⭐ Избранное</button>
                </div>

                <div class="search-box">
                    <input type="text" id="searchInput" placeholder="🔍 Поиск по названию..." onkeyup="filterStations()">
                </div>

                <div class="stations-list" id="stationsList"></div>
            </div>

            <!-- ADD STATION FORM -->
            <div class="add-station-form">
                <div class="form-title">➕ Добавить станцию</div>
                <div class="form-group">
                    <label>Название</label>
                    <input type="text" id="stationName" placeholder="Rock Life" maxlength="50">
                </div>
                <div class="form-group">
                    <label>URL потока (MP3, AAC, HLS)</label>
                    <input type="text" id="stationUrl" placeholder="https://stream.example.com/radio.mp3" maxlength="200">
                </div>
                <div class="form-group">
                    <label>Жанр</label>
                    <input type="text" id="stationGenre" placeholder="Рок, Поп, Электро..." maxlength="30">
                </div>
                <div class="form-group">
                    <label>Страна</label>
                    <input type="text" id="stationCountry" placeholder="Россия, США..." maxlength="30">
                </div>
                <div class="form-buttons">
                    <button class="btn" style="flex: 1;" onclick="addStation()">Добавить</button>
                    <button class="btn btn-secondary" style="flex: 0 0 80px;" onclick="clearForm()">Очистить</button>
                </div>
            </div>
        </div>
    </div>

    <script>
        // ========================================================
        // GLOBAL STATE
        // ========================================================
        let stations = [];
        let currentStation = null;
        let currentTab = 'all';
        let searchQuery = '';
        const ws = new WebSocket('ws://' + window.location.host + '/ws');

        // ========================================================
        // WEBSOCKET HANDLING
        // ========================================================
        ws.onopen = function() {
            console.log('✓ Подключено к серверу');
            showStatus('✓ Подключено', 'info');
        };

        ws.onmessage = function(event) {
            const data = JSON.parse(event.data);
            
            if (data.type === 'update' || data.type === 'init') {
                stations = data.stations || [];
                currentStation = data.current;
                updateUI();
                
                if (data.type === 'init') {
                    loadAudioDevices(data.devices || []);
                }
            }
        };

        ws.onerror = function(error) {
            console.error('WebSocket ошибка:', error);
            showStatus('✗ Ошибка подключения', 'error');
        };

        ws.onclose = function() {
            showStatus('✗ Соединение потеряно', 'error');
        };

        // ========================================================
        // DEVICE LOADING
        // ========================================================
        function loadAudioDevices(devices) {
            const select = document.getElementById('deviceSelect');
            
            if (!devices || devices.length <= 1) {
                fetch('/api/audio-devices')
                    .then(r => r.json())
                    .then(data => {
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
                    });
            } else {
                devices.forEach((device, idx) => {
                    if (idx > 0) {
                        const option = document.createElement('option');
                        option.value = device.id;
                        option.text = device.name;
                        select.appendChild(option);
                    }
                });
                if (devices.length > 1) {
                    showStatus('Найдено USB карт: ' + (devices.length - 1), 'info');
                }
            }
        }

        // ========================================================
        // UI UPDATES
        // ========================================================
        function updateUI() {
            updateCurrentStation();
            updateStationsList();
        }

        function updateCurrentStation() {
            const nameEl = document.getElementById('currentStationName');
            const genreEl = document.getElementById('currentGenre');
            const countryEl = document.getElementById('currentCountry');
            const playBtn = document.getElementById('playBtn');

            if (currentStation) {
                nameEl.textContent = currentStation.name;
                genreEl.textContent = currentStation.genre || '—';
                countryEl.textContent = currentStation.country || '—';
                playBtn.innerHTML = '⏸️ Играет';
                playBtn.classList.add('active');
            } else {
                nameEl.textContent = '—';
                genreEl.textContent = '—';
                countryEl.textContent = '—';
                playBtn.innerHTML = '▶️ Играть';
                playBtn.classList.remove('active');
            }
        }

        function updateStationsList() {
            const listEl = document.getElementById('stationsList');
            let filtered = filterStationsList();

            if (filtered.length === 0) {
                listEl.innerHTML = '<div class="empty-state"><div class="empty-state-icon">📻</div><div>Нет станций</div></div>';
                return;
            }

            listEl.innerHTML = '';
            filtered.forEach(station => {
                const isActive = currentStation && currentStation.id === station.id;
                const div = document.createElement('div');
                div.className = 'station-item' + (isActive ? ' active' : '');
                
                div.innerHTML = '<div class="station-info">' +
                    '<div class="station-name-item">' + escapeHtml(station.name) + '</div>' +
                    '<div class="station-genre">' + escapeHtml(station.genre) + ' • ' + escapeHtml(station.country) + '</div>' +
                    '</div>' +
                    '<div class="station-actions">' +
                    '<button class="icon-btn favorite ' + (station.isFavorite ? 'active' : '') + '" ' +
                    'onclick="toggleFavorite(\'' + station.id + '\', event)">⭐</button>' +
                    '<button class="icon-btn" onclick="playStation(\'' + station.id + '\', event)">▶️</button>' +
                    '<button class="icon-btn" onclick="deleteStation(\'' + station.id + '\', event)">✕</button>' +
                    '</div>';

                listEl.appendChild(div);
            });
        }

        function filterStationsList() {
            let filtered = stations;

            if (currentTab === 'favorites') {
                filtered = filtered.filter(s => s.isFavorite);
            }

            if (searchQuery) {
                const query = searchQuery.toLowerCase();
                filtered = filtered.filter(s => 
                    s.name.toLowerCase().includes(query) ||
                    s.genre.toLowerCase().includes(query) ||
                    s.country.toLowerCase().includes(query)
                );
            }

            return filtered;
        }

        // ========================================================
        // TAB SWITCHING
        // ========================================================
        function switchTab(tab) {
            currentTab = tab;
            document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
            event.target.classList.add('active');
            updateStationsList();
        }

        // ========================================================
        // SEARCH
        // ========================================================
        function filterStations() {
            searchQuery = document.getElementById('searchInput').value;
            updateStationsList();
        }

        // ========================================================
        // PLAYER CONTROLS
        // ========================================================
        function playStation(stationId, e) {
            if (e) e.stopPropagation();
            ws.send(JSON.stringify({
                action: 'play',
                stationId: stationId
            }));
        }

        function playCurrentStation() {
            if (currentStation) {
                playStation(currentStation.id);
            } else if (stations.length > 0) {
                playStation(stations[0].id);
            }
        }

        function stopPlayback() {
            ws.send(JSON.stringify({
                action: 'stop'
            }));
        }

        function toggleFavorite(stationId, e) {
            if (e) e.stopPropagation();
            ws.send(JSON.stringify({
                action: 'toggleFavorite',
                stationId: stationId
            }));
        }

        function deleteStation(stationId, e) {
            if (e) e.stopPropagation();
            if (confirm('Удалить станцию?')) {
                ws.send(JSON.stringify({
                    action: 'deleteStation',
                    stationId: stationId
                }));
            }
        }

        // ========================================================
        // VOLUME & DEVICE
        // ========================================================
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

        // ========================================================
        // ADD STATION
        // ========================================================
        function addStation() {
            const name = document.getElementById('stationName').value.trim();
            const url = document.getElementById('stationUrl').value.trim();
            const genre = document.getElementById('stationGenre').value.trim();
            const country = document.getElementById('stationCountry').value.trim();

            if (!name || !url || !genre || !country) {
                showStatus('⚠️ Заполните все поля', 'error');
                return;
            }

            if (!isValidUrl(url)) {
                showStatus('⚠️ URL должен начинаться с http:// или https://', 'error');
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

            clearForm();
            showStatus('✓ Станция добавлена', 'success');
        }

        function clearForm() {
            document.getElementById('stationName').value = '';
            document.getElementById('stationUrl').value = '';
            document.getElementById('stationGenre').value = '';
            document.getElementById('stationCountry').value = '';
        }

        // ========================================================
        // UTILITIES
        // ========================================================
        function isValidUrl(url) {
            try {
                new URL(url);
                return url.startsWith('http://') || url.startsWith('https://');
            } catch {
                return false;
            }
        }

        function escapeHtml(text) {
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        }

        function showStatus(message, type) {
            const statusEl = document.getElementById('status');
            statusEl.textContent = message;
            statusEl.className = 'status show ' + type;
            
            setTimeout(() => {
                statusEl.classList.remove('show');
            }, 4000);
        }

        // Initialize
        updateUI();
    </script>
</body>
</html>`
}
