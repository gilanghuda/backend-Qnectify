package queries

import (
	"database/sql"
	"errors"

	"github.com/gilanghuda/backend-Quizzo/app/models"
	"github.com/gilanghuda/backend-Quizzo/pkg/utils"
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
