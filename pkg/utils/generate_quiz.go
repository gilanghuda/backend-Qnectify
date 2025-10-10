package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gilanghuda/backend-Quizzo/app/models"
)

func GenerateQuiz(file io.Reader, question_count int, difficulty string) (interface{}, error) {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GOOGLE_API_KEY not set")
	}

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	mimeType := "application/octet-stream"
	if len(fileBytes) >= 512 {
		mimeType = http.DetectContentType(fileBytes[:512])
	} else if len(fileBytes) > 0 {
		mimeType = http.DetectContentType(fileBytes)
	}

	if strings.Contains(mimeType, "zip") || strings.Contains(mimeType, "officedocument") || strings.Contains(mimeType, "msword") {
		return nil, fmt.Errorf("file mime type %s not supported by Gemini. Please extract the archive and upload a supported file (PDF, TXT, HTML, image), or provide the extracted content as text", mimeType)
	}
	b64 := base64.StdEncoding.EncodeToString(fileBytes)
	promptText := fmt.Sprintf(`Berikan HANYA objek JSON mentah (raw JSON) yang valid berdasarkan materi terlampir.

Aturan:
- Buat %d soal pilihan ganda dengan tingkat kesulitan %s.
- Untuk setiap soal sertakan "explanation" (pembahasan singkat) yang menjelaskan jawaban yang benar.
- Judul ("title") harus dihasilkan secara otomatis berdasarkan isi materi.
- JANGAN sertakan teks pengantar, penjelasan, atau markdown format seperti %s.
- Strukturnya harus seperti ini:
{
  "title": "Judul yang relevan dengan materi",
  "questions": [
    {
      "question": "Isi pertanyaan...",
      "options": ["A. Opsi 1", "B. Opsi 2", "C. Opsi 3", "D. Opsi 4"],
      "correct_answer": "A",
      "explanation": "Penjelasan singkat mengapa jawaban benar"
    }
  ]
}
`, question_count, difficulty, "```json")

	payload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{"text": promptText},
					{"inline_data": map[string]string{"mime_type": mimeType, "data": b64}},
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

	clean := strings.ReplaceAll(result, "```json", "")
	clean = strings.ReplaceAll(clean, "```", "")
	clean = strings.TrimSpace(clean)

	var aiResp struct {
		Title     string `json:"title"`
		Questions []struct {
			Question      string   `json:"question"`
			Options       []string `json:"options"`
			CorrectAnswer string   `json:"correct_answer"`
			Explanation   string   `json:"explanation"`
		} `json:"questions"`
	}

	if err := json.Unmarshal([]byte(clean), &aiResp); err != nil {
		return map[string]interface{}{
			"raw_response": result,
			"error":        fmt.Sprintf("failed to parse generated JSON: %v", err),
		}, nil
	}

	quiz := models.Quiz{
		Title:      aiResp.Title,
		Difficulty: difficulty,
		Questions:  []models.Question{},
	}

	for _, q := range aiResp.Questions {
		mq := models.Question{
			Question:    q.Question,
			Explanation: q.Explanation,
			Options:     []models.Option{},
		}
		for _, opt := range q.Options {
			label := ""
			content := strings.TrimSpace(opt)
			if len(content) >= 2 {
				first := strings.TrimSpace(content[:1])
				sep := content[1]
				if (sep == '.' || sep == ')') && strings.ToUpper(first) >= "A" && strings.ToUpper(first) <= "Z" {
					label = strings.ToUpper(first)
					content = strings.TrimSpace(content[2:])
				}
			}
			isCorrect := false
			if label != "" && strings.ToUpper(q.CorrectAnswer) == label {
				isCorrect = true
			} else if q.CorrectAnswer != "" && strings.EqualFold(strings.TrimSpace(q.CorrectAnswer), content) {
				isCorrect = true
			}

			mo := models.Option{
				Content:   content,
				IsCorrect: isCorrect,
			}
			mq.Options = append(mq.Options, mo)
		}
		quiz.Questions = append(quiz.Questions, mq)
	}

	return quiz, nil
}
