package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/aponysus/lectio/internal/server/httpx"
	"github.com/aponysus/lectio/internal/store"
)

type ExportHandler struct {
	Store *store.Store
}

func (h ExportHandler) Download(w http.ResponseWriter, r *http.Request) {
	payload, err := h.Store.ExportData(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to export data")
		return
	}

	filename := "lectio-export-" + time.Now().UTC().Format("2006-01-02") + ".json"
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w.WriteHeader(http.StatusOK)

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(payload); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to write export")
		return
	}
}
