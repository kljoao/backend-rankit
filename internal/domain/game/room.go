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

	PendingPlayers map[string]*Player // Map[SessionID]*Player (Aguardando aprovação)
	Players        map[string]*Player // Map[SessionID]*Player (Aprovados)
	Answers        map[string]*Answer // Map[PlayerID]*Answer (da pergunta atual)

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
		PendingPlayers:       make(map[string]*Player),
		Answers:              make(map[string]*Answer),
	}
}

// --- Métodos de Controle do Jogo (State Machine) ---

// JoinRequest adiciona um jogador à lista de pendentes.
func (r *Room) JoinRequest(sessionID, nickname string) (*Player, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 1. Bloqueio de entrada se sala já iniciada
	if r.Status != StateLobby {
		// Se já é player aprovado, permite reconexão (lógica abaixo)
		if p, ok := r.Players[sessionID]; ok {
			p.Connected = true
			return p, nil
		}
		return nil, ErrSalaIniciada
	}

	// 2. Se já aprovado, retorna ok
	if p, ok := r.Players[sessionID]; ok {
		p.Connected = true
		return p, nil
	}

	// 3. Adiciona aos pendentes
	// Check duplicidade de nickname (opcional, mas bom)
	for _, p := range r.Players {
		if p.Nickname == nickname {
			return nil, errors.New("apelido já em uso na sala")
		}
	}

	p := &Player{
		ID:        sessionID,
		Nickname:  nickname,
		Score:     0,
		Connected: true,
	}
	r.PendingPlayers[sessionID] = p
	return p, nil
}

// ApprovePlayer move o jogador de pendente para aprovado.
func (r *Room) ApprovePlayer(sessionID string) (*Player, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, ok := r.PendingPlayers[sessionID]
	if !ok {
		// Se já estiver em Players, ignora erro
		if approved, exists := r.Players[sessionID]; exists {
			return approved, nil
		}
		return nil, errors.New("jogador não encontrada na lista de pendentes")
	}

	delete(r.PendingPlayers, sessionID)
	r.Players[sessionID] = p
	return p, nil
}

// RejectPlayer remove o jogador da lista de pendentes.
func (r *Room) RejectPlayer(sessionID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.PendingPlayers[sessionID]; !ok {
		return errors.New("jogador não encontrado na lista de pendentes")
	}
	delete(r.PendingPlayers, sessionID)
	return nil
}

// RemovePlayer remove um jogador aprovado da sala (Kick).
func (r *Room) RemovePlayer(playerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.Players[playerID]; !ok {
		return errors.New("jogador não encontrado na sala")
	}

	delete(r.Players, playerID)
	// Também remove resposta se houver
	delete(r.Answers, playerID)

	return nil
}

// Join (Legado/Direto) - Mantido para compatibilidade se necessário, ou removido/adaptado
func (r *Room) Join(playerID, nickname string) (*Player, error) {
	// Redireciona para JoinRequest por padrão, ou mantém lógica antiga
	// Se quisermos manter compatibilidade sem moderação, podemos usar este.
	// Mas o requisito pede moderação.
	return r.JoinRequest(playerID, nickname)
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
