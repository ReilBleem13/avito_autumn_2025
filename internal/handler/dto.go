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

type createPullRequestDTO struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
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
