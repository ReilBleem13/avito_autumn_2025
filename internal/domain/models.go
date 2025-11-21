package domain

import "time"

type TeamMember struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"` // поменять поле на team name
}

type Team struct {
	TeamName string       `json:"team_name"`
	Members  []TeamMember `json:"members"`
}

type User struct {
	UserID   string `db:"user_id" json:"user_id"`
	Username string `db:"username" json:"username"`
	IsActive bool   `db:"is_active" json:"is_active"`
}

type PRStatus string

const (
	PRStatusOpen   PRStatus = "OPEN"
	PRStatusMerged PRStatus = "MERGED"
)

func (s PRStatus) String() string {
	return string(s)
}

type PullRequest struct {
	PullRequestID     string     `json:"pull_request_id"`
	PullRequestName   string     `json:"pull_request_name"`
	AuthorID          string     `json:"author_id"`
	Status            PRStatus   `json:"status"`
	AssignedReviewers []string   `json:"assigned_reviewers"`
	CreatedAt         *time.Time `json:"createdAt,omitempty"`
	MergedAt          *time.Time `json:"mergedAt,omitempty"`
}

type PullRequestShort struct {
	PullRequestID   string   `json:"pull_request_id"`
	PullRequestName string   `json:"pull_request_name"`
	AuthorID        string   `json:"author_id"`
	Status          PRStatus `json:"status"`
}
