package security

import (
	"golang.org/x/crypto/bcrypt"
)

// BcryptHasher implementa a interface PasswordHasher usando bcrypt.
type BcryptHasher struct{}

// NewBcryptHasher cria uma nova instância de BcryptHasher.
func NewBcryptHasher() *BcryptHasher {
	return &BcryptHasher{}
}

// HashPassword gera um hash seguro da senha.
func (h *BcryptHasher) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// ComparePassword compara uma senha em texto plano com um hash.
func (h *BcryptHasher) ComparePassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
