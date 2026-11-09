package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/enterprise-erp/core/internal/accounting"
	"github.com/enterprise-erp/core/internal/ai"
	"github.com/enterprise-erp/core/internal/config"
	"github.com/enterprise-erp/core/internal/database"
	"github.com/enterprise-erp/core/internal/doctype"
	"github.com/enterprise-erp/core/internal/hr"
	"github.com/enterprise-erp/core/internal/inventory"
	"github.com/enterprise-erp/core/internal/module"
	"github.com/enterprise-erp/core/internal/party"
	"github.com/enterprise-erp/core/internal/queue"
	"github.com/enterprise-erp/core/internal/security"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func NewRouter(cfg *config.Config, db *database.DB, wp *queue.WorkerPool) http.Handler {
	r := chi.NewRouter()

	// Essential Middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS Configuration for React SPA
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSAllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Request-ID"},
		ExposedHeaders:   []string{"Link", "Content-Length", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health Check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","timestamp":"%s","version":"1.0.0"}`, time.Now().UTC().Format(time.RFC3339))
	})

	// Server-Sent Events (SSE) General Heartbeat Stream
	r.Get("/api/v1/events", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "event: connected\ndata: {\"message\":\"Connected to real-time ERP event stream\"}\n\n")
		flusher.Flush()

		ticker := time.NewTicker(25 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-r.Context().Done():
				return
			case t := <-ticker.C:
				fmt.Fprintf(w, "event: heartbeat\ndata: {\"time\":\"%s\"}\n\n", t.UTC().Format(time.RFC3339))
				flusher.Flush()
			}
		}
	})

	// Services & Handlers
	accountingService := accounting.NewService(db)
	accountingHandler := accounting.NewHandler(accountingService)
	salesInvoiceHandler := doctype.NewSalesInvoiceHandler(db)
	paymentHandler := doctype.NewPaymentEntryHandler(db)
	inventoryService := inventory.NewService(db)
	inventoryHandler := inventory.NewHandler(inventoryService)
	idpHandler := ai.NewIDPHandler(wp)
	userHandler := security.NewUserHandler(db)
	partyHandler := party.NewHandler(db)
	hrService := hr.NewService(db, accountingService)
	hrHandler := hr.NewHandler(hrService)

	// API v1 Routes
	r.Route("/api/v1", func(api chi.Router) {
		// Human Resources & Payroll Subsystem (ERPNext Business Fields + Tryton docstatus)
		api.Get("/hr/employees", hrHandler.ListEmployees)
		api.Get("/hr/departments", hrHandler.ListDepartments)
		api.Get("/hr/designations", hrHandler.ListDesignations)
		api.Get("/hr/salary-slips", hrHandler.ListSalarySlips)
		api.Post("/hr/salary-slips", hrHandler.CreateSalarySlip)
		api.Post("/hr/salary-slips/{id}/submit", hrHandler.SubmitSalarySlip)
		api.Post("/hr/salary-slips/{id}/cancel", hrHandler.CancelSalarySlip)
		api.Get("/hr/leaves", hrHandler.ListLeaves)
		api.Post("/hr/leaves", hrHandler.CreateLeave)
		api.Post("/hr/leaves/{id}/approve", hrHandler.ApproveLeave)
		api.Get("/hr/attendance", hrHandler.ListAttendance)
		api.Post("/hr/attendance/check-in", hrHandler.RecordCheckIn)
		api.Get("/hr/claims", hrHandler.ListClaims)
		api.Post("/hr/claims", hrHandler.CreateClaim)
		api.Post("/hr/claims/{id}/approve", hrHandler.ApproveClaim)
		api.Post("/hr/claims/{id}/cancel", hrHandler.CancelClaim)

		// Horilla Advanced: Loans & Advance Salary
		api.Get("/hr/loans", hrHandler.ListLoans)
		api.Post("/hr/loans", hrHandler.CreateLoan)
		api.Post("/hr/loans/{id}/disburse", hrHandler.DisburseLoan)
		api.Post("/hr/loans/{id}/cancel", hrHandler.CancelLoan)

		// Horilla Advanced: Recruitment ATS Pipeline
		api.Get("/hr/recruitment/jobs", hrHandler.ListJobOpenings)
		api.Post("/hr/recruitment/jobs", hrHandler.CreateJobOpening)
		api.Get("/hr/recruitment/candidates", hrHandler.ListCandidates)
		api.Post("/hr/recruitment/candidates", hrHandler.CreateCandidate)
		api.Patch("/hr/recruitment/candidates/{id}/stage", hrHandler.UpdateCandidateStage)
		api.Post("/hr/recruitment/candidates/{id}/promote", hrHandler.PromoteCandidate)

		// Horilla Advanced: Company Asset Custody & Allocation
		api.Get("/hr/assets", hrHandler.ListAssets)
		api.Post("/hr/assets", hrHandler.CreateAsset)
		api.Get("/hr/assets/allocations", hrHandler.ListAllocations)
		api.Post("/hr/assets/allocate", hrHandler.AllocateAsset)
		api.Post("/hr/assets/allocations/{id}/return", hrHandler.ReturnAsset)
		// Core User & Identity Management (Tryton res.user & JWT Context)
		api.Post("/auth/login", userHandler.Login)
		api.Post("/users/switch-branch", userHandler.SwitchBranch)
		api.Get("/users", userHandler.ListUsers)
		api.Post("/users/{id}/toggle-status", userHandler.ToggleUserStatus)

		// Parties & Universal Legal Entities (Tryton party.party)
		api.Get("/parties", partyHandler.ListParties)
		api.Post("/parties", partyHandler.CreateParty)
		api.Get("/parties/{id}", partyHandler.GetPartyByID)
		api.Get("/companies/hierarchy", partyHandler.GetCompanyHierarchy)

		// Dynamic Module Manifest (ERPNext-style modularity & fault-tolerance)
		api.Get("/modules", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			modules := module.GetRegistry().ListAll()
			_ = json.NewEncoder(w).Encode(modules)
		})

		// Accounts (Chart of Accounts & Trial Balance)
		api.Post("/accounts", accountingHandler.CreateAccount)
		api.Get("/accounts", accountingHandler.ListAccounts)
		api.Get("/accounting/trial-balance", accountingHandler.GetTrialBalance)
		api.Get("/accounting/party-ledger", accountingHandler.GetPartyLedger)
		api.Post("/accounting/inter-branch-transfer", accountingHandler.CreateInterBranchTransfer)

		// Journal Entries & General Ledger
		api.Post("/journal-entries", accountingHandler.PostJournalEntry)
		api.Get("/ledger", accountingHandler.GetLedgerEntries)

		// Sales Invoices (ERPNext DocType + Tryton Strict GL Posting)
		api.Post("/sales-invoices", salesInvoiceHandler.CreateSalesInvoice)
		api.Post("/sales-invoices/{id}/submit", salesInvoiceHandler.SubmitAndPostSalesInvoice)

		// Payment Entries & Automated Forex Gain/Loss Engine
		api.Post("/payment-entries", paymentHandler.CreatePaymentEntry)
		api.Post("/payment-entries/{id}/submit", paymentHandler.SubmitAndPostPayment)

		// Double-Entry Inventory & Perpetual COGS Auto-Posting
		api.Post("/inventory/receive", inventoryHandler.ReceiveStock)
		api.Post("/inventory/delivery", inventoryHandler.DeliverStock)

		// Intelligent Document Processing (Gemini 1.5 Flash + SSE)
		api.Post("/idp/upload", idpHandler.UploadDocument)
		api.Get("/idp/stream", idpHandler.StreamEvents)

		// Claude 3.5 Sonnet Natural Language Action Brain
		actionHandler := ai.NewActionHandler()
		api.Post("/ai/action", actionHandler.ExecuteActionQuery)
	})

	return r
}
