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

type EngagementHandler struct {
	Store *store.Store
}

type engagementRequest struct {
	SourceID           string `json:"source_id"`
	EngagedAt          string `json:"engaged_at"`
	PortionLabel       string `json:"portion_label"`
	Reflection         string `json:"reflection"`
	WhyItMatters       string `json:"why_it_matters"`
	SourceLanguage     string `json:"source_language"`
	ReflectionLanguage string `json:"reflection_language"`
	AccessMode         string `json:"access_mode"`
	RevisitPriority    *int   `json:"revisit_priority"`
	IsRereadOrRewatch  bool   `json:"is_reread_or_rewatch"`
}

func (h EngagementHandler) List(w http.ResponseWriter, r *http.Request) {
	limit := 20
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "validation_error", "limit must be an integer")
			return
		}
		limit = parsed
	}

	filters, err := validation.NormalizeEngagementFilters(model.EngagementFilters{
		SourceID:   r.URL.Query().Get("source_id"),
		AccessMode: r.URL.Query().Get("access_mode"),
		Limit:      limit,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	engagements, err := h.Store.ListEngagements(r.Context(), filters)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to list engagements")
		return
	}

	httpx.WriteData(w, http.StatusOK, engagements)
}

func (h EngagementHandler) Get(w http.ResponseWriter, r *http.Request) {
	engagement, err := h.Store.GetEngagement(r.Context(), chi.URLParam(r, "engagementID"))
	if err != nil {
		h.writeStoreError(w, err, "failed to load engagement")
		return
	}

	httpx.WriteData(w, http.StatusOK, engagement)
}

func (h EngagementHandler) Create(w http.ResponseWriter, r *http.Request) {
	input, ok := h.decodeAndValidate(w, r)
	if !ok {
		return
	}

	engagement, err := h.Store.CreateEngagement(r.Context(), input)
	if err != nil {
		h.writeStoreError(w, err, "failed to create engagement")
		return
	}

	httpx.WriteData(w, http.StatusCreated, engagement)
}

func (h EngagementHandler) Update(w http.ResponseWriter, r *http.Request) {
	input, ok := h.decodeAndValidate(w, r)
	if !ok {
		return
	}

	engagement, err := h.Store.UpdateEngagement(r.Context(), chi.URLParam(r, "engagementID"), input)
	if err != nil {
		h.writeStoreError(w, err, "failed to update engagement")
		return
	}

	httpx.WriteData(w, http.StatusOK, engagement)
}

func (h EngagementHandler) Archive(w http.ResponseWriter, r *http.Request) {
	if err := h.Store.ArchiveEngagement(r.Context(), chi.URLParam(r, "engagementID")); err != nil {
		h.writeStoreError(w, err, "failed to archive engagement")
		return
	}

	httpx.WriteJSON(w, http.StatusNoContent, nil)
}

func (h EngagementHandler) decodeAndValidate(w http.ResponseWriter, r *http.Request) (model.EngagementInput, bool) {
	var req engagementRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "validation_error", "invalid engagement payload")
		return model.EngagementInput{}, false
	}

	input, err := validation.NormalizeEngagementInput(model.EngagementInput{
		SourceID:           req.SourceID,
		EngagedAt:          req.EngagedAt,
		PortionLabel:       req.PortionLabel,
		Reflection:         req.Reflection,
		WhyItMatters:       req.WhyItMatters,
		SourceLanguage:     req.SourceLanguage,
		ReflectionLanguage: req.ReflectionLanguage,
		AccessMode:         req.AccessMode,
		RevisitPriority:    req.RevisitPriority,
		IsRereadOrRewatch:  req.IsRereadOrRewatch,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "validation_error", err.Error())
		return model.EngagementInput{}, false
	}

	return input, true
}

func (h EngagementHandler) writeStoreError(w http.ResponseWriter, err error, internalMessage string) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		httpx.WriteError(w, http.StatusNotFound, "not_found", "engagement or source not found")
	default:
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", internalMessage)
	}
}
