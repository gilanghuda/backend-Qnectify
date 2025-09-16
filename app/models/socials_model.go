package models

import "github.com/google/uuid"

type Social struct {
	ID         uuid.UUID `json:"id"`
	FollowerID uuid.UUID `json:"follower_id"`
	Following  uuid.UUID `json:"following"`
}
