package me

import (
	"encoding/json"
	"net/http"
)

type MeHandler struct {
}

func NewMeHandler() *MeHandler {
	return &MeHandler{}
}

func (h *MeHandler) Balance(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, map[string]any{
		"balance": 50,
	})
	return
}

func (h *MeHandler) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, `{"error":"failed to encode json"}`, http.StatusInternalServerError)
	}
}
