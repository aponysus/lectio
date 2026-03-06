package handlers

import (
	"net/http"

	"github.com/aponysus/lectio/internal/server/httpx"
)

type HealthHandler struct{}

func (h HealthHandler) Get(w http.ResponseWriter, r *http.Request) {
	httpx.WriteData(w, http.StatusOK, map[string]string{"status": "ok"})
}
