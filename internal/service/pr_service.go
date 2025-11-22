package service

import (
	"ReilBleem13/pull_requests_service/internal/domain"
	"context"
	"database/sql"
	"errors"
	"math/rand"

	"github.com/theartofdevel/logging"
)

func (s *Service) CreatePullRequest(ctx context.Context, prID, prName, authorID string) (*domain.PullRequest, error) {
	s.logger.Info("attempt to create pr",
		logging.StringAttr("prID", prID),
		logging.StringAttr("prName", prName),
		logging.StringAttr("authorID", authorID),
	)

	if prID == "" {
		s.logger.Error("failed to create pull request")
		return nil, domain.ErrInvalidRequest("pr_id is empty")
	}

	if prName == "" {
		s.logger.Error("failed to create pull request",
			logging.StringAttr("prID", prID),
		)
		return nil, domain.ErrInvalidRequest("pr_name is empty")
	}

	if authorID == "" {
		s.logger.Error("failed to create pull request",
			logging.StringAttr("prID", prID),
			logging.StringAttr("prName", prName),
		)
		return nil, domain.ErrInvalidRequest("author_id is empty")
	}

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

func (s *Service) MergePullRequest(ctx context.Context, prID string) (*domain.PullRequest, error) {
	s.logger.Info("attempt to merge pr",
		logging.StringAttr("prID", prID),
	)

	if prID == "" {
		s.logger.Error("failed to create pull request")
		return nil, domain.ErrInvalidRequest("pr_id is empty")
	}

	err := s.prs.Merge(ctx, prID)
	if err != nil {
		s.logger.Error("failed to merge pr",
			logging.StringAttr("prID", prID),
		)
		return nil, err
	}

	pullRequest, err := s.prs.GetPullRequest(ctx, prID)
	if err != nil {
		s.logger.Error("failed to get pull request",
			logging.StringAttr("prID", prID),
		)
		return nil, err
	}

	s.logger.Info("pr was successfully merged",
		logging.StringAttr("prID", prID),
	)

	return pullRequest, nil
}

func (s *Service) ReAssign(ctx context.Context, prID, oldReviewerID string) (*domain.PullRequest, string, error) {
	s.logger.Info("attempt to reassign",
		logging.StringAttr("prID", prID),
		logging.StringAttr("oldReviewerID", oldReviewerID),
	)

	if prID == "" {
		s.logger.Error("failed to reassign")
		return nil, "", domain.ErrInvalidRequest("pr_id is empty")
	}

	if oldReviewerID == "" {
		s.logger.Error("failed to reassign",
			logging.StringAttr("prID", prID),
		)
		return nil, "", domain.ErrInvalidRequest("old_reviewer_id is empty")
	}

	pullRequest, err := s.prs.GetPullRequest(ctx, prID)
	if err != nil {
		s.logger.Error("failed to reassign",
			logging.StringAttr("prID", prID),
			logging.StringAttr("oldReviewerID", oldReviewerID),
		)
		return nil, "", err
	}

	if pullRequest.Status == domain.PRStatusMerged {
		s.logger.Error("failed to reassign, pr already merged",
			logging.StringAttr("prID", prID),
			logging.StringAttr("oldReviewerID", oldReviewerID),
		)
		return nil, "", domain.ErrPRMerged()
	}

	isOldIn := false
	for _, ar := range pullRequest.AssignedReviewers {
		if ar == oldReviewerID {
			isOldIn = true
			break
		}
	}

	if !isOldIn {
		s.logger.Error("failed to reassign, old reviewer is not assigned",
			logging.StringAttr("prID", prID),
			logging.StringAttr("oldReviewerID", oldReviewerID),
		)
		return nil, "", domain.ErrNotAssigned()
	}

	teamName, err := s.users.GetTeamName(ctx, oldReviewerID)
	if err != nil {
		s.logger.Error("failed to reassign, failed to get team name",
			logging.StringAttr("prID", prID),
			logging.StringAttr("oldReviewerID", oldReviewerID),
		)
		return nil, "", err
	}

	teamMembers, err := s.teams.Get(ctx, teamName)
	if err != nil {
		s.logger.Error("failed to reassign, failed to get team members",
			logging.StringAttr("prID", prID),
			logging.StringAttr("oldReviewerID", oldReviewerID),
		)
		return nil, "", err
	}

	// добавить проверку, что новый кандидат это не автор!
	candidates := make([]string, 0)
	for _, tm := range teamMembers {
		if tm.UserID == oldReviewerID {
			continue
		}

		alreadyAssigned := false
		for _, r := range pullRequest.AssignedReviewers {
			if r == tm.UserID {
				alreadyAssigned = true
				break
			}
		}

		if alreadyAssigned || tm.UserID == pullRequest.AuthorID {
			continue
		}
		candidates = append(candidates, tm.UserID)
	}

	if len(candidates) == 0 {
		s.logger.Error("failed to reassign, failed to get team members",
			logging.StringAttr("prID", prID),
			logging.StringAttr("oldReviewerID", oldReviewerID),
		)
		return nil, "", domain.ErrNoCandidate()
	}

	newReviewerID := candidates[rand.Intn(len(candidates))]

	if err := s.prs.ReAssign(ctx, prID, oldReviewerID, newReviewerID); err != nil {
		s.logger.Error("failed to reassign, failed to replace reviewers",
			logging.StringAttr("prID", prID),
			logging.StringAttr("oldReviewerID", oldReviewerID),
			logging.StringAttr("newReviewerID", newReviewerID),
		)
		return nil, "", err
	}

	updatedPullRequest, err := s.prs.GetPullRequest(ctx, prID)
	if err != nil {
		s.logger.Error("failed to reassign",
			logging.StringAttr("prID", prID),
			logging.StringAttr("oldReviewerID", oldReviewerID),
		)
		return nil, "", err
	}

	s.logger.Info("pr was successfully reassigned",
		logging.StringAttr("prID", prID),
		logging.StringAttr("oldReviewerID", oldReviewerID),
		logging.StringAttr("newReviewerID", newReviewerID),
	)

	return updatedPullRequest, newReviewerID, nil
}
