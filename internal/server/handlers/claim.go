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

type ClaimHandler struct {
	Store *store.Store
}

type claimRequest struct {
	Text               string   `json:"text"`
	ClaimType          string   `json:"claim_type"`
	Confidence         *int     `json:"confidence"`
	Status             string   `json:"status"`
	OriginEngagementID string   `json:"origin_engagement_id"`
	Notes              string   `json:"notes"`
	InquiryIDs         []string `json:"inquiry_ids"`
}

func (h ClaimHandler) List(w http.ResponseWriter, r *http.Request) {
	limit := 20
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "validation_error", "limit must be an integer")
			return
		}
		limit = parsed
	}

	filters, err := validation.NormalizeClaimFilters(model.ClaimFilters{
		Query: r.URL.Query().Get("q"),
		Limit: limit,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	claims, err := h.Store.ListClaims(r.Context(), filters)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to list claims")
		return
	}

	httpx.WriteData(w, http.StatusOK, claims)
}

func (h ClaimHandler) ListInquiryClaims(w http.ResponseWriter, r *http.Request) {
	claims, err := h.Store.ListInquiryClaims(r.Context(), chi.URLParam(r, "inquiryID"))
	if err != nil {
		h.writeStoreError(w, err, "failed to list inquiry claims")
		return
	}

	httpx.WriteData(w, http.StatusOK, claims)
}

func (h ClaimHandler) ListEngagementClaims(w http.ResponseWriter, r *http.Request) {
	claims, err := h.Store.ListEngagementClaims(r.Context(), chi.URLParam(r, "engagementID"))
	if err != nil {
		h.writeStoreError(w, err, "failed to list engagement claims")
		return
	}

	httpx.WriteData(w, http.StatusOK, claims)
}

func (h ClaimHandler) Get(w http.ResponseWriter, r *http.Request) {
	claim, err := h.Store.GetClaim(r.Context(), chi.URLParam(r, "claimID"))
	if err != nil {
		h.writeStoreError(w, err, "failed to load claim")
		return
	}

	httpx.WriteData(w, http.StatusOK, claim)
}

func (h ClaimHandler) Create(w http.ResponseWriter, r *http.Request) {
	input, inquiryIDs, ok := h.decodeAndValidate(w, r)
	if !ok {
		return
	}

	claim, err := h.Store.CreateClaim(r.Context(), input, inquiryIDs)
	if err != nil {
		h.writeStoreError(w, err, "failed to create claim")
		return
	}

	httpx.WriteData(w, http.StatusCreated, claim)
}

func (h ClaimHandler) Update(w http.ResponseWriter, r *http.Request) {
	input, _, ok := h.decodeAndValidate(w, r)
	if !ok {
		return
	}

	claim, err := h.Store.UpdateClaim(r.Context(), chi.URLParam(r, "claimID"), input)
	if err != nil {
		h.writeStoreError(w, err, "failed to update claim")
		return
	}

	httpx.WriteData(w, http.StatusOK, claim)
}

func (h ClaimHandler) ReplaceInquiries(w http.ResponseWriter, r *http.Request) {
	var req claimRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "validation_error", "invalid claim payload")
		return
	}

	if err := h.Store.ReplaceClaimInquiries(
		r.Context(),
		chi.URLParam(r, "claimID"),
		validation.NormalizeInquiryIDs(req.InquiryIDs),
	); err != nil {
		h.writeStoreError(w, err, "failed to replace claim inquiries")
		return
	}

	httpx.WriteJSON(w, http.StatusNoContent, nil)
}

func (h ClaimHandler) Archive(w http.ResponseWriter, r *http.Request) {
	if err := h.Store.ArchiveClaim(r.Context(), chi.URLParam(r, "claimID")); err != nil {
		h.writeStoreError(w, err, "failed to archive claim")
		return
	}

	httpx.WriteJSON(w, http.StatusNoContent, nil)
}

func (h ClaimHandler) decodeAndValidate(w http.ResponseWriter, r *http.Request) (model.ClaimInput, []string, bool) {
	var req claimRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "validation_error", "invalid claim payload")
		return model.ClaimInput{}, nil, false
	}

	input, err := validation.NormalizeClaimInput(model.ClaimInput{
		Text:               req.Text,
		ClaimType:          req.ClaimType,
		Confidence:         req.Confidence,
		Status:             req.Status,
		OriginEngagementID: req.OriginEngagementID,
		Notes:              req.Notes,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "validation_error", err.Error())
		return model.ClaimInput{}, nil, false
	}

	return input, validation.NormalizeInquiryIDs(req.InquiryIDs), true
}

func (h ClaimHandler) writeStoreError(w http.ResponseWriter, err error, internalMessage string) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		httpx.WriteError(w, http.StatusNotFound, "not_found", "claim, inquiry, or engagement not found")
	default:
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", internalMessage)
	}
}
