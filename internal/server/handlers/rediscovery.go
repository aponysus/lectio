package handlers

import (
	"errors"
	"net/http"

	"github.com/aponysus/lectio/internal/server/httpx"
	"github.com/aponysus/lectio/internal/store"
	"github.com/go-chi/chi/v5"
)

type RediscoveryHandler struct {
	Store *store.Store
}

func (h RediscoveryHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, ok := parseListLimit(w, r, 6)
	if !ok {
		return
	}

	items, err := h.Store.ListRediscoveryItems(r.Context(), limit)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to list rediscovery items")
		return
	}

	httpx.WriteData(w, http.StatusOK, items)
}

func (h RediscoveryHandler) Dismiss(w http.ResponseWriter, r *http.Request) {
	if err := h.Store.DismissRediscoveryItem(r.Context(), chi.URLParam(r, "itemID")); err != nil {
		h.writeStoreError(w, err, "failed to dismiss rediscovery item")
		return
	}

	httpx.WriteJSON(w, http.StatusNoContent, nil)
}

func (h RediscoveryHandler) MarkActedOn(w http.ResponseWriter, r *http.Request) {
	if err := h.Store.MarkRediscoveryItemActedOn(r.Context(), chi.URLParam(r, "itemID")); err != nil {
		h.writeStoreError(w, err, "failed to mark rediscovery item acted on")
		return
	}

	httpx.WriteJSON(w, http.StatusNoContent, nil)
}

func (h RediscoveryHandler) writeStoreError(w http.ResponseWriter, err error, internalMessage string) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		httpx.WriteError(w, http.StatusNotFound, "not_found", "rediscovery item not found")
	default:
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", internalMessage)
	}
}
