package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID      `json:"id" validate:"required,uuid"`
	Username     string         `json:"username" validate:"required,lte=50"`
	Email        string         `json:"email" validate:"required,email,lte=255"`
	PasswordHash string         `json:"password_hash,omitempty" validate:"required,lte=255"`
	ExpPoints    sql.NullString `json:"exp_point,omitempty" validate:"omitempty,lte=25"`
	UserRole     string         `json:"user_role" validate:"required,lte=25"`
	ImageURL     sql.NullString `json:"image_url,omitempty" validate:"omitempty,lte=255"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}
