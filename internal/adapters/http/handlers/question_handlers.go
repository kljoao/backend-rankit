package handlers

import (
	"encoding/json"
	"net/http"
	"rankit/internal/adapters/http/middlewares"
	"rankit/internal/application/usecases"
	"rankit/internal/domain/quiz"

	"github.com/go-chi/chi/v5"
)

type QuestionHandler struct {
	questionUC *usecases.QuestionUseCases
}

func NewQuestionHandler(questionUC *usecases.QuestionUseCases) *QuestionHandler {
	return &QuestionHandler{questionUC: questionUC}
}

// AddQuestion godoc
// @Summary Adiciona pergunta ao quiz
// @Description Adiciona uma nova pergunta ao quiz em DRAFT.
// @Tags Questions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID do Quiz"
// @Param body body usecases.AddQuestionInput true "Dados da Pergunta"
// @Success 201 {object} quiz.Question
// @Failure 400 "Dados inválidos"
// @Router /quizzes/{id}/questions [post]
func (h *QuestionHandler) AddQuestion(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middlewares.UserIDKey).(string)
	quizID := chi.URLParam(r, "id")

	var input usecases.AddQuestionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	input.QuizID = quizID
	input.TeacherID = userID

	q, err := h.questionUC.AddQuestion(r.Context(), input)
	if err != nil {
		if err == quiz.ErrEnunciadoObrigatorio || err == quiz.ErrAlternativaVazia || err == quiz.ErrIndiceInvalido {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err == quiz.ErrQuizPublicadoNaoEdita {
			http.Error(w, err.Error(), http.StatusConflict) // 409 é melhor aqui
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(q)
}

// UpdateQuestion godoc
// @Summary Atualiza pergunta
// @Tags Questions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID do Quiz"
// @Param questionId path string true "ID da Pergunta"
// @Param body body usecases.UpdateQuestionInput true "Dados da Pergunta"
// @Success 200 {object} quiz.Question
// @Router /quizzes/{id}/questions/{questionId} [put]
func (h *QuestionHandler) UpdateQuestion(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middlewares.UserIDKey).(string)
	quizID := chi.URLParam(r, "id")
	questionID := chi.URLParam(r, "questionId")

	var input usecases.UpdateQuestionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	input.QuizID = quizID
	input.QuestionID = questionID
	input.TeacherID = userID

	q, err := h.questionUC.UpdateQuestion(r.Context(), input)
	if err != nil {
		// Mesmos erros...
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(q)
}

// RemoveQuestion godoc
// @Summary Remove pergunta
// @Tags Questions
// @Security BearerAuth
// @Param id path string true "ID do Quiz"
// @Param questionId path string true "ID da Pergunta"
// @Success 204 "No Content"
// @Router /quizzes/{id}/questions/{questionId} [delete]
func (h *QuestionHandler) RemoveQuestion(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middlewares.UserIDKey).(string)
	quizID := chi.URLParam(r, "id")
	questionID := chi.URLParam(r, "questionId")

	if err := h.questionUC.RemoveQuestion(r.Context(), quizID, questionID, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ReorderQuestions godoc
// @Summary Reordena perguntas
// @Tags Questions
// @Accept json
// @Security BearerAuth
// @Param id path string true "ID do Quiz"
// @Param body body map[string][]string true "Lista com order (Ids)"
// @Success 200 "OK"
// @Router /quizzes/{id}/questions/reorder [post]
func (h *QuestionHandler) ReorderQuestions(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middlewares.UserIDKey).(string)
	quizID := chi.URLParam(r, "id")

	var input struct {
		Order []string `json:"order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if err := h.questionUC.ReorderQuestions(r.Context(), quizID, userID, input.Order); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
