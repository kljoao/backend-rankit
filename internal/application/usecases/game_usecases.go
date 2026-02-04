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
	historyUC *HistoryUseCases
}

func NewGameUseCases(
	gameRepo ports.GameRepository,
	quizRepo ports.QuizRepository,
	hub ports.RealTimeHub,
	historyUC *HistoryUseCases,
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

	// Gera ID curto para a sala
	roomID := uuid.NewString()[:6]

	room := game.NewRoom(roomID, teacherID, q)
	if err := uc.gameRepo.SaveRoom(room); err != nil {
		return nil, err
	}

	return room, nil
}

// JoinRoom adiciona um aluno (solicita entrada).
func (uc *GameUseCases) JoinRoom(roomID, nickname, sessionID string) (*game.Player, error) {
	room, err := uc.gameRepo.FindRoomByID(roomID)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, errors.New("sala não encontrada")
	}

	// Tenta entrar (cai em pendente ou rejeita se iniciado)
	player, err := room.JoinRequest(sessionID, nickname)
	if err != nil {
		return nil, err
	}

	// Se o jogador JÁ estava em Players (reconectou), enviamos o estado e broadcast de volta
	if _, approved := room.Players[sessionID]; approved {
		uc.hub.BroadcastToRoom(roomID, map[string]interface{}{
			"type":    "player_joined",
			"payload": player,
		})
		uc.hub.SendToPlayer(sessionID, map[string]interface{}{
			"type":    "room_state",
			"payload": room.GetStateSnapshot(),
		})
		return player, nil
	}

	// Agora ele está em PENDING.
	// 1. Notifica a sala (Professor deve filtrar por type)
	uc.hub.BroadcastToRoom(roomID, map[string]interface{}{
		"type": "player_request_entry",
		"payload": map[string]string{
			"nickname":     nickname,
			"connectionId": sessionID,
		},
	})

	// 2. Avisa o aluno que está pendente
	uc.hub.SendToPlayer(sessionID, map[string]interface{}{
		"type":    "entry_pending",
		"payload": "Aguardando aprovação do professor",
	})

	return player, nil
}

// ModerateEntry aceita ou rejeita um aluno.
func (uc *GameUseCases) ModerateEntry(roomID, teacherID, targetConnectionID, action string) error {
	room, err := uc.gameRepo.FindRoomByID(roomID)
	if err != nil || room == nil {
		return errors.New("sala não encontrada")
	}
	if room.TeacherID != teacherID {
		return ErrNaoAutorizado
	}

	if action == "ACCEPT" {
		player, err := room.ApprovePlayer(targetConnectionID)
		if err != nil {
			return err
		}

		// Notifica sucesso: Player entrou de fato
		uc.hub.BroadcastToRoom(roomID, map[string]interface{}{
			"type":    "player_joined",
			"payload": player,
		})
		// Envia estado para o aluno liberado
		uc.hub.SendToPlayer(targetConnectionID, map[string]interface{}{
			"type":    "room_state",
			"payload": room.GetStateSnapshot(),
		})

	} else if action == "REJECT" {
		if err := room.RejectPlayer(targetConnectionID); err != nil {
			return err
		}
		// Avisa o aluno e desconecta (opcional)
		uc.hub.SendToPlayer(targetConnectionID, map[string]interface{}{
			"type":    "error",
			"payload": "Entrada negada pelo professor",
		})
	} else {
		return errors.New("ação inválida (use ACCEPT ou REJECT)")
	}

	return nil
}

// KickPlayer remove um jogador aprovado da sala.
func (uc *GameUseCases) KickPlayer(roomID, teacherID, targetConnectionID string) error {
	room, err := uc.gameRepo.FindRoomByID(roomID)
	if err != nil || room == nil {
		return errors.New("sala não encontrada")
	}
	if room.TeacherID != teacherID {
		return ErrNaoAutorizado
	}

	if err := room.RemovePlayer(targetConnectionID); err != nil {
		return err
	}

	// 1. Notifica o jogador expulso
	uc.hub.SendToPlayer(targetConnectionID, map[string]interface{}{
		"type":    "error",
		"payload": "Você foi removido da sala pelo professor",
	})

	// 2. Notifica a sala (opcional: player_left ou leaderboard_update)
	uc.hub.BroadcastToRoom(roomID, map[string]interface{}{
		"type":    "leaderboard_update",
		"payload": room.GetLeaderboard(),
	})

	return nil
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
