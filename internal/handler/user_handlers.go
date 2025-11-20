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
		User:     user,
		TeamName: teamName,
	}

	writeJSON(w, http.StatusOK, response)
}
