package quiz

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEnunciadoObrigatorio = errors.New("o enunciado (prompt) é obrigatório")
	ErrAlternativaVazia     = errors.New("todas as 4 alternativas devem ser preenchidas")
	ErrIndiceInvalido       = errors.New("o índice da resposta correta deve ser entre 0 e 3")
)

// Question representa uma pergunta de múltipla escolha.
type Question struct {
	ID           string    `json:"id"`
	QuizID       string    `json:"quizId"`
	Prompt       string    `json:"prompt"`       // Enunciado
	OptionA      string    `json:"optionA"`      // 0
	OptionB      string    `json:"optionB"`      // 1
	OptionC      string    `json:"optionC"`      // 2
	OptionD      string    `json:"optionD"`      // 3
	CorrectIndex int       `json:"correctIndex"` // 0..3
	SortOrder    int       `json:"sortOrder"`    // Ordem na lista
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// NewQuestion cria uma nova pergunta.
func NewQuestion(quizID, prompt, optA, optB, optC, optD string, correctIndex, order int) (*Question, error) {
	q := &Question{
		ID:           uuid.NewString(),
		QuizID:       quizID,
		Prompt:       prompt,
		OptionA:      optA,
		OptionB:      optB,
		OptionC:      optC,
		OptionD:      optD,
		CorrectIndex: correctIndex,
		SortOrder:    order,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := q.Validate(); err != nil {
		return nil, err
	}

	return q, nil
}

// Validate verifica se a pergunta é válida.
func (q *Question) Validate() error {
	if q.Prompt == "" {
		return ErrEnunciadoObrigatorio
	}
	if q.OptionA == "" || q.OptionB == "" || q.OptionC == "" || q.OptionD == "" {
		return ErrAlternativaVazia
	}
	if q.CorrectIndex < 0 || q.CorrectIndex > 3 {
		return ErrIndiceInvalido
	}
	return nil
}

// Update atualiza os dados da pergunta.
func (q *Question) Update(prompt, optA, optB, optC, optD string, correctIndex int) error {
	q.Prompt = prompt
	q.OptionA = optA
	q.OptionB = optB
	q.OptionC = optC
	q.OptionD = optD
	q.CorrectIndex = correctIndex
	q.UpdatedAt = time.Now()

	return q.Validate()
}
