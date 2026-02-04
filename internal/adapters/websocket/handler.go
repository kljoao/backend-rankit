package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"rankit/internal/application/usecases"

	"github.com/google/uuid"
)

// WebSocketHandler gerencia o upgrade e o roteamento de eventos.
type WebSocketHandler struct {
	hub    *Hub
	gameUC *usecases.GameUseCases
}

func NewWebSocketHandler(hub *Hub, gameUC *usecases.GameUseCases) *WebSocketHandler {
	handler := &WebSocketHandler{
		hub:    hub,
		gameUC: gameUC,
	}

	// Registra o callback no Hub
	hub.EventHandler = handler.HandleEvent
	return handler
}

// HandleWS faz o upgrade da conexão HTTP para WebSocket.
func (h *WebSocketHandler) HandleWS(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("roomId")
	if roomID == "" {
		roomID = r.URL.Query().Get("room")
	}

	if roomID == "" {
		http.Error(w, "Room ID required (roomId)", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	sessionID := uuid.NewString()

	client := &Client{
		Hub:      h.hub,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		RoomID:   roomID,
		PlayerID: sessionID,
	}

	client.Hub.register <- client

	go client.writePump()
	go client.readPump()
}

// HandleEvent processa mensagens vindas dos clientes (Router de Eventos).
func (h *WebSocketHandler) HandleEvent(client *Client, msg Envelope) {
	switch msg.Type {
	case "join_room":
		var payload struct {
			Nickname string `json:"nickname"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err == nil {
			_, err := h.gameUC.JoinRoom(client.RoomID, payload.Nickname, client.PlayerID)
			if err != nil {
				h.sendError(client.PlayerID, err.Error())
			}
		}

	case "teacher_moderate_entry":
		var payload struct {
			TeacherID    string `json:"teacherId"`
			ConnectionID string `json:"connectionId"`
			Action       string `json:"action"` // ACCEPT | REJECT
		}
		if err := json.Unmarshal(msg.Payload, &payload); err == nil {
			if err := h.gameUC.ModerateEntry(client.RoomID, payload.TeacherID, payload.ConnectionID, payload.Action); err != nil {
				h.sendError(client.PlayerID, err.Error())
			}
		}

	case "teacher_kick_player":
		var payload struct {
			TeacherID    string `json:"teacherId"`
			ConnectionID string `json:"connectionId"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err == nil {
			if err := h.gameUC.KickPlayer(client.RoomID, payload.TeacherID, payload.ConnectionID); err != nil {
				h.sendError(client.PlayerID, err.Error())
			}
		}

	case "teacher_open_question":
		var payload struct {
			TeacherID string `json:"teacherId"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err == nil {
			if err := h.gameUC.OpenQuestion(client.RoomID, payload.TeacherID); err != nil {
				h.sendError(client.PlayerID, err.Error())
			}
		}

	case "submit_answer":
		var payload struct {
			AnswerIndex int `json:"answerIndex"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err == nil {
			if err := h.gameUC.SubmitAnswer(client.RoomID, client.PlayerID, payload.AnswerIndex); err != nil {
				h.sendError(client.PlayerID, err.Error())
			}
		}

	case "teacher_reveal":
		var payload struct {
			TeacherID string `json:"teacherId"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err == nil {
			if err := h.gameUC.RevealQuestion(client.RoomID, payload.TeacherID); err != nil {
				h.sendError(client.PlayerID, err.Error())
			}
		}

	default:
		log.Printf("Evento desconhecido: %s", msg.Type)
	}
}

func (h *WebSocketHandler) sendError(playerID, errorMsg string) {
	h.hub.SendToPlayer(playerID, map[string]interface{}{
		"type":    "error",
		"payload": errorMsg,
	})
}
