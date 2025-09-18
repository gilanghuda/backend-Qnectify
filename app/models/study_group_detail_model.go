package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type StudyGroupMember struct {
	UserID   uuid.UUID      `json:"user_id"`
	Username string         `json:"username"`
	ImageURL sql.NullString `json:"image_url,omitempty"`
	JoinedAt time.Time      `json:"joined_at"`
}

type StudyGroupMemberScore struct {
	UserID     uuid.UUID      `json:"user_id"`
	Username   string         `json:"username"`
	ImageURL   sql.NullString `json:"image_url,omitempty"`
	TotalScore int            `json:"total_score"`
}

type StudyGroupDetail struct {
	Group       StudyGroup              `json:"group"`
	Members     []StudyGroupMember      `json:"members"`
	Leaderboard []StudyGroupMemberScore `json:"leaderboard"`
}
