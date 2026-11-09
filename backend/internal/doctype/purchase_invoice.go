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
	ErrPurchaseInvoiceItemsRequired  = errors.New("purchase invoice requires at least one item")
	ErrPurchaseAccountMappingMissing = errors.New("accounts payable and expense/stock accounts are required")
	ErrUnbalancedPurchaseGL          = errors.New("Tryton Invariant: generated GL entries for purchase invoice are unbalanced")
)

type PurchaseInvoiceItem struct {
	ID          uuid.UUID `json:"id"`
	ParentID    uuid.UUID `json:"parent_id"`
	ItemCode    string    `json:"item_code"`
	ItemName    string    `json:"item_name"`
	Description string    `json:"description"`
	Qty         Money     `json:"qty"`
	Rate        Money     `json:"rate"`
	Amount      Money     `json:"amount"` // Qty * Rate
	TaxRate     Money     `json:"tax_rate"`
	TaxAmount   Money     `json:"tax_amount"`
	BaseRate    Money     `json:"base_rate"`
	BaseAmount  Money     `json:"base_amount"`
}

type PurchaseInvoice struct {
	BaseDocument
	CompanyID         uuid.UUID             `json:"company_id"`
	InvoiceNumber     string                `json:"invoice_number"` // PINV-YYYY-#####
	SupplierInvoiceNo string                `json:"supplier_invoice_no,omitempty"`
	SupplierID        *uuid.UUID            `json:"supplier_id,omitempty"`
	SupplierName      string                `json:"supplier_name"`
	PostingDate       time.Time             `json:"posting_date"`
	DueDate           time.Time             `json:"due_date"`
	Currency          string                `json:"currency"`
	ConversionRate    Money                 `json:"conversion_rate"`
	NetTotal          Money                 `json:"net_total"`
	TaxTotal          Money                 `json:"tax_total"`
	GrandTotal        Money                 `json:"grand_total"`
	BaseNetTotal      Money                 `json:"base_net_total"`
	BaseTaxTotal      Money                 `json:"base_tax_total"`
	BaseGrandTotal    Money                 `json:"base_grand_total"`
	CreditToAccountID uuid.UUID             `json:"credit_to_account_id"`     // Accounts Payable (2001)
	ExpenseAccountID  uuid.UUID             `json:"expense_account_id"`       // Stock in Hand (1030) or Expense
	TaxAccountID      *uuid.UUID            `json:"tax_account_id,omitempty"` // Input VAT / Tax Receivable
	Items             []PurchaseInvoiceItem `json:"items"`
	Remarks           string                `json:"remarks"`
	PostedAt          *time.Time            `json:"posted_at,omitempty"`
	CancelledAt       *time.Time            `json:"cancelled_at,omitempty"`
}

func init() {
	GetRegistry().Register(&DocTypeDefinition{
		Name:          "Purchase Invoice",
		Module:        "Procurement",
		TableName:     "purchase_invoices",
		IsSubmittable: true,
	})
}

// CalculateTotals computes line totals, taxes, and functional base currency amounts
func (pi *PurchaseInvoice) CalculateTotals() error {
	if len(pi.Items) == 0 {
		return ErrPurchaseInvoiceItemsRequired
	}

	if pi.ConversionRate.IsZero() {
		pi.ConversionRate = decimal.NewFromInt(1)
	}

	netTotal := decimal.Zero
	taxTotal := decimal.Zero

	for i := range pi.Items {
		item := &pi.Items[i]
		item.Amount = item.Qty.Mul(item.Rate).Round(6)

		if !item.TaxRate.IsZero() {
			item.TaxAmount = item.Amount.Mul(item.TaxRate).Div(decimal.NewFromInt(100)).Round(6)
		} else {
			item.TaxAmount = decimal.Zero
		}

		item.BaseRate = item.Rate.Mul(pi.ConversionRate).Round(6)
		item.BaseAmount = item.Amount.Mul(pi.ConversionRate).Round(6)

		netTotal = netTotal.Add(item.Amount)
		taxTotal = taxTotal.Add(item.TaxAmount)
	}

	pi.NetTotal = netTotal
	pi.TaxTotal = taxTotal
	pi.GrandTotal = netTotal.Add(taxTotal)

	pi.BaseNetTotal = pi.NetTotal.Mul(pi.ConversionRate).Round(6)
	pi.BaseTaxTotal = pi.TaxTotal.Mul(pi.ConversionRate).Round(6)
	pi.BaseGrandTotal = pi.GrandTotal.Mul(pi.ConversionRate).Round(6)

	return nil
}

// GenerateGLEntries constructs balanced, audit-proof GL Entries for Accounts Payable
func (pi *PurchaseInvoice) GenerateGLEntries() ([]accounting.GLEntry, error) {
	if pi.CreditToAccountID == uuid.Nil || pi.ExpenseAccountID == uuid.Nil {
		return nil, ErrPurchaseAccountMappingMissing
	}

	if pi.GrandTotal.IsZero() {
		if err := pi.CalculateTotals(); err != nil {
			return nil, err
		}
	}

	var entries []accounting.GLEntry
	remarks := fmt.Sprintf("Purchase Invoice: %s - %s", pi.InvoiceNumber, pi.SupplierName)
	partyType := "Supplier"

	// 1. Credit to Accounts Payable (Trade Creditors Liability)
	entries = append(entries, accounting.GLEntry{
		ID:           uuid.New(),
		CompanyID:    pi.CompanyID,
		PostingDate:  pi.PostingDate,
		VoucherType:  "Purchase Invoice",
		VoucherID:    pi.ID,
		VoucherNo:    pi.InvoiceNumber,
		AccountID:    pi.CreditToAccountID,
		AccountCode:  "2001",
		PartyType:    &partyType,
		PartyID:      pi.SupplierID,
		Currency:     pi.Currency,
		Debit:        decimal.Zero,
		Credit:       pi.GrandTotal,
		ExchangeRate: pi.ConversionRate,
		DebitBase:    decimal.Zero,
		CreditBase:   pi.BaseGrandTotal,
		Remarks:      remarks,
		CreatedAt:    time.Now().UTC(),
	})

	// 2. Debit to Expense / Stock In Hand (Asset)
	entries = append(entries, accounting.GLEntry{
		ID:           uuid.New(),
		CompanyID:    pi.CompanyID,
		PostingDate:  pi.PostingDate,
		VoucherType:  "Purchase Invoice",
		VoucherID:    pi.ID,
		VoucherNo:    pi.InvoiceNumber,
		AccountID:    pi.ExpenseAccountID,
		AccountCode:  "1030", // or expense
		Currency:     pi.Currency,
		Debit:        pi.NetTotal,
		Credit:       decimal.Zero,
		ExchangeRate: pi.ConversionRate,
		DebitBase:    pi.BaseNetTotal,
		CreditBase:   decimal.Zero,
		Remarks:      remarks,
		CreatedAt:    time.Now().UTC(),
	})

	// 3. Debit to Input VAT / Tax Receivable (if applicable)
	if !pi.TaxTotal.IsZero() && pi.TaxAccountID != nil {
		entries = append(entries, accounting.GLEntry{
			ID:           uuid.New(),
			CompanyID:    pi.CompanyID,
			PostingDate:  pi.PostingDate,
			VoucherType:  "Purchase Invoice",
			VoucherID:    pi.ID,
			VoucherNo:    pi.InvoiceNumber,
			AccountID:    *pi.TaxAccountID,
			AccountCode:  "1050",
			Currency:     pi.Currency,
			Debit:        pi.TaxTotal,
			Credit:       decimal.Zero,
			ExchangeRate: pi.ConversionRate,
			DebitBase:    pi.BaseTaxTotal,
			CreditBase:   decimal.Zero,
			Remarks:      remarks + " (Input VAT)",
			CreatedAt:    time.Now().UTC(),
		})
	}

	// 4. Tryton Invariant Check: Sum(DebitBase) == Sum(CreditBase)
	totalDebitBase := decimal.Zero
	totalCreditBase := decimal.Zero
	for _, e := range entries {
		totalDebitBase = totalDebitBase.Add(e.DebitBase)
		totalCreditBase = totalCreditBase.Add(e.CreditBase)
	}

	diff := totalDebitBase.Sub(totalCreditBase).Abs()

	// Multi-currency Round-off Guard (Account 5090)
	if !diff.IsZero() && diff.LessThanOrEqual(decimal.NewFromFloat(0.05)) {
		roundOffAccountID := uuid.MustParse("00000000-0000-0000-0000-000000005090")
		if totalDebitBase.LessThan(totalCreditBase) {
			// Under-debit: add debit to round-off
			entries = append(entries, accounting.GLEntry{
				ID:           uuid.New(),
				CompanyID:    pi.CompanyID,
				PostingDate:  pi.PostingDate,
				VoucherType:  "Purchase Invoice",
				VoucherID:    pi.ID,
				VoucherNo:    pi.InvoiceNumber,
				AccountID:    roundOffAccountID,
				AccountCode:  "5090",
				Currency:     pi.Currency,
				Debit:        decimal.Zero,
				Credit:       decimal.Zero,
				ExchangeRate: pi.ConversionRate,
				DebitBase:    diff,
				CreditBase:   decimal.Zero,
				Remarks:      "Multi-Currency Round-Off Adjustment",
				CreatedAt:    time.Now().UTC(),
			})
		} else {
			// Over-debit: add credit to round-off
			entries = append(entries, accounting.GLEntry{
				ID:           uuid.New(),
				CompanyID:    pi.CompanyID,
				PostingDate:  pi.PostingDate,
				VoucherType:  "Purchase Invoice",
				VoucherID:    pi.ID,
				VoucherNo:    pi.InvoiceNumber,
				AccountID:    roundOffAccountID,
				AccountCode:  "5090",
				Currency:     pi.Currency,
				Debit:        decimal.Zero,
				Credit:       decimal.Zero,
				ExchangeRate: pi.ConversionRate,
				DebitBase:    decimal.Zero,
				CreditBase:   diff,
				Remarks:      "Multi-Currency Round-Off Adjustment",
				CreatedAt:    time.Now().UTC(),
			})
		}
	} else if !diff.IsZero() {
		return nil, fmt.Errorf("%w: debit=%s, credit=%s, diff=%s", ErrUnbalancedPurchaseGL, totalDebitBase, totalCreditBase, diff)
	}

	return entries, nil
}

func (pi *PurchaseInvoice) Submit() ([]accounting.GLEntry, error) {
	if pi.DocStatus != DocDraft {
		return nil, fmt.Errorf("cannot submit purchase invoice in docstatus %d (must be Draft)", pi.DocStatus)
	}
	if err := pi.CalculateTotals(); err != nil {
		return nil, err
	}
	glEntries, err := pi.GenerateGLEntries()
	if err != nil {
		return nil, err
	}

	pi.DocStatus = DocSubmitted
	now := time.Now().UTC()
	pi.PostedAt = &now
	pi.UpdatedAt = now
	return glEntries, nil
}

func (pi *PurchaseInvoice) Cancel() ([]accounting.GLEntry, error) {
	if pi.DocStatus != DocSubmitted {
		return nil, fmt.Errorf("cannot cancel purchase invoice in docstatus %d (must be Submitted)", pi.DocStatus)
	}

	postedEntries, err := pi.GenerateGLEntries()
	if err != nil {
		return nil, err
	}

	// Tryton Contra Reversal: Swap Debits and Credits
	var contraEntries []accounting.GLEntry
	for _, entry := range postedEntries {
		contra := entry
		contra.ID = uuid.New()
		contra.Debit = entry.Credit
		contra.Credit = entry.Debit
		contra.DebitBase = entry.CreditBase
		contra.CreditBase = entry.DebitBase
		contra.Remarks = fmt.Sprintf("Contra Reversal of %s (%s)", entry.VoucherNo, entry.Remarks)
		contra.CreatedAt = time.Now().UTC()
		contraEntries = append(contraEntries, contra)
	}

	pi.DocStatus = DocCancelled
	now := time.Now().UTC()
	pi.CancelledAt = &now
	pi.UpdatedAt = now
	return contraEntries, nil
}
