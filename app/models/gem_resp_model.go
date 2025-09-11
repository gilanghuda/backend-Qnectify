package models

type AiResp struct {
	Questions []struct {
		Question      string   `json:"question"`
		Options       []string `json:"options"`
		CorrectAnswer string   `json:"correct_answer"`
	} `json:"questions"`
}
