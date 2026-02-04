package teacher

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNomeObrigatorio  = errors.New("o nome é obrigatório")
	ErrEmailInvalido    = errors.New("o email é inválido")
	ErrSenhaCurta       = errors.New("a senha deve ter no mínimo 6 caracteres")
)

// Teacher representa um professor no sistema RankIt.
type Teacher struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Oculta o hash da senha no JSON
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// NewTeacher cria uma nova instância de Teacher com validações.
func NewTeacher(name, email, password string) (*Teacher, error) {
	if name == "" {
		return nil, ErrNomeObrigatorio
	}
	if !isEmailValid(email) {
		return nil, ErrEmailInvalido
	}
	if len(password) < 6 {
		return nil, ErrSenhaCurta
	}

	return &Teacher{
		ID:        uuid.NewString(), // Gera UUID v4
		Name:      name,
		Email:     email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		// PasswordHash deve ser definido externamente via serviço de hash
	}, nil
}

// SetPassword define o hash da senha.
func (t *Teacher) SetPassword(hash string) {
	t.PasswordHash = hash
}

// Validação simples de email usando regex.
func isEmailValid(email string) bool {
	// Regex simplificado para validação de email
	regex := `^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`
	match, _ := regexp.MatchString(regex, email)
	return match
}
