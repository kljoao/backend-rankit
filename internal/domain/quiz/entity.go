package quiz

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Estados do Quiz
const (
	StatusRascunho  = "DRAFT"
	StatusPublicado = "PUBLISHED"
)

var (
	ErrTituloObrigatorio     = errors.New("o título é obrigatório")
	ErrQuizPublicadoNaoEdita = errors.New("não é permitido editar um quiz publicado")
	ErrQuizSemPerguntas      = errors.New("o quiz deve ter pelo menos uma pergunta para ser publicado")
	ErrPerguntaInvalida      = errors.New("existem perguntas inválidas no quiz")
)

// Quiz representa um conjunto de perguntas criado por um professor.
type Quiz struct {
	ID          string     `json:"id"`
	TeacherID   string     `json:"teacherId"` // Owner
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	Subject     string     `json:"subject,omitempty"` // Disciplina (ex: História)
	Grade       string     `json:"grade,omitempty"`   // Série (ex: 7º Ano)
	Status      string     `json:"status"`            // DRAFT | PUBLISHED
	Questions   []Question `json:"questions,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

// NewQuiz cria um novo rascunho de quiz.
func NewQuiz(teacherID, title, description, subject, grade string) (*Quiz, error) {
	if title == "" {
		return nil, ErrTituloObrigatorio
	}

	now := time.Now()
	return &Quiz{
		ID:          uuid.NewString(),
		TeacherID:   teacherID,
		Title:       title,
		Description: description,
		Subject:     subject,
		Grade:       grade,
		Status:      StatusRascunho,
		CreatedAt:   now,
		UpdatedAt:   now,
		Questions:   []Question{},
	}, nil
}

// CanEdit verifica se o quiz pode ser alterado (apenas Rascunhos).
func (q *Quiz) CanEdit() error {
	if q.Status != StatusRascunho {
		return ErrQuizPublicadoNaoEdita
	}
	return nil
}

// Publish tenta alterar o status para PUBLISHED.
func (q *Quiz) Publish() error {
	if len(q.Questions) == 0 {
		return ErrQuizSemPerguntas
	}

	for _, question := range q.Questions {
		if err := question.Validate(); err != nil {
			return ErrPerguntaInvalida
		}
	}

	q.Status = StatusPublicado
	q.UpdatedAt = time.Now()
	return nil
}

// UpdateMetadata atualiza dados básicos do quiz.
func (q *Quiz) UpdateMetadata(title, description, subject, grade string) error {
	if err := q.CanEdit(); err != nil {
		return err
	}
	if title == "" {
		return ErrTituloObrigatorio
	}

	q.Title = title
	q.Description = description
	q.Subject = subject
	q.Grade = grade
	q.UpdatedAt = time.Now()
	return nil
}
