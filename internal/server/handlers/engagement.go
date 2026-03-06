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
	SourceID           string                                `json:"source_id"`
	EngagedAt          string                                `json:"engaged_at"`
	PortionLabel       string                                `json:"portion_label"`
	Reflection         string                                `json:"reflection"`
	WhyItMatters       string                                `json:"why_it_matters"`
	SourceLanguage     string                                `json:"source_language"`
	ReflectionLanguage string                                `json:"reflection_language"`
	AccessMode         string                                `json:"access_mode"`
	RevisitPriority    *int                                  `json:"revisit_priority"`
	IsRereadOrRewatch  bool                                  `json:"is_reread_or_rewatch"`
	InquiryIDs         []string                              `json:"inquiry_ids"`
	InlineInquiries    []engagementInlineInquiryRequest      `json:"inline_inquiries"`
	Claims             []engagementInlineClaimRequest        `json:"claims"`
	LanguageNotes      []engagementInlineLanguageNoteRequest `json:"language_notes"`
}

type engagementInlineInquiryRequest struct {
	Title        string `json:"title"`
	Question     string `json:"question"`
	Status       string `json:"status"`
	WhyItMatters string `json:"why_it_matters"`
	CurrentView  string `json:"current_view"`
	OpenTensions string `json:"open_tensions"`
}

type engagementInlineClaimRequest struct {
	Text       string `json:"text"`
	ClaimType  string `json:"claim_type"`
	Confidence *int   `json:"confidence"`
	Status     string `json:"status"`
	Notes      string `json:"notes"`
}

type engagementInlineLanguageNoteRequest struct {
	Term     string `json:"term"`
	Language string `json:"language"`
	NoteType string `json:"note_type"`
	Content  string `json:"content"`
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

	hasLanguageNotes := false
	if raw := r.URL.Query().Get("has_language_notes"); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "validation_error", "has_language_notes must be a boolean")
			return
		}
		hasLanguageNotes = parsed
	}

	filters, err := validation.NormalizeEngagementFilters(model.EngagementFilters{
		Query:            r.URL.Query().Get("q"),
		SourceID:         r.URL.Query().Get("source_id"),
		AccessMode:       r.URL.Query().Get("access_mode"),
		HasLanguageNotes: hasLanguageNotes,
		Limit:            limit,
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
	input, ok := h.decodeAndValidateCapture(w, r)
	if !ok {
		return
	}

	engagement, err := h.Store.CreateEngagementCapture(r.Context(), input)
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

func (h EngagementHandler) decodeAndValidateCapture(w http.ResponseWriter, r *http.Request) (model.EngagementCaptureInput, bool) {
	var req engagementRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "validation_error", "invalid engagement payload")
		return model.EngagementCaptureInput{}, false
	}

	engagementInput, err := validation.NormalizeEngagementInput(model.EngagementInput{
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
		return model.EngagementCaptureInput{}, false
	}

	inquiryIDs := validation.NormalizeInquiryIDs(req.InquiryIDs)
	inlineInquiries := make([]model.InquiryInput, 0, len(req.InlineInquiries))
	for index, inquiry := range req.InlineInquiries {
		status := inquiry.Status
		if status == "" {
			status = string(model.InquiryStatusActive)
		}

		normalized, err := validation.NormalizeInquiryInput(model.InquiryInput{
			Title:        inquiry.Title,
			Question:     inquiry.Question,
			Status:       status,
			WhyItMatters: inquiry.WhyItMatters,
			CurrentView:  inquiry.CurrentView,
			OpenTensions: inquiry.OpenTensions,
		})
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "validation_error", prefixedFieldError(err, "inline_inquiries", index))
			return model.EngagementCaptureInput{}, false
		}
		inlineInquiries = append(inlineInquiries, normalized)
	}

	if len(req.Claims) > 3 {
		httpx.WriteError(w, http.StatusBadRequest, "validation_error", "claims must contain at most 3 items")
		return model.EngagementCaptureInput{}, false
	}

	claims := make([]model.ClaimInput, 0, len(req.Claims))
	for index, claim := range req.Claims {
		normalized, err := validation.NormalizeClaimInput(model.ClaimInput{
			Text:       claim.Text,
			ClaimType:  claim.ClaimType,
			Confidence: claim.Confidence,
			Status:     claim.Status,
			Notes:      claim.Notes,
		})
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "validation_error", prefixedFieldError(err, "claims", index))
			return model.EngagementCaptureInput{}, false
		}
		claims = append(claims, normalized)
	}

	languageNotes := make([]model.LanguageNoteInput, 0, len(req.LanguageNotes))
	for index, note := range req.LanguageNotes {
		normalized, err := validation.NormalizeLanguageNoteInput(model.LanguageNoteInput{
			Term:     note.Term,
			Language: note.Language,
			NoteType: note.NoteType,
			Content:  note.Content,
		})
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "validation_error", prefixedFieldError(err, "language_notes", index))
			return model.EngagementCaptureInput{}, false
		}
		languageNotes = append(languageNotes, normalized)
	}

	return model.EngagementCaptureInput{
		Engagement:      engagementInput,
		InquiryIDs:      inquiryIDs,
		InlineInquiries: inlineInquiries,
		Claims:          claims,
		LanguageNotes:   languageNotes,
	}, true
}

func prefixedFieldError(err error, collection string, index int) string {
	if fieldError, ok := err.(validation.FieldError); ok {
		return collection + "[" + strconv.Itoa(index) + "]." + fieldError.Field + " " + fieldError.Message
	}

	return err.Error()
}

func (h EngagementHandler) writeStoreError(w http.ResponseWriter, err error, internalMessage string) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		httpx.WriteError(w, http.StatusNotFound, "not_found", "engagement or source not found")
	default:
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", internalMessage)
	}
}
