package handler

import (
	"net/http"

	"github.com/sealsurlaw/gouvre/response"
)

func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.WriteHeader(http.StatusOK)
		return
	} else {
		response.SendMethodNotFound(w)
		return
	}
}
