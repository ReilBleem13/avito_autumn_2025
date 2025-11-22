package service

import (
	"ReilBleem13/pull_requests_service/internal/domain"
	"context"
)

type TeamRepositoryInterface interface {
	Create(ctx context.Context, teamName string, users []domain.User) error
	Get(ctx context.Context, teamName string) ([]domain.User, error)
}
type UserRepositoryInterface interface {
	SetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, string, error)
	GetUser(ctx context.Context, userID string) (*domain.User, error)
	GetTeamName(ctx context.Context, userID string) (string, error)
}

type PullRequestRepositoryInterface interface {
	GetActiveTeamMembers(ctx context.Context, teamName, authorID string) ([]domain.User, error)
	GetPullRequest(ctx context.Context, prID string) (*domain.PullRequest, error)
	Create(ctx context.Context, prID, prName, authorID string, assignedUsers []string) error
	Merge(ctx context.Context, prID string) error
	ReAssign(ctx context.Context, prID, oldReviewerID, newReviewerID string) error
}

type LoggerInterfaces interface {
	Debug(msg string, params ...any)
	Info(msg string, params ...any)
	Warn(msg string, params ...any)
	Error(msg string, params ...any)
}
