package handler

import (
	"ReilBleem13/pull_requests_service/internal/domain"
	"encoding/json"
	"net/http"

	"github.com/theartofdevel/logging"
)

// POST /pullRequest/create
func (h *Handler) handlePullRequestCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req createPullRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid request body", logging.ErrAttr(err))
		h.WriteError(w, domain.ErrInvalidRequest("invalid json payload"))
		return
	}

	createdPR, err := h.svc.CreatePullRequest(r.Context(), req.PullRequestID, req.PullRequestName, req.AuthorID)
	if err != nil {
		h.WriteError(w, err)
		return
	}

	// добавить структуру для ответа
	writeJSON(w, 201, map[string]any{"pr": createdPR})
}

// POST /pullRequest/merge
func (h *Handler) handlePullRequestMerge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req doMergedRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid request body", logging.ErrAttr(err))
		h.WriteError(w, domain.ErrInvalidRequest("invalid json payload"))
		return
	}

	pullRequest, err := h.svc.MergePullRequest(r.Context(), req.PullRequestID)
	if err != nil {
		h.WriteError(w, err)
		return
	}

	response := getMergedDTO{
		PullRequestID:     pullRequest.PullRequestID,
		PullRequestName:   pullRequest.PullRequestName,
		AuthorID:          pullRequest.AuthorID,
		Status:            pullRequest.Status,
		AssignedReviewers: pullRequest.AssignedReviewers,
		MergedAt:          pullRequest.MergedAt,
	}

	writeJSON(w, 200, map[string]any{"pr": response})
}

// POST /pullRequest/create
func (h *Handler) handlePullRequestReassign(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req reassignDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid request body", logging.ErrAttr(err))
		h.WriteError(w, domain.ErrInvalidRequest("invalid json payload"))
		return
	}

	pullRequest, newReviewerID, err := h.svc.ReAssign(r.Context(), req.PullRequestID, req.OldUserID)
	if err != nil {
		h.WriteError(w, err)
		return
	}

	response := getAssignedDTO{
		PullRequestID:     pullRequest.PullRequestID,
		PullRequestName:   pullRequest.PullRequestName,
		AuthorID:          pullRequest.AuthorID,
		Status:            pullRequest.Status,
		AssignedReviewers: pullRequest.AssignedReviewers,
	}

	writeJSON(w, 200,
		map[string]any{
			"pr":          response,
			"replaced_by": newReviewerID,
		},
	)
}
