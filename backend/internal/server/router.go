package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/enterprise-erp/core/internal/accounting"
	"github.com/enterprise-erp/core/internal/config"
	"github.com/enterprise-erp/core/internal/database"
	"github.com/enterprise-erp/core/internal/doctype"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func NewRouter(cfg *config.Config, db *database.DB) http.Handler {
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

	// Server-Sent Events (SSE) Endpoint for real-time client push
	r.Get("/api/v1/events", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		// Initial connection handshake
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

	// API v1 Routes
	r.Route("/api/v1", func(api chi.Router) {
		// Accounts (Chart of Accounts)
		api.Post("/accounts", accountingHandler.CreateAccount)
		api.Get("/accounts", accountingHandler.ListAccounts)

		// Journal Entries & General Ledger
		api.Post("/journal-entries", accountingHandler.PostJournalEntry)
		api.Get("/ledger", accountingHandler.GetLedgerEntries)

		// Sales Invoices (ERPNext DocType + Tryton Strict GL Posting)
		api.Post("/sales-invoices", salesInvoiceHandler.CreateSalesInvoice)
		api.Post("/sales-invoices/{id}/submit", salesInvoiceHandler.SubmitAndPostSalesInvoice)

		// Payment Entries & Automated Forex Gain/Loss Engine
		api.Post("/payment-entries", paymentHandler.CreatePaymentEntry)
		api.Post("/payment-entries/{id}/submit", paymentHandler.SubmitAndPostPayment)
	})

	return r
}
