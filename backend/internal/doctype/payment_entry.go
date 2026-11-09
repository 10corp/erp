package doctype

import (
	"errors"
	"fmt"
	"time"

	"github.com/enterprise-erp/core/internal/accounting"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrPaymentAmountZero         = errors.New("payment amount must be greater than zero")
	ErrExchangeRateZero          = errors.New("exchange rate must be greater than zero")
	ErrUnbalancedPaymentGL       = errors.New("Tryton Invariant: generated GL entries for payment are unbalanced")
	ErrAccountsRequired          = errors.New("bank/cash account and receivable/payable account are required")
	ErrCannotCancelUnpostedPayment = errors.New("can only cancel a posted/submitted payment")
)

type PaymentType string

const (
	PaymentReceive PaymentType = "Receive" // Customer Payment
	PaymentPay     PaymentType = "Pay"     // Supplier Payment
)

// PaymentEntry combines ERPNext's payment_entry.json schema with Tryton's Forex Realized Gain/Loss engine
type PaymentEntry struct {
	BaseDocument
	CompanyID              uuid.UUID   `json:"company_id"`
	PaymentNumber          string      `json:"payment_number"` // PAY-YYYY-#####
	PaymentType            PaymentType `json:"payment_type"`
	PartyType              string      `json:"party_type"` // 'Customer', 'Supplier'
	PartyID                *uuid.UUID  `json:"party_id,omitempty"`
	PartyName              string      `json:"party_name"`
	PostingDate            time.Time   `json:"posting_date"`
	PaidToAccountID        uuid.UUID   `json:"paid_to_account_id"`   // Bank / Cash
	PaidFromAccountID      uuid.UUID   `json:"paid_from_account_id"` // Accounts Receivable
	Currency               string      `json:"currency"`
	ExchangeRate           Money       `json:"exchange_rate"` // Payment date exchange rate
	PaidAmount             Money       `json:"paid_amount"`
	BasePaidAmount         Money       `json:"base_paid_amount"`
	AllocatedAmount        Money       `json:"allocated_amount"`
	ReferenceInvoiceID     *uuid.UUID  `json:"reference_invoice_id,omitempty"`
	ReferenceInvoiceNumber string      `json:"reference_invoice_number,omitempty"`
	ReferenceInvoiceRate   Money       `json:"reference_invoice_rate"` // Original invoice conversion rate
	ForexGainLoss          Money       `json:"forex_gain_loss"`        // Positive = Gain, Negative = Loss
	ForexAccountID         *uuid.UUID  `json:"forex_account_id,omitempty"`
	Remarks                string      `json:"remarks"`
	PostedAt               *time.Time  `json:"posted_at,omitempty"`
	CancelledAt            *time.Time  `json:"cancelled_at,omitempty"`
}

func init() {
	GetRegistry().Register(&DocTypeDefinition{
		Name:          "PaymentEntry",
		Module:        "Accounts",
		IsSubmittable: true,
		TableName:     "payment_entries",
		Fields: map[string]DocField{
			"payment_number":   {FieldName: "payment_number", Label: "Payment #", FieldType: TypeData, Unique: true, Required: true},
			"payment_type":     {FieldName: "payment_type", Label: "Type", FieldType: TypeSelect, Required: true},
			"party_name":       {FieldName: "party_name", Label: "Party", FieldType: TypeData, Required: true},
			"posting_date":     {FieldName: "posting_date", Label: "Posting Date", FieldType: TypeDate, Required: true},
			"currency":         {FieldName: "currency", Label: "Currency", FieldType: TypeLink, Options: "Currency", Required: true},
			"exchange_rate":    {FieldName: "exchange_rate", Label: "Payment Rate", FieldType: TypeFloat, Required: true},
			"paid_amount":      {FieldName: "paid_amount", Label: "Paid Amount", FieldType: TypeCurrency, Required: true},
			"paid_to_account":  {FieldName: "paid_to_account_id", Label: "Bank / Cash Account", FieldType: TypeLink, Options: "Account", Required: true},
			"paid_from_account":{FieldName: "paid_from_account_id", Label: "Receivable / Payable Account", FieldType: TypeLink, Options: "Account", Required: true},
		},
		FieldOrder: []string{"payment_number", "payment_type", "party_name", "posting_date", "currency", "exchange_rate", "paid_amount", "paid_to_account", "paid_from_account"},
	})
}

// CalculateForex reconciles the payment against the invoice and calculates Realized Forex Gain or Loss with zero float
func (pe *PaymentEntry) CalculateForex() error {
	if pe.PaidAmount.IsZero() || pe.PaidAmount.IsNegative() {
		return ErrPaymentAmountZero
	}

	if pe.ExchangeRate.IsZero() || pe.ExchangeRate.IsNegative() {
		return ErrExchangeRateZero
	}

	// Calculate base paid amount at payment date rate
	pe.BasePaidAmount = pe.PaidAmount.Mul(pe.ExchangeRate)

	// If linked to an invoice, compute forex difference
	if pe.ReferenceInvoiceRate.IsPositive() {
		if pe.AllocatedAmount.IsZero() {
			pe.AllocatedAmount = pe.PaidAmount
		}

		// Relieve Accounts Receivable in base currency at the original invoice rate
		baseARRelief := pe.AllocatedAmount.Mul(pe.ReferenceInvoiceRate)

		// Forex Variance = Base Bank Received - Base AR Relieved
		// For customer payment: if we receive 118 BDT per USD and invoice was 115 BDT per USD -> Gain of +3 BDT per USD
		pe.ForexGainLoss = pe.BasePaidAmount.Sub(baseARRelief)
	} else {
		pe.ReferenceInvoiceRate = pe.ExchangeRate
		pe.ForexGainLoss = decimal.Zero
	}

	return nil
}

// GenerateGLEntries generates Tryton-compliant double-entry ledger rows:
// 1. Debit Bank/Cash (BasePaidAmount = PaidAmount * ExchangeRate)
// 2. Credit Accounts Receivable (BaseAllocated = PaidAmount * ReferenceInvoiceRate)
// 3. Forex Gain / Loss:
//    - If Gain (> 0): Credit 4200 - Forex Realized Gain (Base = ForexGainLoss)
//    - If Loss (< 0): Debit 5200 - Forex Realized Loss (Base = Abs(ForexGainLoss))
func (pe *PaymentEntry) GenerateGLEntries(bankCode, arCode, forexGainCode, forexLossCode string, forexGainAccountID, forexLossAccountID uuid.UUID) ([]accounting.GLEntry, error) {
	if err := pe.CalculateForex(); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	var entries []accounting.GLEntry

	// 1. Debit Bank / Cash Account (Realized Cash Received at Payment Exchange Rate)
	entries = append(entries, accounting.GLEntry{
		ID:           uuid.New(),
		CompanyID:    pe.CompanyID,
		PostingDate:  pe.PostingDate,
		VoucherType:  "Payment Entry",
		VoucherID:    pe.ID,
		VoucherNo:    pe.PaymentNumber,
		AccountID:    pe.PaidToAccountID,
		AccountCode:  bankCode,
		PartyType:    nil,
		Currency:     pe.Currency,
		Debit:        pe.PaidAmount,
		Credit:       decimal.Zero,
		ExchangeRate: pe.ExchangeRate,
		DebitBase:    pe.BasePaidAmount,
		CreditBase:   decimal.Zero,
		Remarks:      fmt.Sprintf("Payment %s received from %s", pe.PaymentNumber, pe.PartyName),
		IsCancelled:  false,
		CreatedAt:    now,
	})

	// 2. Credit Accounts Receivable (Relieve Customer Debt at Original Invoice Rate)
	baseARRelief := pe.AllocatedAmount.Mul(pe.ReferenceInvoiceRate)
	entries = append(entries, accounting.GLEntry{
		ID:           uuid.New(),
		CompanyID:    pe.CompanyID,
		PostingDate:  pe.PostingDate,
		VoucherType:  "Payment Entry",
		VoucherID:    pe.ID,
		VoucherNo:    pe.PaymentNumber,
		AccountID:    pe.PaidFromAccountID,
		AccountCode:  arCode,
		PartyType:    &pe.PartyName,
		Currency:     pe.Currency,
		Debit:        decimal.Zero,
		Credit:       pe.AllocatedAmount,
		ExchangeRate: pe.ReferenceInvoiceRate,
		DebitBase:    decimal.Zero,
		CreditBase:   baseARRelief,
		Remarks:      fmt.Sprintf("Payment %s against invoice %s", pe.PaymentNumber, pe.ReferenceInvoiceNumber),
		IsCancelled:  false,
		CreatedAt:    now,
	})

	// 3. Automated Realized Forex Gain or Loss Entry
	if !pe.ForexGainLoss.IsZero() {
		if pe.ForexGainLoss.IsPositive() {
			// FOREX GAIN: Credit Income Account 4200
			gainAmount := pe.ForexGainLoss
			entries = append(entries, accounting.GLEntry{
				ID:           uuid.New(),
				CompanyID:    pe.CompanyID,
				PostingDate:  pe.PostingDate,
				VoucherType:  "Payment Entry",
				VoucherID:    pe.ID,
				VoucherNo:    pe.PaymentNumber,
				AccountID:    forexGainAccountID,
				AccountCode:  forexGainCode,
				PartyType:    nil,
				Currency:     pe.Currency,
				Debit:        decimal.Zero,
				Credit:       decimal.Zero, // Forex variances exist strictly in functional base currency
				ExchangeRate: pe.ExchangeRate,
				DebitBase:    decimal.Zero,
				CreditBase:   gainAmount,
				Remarks:      fmt.Sprintf("Forex Realized Gain on %s settlement (rate: %s vs %s)", pe.ReferenceInvoiceNumber, pe.ExchangeRate.String(), pe.ReferenceInvoiceRate.String()),
				IsCancelled:  false,
				CreatedAt:    now,
			})
		} else {
			// FOREX LOSS: Debit Expense Account 5200
			lossAmount := pe.ForexGainLoss.Abs()
			entries = append(entries, accounting.GLEntry{
				ID:           uuid.New(),
				CompanyID:    pe.CompanyID,
				PostingDate:  pe.PostingDate,
				VoucherType:  "Payment Entry",
				VoucherID:    pe.ID,
				VoucherNo:    pe.PaymentNumber,
				AccountID:    forexLossAccountID,
				AccountCode:  forexLossCode,
				PartyType:    nil,
				Currency:     pe.Currency,
				Debit:        decimal.Zero,
				Credit:       decimal.Zero,
				ExchangeRate: pe.ExchangeRate,
				DebitBase:    lossAmount,
				CreditBase:   decimal.Zero,
				Remarks:      fmt.Sprintf("Forex Realized Loss on %s settlement (rate: %s vs %s)", pe.ReferenceInvoiceNumber, pe.ExchangeRate.String(), pe.ReferenceInvoiceRate.String()),
				IsCancelled:  false,
				CreatedAt:    now,
			})
		}
	}

	// Tryton Balance Validation Invariant: SUM(Debit Base) == SUM(Credit Base)
	totalDebitBase := decimal.Zero
	totalCreditBase := decimal.Zero
	for _, entry := range entries {
		totalDebitBase = totalDebitBase.Add(entry.DebitBase)
		totalCreditBase = totalCreditBase.Add(entry.CreditBase)
	}

	if !totalDebitBase.Equal(totalCreditBase) {
		return nil, fmt.Errorf("%w: debit=%s, credit=%s, diff=%s",
			ErrUnbalancedPaymentGL, totalDebitBase.String(), totalCreditBase.String(), totalDebitBase.Sub(totalCreditBase).String())
	}

	return entries, nil
}

// GenerateReversalGLEntries creates Tryton-style contra reversal lines when payment is cancelled
func (pe *PaymentEntry) GenerateReversalGLEntries(originalEntries []accounting.GLEntry) []accounting.GLEntry {
	now := time.Now().UTC()
	var reversals []accounting.GLEntry

	for _, orig := range originalEntries {
		reversals = append(reversals, accounting.GLEntry{
			ID:           uuid.New(),
			CompanyID:    orig.CompanyID,
			PostingDate:  pe.PostingDate,
			VoucherType:  "Payment Entry (Reversal)",
			VoucherID:    orig.VoucherID,
			VoucherNo:    orig.VoucherNo,
			AccountID:    orig.AccountID,
			AccountCode:  orig.AccountCode,
			PartyType:    orig.PartyType,
			PartyID:      orig.PartyID,
			Currency:     orig.Currency,
			// Swap Debit and Credit
			Debit:        orig.Credit,
			Credit:       orig.Debit,
			ExchangeRate: orig.ExchangeRate,
			DebitBase:    orig.CreditBase,
			CreditBase:   orig.DebitBase,
			Remarks:      fmt.Sprintf("Reversal contra entry for %s", orig.VoucherNo),
			IsCancelled:  true,
			CreatedAt:    now,
		})
	}

	return reversals
}
