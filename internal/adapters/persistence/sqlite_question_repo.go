package persistence

import (
	"context"
	"database/sql"
	"rankit/internal/domain/quiz"
)

type SQLiteQuestionRepository struct {
	db *sql.DB
}

func NewSQLiteQuestionRepository(db *sql.DB) *SQLiteQuestionRepository {
	return &SQLiteQuestionRepository{db: db}
}

func (r *SQLiteQuestionRepository) Save(ctx context.Context, q *quiz.Question) error {
	// Chama lógica reutilizável ou implementa direto
	query := `
		INSERT INTO questions (id, quiz_id, prompt, option_a, option_b, option_c, option_d, correct_index, sort_order, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		q.ID, q.QuizID, q.Prompt,
		q.OptionA, q.OptionB, q.OptionC, q.OptionD,
		q.CorrectIndex, q.SortOrder,
		q.CreatedAt, q.UpdatedAt,
	)
	return err
}

func (r *SQLiteQuestionRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM questions WHERE id = ?", id)
	return err
}

func (r *SQLiteQuestionRepository) FindByQuizID(ctx context.Context, quizID string) ([]*quiz.Question, error) {
	query := `
		SELECT id, quiz_id, prompt, option_a, option_b, option_c, option_d, correct_index, sort_order, created_at, updated_at
		FROM questions
		WHERE quiz_id = ?
		ORDER BY sort_order ASC
	`
	rows, err := r.db.QueryContext(ctx, query, quizID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []*quiz.Question
	for rows.Next() {
		var q quiz.Question
		if err := rows.Scan(
			&q.ID, &q.QuizID, &q.Prompt,
			&q.OptionA, &q.OptionB, &q.OptionC, &q.OptionD,
			&q.CorrectIndex, &q.SortOrder,
			&q.CreatedAt, &q.UpdatedAt,
		); err != nil {
			return nil, err
		}
		questions = append(questions, &q)
	}
	return questions, nil
}

func (r *SQLiteQuestionRepository) ReorderQuestions(ctx context.Context, quizID string, questions []*quiz.Question) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, "UPDATE questions SET sort_order = ? WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, q := range questions {
		// Segurança: update where quiz_id = ? (se quiser garantir, mas id é unique)
		if _, err := stmt.ExecContext(ctx, q.SortOrder, q.ID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Método extra para Update que não está na interface genérica mas é útil?
// A interface diz Save, Delete, etc. Se precisar Update, adicionamos na interface.
// Vamos adicionar Update na interface depois ou usar Save como upsert? Simplifique: ADICIONAR Update na interface QuestionRepo.
func (r *SQLiteQuestionRepository) Update(ctx context.Context, q *quiz.Question) error {
	query := `
		UPDATE questions 
		SET prompt = ?, option_a = ?, option_b = ?, option_c = ?, option_d = ?, correct_index = ?, updated_at = ?
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query,
		q.Prompt, q.OptionA, q.OptionB, q.OptionC, q.OptionD, q.CorrectIndex, q.UpdatedAt, q.ID,
	)
	return err
}
