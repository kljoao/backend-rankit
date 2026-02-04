package handlers

import (
	"encoding/json"
	"net/http"
	"rankit/internal/adapters/http/middlewares"
	"rankit/internal/application/usecases"

	"github.com/go-chi/chi/v5"
)

type GameHandler struct {
	gameUC *usecases.GameUseCases
}

func NewGameHandler(gameUC *usecases.GameUseCases) *GameHandler {
	return &GameHandler{gameUC: gameUC}
}

// CreateRoom godoc
// @Summary Cria uma sala de jogo
// @Description Cria uma nova sala a partir de um quiz PUBLISHED.
// @Tags Rooms
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body map[string]string true "payload: {quizId: uuid}"
// @Success 201 {object} game.Room
// @Failure 400 "Quiz inválido"
// @Router /rooms [post]
func (h *GameHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middlewares.UserIDKey).(string)

	var input struct {
		QuizID string `json:"quizId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	room, err := h.gameUC.CreateRoom(r.Context(), userID, input.QuizID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(room)
}

// GetRoom godoc
// @Summary Obtém dados da sala
// @Tags Rooms
// @Produce json
// @Param id path string true "Room ID"
// @Success 200 {object} game.Room
// @Failure 404 "Sala não encontrada"
// @Router /rooms/{id} [get]
func (h *GameHandler) GetRoom(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "id")

	room, err := h.gameUC.GetRoom(r.Context(), roomID)
	if err != nil {
		http.Error(w, "Sala não encontrada", http.StatusNotFound)
		return
	}
	if room == nil {
		http.Error(w, "Sala não encontrada", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(room)
}
