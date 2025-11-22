package repository

import (
	"ReilBleem13/pull_requests_service/internal/domain"
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{
		db: db,
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
			return nil, "", domain.ErrNotFound()
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
			return nil, "", domain.ErrNotFound()
		}
		return nil, "", err
	}

	return &user, teamName, nil
}

func (u *UserRepository) GetUser(ctx context.Context, userID string) (*domain.User, error) {
	getUserQuery := `
		SELECT user_id, username, is_active 
		FROM users
		WHERE user_id = $1
	`

	var user domain.User
	if err := u.db.GetContext(ctx, &user, getUserQuery, userID); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user is not exist %w", domain.ErrNotFound())
		}
	}
	return &user, nil
}

func (u *UserRepository) GetTeamName(ctx context.Context, userID string) (string, error) {
	getTeamNameQuery := `
		SELECT team_name 
		FROM team_members
		WHERE user_id = $1
		ORDER BY team_name
		LIMIT 1
	`

	var teamName string
	if err := u.db.GetContext(ctx, &teamName, getTeamNameQuery, userID); err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("user is not exist %w", domain.ErrNotFound())
		}
	}
	return teamName, nil
}
