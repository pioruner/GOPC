package loger

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

func Create(name string) *log.Logger {
	execPath, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to get executable path: %s", err)
	}
	execDir := filepath.Dir(execPath)

	// Определяем путь для логов
	logDir := filepath.Join(execDir, "logs")
	logFile := filepath.Join(logDir, name+".log")

	if err := os.MkdirAll(logDir, 0755); err != nil { // Создаем папку
		log.Fatalf("Failed to create log directory: %s", err)
	}

	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666) // Открываем файл для логов
	if err != nil {
		log.Fatalf("Failed to open log file: %s", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Error closing file log: %v", err)
		}
	}()

	multiWriter := io.MultiWriter(os.Stdout, file)                                    // Настраиваем output (в терминал и файл)
	logger := log.New(multiWriter, "["+name+"] ", log.Ldate|log.Ltime|log.Lshortfile) // Создаём новый лог
	return logger
}
