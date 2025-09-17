package queries

import (
	"database/sql"
	"errors"
	"time"

	"github.com/gilanghuda/backend-Quizzo/app/models"
	"github.com/gilanghuda/backend-Quizzo/pkg/utils"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type StudyGroupQueries struct {
	DB *sql.DB
}

func (q *StudyGroupQueries) CreateStudyGroup(sg *models.StudyGroup) (*models.StudyGroup, error) {
	const maxAttempt = 3

	for i := 0; i < maxAttempt; i++ {
		inviteCode, err := utils.GenerateInviteCode(8)
		if err != nil {
			return nil, err
		}

		query := `
			INSERT INTO study_group
				(name, description, invite_code, max_member, is_private, created_by)
			VALUES
				($1, $2, $3, $4, $5, $6)
			RETURNING id, name, description, invite_code, member_count, max_member, is_private, created_by, created_at, updated_at
		`

		var created models.StudyGroup
		err = q.DB.QueryRow(
			query,
			sg.Name,
			sg.Description,
			inviteCode,
			sg.MaxMember,
			sg.IsPrivate,
			sg.CreatedBy,
		).Scan(
			&created.ID,
			&created.Name,
			&created.Description,
			&created.InviteCode,
			&created.MemberCount,
			&created.MaxMember,
			&created.IsPrivate,
			&created.CreatedBy,
			&created.CreatedAt,
			&created.UpdatedAt,
		)

		if err == nil {
			return &created, nil
		}
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			continue
		}
		return nil, err
	}

	return nil, errors.New("failed to generate unique invite code after multiple attempts")
}

func (q *StudyGroupQueries) GetStudyGroup(id uuid.UUID) (*models.StudyGroup, error) {
	query := `SELECT id, name, description, invite_code, member_count, max_member, is_private, created_by, created_at, updated_at FROM study_group WHERE id = $1`
	var sg models.StudyGroup
	row := q.DB.QueryRow(query, id)
	if err := row.Scan(&sg.ID, &sg.Name, &sg.Description, &sg.InviteCode, &sg.MemberCount, &sg.MaxMember, &sg.IsPrivate, &sg.CreatedBy, &sg.CreatedAt, &sg.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &sg, nil
}

func (q *StudyGroupQueries) UpdateStudyGroup(sg *models.StudyGroup) error {
	query := `UPDATE study_group SET name=$1, description=$2, max_member=$3, is_private=$4, updated_at=$5 WHERE id=$6`
	_, err := q.DB.Exec(query, sg.Name, sg.Description, sg.MaxMember, sg.IsPrivate, time.Now(), sg.ID)
	return err
}

func (q *StudyGroupQueries) DeleteStudyGroup(id uuid.UUID) error {
	query := `DELETE FROM study_group WHERE id=$1`
	_, err := q.DB.Exec(query, id)
	return err
}

func (q *StudyGroupQueries) JoinStudyGroup(groupID, userID uuid.UUID) error {
	res, err := q.DB.Exec(`INSERT INTO study_group_member (group_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, groupID, userID)
	if err != nil {
		return err
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if ra > 0 {
		_, err = q.DB.Exec(`UPDATE study_group SET member_count = member_count + 1 WHERE id = $1`, groupID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (q *StudyGroupQueries) GetAllStudyGroups(limit, offset int) ([]models.StudyGroup, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	query := `SELECT id, name, description, invite_code, member_count, max_member, is_private, created_by, created_at, updated_at FROM study_group ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := q.DB.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := []models.StudyGroup{}
	for rows.Next() {
		var sg models.StudyGroup
		if err := rows.Scan(&sg.ID, &sg.Name, &sg.Description, &sg.InviteCode, &sg.MemberCount, &sg.MaxMember, &sg.IsPrivate, &sg.CreatedBy, &sg.CreatedAt, &sg.UpdatedAt); err != nil {
			return nil, err
		}
		res = append(res, sg)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (q *StudyGroupQueries) GetStudyGroupsForUser(userID uuid.UUID, limit, offset int) ([]models.StudyGroup, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	query := `SELECT sg.id, sg.name, sg.description, sg.invite_code, sg.member_count, sg.max_member, sg.is_private, sg.created_by, sg.created_at, sg.updated_at
	FROM study_group sg
	JOIN study_group_member sgm ON sgm.group_id = sg.id
	WHERE sgm.user_id = $1
	ORDER BY sg.created_at DESC
	LIMIT $2 OFFSET $3`

	rows, err := q.DB.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := []models.StudyGroup{}
	for rows.Next() {
		var sg models.StudyGroup
		if err := rows.Scan(&sg.ID, &sg.Name, &sg.Description, &sg.InviteCode, &sg.MemberCount, &sg.MaxMember, &sg.IsPrivate, &sg.CreatedBy, &sg.CreatedAt, &sg.UpdatedAt); err != nil {
			return nil, err
		}
		res = append(res, sg)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}
