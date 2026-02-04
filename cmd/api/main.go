package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"

	httpadapter "rankit/internal/adapters/http"
	"rankit/internal/adapters/http/handlers"
	"rankit/internal/adapters/persistence"
	"rankit/internal/adapters/security"
	"rankit/internal/adapters/websocket"
	"rankit/internal/application/usecases"
	"rankit/internal/infra/config"
	infraDB "rankit/internal/infra/db"
	"rankit/internal/infra/logger"

	_ "rankit/docs"
)

// @title RankIt API
// @version 1.0
// @description Backend para a plataforma RankIt (Gamificação em sala de aula).
// @contact.name Suporte RankIt
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @externalDocs.description  Documentação WebSocket (Não interativo via Swagger)
// @externalDocs.url          /walkthrough.md
func main() {
	// 1. Configuração e Logger
	logger.Init()
	cfg := config.Load()

	// 2. Banco de Dados
	db, err := infraDB.NewSQLiteConnection(cfg.Database.DSN)
	if err != nil {
		logger.Error("Não foi possível conectar ao banco", "erro", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := runMigrations(db); err != nil {
		logger.Error("Falha na migração", "erro", err)
		os.Exit(1)
	}

	// 3a. Adapters (Driving - Persistence)
	teacherRepo := persistence.NewSQLiteTeacherRepository(db)
	quizRepo := persistence.NewSQLiteQuizRepository(db)
	questionRepo := persistence.NewSQLiteQuestionRepository(db)

	// Novo - Repositório In-Memory
	gameRepo := persistence.NewInMemoryGameRepository()
	// Novo - Repositório Histórico
	historyRepo := persistence.NewSQLiteHistoryRepository(db)

	hasher := security.NewBcryptHasher()
	tokenService := security.NewJWTService(cfg.JWTSecret)

	// 3b. Adapters (Driving - WebSocket Hub)
	wsHub := websocket.NewHub()
	// Inicia o Hub em background
	go wsHub.Run()

	// 4. Application (Use Cases)
	registerUC := usecases.NewRegisterTeacherUseCase(teacherRepo, hasher)
	loginUC := usecases.NewLoginTeacherUseCase(teacherRepo, hasher, tokenService)
	getMeUC := usecases.NewGetMeUseCase(teacherRepo)

	quizUC := usecases.NewQuizUseCases(quizRepo)
	questionUC := usecases.NewQuestionUseCases(quizRepo, questionRepo)

	// Novo - Use Case de Jogo
	historyUC := usecases.NewHistoryUseCases(historyRepo, gameRepo)
	gameUC := usecases.NewGameUseCases(gameRepo, quizRepo, wsHub, historyUC)

	// 5. Adapters (Driven - Handlers)
	authHandler := handlers.NewAuthHandler(registerUC, loginUC, getMeUC)
	quizHandler := handlers.NewQuizHandler(quizUC)
	questionHandler := handlers.NewQuestionHandler(questionUC)
	gameHandler := handlers.NewGameHandler(gameUC)
	reportHandler := handlers.NewReportHandler(historyUC)

	wsHandler := websocket.NewWebSocketHandler(wsHub, gameUC)

	// 6. Router
	router := httpadapter.NewRouter(
		authHandler,
		quizHandler,
		questionHandler,
		gameHandler,
		reportHandler,
		wsHandler,
		tokenService,
	)

	// 7. Servidor
	logger.Info("Iniciando servidor", "porta", cfg.Port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", cfg.Port), router); err != nil {
		logger.Error("Falha no servidor HTTP", "erro", err)
	}
}

func runMigrations(db *sql.DB) error {
	files, err := os.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("erro ao ler diretório migrations: %w", err)
	}

	var filenames []string
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".sql" {
			filenames = append(filenames, f.Name())
		}
	}
	sort.Strings(filenames)

	for _, filename := range filenames {
		path := filepath.Join("migrations", filename)
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("erro ao ler %s: %w", filename, err)
		}

		logger.Info("Executando migração", "arquivo", filename)
		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("erro ao executar %s: %w", filename, err)
		}
	}
	return nil
}
