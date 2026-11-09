package hr

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ListEmployees(w http.ResponseWriter, r *http.Request) {
	companyID := parseCompanyID(r)
	employees, err := h.service.ListEmployees(r.Context(), companyID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, employees)
}

func (h *Handler) ListDepartments(w http.ResponseWriter, r *http.Request) {
	companyID := parseCompanyID(r)
	depts, err := h.service.ListDepartments(r.Context(), companyID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, depts)
}

func (h *Handler) ListDesignations(w http.ResponseWriter, r *http.Request) {
	companyID := parseCompanyID(r)
	desigs, err := h.service.ListDesignations(r.Context(), companyID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, desigs)
}

func (h *Handler) ListSalarySlips(w http.ResponseWriter, r *http.Request) {
	companyID := parseCompanyID(r)
	month, _ := strconv.Atoi(r.URL.Query().Get("month"))
	year, _ := strconv.Atoi(r.URL.Query().Get("year"))

	slips, err := h.service.ListSalarySlips(r.Context(), companyID, month, year)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, slips)
}

func (h *Handler) CreateSalarySlip(w http.ResponseWriter, r *http.Request) {
	var req CreateSalarySlipRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request payload: "+err.Error())
		return
	}

	slip, err := h.service.CreateSalarySlip(r.Context(), req)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, slip)
}

func (h *Handler) SubmitSalarySlip(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	slipID, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid slip id format")
		return
	}

	slip, err := h.service.SubmitSalarySlip(r.Context(), slipID)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, slip)
}

func (h *Handler) CancelSalarySlip(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	slipID, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid slip id format")
		return
	}

	slip, err := h.service.CancelSalarySlip(r.Context(), slipID)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, slip)
}

func (h *Handler) ListLeaves(w http.ResponseWriter, r *http.Request) {
	companyID := parseCompanyID(r)
	leaves, err := h.service.ListLeaves(r.Context(), companyID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, leaves)
}

func (h *Handler) CreateLeave(w http.ResponseWriter, r *http.Request) {
	var req CreateLeaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request payload: "+err.Error())
		return
	}

	leave, err := h.service.CreateLeave(r.Context(), req)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, leave)
}

func (h *Handler) ApproveLeave(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	leaveID, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid leave id format")
		return
	}

	leave, err := h.service.ApproveLeave(r.Context(), leaveID, "System Administrator (CFO)")
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, leave)
}

func (h *Handler) ListAttendance(w http.ResponseWriter, r *http.Request) {
	companyID := parseCompanyID(r)
	records, err := h.service.ListAttendance(r.Context(), companyID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, records)
}

func (h *Handler) RecordCheckIn(w http.ResponseWriter, r *http.Request) {
	var req CheckInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid payload: "+err.Error())
		return
	}

	rec, err := h.service.RecordCheckIn(r.Context(), req)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, rec)
}

func (h *Handler) ListClaims(w http.ResponseWriter, r *http.Request) {
	companyID := parseCompanyID(r)
	claims, err := h.service.ListClaims(r.Context(), companyID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, claims)
}

func (h *Handler) CreateClaim(w http.ResponseWriter, r *http.Request) {
	var req CreateClaimRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid payload: "+err.Error())
		return
	}

	claim, err := h.service.CreateClaim(r.Context(), req)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, claim)
}

func (h *Handler) ApproveClaim(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	claimID, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid claim id format")
		return
	}

	claim, err := h.service.ApproveClaim(r.Context(), claimID)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, claim)
}

func (h *Handler) CancelClaim(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	claimID, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid claim id format")
		return
	}

	claim, err := h.service.CancelClaim(r.Context(), claimID)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, claim)
}

// ============================================================================
// Horilla Advanced Pillar 1: Loans & Advances
// ============================================================================

func (h *Handler) ListLoans(w http.ResponseWriter, r *http.Request) {
	companyID := parseCompanyID(r)
	loans, err := h.service.ListLoans(r.Context(), companyID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, loans)
}

func (h *Handler) CreateLoan(w http.ResponseWriter, r *http.Request) {
	var req CreateLoanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	loan, err := h.service.CreateLoan(r.Context(), req)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, loan)
}

func (h *Handler) DisburseLoan(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	loanID, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid loan id format")
		return
	}

	loan, err := h.service.DisburseLoan(r.Context(), loanID)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, loan)
}

func (h *Handler) CancelLoan(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	loanID, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid loan id format")
		return
	}

	loan, err := h.service.CancelLoan(r.Context(), loanID)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, loan)
}

// ============================================================================
// Horilla Advanced Pillar 2: Recruitment & ATS Pipeline
// ============================================================================

func (h *Handler) ListJobOpenings(w http.ResponseWriter, r *http.Request) {
	companyID := parseCompanyID(r)
	openings, err := h.service.ListJobOpenings(r.Context(), companyID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, openings)
}

func (h *Handler) CreateJobOpening(w http.ResponseWriter, r *http.Request) {
	var req CreateJobOpeningRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	opening, err := h.service.CreateJobOpening(r.Context(), req)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, opening)
}

func (h *Handler) ListCandidates(w http.ResponseWriter, r *http.Request) {
	companyID := parseCompanyID(r)
	var openingID *uuid.UUID
	if opStr := r.URL.Query().Get("job_opening_id"); opStr != "" {
		if id, err := uuid.Parse(opStr); err == nil {
			openingID = &id
		}
	}

	candidates, err := h.service.ListCandidates(r.Context(), companyID, openingID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, candidates)
}

func (h *Handler) CreateCandidate(w http.ResponseWriter, r *http.Request) {
	var req CreateCandidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	candidate, err := h.service.CreateCandidate(r.Context(), req)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, candidate)
}

func (h *Handler) UpdateCandidateStage(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	candidateID, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid candidate id format")
		return
	}

	var req UpdateCandidateStageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	candidate, err := h.service.UpdateCandidateStage(r.Context(), candidateID, req)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, candidate)
}

func (h *Handler) PromoteCandidate(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	candidateID, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid candidate id format")
		return
	}

	emp, err := h.service.PromoteCandidateToEmployee(r.Context(), candidateID)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, emp)
}

// ============================================================================
// Horilla Advanced Pillar 3: Company Asset Custody & Allocation
// ============================================================================

func (h *Handler) ListAssets(w http.ResponseWriter, r *http.Request) {
	companyID := parseCompanyID(r)
	assets, err := h.service.ListAssets(r.Context(), companyID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, assets)
}

func (h *Handler) CreateAsset(w http.ResponseWriter, r *http.Request) {
	var req CreateAssetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	asset, err := h.service.CreateAsset(r.Context(), req)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, asset)
}

func (h *Handler) ListAllocations(w http.ResponseWriter, r *http.Request) {
	var empID *uuid.UUID
	if empStr := r.URL.Query().Get("employee_id"); empStr != "" {
		if id, err := uuid.Parse(empStr); err == nil {
			empID = &id
		}
	}

	allocs, err := h.service.ListAllocations(r.Context(), empID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, allocs)
}

func (h *Handler) AllocateAsset(w http.ResponseWriter, r *http.Request) {
	var req AllocateAssetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	alloc, err := h.service.AllocateAsset(r.Context(), req)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, alloc)
}

func (h *Handler) ReturnAsset(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	allocID, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid allocation id format")
		return
	}

	var req ReturnAssetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	alloc, err := h.service.ReturnAsset(r.Context(), allocID, req)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, alloc)
}

func parseCompanyID(r *http.Request) uuid.UUID {
	cidStr := r.URL.Query().Get("company_id")
	if cidStr != "" {
		if id, err := uuid.Parse(cidStr); err == nil {
			return id
		}
	}
	return uuid.MustParse("00000000-0000-0000-0000-000000000001")
}

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
