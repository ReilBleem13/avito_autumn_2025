package repository

import (
	"ReilBleem13/pull_requests_service/internal/domain"
	"ReilBleem13/pull_requests_service/internal/repository/database"
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type TeamRepository struct {
	db *sqlx.DB
}

func NewTeamRepository(db *database.PostgresDB) *TeamRepository {
	return &TeamRepository{
		db: db.Client(),
	}
}

func (t *TeamRepository) Create(ctx context.Context, teamName string, users []domain.User) error {
	tx, err := t.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	createTeamQuery := `
		INSERT INTO teams (team_name) VALUES ($1)`
	_, err = tx.ExecContext(ctx, createTeamQuery, teamName)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return domain.ErrTeamExists()
		}
		return err
	}

	createUserQuery := `
		INSERT INTO users (user_id, username, is_active)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, username) DO NOTHING
	`

	createTeamMember := `
		INSERT INTO team_members (user_id, team_name)
		VALUES ($1, $2)
		ON CONFLICT (user_id, team_name) DO NOTHING
	`

	for _, user := range users {
		_, err := tx.ExecContext(ctx, createUserQuery, user.UserID, user.Username, user.IsActive)
		if err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, createTeamMember, user.UserID, teamName)
		if err != nil {
			return err
		}
	}
	// добавить проверку если уже существует
	return nil
}

func (t *TeamRepository) Get(ctx context.Context, teamName string) ([]domain.User, error) {
	checkQuery := `SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)`

	var exists bool
	if err := t.db.GetContext(ctx, &exists, checkQuery, teamName); err != nil {
		return nil, err
	}

	if !exists {
		return nil, domain.ErrNotFound()
	}

	getTeamMembersQuery := `
		SELECT u.user_id, u.username, u.is_active
		FROM team_members tm
		JOIN users u ON u.user_id = tm.user_id
		WHERE tm.team_name = $1
		ORDER BY u.created_at
	`

	var teamMembers []domain.User
	if err := t.db.SelectContext(ctx, &teamMembers, getTeamMembersQuery, teamName); err != nil {
		return nil, err
	}
	return teamMembers, nil
}
