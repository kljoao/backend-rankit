package usecases

import (
	"context"
	"errors"
	"rankit/internal/domain/quiz"
	"rankit/internal/ports"
)

var (
	ErrNaoAutorizado     = errors.New("você não tem permissão para acessar este quiz")
	ErrQuizNaoEncontrado = errors.New("quiz não encontrado")
)

// ------ QUIZ METHODS ------

type QuizUseCases struct {
	quizRepo ports.QuizRepository
}

func NewQuizUseCases(quizRepo ports.QuizRepository) *QuizUseCases {
	return &QuizUseCases{quizRepo: quizRepo}
}

type CreateQuizInput struct {
	TeacherID   string
	Title       string
	Description string
	Subject     string
	Grade       string
}

func (uc *QuizUseCases) CreateQuiz(ctx context.Context, input CreateQuizInput) (*quiz.Quiz, error) {
	q, err := quiz.NewQuiz(input.TeacherID, input.Title, input.Description, input.Subject, input.Grade)
	if err != nil {
		return nil, err
	}

	if err := uc.quizRepo.Save(ctx, q); err != nil {
		return nil, err
	}

	return q, nil
}

func (uc *QuizUseCases) ListQuizzes(ctx context.Context, teacherID string) ([]*quiz.Quiz, error) {
	return uc.quizRepo.FindByTeacherID(ctx, teacherID)
}

func (uc *QuizUseCases) GetQuizByID(ctx context.Context, quizID, teacherID string) (*quiz.Quiz, error) {
	q, err := uc.quizRepo.FindByID(ctx, quizID)
	if err != nil {
		return nil, err
	}
	if q == nil {
		return nil, ErrQuizNaoEncontrado
	}

	// Validação de Ownership
	if q.TeacherID != teacherID {
		return nil, ErrNaoAutorizado
	}

	return q, nil
}

type UpdateQuizInput struct {
	QuizID      string
	TeacherID   string
	Title       string
	Description string
	Subject     string
	Grade       string
}

func (uc *QuizUseCases) UpdateQuiz(ctx context.Context, input UpdateQuizInput) (*quiz.Quiz, error) {
	q, err := uc.GetQuizByID(ctx, input.QuizID, input.TeacherID)
	if err != nil {
		return nil, err
	}

	// Domínio checa se pode editar (ex: se é DRAFT)
	if err := q.UpdateMetadata(input.Title, input.Description, input.Subject, input.Grade); err != nil {
		return nil, err
	}

	if err := uc.quizRepo.Update(ctx, q); err != nil {
		return nil, err
	}

	return q, nil
}

func (uc *QuizUseCases) DeleteQuiz(ctx context.Context, quizID, teacherID string) error {
	q, err := uc.GetQuizByID(ctx, quizID, teacherID)
	if err != nil {
		return err
	}

	if err := q.CanEdit(); err != nil {
		return err // Não deletar se publicado (se for regra de negócio)
	}

	return uc.quizRepo.Delete(ctx, quizID)
}

func (uc *QuizUseCases) PublishQuiz(ctx context.Context, quizID, teacherID string) (*quiz.Quiz, error) {
	q, err := uc.GetQuizByID(ctx, quizID, teacherID)
	if err != nil {
		return nil, err
	}

	// Verifica regras para publicar (tem perguntas validadas, etc)
	if err := q.Publish(); err != nil {
		return nil, err
	}

	if err := uc.quizRepo.Update(ctx, q); err != nil {
		return nil, err
	}

	return q, nil
}
