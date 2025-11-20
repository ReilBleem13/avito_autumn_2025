package handler

import (
	"ReilBleem13/pull_requests_service/internal/domain"
	"encoding/json"
	"net/http"

	"github.com/theartofdevel/logging"
)

// POST /team/add
func (h *Handler) handleCreateTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req createTeamDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid request body", logging.ErrAttr(err))
		h.WriteError(w, domain.ErrInvalidRequest("invalid json payload"))
		return
	}

	if err := h.svc.Create(r.Context(), req.TeamName, req.Members); err != nil {
		h.WriteError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"status": "team created"})
}

func (h *Handler) handleGetTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	teamName := r.URL.Query().Get("team_name")
	users, err := h.svc.Get(r.Context(), teamName)
	if err != nil {
		h.WriteError(w, err)
		return
	}

	response := getTeamUsersDTO{
		TeamName: teamName,
		Members:  users,
	}

	writeJSON(w, http.StatusOK, response)
}
