package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
	DebugLogger *log.Logger
)

// SetupLogger mengkonfigurasi logger dengan rotating files
func SetupLogger() {
	logPath := GetEnv("LOG_FILE_PATH", "./logs")

	// Buat direktori logs jika belum ada
	if err := os.MkdirAll(logPath, 0755); err != nil {
		log.Fatal("Failed to create logs directory:", err)
	}

	// Format tanggal untuk nama file
	dateStr := time.Now().Format("2006-01-02")

	// File untuk info log
	infoFile, err := os.OpenFile(
		filepath.Join(logPath, fmt.Sprintf("info-%s.log", dateStr)),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		log.Fatal("Failed to open info log file:", err)
	}

	// File untuk error log
	errorFile, err := os.OpenFile(
		filepath.Join(logPath, fmt.Sprintf("error-%s.log", dateStr)),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		log.Fatal("Failed to open error log file:", err)
	}

	// File untuk debug log
	debugFile, err := os.OpenFile(
		filepath.Join(logPath, fmt.Sprintf("debug-%s.log", dateStr)),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		log.Fatal("Failed to open debug log file:", err)
	}

	// Setup loggers
	InfoLogger = log.New(infoFile, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(errorFile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	DebugLogger = log.New(debugFile, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)

	log.Println("Logger setup completed")
}