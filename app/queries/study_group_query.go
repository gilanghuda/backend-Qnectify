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

func (q *StudyGroupQueries) GetStudyGroupDetail(groupID uuid.UUID) (*models.StudyGroupDetail, error) {
	// fetch group
	group, err := q.GetStudyGroup(groupID)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, nil
	}

	// fetch members
	membersQuery := `
	SELECT u.uid, u.username, u.image_url, sgm.joined_at
	FROM study_group_member sgm
	JOIN users u ON u.uid = sgm.user_id
	WHERE sgm.group_id = $1
	ORDER BY sgm.joined_at ASC
	`
	rows, err := q.DB.Query(membersQuery, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := []models.StudyGroupMember{}
	for rows.Next() {
		var m models.StudyGroupMember
		if err := rows.Scan(&m.UserID, &m.Username, &m.ImageURL, &m.JoinedAt); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// fetch leaderboard: sum of scores per user in this group
	leaderQuery := `
	SELECT u.uid, u.username, u.image_url, COALESCE(SUM(a.score),0) as total_score
	FROM study_group_member sgm
	JOIN users u ON u.uid = sgm.user_id
	LEFT JOIN attempts_quiz a ON a.user_id = sgm.user_id
	WHERE sgm.group_id = $1
	GROUP BY u.uid, u.username, u.image_url
	ORDER BY total_score DESC
	LIMIT 50
	`
	lrows, err := q.DB.Query(leaderQuery, groupID)
	if err != nil {
		return nil, err
	}
	defer lrows.Close()

	leaderboard := []models.StudyGroupMemberScore{}
	for lrows.Next() {
		var e models.StudyGroupMemberScore
		if err := lrows.Scan(&e.UserID, &e.Username, &e.ImageURL, &e.TotalScore); err != nil {
			return nil, err
		}
		leaderboard = append(leaderboard, e)
	}
	if err := lrows.Err(); err != nil {
		return nil, err
	}

	detail := &models.StudyGroupDetail{
		Group:       *group,
		Members:     members,
		Leaderboard: leaderboard,
	}
	return detail, nil
}
