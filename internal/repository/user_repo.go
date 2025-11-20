package repository

import (
	"ReilBleem13/pull_requests_service/internal/domain"
	"ReilBleem13/pull_requests_service/internal/repository/database"
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *database.PostgresDB) *UserRepository {
	return &UserRepository{
		db: db.Client(),
	}
}

func (u *UserRepository) SetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, string, error) {
	updateQuery := `
		UPDATE users
		SET is_active = $2, updated_at = NOW()
		WHERE user_id = $1
		RETURNING user_id, username, is_active
	`

	var user domain.User
	if err := u.db.QueryRowContext(ctx, updateQuery, userID, isActive).
		Scan(&user.UserID, &user.Username, &user.IsActive); err != nil {
		if err == sql.ErrNoRows {
			return nil, "", sql.ErrNoRows
		}
		return nil, "", err
	}

	getTeamQuery := `
		SELECT team_name
		FROM team_members
		WHERE user_id = $1
		ORDER BY team_name
		LIMIT 1
	`

	var teamName string
	if err := u.db.GetContext(ctx, &teamName, getTeamQuery, userID); err != nil {
		if err == sql.ErrNoRows {
			return nil, "", sql.ErrNoRows
		}
		return nil, "", err
	}

	return &user, teamName, nil
}
