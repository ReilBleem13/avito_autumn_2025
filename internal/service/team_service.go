package service

import (
	"ReilBleem13/pull_requests_service/internal/domain"
	"context"
	"database/sql"
	"errors"

	"github.com/theartofdevel/logging"
)

func (s *Service) CreateTeam(ctx context.Context, teamName string, users []domain.User) error {
	s.logger.Info("attempt to create team",
		logging.StringAttr("team_name", teamName),
		logging.IntAttr("quantity of users", len(users)),
	)

	if teamName == "" {
		s.logger.Error("failed to create team",
			logging.StringAttr("error", "team_name is empty"),
		)
		return domain.ErrInvalidRequest("team_name is empty")
	}

	if len(users) == 0 {
		s.logger.Error("failed to create team",
			logging.StringAttr("team_name", teamName),
			logging.StringAttr("error", "count of users is 0"),
		)
		return domain.ErrInvalidRequest("team_users is empty")
	}

	if err := s.teams.Create(ctx, teamName, users); err != nil {
		s.logger.Error("failed to create team",
			logging.StringAttr("team_name", teamName),
			logging.ErrAttr(err),
		)
		return err
	}

	s.logger.Info("team was succeccfully created",
		logging.StringAttr("team_name", teamName),
		logging.IntAttr("quantity of users", len(users)),
	)
	return nil
}

func (s *Service) GetTeam(ctx context.Context, teamName string) ([]domain.User, error) {
	s.logger.Info("attempt to get team members",
		logging.StringAttr("team_name", teamName),
	)

	teamMembers, err := s.teams.Get(ctx, teamName)
	if err != nil {
		s.logger.Error("failed to get team members",
			logging.StringAttr("team_name", teamName),
			logging.ErrAttr(err),
		)

		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound()
		}
		return nil, err
	}

	if len(teamMembers) == 0 {
		s.logger.Info("team exists but has no members",
			logging.StringAttr("team_name", teamName))
		return nil, domain.ErrNotFound()
	}

	s.logger.Info("team members was successfully received",
		logging.StringAttr("team_name", teamName),
		logging.IntAttr("quantity of team members", len(teamMembers)),
	)
	return teamMembers, nil
}
