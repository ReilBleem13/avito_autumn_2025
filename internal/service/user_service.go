package service

import (
	"ReilBleem13/pull_requests_service/internal/domain"
	"context"
	"database/sql"
	"errors"

	"github.com/theartofdevel/logging"
)

func (s *Service) SetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, string, error) {
	s.logger.Info("attempt to set user status",
		logging.StringAttr("userID", userID),
		logging.BoolAttr("status", isActive),
	)

	if userID == "" {
		s.logger.Error("failed to set user status")
		return nil, "", domain.ErrInvalidRequest("user_id is empty")
	}

	user, teamName, err := s.users.SetIsActive(ctx, userID, isActive)
	if err != nil {
		s.logger.Error("failed to set user status",
			logging.StringAttr("userID", userID),
			logging.ErrAttr(err),
		)

		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", domain.ErrNotFound()
		}

		return nil, "", err
	}

	s.logger.Info("team was succeccfully set",
		logging.StringAttr("userID", userID),
		logging.BoolAttr("status", isActive),
	)
	return user, teamName, nil
}

func (s *Service) GetReview(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
	s.logger.Info("attempt to get review",
		logging.StringAttr("userID", userID),
	)

	if userID == "" {
		s.logger.Error("failed to set user status")
		return nil, domain.ErrInvalidRequest("user_id is empty")
	}

	pullRequests, err := s.prs.GetPullRequestByID(ctx, userID)
	if err != nil {
		s.logger.Error("failed to get pull requests",
			logging.StringAttr("userID", userID),
			logging.ErrAttr(err),
		)
		return nil, err
	}

	s.logger.Info("pull request was successfully received",
		logging.StringAttr("userID", userID),
		logging.IntAttr("count pr", len(pullRequests)),
	)

	return pullRequests, nil
}
