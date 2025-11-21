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
		s.logger.Error("failed to set user status",
			logging.StringAttr("userID", userID),
			logging.StringAttr("error", "user_id is empty"),
		)
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
