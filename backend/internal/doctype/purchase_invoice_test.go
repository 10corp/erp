package doctype

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestPurchaseInvoiceGLGeneration(t *testing.T) {
	companyID := uuid.New()
	apAccountID := uuid.New()
	stockAccountID := uuid.New()
	vatAccountID := uuid.New()

	pi := &PurchaseInvoice{
		CompanyID:         companyID,
		InvoiceNumber:     "PINV-2026-0001",
		SupplierName:      "Silicon Valley Chipsets Inc",
		PostingDate:       time.Date(2026, 11, 10, 0, 0, 0, 0, time.UTC),
		DueDate:           time.Date(2026, 12, 10, 0, 0, 0, 0, time.UTC),
		Currency:          "USD",
		ConversionRate:    decimal.NewFromFloat(118.50), // 1 USD = 118.50 BDT
		CreditToAccountID: apAccountID,
		ExpenseAccountID:  stockAccountID,
		TaxAccountID:      &vatAccountID,
		Items: []PurchaseInvoiceItem{
			{
				ItemCode: "ITEM-CHIP-01",
				ItemName: "AI Processor Unit",
				Qty:      decimal.NewFromInt(10),
				Rate:     decimal.NewFromInt(500), // $5,000 net
				TaxRate:  decimal.NewFromInt(10),  // 10% VAT = $500
			},
		},
	}

	entries, err := pi.Submit()
	if err != nil {
		t.Fatalf("expected successful submit, got error: %v", err)
	}

	if pi.DocStatus != DocSubmitted {
		t.Errorf("expected docstatus 1, got %d", pi.DocStatus)
	}

	// Grand total: $5,500. Base grand total: $5,500 * 118.50 = 651,750 BDT
	expectedGrandTotal := decimal.NewFromInt(5500)
	expectedBaseGrand := decimal.NewFromFloat(651750.00)

	if !pi.GrandTotal.Equal(expectedGrandTotal) {
		t.Errorf("expected grand total %s, got %s", expectedGrandTotal, pi.GrandTotal)
	}
	if !pi.BaseGrandTotal.Equal(expectedBaseGrand) {
		t.Errorf("expected base grand total %s, got %s", expectedBaseGrand, pi.BaseGrandTotal)
	}

	// Verify Tryton Invariant: Sum(DebitBase) == Sum(CreditBase)
	totalDebitBase := decimal.Zero
	totalCreditBase := decimal.Zero

	for _, e := range entries {
		totalDebitBase = totalDebitBase.Add(e.DebitBase)
		totalCreditBase = totalCreditBase.Add(e.CreditBase)
	}

	if !totalDebitBase.Equal(totalCreditBase) {
		t.Fatalf("Tryton balance invariant violated! DebitBase=%s, CreditBase=%s", totalDebitBase, totalCreditBase)
	}
}

func TestPurchaseInvoiceContraReversal(t *testing.T) {
	companyID := uuid.New()
	apAccountID := uuid.New()
	stockAccountID := uuid.New()

	pi := &PurchaseInvoice{
		CompanyID:         companyID,
		InvoiceNumber:     "PINV-2026-0002",
		SupplierName:      "Tokyo Optoelectronics",
		PostingDate:       time.Now().UTC(),
		DueDate:           time.Now().UTC().AddDate(0, 1, 0),
		Currency:          "BDT",
		ConversionRate:    decimal.NewFromInt(1),
		CreditToAccountID: apAccountID,
		ExpenseAccountID:  stockAccountID,
		Items: []PurchaseInvoiceItem{
			{
				ItemCode: "ITEM-OPT-02",
				ItemName: "Fiber Transceiver",
				Qty:      decimal.NewFromInt(50),
				Rate:     decimal.NewFromInt(1200),
			},
		},
	}

	originalEntries, err := pi.Submit()
	if err != nil {
		t.Fatalf("submit failed: %v", err)
	}

	contraEntries, err := pi.Cancel()
	if err != nil {
		t.Fatalf("cancel failed: %v", err)
	}

	if pi.DocStatus != DocCancelled {
		t.Errorf("expected docstatus 2 (Cancelled), got %d", pi.DocStatus)
	}

	if len(originalEntries) != len(contraEntries) {
		t.Fatalf("expected %d contra entries, got %d", len(originalEntries), len(contraEntries))
	}

	// Verify mirror-image contra reversal
	for i := range originalEntries {
		orig := originalEntries[i]
		contra := contraEntries[i]

		if !orig.Debit.Equal(contra.Credit) || !orig.Credit.Equal(contra.Debit) {
			t.Errorf("contra entry did not properly invert debits/credits! orig: D=%s, C=%s; contra: D=%s, C=%s",
				orig.Debit, orig.Credit, contra.Debit, contra.Credit)
		}
	}
}

func TestPurchaseOrderStateAndTotals(t *testing.T) {
	po := &PurchaseOrder{
		PONumber:        "PO-2026-0001",
		SupplierName:    "Shenzhen Assemblies",
		TransactionDate: time.Now().UTC(),
		Currency:        "USD",
		ConversionRate:  decimal.NewFromFloat(118.00),
		Items: []PurchaseOrderItem{
			{
				ItemCode: "ITEM-PCB-01",
				ItemName: "Bare Motherboard PCB",
				Qty:      decimal.NewFromInt(100),
				Rate:     decimal.NewFromFloat(25.50),
				TaxRate:  decimal.NewFromInt(5), // 5% tax
			},
		},
	}

	if err := po.Submit(); err != nil {
		t.Fatalf("PO submit failed: %v", err)
	}

	if po.DocStatus != DocSubmitted {
		t.Errorf("expected PO docstatus 1, got %d", po.DocStatus)
	}

	// 100 * 25.50 = 2,550.00 Net; Tax = 127.50; Grand Total = 2,677.50
	expectedNet := decimal.NewFromFloat(2550.00)
	expectedTax := decimal.NewFromFloat(127.50)
	expectedGrand := decimal.NewFromFloat(2677.50)

	if !po.NetTotal.Equal(expectedNet) {
		t.Errorf("expected net total %s, got %s", expectedNet, po.NetTotal)
	}
	if !po.TaxTotal.Equal(expectedTax) {
		t.Errorf("expected tax total %s, got %s", expectedTax, po.TaxTotal)
	}
	if !po.GrandTotal.Equal(expectedGrand) {
		t.Errorf("expected grand total %s, got %s", expectedGrand, po.GrandTotal)
	}
}
