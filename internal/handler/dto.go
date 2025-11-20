package handler

import "ReilBleem13/pull_requests_service/internal/domain"

// requests
type createTeamDTO struct {
	TeamName string        `json:"team_name"`
	Members  []domain.User `json:"members"`
}

// response
type getTeamUsersDTO struct {
	TeamName string        `json:"team_name"`
	Members  []domain.User `json:"members"`
}
