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
	// Configura o handler de eventos do Hub para usar este Controller
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
	// Query params para identificar sala e (opcionalmente) player
	roomID := r.URL.Query().Get("room")
	if roomID == "" {
		http.Error(w, "Room ID required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	// Gera ID de sessão se não tiver
	sessionID := uuid.NewString()

	client := &Client{
		Hub:      h.hub,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		RoomID:   roomID,
		PlayerID: sessionID, // Inicialmente sessionID, pode ser associado a Player real no Join
	}

	client.Hub.register <- client

	go client.writePump()
	go client.readPump()
}

// HandleEvent processa mensagens vindas dos clientes (Router de Eventos).
func (h *WebSocketHandler) HandleEvent(client *Client, msg Envelope) {
	// Parse do payload dependendo do tipo
	// msg.Type determina qual UseCase chamar.

	switch msg.Type {
	case "join_room":
		var payload struct {
			Nickname string `json:"nickname"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err == nil {
			// Adiciona player na sala
			// O sessionID é o PlayerID neste contexto (aluno anônimo)
			_, err := h.gameUC.JoinRoom(client.RoomID, payload.Nickname, client.PlayerID)
			if err != nil {
				h.sendError(client.PlayerID, err.Error())
			}
		}

	case "teacher_open_question":
		// Payload: {teacherId} - Idealmente viria do token JWT na conexão WS
		// Simplificação: TeacherID enviado no payload ou confiamos se souber o RoomID?
		// Vamos exigir teacherID no payload por enquanto, mas inseguro.
		// TODO: Autenticar WS com JWT.
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

	// ... NextQuestion, FinishRoom ...

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
