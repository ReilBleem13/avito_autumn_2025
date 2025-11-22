package handler

import (
	"ReilBleem13/pull_requests_service/internal/domain"
	"encoding/json"
	"net/http"

	"github.com/theartofdevel/logging"
)

// POST /user/setIsActive
func (h *Handler) handleSetIsActive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req setIsActiveDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid request body", logging.ErrAttr(err))
		h.WriteError(w, domain.ErrInvalidRequest("invalid json payload"))
		return
	}

	user, teamName, err := h.svc.SetIsActive(r.Context(), req.UserID, req.IsActive)
	if err != nil {
		h.WriteError(w, err)
		return
	}

	response := getSetUserDTO{
		UserID:   user.UserID,
		Username: user.Username,
		TeamName: teamName,
		IsActive: user.IsActive,
	}

	writeJSON(w, http.StatusOK, map[string]any{"user": response})
}

func (h *Handler) handleGetReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("user_id")

	pullRequestsShort, err := h.svc.GetReview(r.Context(), userID)
	if err != nil {
		h.WriteError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"user_id":       userID,
		"pull_requests": pullRequestsShort,
	})
}
