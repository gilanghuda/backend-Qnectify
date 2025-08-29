package models

import "github.com/google/uuid"

type Option struct {
	ID         uuid.UUID `json:"id,omitempty"`
	QuestionID uuid.UUID `json:"question_id,omitempty"`
	Content    string    `json:"content"`
	IsCorrect  bool      `json:"is_correct"`
}
