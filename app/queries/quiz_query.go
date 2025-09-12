package queries

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gilanghuda/backend-Quizzo/app/models"
	"github.com/google/uuid"
)

type QuizQueries struct {
	DB *sql.DB
}

func (q *QuizQueries) InsertQuiz(quiz models.Quiz, createdBy string, description string) (string, error) {
	userID, err := uuid.Parse(createdBy)
	if err != nil {
		return "", err
	}

	var quizID string
	query := `INSERT INTO quizzes (title, description, difficulty_level, time_limit, created_by) VALUES ($1,$2,$3,$4,$5) RETURNING id`
	if err := q.DB.QueryRow(query, quiz.Title, description, quiz.Difficulty, nil, userID).Scan(&quizID); err != nil {
		return "", err
	}
	return quizID, nil
}

func (q *QuizQueries) InsertQuestionsBulk(quizID string, questions []models.Question) ([]string, error) {
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
	rows, err := q.DB.Query(query, args...)
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

func (q *QuizQueries) InsertOption(questionID string, opt models.Option) (string, error) {
	var optID string
	optQuery := `INSERT INTO quiz_options (question_id, content, is_correct) VALUES ($1,$2,$3) RETURNING id`
	if err := q.DB.QueryRow(optQuery, questionID, opt.Content, opt.IsCorrect).Scan(&optID); err != nil {
		return "", err
	}
	return optID, nil
}

func (q *QuizQueries) InsertOptionsBulk(questionIDs []string, questions []models.Question) error {
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
	if _, err := q.DB.Exec(query, args...); err != nil {
		return err
	}
	return nil
}

func (q *QuizQueries) GetQuizByUserId(userID string) (*models.Quiz, error) {
	var quiz models.Quiz
	var questionsJSON []byte

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	query := `
        SELECT 
            q.id,
            q.title,
            q.description,
            q.difficulty_level,
            q.time_limit,
            q.created_by,
            COALESCE(json_agg(
                json_build_object(
                    'id', qq.id,
                    'quiz_id', qq.quiz_id,
                    'question_text', qq.question_text,
                    'options', (
                        SELECT json_agg(
                            json_build_object(
                                'id', qo.id,
                                'question_id', qo.question_id,
                                'content', qo.content,
                                'is_correct', qo.is_correct
                            )
                        )
                        FROM quiz_options qo
                        WHERE qo.question_id = qq.id
                    )
                )
            ), '[]') AS questions
        FROM quizzes q
        JOIN quiz_questions qq ON qq.quiz_id = q.id
        WHERE q.created_by = $1
        GROUP BY q.id, q.title, q.description, q.difficulty_level, q.time_limit, q.created_by;
    `

	err = q.DB.QueryRow(query, userUUID).Scan(
		&quiz.ID,
		&quiz.Title,
		&quiz.Description,
		&quiz.Difficulty,
		&quiz.TimeLimit,
		&quiz.CreatedBy,
		&questionsJSON,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(questionsJSON, &quiz.Questions); err != nil {
		return nil, err
	}

	return &quiz, nil
}
