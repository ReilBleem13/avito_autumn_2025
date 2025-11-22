package handler

import (
	"ReilBleem13/pull_requests_service/internal/service"
	"net/http"
)

type Handler struct {
	svc    *service.Service
	logger service.LoggerInterfaces
}

func NewHandler(svc *service.Service, logger service.LoggerInterfaces) *Handler {
	return &Handler{
		svc:    svc,
		logger: logger,
	}
}

func NewRouter(svc *service.Service, logger service.LoggerInterfaces) *http.ServeMux {
	h := NewHandler(svc, logger)

	mux := http.NewServeMux()

	mux.HandleFunc("/team/add", h.handleCreateTeam)
	mux.HandleFunc("/team/get", h.handleGetTeam)

	mux.HandleFunc("/users/setIsActive", h.handleSetIsActive)
	mux.HandleFunc("/users/getReview", h.handleGetReview)

	mux.HandleFunc("/pullRequest/create", h.handlePullRequestCreate)
	mux.HandleFunc("/pullRequest/merge", h.handlePullRequestMerge)
	mux.HandleFunc("/pullRequest/reassign", h.handlePullRequestReassign)

	// mux.HandleFunc("/stats", nil)
	mux.HandleFunc("/health", h.handleHealth)

	return mux
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func NewServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:    addr,
		Handler: handler,
	}
}
