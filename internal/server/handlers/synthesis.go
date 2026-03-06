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

type SynthesisHandler struct {
	Store *store.Store
}

type synthesisRequest struct {
	Title     string `json:"title"`
	Body      string `json:"body"`
	Type      string `json:"type"`
	InquiryID string `json:"inquiry_id"`
	Notes     string `json:"notes"`
}

func (h SynthesisHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, ok := parseListLimit(w, r, 50)
	if !ok {
		return
	}

	filters, err := validation.NormalizeSynthesisFilters(model.SynthesisFilters{
		Limit: limit,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	syntheses, err := h.Store.ListSyntheses(r.Context(), filters)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to list syntheses")
		return
	}

	httpx.WriteData(w, http.StatusOK, syntheses)
}

func (h SynthesisHandler) ListInquirySyntheses(w http.ResponseWriter, r *http.Request) {
	limit, ok := parseListLimit(w, r, 50)
	if !ok {
		return
	}

	syntheses, err := h.Store.ListInquirySyntheses(r.Context(), chi.URLParam(r, "inquiryID"), limit)
	if err != nil {
		h.writeStoreError(w, err, "failed to list inquiry syntheses")
		return
	}

	httpx.WriteData(w, http.StatusOK, syntheses)
}

func (h SynthesisHandler) Get(w http.ResponseWriter, r *http.Request) {
	synthesis, err := h.Store.GetSynthesis(r.Context(), chi.URLParam(r, "synthesisID"))
	if err != nil {
		h.writeStoreError(w, err, "failed to load synthesis")
		return
	}

	httpx.WriteData(w, http.StatusOK, synthesis)
}

func (h SynthesisHandler) Create(w http.ResponseWriter, r *http.Request) {
	input, ok := h.decodeAndValidate(w, r)
	if !ok {
		return
	}

	synthesis, err := h.Store.CreateSynthesis(r.Context(), input)
	if err != nil {
		h.writeStoreError(w, err, "failed to create synthesis")
		return
	}

	httpx.WriteData(w, http.StatusCreated, synthesis)
}

func (h SynthesisHandler) Update(w http.ResponseWriter, r *http.Request) {
	input, ok := h.decodeAndValidate(w, r)
	if !ok {
		return
	}

	synthesis, err := h.Store.UpdateSynthesis(r.Context(), chi.URLParam(r, "synthesisID"), input)
	if err != nil {
		h.writeStoreError(w, err, "failed to update synthesis")
		return
	}

	httpx.WriteData(w, http.StatusOK, synthesis)
}

func (h SynthesisHandler) Archive(w http.ResponseWriter, r *http.Request) {
	if err := h.Store.ArchiveSynthesis(r.Context(), chi.URLParam(r, "synthesisID")); err != nil {
		h.writeStoreError(w, err, "failed to archive synthesis")
		return
	}

	httpx.WriteJSON(w, http.StatusNoContent, nil)
}

func (h SynthesisHandler) decodeAndValidate(w http.ResponseWriter, r *http.Request) (model.SynthesisInput, bool) {
	var req synthesisRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "validation_error", "invalid synthesis payload")
		return model.SynthesisInput{}, false
	}

	input, err := validation.NormalizeSynthesisInput(model.SynthesisInput{
		Title:     req.Title,
		Body:      req.Body,
		Type:      req.Type,
		InquiryID: req.InquiryID,
		Notes:     req.Notes,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "validation_error", err.Error())
		return model.SynthesisInput{}, false
	}

	return input, true
}

func (h SynthesisHandler) writeStoreError(w http.ResponseWriter, err error, internalMessage string) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		httpx.WriteError(w, http.StatusNotFound, "not_found", "synthesis or inquiry not found")
	default:
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", internalMessage)
	}
}

func parseListLimit(w http.ResponseWriter, r *http.Request, defaultLimit int) (int, bool) {
	limit := defaultLimit
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "validation_error", "limit must be an integer")
			return 0, false
		}
		limit = parsed
	}

	return limit, true
}
