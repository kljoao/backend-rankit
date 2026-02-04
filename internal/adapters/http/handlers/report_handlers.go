package handlers

import (
	"encoding/json"
	"net/http"
	"rankit/internal/adapters/http/middlewares"
	"rankit/internal/application/usecases"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type ReportHandler struct {
	historyUC *usecases.HistoryUseCases
}

func NewReportHandler(historyUC *usecases.HistoryUseCases) *ReportHandler {
	return &ReportHandler{historyUC: historyUC}
}

// ListRooms godoc
// @Summary Histórico de Salas
// @Description Lista salas finalizadas do professor logado.
// @Tags Reports
// @Produce json
// @Param page query int false "Página (default 1)"
// @Param limit query int false "Limite (default 20)"
// @Success 200 {array} history.RoomHistory
// @Security BearerAuth
// @Router /reports/rooms [get]
func (h *ReportHandler) ListRooms(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middlewares.UserIDKey).(string)

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 20
	}

	rooms, err := h.historyUC.ListRooms(r.Context(), userID, page, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(rooms)
}

// GetRoomDetail godoc
// @Summary Detalhe da Sala
// @Description Retorna detalhes completos de uma sala, incluindo ranking e stats.
// @Tags Reports
// @Produce json
// @Param id path string true "History Room ID"
// @Success 200 {object} history.RoomHistory
// @Failure 404 "Sala não encontrada"
// @Security BearerAuth
// @Router /reports/rooms/{id} [get]
func (h *ReportHandler) GetRoomDetail(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middlewares.UserIDKey).(string)
	id := chi.URLParam(r, "id")

	room, err := h.historyUC.GetRoomDetail(r.Context(), id, userID)
	if err != nil {
		// Tratamento de erro simples: assume 404 se não encontrado ou não autorizado
		http.Error(w, "Sala não encontrada ou acesso negado", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(room)
}

// GetQuizStats godoc
// @Summary Estatísticas do Quiz
// @Description Retorna métricas agregadas de um quiz.
// @Tags Reports
// @Produce json
// @Param id path string true "Quiz ID"
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /reports/quizzes/{id} [get]
func (h *ReportHandler) GetQuizStats(w http.ResponseWriter, r *http.Request) {
	_ = r.Context().Value(middlewares.UserIDKey).(string)
	quizID := chi.URLParam(r, "id")

	// TODO: Validar se quiz pertence ao teacher?
	// O usecase GetQuizStats atual é genérico.

	stats, err := h.historyUC.GetQuizStats(r.Context(), quizID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(stats)
}
