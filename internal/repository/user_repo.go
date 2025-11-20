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

func (u *UserRepository) SetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error) {
	query := `
		UPDATE users
		SET is_active = $2, updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $1
		RETURNING user_id, username, is_active
	`

	var user domain.User
	if err := u.db.QueryRowContext(ctx, query, userID, isActive).
		Scan(&user.UserID, &user.Username, &user.IsActive); err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	return nil, nil
}
