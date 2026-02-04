package config

import (
	"os"
	"strconv"
)

// Config contém as configurações da aplicação.
type Config struct {
	Port      string
	Database  DatabaseConfig
	JWTSecret string
}

type DatabaseConfig struct {
	Driver string
	DSN    string // Data Source Name (caminho do arquivo SQLite ou URL do Postgres)
}

// Load carrega as configurações das variáveis de ambiente ou usa padrões.
func Load() *Config {
	return &Config{
		Port: getEnv("PORT", "8080"),
		Database: DatabaseConfig{
			Driver: getEnv("DB_DRIVER", "sqlite3"), // ncruces usa "sqlite3"
			DSN:    getEnv("DB_DSN", "./rankit.db"),
		},
		JWTSecret: getEnv("JWT_SECRET", "segredo_padrao_para_desenvolvimento"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}
