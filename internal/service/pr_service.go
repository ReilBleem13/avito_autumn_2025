package service

import (
	"ReilBleem13/pull_requests_service/internal/domain"
	"context"
	"database/sql"
	"errors"

	"github.com/theartofdevel/logging"
)

func (s *Service) CreatePullRequest(ctx context.Context, prID, prName, authorID string) (*domain.PullRequest, error) {
	s.logger.Info("attempt to create pr",
		logging.StringAttr("prID", prID),
		logging.StringAttr("prName", prName),
		logging.StringAttr("authorID", authorID),
	)

	_, err := s.users.GetUser(ctx, authorID)
	if err != nil {
		s.logger.Error("failed to create pr",
			logging.StringAttr("prID", prID),
			logging.StringAttr("prName", prName),
			logging.StringAttr("authorID", authorID),
			logging.ErrAttr(err),
		)

		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound()
		}
		return nil, err
	}

	teamName, err := s.users.GetTeamName(ctx, authorID)
	if err != nil {
		s.logger.Error("failed to create pr",
			logging.StringAttr("prID", prID),
			logging.StringAttr("prName", prName),
			logging.StringAttr("authorID", authorID),
			logging.ErrAttr(err),
		)
		return nil, err
	}

	candidates, err := s.prs.GetActiveTeamMembers(ctx, teamName, authorID)
	if err != nil {
		s.logger.Error("failed to create pr",
			logging.StringAttr("prID", prID),
			logging.StringAttr("prName", prName),
			logging.StringAttr("authorID", authorID),
			logging.ErrAttr(err),
		)
		return nil, err
	}

	assignedUsers := make([]string, 0, 2)
	for i := 0; i < len(candidates) && i < 2; i++ {
		assignedUsers = append(assignedUsers, candidates[i].UserID)
	}

	if err := s.prs.Create(ctx, prID, prName, authorID, assignedUsers); err != nil {
		s.logger.Error("failed to create pr",
			logging.StringAttr("prID", prID),
			logging.StringAttr("prName", prName),
			logging.StringAttr("authorID", authorID),
			logging.ErrAttr(err),
		)
		return nil, err
	}

	s.logger.Info("pr was successfully created",
		logging.StringAttr("prID", prID),
		logging.StringAttr("prName", prName),
		logging.StringAttr("authorID", authorID),
	)
	return &domain.PullRequest{
		PullRequestID:     prID,
		PullRequestName:   prName,
		AuthorID:          authorID,
		Status:            domain.PRStatusOpen,
		AssignedReviewers: assignedUsers,
	}, nil
}
