package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type ErrorEnvelope struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload == nil {
		return
	}
	_ = json.NewEncoder(w).Encode(payload)
}

func WriteData(w http.ResponseWriter, status int, data any) {
	WriteJSON(w, status, map[string]any{"data": data})
}

func WriteError(w http.ResponseWriter, status int, code, message string) {
	envelope := ErrorEnvelope{}
	envelope.Error.Code = code
	envelope.Error.Message = message
	WriteJSON(w, status, envelope)
}

func DecodeJSON(r *http.Request, dst any) error {
	if r.Body == nil {
		return errors.New("request body required")
	}

	defer r.Body.Close()

	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()

	return decoder.Decode(dst)
}
