package service

import (
	"ReilBleem13/pull_requests_service/internal/domain"
	"context"
)

type TeamRepositoryInterface interface {
	Create(ctx context.Context, teamName string, users []domain.User) error
	Get(ctx context.Context, teamName string) ([]domain.User, error)
}
type UserRepositoryInterface interface{}
type PullRequestRepositoryInterface interface{}

type LoggerInterfaces interface {
	Debug(msg string, params ...any)
	Info(msg string, params ...any)
	Warn(msg string, params ...any)
	Error(msg string, params ...any)
}
