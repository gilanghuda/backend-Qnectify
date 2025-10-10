package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Quiz struct {
	ID             uuid.UUID     `json:"id,omitempty"`
	Title          string        `json:"title"`
	Description    string        `json:"description"`
	Difficulty     string        `json:"difficulty"`
	TimeLimit      sql.NullInt64 `json:"time_limit"`
	CreatedBy      string        `json:"created_by"`
	Attempts       int           `json:"attempts,omitempty"`
	TotalQuestions *int          `json:"total_questions,omitempty"`
	Questions      []Question    `json:"questions"`
	CreatedAt      time.Time     `json:"created_at,omitempty"`
}
