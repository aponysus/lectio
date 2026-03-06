package handlers

import (
	"errors"
	"net/http"

	"github.com/aponysus/lectio/internal/server/httpx"
	"github.com/aponysus/lectio/internal/store"
	"github.com/aponysus/lectio/internal/validation"
	"github.com/go-chi/chi/v5"
)

type EngagementInquiryHandler struct {
	Store *store.Store
}

type replaceEngagementInquiryRequest struct {
	InquiryIDs []string `json:"inquiry_ids"`
}

func (h EngagementInquiryHandler) List(w http.ResponseWriter, r *http.Request) {
	inquiries, err := h.Store.ListEngagementInquiries(r.Context(), chi.URLParam(r, "engagementID"))
	if err != nil {
		h.writeStoreError(w, err, "failed to load engagement inquiries")
		return
	}

	httpx.WriteData(w, http.StatusOK, inquiries)
}

func (h EngagementInquiryHandler) Replace(w http.ResponseWriter, r *http.Request) {
	var req replaceEngagementInquiryRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "validation_error", "invalid engagement inquiries payload")
		return
	}

	if err := h.Store.ReplaceEngagementInquiries(
		r.Context(),
		chi.URLParam(r, "engagementID"),
		validation.NormalizeInquiryIDs(req.InquiryIDs),
	); err != nil {
		h.writeStoreError(w, err, "failed to replace engagement inquiries")
		return
	}

	httpx.WriteJSON(w, http.StatusNoContent, nil)
}

func (h EngagementInquiryHandler) writeStoreError(w http.ResponseWriter, err error, internalMessage string) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		httpx.WriteError(w, http.StatusNotFound, "not_found", "engagement or inquiry not found")
	default:
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", internalMessage)
	}
}
