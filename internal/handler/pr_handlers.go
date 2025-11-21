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

	writeJSON(w, 201, createdPR)
}

// POST /pullRequest/merge
func (h *Handler) handlePullRequestMerge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

}

// POST /pullRequest/create
func (h *Handler) handlePullRequestReassign(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

}
