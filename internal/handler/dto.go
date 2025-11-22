package handler

import (
	"ReilBleem13/pull_requests_service/internal/domain"
	"time"
)

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

type doMergedRequestDTO struct {
	PullRequestID string `json:"pull_request_id"`
}

type reassignDTO struct {
	PullRequestID string `json:"pull_request_id"`
	OldUserID     string `json:"old_user_id"`
}

// response
type getTeamUsersDTO struct {
	TeamName string        `json:"team_name"`
	Members  []domain.User `json:"members"`
}

type getSetUserDTO struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

type getMergedDTO struct {
	PullRequestID     string          `json:"pull_request_id"`
	PullRequestName   string          `json:"pull_request_name"`
	AuthorID          string          `json:"author_id"`
	Status            domain.PRStatus `json:"status"`
	AssignedReviewers []string        `json:"assigned_reviewers"`
	MergedAt          *time.Time      `json:"merged_at"`
}

type getAssignedDTO struct {
	PullRequestID     string          `json:"pull_request_id"`
	PullRequestName   string          `json:"pull_request_name"`
	AuthorID          string          `json:"author_id"`
	Status            domain.PRStatus `json:"status"`
	AssignedReviewers []string        `json:"assigned_reviewers"`
}
