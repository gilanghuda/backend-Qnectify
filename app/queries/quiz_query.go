package queries

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/gilanghuda/backend-Quizzo/app/models"
	"github.com/google/uuid"
)

func InsertQuiz(tx *sql.Tx, quiz models.Quiz, createdBy string, description string) (string, error) {
	userID, err := uuid.Parse(createdBy)
	if err != nil {
		return "", err
	}

	var quizID string
	query := `INSERT INTO quizzes (title, description, difficulty_level, time_limit, created_by) VALUES ($1,$2,$3,$4,$5) RETURNING id`
	if err := tx.QueryRow(query, quiz.Title, description, quiz.Difficulty, nil, userID).Scan(&quizID); err != nil {
		return "", err
	}
	return quizID, nil
}

func InsertQuestionsBulk(tx *sql.Tx, quizID string, questions []models.Question) ([]string, error) {
	if len(questions) == 0 {
		return nil, nil
	}

	var args []interface{}
	vals := make([]string, 0, len(questions))
	idx := 1
	for _, q := range questions {
		vals = append(vals, fmt.Sprintf("($%d,$%d)", idx, idx+1))
		args = append(args, quizID, q.Question)
		idx += 2
	}

	query := fmt.Sprintf("INSERT INTO quiz_questions (quiz_id, question_text) VALUES %s RETURNING id", strings.Join(vals, ","))
	rows, err := tx.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ids, nil
}

func InsertOption(tx *sql.Tx, questionID string, opt models.Option) (string, error) {
	var optID string
	optQuery := `INSERT INTO quiz_options (question_id, content, is_correct) VALUES ($1,$2,$3) RETURNING id`
	if err := tx.QueryRow(optQuery, questionID, opt.Content, opt.IsCorrect).Scan(&optID); err != nil {
		return "", err
	}
	return optID, nil
}

func InsertOptionsBulk(tx *sql.Tx, questionIDs []string, questions []models.Question) error {
	var args []interface{}
	vals := make([]string, 0)
	idx := 1
	for qi, q := range questions {
		qID := ""
		if qi < len(questionIDs) {
			qID = questionIDs[qi]
		} else {
			return fmt.Errorf("questionIDs length mismatch")
		}
		for _, opt := range q.Options {
			vals = append(vals, fmt.Sprintf("($%d,$%d,$%d)", idx, idx+1, idx+2))
			args = append(args, qID, opt.Content, opt.IsCorrect)
			idx += 3
		}
	}

	if len(vals) == 0 {
		return nil
	}

	query := fmt.Sprintf("INSERT INTO quiz_options (question_id, content, is_correct) VALUES %s", strings.Join(vals, ","))
	if _, err := tx.Exec(query, args...); err != nil {
		return err
	}
	return nil
}
