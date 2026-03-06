package handlers

import (
	"errors"
	"net/http"

	"github.com/aponysus/lectio/internal/model"
	"github.com/aponysus/lectio/internal/server/httpx"
	"github.com/aponysus/lectio/internal/store"
	"github.com/aponysus/lectio/internal/validation"
	"github.com/go-chi/chi/v5"
)

type LanguageNoteHandler struct {
	Store *store.Store
}

type languageNoteRequest struct {
	EngagementID string `json:"engagement_id"`
	Term         string `json:"term"`
	Language     string `json:"language"`
	NoteType     string `json:"note_type"`
	Content      string `json:"content"`
}

func (h LanguageNoteHandler) ListEngagementLanguageNotes(w http.ResponseWriter, r *http.Request) {
	notes, err := h.Store.ListEngagementLanguageNotes(r.Context(), chi.URLParam(r, "engagementID"))
	if err != nil {
		h.writeStoreError(w, err, "failed to list engagement language notes")
		return
	}

	httpx.WriteData(w, http.StatusOK, notes)
}

func (h LanguageNoteHandler) Create(w http.ResponseWriter, r *http.Request) {
	input, ok := h.decodeAndValidate(w, r)
	if !ok {
		return
	}

	if input.EngagementID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "validation_error", "engagement_id is required")
		return
	}

	note, err := h.Store.CreateLanguageNote(r.Context(), input)
	if err != nil {
		h.writeStoreError(w, err, "failed to create language note")
		return
	}

	httpx.WriteData(w, http.StatusCreated, note)
}

func (h LanguageNoteHandler) Update(w http.ResponseWriter, r *http.Request) {
	input, ok := h.decodeAndValidate(w, r)
	if !ok {
		return
	}

	note, err := h.Store.UpdateLanguageNote(r.Context(), chi.URLParam(r, "languageNoteID"), input)
	if err != nil {
		h.writeStoreError(w, err, "failed to update language note")
		return
	}

	httpx.WriteData(w, http.StatusOK, note)
}

func (h LanguageNoteHandler) Archive(w http.ResponseWriter, r *http.Request) {
	if err := h.Store.ArchiveLanguageNote(r.Context(), chi.URLParam(r, "languageNoteID")); err != nil {
		h.writeStoreError(w, err, "failed to archive language note")
		return
	}

	httpx.WriteJSON(w, http.StatusNoContent, nil)
}

func (h LanguageNoteHandler) decodeAndValidate(w http.ResponseWriter, r *http.Request) (model.LanguageNoteInput, bool) {
	var req languageNoteRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "validation_error", "invalid language note payload")
		return model.LanguageNoteInput{}, false
	}

	input, err := validation.NormalizeLanguageNoteInput(model.LanguageNoteInput{
		EngagementID: req.EngagementID,
		Term:         req.Term,
		Language:     req.Language,
		NoteType:     req.NoteType,
		Content:      req.Content,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "validation_error", err.Error())
		return model.LanguageNoteInput{}, false
	}

	return input, true
}

func (h LanguageNoteHandler) writeStoreError(w http.ResponseWriter, err error, internalMessage string) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		httpx.WriteError(w, http.StatusNotFound, "not_found", "language note or engagement not found")
	default:
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", internalMessage)
	}
}
