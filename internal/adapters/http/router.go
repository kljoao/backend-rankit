package httpadapter

import (
	"fmt"
	"net/http"
	"rankit/internal/adapters/http/handlers"
	"rankit/internal/adapters/http/middlewares"
	"rankit/internal/adapters/websocket"
	"rankit/internal/ports"

	_ "rankit/docs"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// NewRouter configura as rotas e middlewares.
func NewRouter(
	authHandler *handlers.AuthHandler,
	quizHandler *handlers.QuizHandler,
	questionHandler *handlers.QuestionHandler,
	gameHandler *handlers.GameHandler,
	reportHandler *handlers.ReportHandler,
	wsHandler *websocket.WebSocketHandler,
	tokenService ports.TokenService,
) http.Handler {
	r := chi.NewRouter()

	// Middlewares globais
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Configuração CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Rota de Health Check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Swagger
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("http://localhost:8080/swagger/doc.json")),
	))

	// WebSocket Endpoint
	r.Get("/ws", wsHandler.HandleWS)

	// Grupo de rotas de Auth
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)

		// Rotas protegidas (Auth)
		r.Group(func(r chi.Router) {
			r.Use(middlewares.AuthMiddleware(tokenService))
			r.Get("/me", authHandler.GetMe)
		})
	})

	// Grupo de rotas de Quizzes (Protegidas)
	r.Route("/quizzes", func(r chi.Router) {
		r.Use(middlewares.AuthMiddleware(tokenService))

		r.Post("/", quizHandler.CreateQuiz)
		r.Get("/", quizHandler.ListQuizzes)
		r.Get("/{id}", quizHandler.GetQuiz)
		r.Put("/{id}", quizHandler.UpdateQuiz)
		r.Delete("/{id}", quizHandler.DeleteQuiz)
		r.Post("/{id}/publish", quizHandler.PublishQuiz)

		// Sub-rotas de Questions
		r.Route("/{id}/questions", func(r chi.Router) {
			r.Post("/", questionHandler.AddQuestion)
			r.Post("/reorder", questionHandler.ReorderQuestions)

			r.Put("/{questionId}", questionHandler.UpdateQuestion)
			r.Delete("/{questionId}", questionHandler.RemoveQuestion)
		})
	})

	// Grupo de rotas de Salas (Game)
	r.Route("/rooms", func(r chi.Router) {
		// Criar sala exige autenticação (Professor)
		r.Group(func(r chi.Router) {
			r.Use(middlewares.AuthMiddleware(tokenService))
			r.Post("/", gameHandler.CreateRoom)
		})

		// Visualizar detalhes da sala pode ser público (para alunos confirmarem info)
		// Ou restrito. O requisito diz "GET /rooms/{roomId}/join-info"
		// Vamos expor GET /rooms/{id} como público por enquanto (join-info)
		r.Get("/{id}", gameHandler.GetRoom)
	})

	// Grupo de rotas de Relatórios (Protegidas)
	r.Route("/reports", func(r chi.Router) {
		r.Use(middlewares.AuthMiddleware(tokenService))

		r.Get("/rooms", reportHandler.ListRooms)
		r.Get("/rooms/{id}", reportHandler.GetRoomDetail)
		r.Get("/quizzes/{id}", reportHandler.GetQuizStats)
	})

	return r
}
