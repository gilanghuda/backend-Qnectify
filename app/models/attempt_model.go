package models

import (
	"time"

	"github.com/google/uuid"
)

type Attempt struct {
	ID             uuid.UUID `json:"id,omitempty"`
	QuizID         uuid.UUID `json:"quiz_id,omitempty"`
	UserID         uuid.UUID `json:"user_id,omitempty"`
	Score          int       `json:"score"`
	TotalQuestions int       `json:"total_questions"`
	SubmittedAt    time.Time `json:"submitted_at"`
	IsCompleted    bool      `json:"is_completed"`
}
