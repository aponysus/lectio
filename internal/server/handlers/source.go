package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/aponysus/lectio/internal/model"
	"github.com/aponysus/lectio/internal/server/httpx"
	"github.com/aponysus/lectio/internal/store"
	"github.com/aponysus/lectio/internal/validation"
	"github.com/go-chi/chi/v5"
)

type SourceHandler struct {
	Store *store.Store
}

type sourceRequest struct {
	Title            string `json:"title"`
	Medium           string `json:"medium"`
	Creator          string `json:"creator"`
	Year             *int   `json:"year"`
	OriginalLanguage string `json:"original_language"`
	CultureOrContext string `json:"culture_or_context"`
	Notes            string `json:"notes"`
}

func (h SourceHandler) List(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "validation_error", "limit must be an integer")
			return
		}
		limit = parsed
	}

	filters, err := validation.NormalizeSourceFilters(model.SourceFilters{
		Query:            r.URL.Query().Get("q"),
		Medium:           r.URL.Query().Get("medium"),
		OriginalLanguage: r.URL.Query().Get("original_language"),
		Sort:             r.URL.Query().Get("sort"),
		Limit:            limit,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	sources, err := h.Store.ListSources(r.Context(), filters)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to list sources")
		return
	}

	httpx.WriteData(w, http.StatusOK, sources)
}

func (h SourceHandler) Get(w http.ResponseWriter, r *http.Request) {
	source, err := h.Store.GetSource(r.Context(), chi.URLParam(r, "sourceID"))
	if err != nil {
		h.writeStoreError(w, err, "failed to load source")
		return
	}

	httpx.WriteData(w, http.StatusOK, source)
}

func (h SourceHandler) Create(w http.ResponseWriter, r *http.Request) {
	input, ok := h.decodeAndValidate(w, r)
	if !ok {
		return
	}

	source, err := h.Store.CreateSource(r.Context(), input)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to create source")
		return
	}

	httpx.WriteData(w, http.StatusCreated, source)
}

func (h SourceHandler) Update(w http.ResponseWriter, r *http.Request) {
	input, ok := h.decodeAndValidate(w, r)
	if !ok {
		return
	}

	source, err := h.Store.UpdateSource(r.Context(), chi.URLParam(r, "sourceID"), input)
	if err != nil {
		h.writeStoreError(w, err, "failed to update source")
		return
	}

	httpx.WriteData(w, http.StatusOK, source)
}

func (h SourceHandler) Archive(w http.ResponseWriter, r *http.Request) {
	if err := h.Store.ArchiveSource(r.Context(), chi.URLParam(r, "sourceID")); err != nil {
		h.writeStoreError(w, err, "failed to archive source")
		return
	}

	httpx.WriteJSON(w, http.StatusNoContent, nil)
}

func (h SourceHandler) decodeAndValidate(w http.ResponseWriter, r *http.Request) (model.SourceInput, bool) {
	var req sourceRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "validation_error", "invalid source payload")
		return model.SourceInput{}, false
	}

	input, err := validation.NormalizeSourceInput(model.SourceInput{
		Title:            req.Title,
		Medium:           req.Medium,
		Creator:          req.Creator,
		Year:             req.Year,
		OriginalLanguage: req.OriginalLanguage,
		CultureOrContext: req.CultureOrContext,
		Notes:            req.Notes,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "validation_error", err.Error())
		return model.SourceInput{}, false
	}

	return input, true
}

func (h SourceHandler) writeStoreError(w http.ResponseWriter, err error, internalMessage string) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		httpx.WriteError(w, http.StatusNotFound, "not_found", "source not found")
	default:
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", internalMessage)
	}
}
