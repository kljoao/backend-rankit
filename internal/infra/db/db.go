package db

import (
	"database/sql"
	"rankit/internal/infra/logger"

	_ "github.com/ncruces/go-sqlite3/driver" // Driver SQLite via Wazero (Pure Go)
	_ "github.com/ncruces/go-sqlite3/embed"  // Embed binary
)

// NewSQLiteConnection abre uma conexão com o banco de dados SQLite.
func NewSQLiteConnection(dsn string) (*sql.DB, error) {
	// Driver "sqlite3"
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		logger.Error("Falha ao abrir conexão com banco de dados", "erro", err)
		return nil, err
	}

	if err := db.Ping(); err != nil {
		logger.Error("Falha ao conectar com banco de dados (ping)", "erro", err)
		return nil, err
	}

	logger.Info("Conectado ao banco de dados SQLite com sucesso", "dsn", dsn)
	return db, nil
}
