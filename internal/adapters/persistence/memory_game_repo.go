package persistence

import (
	"errors"
	"rankit/internal/domain/game"
	"rankit/internal/ports"
	"sync"
)

// InMemoryGameRepository implementa GameRepository usando memória RAM.
type InMemoryGameRepository struct {
	rooms sync.Map // Map[string]*game.Room
}

func NewInMemoryGameRepository() ports.GameRepository {
	return &InMemoryGameRepository{}
}

func (r *InMemoryGameRepository) SaveRoom(room *game.Room) error {
	r.rooms.Store(room.ID, room)
	return nil
}

func (r *InMemoryGameRepository) FindRoomByID(id string) (*game.Room, error) {
	val, ok := r.rooms.Load(id)
	if !ok {
		return nil, nil // Não encontrado (sem erro)
	}

	room, ok := val.(*game.Room)
	if !ok {
		return nil, errors.New("erro de tipo no repositório de jogos")
	}
	return room, nil
}

func (r *InMemoryGameRepository) DeleteRoom(id string) error {
	r.rooms.Delete(id)
	return nil
}
