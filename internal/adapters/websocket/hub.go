package websocket

import (
	"encoding/json"
	"log"
	"sync"
)

// HubMessage envolve a mensagem e o cliente remetente.
type HubMessage struct {
	Client  *Client
	Content Envelope
}

// Hub implementa ports.RealTimeHub.
type Hub struct {
	clients    map[*Client]bool
	rooms      map[string]map[*Client]bool
	register   chan *Client
	unregister chan *Client

	// IncomingMsgs é o canal onde o Hub recebe comandos dos clientes
	IncomingMsgs chan HubMessage

	// Handler processa eventos de negócio (injetado via setter ou campo)
	EventHandler func(*Client, Envelope)

	// Mapeia PlayerID -> Client (para envio direto)
	playerSessions map[string]*Client

	mu sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		clients:        make(map[*Client]bool),
		rooms:          make(map[string]map[*Client]bool),
		playerSessions: make(map[string]*Client),
		IncomingMsgs:   make(chan HubMessage),
	}
}

// Implementação da interface RealTimeHub
func (h *Hub) BroadcastToRoom(roomID string, message interface{}) {
	bytes, err := json.Marshal(message)
	if err != nil {
		log.Println("Erro ao serializar broadcast:", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.rooms[roomID]; ok {
		for client := range clients {
			select {
			case client.Send <- bytes:
			default:
				close(client.Send)
				delete(h.clients, client)
				delete(clients, client)
			}
		}
	}
}

func (h *Hub) SendToPlayer(playerID string, message interface{}) {
	bytes, err := json.Marshal(message)
	if err != nil {
		log.Println("Erro ao serializar mensagem direta:", err)
		return
	}

	h.mu.RLock()
	client, ok := h.playerSessions[playerID]
	h.mu.RUnlock()

	if ok {
		select {
		case client.Send <- bytes:
		default:
			// Falha no envio
		}
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			if _, ok := h.rooms[client.RoomID]; !ok {
				h.rooms[client.RoomID] = make(map[*Client]bool)
			}
			h.rooms[client.RoomID][client] = true

			if client.PlayerID != "" {
				h.playerSessions[client.PlayerID] = client
			}
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				if clients, ok := h.rooms[client.RoomID]; ok {
					delete(clients, client)
					if len(clients) == 0 {
						delete(h.rooms, client.RoomID)
					}
				}
				if client.PlayerID != "" {
					delete(h.playerSessions, client.PlayerID)
				}
				close(client.Send)
			}
			h.mu.Unlock()

		case msg := <-h.IncomingMsgs:
			// Delega para o handler de negócio
			if h.EventHandler != nil {
				go h.EventHandler(msg.Client, msg.Content)
			}
		}
	}
}
