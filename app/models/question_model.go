package models

import "github.com/google/uuid"

type Question struct {
	ID          uuid.UUID `json:"id,omitempty"`
	QuizID      uuid.UUID `json:"quiz_id,omitempty"`
	Question    string    `json:"question_text"`
	Options     []Option  `json:"options"`
	Explanation string    `json:"explanation"`
}
