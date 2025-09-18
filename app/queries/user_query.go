package queries

import (
	"database/sql"
	"errors"
	"log"

	"github.com/gilanghuda/backend-Quizzo/app/models"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type UserQueries struct {
	DB *sql.DB
}

func (q *UserQueries) GetUserByID(id uuid.UUID) (models.User, error) {
	user := models.User{}

	query := `SELECT u.uid, u.username, u.user_role, u.email, u.password, u.exp_point, u.image_url, u.created_at, u.updated_at,
		COALESCE(followers.count, 0) as follower_count,
		COALESCE(following.count, 0) as following_count
		FROM users u
		LEFT JOIN (
			SELECT following as user_id, COUNT(*) as count FROM socials GROUP BY following
		) followers ON followers.user_id = u.uid
		LEFT JOIN (
			SELECT follower_id as user_id, COUNT(*) as count FROM socials GROUP BY follower_id
		) following ON following.user_id = u.uid
		WHERE u.uid = $1`

	err := q.DB.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.UserRole,
		&user.Email,
		&user.PasswordHash,
		&user.ExpPoints,
		&user.ImageURL,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.FollowerCount,
		&user.FollowingCount,
	)

	if err != nil {
		log.Println("Error querying user by ID:", err)
		if err == sql.ErrNoRows {
			return user, errors.New("user not found")
		}
		return user, errors.New("unable to get user, DB error")
	}

	return user, nil
}

func (q *UserQueries) GetUserByEmail(email string) (models.User, error) {
	user := models.User{}

	query := `SELECT uid, username, email, password, created_at, updated_at
			  FROM users WHERE email = $1`

	err := q.DB.QueryRow(query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return user, errors.New("user not found")
		}
		return user, errors.New("unable to get user, DB error")
	}

	return user, nil
}

func (q *UserQueries) CreateUser(u *models.User) error {
	query := `INSERT INTO users (uid, username, user_role, email, password, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := q.DB.Exec(query,
		u.ID,
		u.Username,
		u.UserRole,
		u.Email,
		u.PasswordHash,
		u.CreatedAt,
		u.UpdatedAt,
	)

	if err != nil {
		return errors.New("unable to create user, DB error")
	}

	return nil
}

func (q *UserQueries) DeleteUser(id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`

	res, err := q.DB.Exec(query, id)
	if err != nil {
		return errors.New("unable to delete user, DB error")
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("no user deleted")
	}

	return nil
}

func (q *UserQueries) FollowUser(follower uuid.UUID, following uuid.UUID) error {
	if follower == following {
		return errors.New("cannot follow yourself")
	}

	query := `INSERT INTO socials (follower_id, following) VALUES ($1, $2)`
	if _, err := q.DB.Exec(query, follower, following); err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return errors.New("already following")
		}
		return err
	}
	return nil
}

func (q *UserQueries) UnfollowUser(follower uuid.UUID, following uuid.UUID) error {
	query := `DELETE FROM socials WHERE follower_id = $1 AND following = $2`
	res, err := q.DB.Exec(query, follower, following)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("not following")
	}
	return nil
}

func (q *UserQueries) GetRecommendedUsers(userID uuid.UUID, limit int) ([]models.RecommendedUser, error) {
	query := `
	SELECT u.uid, u.username, u.email, COALESCE(f.follower_count,0) as follower_count
	FROM users u
	LEFT JOIN (
		SELECT following as user_id, COUNT(*) as follower_count
		FROM socials
		GROUP BY following
	) f ON f.user_id = u.uid
	WHERE u.uid != $1
	  AND u.uid NOT IN (
		SELECT following FROM socials WHERE follower_id = $1
	  )
	ORDER BY follower_count DESC
	LIMIT $2
	`

	rows, err := q.DB.Query(query, userID, limit)
	if err != nil {
		println("Error executing query:", err.Error())
		return nil, err
	}
	defer rows.Close()

	var res []models.RecommendedUser
	for rows.Next() {
		var r models.RecommendedUser
		if err := rows.Scan(&r.ID, &r.Username, &r.Email, &r.FollowerCount); err != nil {
			return nil, err
		}
		res = append(res, r)
	}
	if err := rows.Err(); err != nil {
		println("Error iterating over rows:", err.Error())
		return nil, err
	}

	return res, nil
}
