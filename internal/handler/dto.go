package handler

import "ReilBleem13/pull_requests_service/internal/domain"

// requests
type createTeamDTO struct {
	TeamName string        `json:"team_name"`
	Members  []domain.User `json:"members"`
}

type setIsActiveDTO struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

// response
type getTeamUsersDTO struct {
	TeamName string        `json:"team_name"`
	Members  []domain.User `json:"members"`
}

type getSetUserDTO struct {
	User     *domain.User `json:"user"`
	TeamName string       `json:"team_name"`
}
