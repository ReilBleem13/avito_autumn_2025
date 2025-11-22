// internal/service/service_pr_integration_test.go

package service_test

import (
	"context"
	"math/rand"
	"testing"

	"ReilBleem13/pull_requests_service/internal/domain"
	"ReilBleem13/pull_requests_service/internal/repository"
	"ReilBleem13/pull_requests_service/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_CreatePullRequest_Integration(t *testing.T) {
	db := setupTestDatabase(t)

	userRepo := repository.NewUserRepository(db)
	teamRepo := repository.NewTeamRepository(db)
	prRepo := repository.NewPullRequestRepository(db)

	svc := service.NewService(userRepo, teamRepo, prRepo, &mockLogger{})
	ctx := context.Background()

	_, err := db.Exec(`
		INSERT INTO teams (team_name) VALUES ('backend');
		INSERT INTO users (user_id, username, is_active) VALUES
		('author-1', 'alice', true),
		('rev-1', 'bob', true),
		('rev-2', 'charlie', true),
		('rev-3', 'dave', true);

		INSERT INTO team_members (team_name, user_id) VALUES
		('backend', 'author-1'),
		('backend', 'rev-1'),
		('backend', 'rev-2'),
		('backend', 'rev-3');
	`)
	require.NoError(t, err)

	t.Run("successfully create PR and assign up to 2 reviewers", func(t *testing.T) {
		pr, err := svc.CreatePullRequest(ctx, "pr-001", "Fix login", "author-1")

		assert.NoError(t, err)
		assert.Equal(t, "pr-001", pr.PullRequestID)
		assert.Equal(t, domain.PRStatusOpen, pr.Status)
		assert.Len(t, pr.AssignedReviewers, 2)
		assert.NotContains(t, pr.AssignedReviewers, "author-1")
	})

	t.Run("fail on empty prID", func(t *testing.T) {
		pr, err := svc.CreatePullRequest(ctx, "", "Title", "author-1")
		assert.Error(t, err)
		assert.Nil(t, pr)
		assertAppError(t, err, domain.CodeInvalidRequest, "pr_id is empty")
	})

	t.Run("fail on empty prName", func(t *testing.T) {
		pr, err := svc.CreatePullRequest(ctx, "pr-002", "", "author-1")
		assert.Error(t, err)
		assert.Nil(t, pr)
		assertAppError(t, err, domain.CodeInvalidRequest, "pr_name is empty")
	})

	t.Run("fail on empty authorID", func(t *testing.T) {
		pr, err := svc.CreatePullRequest(ctx, "pr-003", "Title", "")
		assert.Error(t, err)
		assert.Nil(t, pr)
		assertAppError(t, err, domain.CodeInvalidRequest, "author_id is empty")
	})

	t.Run("fail when author not found", func(t *testing.T) {
		pr, err := svc.CreatePullRequest(ctx, "pr-004", "Title", "ghost")
		assert.Error(t, err)
		assert.Nil(t, pr)
		assertAppError(t, err, domain.CodeNotFound)
	})

	t.Run("assign fewer than 2 if not enough active members", func(t *testing.T) {
		_, err := db.Exec(`UPDATE users SET is_active = false WHERE user_id IN ('rev-2', 'rev-3')`)
		require.NoError(t, err)

		pr, err := svc.CreatePullRequest(ctx, "pr-005", "Only one reviewer", "author-1")
		assert.NoError(t, err)
		assert.Len(t, pr.AssignedReviewers, 1)
		assert.Equal(t, "rev-1", pr.AssignedReviewers[0])
	})
}

func TestService_MergePullRequest_Integration(t *testing.T) {
	db := setupTestDatabase(t)

	userRepo := repository.NewUserRepository(db)
	teamRepo := repository.NewTeamRepository(db)
	prRepo := repository.NewPullRequestRepository(db)

	svc := service.NewService(userRepo, teamRepo, prRepo, &mockLogger{})
	ctx := context.Background()

	_, err := db.Exec(`
		INSERT INTO teams (team_name) VALUES ('backend');
		INSERT INTO users (user_id, username) VALUES ('author-1', 'alice');
		INSERT INTO team_members (team_name, user_id) VALUES ('backend', 'author-1');
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status) 
		VALUES ('pr-open', 'Open PR', 'author-1', 'OPEN');
	`)
	require.NoError(t, err)

	t.Run("successfully merge open PR", func(t *testing.T) {
		pr, err := svc.MergePullRequest(ctx, "pr-open")

		assert.NoError(t, err)
		assert.Equal(t, domain.PRStatusMerged, pr.Status)
		assert.NotEmpty(t, pr.MergedAt)
	})

	t.Run("fail on empty prID", func(t *testing.T) {
		pr, err := svc.MergePullRequest(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, pr)
		assertAppError(t, err, domain.CodeInvalidRequest, "pr_id is empty")
	})

	t.Run("fail when PR not found", func(t *testing.T) {
		pr, err := svc.MergePullRequest(ctx, "ghost-pr")
		assert.Error(t, err)
		assert.Nil(t, pr)
		assertAppError(t, err, domain.CodeNotFound)
	})
}

func TestService_ReAssign_Integration(t *testing.T) {
	rand.Seed(42)

	db := setupTestDatabase(t)

	userRepo := repository.NewUserRepository(db)
	teamRepo := repository.NewTeamRepository(db)
	prRepo := repository.NewPullRequestRepository(db)
	svc := service.NewService(userRepo, teamRepo, prRepo, &mockLogger{})
	ctx := context.Background()

	setupReAssignTest := func(t *testing.T, teamName string) {
		tx, err := db.BeginTx(ctx, nil)
		require.NoError(t, err)
		defer tx.Rollback()

		exec := func(query string, args ...any) {
			_, err := tx.Exec(query, args...)
			require.NoError(t, err)
		}

		exec(`INSERT INTO teams (team_name) VALUES ($1) ON CONFLICT DO NOTHING`, teamName)

		exec(`
		INSERT INTO users (user_id, username, is_active) VALUES
			('author', 'Author', true),
			('old', 'OldReviewer', true),
			('c1', 'C1', true),
			('c2', 'C2', true),
			('c3', 'C3', true)
		ON CONFLICT (user_id) DO NOTHING`)

		exec(`
		INSERT INTO team_members (team_name, user_id) VALUES
			($1,'author'), ($1,'old'), ($1,'c1'), ($1,'c2'), ($1,'c3')
		ON CONFLICT DO NOTHING`, teamName)

		exec(`DELETE FROM pull_request_reviewers WHERE pull_request_id = 'pr-reassign'`)
		exec(`DELETE FROM pull_requests WHERE pull_request_id = 'pr-reassign'`)

		exec(`
		INSERT INTO pull_requests 
		(pull_request_id, pull_request_name, author_id, status)
		VALUES ('pr-reassign', 'Test PR', 'author', 'OPEN')`)

		exec(`
		INSERT INTO pull_request_reviewers (pull_request_id, user_id)
		VALUES ('pr-reassign', 'old')`)

		require.NoError(t, tx.Commit())
	}

	t.Run("successfully reassign reviewer", func(t *testing.T) {
		setupReAssignTest(t, "team-success")
		pr, newID, err := svc.ReAssign(ctx, "pr-reassign", "old")
		assert.NoError(t, err)
		assert.Contains(t, []string{"c1", "c2", "c3"}, newID)
		assert.Contains(t, pr.AssignedReviewers, newID)
		assert.NotContains(t, pr.AssignedReviewers, "old")
	})

	t.Run("fail on merged PR", func(t *testing.T) {
		setupReAssignTest(t, "team-merged")
		_, err := svc.MergePullRequest(ctx, "pr-reassign")
		require.NoError(t, err)

		pr, _, err := svc.ReAssign(ctx, "pr-reassign", "old")
		assert.Error(t, err)
		assert.Nil(t, pr)
		assertAppError(t, err, domain.CodePRMerged)
	})

	t.Run("fail when old reviewer not assigned", func(t *testing.T) {
		setupReAssignTest(t, "team-not-assigned")
		pr, _, err := svc.ReAssign(ctx, "pr-reassign", "ghost")
		assert.Error(t, err)
		assert.Nil(t, pr)
		assertAppError(t, err, domain.CodeNotAssigned)
	})

	t.Run("fail when no replacement candidate", func(t *testing.T) {
		tx, err := db.BeginTx(ctx, nil)
		require.NoError(t, err)
		defer tx.Rollback()

		_, err = tx.Exec(`INSERT INTO teams (team_name) VALUES ('team-no-candidate') ON CONFLICT DO NOTHING`)
		require.NoError(t, err)

		_, err = tx.Exec(`
			INSERT INTO users (user_id, username, is_active) VALUES
				('author', 'Author', true),
				('old', 'Old', true),
				('c1', 'C1', true),
				('c2', 'C2', true),
				('c3', 'C3', true)
			ON CONFLICT (user_id) DO NOTHING`)
		require.NoError(t, err)

		_, err = tx.Exec(`
			INSERT INTO team_members (team_name, user_id) VALUES
				('team-no-candidate', 'author'),
				('team-no-candidate', 'old'),
				('team-no-candidate', 'c1'),
				('team-no-candidate', 'c2'),
				('team-no-candidate', 'c3')
			ON CONFLICT DO NOTHING`)
		require.NoError(t, err)

		_, err = tx.Exec(`
			DELETE FROM pull_request_reviewers WHERE pull_request_id = 'pr-reassign';
			DELETE FROM pull_requests WHERE pull_request_id = 'pr-reassign';
		`)
		require.NoError(t, err)

		_, err = tx.Exec(`
			INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status)
			VALUES ('pr-reassign', 'Test PR', 'author', 'OPEN')
		`)
		require.NoError(t, err)

		_, err = tx.Exec(`
			INSERT INTO pull_request_reviewers (pull_request_id, user_id)
			VALUES ('pr-reassign', 'old')
		`)
		require.NoError(t, err)

		_, err = tx.Exec(`UPDATE users SET is_active = false WHERE user_id != 'author'`)
		require.NoError(t, err)

		require.NoError(t, tx.Commit())

		pr, _, err := svc.ReAssign(ctx, "pr-reassign", "old")
		assert.Error(t, err)
		assert.Nil(t, pr)
		assertAppError(t, err, domain.CodeNoCandidate)
	})
}
