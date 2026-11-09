package doctype

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestSalesInvoiceTaxAndGLGeneration(t *testing.T) {
	companyID := uuid.New()
	arAccountID := uuid.New()
	incomeAccountID := uuid.New()
	taxAccountID := uuid.New()

	invoice := &SalesInvoice{
		BaseDocument: BaseDocument{
			ID:        uuid.New(),
			DocStatus: DocDraft,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		CompanyID:        companyID,
		InvoiceNumber:    "SINV-2026-00042",
		CustomerName:     "Acme International Corp",
		PostingDate:      time.Now(),
		DueDate:          time.Now().AddDate(0, 1, 0),
		Currency:         "USD",
		ConversionRate:   decimal.RequireFromString("117.85"), // 117.85 BDT per USD
		DebitToAccountID: arAccountID,
		IncomeAccountID:  incomeAccountID,
		TaxAccountID:     &taxAccountID,
		Items: []SalesInvoiceItem{
			{
				ID:          uuid.New(),
				ItemCode:    "PROD-001",
				ItemName:    "Enterprise Core Server",
				Qty:         decimal.RequireFromString("2.0"),
				Rate:        decimal.RequireFromString("5000.00"), // 2 * 5000 = 10,000 USD
				TaxRate:     decimal.RequireFromString("15.0"),    // 15% VAT = 1,500 USD
			},
			{
				ID:          uuid.New(),
				ItemCode:    "SERV-002",
				ItemName:    "Annual Maintenance SLA",
				Qty:         decimal.RequireFromString("1.0"),
				Rate:        decimal.RequireFromString("2500.00"), // 1 * 2500 = 2,500 USD
				TaxRate:     decimal.RequireFromString("15.0"),    // 15% VAT = 375 USD
			},
		},
	}

	// 1. Verify Calculations
	err := invoice.CalculateTotals()
	if err != nil {
		t.Fatalf("CalculateTotals failed: %v", err)
	}

	// Net Total: 10,000 + 2,500 = 12,500 USD
	expectedNetUSD := decimal.RequireFromString("12500.00")
	if !invoice.NetTotal.Equal(expectedNetUSD) {
		t.Fatalf("NetTotal mismatch: got %s, expected %s", invoice.NetTotal.String(), expectedNetUSD.String())
	}

	// Tax Total: 1,500 + 375 = 1,875 USD
	expectedTaxUSD := decimal.RequireFromString("1875.00")
	if !invoice.TaxTotal.Equal(expectedTaxUSD) {
		t.Fatalf("TaxTotal mismatch: got %s, expected %s", invoice.TaxTotal.String(), expectedTaxUSD.String())
	}

	// Grand Total: 12,500 + 1,875 = 14,375 USD
	expectedGrandUSD := decimal.RequireFromString("14375.00")
	if !invoice.GrandTotal.Equal(expectedGrandUSD) {
		t.Fatalf("GrandTotal mismatch: got %s, expected %s", invoice.GrandTotal.String(), expectedGrandUSD.String())
	}

	// Base Grand Total: 14,375 * 117.85 = 1,694,093.75 BDT
	expectedBaseGrandBDT := decimal.RequireFromString("1694093.75")
	if !invoice.BaseGrandTotal.Equal(expectedBaseGrandBDT) {
		t.Fatalf("BaseGrandTotal mismatch: got %s, expected %s", invoice.BaseGrandTotal.String(), expectedBaseGrandBDT.String())
	}

	// 2. Generate and Verify Double-Entry GL Ledger Entries with Round Off Guard
	glEntries, err := invoice.GenerateGLEntries("1020", "4001", "2050", nil, "5090")
	if err != nil {
		t.Fatalf("GenerateGLEntries failed: %v", err)
	}

	if len(glEntries) != 3 {
		t.Fatalf("Expected 3 GL lines (AR, Revenue, Tax), got %d", len(glEntries))
	}

	// Verify Line 1: Debit AR
	arLine := glEntries[0]
	if !arLine.DebitBase.Equal(expectedBaseGrandBDT) || !arLine.CreditBase.IsZero() {
		t.Fatalf("AR Line invalid: debit_base=%s, credit_base=%s", arLine.DebitBase.String(), arLine.CreditBase.String())
	}

	// Verify Line 2: Credit Sales Revenue (12,500 * 117.85 = 1,473,125.00 BDT)
	revLine := glEntries[1]
	expectedRevBaseBDT := decimal.RequireFromString("1473125.00")
	if !revLine.CreditBase.Equal(expectedRevBaseBDT) || !revLine.DebitBase.IsZero() {
		t.Fatalf("Revenue Line invalid: debit_base=%s, credit_base=%s", revLine.DebitBase.String(), revLine.CreditBase.String())
	}

	// Verify Line 3: Credit Tax Payable (1,875 * 117.85 = 220,968.75 BDT)
	taxLine := glEntries[2]
	expectedTaxBaseBDT := decimal.RequireFromString("220968.75")
	if !taxLine.CreditBase.Equal(expectedTaxBaseBDT) || !taxLine.DebitBase.IsZero() {
		t.Fatalf("Tax Line invalid: debit_base=%s, credit_base=%s", taxLine.DebitBase.String(), taxLine.CreditBase.String())
	}

	// Verify Mathematical Double-Entry Invariant: Debit == Credit (1,694,093.75 == 1,473,125.00 + 220,968.75)
	if !revLine.CreditBase.Add(taxLine.CreditBase).Equal(arLine.DebitBase) {
		t.Fatalf("Double-entry balance check failed!")
	}

	// 3. Test Tryton Reversal Contra Generation
	reversals := invoice.GenerateReversalGLEntries(glEntries)
	if len(reversals) != 3 {
		t.Fatalf("Expected 3 reversal lines, got %d", len(reversals))
	}

	// AR reversal must now be CREDIT
	if !reversals[0].CreditBase.Equal(expectedBaseGrandBDT) || !reversals[0].DebitBase.IsZero() {
		t.Fatalf("Reversal AR line invalid")
	}

	// Revenue reversal must now be DEBIT
	if !reversals[1].DebitBase.Equal(expectedRevBaseBDT) || !reversals[1].CreditBase.IsZero() {
		t.Fatalf("Reversal Revenue line invalid")
	}
}

func TestRoundOffAccountGuard(t *testing.T) {
	companyID := uuid.New()
	invoice := &SalesInvoice{
		BaseDocument: BaseDocument{
			ID:        uuid.New(),
			DocStatus: DocDraft,
		},
		CompanyID:        companyID,
		InvoiceNumber:    "SINV-2026-ROUNDTEST",
		CustomerName:     "Test Rounding Client",
		PostingDate:      time.Now(),
		Currency:         "USD",
		ConversionRate:   decimal.RequireFromString("117.850000"),
		DebitToAccountID: uuid.New(),
		IncomeAccountID:  uuid.New(),
		Items: []SalesInvoiceItem{
			{
				ItemCode: "ITEM-X",
				Qty:      decimal.RequireFromString("1.0"),
				Rate:     decimal.RequireFromString("100.00"),
			},
		},
	}

	_ = invoice.CalculateTotals()

	// Artificially induce a microscopic 0.000005 multi-currency rounding fraction on BaseGrandTotal
	invoice.BaseGrandTotal = invoice.BaseGrandTotal.Add(decimal.RequireFromString("0.000005"))

	roundOffAccountID := uuid.New()
	glEntries, err := invoice.GenerateGLEntries("1020", "4001", "2050", &roundOffAccountID, "5090")
	if err != nil {
		t.Fatalf("GenerateGLEntries failed with rounding guard: %v", err)
	}

	// Must contain: 1. Debit AR, 2. Credit Revenue, 3. Credit Round Off (5090)
	if len(glEntries) != 3 {
		t.Fatalf("Expected 3 entries (AR, Revenue, RoundOff), got %d", len(glEntries))
	}

	roundOffLine := glEntries[2]
	if roundOffLine.AccountCode != "5090" {
		t.Fatalf("Expected AccountCode 5090, got %s", roundOffLine.AccountCode)
	}

	expectedRoundDiff := decimal.RequireFromString("0.000005")
	if !roundOffLine.CreditBase.Equal(expectedRoundDiff) {
		t.Fatalf("Round off amount mismatch: got %s, expected %s", roundOffLine.CreditBase.String(), expectedRoundDiff.String())
	}

	// Verify exact balance invariant
	totalDebit := decimal.Zero
	totalCredit := decimal.Zero
	for _, e := range glEntries {
		totalDebit = totalDebit.Add(e.DebitBase)
		totalCredit = totalCredit.Add(e.CreditBase)
	}
	if !totalDebit.Equal(totalCredit) {
		t.Fatalf("Balance invariant failed: debit=%s, credit=%s", totalDebit.String(), totalCredit.String())
	}
}
