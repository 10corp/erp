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

type PaymentEntryHandler struct {
	db *database.DB
}

func NewPaymentEntryHandler(db *database.DB) *PaymentEntryHandler {
	return &PaymentEntryHandler{db: db}
}

type CreatePaymentRequest struct {
	CompanyID              uuid.UUID `json:"company_id"`
	PaymentType            PaymentType `json:"payment_type"`
	PartyName              string    `json:"party_name"`
	PostingDate            string    `json:"posting_date"` // YYYY-MM-DD
	PaidToAccountID        uuid.UUID `json:"paid_to_account_id"`   // Bank / Cash
	PaidFromAccountID      uuid.UUID `json:"paid_from_account_id"` // Accounts Receivable
	Currency               string    `json:"currency"`
	ExchangeRate           string    `json:"exchange_rate"` // Payment date rate
	PaidAmount             string    `json:"paid_amount"`
	AllocatedAmount        string    `json:"allocated_amount"`
	ReferenceInvoiceID     *uuid.UUID `json:"reference_invoice_id,omitempty"`
	ReferenceInvoiceNumber string    `json:"reference_invoice_number,omitempty"`
	Remarks                string    `json:"remarks"`
}

// CreatePaymentEntry handles drafting a customer or supplier payment
func (h *PaymentEntryHandler) CreatePaymentEntry(w http.ResponseWriter, r *http.Request) {
	var req CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json payload"})
		return
	}

	postingDate, err := time.Parse("2006-01-02", req.PostingDate)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid posting_date format"})
		return
	}

	paidAmount, err := decimal.NewFromString(req.PaidAmount)
	if err != nil || paidAmount.IsZero() || paidAmount.IsNegative() {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "paid_amount must be greater than zero"})
		return
	}

	exchangeRate, err := decimal.NewFromString(req.ExchangeRate)
	if err != nil || exchangeRate.IsZero() {
		exchangeRate = decimal.NewFromInt(1)
	}

	allocatedAmount := paidAmount
	if req.AllocatedAmount != "" {
		if a, err := decimal.NewFromString(req.AllocatedAmount); err == nil && a.IsPositive() {
			allocatedAmount = a
		}
	}

	ctx := r.Context()
	var invoiceRate Money = exchangeRate

	// If linked to an existing Sales Invoice, fetch the original conversion rate
	if req.ReferenceInvoiceID != nil {
		err := h.db.Pool.QueryRow(ctx, `
			SELECT conversion_rate, invoice_number 
			FROM sales_invoices 
			WHERE id = $1;
		`, req.ReferenceInvoiceID).Scan(&invoiceRate, &req.ReferenceInvoiceNumber)
		if err != nil && err != pgx.ErrNoRows {
			respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch linked invoice"})
			return
		}
	}

	paymentID := uuid.New()
	paymentNumber := fmt.Sprintf("PAY-%s-%s", postingDate.Format("2006"), paymentID.String()[:8])

	payment := &PaymentEntry{
		BaseDocument: BaseDocument{
			ID:        paymentID,
			DocStatus: DocDraft,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
		CompanyID:              req.CompanyID,
		PaymentNumber:          paymentNumber,
		PaymentType:            req.PaymentType,
		PartyType:              "Customer",
		PartyName:              req.PartyName,
		PostingDate:            postingDate,
		PaidToAccountID:        req.PaidToAccountID,
		PaidFromAccountID:      req.PaidFromAccountID,
		Currency:               req.Currency,
		ExchangeRate:           exchangeRate,
		PaidAmount:             paidAmount,
		AllocatedAmount:        allocatedAmount,
		ReferenceInvoiceID:     req.ReferenceInvoiceID,
		ReferenceInvoiceNumber: req.ReferenceInvoiceNumber,
		ReferenceInvoiceRate:   invoiceRate,
		Remarks:                req.Remarks,
	}

	if err := payment.CalculateForex(); err != nil {
		respondJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		return
	}

	// Insert Draft into PostgreSQL
	insertQuery := `
		INSERT INTO payment_entries (
			id, company_id, payment_number, payment_type, party_type, party_name,
			posting_date, paid_to_account_id, paid_from_account_id,
			currency, exchange_rate, paid_amount, base_paid_amount, allocated_amount,
			reference_invoice_id, reference_invoice_number, reference_invoice_rate,
			forex_gain_loss, docstatus, remarks, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22);
	`
	_, err = h.db.Pool.Exec(ctx, insertQuery,
		payment.ID, payment.CompanyID, payment.PaymentNumber, payment.PaymentType, payment.PartyType, payment.PartyName,
		payment.PostingDate, payment.PaidToAccountID, payment.PaidFromAccountID,
		payment.Currency, payment.ExchangeRate, payment.PaidAmount, payment.BasePaidAmount, payment.AllocatedAmount,
		payment.ReferenceInvoiceID, payment.ReferenceInvoiceNumber, payment.ReferenceInvoiceRate,
		payment.ForexGainLoss, payment.DocStatus, payment.Remarks, payment.CreatedAt, payment.UpdatedAt,
	)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusCreated, payment)
}

// SubmitAndPostPayment submits the payment and posts atomic balanced double-entry rows to partitioned gl_entries
func (h *PaymentEntryHandler) SubmitAndPostPayment(w http.ResponseWriter, r *http.Request) {
	paymentIDStr := chi.URLParam(r, "id")
	paymentID, err := uuid.Parse(paymentIDStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid payment id"})
		return
	}

	ctx := r.Context()
	tx, err := h.db.Pool.Begin(ctx)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to start transaction"})
		return
	}
	defer tx.Rollback(ctx)

	var pe PaymentEntry
	query := `
		SELECT id, company_id, payment_number, payment_type, party_type, party_name,
		       posting_date, paid_to_account_id, paid_from_account_id,
		       currency, exchange_rate, paid_amount, base_paid_amount, allocated_amount,
		       reference_invoice_id, reference_invoice_number, reference_invoice_rate,
		       forex_gain_loss, docstatus
		FROM payment_entries
		WHERE id = $1
		FOR UPDATE;
	`
	err = tx.QueryRow(ctx, query, paymentID).Scan(
		&pe.ID, &pe.CompanyID, &pe.PaymentNumber, &pe.PaymentType, &pe.PartyType, &pe.PartyName,
		&pe.PostingDate, &pe.PaidToAccountID, &pe.PaidFromAccountID,
		&pe.Currency, &pe.ExchangeRate, &pe.PaidAmount, &pe.BasePaidAmount, &pe.AllocatedAmount,
		&pe.ReferenceInvoiceID, &pe.ReferenceInvoiceNumber, &pe.ReferenceInvoiceRate,
		&pe.ForexGainLoss, &pe.DocStatus,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			respondJSON(w, http.StatusNotFound, map[string]string{"error": "payment not found"})
			return
		}
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// Tryton Immutability Check: Can only submit if DocStatus == 0 (Draft)
	if pe.DocStatus != DocDraft {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "payment is already submitted or cancelled"})
		return
	}

	// Fetch Account Codes
	var bankCode, arCode string
	_ = tx.QueryRow(ctx, "SELECT code FROM accounts WHERE id = $1", pe.PaidToAccountID).Scan(&bankCode)
	_ = tx.QueryRow(ctx, "SELECT code FROM accounts WHERE id = $1", pe.PaidFromAccountID).Scan(&arCode)

	// Fetch or default Forex Accounts: 4200 (Gain) and 5200 (Loss)
	var forexGainID, forexLossID uuid.UUID
	err = tx.QueryRow(ctx, "SELECT id FROM accounts WHERE company_id = $1 AND code = '4200'", pe.CompanyID).Scan(&forexGainID)
	if err != nil {
		forexGainID = uuid.MustParse("00000000-0000-0000-0000-000000004200")
	}

	err = tx.QueryRow(ctx, "SELECT id FROM accounts WHERE company_id = $1 AND code = '5200'", pe.CompanyID).Scan(&forexLossID)
	if err != nil {
		forexLossID = uuid.MustParse("00000000-0000-0000-0000-000000005200")
	}

	// Generate Tryton-Strict Balanced GL Ledger Entries with Automated Forex Gain/Loss
	glEntries, err := pe.GenerateGLEntries(bankCode, arCode, "4200", "5200", forexGainID, forexLossID)
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

	// Update Payment State to Submitted (1)
	updatePaymentQuery := `
		UPDATE payment_entries
		SET docstatus = 1, posted_at = $1, updated_at = $1
		WHERE id = $2;
	`
	_, err = tx.Exec(ctx, updatePaymentQuery, now, pe.ID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	if err := tx.Commit(ctx); err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "commit failed"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":          "posted",
		"payment_id":      pe.ID,
		"gl_entries":      len(glEntries),
		"forex_gain_loss": pe.ForexGainLoss.String(),
		"message":         "Payment successfully posted to General Ledger with automatic Forex Gain/Loss reconciliation.",
	})
}
