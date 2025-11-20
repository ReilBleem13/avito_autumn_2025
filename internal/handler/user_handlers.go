package handler

import "net/http"

// POST /user/setIsActive

func (h *Handler) handleSetIsActive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

}
