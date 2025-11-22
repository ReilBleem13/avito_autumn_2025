// internal/service/service_team_integration_test.go

package service_test

import (
	"context"
	"testing"

	"ReilBleem13/pull_requests_service/internal/domain"
	"ReilBleem13/pull_requests_service/internal/repository"
	"ReilBleem13/pull_requests_service/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_CreateTeam_Integration(t *testing.T) {
	db := setupTestDatabase(t)

	userRepo := repository.NewUserRepository(db)
	teamRepo := repository.NewTeamRepository(db)
	prRepo := repository.NewPullRequestRepository(db)

	svc := service.NewService(userRepo, teamRepo, prRepo, &mockLogger{})
	ctx := context.Background()

	_, err := db.Exec(`
		INSERT INTO users (user_id, username, is_active) VALUES
		('alice', 'Alice', true),
		('bob', 'Bob', true),
		('charlie', 'Charlie', true)
		ON CONFLICT DO NOTHING;
	`)
	require.NoError(t, err)

	t.Run("successfully create team with members", func(t *testing.T) {
		users := []domain.User{
			{UserID: "alice", Username: "Alice", IsActive: true},
			{UserID: "bob", Username: "Bob", IsActive: true},
		}

		err := svc.CreateTeam(ctx, "golang-squad", users)
		assert.NoError(t, err)

		var count int
		err = db.Get(&count, `SELECT COUNT(*) FROM teams WHERE team_name = 'golang-squad'`)
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		err = db.Get(&count, `SELECT COUNT(*) FROM team_members WHERE team_name = 'golang-squad'`)
		require.NoError(t, err)
		assert.Equal(t, 2, count)
	})

	t.Run("fail when team_name is empty", func(t *testing.T) {
		err := svc.CreateTeam(ctx, "", []domain.User{{UserID: "alice"}})
		assert.Error(t, err)
		assertAppError(t, err, domain.CodeInvalidRequest, "team_name is empty")
	})

	t.Run("fail when users slice is empty", func(t *testing.T) {
		err := svc.CreateTeam(ctx, "empty-team", []domain.User{})
		assert.Error(t, err)
		assertAppError(t, err, domain.CodeInvalidRequest, "team_users is empty")
	})

	t.Run("fail when team already exists", func(t *testing.T) {
		users := []domain.User{{UserID: "charlie", Username: "Charlie"}}

		err := svc.CreateTeam(ctx, "duplicate-team", users)
		require.NoError(t, err)

		err = svc.CreateTeam(ctx, "duplicate-team", users)

		assert.Error(t, err)
		assertAppError(t, err, domain.CodeTeamExists)
	})
}

func TestService_GetTeam_Integration(t *testing.T) {
	db := setupTestDatabase(t)

	userRepo := repository.NewUserRepository(db)
	teamRepo := repository.NewTeamRepository(db)
	prRepo := repository.NewPullRequestRepository(db)

	svc := service.NewService(userRepo, teamRepo, prRepo, &mockLogger{})
	ctx := context.Background()

	_, err := db.Exec(`
		INSERT INTO users (user_id, username, is_active) VALUES
		('alice', 'Alice', true),
		('bob', 'Bob', true),
		('charlie', 'Charlie', false);

		INSERT INTO teams (team_name) VALUES 
		('backend'), ('frontend'), ('empty-team');

		INSERT INTO team_members (team_name, user_id) VALUES
		('backend', 'alice'),
		('backend', 'bob'),
		('frontend', 'charlie');
	`)
	require.NoError(t, err)

	t.Run("successfully get team members", func(t *testing.T) {
		members, err := svc.GetTeam(ctx, "backend")

		assert.NoError(t, err)
		assert.Len(t, members, 2)

		ids := map[string]bool{}
		for _, m := range members {
			ids[m.UserID] = true
		}
		assert.Contains(t, ids, "alice")
		assert.Contains(t, ids, "bob")
	})

	t.Run("return ErrNotFound when team does not exist", func(t *testing.T) {
		members, err := svc.GetTeam(ctx, "nonexistent-team")

		assert.Error(t, err)
		assert.Nil(t, members)
		assertAppError(t, err, domain.CodeNotFound)
	})

	t.Run("return ErrNotFound when team exists but has no members", func(t *testing.T) {
		_, err := db.Exec(`INSERT INTO teams (team_name) VALUES ('empty-team') ON CONFLICT DO NOTHING`)
		require.NoError(t, err)

		members, err := svc.GetTeam(ctx, "empty-team")

		assert.Error(t, err)
		assert.Nil(t, members)
		assertAppError(t, err, domain.CodeNotFound)
	})
}
