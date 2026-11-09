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
	ErrInvoiceItemsRequired = errors.New("sales invoice requires at least one item")
	ErrAccountMappingRequired = errors.New("accounts receivable and sales income accounts are required")
	ErrUnbalancedInvoiceGL   = errors.New("Tryton Invariant: generated GL entries for invoice are unbalanced")
)

type SalesInvoiceItem struct {
	ID          uuid.UUID `json:"id"`
	ParentID    uuid.UUID `json:"parent_id"`
	ItemCode    string    `json:"item_code"`
	ItemName    string    `json:"item_name"`
	Description string    `json:"description"`
	Qty         Money     `json:"qty"`        // shopspring/decimal
	Rate        Money     `json:"rate"`       // Transaction currency rate
	Amount      Money     `json:"amount"`     // Qty * Rate
	TaxRate     Money     `json:"tax_rate"`   // Percentage (e.g. 15.00 for 15% VAT)
	TaxAmount   Money     `json:"tax_amount"` // Amount * (TaxRate / 100)
	BaseRate    Money     `json:"base_rate"`
	BaseAmount  Money     `json:"base_amount"`
}

type SalesInvoice struct {
	BaseDocument
	CompanyID         uuid.UUID          `json:"company_id"`
	InvoiceNumber     string             `json:"invoice_number"` // SINV-YYYY-#####
	CustomerID        *uuid.UUID         `json:"customer_id,omitempty"`
	CustomerName      string             `json:"customer_name"`
	PostingDate       time.Time          `json:"posting_date"`
	DueDate           time.Time          `json:"due_date"`
	Currency          string             `json:"currency"`
	ConversionRate    Money              `json:"conversion_rate"`
	NetTotal          Money              `json:"net_total"`
	TaxTotal          Money              `json:"tax_total"`
	GrandTotal        Money              `json:"grand_total"`
	BaseNetTotal      Money              `json:"base_net_total"`
	BaseTaxTotal      Money              `json:"base_tax_total"`
	BaseGrandTotal    Money              `json:"base_grand_total"`
	DebitToAccountID  uuid.UUID          `json:"debit_to_account_id"` // Accounts Receivable
	IncomeAccountID   uuid.UUID          `json:"income_account_id"`   // Sales Revenue
	TaxAccountID      *uuid.UUID         `json:"tax_account_id,omitempty"` // Output VAT Payable
	Items             []SalesInvoiceItem `json:"items"`
	Remarks           string             `json:"remarks"`
	PostedAt          *time.Time         `json:"posted_at,omitempty"`
	CancelledAt       *time.Time         `json:"cancelled_at,omitempty"`
}

func init() {
	GetRegistry().Register(&DocTypeDefinition{
		Name:          "SalesInvoice",
		Module:        "Accounts",
		IsSubmittable: true,
		TableName:     "sales_invoices",
		Fields: map[string]DocField{
			"invoice_number":   {FieldName: "invoice_number", Label: "Invoice #", FieldType: TypeData, Unique: true, Required: true},
			"customer_name":    {FieldName: "customer_name", Label: "Customer", FieldType: TypeData, Required: true},
			"posting_date":     {FieldName: "posting_date", Label: "Posting Date", FieldType: TypeDate, Required: true},
			"due_date":         {FieldName: "due_date", Label: "Due Date", FieldType: TypeDate, Required: true},
			"currency":         {FieldName: "currency", Label: "Currency", FieldType: TypeLink, Options: "Currency", Required: true},
			"conversion_rate":  {FieldName: "conversion_rate", Label: "Exchange Rate", FieldType: TypeFloat, Required: true},
			"grand_total":      {FieldName: "grand_total", Label: "Grand Total", FieldType: TypeCurrency, ReadOnly: true},
			"debit_to_account": {FieldName: "debit_to_account_id", Label: "Receivable Account", FieldType: TypeLink, Options: "Account", Required: true},
			"income_account":   {FieldName: "income_account_id", Label: "Income Account", FieldType: TypeLink, Options: "Account", Required: true},
			"items":            {FieldName: "items", Label: "Items", FieldType: TypeTable, Options: "SalesInvoiceItem", Required: true},
		},
		FieldOrder: []string{"invoice_number", "customer_name", "posting_date", "due_date", "currency", "conversion_rate", "grand_total", "debit_to_account", "income_account", "items"},
	})
}

// CalculateTotals calculates taxes, item sub-totals and converts into functional base currency with zero-float precision
func (si *SalesInvoice) CalculateTotals() error {
	if len(si.Items) == 0 {
		return ErrInvoiceItemsRequired
	}

	if si.ConversionRate.IsZero() || si.ConversionRate.IsNegative() {
		si.ConversionRate = decimal.NewFromInt(1)
	}

	netTotal := decimal.Zero
	taxTotal := decimal.Zero
	hundred := decimal.NewFromInt(100)

	for i := range si.Items {
		item := &si.Items[i]
		if item.Qty.IsNegative() || item.Rate.IsNegative() {
			return ErrNegativeQuantity
		}

		item.Amount = item.Qty.Mul(item.Rate)
		if item.TaxRate.IsPositive() {
			item.TaxAmount = item.Amount.Mul(item.TaxRate).Div(hundred)
		} else {
			item.TaxAmount = decimal.Zero
		}

		item.BaseRate = item.Rate.Mul(si.ConversionRate)
		item.BaseAmount = item.Amount.Mul(si.ConversionRate)

		netTotal = netTotal.Add(item.Amount)
		taxTotal = taxTotal.Add(item.TaxAmount)
	}

	si.NetTotal = netTotal
	si.TaxTotal = taxTotal
	si.GrandTotal = netTotal.Add(taxTotal)

	si.BaseNetTotal = netTotal.Mul(si.ConversionRate)
	si.BaseTaxTotal = taxTotal.Mul(si.ConversionRate)
	si.BaseGrandTotal = si.GrandTotal.Mul(si.ConversionRate)

	return nil
}

// GenerateGLEntries produces immutable double-entry rows conforming strictly to Tryton GAAP/IFRS rules:
// Line 1: Debit Accounts Receivable (Full Grand Total in Base Currency)
// Line 2: Credit Sales Revenue (Net Total in Base Currency)
// Line 3: Credit VAT/Tax Payable (Tax Total in Base Currency, if applicable)
// Line 4: Round Off Guard (account_code: 5090) if any minute multi-currency fraction exists
func (si *SalesInvoice) GenerateGLEntries(receivableCode, incomeCode, taxCode string, roundOffAccountID *uuid.UUID, roundOffCode string) ([]accounting.GLEntry, error) {
	if si.GrandTotal.IsZero() {
		if err := si.CalculateTotals(); err != nil {
			return nil, err
		}
	}

	now := time.Now().UTC()
	var entries []accounting.GLEntry

	// 1. Debit Accounts Receivable (Trade Debtors)
	entries = append(entries, accounting.GLEntry{
		ID:           uuid.New(),
		CompanyID:    si.CompanyID,
		PostingDate:  si.PostingDate,
		VoucherType:  "Sales Invoice",
		VoucherID:    si.ID,
		VoucherNo:    si.InvoiceNumber,
		AccountID:    si.DebitToAccountID,
		AccountCode:  receivableCode,
		PartyType:    &si.CustomerName,
		Currency:     si.Currency,
		Debit:        si.GrandTotal,
		Credit:       decimal.Zero,
		ExchangeRate: si.ConversionRate,
		DebitBase:    si.BaseGrandTotal,
		CreditBase:   decimal.Zero,
		Remarks:      fmt.Sprintf("Invoice %s receivable debit for %s", si.InvoiceNumber, si.CustomerName),
		IsCancelled:  false,
		CreatedAt:    now,
	})

	// 2. Credit Sales Revenue
	entries = append(entries, accounting.GLEntry{
		ID:           uuid.New(),
		CompanyID:    si.CompanyID,
		PostingDate:  si.PostingDate,
		VoucherType:  "Sales Invoice",
		VoucherID:    si.ID,
		VoucherNo:    si.InvoiceNumber,
		AccountID:    si.IncomeAccountID,
		AccountCode:  incomeCode,
		PartyType:    nil,
		Currency:     si.Currency,
		Debit:        decimal.Zero,
		Credit:       si.NetTotal,
		ExchangeRate: si.ConversionRate,
		DebitBase:    decimal.Zero,
		CreditBase:   si.BaseNetTotal,
		Remarks:      fmt.Sprintf("Invoice %s revenue recognition", si.InvoiceNumber),
		IsCancelled:  false,
		CreatedAt:    now,
	})

	// 3. Credit Output Tax / VAT Payable (if tax > 0)
	if si.BaseTaxTotal.IsPositive() && si.TaxAccountID != nil {
		entries = append(entries, accounting.GLEntry{
			ID:           uuid.New(),
			CompanyID:    si.CompanyID,
			PostingDate:  si.PostingDate,
			VoucherType:  "Sales Invoice",
			VoucherID:    si.ID,
			VoucherNo:    si.InvoiceNumber,
			AccountID:    *si.TaxAccountID,
			AccountCode:  taxCode,
			PartyType:    nil,
			Currency:     si.Currency,
			Debit:        decimal.Zero,
			Credit:       si.TaxTotal,
			ExchangeRate: si.ConversionRate,
			DebitBase:    decimal.Zero,
			CreditBase:   si.BaseTaxTotal,
			Remarks:      fmt.Sprintf("Invoice %s output tax payable", si.InvoiceNumber),
			IsCancelled:  false,
			CreatedAt:    now,
		})
	}

	// Calculate sub-totals before round-off guard
	totalDebitBase := decimal.Zero
	totalCreditBase := decimal.Zero
	for _, entry := range entries {
		totalDebitBase = totalDebitBase.Add(entry.DebitBase)
		totalCreditBase = totalCreditBase.Add(entry.CreditBase)
	}

	// 4. Pro-Tip Rounding Guard: Account 5090 (Round Off Account)
	// Catches microscopic multi-currency rounding differences (<= 0.05 BDT) without altering tax or revenue
	diff := totalDebitBase.Sub(totalCreditBase)
	maxRoundingTolerance := decimal.RequireFromString("0.050000")

	if !diff.IsZero() && diff.Abs().LessThanOrEqual(maxRoundingTolerance) {
		if roundOffCode == "" {
			roundOffCode = "5090"
		}
		roAccountID := uuid.MustParse("00000000-0000-0000-0000-000000005090")
		if roundOffAccountID != nil && *roundOffAccountID != uuid.Nil {
			roAccountID = *roundOffAccountID
		}

		roundAmount := diff.Abs()
		if diff.IsPositive() {
			// Debit > Credit: add Credit to Round Off line
			entries = append(entries, accounting.GLEntry{
				ID:           uuid.New(),
				CompanyID:    si.CompanyID,
				PostingDate:  si.PostingDate,
				VoucherType:  "Sales Invoice",
				VoucherID:    si.ID,
				VoucherNo:    si.InvoiceNumber,
				AccountID:    roAccountID,
				AccountCode:  roundOffCode,
				Currency:     si.Currency,
				Debit:        decimal.Zero,
				Credit:       decimal.Zero, // Rounding adjustment applies strictly in base ledger currency
				ExchangeRate: si.ConversionRate,
				DebitBase:    decimal.Zero,
				CreditBase:   roundAmount,
				Remarks:      fmt.Sprintf("Invoice %s automatic round-off adjustment", si.InvoiceNumber),
				IsCancelled:  false,
				CreatedAt:    now,
			})
			totalCreditBase = totalCreditBase.Add(roundAmount)
		} else {
			// Credit > Debit: add Debit to Round Off line
			entries = append(entries, accounting.GLEntry{
				ID:           uuid.New(),
				CompanyID:    si.CompanyID,
				PostingDate:  si.PostingDate,
				VoucherType:  "Sales Invoice",
				VoucherID:    si.ID,
				VoucherNo:    si.InvoiceNumber,
				AccountID:    roAccountID,
				AccountCode:  roundOffCode,
				Currency:     si.Currency,
				Debit:        decimal.Zero,
				Credit:       decimal.Zero,
				ExchangeRate: si.ConversionRate,
				DebitBase:    roundAmount,
				CreditBase:   decimal.Zero,
				Remarks:      fmt.Sprintf("Invoice %s automatic round-off adjustment", si.InvoiceNumber),
				IsCancelled:  false,
				CreatedAt:    now,
			})
			totalDebitBase = totalDebitBase.Add(roundAmount)
		}
	}

	// Tryton Balance Validation Invariant: SUM(Debit Base) == SUM(Credit Base)
	if !totalDebitBase.Equal(totalCreditBase) {
		return nil, fmt.Errorf("%w: debit=%s, credit=%s, diff=%s",
			ErrUnbalancedInvoiceGL, totalDebitBase.String(), totalCreditBase.String(), totalDebitBase.Sub(totalCreditBase).String())
	}

	return entries, nil
}

// GenerateReversalGLEntries generates exact Tryton-style reversal contra entries for audit cancellation
func (si *SalesInvoice) GenerateReversalGLEntries(originalEntries []accounting.GLEntry) []accounting.GLEntry {
	now := time.Now().UTC()
	var reversals []accounting.GLEntry

	for _, orig := range originalEntries {
		rev := accounting.GLEntry{
			ID:           uuid.New(),
			CompanyID:    orig.CompanyID,
			PostingDate:  si.PostingDate,
			VoucherType:  "Sales Invoice (Reversal)",
			VoucherID:    orig.VoucherID,
			VoucherNo:    orig.VoucherNo,
			AccountID:    orig.AccountID,
			AccountCode:  orig.AccountCode,
			PartyType:    orig.PartyType,
			PartyID:      orig.PartyID,
			Currency:     orig.Currency,
			// Swap Debit and Credit to produce an exact mathematical contra entry
			Debit:        orig.Credit,
			Credit:       orig.Debit,
			ExchangeRate: orig.ExchangeRate,
			DebitBase:    orig.CreditBase,
			CreditBase:   orig.DebitBase,
			Remarks:      fmt.Sprintf("Reversal contra entry for %s", orig.VoucherNo),
			IsCancelled:  true,
			CreatedAt:    now,
		}
		reversals = append(reversals, rev)
	}

	return reversals
}
