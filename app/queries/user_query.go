package queries

import (
	"database/sql"
	"errors"
	"log"

	"github.com/gilanghuda/backend-Quizzo/app/models"
	"github.com/google/uuid"
)

type UserQueries struct {
	DB *sql.DB
}

func (q *UserQueries) GetUserByID(id uuid.UUID) (models.User, error) {
	user := models.User{}

	query := `SELECT uid, username, user_role, email, password, exp_point, image_url, created_at, updated_at
			  FROM users WHERE uid = $1`

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
		log.Println("Error querying user by email:", err)
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
