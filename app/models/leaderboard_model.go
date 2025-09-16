package models

import "github.com/google/uuid"

type UserLeaderboardEntry struct {
	UserID     uuid.UUID `json:"user_id"`
	Username   string    `json:"username"`
	ImageURL   *string   `json:"image_url,omitempty"`
	TotalScore int       `json:"total_score"`
}

type StudyGroupLeaderboardEntry struct {
	GroupID     uuid.UUID `json:"group_id"`
	Name        string    `json:"name"`
	MemberCount int       `json:"member_count"`
	TotalScore  int       `json:"total_score"`
}
