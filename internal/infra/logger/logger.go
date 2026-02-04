package logger

import (
	"log/slog"
	"os"
)

// Init inicializa o logger global.
func Init() {
	// Cria um logger JSON estruturado
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

// LogInfo registra uma mensagem de informação.
func Info(msg string, args ...any) {
	slog.Info(msg, args...)
}

// LogError registra uma mensagem de erro.
func Error(msg string, args ...any) {
	slog.Error(msg, args...)
}
