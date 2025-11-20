package domain

type ErrorCode string

const (
	CodeTeamExists  ErrorCode = "TEAM_EXISTS"
	CodePRExists    ErrorCode = "PR_EXISTS"
	CodePRMerged    ErrorCode = "PR_MERGED"
	CodeNotAssigned ErrorCode = "NOT_ASSIGNED"
	CodeNoCandidate ErrorCode = "NO_CANDIDATE"
	CodeNotFound    ErrorCode = "NOT_FOUND"

	CodeInvalidRequest   ErrorCode = "INVALID_REQUEST"
	CodeTeamNameEmpty    ErrorCode = "TEAM_NAME_EMPTY"
	CodeTeamMembersEmpty ErrorCode = "TEAM_MEMBERS_EMPTY"
	CodeInternalError    ErrorCode = "INTERNAL_SERVER_ERROR"
)

type AppError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Cause   error     `json:"-"`
}

func (e *AppError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return string(e.Code)
}

func (e *AppError) Unwrap() error { return e.Cause }

func ErrTeamExists() error {
	return &AppError{Code: CodeTeamExists, Message: "team_name already exists"}
}

func ErrPRExists() error {
	return &AppError{Code: CodePRExists, Message: "PR id already exists"}
}

func ErrPRMerged() error {
	return &AppError{Code: CodePRMerged, Message: "cannot reassign on merged PR"}
}

func ErrNotAssigned() error {
	return &AppError{Code: CodeNotAssigned, Message: "reviewer is not assigned to this PR"}
}

func ErrNoCandidate() error {
	return &AppError{Code: CodeNoCandidate, Message: "no active replacement candidate in team"}
}

func ErrNotFound() error {
	return &AppError{Code: CodeNotFound, Message: "resource not found"}
}

func ErrInvalidRequest(msg string) error {
	return &AppError{Code: CodeInvalidRequest, Message: msg}
}
