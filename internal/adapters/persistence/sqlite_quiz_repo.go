package persistence

import (
	"context"
	"database/sql"
	"errors"
	"rankit/internal/domain/quiz"
)

type SQLiteQuizRepository struct {
	db *sql.DB
}

func NewSQLiteQuizRepository(db *sql.DB) *SQLiteQuizRepository {
	return &SQLiteQuizRepository{db: db}
}

// ------ QUIZ METHODS ------

func (r *SQLiteQuizRepository) Save(ctx context.Context, q *quiz.Quiz) error {
	query := `
		INSERT INTO quizzes (id, teacher_id, title, description, subject, grade, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		q.ID, q.TeacherID, q.Title, q.Description, q.Subject, q.Grade, q.Status, q.CreatedAt, q.UpdatedAt,
	)
	return err
}

func (r *SQLiteQuizRepository) Update(ctx context.Context, q *quiz.Quiz) error {
	query := `
		UPDATE quizzes 
		SET title = ?, description = ?, subject = ?, grade = ?, status = ?, updated_at = ?
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query,
		q.Title, q.Description, q.Subject, q.Grade, q.Status, q.UpdatedAt, q.ID,
	)
	return err
}

func (r *SQLiteQuizRepository) FindByID(ctx context.Context, id string) (*quiz.Quiz, error) {
	query := `
		SELECT id, teacher_id, title, description, subject, grade, status, created_at, updated_at
		FROM quizzes WHERE id = ?
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var q quiz.Quiz
	var desc, subj, grade sql.NullString // Handle nullables

	err := row.Scan(
		&q.ID, &q.TeacherID, &q.Title, &desc, &subj, &grade,
		&q.Status, &q.CreatedAt, &q.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Not found
		}
		return nil, err
	}

	q.Description = desc.String
	q.Subject = subj.String
	q.Grade = grade.String

	// Busca Questions associadas
	questions, err := r.FindByQuizID(ctx, q.ID)
	if err != nil {
		return nil, err
	}
	q.Questions = make([]quiz.Question, len(questions))
	for i, ptr := range questions {
		q.Questions[i] = *ptr
	}

	return &q, nil
}

func (r *SQLiteQuizRepository) FindByTeacherID(ctx context.Context, teacherID string) ([]*quiz.Quiz, error) {
	query := `
		SELECT id, teacher_id, title, description, subject, grade, status, created_at, updated_at
		FROM quizzes WHERE teacher_id = ? ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, teacherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var quizzes []*quiz.Quiz
	for rows.Next() {
		var q quiz.Quiz
		var desc, subj, grade sql.NullString

		if err := rows.Scan(
			&q.ID, &q.TeacherID, &q.Title, &desc, &subj, &grade,
			&q.Status, &q.CreatedAt, &q.UpdatedAt,
		); err != nil {
			return nil, err
		}
		q.Description = desc.String
		q.Subject = subj.String
		q.Grade = grade.String

		// NOTA: Para listagem, geralmente não carregamos todas as questions para não pesar.
		// Deixamos questions vazio nesta query.
		quizzes = append(quizzes, &q)
	}
	return quizzes, nil
}

func (r *SQLiteQuizRepository) Delete(ctx context.Context, id string) error {
	// Cascade delete via FK (se configurado) ou manual.
	// SQLite suporta FK cascade se habilitado, mas vamos garantir na query.
	// Vamos confiar no ON DELETE CASCADE do schema.
	_, err := r.db.ExecContext(ctx, "DELETE FROM quizzes WHERE id = ?", id)
	return err
}

// ------ QUESTION METHODS (Implementação de QuestionRepository interface localmente ou via composition) ------

// FindByQuizID é usado tanto internamente quanto externamente.
func (r *SQLiteQuizRepository) FindByQuizID(ctx context.Context, quizID string) ([]*quiz.Question, error) {
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

func (r *SQLiteQuizRepository) SaveQuestion(ctx context.Context, q *quiz.Question) error {
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

func (r *SQLiteQuizRepository) DeleteQuestion(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM questions WHERE id = ?", id)
	return err
}

// UpdateQuestion atualiza uma pergunta existente
func (r *SQLiteQuizRepository) UpdateQuestion(ctx context.Context, q *quiz.Question) error {
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

// ReorderQuestions atualiza a ordem de várias perguntas.
func (r *SQLiteQuizRepository) ReorderQuestions(ctx context.Context, quizID string, questions []*quiz.Question) error {
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
		if _, err := stmt.ExecContext(ctx, q.SortOrder, q.ID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Adapter compliance check
// Precisamos garantir que SQLiteQuizRepository implemente ambas as interfaces de ports?
// Sim, ou retornar structs separados. Para simplicidade, vamos implementar métodos com nomes compatíveis
// e fazer o cast ou adapter.
//
// Interface methods in ports:
// QuizRepository: Save, FindByID, FindByTeacherID, Delete, Update
// QuestionRepository: Save, Delete, FindByQuizID, ReorderQuestions (Save -> chama SaveQuestion)
//
// Vamos ajustar os nomes acima para bater com interface ou criar wrappers.
//
// Ajuste: SaveQuestion -> Save (precisa de assinatura diferente? Go não suporta overload).
// Solução: Teremos 2 structs ou métodos específicos.
//
// Vamos manter SQLiteQuizRepository implementando QuizRepository.
// E criar SQLiteQuestionRepository implementando QuestionRepository.
