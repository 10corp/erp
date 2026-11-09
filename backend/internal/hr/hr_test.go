package hr

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestHRServiceLifecycle(t *testing.T) {
	ctx := context.Background()
	svc := NewService(nil, nil)

	// 1. List Employees
	emps, err := svc.ListEmployees(ctx, uuid.New())
	if err != nil {
		t.Fatalf("failed to list employees: %v", err)
	}
	if len(emps) < 3 {
		t.Fatalf("expected at least 3 employees, got %d", len(emps))
	}

	emp := emps[0]

	// 2. Create Salary Slip
	req := CreateSalarySlipRequest{
		CompanyID:     emp.CompanyID,
		EmployeeID:    emp.ID,
		PostingDate:   "2026-11-30",
		Month:         11,
		Year:          2026,
		BaseSalary:    decimal.RequireFromString("150000.00"),
		HouseRent:     decimal.RequireFromString("60000.00"),
		Medical:       decimal.RequireFromString("15000.00"),
		Conveyance:    decimal.RequireFromString("10000.00"),
		ProvidentFund: decimal.RequireFromString("15000.00"),
		TaxDeduction:  decimal.RequireFromString("20000.00"),
		Remarks:       "November Executive Salary",
	}

	slip, err := svc.CreateSalarySlip(ctx, req)
	if err != nil {
		t.Fatalf("failed to create salary slip: %v", err)
	}

	if slip.DocStatus != DocDraft {
		t.Fatalf("expected slip docstatus=0 (Draft), got %d", slip.DocStatus)
	}

	// 150k + 60k + 15k + 10k = 235k Gross
	if !slip.GrossPay.Equal(decimal.RequireFromString("235000.00")) {
		t.Errorf("gross pay mismatch: got %s, expected 235000.00", slip.GrossPay)
	}
	// 15k + 20k = 35k Deductions
	if !slip.TotalDeduction.Equal(decimal.RequireFromString("35000.00")) {
		t.Errorf("deduction mismatch: got %s, expected 35000.00", slip.TotalDeduction)
	}
	// 235k - 35k = 200k Net Pay
	if !slip.NetPay.Equal(decimal.RequireFromString("200000.00")) {
		t.Errorf("net pay mismatch: got %s, expected 200000.00", slip.NetPay)
	}

	// 3. Submit Salary Slip (docstatus 0 -> 1)
	submittedSlip, err := svc.SubmitSalarySlip(ctx, slip.ID)
	if err != nil {
		t.Fatalf("failed to submit salary slip: %v", err)
	}
	if submittedSlip.DocStatus != DocSubmitted {
		t.Fatalf("expected submitted docstatus=1, got %d", submittedSlip.DocStatus)
	}
	if submittedSlip.JournalEntryID == nil {
		t.Fatalf("expected linked JournalEntryID on submitted salary slip")
	}

	// 4. Cancel Salary Slip (docstatus 1 -> 2)
	cancelledSlip, err := svc.CancelSalarySlip(ctx, slip.ID)
	if err != nil {
		t.Fatalf("failed to cancel salary slip: %v", err)
	}
	if cancelledSlip.DocStatus != DocCancelled {
		t.Fatalf("expected cancelled docstatus=2, got %d", cancelledSlip.DocStatus)
	}

	// 5. Leave Application & Approval
	leaveReq := CreateLeaveRequest{
		CompanyID:  emp.CompanyID,
		EmployeeID: emp.ID,
		LeaveType:  "Casual Leave",
		FromDate:   "2026-11-20",
		ToDate:     "2026-11-21",
		TotalDays:  2,
		Reason:     "Personal Business",
	}

	leave, err := svc.CreateLeave(ctx, leaveReq)
	if err != nil {
		t.Fatalf("failed to create leave: %v", err)
	}
	if leave.DocStatus != DocDraft {
		t.Fatalf("expected leave docstatus=0, got %d", leave.DocStatus)
	}

	approvedLeave, err := svc.ApproveLeave(ctx, leave.ID, "System CFO")
	if err != nil {
		t.Fatalf("failed to approve leave: %v", err)
	}
	if approvedLeave.DocStatus != DocSubmitted {
		t.Fatalf("expected approved leave docstatus=1, got %d", approvedLeave.DocStatus)
	}

	// 6. Horilla Attendance Check-In Test
	checkInReq := CheckInRequest{
		CompanyID:  emp.CompanyID,
		EmployeeID: emp.ID,
		Source:     "Biometric",
	}
	attRecord, err := svc.RecordCheckIn(ctx, checkInReq)
	if err != nil {
		t.Fatalf("failed to record check-in: %v", err)
	}
	if attRecord.EmployeeID != emp.ID || attRecord.CheckIn == "" {
		t.Fatalf("invalid check-in record generated: %+v", attRecord)
	}

	// 7. Horilla Expense Reimbursement Claim Test
	claimReq := CreateClaimRequest{
		CompanyID:   emp.CompanyID,
		EmployeeID:  emp.ID,
		Category:    "Travel",
		Amount:      decimal.RequireFromString("4500.00"),
		Description: "Inter-branch train fare and meals",
		ReceiptRef:  "REC-TRAIN-01.pdf",
	}
	claim, err := svc.CreateClaim(ctx, claimReq)
	if err != nil {
		t.Fatalf("failed to create claim: %v", err)
	}
	if claim.DocStatus != DocDraft {
		t.Fatalf("expected claim docstatus=0, got %d", claim.DocStatus)
	}

	approvedClaim, err := svc.ApproveClaim(ctx, claim.ID)
	if err != nil {
		t.Fatalf("failed to approve claim: %v", err)
	}
	if approvedClaim.DocStatus != DocSubmitted || approvedClaim.JournalEntryID == nil {
		t.Fatalf("expected approved claim with GL Journal voucher linked")
	}

	cancelledClaim, err := svc.CancelClaim(ctx, claim.ID)
	if err != nil {
		t.Fatalf("failed to cancel claim: %v", err)
	}
	if cancelledClaim.DocStatus != DocCancelled {
		t.Fatalf("expected cancelled claim docstatus=2, got %d", cancelledClaim.DocStatus)
	}
}
