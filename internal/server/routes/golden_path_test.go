package routes

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/aponysus/lectio/internal/model"
)

func TestGoldenPathCaptureFlow(t *testing.T) {
	t.Parallel()

	client, _ := newArchiveTestServer(t)
	csrfToken := loginArchiveTestClient(t, client)

	source := mutateAPIData[model.Source](t, client, http.MethodPost, "/api/sources", csrfToken, `{
	  "title": "Golden Path Source",
	  "medium": "BOOK",
	  "creator": "Test Author",
	  "year": 2024,
	  "original_language": "en",
	  "culture_or_context": "Regression harness",
	  "notes": ""
	}`, http.StatusCreated)

	inquiry := mutateAPIData[model.Inquiry](t, client, http.MethodPost, "/api/inquiries", csrfToken, `{
	  "title": "What survives the golden path?",
	  "question": "Does the MVP support the full loop from source to synthesis?",
	  "status": "ACTIVE",
	  "why_it_matters": "This is the end-to-end capture test.",
	  "current_view": "",
	  "open_tensions": ""
	}`, http.StatusCreated)

	engagement := mutateAPIData[model.Engagement](t, client, http.MethodPost, "/api/engagements", csrfToken, `{
	  "source_id": "`+source.ID+`",
	  "engaged_at": "2026-03-06T18:00:00Z",
	  "portion_label": "Chapter 2",
	  "reflection": "The MVP should create the engagement, claims, note, and inquiry links in one transaction.",
	  "why_it_matters": "It proves the main loop is coherent.",
	  "source_language": "en",
	  "reflection_language": "en",
	  "access_mode": "ORIGINAL",
	  "revisit_priority": 4,
	  "is_reread_or_rewatch": false,
	  "inquiry_ids": ["`+inquiry.ID+`"],
	  "claims": [
	    {
	      "text": "The capture flow is the system's primary command boundary.",
	      "claim_type": "INTERPRETATION",
	      "confidence": 4,
	      "status": "ACTIVE",
	      "notes": ""
	    },
	    {
	      "text": "Nested records should not require manual recovery after a partial failure.",
	      "claim_type": "OBSERVATION",
	      "confidence": 5,
	      "status": "TENTATIVE",
	      "notes": "Golden path claim"
	    }
	  ],
	  "language_notes": [
	    {
	      "term": "praxis",
	      "language": "grc",
	      "note_type": "TRANSLATION",
	      "content": "Language note created inline with the engagement."
	    }
	  ]
	}`, http.StatusCreated)

	linkedInquiries := getAPIData[[]model.InquirySummary](t, client, "/api/engagements/"+engagement.ID+"/inquiries")
	if len(linkedInquiries) != 1 || linkedInquiries[0].ID != inquiry.ID {
		t.Fatalf("expected engagement to be linked to inquiry %s, got %+v", inquiry.ID, linkedInquiries)
	}

	inquiryDetail := getAPIData[model.Inquiry](t, client, "/api/inquiries/"+inquiry.ID)
	if inquiryDetail.EngagementCount != 1 {
		t.Fatalf("expected inquiry engagement count 1, got %d", inquiryDetail.EngagementCount)
	}
	if inquiryDetail.ClaimCount != 2 {
		t.Fatalf("expected inquiry claim count 2, got %d", inquiryDetail.ClaimCount)
	}

	inquiryEngagements := getAPIData[[]model.Engagement](t, client, "/api/inquiries/"+inquiry.ID+"/engagements")
	if len(inquiryEngagements) != 1 || inquiryEngagements[0].ID != engagement.ID {
		t.Fatalf("expected inquiry engagements to contain %s, got %+v", engagement.ID, inquiryEngagements)
	}

	inquiryClaims := getAPIData[[]model.Claim](t, client, "/api/inquiries/"+inquiry.ID+"/claims")
	if len(inquiryClaims) != 2 {
		t.Fatalf("expected 2 inquiry claims, got %d", len(inquiryClaims))
	}

	engagementNotes := getAPIData[[]model.LanguageNote](t, client, "/api/engagements/"+engagement.ID+"/language-notes")
	if len(engagementNotes) != 1 {
		t.Fatalf("expected 1 engagement language note, got %d", len(engagementNotes))
	}

	synthesis := mutateAPIData[model.Synthesis](t, client, http.MethodPost, "/api/syntheses", csrfToken, `{
	  "title": "Golden path synthesis",
	  "body": "The inquiry now has enough material to support a first compression pass.",
	  "type": "CHECKPOINT",
	  "inquiry_id": "`+inquiry.ID+`",
	  "notes": ""
	}`, http.StatusCreated)

	inquirySyntheses := getAPIData[[]model.Synthesis](t, client, "/api/inquiries/"+inquiry.ID+"/syntheses")
	if len(inquirySyntheses) != 1 || inquirySyntheses[0].ID != synthesis.ID {
		t.Fatalf("expected inquiry syntheses to contain %s, got %+v", synthesis.ID, inquirySyntheses)
	}

	dashboardStatus := getAPIData[[]model.Inquiry](t, client, "/api/inquiries/eligible-for-synthesis")
	if len(dashboardStatus) != 0 {
		t.Fatalf("expected no synthesis-eligible inquiries after creating synthesis, got %+v", dashboardStatus)
	}

	_ = getAPIData[[]model.RediscoveryItem](t, client, "/api/rediscovery/items?limit=6")
}

func mutateAPIData[T any](t *testing.T, client *archiveTestClient, method, path, csrfToken, body string, wantStatus int) T {
	t.Helper()

	resp := client.Do(method, path, csrfToken, bytes.NewBufferString(body))
	defer resp.Body.Close()

	if resp.StatusCode != wantStatus {
		payload, _ := io.ReadAll(resp.Body)
		t.Fatalf("%s %s returned status %d, want %d, body=%s", method, path, resp.StatusCode, wantStatus, strings.TrimSpace(string(payload)))
	}

	var envelope apiEnvelope[T]
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		t.Fatalf("decode %s %s response: %v", method, path, err)
	}

	return envelope.Data
}
