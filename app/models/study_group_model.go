package models

import (
	"time"

	"github.com/google/uuid"
)

type StudyGroup struct {
	ID          uuid.UUID `json:"id" `
	Name        string    `json:"name" validate:"required,lte=100"`
	Description *string   `json:"description,omitempty" validate:"lte=500"`
	InviteCode  *string   `json:"invite_code"`
	MemberCount int       `json:"member_count" `
	MaxMember   int       `json:"max_member" validate:"required,min=1"`
	IsPrivate   bool      `json:"is_private"`
	CreatedBy   uuid.UUID `json:"created_by" `
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" `
}
