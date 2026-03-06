package handlers

import (
	"net/http"

	"github.com/aponysus/lectio/internal/config"
	"github.com/aponysus/lectio/internal/server/httpx"
	"github.com/aponysus/lectio/internal/store"
)

type SystemHandler struct {
	Config config.Config
	Store  *store.Store
}

func (h SystemHandler) Status(w http.ResponseWriter, r *http.Request) {
	status, err := h.Store.SystemStatus(r.Context(), h.Config.Env)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to load system status")
		return
	}

	httpx.WriteData(w, http.StatusOK, status)
}
