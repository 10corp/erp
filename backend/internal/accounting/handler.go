package accounting

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var req CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	account, err := h.service.CreateAccount(r.Context(), req)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusCreated, account)
}

func (h *Handler) ListAccounts(w http.ResponseWriter, r *http.Request) {
	companyIDStr := r.URL.Query().Get("company_id")
	if companyIDStr == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "company_id query param is required"})
		return
	}

	companyID, err := uuid.Parse(companyIDStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid company_id format"})
		return
	}

	accounts, err := h.service.ListAccounts(r.Context(), companyID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"data":  accounts,
		"total": len(accounts),
	})
}

func (h *Handler) PostJournalEntry(w http.ResponseWriter, r *http.Request) {
	var req CreateJournalEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json payload"})
		return
	}

	entry, err := h.service.PostJournalEntry(r.Context(), req)
	if err != nil {
		respondJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusCreated, entry)
}

func (h *Handler) GetLedgerEntries(w http.ResponseWriter, r *http.Request) {
	companyIDStr := r.URL.Query().Get("company_id")
	if companyIDStr == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "company_id query param is required"})
		return
	}

	companyID, err := uuid.Parse(companyIDStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid company_id"})
		return
	}

	limit := 50
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0
	if oStr := r.URL.Query().Get("offset"); oStr != "" {
		if o, err := strconv.Atoi(oStr); err == nil && o >= 0 {
			offset = o
		}
	}

	entries, err := h.service.GetLedgerEntries(r.Context(), companyID, limit, offset)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"data":   entries,
		"limit":  limit,
		"offset": offset,
		"count":  len(entries),
	})
}

// GetTrialBalance handles GET /api/v1/accounting/trial-balance?company_id=...&consolidated=true
func (h *Handler) GetTrialBalance(w http.ResponseWriter, r *http.Request) {
	companyIDStr := r.URL.Query().Get("company_id")
	var companyID uuid.UUID
	var err error

	if companyIDStr != "" {
		companyID, err = uuid.Parse(companyIDStr)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid company_id"})
			return
		}
	} else {
		// Default to Global HQ
		companyID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	}

	consolidated := r.URL.Query().Get("consolidated") == "true"

	report, err := h.service.GetTrialBalance(r.Context(), companyID, consolidated)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, report)
}

// CreateInterBranchTransfer handles POST /api/v1/accounting/inter-branch-transfer
func (h *Handler) CreateInterBranchTransfer(w http.ResponseWriter, r *http.Request) {
	var req InterBranchTransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json payload: " + err.Error()})
		return
	}

	result, err := h.service.CreateInterBranchTransfer(r.Context(), req)
	if err != nil {
		respondJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusCreated, result)
}

// GetPartyLedger handles GET /api/v1/accounting/party-ledger?party_id=...&company_id=...&consolidated=true
func (h *Handler) GetPartyLedger(w http.ResponseWriter, r *http.Request) {
	partyIDStr := r.URL.Query().Get("party_id")
	if partyIDStr == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "party_id query param is required"})
		return
	}

	partyID, err := uuid.Parse(partyIDStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid party_id format"})
		return
	}

	companyIDStr := r.URL.Query().Get("company_id")
	var companyID uuid.UUID
	if companyIDStr != "" {
		companyID, err = uuid.Parse(companyIDStr)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid company_id"})
			return
		}
	} else {
		companyID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	}

	consolidated := r.URL.Query().Get("consolidated") == "true"

	report, err := h.service.GetPartyLedger(r.Context(), companyID, partyID, consolidated)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, report)
}

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
