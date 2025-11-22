package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"ReilBleem13/pull_requests_service/internal/domain"
	"ReilBleem13/pull_requests_service/internal/repository"
	"ReilBleem13/pull_requests_service/internal/service"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type mockLogger struct{}

func (m *mockLogger) Debug(string, ...any) {}
func (m *mockLogger) Info(string, ...any)  {}
func (m *mockLogger) Warn(string, ...any)  {}
func (m *mockLogger) Error(string, ...any) {}

func assertAppError(t *testing.T, err error, expectedCode domain.ErrorCode, messageContains ...string) {
	var appErr *domain.AppError
	require.True(t, errors.As(err, &appErr), "error must be *domain.AppError")
	assert.Equal(t, expectedCode, appErr.Code)

	if len(messageContains) > 0 {
		assert.Contains(t, appErr.Message, messageContains[0])
	}
}

func cleanupDatabase(db *sqlx.DB) {
	tables := []string{
		"pull_request_reviewers",
		"pull_requests",
		"team_members",
		"teams",
		"users",
	}

	for _, table := range tables {
		_, err := db.Exec(`TRUNCATE TABLE ` + table + ` RESTART IDENTITY CASCADE`)
		if err != nil {
			panic("failed to truncate table " + table + ": " + err.Error())
		}
	}
}

func setupTestDatabase(t *testing.T) *sqlx.DB {
	t.Helper()

	ctx := context.Background()

	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(15*time.Second),
		),
	)
	require.NoError(t, err)

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sqlx.Open("postgres", connStr)
	require.NoError(t, err)
	require.NoError(t, db.Ping())

	schema := `
		CREATE TABLE users (
		    user_id     TEXT        PRIMARY KEY,
		    username    TEXT        NOT NULL,                  
		    is_active   BOOLEAN     NOT NULL DEFAULT true,
		    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE teams (
		    team_name   TEXT        PRIMARY KEY,
		    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE team_members (
		    team_name   TEXT NOT NULL REFERENCES teams(team_name)   ON DELETE CASCADE,
		    user_id     TEXT NOT NULL REFERENCES users(user_id)     ON DELETE CASCADE,
		    joined_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		    PRIMARY KEY (team_name, user_id)
		);

		CREATE TABLE pull_requests (
		    pull_request_id     TEXT        PRIMARY KEY,
		    pull_request_name   TEXT        NOT NULL,
		    author_id           TEXT        NOT NULL REFERENCES users(user_id),
		    status              TEXT        NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'MERGED')),
		    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		    merged_at           TIMESTAMPTZ NULL
		);

		CREATE TABLE pull_request_reviewers (
		    pull_request_id TEXT NOT NULL REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE,
		    user_id         TEXT NOT NULL REFERENCES users(user_id)         ON DELETE CASCADE,
		    assigned_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		    PRIMARY KEY (pull_request_id, user_id)
		);

		CREATE INDEX idx_team_members_team_name ON team_members(team_name);
		CREATE INDEX idx_team_members_user_id ON team_members(user_id);
		CREATE INDEX idx_pr_status ON pull_requests(status);
		CREATE INDEX idx_pr_author_id ON pull_requests(author_id);
		CREATE INDEX idx_reviewers_pull_request_id ON pull_request_reviewers(pull_request_id);
		CREATE INDEX idx_reviewers_user_id ON pull_request_reviewers(user_id);
	`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	t.Cleanup(func() {
		cleanupDatabase(db)
		_ = db.Close()
		_ = container.Terminate(context.Background())
	})

	return db
}

func TestService_SetIsActive_Integration(t *testing.T) {
	db := setupTestDatabase(t)

	userRepo := repository.NewUserRepository(db)
	teamRepo := repository.NewTeamRepository(db)
	prRepo := repository.NewPullRequestRepository(db)

	svc := service.NewService(userRepo, teamRepo, prRepo, &mockLogger{})
	ctx := context.Background()

	_, err := db.Exec(`
		INSERT INTO teams (team_name) VALUES ('backend');
		INSERT INTO users (user_id, username) VALUES ('user-123', 'alice');
		INSERT INTO team_members (team_name, user_id) VALUES ('backend', 'user-123');
	`)
	require.NoError(t, err)

	t.Run("successfully deactivate user", func(t *testing.T) {
		user, teamName, err := svc.SetIsActive(ctx, "user-123", false)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.False(t, user.IsActive)
		assert.Equal(t, "backend", teamName)

		var isActive bool
		err = db.Get(&isActive, "SELECT is_active FROM users WHERE user_id = $1", "user-123")
		require.NoError(t, err)
		assert.False(t, isActive)
	})

	t.Run("return ErrNotFound for non-existent user", func(t *testing.T) {
		_, _, err := svc.SetIsActive(ctx, "ghost-user-999", true)

		assert.Error(t, err)
		assertAppError(t, err, domain.CodeNotFound)
	})

	t.Run("return ErrInvalidRequest on empty userID", func(t *testing.T) {
		_, _, err := svc.SetIsActive(ctx, "", true)

		assert.Error(t, err)
		assertAppError(t, err, domain.CodeInvalidRequest, "user_id is empty")
	})
}

func TestService_GetReview_Integration(t *testing.T) {
	db := setupTestDatabase(t)

	userRepo := repository.NewUserRepository(db)
	teamRepo := repository.NewTeamRepository(db)
	prRepo := repository.NewPullRequestRepository(db)
	svc := service.NewService(userRepo, teamRepo, prRepo, &mockLogger{})
	ctx := context.Background()

	// Пользователь существует, но у него нет PR на ревью
	_, err := db.Exec(`
		INSERT INTO teams (team_name) VALUES ('dev');
		INSERT INTO users (user_id, username, is_active) VALUES ('lonely', 'Lonely', true);
		INSERT INTO team_members (team_name, user_id) VALUES ('dev', 'lonely');
	`)
	require.NoError(t, err)

	t.Run("return empty slice when no PRs assigned", func(t *testing.T) {
		prs, err := svc.GetReview(ctx, "lonely")
		assert.Error(t, err)
		assert.Empty(t, prs)
		assertAppError(t, err, domain.CodeNotFound)
	})

	t.Run("return ErrInvalidRequest on empty userID", func(t *testing.T) {
		prs, err := svc.GetReview(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, prs)
		assertAppError(t, err, domain.CodeInvalidRequest, "user_id is empty")
	})
}
