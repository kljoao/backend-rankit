package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"rankit/internal/domain/history"
	"time"
)

type SQLiteHistoryRepository struct {
	db *sql.DB
}

func NewSQLiteHistoryRepository(db *sql.DB) *SQLiteHistoryRepository {
	return &SQLiteHistoryRepository{db: db}
}

// SaveHistory salva o histórico completo de uma sala (transactional).
func (r *SQLiteHistoryRepository) SaveHistory(ctx context.Context, h *history.RoomHistory) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Save Room History
	queryRoom := `
		INSERT INTO rooms_history (id, room_id, teacher_id, quiz_id, quiz_title_snapshot, status, total_questions, started_at, finished_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err = tx.ExecContext(ctx, queryRoom,
		h.ID, h.RoomID, h.TeacherID, h.QuizID, h.QuizTitleSnapshot,
		h.Status, h.TotalQuestions, h.StartedAt, h.FinishedAt, h.CreatedAt,
	)
	if err != nil {
		return err
	}

	// 2. Save Room Players
	queryPlayer := `
		INSERT INTO room_players (id, room_history_id, player_runtime_id, nickname, score, correct_count, wrong_count, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	for _, p := range h.Players {
		_, err = tx.ExecContext(ctx, queryPlayer,
			p.ID, h.ID, p.PlayerRuntimeID, p.Nickname, p.Score, p.CorrectCount, p.WrongCount, time.Now(),
		)
		if err != nil {
			return err
		}
	}

	// 3. Save Room Questions
	queryQuestion := `
		INSERT INTO room_questions (id, room_history_id, question_index, question_id, prompt_snapshot, correct_index, count_a, count_b, count_c, count_d, correct_count, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	for _, q := range h.Questions {
		_, err = tx.ExecContext(ctx, queryQuestion,
			q.ID, h.ID, q.QuestionIndex, q.QuestionID, q.PromptSnapshot, q.CorrectIndex,
			q.CountA, q.CountB, q.CountC, q.CountD, q.CorrectCount, time.Now(),
		)
		if err != nil {
			return err
		}
	}

	// 4. Save Room Answers
	queryAnswer := `
		INSERT INTO room_answers (id, room_history_id, question_index, room_player_id, selected_index, is_correct, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	for _, a := range h.Answers {
		_, err = tx.ExecContext(ctx, queryAnswer,
			a.ID, h.ID, a.QuestionIndex, a.RoomPlayerID, a.SelectedIndex, a.IsCorrect, time.Now(),
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// ListByTeacherID lista histórico paginado.
func (r *SQLiteHistoryRepository) ListByTeacherID(ctx context.Context, teacherID string, limit, offset int) ([]*history.RoomHistory, error) {
	query := `
		SELECT id, room_id, teacher_id, quiz_id, quiz_title_snapshot, status, total_questions, started_at, finished_at, created_at
		FROM rooms_history
		WHERE teacher_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.QueryContext(ctx, query, teacherID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var histories []*history.RoomHistory
	for rows.Next() {
		var h history.RoomHistory
		if err := rows.Scan(
			&h.ID, &h.RoomID, &h.TeacherID, &h.QuizID, &h.QuizTitleSnapshot,
			&h.Status, &h.TotalQuestions, &h.StartedAt, &h.FinishedAt, &h.CreatedAt,
		); err != nil {
			return nil, err
		}
		histories = append(histories, &h)
	}
	return histories, nil
}

// GetByID busca histórico detalhado.
func (r *SQLiteHistoryRepository) GetByID(ctx context.Context, id string) (*history.RoomHistory, error) {
	query := `
		SELECT id, room_id, teacher_id, quiz_id, quiz_title_snapshot, status, total_questions, started_at, finished_at, created_at
		FROM rooms_history
		WHERE id = ?
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var h history.RoomHistory
	if err := row.Scan(
		&h.ID, &h.RoomID, &h.TeacherID, &h.QuizID, &h.QuizTitleSnapshot,
		&h.Status, &h.TotalQuestions, &h.StartedAt, &h.FinishedAt, &h.CreatedAt,
	); err != nil {
		return nil, err
	}

	// Carrega Players
	pRows, err := r.db.QueryContext(ctx, "SELECT id, player_runtime_id, nickname, score, correct_count, wrong_count FROM room_players WHERE room_history_id = ?", h.ID)
	if err != nil {
		return nil, err
	}
	defer pRows.Close()

	for pRows.Next() {
		var p history.PlayerStats
		p.RoomHistoryID = h.ID
		if err := pRows.Scan(&p.ID, &p.PlayerRuntimeID, &p.Nickname, &p.Score, &p.CorrectCount, &p.WrongCount); err != nil {
			return nil, err
		}
		h.Players = append(h.Players, p)
	}

	// Carrega Questions Stats
	qRows, err := r.db.QueryContext(ctx, "SELECT id, question_index, question_id, correct_index, count_a, count_b, count_c, count_d, correct_count FROM room_questions WHERE room_history_id = ?", h.ID)
	if err != nil {
		return nil, err
	}
	defer qRows.Close()

	for qRows.Next() {
		var q history.QuestionStats
		q.RoomHistoryID = h.ID
		if err := qRows.Scan(&q.ID, &q.QuestionIndex, &q.QuestionID, &q.CorrectIndex, &q.CountA, &q.CountB, &q.CountC, &q.CountD, &q.CorrectCount); err != nil {
			return nil, err
		}
		h.Questions = append(h.Questions, q)
	}

	return &h, nil
}

// GetQuizStats retorna estatísticas agregadas de um quiz.
func (r *SQLiteHistoryRepository) GetQuizStats(ctx context.Context, quizID string) (map[string]interface{}, error) {
	// Query agregada (Corrigida e simplificada)
	query := `
		SELECT 
			COUNT(id) as total_rooms
		FROM rooms_history
		WHERE quiz_id = ?
	`
	row := r.db.QueryRowContext(ctx, query, quizID)

	var totalRooms int64
	// var avgQuestions float64 // Removido pois não estamos usando

	// Handle nulls if 0 rows
	var totalRoomsNull sql.NullInt64

	if err := row.Scan(&totalRoomsNull); err != nil {
		return nil, err
	}
	totalRooms = totalRoomsNull.Int64

	// Query avg participants
	queryParticipants := `
		SELECT COUNT(p.id) 
		FROM room_players p
		JOIN rooms_history h ON p.room_history_id = h.id
		WHERE h.quiz_id = ?
	`

	var totalParticipants int
	if err := r.db.QueryRowContext(ctx, queryParticipants, quizID).Scan(&totalParticipants); err != nil {
		// Log error or ignore
	}

	avgParticipants := 0.0
	if totalRooms > 0 {
		avgParticipants = float64(totalParticipants) / float64(totalRooms)
	}

	return map[string]interface{}{
		"quizId":          quizID,
		"totalRooms":      totalRooms,
		"avgParticipants": avgParticipants,
	}, nil
}

func toJson(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}
