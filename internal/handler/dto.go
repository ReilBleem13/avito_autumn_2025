package handler

import "ReilBleem13/pull_requests_service/internal/domain"

type createTeamDTO struct {
	TeamName string        `json:"team_name"`
	Members  []domain.User `json:"members"`
}
