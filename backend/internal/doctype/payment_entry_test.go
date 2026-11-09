package doctype

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestPaymentEntryForexRealizedGain(t *testing.T) {
	companyID := uuid.New()
	bankAccountID := uuid.New()
	arAccountID := uuid.New()
	forexGainID := uuid.New()
	forexLossID := uuid.New()

	payment := &PaymentEntry{
		BaseDocument: BaseDocument{
			ID:        uuid.New(),
			DocStatus: DocDraft,
			CreatedAt: time.Now(),
		},
		CompanyID:              companyID,
		PaymentNumber:          "PAY-2026-00001",
		PaymentType:            PaymentReceive,
		PartyType:              "Customer",
		PartyName:              "Acme Global Client",
		PostingDate:            time.Now(),
		PaidToAccountID:        bankAccountID, // Bank Account
		PaidFromAccountID:      arAccountID,   // Accounts Receivable
		Currency:               "USD",
		ExchangeRate:           decimal.RequireFromString("118.000000"), // Payment date: 118 BDT/USD
		PaidAmount:             decimal.RequireFromString("10000.00"),   // $10,000 USD
		AllocatedAmount:        decimal.RequireFromString("10000.00"),
		ReferenceInvoiceNumber: "SINV-2026-00042",
		ReferenceInvoiceRate:   decimal.RequireFromString("115.000000"), // Invoice date: 115 BDT/USD
	}

	// 1. Calculate Forex
	err := payment.CalculateForex()
	if err != nil {
		t.Fatalf("CalculateForex failed: %v", err)
	}

	// Base Paid Amount = 10,000 * 118 = 1,180,000 BDT
	expectedBasePaid := decimal.RequireFromString("1180000.000000")
	if !payment.BasePaidAmount.Equal(expectedBasePaid) {
		t.Fatalf("BasePaidAmount mismatch: got %s, expected %s", payment.BasePaidAmount.String(), expectedBasePaid.String())
	}

	// Forex Gain = 1,180,000 - (10,000 * 115) = +30,000 BDT
	expectedGain := decimal.RequireFromString("30000.000000")
	if !payment.ForexGainLoss.Equal(expectedGain) {
		t.Fatalf("ForexGain mismatch: got %s, expected %s", payment.ForexGainLoss.String(), expectedGain.String())
	}

	// 2. Generate and Verify Double-Entry GL Ledger Entries
	glEntries, err := payment.GenerateGLEntries("1001", "1020", "4200", "5200", forexGainID, forexLossID)
	if err != nil {
		t.Fatalf("GenerateGLEntries failed: %v", err)
	}

	if len(glEntries) != 3 {
		t.Fatalf("Expected 3 GL lines (Bank, AR, Forex Gain), got %d", len(glEntries))
	}

	// Verify Line 1: Debit Bank (1,180,000 BDT)
	bankLine := glEntries[0]
	if bankLine.AccountCode != "1001" || !bankLine.DebitBase.Equal(expectedBasePaid) || !bankLine.CreditBase.IsZero() {
		t.Fatalf("Bank Line invalid: debit_base=%s, credit_base=%s", bankLine.DebitBase.String(), bankLine.CreditBase.String())
	}

	// Verify Line 2: Credit AR (1,150,000 BDT)
	arLine := glEntries[1]
	expectedARRelief := decimal.RequireFromString("1150000.000000")
	if arLine.AccountCode != "1020" || !arLine.CreditBase.Equal(expectedARRelief) || !arLine.DebitBase.IsZero() {
		t.Fatalf("AR Line invalid: debit_base=%s, credit_base=%s", arLine.DebitBase.String(), arLine.CreditBase.String())
	}

	// Verify Line 3: Credit Forex Realized Gain (4200) (30,000 BDT)
	gainLine := glEntries[2]
	if gainLine.AccountCode != "4200" || !gainLine.CreditBase.Equal(expectedGain) || !gainLine.DebitBase.IsZero() {
		t.Fatalf("Forex Gain Line invalid: credit_base=%s", gainLine.CreditBase.String())
	}

	// Invariant check: Debit Bank (1,180,000) == Credit AR (1,150,000) + Credit Forex Gain (30,000)
	if !bankLine.DebitBase.Equal(arLine.CreditBase.Add(gainLine.CreditBase)) {
		t.Fatalf("Double-entry balance check failed for Forex Gain!")
	}
}

func TestPaymentEntryForexRealizedLoss(t *testing.T) {
	companyID := uuid.New()
	bankAccountID := uuid.New()
	arAccountID := uuid.New()
	forexGainID := uuid.New()
	forexLossID := uuid.New()

	payment := &PaymentEntry{
		BaseDocument: BaseDocument{
			ID:        uuid.New(),
			DocStatus: DocDraft,
		},
		CompanyID:              companyID,
		PaymentNumber:          "PAY-2026-00002",
		PaymentType:            PaymentReceive,
		PartyType:              "Customer",
		PartyName:              "Acme Global Client",
		PostingDate:            time.Now(),
		PaidToAccountID:        bankAccountID,
		PaidFromAccountID:      arAccountID,
		Currency:               "USD",
		ExchangeRate:           decimal.RequireFromString("114.000000"), // Payment date rate dropped to 114 BDT/USD
		PaidAmount:             decimal.RequireFromString("10000.00"),   // $10,000 USD
		AllocatedAmount:        decimal.RequireFromString("10000.00"),
		ReferenceInvoiceNumber: "SINV-2026-00042",
		ReferenceInvoiceRate:   decimal.RequireFromString("118.000000"), // Invoice date was 118 BDT/USD
	}

	// 1. Calculate Forex
	err := payment.CalculateForex()
	if err != nil {
		t.Fatalf("CalculateForex failed: %v", err)
	}

	// Bank Received = 10,000 * 114 = 1,140,000 BDT
	expectedBasePaid := decimal.RequireFromString("1140000.000000")
	if !payment.BasePaidAmount.Equal(expectedBasePaid) {
		t.Fatalf("BasePaidAmount mismatch: got %s", payment.BasePaidAmount.String())
	}

	// Forex Loss = 1,140,000 - 1,180,000 = -40,000 BDT
	expectedLoss := decimal.RequireFromString("-40000.000000")
	if !payment.ForexGainLoss.Equal(expectedLoss) {
		t.Fatalf("ForexLoss mismatch: got %s, expected %s", payment.ForexGainLoss.String(), expectedLoss.String())
	}

	// 2. Generate and Verify Double-Entry GL Ledger Entries
	glEntries, err := payment.GenerateGLEntries("1001", "1020", "4200", "5200", forexGainID, forexLossID)
	if err != nil {
		t.Fatalf("GenerateGLEntries failed: %v", err)
	}

	if len(glEntries) != 3 {
		t.Fatalf("Expected 3 GL lines, got %d", len(glEntries))
	}

	// Verify Line 1: Debit Bank (1,140,000 BDT)
	bankLine := glEntries[0]
	// Verify Line 2: Credit AR (1,180,000 BDT)
	arLine := glEntries[1]
	// Verify Line 3: Debit Forex Realized Loss (5200) (40,000 BDT)
	lossLine := glEntries[2]
	if lossLine.AccountCode != "5200" || !lossLine.DebitBase.Equal(decimal.RequireFromString("40000.000000")) {
		t.Fatalf("Forex Loss Line invalid: debit_base=%s", lossLine.DebitBase.String())
	}

	// Invariant check: Debit Bank (1,140,000) + Debit Loss (40,000) == Credit AR (1,180,000)
	if !bankLine.DebitBase.Add(lossLine.DebitBase).Equal(arLine.CreditBase) {
		t.Fatalf("Double-entry balance check failed for Forex Loss!")
	}

	// 3. Test Tryton Reversal Contra Generation
	reversals := payment.GenerateReversalGLEntries(glEntries)
	if len(reversals) != 3 {
		t.Fatalf("Expected 3 reversal lines, got %d", len(reversals))
	}
	// Reversal Bank must be Credit
	if !reversals[0].CreditBase.Equal(expectedBasePaid) || !reversals[0].DebitBase.IsZero() {
		t.Fatalf("Reversal Bank line invalid")
	}
	// Reversal Loss must be Credit
	if !reversals[2].CreditBase.Equal(decimal.RequireFromString("40000.000000")) || !reversals[2].DebitBase.IsZero() {
		t.Fatalf("Reversal Loss line invalid")
	}
}
