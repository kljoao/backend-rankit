package game

import (
	"errors"
	"rankit/internal/domain/quiz"
	"sync"
	"time"
)

// Estados da Sala (State Machine)
const (
	StateLobby    = "LOBBY"
	StateOpen     = "OPEN"
	StateRevealed = "REVEALED"
	StateFinished = "FINISHED"
)

var (
	ErrSalaIniciada       = errors.New("a sala já foi iniciada")
	ErrSalaNaoAberta      = errors.New("a pergunta não está aberta para respostas")
	ErrJogoFinalizado     = errors.New("o jogo já foi finalizado")
	ErrPermissaoProfessor = errors.New("apenas o professor pode realizar esta ação")
)

// Player representa um aluno na sala.
type Player struct {
	ID        string `json:"id"` // Session ID / Socket ID
	Nickname  string `json:"nickname"`
	Score     int    `json:"score"`
	Connected bool   `json:"connected"`
}

// Answer representa a resposta de um aluno para a pergunta atual.
type Answer struct {
	PlayerID    string
	AnswerIndex int // 0..3
	SubmittedAt time.Time
}

// Room representa uma sala de aula ao vivo.
// Mantém o estado do jogo em memória.
type Room struct {
	ID        string
	TeacherID string
	Quiz      *quiz.Quiz

	Status               string
	CurrentQuestionIndex int

	Players map[string]*Player // Map[PlayerID]*Player
	Answers map[string]*Answer // Map[PlayerID]*Answer (da pergunta atual)

	mu sync.RWMutex // Mutex para garantir thread-safety
}

// NewRoom cria uma nova sala.
func NewRoom(id, teacherID string, q *quiz.Quiz) *Room {
	return &Room{
		ID:                   id,
		TeacherID:            teacherID,
		Quiz:                 q,
		Status:               StateLobby,
		CurrentQuestionIndex: -1, // Ainda não começou
		Players:              make(map[string]*Player),
		Answers:              make(map[string]*Answer),
	}
}

// --- Métodos de Controle do Jogo (State Machine) ---

// Join adiciona um jogador à sala.
func (r *Room) Join(playerID, nickname string) (*Player, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Se já existe, reconecta
	if p, exists := r.Players[playerID]; exists {
		p.Connected = true
		return p, nil
	}

	// Regra: Não entrar se jogo finalizado? (Opcional, mas vamos permitir spectate)

	p := &Player{
		ID:        playerID,
		Nickname:  nickname,
		Score:     0,
		Connected: true,
	}
	r.Players[playerID] = p
	return p, nil
}

// NextQuestion avança para a próxima pergunta (ou inicia a primeira).
// Estado: OPEN
func (r *Room) NextQuestion() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Status == StateFinished {
		return ErrJogoFinalizado
	}

	nextIndex := r.CurrentQuestionIndex + 1
	if nextIndex >= len(r.Quiz.Questions) {
		r.Status = StateFinished
		return nil
	}

	r.CurrentQuestionIndex = nextIndex
	r.Status = StateOpen
	r.Answers = make(map[string]*Answer) // Limpa respostas da rodada anterior

	return nil
}

// RevealQuestion revela a resposta da pergunta atual e calcula pontuação.
// Estado: REVEALED
func (r *Room) RevealQuestion() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Status != StateOpen {
		return errors.New("a pergunta não está aberta")
	}

	currentQ := r.Quiz.Questions[r.CurrentQuestionIndex]

	// Calcula pontuação
	for _, ans := range r.Answers {
		if ans.AnswerIndex == currentQ.CorrectIndex {
			if p, exists := r.Players[ans.PlayerID]; exists {
				p.Score += 10
			}
		}
	}

	r.Status = StateRevealed
	return nil
}

// SubmitAnswer registra a resposta de um aluno.
func (r *Room) SubmitAnswer(playerID string, answerIndex int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Status != StateOpen {
		return ErrSalaNaoAberta
	}

	if _, exists := r.Players[playerID]; !exists {
		return errors.New("jogador não está na sala")
	}

	r.Answers[playerID] = &Answer{
		PlayerID:    playerID,
		AnswerIndex: answerIndex,
		SubmittedAt: time.Now(),
	}

	return nil
}

// StateSnapshot retorna o estado atual para enviar ao cliente.
type RoomStateDTO struct {
	Status               string         `json:"status"`
	CurrentQuestion      *quiz.Question `json:"currentQuestion,omitempty"`
	TotalQuestions       int            `json:"totalQuestions"`
	CurrentQuestionIndex int            `json:"currentQuestionIndex"`
	PlayersCount         int            `json:"playersCount"`
	AnswersCount         int            `json:"answersCount"`           // Quantos responderam
	CorrectIndex         int            `json:"correctIndex,omitempty"` // Só enviado se REVEALED
}

func (r *Room) GetStateSnapshot() RoomStateDTO {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var currentQ *quiz.Question
	correctIndex := -1

	if r.CurrentQuestionIndex >= 0 && r.CurrentQuestionIndex < len(r.Quiz.Questions) {
		q := r.Quiz.Questions[r.CurrentQuestionIndex]
		// Clona para não expor correct index se OPEN
		qCopy := q
		if r.Status == StateOpen {
			qCopy.CorrectIndex = -1
		} else if r.Status == StateRevealed {
			correctIndex = q.CorrectIndex
		}
		currentQ = &qCopy
	}

	return RoomStateDTO{
		Status:               r.Status,
		CurrentQuestion:      currentQ,
		TotalQuestions:       len(r.Quiz.Questions),
		CurrentQuestionIndex: r.CurrentQuestionIndex,
		PlayersCount:         len(r.Players),
		AnswersCount:         len(r.Answers),
		CorrectIndex:         correctIndex,
	}
}

// GetLeaderboard retorna a lista de jogadores ordenada por score.
func (r *Room) GetLeaderboard() []*Player {
	r.mu.RLock()
	defer r.mu.RUnlock()

	players := make([]*Player, 0, len(r.Players))
	for _, p := range r.Players {
		players = append(players, p)
	}
	// A ordenação será feita no UseCase ou aqui?
	// Vamos deixar array simples aqui, ordenação pode ser feita no output.
	return players
}
