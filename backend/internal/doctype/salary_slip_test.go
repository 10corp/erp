package doctype

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestSalarySlipLifecycleAndAccountingPost(t *testing.T) {
	slip := &SalarySlip{
		BaseDocument: BaseDocument{
			ID:        uuid.New(),
			DocStatus: DocDraft, // 0 = Draft
			CreatedAt: time.Now().UTC(),
		},
		SlipNumber:    "SLIP-2026-11-001",
		EmployeeID:    uuid.New(),
		EmployeeName:  "Rahim (Senior Sales Officer)",
		Month:         11,
		Year:          2026,
		BaseSalary:    decimal.RequireFromString("60000.00"),
		HouseRent:     decimal.RequireFromString("30000.00"),
		Medical:       decimal.RequireFromString("5000.00"),
		Conveyance:    decimal.RequireFromString("5000.00"),
		ProvidentFund: decimal.RequireFromString("6000.00"),
		TaxDeduction:  decimal.RequireFromString("4000.00"),
	}

	// 1. Calculate Totals
	slip.CalculateTotals()

	// Gross: 60k + 30k + 5k + 5k = 100,000.00
	if !slip.GrossPay.Equal(decimal.RequireFromString("100000.00")) {
		t.Fatalf("expected gross pay 100000.00, got %s", slip.GrossPay)
	}

	// Deductions: 6k + 4k = 10,000.00
	if !slip.TotalDeduction.Equal(decimal.RequireFromString("10000.00")) {
		t.Fatalf("expected total deduction 10000.00, got %s", slip.TotalDeduction)
	}

	// Net Pay: 100k - 10k = 90,000.00
	if !slip.NetPay.Equal(decimal.RequireFromString("90000.00")) {
		t.Fatalf("expected net pay 90000.00, got %s", slip.NetPay)
	}

	// 2. Submit Document (docstatus 0 -> 1)
	if err := slip.Submit(); err != nil {
		t.Fatalf("expected successful submission, got: %v", err)
	}
	if slip.DocStatus != DocSubmitted {
		t.Fatalf("expected docstatus=1 (Submitted), got %d", slip.DocStatus)
	}

	// Cannot submit an already submitted slip
	if err := slip.Submit(); err != ErrSalarySlipNotDraft {
		t.Fatalf("expected ErrSalarySlipNotDraft, got %v", err)
	}

	// 3. Generate General Ledger Lines (Debit Expense == Credit Payable + Deductions)
	expenseAcc := uuid.New()
	payableAcc := uuid.New()
	withholdingAcc := uuid.New()

	glLines := slip.GenerateGLEntryLines(expenseAcc, payableAcc, withholdingAcc)
	if len(glLines) != 3 {
		t.Fatalf("expected 3 GL lines, got %d", len(glLines))
	}

	var sumDebit, sumCredit decimal.Decimal
	for _, l := range glLines {
		sumDebit = sumDebit.Add(l["debit"].(decimal.Decimal))
		sumCredit = sumCredit.Add(l["credit"].(decimal.Decimal))
	}

	if !sumDebit.Equal(sumCredit) {
		t.Fatalf("GL entries unbalanced: debit=%s, credit=%s", sumDebit, sumCredit)
	}
	if !sumDebit.Equal(decimal.RequireFromString("100000.00")) {
		t.Fatalf("GL debit sum must equal gross pay 100000.00, got %s", sumDebit)
	}

	// 4. Cancel Document (docstatus 1 -> 2)
	if err := slip.Cancel(); err != nil {
		t.Fatalf("expected successful cancellation, got: %v", err)
	}
	if slip.DocStatus != DocCancelled {
		t.Fatalf("expected docstatus=2 (Cancelled), got %d", slip.DocStatus)
	}
}
