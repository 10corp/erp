package doctype

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/enterprise-erp/core/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

type SalesInvoiceHandler struct {
	db *database.DB
}

func NewSalesInvoiceHandler(db *database.DB) *SalesInvoiceHandler {
	return &SalesInvoiceHandler{db: db}
}

type CreateInvoiceRequest struct {
	CompanyID        uuid.UUID `json:"company_id"`
	CustomerName     string    `json:"customer_name"`
	PostingDate      string    `json:"posting_date"` // YYYY-MM-DD
	DueDate          string    `json:"due_date"`
	Currency         string    `json:"currency"`
	ConversionRate   string    `json:"conversion_rate"`
	DebitToAccountID uuid.UUID `json:"debit_to_account_id"`
	IncomeAccountID  uuid.UUID `json:"income_account_id"`
	TaxAccountID     *uuid.UUID `json:"tax_account_id,omitempty"`
	Remarks          string    `json:"remarks"`
	Items            []struct {
		ItemCode    string `json:"item_code"`
		ItemName    string `json:"item_name"`
		Description string `json:"description"`
		Qty         string `json:"qty"`
		Rate        string `json:"rate"`
		TaxRate     string `json:"tax_rate"`
	} `json:"items"`
}

// CreateSalesInvoice handles creating a Draft Sales Invoice
func (h *SalesInvoiceHandler) CreateSalesInvoice(w http.ResponseWriter, r *http.Request) {
	var req CreateInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json payload"})
		return
	}

	postingDate, err := time.Parse("2006-01-02", req.PostingDate)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid posting_date format"})
		return
	}

	dueDate, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		dueDate = postingDate.AddDate(0, 1, 0)
	}

	conversionRate, err := decimal.NewFromString(req.ConversionRate)
	if err != nil || conversionRate.IsZero() {
		conversionRate = decimal.NewFromInt(1)
	}

	invoiceID := uuid.New()
	invoiceNumber := fmt.Sprintf("SINV-%s-%s", postingDate.Format("2006"), invoiceID.String()[:8])

	invoice := &SalesInvoice{
		BaseDocument: BaseDocument{
			ID:        invoiceID,
			DocStatus: DocDraft,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
		CompanyID:        req.CompanyID,
		InvoiceNumber:    invoiceNumber,
		CustomerName:     req.CustomerName,
		PostingDate:      postingDate,
		DueDate:          dueDate,
		Currency:         req.Currency,
		ConversionRate:   conversionRate,
		DebitToAccountID: req.DebitToAccountID,
		IncomeAccountID:  req.IncomeAccountID,
		TaxAccountID:     req.TaxAccountID,
		Remarks:          req.Remarks,
	}

	for _, itemInput := range req.Items {
		qty, _ := decimal.NewFromString(itemInput.Qty)
		rate, _ := decimal.NewFromString(itemInput.Rate)
		taxRate, _ := decimal.NewFromString(itemInput.TaxRate)

		invoice.Items = append(invoice.Items, SalesInvoiceItem{
			ID:          uuid.New(),
			ParentID:    invoiceID,
			ItemCode:    itemInput.ItemCode,
			ItemName:    itemInput.ItemName,
			Description: itemInput.Description,
			Qty:         qty,
			Rate:        rate,
			TaxRate:     taxRate,
		})
	}

	if err := invoice.CalculateTotals(); err != nil {
		respondJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		return
	}

	// Persist Draft Invoice into PostgreSQL
	tx, err := h.db.Pool.Begin(r.Context())
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to begin tx"})
		return
	}
	defer tx.Rollback(r.Context())

	insertInvQuery := `
		INSERT INTO sales_invoices (
			id, company_id, invoice_number, customer_name, posting_date, due_date,
			currency, conversion_rate, net_total, tax_total, grand_total,
			base_net_total, base_tax_total, base_grand_total,
			debit_to_account_id, income_account_id, tax_account_id,
			docstatus, remarks, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21);
	`
	_, err = tx.Exec(r.Context(), insertInvQuery,
		invoice.ID, invoice.CompanyID, invoice.InvoiceNumber, invoice.CustomerName,
		invoice.PostingDate, invoice.DueDate, invoice.Currency, invoice.ConversionRate,
		invoice.NetTotal, invoice.TaxTotal, invoice.GrandTotal,
		invoice.BaseNetTotal, invoice.BaseTaxTotal, invoice.BaseGrandTotal,
		invoice.DebitToAccountID, invoice.IncomeAccountID, invoice.TaxAccountID,
		invoice.DocStatus, invoice.Remarks, invoice.CreatedAt, invoice.UpdatedAt,
	)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	insertItemQuery := `
		INSERT INTO sales_invoice_items (
			id, parent_id, item_code, item_name, description, qty, rate, amount,
			tax_rate, tax_amount, base_rate, base_amount, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13);
	`
	for _, it := range invoice.Items {
		_, err = tx.Exec(r.Context(), insertItemQuery,
			it.ID, it.ParentID, it.ItemCode, it.ItemName, it.Description,
			it.Qty, it.Rate, it.Amount, it.TaxRate, it.TaxAmount, it.BaseRate, it.BaseAmount, invoice.CreatedAt,
		)
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	}

	if err := tx.Commit(r.Context()); err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to commit invoice"})
		return
	}

	respondJSON(w, http.StatusCreated, invoice)
}

// SubmitAndPostSalesInvoice posts invoice lines to partitioned gl_entries in one atomic ACID transaction
func (h *SalesInvoiceHandler) SubmitAndPostSalesInvoice(w http.ResponseWriter, r *http.Request) {
	invoiceIDStr := chi.URLParam(r, "id")
	invoiceID, err := uuid.Parse(invoiceIDStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid invoice id"})
		return
	}

	ctx := r.Context()
	tx, err := h.db.Pool.Begin(ctx)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to start transaction"})
		return
	}
	defer tx.Rollback(ctx)

	// Fetch invoice
	var inv SalesInvoice
	query := `
		SELECT id, company_id, invoice_number, customer_name, posting_date, due_date,
		       currency, conversion_rate, net_total, tax_total, grand_total,
		       base_net_total, base_tax_total, base_grand_total,
		       debit_to_account_id, income_account_id, tax_account_id, docstatus
		FROM sales_invoices
		WHERE id = $1
		FOR UPDATE;
	`
	err = tx.QueryRow(ctx, query, invoiceID).Scan(
		&inv.ID, &inv.CompanyID, &inv.InvoiceNumber, &inv.CustomerName, &inv.PostingDate, &inv.DueDate,
		&inv.Currency, &inv.ConversionRate, &inv.NetTotal, &inv.TaxTotal, &inv.GrandTotal,
		&inv.BaseNetTotal, &inv.BaseTaxTotal, &inv.BaseGrandTotal,
		&inv.DebitToAccountID, &inv.IncomeAccountID, &inv.TaxAccountID, &inv.DocStatus,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			respondJSON(w, http.StatusNotFound, map[string]string{"error": "sales invoice not found"})
			return
		}
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// Tryton Immutability Check: Can only submit if DocStatus == 0 (Draft)
	if inv.DocStatus != DocDraft {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invoice is already submitted or cancelled"})
		return
	}

	// Fetch Account Codes
	var arCode, incomeCode, taxCode string
	_ = tx.QueryRow(ctx, "SELECT code FROM accounts WHERE id = $1", inv.DebitToAccountID).Scan(&arCode)
	_ = tx.QueryRow(ctx, "SELECT code FROM accounts WHERE id = $1", inv.IncomeAccountID).Scan(&incomeCode)
	if inv.TaxAccountID != nil {
		_ = tx.QueryRow(ctx, "SELECT code FROM accounts WHERE id = $1", *inv.TaxAccountID).Scan(&taxCode)
	}

	// Generate Tryton-Strict Balanced GL Ledger Entries with Account 5090 Round Off Guard
	glEntries, err := inv.GenerateGLEntries(arCode, incomeCode, taxCode, nil, "5090")
	if err != nil {
		respondJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		return
	}

	// Insert into Partitioned gl_entries table
	insertGLQuery := `
		INSERT INTO gl_entries (
			id, company_id, posting_date, voucher_type, voucher_id, voucher_no,
			account_id, account_code, party_type, currency,
			debit, credit, exchange_rate, debit_base, credit_base, remarks, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17);
	`
	now := time.Now().UTC()
	for _, gl := range glEntries {
		_, err = tx.Exec(ctx, insertGLQuery,
			gl.ID, gl.CompanyID, gl.PostingDate, gl.VoucherType, gl.VoucherID, gl.VoucherNo,
			gl.AccountID, gl.AccountCode, gl.PartyType, gl.Currency,
			gl.Debit, gl.Credit, gl.ExchangeRate, gl.DebitBase, gl.CreditBase, gl.Remarks, now,
		)
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	}

	// Update Invoice State to Submitted (1)
	updateInvQuery := `
		UPDATE sales_invoices
		SET docstatus = 1, posted_at = $1, updated_at = $1
		WHERE id = $2;
	`
	_, err = tx.Exec(ctx, updateInvQuery, now, inv.ID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	if err := tx.Commit(ctx); err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "commit failed"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":     "posted",
		"invoice_id": inv.ID,
		"gl_entries": len(glEntries),
		"message":    "Invoice successfully posted to General Ledger partition with full double-entry balance.",
	})
}

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
