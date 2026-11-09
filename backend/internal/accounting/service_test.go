package accounting

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestZeroFloatPrecision(t *testing.T) {
	// Classic floating point trap: 0.1 + 0.2 == 0.30000000000000004 in float64
	d1 := decimal.RequireFromString("0.1")
	d2 := decimal.RequireFromString("0.2")
	expected := decimal.RequireFromString("0.3")

	result := d1.Add(d2)
	if !result.Equal(expected) {
		t.Fatalf("Decimal precision failed: got %s, expected %s", result.String(), expected.String())
	}
}

func TestMultiCurrencyExchangeRateMath(t *testing.T) {
	// USD 1,250.75 @ Exchange Rate 117.85 BDT/USD
	amountUSD := decimal.RequireFromString("1250.75")
	exchangeRate := decimal.RequireFromString("117.85")

	amountBDT := amountUSD.Mul(exchangeRate)
	// 1250.75 * 117.85 = 147400.8875
	expectedBDT := decimal.RequireFromString("147400.8875")

	if !amountBDT.Equal(expectedBDT) {
		t.Fatalf("Exchange rate math failed: got %s, expected %s", amountBDT.String(), expectedBDT.String())
	}
}

func TestUnbalancedDebitCreditRule(t *testing.T) {
	debitTotal := decimal.RequireFromString("500.00")
	creditTotal := decimal.RequireFromString("499.99")

	if debitTotal.Equal(creditTotal) {
		t.Fatalf("Unbalanced amounts incorrectly reported as equal")
	}

	diff := debitTotal.Sub(creditTotal)
	if !diff.Equal(decimal.RequireFromString("0.01")) {
		t.Fatalf("Discrepancy calculation error: got %s", diff.String())
	}
}

func TestTrialBalanceZeroFloatInvariant(t *testing.T) {
	svc := &Service{}
	report, err := svc.GetTrialBalance(nil, [16]byte{1}, true)
	if err != nil {
		t.Fatalf("failed to generate trial balance: %v", err)
	}

	if !report.IsBalanced {
		t.Fatalf("expected trial balance to be strictly balanced, diff=%s", report.Difference)
	}

	if !report.TotalDebit.Equal(report.TotalCredit) {
		t.Errorf("debit (%s) != credit (%s)", report.TotalDebit, report.TotalCredit)
	}

	if !report.IsConsolidated {
		t.Error("expected consolidated flag to be true")
	}
}

func TestInterBranchTransfer(t *testing.T) {
	svc := &Service{}
	fromID := [16]byte{1}
	toID := [16]byte{2}

	// 1. Valid transfer between two distinct branches
	req := InterBranchTransferRequest{
		FromCompanyID: fromID,
		ToCompanyID:   toID,
		Amount:        decimal.RequireFromString("150000.00"),
		Currency:      "BDT",
		Remarks:       "Inventory replenishment transfer",
	}

	res, err := svc.CreateInterBranchTransfer(nil, req)
	if err != nil {
		t.Fatalf("expected successful inter-branch transfer, got error: %v", err)
	}

	if !res.IsBalanced {
		t.Error("expected inter-branch transfer clearing vouchers to be balanced")
	}

	// 2. Transfer between identical branches MUST fail
	invalidReq := req
	invalidReq.ToCompanyID = fromID

	_, err = svc.CreateInterBranchTransfer(nil, invalidReq)
	if err != ErrSameBranchTransfer {
		t.Errorf("expected ErrSameBranchTransfer, got: %v", err)
	}
}

func TestPartyLedgerRunningBalance(t *testing.T) {
	svc := &Service{}
	partyID := [16]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x01} // Acme
	companyID := [16]byte{1}

	report, err := svc.GetPartyLedger(nil, companyID, partyID, true)
	if err != nil {
		t.Fatalf("failed to get party ledger: %v", err)
	}

	if len(report.Lines) == 0 {
		t.Fatal("expected party ledger lines, got 0")
	}

	// Verify running balance invariant: Each line's balance must equal previous balance + debit - credit
	var running decimal.Decimal = report.OpeningBalance
	for idx, line := range report.Lines {
		running = running.Add(line.Debit).Sub(line.Credit)
		if !line.Balance.Equal(running) {
			t.Errorf("line %d running balance mismatch: got %s, expected %s", idx, line.Balance, running)
		}
	}

	// Final running balance must equal closing balance
	if !report.ClosingBalance.Equal(running) {
		t.Errorf("closing balance mismatch: got %s, expected %s", report.ClosingBalance, running)
	}
}
