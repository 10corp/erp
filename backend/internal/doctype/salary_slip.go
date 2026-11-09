package doctype

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrSalarySlipNotDraft      = errors.New("salary slip is not in draft state")
	ErrSalarySlipNotSubmitted  = errors.New("only submitted salary slips can be cancelled")
	ErrSalarySlipMathMismatch  = errors.New("net pay must strictly equal gross pay minus total deductions")
)

// SalarySlip represents ERPNext's clean HR Payroll document powered by Tryton's docstatus lifecycle
type SalarySlip struct {
	BaseDocument
	TenantID        uuid.UUID       `json:"tenant_id"`
	CompanyID       uuid.UUID       `json:"company_id"`
	SlipNumber      string          `json:"slip_number"`
	EmployeeID      uuid.UUID       `json:"employee_id"`
	EmployeeName    string          `json:"employee_name"`
	Designation     string          `json:"designation"`
	Department      string          `json:"department"`
	PostingDate     time.Time       `json:"posting_date"`
	Month           int16           `json:"month"`
	Year            int             `json:"year"`
	
	// Zero-Float Earnings
	BaseSalary      Money           `json:"base_salary"`
	HouseRent       Money           `json:"house_rent"`
	Medical         Money           `json:"medical"`
	Conveyance      Money           `json:"conveyance"`
	GrossPay        Money           `json:"gross_pay"`
	
	// Zero-Float Deductions
	ProvidentFund   Money           `json:"provident_fund"`
	TaxDeduction    Money           `json:"tax_deduction"`
	TotalDeduction  Money           `json:"total_deduction"`
	
	// Net Pay
	NetPay          Money           `json:"net_pay"`
	
	JournalEntryID  *uuid.UUID      `json:"journal_entry_id,omitempty"`
	Remarks         string          `json:"remarks"`
}

func init() {
	GetRegistry().Register(&DocTypeDefinition{
		Name:          "Salary Slip",
		Module:        "Human Resources & Payroll",
		IsSubmittable: true,
		TableName:     "hr_salary_slips",
		Fields: map[string]DocField{
			"slip_number":     {FieldName: "slip_number", Label: "Slip Number", FieldType: TypeData, Required: true, Unique: true},
			"employee_id":     {FieldName: "employee_id", Label: "Employee", FieldType: TypeLink, Options: "Employee", Required: true},
			"month":           {FieldName: "month", Label: "Month", FieldType: TypeInt, Required: true},
			"year":            {FieldName: "year", Label: "Year", FieldType: TypeInt, Required: true},
			"gross_pay":       {FieldName: "gross_pay", Label: "Gross Pay", FieldType: TypeCurrency, Required: true},
			"total_deduction": {FieldName: "total_deduction", Label: "Total Deductions", FieldType: TypeCurrency, Required: true},
			"net_pay":         {FieldName: "net_pay", Label: "Net Salary", FieldType: TypeCurrency, Required: true},
		},
		FieldOrder: []string{"slip_number", "employee_id", "month", "year", "gross_pay", "total_deduction", "net_pay"},
	})
}

// CalculateTotals enforces mathematical integrity using shopspring/decimal
func (s *SalarySlip) CalculateTotals() {
	s.GrossPay = s.BaseSalary.Add(s.HouseRent).Add(s.Medical).Add(s.Conveyance)
	s.TotalDeduction = s.ProvidentFund.Add(s.TaxDeduction)
	s.NetPay = s.GrossPay.Sub(s.TotalDeduction)
}

// Submit transitions docstatus from 0 (Draft) to 1 (Submitted), validating zero-float invariants
func (s *SalarySlip) Submit() error {
	if s.DocStatus != DocDraft {
		return ErrSalarySlipNotDraft
	}

	calculatedNet := s.GrossPay.Sub(s.TotalDeduction)
	if !s.NetPay.Equal(calculatedNet) {
		return fmt.Errorf("%w: net_pay=%s, calculated=%s", ErrSalarySlipMathMismatch, s.NetPay, calculatedNet)
	}

	if s.NetPay.IsNegative() {
		return errors.New("net pay cannot be negative")
	}

	s.DocStatus = DocSubmitted
	s.UpdatedAt = time.Now().UTC()
	return nil
}

// GenerateGLEntryLines produces double-entry balanced lines for salary expense vs payables
func (s *SalarySlip) GenerateGLEntryLines(salaryExpenseAccID, salaryPayableAccID, taxWithholdingAccID uuid.UUID) []map[string]interface{} {
	lines := []map[string]interface{}{
		{
			"account_id":  salaryExpenseAccID,
			"account_code": "5020",
			"debit":       s.GrossPay,
			"credit":      decimal.Zero,
			"remarks":     fmt.Sprintf("Salary Expense for %s (%s)", s.EmployeeName, s.SlipNumber),
		},
		{
			"account_id":  salaryPayableAccID,
			"account_code": "2010",
			"debit":       decimal.Zero,
			"credit":      s.NetPay,
			"remarks":     fmt.Sprintf("Net Salary Payable to %s (%s)", s.EmployeeName, s.SlipNumber),
		},
	}

	if s.TotalDeduction.GreaterThan(decimal.Zero) {
		lines = append(lines, map[string]interface{}{
			"account_id":  taxWithholdingAccID,
			"account_code": "2020",
			"debit":       decimal.Zero,
			"credit":      s.TotalDeduction,
			"remarks":     fmt.Sprintf("Statutory Tax & PF Withheld for %s (%s)", s.EmployeeName, s.SlipNumber),
		})
	}

	return lines
}

// Cancel transitions docstatus from 1 (Submitted) to 2 (Cancelled) via contra reversal
func (s *SalarySlip) Cancel() error {
	if s.DocStatus != DocSubmitted {
		return ErrSalarySlipNotSubmitted
	}

	s.DocStatus = DocCancelled
	s.UpdatedAt = time.Now().UTC()
	return nil
}
