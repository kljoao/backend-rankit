package usecases

import (
	"context"
	"rankit/internal/domain/game"
	"rankit/internal/domain/history"
	"rankit/internal/ports"
	"time"

	"github.com/google/uuid"
)

type HistoryUseCases struct {
	historyRepo ports.HistoryRepository
	gameRepo    ports.GameRepository
}

func NewHistoryUseCases(historyRepo ports.HistoryRepository, gameRepo ports.GameRepository) *HistoryUseCases {
	return &HistoryUseCases{
		historyRepo: historyRepo,
		gameRepo:    gameRepo,
	}
}

// ArchiveRoom converte uma sala de jogo em histórico persistente.
func (uc *HistoryUseCases) ArchiveRoom(ctx context.Context, room *game.Room) error {
	// Mapeia Game -> History
	h := &history.RoomHistory{
		ID:                uuid.NewString(),
		RoomID:            room.ID,
		TeacherID:         room.TeacherID,
		QuizID:            room.Quiz.ID,
		QuizTitleSnapshot: room.Quiz.Title,
		Status:            room.Status,
		TotalQuestions:    len(room.Quiz.Questions),
		StartedAt:         time.Now(), // Aproximação
		FinishedAt:        time.Now(),
		CreatedAt:         time.Now(),
	}

	players := room.GetLeaderboard()
	for _, p := range players {
		correct := p.Score / 10
		hP := history.PlayerStats{
			ID:              uuid.NewString(),
			RoomHistoryID:   h.ID,
			PlayerRuntimeID: p.ID,
			Nickname:        p.Nickname,
			Score:           p.Score,
			CorrectCount:    correct,
			WrongCount:      0,
		}
		h.Players = append(h.Players, hP)
	}

	// Não salvamos perguntas individuais por enquanto pois room in-memory não tem histórico.

	return uc.historyRepo.SaveHistory(ctx, h)
}

// ------ REPORT METHODS ------

func (uc *HistoryUseCases) ListRooms(ctx context.Context, teacherID string, page, limit int) ([]*history.RoomHistory, error) {
	offset := (page - 1) * limit
	return uc.historyRepo.ListByTeacherID(ctx, teacherID, limit, offset)
}

func (uc *HistoryUseCases) GetRoomDetail(ctx context.Context, id, teacherID string) (*history.RoomHistory, error) {
	h, err := uc.historyRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	// Valida ownership
	if h.TeacherID != teacherID {
		return nil, ErrNaoAutorizado // Removeu 'ports.'
	}
	return h, nil
}

func (uc *HistoryUseCases) GetQuizStats(ctx context.Context, quizID string) (map[string]interface{}, error) {
	return uc.historyRepo.GetQuizStats(ctx, quizID)
}
