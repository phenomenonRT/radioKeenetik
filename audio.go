package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// AudioPlayer плеер для воспроизведения аудио
type AudioPlayer struct {
	cmd     *exec.Cmd
	process *os.Process
	ctx     context.Context
	cancel  context.CancelFunc
	station *Station
	state   PlaybackState
	mu      sync.RWMutex
	device  string // Текущее устройство (hw:1,0 или default)
	volume  int    // 0-100
	player  string // mpv, ffmpeg, mplayer, ffplay
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
		cmd = ap.buildMpvCommand(url)
	case "ffplay":
		cmd = ap.buildFfplayCommand(url)
	case "mplayer":
		cmd = ap.buildMplayerCommand(url)
	case "play":
		cmd = ap.buildSoxCommand(url)
	default:
		return fmt.Errorf("неизвестный плеер: %s", playerName)
	}

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

// buildMpvCommand собирает команду для mpv
func (ap *AudioPlayer) buildMpvCommand(url string) *exec.Cmd {
	args := []string{
		"--no-terminal",
		"--really-quiet",
		fmt.Sprintf("--volume=%d", ap.volume),
	}

	if ap.device != "default" {
		args = append(args, fmt.Sprintf("--ao=alsa:device=%s", ap.device))
	}

	args = append(args, url)
	log.Printf("▶ Запуск mpv: %s", url)
	return exec.CommandContext(ap.ctx, ap.player, args...)
}

// buildFfplayCommand собирает команду для ffplay
func (ap *AudioPlayer) buildFfplayCommand(url string) *exec.Cmd {
	args := []string{
		"-nodisp",
		"-autoexit",
		"-loglevel", "quiet",
		url,
	}
	log.Printf("▶ Запуск ffplay: %s", url)
	return exec.CommandContext(ap.ctx, ap.player, args...)
}

// buildMplayerCommand собирает команду для mplayer
func (ap *AudioPlayer) buildMplayerCommand(url string) *exec.Cmd {
	args := []string{
		"-really-quiet",
		"-msglevel", "all=0",
		fmt.Sprintf("-volume=%d", ap.volume),
	}

	if ap.device != "default" {
		args = append(args, "-ao", fmt.Sprintf("alsa:device=%s", ap.device))
	}

	args = append(args, url)
	log.Printf("▶ Запуск mplayer: %s", url)
	return exec.CommandContext(ap.ctx, ap.player, args...)
}

// buildSoxCommand собирает команду для sox play
func (ap *AudioPlayer) buildSoxCommand(url string) *exec.Cmd {
	args := []string{"-q", "-d", url}
	log.Printf("▶ Запуск sox play: %s", url)
	return exec.CommandContext(ap.ctx, ap.player, args...)
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

// SetVolume устанавливает громкость (0-100)
func (ap *AudioPlayer) SetVolume(vol int) {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	// Валидация
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
		parts := strings.Split(ap.device, ":")
		if len(parts) > 1 {
			cardNum := strings.Split(parts[1], ",")[0]
			exec.Command("amixer", "-c", cardNum, "set", "Master", fmt.Sprintf("%d%%", vol)).Run()
		}
	} else {
		// Для default используем pactl (PulseAudio)
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

// ============================================================================
// AUDIO DEVICES
// ============================================================================

// AudioDevice информация об аудиокарте
type AudioDevice struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// GetAudioDevices возвращает список USB аудиокарт
func GetAudioDevices() []AudioDevice {
	devices := []AudioDevice{
		{ID: "default", Name: "По умолчанию"},
	}

	cmd := exec.Command("aplay", "-l")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "card") {
				if parts := strings.Split(line, "card"); len(parts) > 1 {
					cardInfo := strings.Split(parts[1], ":")[0]
					cardNum := strings.TrimSpace(cardInfo)

					if strings.Contains(line, "USB") {
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
