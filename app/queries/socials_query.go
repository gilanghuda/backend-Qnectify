package queries

import (
	"database/sql"
	"fmt"

	"github.com/gilanghuda/backend-Quizzo/app/models"
	"github.com/google/uuid"
)

type SocialsQueries struct {
	DB *sql.DB
}

func (q *SocialsQueries) HasLiked(quizID, userID string) (bool, error) {
	var cnt int
	idQ, err := uuid.Parse(quizID)
	if err != nil {
		return false, err
	}
	uID, err := uuid.Parse(userID)
	if err != nil {
		return false, err
	}
	if err := q.DB.QueryRow(`SELECT COUNT(*) FROM likes WHERE quiz_id = $1 AND liked_by = $2`, idQ, uID).Scan(&cnt); err != nil {
		return false, err
	}
	return cnt > 0, nil
}

func (q *SocialsQueries) ToggleLike(quizID, userID string) (bool, error) {
	idQ, err := uuid.Parse(quizID)
	if err != nil {
		return false, err
	}
	uID, err := uuid.Parse(userID)
	if err != nil {
		return false, err
	}

	liked, err := q.HasLiked(quizID, userID)
	if err != nil {
		return false, err
	}
	if liked {
		if _, err := q.DB.Exec(`DELETE FROM likes WHERE quiz_id = $1 AND liked_by = $2`, idQ, uID); err != nil {
			return false, err
		}
		return false, nil
	}
	var id string
	if err := q.DB.QueryRow(`INSERT INTO likes (quiz_id, liked_by) VALUES ($1,$2) RETURNING id`, idQ, uID).Scan(&id); err != nil {
		return false, err
	}
	return true, nil
}

func (q *SocialsQueries) CountLikes(quizID string) (int, error) {
	idQ, err := uuid.Parse(quizID)
	if err != nil {
		return 0, err
	}
	var cnt int
	if err := q.DB.QueryRow(`SELECT COUNT(*) FROM likes WHERE quiz_id = $1`, idQ).Scan(&cnt); err != nil {
		return 0, err
	}
	return cnt, nil
}

func (q *SocialsQueries) AddComment(quizID, userID, content string) (string, error) {
	idQ, err := uuid.Parse(quizID)
	if err != nil {
		return "", err
	}
	uID, err := uuid.Parse(userID)
	if err != nil {
		return "", err
	}
	var commentID string
	query := `INSERT INTO comments (quiz_id, content, commenter_by) VALUES ($1,$2,$3) RETURNING id`
	if err := q.DB.QueryRow(query, idQ, content, uID).Scan(&commentID); err != nil {
		return "", err
	}
	return commentID, nil
}

func (q *SocialsQueries) DeleteComment(commentID, userID string) error {
	cid, err := uuid.Parse(commentID)
	if err != nil {
		return err
	}
	uID, err := uuid.Parse(userID)
	if err != nil {
		return err
	}
	res, err := q.DB.Exec(`DELETE FROM comments WHERE id = $1 AND commenter_by = $2`, cid, uID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("not found or not owner")
	}
	return nil
}

func (q *SocialsQueries) GetComments(quizID string, limit int) ([]models.Comment, error) {
	idQ, err := uuid.Parse(quizID)
	if err != nil {
		return nil, err
	}
	base := `SELECT id, quiz_id, content, commenter_by, created_at FROM comments WHERE quiz_id = $1 ORDER BY created_at DESC`
	if limit > 0 {
		base += fmt.Sprintf(" LIMIT %d", limit)
	}
	rows, err := q.DB.Query(base, idQ)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := []models.Comment{}
	for rows.Next() {
		var c models.Comment
		if err := rows.Scan(&c.ID, &c.QuizID, &c.Content, &c.CommenterBy, &c.CreatedAt); err != nil {
			return nil, err
		}
		res = append(res, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}
