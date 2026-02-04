package usecases

import (
	"context"
	"errors"
	"rankit/internal/domain/game"
	"rankit/internal/domain/quiz"
	"rankit/internal/ports"

	"github.com/google/uuid"
)

type GameUseCases struct {
	gameRepo  ports.GameRepository
	quizRepo  ports.QuizRepository
	hub       ports.RealTimeHub
	historyUC *HistoryUseCases // Injeção corrigida
}

func NewGameUseCases(
	gameRepo ports.GameRepository,
	quizRepo ports.QuizRepository,
	hub ports.RealTimeHub,
	historyUC *HistoryUseCases, // Argumento adicionado
) *GameUseCases {
	return &GameUseCases{
		gameRepo:  gameRepo,
		quizRepo:  quizRepo,
		hub:       hub,
		historyUC: historyUC,
	}
}

// CreateRoom cria uma sala a partir de um quiz PUBLISHED.
func (uc *GameUseCases) CreateRoom(ctx context.Context, teacherID, quizID string) (*game.Room, error) {
	q, err := uc.quizRepo.FindByID(ctx, quizID)
	if err != nil {
		return nil, err
	}
	if q == nil {
		return nil, errors.New("quiz não encontrado")
	}
	if q.TeacherID != teacherID {
		return nil, ErrNaoAutorizado
	}
	if q.Status != quiz.StatusPublicado {
		return nil, errors.New("apenas quizzes publicados podem ser jogados")
	}

	// Gera ID curto para a sala (simulado com UUID por enquanto)
	roomID := uuid.NewString()[:6]

	room := game.NewRoom(roomID, teacherID, q)
	if err := uc.gameRepo.SaveRoom(room); err != nil {
		return nil, err
	}

	return room, nil
}

// JoinRoom adiciona um jogador à sala.
func (uc *GameUseCases) JoinRoom(roomID, nickname, sessionID string) (*game.Player, error) {
	room, err := uc.gameRepo.FindRoomByID(roomID)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, errors.New("sala não encontrada")
	}

	player, err := room.Join(sessionID, nickname)
	if err != nil {
		return nil, err
	}

	// Notifica a sala sobre o novo jogador
	uc.hub.BroadcastToRoom(roomID, map[string]interface{}{
		"type":    "player_joined",
		"payload": player,
	})

	// Envia o estado atual para o jogador que entrou
	uc.hub.SendToPlayer(sessionID, map[string]interface{}{
		"type":    "room_state",
		"payload": room.GetStateSnapshot(),
	})

	return player, nil
}

// OpenQuestion abre a próxima pergunta ou a atual.
func (uc *GameUseCases) OpenQuestion(roomID, teacherID string) error {
	room, err := uc.gameRepo.FindRoomByID(roomID)
	if err != nil || room == nil {
		return errors.New("sala não encontrada")
	}
	if room.TeacherID != teacherID {
		return ErrNaoAutorizado
	}

	if err := room.NextQuestion(); err != nil {
		return err
	}

	uc.hub.BroadcastToRoom(roomID, map[string]interface{}{
		"type":    "question_opened",
		"payload": room.GetStateSnapshot(),
	})

	// Verifica se acabou de finalizar o jogo
	if room.Status == game.StateFinished {
		// Arquiva automaticamente
		go func() {
			ctx := context.Background()
			if err := uc.historyUC.ArchiveRoom(ctx, room); err != nil {
				// Log silenciado por enquanto
			}
		}()
	}

	return nil
}

// SubmitAnswer recebe a resposta do aluno.
func (uc *GameUseCases) SubmitAnswer(roomID, playerID string, answerIndex int) error {
	room, err := uc.gameRepo.FindRoomByID(roomID)
	if err != nil || room == nil {
		return errors.New("sala não encontrada")
	}

	if err := room.SubmitAnswer(playerID, answerIndex); err != nil {
		return err
	}

	uc.hub.BroadcastToRoom(roomID, map[string]interface{}{
		"type":    "answer_submitted",
		"payload": map[string]int{"answersCount": len(room.Answers)},
	})

	return nil
}

// RevealQuestion revela o resultado da pergunta atual.
func (uc *GameUseCases) RevealQuestion(roomID, teacherID string) error {
	room, err := uc.gameRepo.FindRoomByID(roomID)
	if err != nil || room == nil {
		return errors.New("sala não encontrada")
	}
	if room.TeacherID != teacherID {
		return ErrNaoAutorizado
	}

	if err := room.RevealQuestion(); err != nil {
		return err
	}

	// Envia resultado e placar
	uc.hub.BroadcastToRoom(roomID, map[string]interface{}{
		"type":    "question_revealed",
		"payload": room.GetStateSnapshot(),
	})

	// Leaderboard update
	uc.hub.BroadcastToRoom(roomID, map[string]interface{}{
		"type":    "leaderboard_update",
		"payload": room.GetLeaderboard(),
	})

	return nil
}

// GetRoom retorna info da sala (para HTTP).
func (uc *GameUseCases) GetRoom(ctx context.Context, roomID string) (*game.Room, error) {
	room, err := uc.gameRepo.FindRoomByID(roomID)
	if err != nil {
		return nil, err
	}
	return room, nil
}
