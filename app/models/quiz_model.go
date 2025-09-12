package models

import (
	"database/sql"

	"github.com/google/uuid"
)

type Quiz struct {
	ID          uuid.UUID      `json:"id,omitempty"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Difficulty  string         `json:"difficulty"`
	TimeLimit   sql.NullString `json:"time_limit"`
	CreatedBy   string         `json:"created_by"`
	Questions   []Question     `json:"questions"`
}
