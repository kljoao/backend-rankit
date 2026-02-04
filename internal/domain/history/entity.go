package history

import "time"

// RoomHistory representa o registro histórico de uma sala executada.
type RoomHistory struct {
	ID                string    `json:"id"`
	RoomID            string    `json:"roomId"`
	TeacherID         string    `json:"teacherId"`
	QuizID            string    `json:"quizId"`
	QuizTitleSnapshot string    `json:"quizTitleSnapshot"`
	Status            string    `json:"status"`
	TotalQuestions    int       `json:"totalQuestions"`
	StartedAt         time.Time `json:"startedAt"`
	FinishedAt        time.Time `json:"finishedAt"`
	CreatedAt         time.Time `json:"createdAt"`

	Players   []PlayerStats   `json:"players,omitempty"`
	Questions []QuestionStats `json:"questions,omitempty"`
	Answers   []PlayerAnswer  `json:"answers,omitempty"`
}

type PlayerStats struct {
	ID              string `json:"id"`
	RoomHistoryID   string `json:"roomHistoryId"`
	PlayerRuntimeID string `json:"playerRuntimeId"`
	Nickname        string `json:"nickname"`
	Score           int    `json:"score"`
	CorrectCount    int    `json:"correctCount"`
	WrongCount      int    `json:"wrongCount"`
}

type QuestionStats struct {
	ID             string `json:"id"`
	RoomHistoryID  string `json:"roomHistoryId"`
	QuestionIndex  int    `json:"questionIndex"`
	QuestionID     string `json:"questionId"`
	PromptSnapshot string `json:"promptSnapshot"`
	CorrectIndex   int    `json:"correctIndex"`
	CountA         int    `json:"countA"`
	CountB         int    `json:"countB"`
	CountC         int    `json:"countC"`
	CountD         int    `json:"countD"`
	CorrectCount   int    `json:"correctCount"`
}

type PlayerAnswer struct {
	ID            string `json:"id"`
	RoomHistoryID string `json:"roomHistoryId"`
	QuestionIndex int    `json:"questionIndex"`
	RoomPlayerID  string `json:"roomPlayerId"`
	SelectedIndex int    `json:"selectedIndex"` // -1 se não respondeu?
	IsCorrect     bool   `json:"isCorrect"`
}
