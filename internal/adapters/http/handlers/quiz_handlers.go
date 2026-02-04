package handlers

import (
	"encoding/json"
	"net/http"
	"rankit/internal/adapters/http/middlewares"
	"rankit/internal/application/usecases"
	"rankit/internal/domain/quiz"

	"github.com/go-chi/chi/v5"
)

type QuizHandler struct {
	quizUC *usecases.QuizUseCases
}

func NewQuizHandler(quizUC *usecases.QuizUseCases) *QuizHandler {
	return &QuizHandler{quizUC: quizUC}
}

// CreateQuiz godoc
// @Summary Cria um novo quiz (Rascunho)
// @Description Cria um quiz vinculado ao professor logado. Status inicial DRAFT.
// @Tags Quizzes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body usecases.CreateQuizInput true "Dados do Quiz"
// @Success 201 {object} quiz.Quiz
// @Failure 401 {object} map[string]string "Não autorizado"
// @Failure 400 {object} map[string]string "Erro de validação"
// @Failure 500 {object} map[string]string "Erro interno"
// @Router /quizzes [post]
func (h *QuizHandler) CreateQuiz(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middlewares.UserIDKey).(string)

	var input usecases.CreateQuizInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	input.TeacherID = userID

	q, err := h.quizUC.CreateQuiz(r.Context(), input)
	if err != nil {
		if err == quiz.ErrTituloObrigatorio {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(q)
}

// ListQuizzes godoc
// @Summary Lista quizzes do professor
// @Description Retorna todos os quizzes do professor logado.
// @Tags Quizzes
// @Produce json
// @Security BearerAuth
// @Success 200 {array} quiz.Quiz
// @Router /quizzes [get]
func (h *QuizHandler) ListQuizzes(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middlewares.UserIDKey).(string)

	quizzes, err := h.quizUC.ListQuizzes(r.Context(), userID)
	if err != nil {
		http.Error(w, "Erro ao listar quizzes", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(quizzes)
}

// GetQuiz godoc
// @Summary Detalha um quiz
// @Description Retorna dados do quiz e suas perguntas.
// @Tags Quizzes
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID do Quiz"
// @Success 200 {object} quiz.Quiz
// @Failure 404 {object} map[string]string "Não encontrado"
// @Router /quizzes/{id} [get]
func (h *QuizHandler) GetQuiz(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middlewares.UserIDKey).(string)
	quizID := chi.URLParam(r, "id")

	q, err := h.quizUC.GetQuizByID(r.Context(), quizID, userID)
	if err != nil {
		if err == usecases.ErrQuizNaoEncontrado {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err == usecases.ErrNaoAutorizado {
			http.Error(w, err.Error(), http.StatusNotFound) // 404 para não vazar
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(q)
}

// UpdateQuiz godoc
// @Summary Atualiza um quiz (Rascunho)
// @Description Atualiza título, descrição, etc. Apenas se DRAFT.
// @Tags Quizzes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID do Quiz"
// @Param body body usecases.UpdateQuizInput true "Dados novos"
// @Success 200 {object} quiz.Quiz
// @Failure 400 {object} map[string]string "Erro validação ou estao já publicado"
// @Router /quizzes/{id} [put]
func (h *QuizHandler) UpdateQuiz(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middlewares.UserIDKey).(string)
	quizID := chi.URLParam(r, "id")

	var input usecases.UpdateQuizInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	input.QuizID = quizID
	input.TeacherID = userID

	q, err := h.quizUC.UpdateQuiz(r.Context(), input)
	if err != nil {
		if err == quiz.ErrQuizPublicadoNaoEdita {
			http.Error(w, err.Error(), http.StatusBadRequest) // ou 409
			return
		}
		if err == usecases.ErrQuizNaoEncontrado {
			http.Error(w, "Quiz não encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(q)
}

// DeleteQuiz godoc
// @Summary Remove um quiz
// @Description Remove quiz e perguntas. Apenas se DRAFT.
// @Tags Quizzes
// @Security BearerAuth
// @Param id path string true "ID do Quiz"
// @Success 204 "No Content"
// @Failure 400 "Não pode deletar quiz publicado"
// @Router /quizzes/{id} [delete]
func (h *QuizHandler) DeleteQuiz(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middlewares.UserIDKey).(string)
	quizID := chi.URLParam(r, "id")

	if err := h.quizUC.DeleteQuiz(r.Context(), quizID, userID); err != nil {
		if err == quiz.ErrQuizPublicadoNaoEdita {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err == usecases.ErrQuizNaoEncontrado {
			http.Error(w, "Quiz não encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// PublishQuiz godoc
// @Summary Publica um quiz
// @Description Altera status para PUBLISHED. Valida perguntas.
// @Tags Quizzes
// @Security BearerAuth
// @Param id path string true "ID do Quiz"
// @Success 200 {object} quiz.Quiz
// @Failure 409 {object} map[string]string "Quiz inválido para publicação"
// @Router /quizzes/{id}/publish [post]
func (h *QuizHandler) PublishQuiz(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middlewares.UserIDKey).(string)
	quizID := chi.URLParam(r, "id")

	q, err := h.quizUC.PublishQuiz(r.Context(), quizID, userID)
	if err != nil {
		if err == quiz.ErrQuizSemPerguntas || err == quiz.ErrPerguntaInvalida {
			http.Error(w, err.Error(), http.StatusConflict) // 409 Conflict
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(q)
}
