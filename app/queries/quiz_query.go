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

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	query := `
   SELECT q.id, q.title, q.description, q.difficulty_level, q.created_by, q.time_limit
	FROM quizzes q
	WHERE q.created_by = $1
	ORDER BY q.created_at DESC;
    `

	err = q.DB.QueryRow(query, userUUID).Scan(
		&quiz.ID,
		&quiz.Title,
		&quiz.Description,
		&quiz.Difficulty,
		&quiz.CreatedBy,
		&quiz.TimeLimit,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &quiz, nil
}

func (q *QuizQueries) GetQuizzesFromFollowing(userID string) ([]models.Quiz, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	query := `
	SELECT q.id, q.title, q.description, q.difficulty_level, q.created_by, 
		COALESCE((SELECT COUNT(*) FROM attempts_quiz a WHERE a.quiz_id = q.id), 0) AS attempts_count,
		q.created_at
	FROM quizzes q
	JOIN socials s ON s.following = q.created_by
	WHERE s.follower_id = $1
	ORDER BY q.created_at DESC
	LIMIT 10
	`

	rows, err := q.DB.Query(query, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []models.Quiz
	for rows.Next() {
		var qs models.Quiz
		if err := rows.Scan(&qs.ID, &qs.Title, &qs.Description, &qs.Difficulty, &qs.CreatedBy, &qs.Attempts, &qs.CreatedAt); err != nil {
			return nil, err
		}
		res = append(res, qs)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (q *QuizQueries) InsertQuizAttempt(quizID, userID string, score, totalQuestions int, isCompleted bool) (string, error) {
	var attemptID string
	query := `INSERT INTO attempts_quiz (quiz_id, user_id, score, total_questions, is_completed) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	if err := q.DB.QueryRow(query, quizID, userID, score, totalQuestions, isCompleted).Scan(&attemptID); err != nil {
		return "", err
	}
	return attemptID, nil
}

func (q *QuizQueries) EvaluateQuizAttempt(quizID string, answers map[string]string) (int, int, error) {
	if len(answers) == 0 {
		var cnt int
		err := q.DB.QueryRow(`SELECT COUNT(*) FROM quiz_questions WHERE quiz_id = $1`, quizID).Scan(&cnt)
		if err != nil {
			return 0, 0, err
		}
		return 0, cnt, nil
	}
	correct := 0
	for qid, oid := range answers {
		var isCorrect bool
		err := q.DB.QueryRow(`SELECT is_correct FROM quiz_options WHERE id = $1 AND question_id = $2`, oid, qid).Scan(&isCorrect)
		if err != nil {
			if err == sql.ErrNoRows {
				continue
			}
			return 0, 0, err
		}
		if isCorrect {
			correct++
		}
	}
	return correct, len(answers), nil
}

func (q *QuizQueries) GetAttemptsForUser(userID string, quizID *string, limit int) ([]models.Attempt, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	base := `SELECT id, quiz_id, user_id, score, total_questions, submitted_at, is_completed FROM attempts_quiz WHERE user_id = $1`
	args := []interface{}{uid}
	if quizID != nil {
		base += ` AND quiz_id = $2`
		args = append(args, *quizID)
	}
	base += ` ORDER BY submitted_at DESC`
	if limit > 0 {
		base += ` LIMIT ` + fmt.Sprintf("%d", limit)
	}

	rows, err := q.DB.Query(base, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := []models.Attempt{}
	for rows.Next() {
		var a models.Attempt
		if err := rows.Scan(&a.ID, &a.QuizID, &a.UserID, &a.Score, &a.TotalQuestions, &a.SubmittedAt, &a.IsCompleted); err != nil {
			return nil, err
		}
		res = append(res, a)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (q *QuizQueries) GetQuizByID(quizID string) (*models.Quiz, error) {
	var quiz models.Quiz
	var questionsJSON []byte

	id, err := uuid.Parse(quizID)
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
            ) FILTER (WHERE qq.id IS NOT NULL), '[]') AS questions
        FROM quizzes q
        LEFT JOIN quiz_questions qq ON qq.quiz_id = q.id
        WHERE q.id = $1
        GROUP BY q.id, q.title, q.description, q.difficulty_level, q.time_limit, q.created_by;
    `

	err = q.DB.QueryRow(query, id).Scan(
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

func (q *QuizQueries) GetUserLeaderboard(limit int) ([]models.UserLeaderboardEntry, error) {
	base := `SELECT u.uid, u.username, u.image_url, COALESCE(SUM(a.score),0) as total_score FROM users u LEFT JOIN attempts_quiz a ON a.user_id = u.uid GROUP BY u.uid, u.username, u.image_url ORDER BY total_score DESC`
	if limit > 0 {
		base += ` LIMIT ` + fmt.Sprintf("%d", limit)
	}

	rows, err := q.DB.Query(base)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := []models.UserLeaderboardEntry{}
	for rows.Next() {
		var e models.UserLeaderboardEntry
		if err := rows.Scan(&e.UserID, &e.Username, &e.ImageURL, &e.TotalScore); err != nil {
			return nil, err
		}
		res = append(res, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (q *QuizQueries) GetStudyGroupLeaderboard(limit int) ([]models.StudyGroupLeaderboardEntry, error) {
	query := `
SELECT sg.id, sg.name, sg.member_count, COALESCE(SUM(a.score),0) AS total_score
FROM study_group sg
LEFT JOIN study_group_member sgm ON sgm.group_id = sg.id
LEFT JOIN attempts_quiz a ON a.user_id = sgm.user_id
GROUP BY sg.id, sg.name, sg.member_count
ORDER BY total_score DESC
`
	if limit > 0 {
		query += ` LIMIT ` + fmt.Sprintf("%d", limit)
	}

	rows, err := q.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := []models.StudyGroupLeaderboardEntry{}
	for rows.Next() {
		var e models.StudyGroupLeaderboardEntry
		if err := rows.Scan(&e.GroupID, &e.Name, &e.MemberCount, &e.TotalScore); err != nil {
			return nil, err
		}
		res = append(res, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}
