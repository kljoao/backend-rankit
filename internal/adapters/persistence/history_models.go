package persistence

import "time"

// Modelos de persistência para Histórico (DTOs do banco)

type RoomHistory struct {
	ID                string    `json:"id"`
	RoomID            string    `json:"roomId"` // ID da sessão WS
	TeacherID         string    `json:"teacherId"`
	QuizID            string    `json:"quizId"`
	QuizTitleSnapshot string    `json:"quizTitleSnapshot"`
	Status            string    `json:"status"`
	TotalQuestions    int       `json:"totalQuestions"`
	StartedAt         time.Time `json:"startedAt"`
	FinishedAt        time.Time `json:"finishedAt"`
	CreatedAt         time.Time `json:"createdAt"`

	// Agregados carregados opcionalmente
	Players   []RoomPlayer   `json:"players,omitempty"`
	Questions []RoomQuestion `json:"questions,omitempty"`
}

type RoomPlayer struct {
	ID              string    `json:"id"`
	RoomHistoryID   string    `json:"roomHistoryId"`
	PlayerRuntimeID string    `json:"playerRuntimeId"`
	Nickname        string    `json:"nickname"`
	Score           int       `json:"score"`
	CorrectCount    int       `json:"correctCount"`
	WrongCount      int       `json:"wrongCount"`
	CreatedAt       time.Time `json:"createdAt"`
}

type RoomQuestion struct {
	ID             string    `json:"id"`
	RoomHistoryID  string    `json:"roomHistoryId"`
	QuestionIndex  int       `json:"questionIndex"`
	QuestionID     string    `json:"questionId"`
	PromptSnapshot string    `json:"promptSnapshot"`
	CorrectIndex   int       `json:"correctIndex"`
	CountA         int       `json:"countA"`
	CountB         int       `json:"countB"`
	CountC         int       `json:"countC"`
	CountD         int       `json:"countD"`
	CorrectCount   int       `json:"correctCount"`
	CreatedAt      time.Time `json:"createdAt"`
}

type RoomAnswer struct {
	ID            string    `json:"id"`
	RoomHistoryID string    `json:"roomHistoryId"`
	QuestionIndex int       `json:"questionIndex"`
	RoomPlayerID  string    `json:"roomPlayerId"`
	SelectedIndex int       `json:"selectedIndex"`
	IsCorrect     bool      `json:"isCorrect"`
	CreatedAt     time.Time `json:"createdAt"`
}
