package me

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/boliev/graphai/internal/domain/user"
	"github.com/boliev/graphai/internal/pkg/vkHelper"
)

type MeHandler struct {
	userService *user.Service
	vkSecureKey string
}

func NewMeHandler(userService *user.Service, vkSecureKey string) *MeHandler {
	return &MeHandler{
		userService: userService,
		vkSecureKey: vkSecureKey,
	}
}

func (h *MeHandler) Balance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	VKUserID, err := vkHelper.GetVKUserID(r, h.vkSecureKey)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("error while requesting balance %v", err.Error())
		return
	}
	u, err := h.userService.FindByVKID(ctx, VKUserID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("error while fetching user for balance %v", err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]any{
		"balance": u.Credits,
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
