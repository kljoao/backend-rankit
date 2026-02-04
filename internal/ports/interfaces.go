package ports

import (
	"context"
	"rankit/internal/domain/game"
	"rankit/internal/domain/history"
	"rankit/internal/domain/quiz"
	"rankit/internal/domain/teacher"
)

// TeacherRepository define as operações de persistência para a entidade Teacher.
type TeacherRepository interface {
	// Create salva um novo professor no banco de dados.
	Create(ctx context.Context, teacher *teacher.Teacher) error

	// FindByEmail busca um professor pelo email. Retorna erro se não encontrar.
	FindByEmail(ctx context.Context, email string) (*teacher.Teacher, error)

	// FindByID busca um professor pelo ID.
	FindByID(ctx context.Context, id string) (*teacher.Teacher, error)
}

// PasswordHasher define o contrato para hash e verificação de senhas.
type PasswordHasher interface {
	// HashPassword gera um hash seguro da senha.
	HashPassword(password string) (string, error)

	// ComparePassword compara uma senha em texto plano com um hash.
	// Retorna nil se forem iguais, ou erro se forem diferentes.
	ComparePassword(hash, password string) error
}

// TokenService define o contrato para geração e validação de tokens JWT.
type TokenService interface {
	// GenerateToken gera um token de acesso para o ID do usuário fornecido.
	GenerateToken(userID string) (string, int64, error)

	// ValidateToken valida o token e retorna o ID do usuário se válido.
	ValidateToken(tokenString string) (string, error)
}

// QuizRepository define persistência para Quizzes.
type QuizRepository interface {
	Save(ctx context.Context, quiz *quiz.Quiz) error
	FindByID(ctx context.Context, id string) (*quiz.Quiz, error)
	FindByTeacherID(ctx context.Context, teacherID string) ([]*quiz.Quiz, error)
	Delete(ctx context.Context, id string) error
	Update(ctx context.Context, quiz *quiz.Quiz) error
}

// QuestionRepository define persistência para Perguntas.
type QuestionRepository interface {
	Save(ctx context.Context, q *quiz.Question) error
	Delete(ctx context.Context, id string) error
	FindByQuizID(ctx context.Context, quizID string) ([]*quiz.Question, error)
	ReorderQuestions(ctx context.Context, quizID string, questions []*quiz.Question) error
	Update(ctx context.Context, q *quiz.Question) error
}

// GameRepository define persistência em memória para Salas de Jogo.
type GameRepository interface {
	SaveRoom(room *game.Room) error
	FindRoomByID(id string) (*game.Room, error)
	DeleteRoom(id string) error
}

// RealTimeHub define contrato para envio de mensagens via WebSocket.
type RealTimeHub interface {
	BroadcastToRoom(roomID string, message interface{})
	SendToPlayer(playerID string, message interface{})
}

// HistoryRepository define persistência de histórico e relatórios.
type HistoryRepository interface {
	SaveHistory(ctx context.Context, history *history.RoomHistory) error
	ListByTeacherID(ctx context.Context, teacherID string, limit, offset int) ([]*history.RoomHistory, error)
	GetByID(ctx context.Context, id string) (*history.RoomHistory, error)
	GetQuizStats(ctx context.Context, quizID string) (map[string]interface{}, error) // Placeholder para retorno simples
}
