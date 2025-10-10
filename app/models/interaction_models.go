package models

import (
	"time"

	"github.com/google/uuid"
)

type Like struct {
	ID        uuid.UUID `json:"id,omitempty"`
	QuizID    uuid.UUID `json:"quiz_id,omitempty"`
	LikedBy   uuid.UUID `json:"liked_by,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

type Comment struct {
	ID          uuid.UUID `json:"id,omitempty"`
	QuizID      uuid.UUID `json:"quiz_id,omitempty"`
	Content     string    `json:"content"`
	CommenterBy uuid.UUID `json:"commenter_by,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
}
