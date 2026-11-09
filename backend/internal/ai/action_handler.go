package ai

import (
	"encoding/json"
	"net/http"
)

type ActionHandler struct {
	brain *ClaudeActionBrain
}

func NewActionHandler() *ActionHandler {
	return &ActionHandler{
		brain: NewClaudeActionBrain(),
	}
}

type AIActionRequest struct {
	Query   string                 `json:"query"`
	Context map[string]interface{} `json:"context,omitempty"`
}

func (h *ActionHandler) ExecuteActionQuery(w http.ResponseWriter, r *http.Request) {
	var req AIActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json payload"})
		return
	}

	action, err := h.brain.ParseUserIntent(r.Context(), req.Query, req.Context)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, action)
}
