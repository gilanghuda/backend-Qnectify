package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gilanghuda/backend-Quizzo/app/models"
)

func GenerateQuizFromContent(content []byte, question_count int, difficulty string) (interface{}, error) {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GOOGLE_API_KEY not set")
	}

	promptText := fmt.Sprintf(`Buatkan %d soal pilihan ganda dengan tingkat kesulitan %s berdasarkan materi berikut. 
Format output sebagai JSON dengan struktur:
{
  "questions": [
    {
      "question": "Pertanyaan",
      "options": ["A. Option1", "B. Option2", "C. Option3", "D. Option4"],
      "correct_answer": "A",
    }
  ]
}

Materi:
%s`, question_count, difficulty, string(content))

	payload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{
						"text": promptText,
					},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature":     0.7,
			"maxOutputTokens": 8127,
		},
	}

	payloadBytes, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key="+apiKey, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	log.Println("Gemini raw response:", string(respBody))

	var geminiResp models.GeminiResponse

	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		return nil, err
	}

	result := ""
	if len(geminiResp.Candidates) > 0 && len(geminiResp.Candidates[0].Content.Parts) > 0 {
		result = geminiResp.Candidates[0].Content.Parts[0].Text
	} else {
		return map[string]interface{}{
			"raw_response": string(respBody),
			"result":       "",
			"error":        "No candidates or parts found in Gemini response",
		}, nil
	}

	return map[string]interface{}{
		"result": result,
	}, nil
}
