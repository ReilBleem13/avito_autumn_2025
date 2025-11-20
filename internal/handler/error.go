package handler

import (
	"ReilBleem13/pull_requests_service/internal/domain"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/theartofdevel/logging"
)

type apiErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeAPIResponse(w http.ResponseWriter, status int, code, message string) {
	resp := apiErrorResponse{}
	resp.Error.Code = code
	resp.Error.Message = message
	writeJSON(w, status, resp)
}

func (h *Handler) WriteError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	var appErr *domain.AppError
	if errors.As(err, &appErr) {
		h.logger.Error("application error",
			logging.StringAttr("code", string(appErr.Code)),
			logging.StringAttr("message", appErr.Message),
			logging.ErrAttr(appErr.Cause),
		)
		writeAPIResponse(w, statusFromCode(appErr.Code), string(appErr.Code), appErr.Message)
		return
	}
	h.logger.Error("unexpected error", logging.ErrAttr(err))
	writeAPIResponse(w, http.StatusInternalServerError, "SERVER_ERROR", "internal server error")
}

func statusFromCode(code domain.ErrorCode) int {
	switch code {
	case domain.CodeTeamExists:
		return http.StatusBadRequest
	case domain.CodePRExists, domain.CodePRMerged, domain.CodeNotAssigned, domain.CodeNoCandidate:
		return http.StatusConflict
	case domain.CodeNotFound:
		return http.StatusNotFound
	default:
		return http.StatusNotFound
	}
}
