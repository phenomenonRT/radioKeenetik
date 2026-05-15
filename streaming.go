package main

import (
	"io"
	"log"
	"net/http"
	"time"
)

// StreamProxy проксирует аудио потоки
type StreamProxy struct {
	stationID string
	url       string
	client    *http.Client
}

// NewStreamProxy создает новый прокси потока
func NewStreamProxy(stationID, url string) *StreamProxy {
	return &StreamProxy{
		stationID: stationID,
		url:       url,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ProxyStream проксирует аудио поток
func (sp *StreamProxy) ProxyStream(w http.ResponseWriter, r *http.Request) error {
	// Создать запрос к потоку
	req, err := http.NewRequest("GET", sp.url, nil)
	if err != nil {
		return err
	}

	// Скопировать заголовки User-Agent
	req.Header.Set("User-Agent", "radioKeenetik/1.0")

	// Получить ответ
	resp, err := sp.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Проверить статус
	if resp.StatusCode != http.StatusOK {
		return io.ErrClosedPipe
	}

	// Скопировать заголовки ответа
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Установить необходимые заголовки для потока
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "audio/mpeg")
	}
	w.Header().Set("Connection", "close")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Проксировать тело ответа
	_, err = io.Copy(w, resp.Body)
	if err != nil && err != io.EOF {
		log.Printf("Error copying stream: %v", err)
	}

	return nil
}

// StreamInfo информация о потоке
type StreamInfo struct {
	StationID  string        `json:"stationId"`
	Status     string        `json:"status"`
	StartTime  time.Time     `json:"startTime"`
	Duration   time.Duration `json:"duration"`
	BytesSent  int64         `json:"bytesSent"`
	Bitrate    string        `json:"bitrate"`
}

// Playlist структура для M3U плейлиста
type Playlist struct {
	Stations []*Station `json:"stations"`
}

// ExportM3U экспортирует список станций в формат M3U
func (rs *RadioServer) ExportM3U() string {
	m3u := "#EXTM3U\n"

	for _, station := range rs.GetStations() {
		// EXTINF format: duration, name
		m3u += "#EXTINF:-1, " + station.Name + "\n"
		// Stream URL
		m3u += "http://localhost:8080/api/stream?id=" + station.ID + "\n"
	}

	return m3u
}

// ExportPLS экспортирует список станций в формат PLS
func (rs *RadioServer) ExportPLS() string {
	pls := "[playlist]\n"
	stations := rs.GetStations()

	for i, station := range stations {
		num := i + 1
		pls += "File" + string(rune(num)) + "=http://localhost:8080/api/stream?id=" + station.ID + "\n"
		pls += "Title" + string(rune(num)) + "=" + station.Name + "\n"
	}

	pls += "NumberOfEntries=" + string(rune(len(stations))) + "\n"
	pls += "Version=2\n"

	return pls
}
