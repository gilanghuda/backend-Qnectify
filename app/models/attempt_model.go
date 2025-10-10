package models

import (
	"time"

	"github.com/google/uuid"
)

// Attempt represents a user's quiz attempt
type Attempt struct {
	ID             uuid.UUID `json:"id,omitempty"`
	QuizID         uuid.UUID `json:"quiz_id,omitempty"`
	UserID         uuid.UUID `json:"user_id,omitempty"`
	Score          int       `json:"score"`
	TotalQuestions int       `json:"total_questions"`
	SubmittedAt    time.Time `json:"submitted_at,omitempty"`
	IsCompleted    bool      `json:"is_completed"`
}

// AttemptAnswer stores a single question's selected option for an attempt
type AttemptAnswer struct {
	QuestionID       uuid.UUID `json:"question_id"`
	SelectedOptionID uuid.UUID `json:"selected_option_id"`
}

// AttemptDetail represents details returned for an attempt including quiz and answers
type AttemptDetail struct {
	Attempt      Attempt         `json:"attempt"`
	Quiz         Quiz            `json:"quiz"`
	Answers      []AttemptAnswer `json:"answers"`
	TotalCorrect int             `json:"total_correct"`
}
