package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
)

func main() {
	// Запуск логирования
	log.Println("")
	log.Println("╔══════════════════════════════════════════════════════════╗")
	log.Println("║  🎙️  radioKeenetik - Интернет-радио на USB аудиокарте  ║")
	log.Println("╚══════════════════════════════════════════════════════════╝")
	log.Println("")

	// Создать сервер
	radioServer := NewRadioServer()

	// Проверить плеер
	if radioServer.player.player == "" {
		log.Println("⚠️  ОШИБКА: Не найден ни один плеер!")
		log.Println("   Установите плеер: sudo apt install -y mpv")
		os.Exit(1)
	}

	log.Printf("✓ Плеер: %s", radioServer.player.player)

	// Проверить аудиокарты
	devices := GetAudioDevices()
	log.Printf("✓ Найдено USB устройств: %d", len(devices)-1)

	// Добавить примеры станций
	radioServer.AddStation(&Station{
		ID:      "1",
		Name:    "Radio Record - Live DJ Sets",
		URL:     "https://orfeyfm.hostingradio.ru:8034/orfeyfm128.mp3",
		Genre:   "Electronic",
		Country: "Russia",
	})

	log.Println("")
	log.Println("════════════════════════════════════════════════════════════")

	// Настроить маршруты
	r := mux.NewRouter()
	radioServer.RegisterRoutes(r)

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("")
		log.Println("⏹️  Выключение сервера...")
		radioServer.StopPlayback()
		os.Exit(0)
	}()

	// Запустить сервер
	addr := ":8080"
	log.Printf("✓ Сервер запущен на http://0.0.0.0%s", addr)
	log.Printf("✓ Откройте в браузере: http://localhost%s", addr)
	log.Println("")
	log.Println("════════════════════════════════════════════════════════════")
	log.Println("")

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("✗ Ошибка сервера: %v", err)
	}
}
