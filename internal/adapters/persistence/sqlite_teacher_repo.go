package persistence

import (
	"context"
	"database/sql"
	"errors"
	"rankit/internal/domain/teacher"
	"rankit/internal/ports"
)

// SQLiteTeacherRepository implementa TeacherRepository para SQLite.
type SQLiteTeacherRepository struct {
	db *sql.DB
}

// NewSQLiteTeacherRepository cria uma nova instância do repositório.
func NewSQLiteTeacherRepository(db *sql.DB) ports.TeacherRepository {
	return &SQLiteTeacherRepository{db: db}
}

// Create insere um novo professor no banco.
func (r *SQLiteTeacherRepository) Create(ctx context.Context, t *teacher.Teacher) error {
	query := `
		INSERT INTO teachers (id, name, email, password_hash, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		t.ID,
		t.Name,
		t.Email,
		t.PasswordHash,
		t.CreatedAt,
		t.UpdatedAt,
	)
	return err
}

// FindByEmail busca um professor pelo email.
func (r *SQLiteTeacherRepository) FindByEmail(ctx context.Context, email string) (*teacher.Teacher, error) {
	query := `
		SELECT id, name, email, password_hash, created_at, updated_at
		FROM teachers
		WHERE email = ?
	`
	row := r.db.QueryRowContext(ctx, query, email)

	var t teacher.Teacher
	err := row.Scan(
		&t.ID,
		&t.Name,
		&t.Email,
		&t.PasswordHash,
		&t.CreatedAt,
		&t.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Não encontrado
		}
		return nil, err
	}

	return &t, nil
}

// FindByID busca um professor pelo ID.
func (r *SQLiteTeacherRepository) FindByID(ctx context.Context, id string) (*teacher.Teacher, error) {
	query := `
		SELECT id, name, email, password_hash, created_at, updated_at
		FROM teachers
		WHERE id = ?
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var t teacher.Teacher
	err := row.Scan(
		&t.ID,
		&t.Name,
		&t.Email,
		&t.PasswordHash,
		&t.CreatedAt,
		&t.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Não encontrado
		}
		return nil, err
	}

	return &t, nil
}
