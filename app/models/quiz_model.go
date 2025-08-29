package models

type Quiz struct {
	ID         uint       `json:"id,omitempty"`
	Title      string     `json:"title"`
	Difficulty string     `json:"difficulty"`
	TimeLimit  string     `json:"time_limit"`
	CreatedBy  string     `json:"created_by"`
	Questions  []Question `json:"questions"`
}
