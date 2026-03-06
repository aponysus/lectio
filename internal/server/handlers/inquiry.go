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

type InquiryHandler struct {
	Store *store.Store
}

type inquiryRequest struct {
	Title        string `json:"title"`
	Question     string `json:"question"`
	Status       string `json:"status"`
	WhyItMatters string `json:"why_it_matters"`
	CurrentView  string `json:"current_view"`
	OpenTensions string `json:"open_tensions"`
}

func (h InquiryHandler) List(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "validation_error", "limit must be an integer")
			return
		}
		limit = parsed
	}

	filters, err := validation.NormalizeInquiryFilters(model.InquiryFilters{
		Query:  r.URL.Query().Get("q"),
		Status: r.URL.Query().Get("status"),
		Limit:  limit,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	inquiries, err := h.Store.ListInquiries(r.Context(), filters)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to list inquiries")
		return
	}

	httpx.WriteData(w, http.StatusOK, inquiries)
}

func (h InquiryHandler) Get(w http.ResponseWriter, r *http.Request) {
	inquiry, err := h.Store.GetInquiry(r.Context(), chi.URLParam(r, "inquiryID"))
	if err != nil {
		h.writeStoreError(w, err, "failed to load inquiry")
		return
	}

	httpx.WriteData(w, http.StatusOK, inquiry)
}

func (h InquiryHandler) ListEligibleForSynthesis(w http.ResponseWriter, r *http.Request) {
	limit, ok := parseListLimit(w, r, 6)
	if !ok {
		return
	}

	inquiries, err := h.Store.ListEligibleForSynthesisInquiries(r.Context(), limit)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to list synthesis-eligible inquiries")
		return
	}

	httpx.WriteData(w, http.StatusOK, inquiries)
}

func (h InquiryHandler) ListEngagements(w http.ResponseWriter, r *http.Request) {
	limit := 20
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "validation_error", "limit must be an integer")
			return
		}
		limit = parsed
	}

	engagements, err := h.Store.ListInquiryEngagements(r.Context(), chi.URLParam(r, "inquiryID"), limit)
	if err != nil {
		h.writeStoreError(w, err, "failed to list inquiry engagements")
		return
	}

	httpx.WriteData(w, http.StatusOK, engagements)
}

func (h InquiryHandler) Create(w http.ResponseWriter, r *http.Request) {
	input, ok := h.decodeAndValidate(w, r)
	if !ok {
		return
	}

	inquiry, err := h.Store.CreateInquiry(r.Context(), input)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to create inquiry")
		return
	}

	httpx.WriteData(w, http.StatusCreated, inquiry)
}

func (h InquiryHandler) Update(w http.ResponseWriter, r *http.Request) {
	input, ok := h.decodeAndValidate(w, r)
	if !ok {
		return
	}

	inquiry, err := h.Store.UpdateInquiry(r.Context(), chi.URLParam(r, "inquiryID"), input)
	if err != nil {
		h.writeStoreError(w, err, "failed to update inquiry")
		return
	}

	httpx.WriteData(w, http.StatusOK, inquiry)
}

func (h InquiryHandler) Archive(w http.ResponseWriter, r *http.Request) {
	if err := h.Store.ArchiveInquiry(r.Context(), chi.URLParam(r, "inquiryID")); err != nil {
		h.writeStoreError(w, err, "failed to archive inquiry")
		return
	}

	httpx.WriteJSON(w, http.StatusNoContent, nil)
}

func (h InquiryHandler) decodeAndValidate(w http.ResponseWriter, r *http.Request) (model.InquiryInput, bool) {
	var req inquiryRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "validation_error", "invalid inquiry payload")
		return model.InquiryInput{}, false
	}

	input, err := validation.NormalizeInquiryInput(model.InquiryInput{
		Title:        req.Title,
		Question:     req.Question,
		Status:       req.Status,
		WhyItMatters: req.WhyItMatters,
		CurrentView:  req.CurrentView,
		OpenTensions: req.OpenTensions,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "validation_error", err.Error())
		return model.InquiryInput{}, false
	}

	return input, true
}

func (h InquiryHandler) writeStoreError(w http.ResponseWriter, err error, internalMessage string) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		httpx.WriteError(w, http.StatusNotFound, "not_found", "inquiry not found")
	default:
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", internalMessage)
	}
}
