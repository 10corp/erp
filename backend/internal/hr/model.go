package hr

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type DocStatus int

const (
	DocDraft     DocStatus = 0 // Editable
	DocSubmitted DocStatus = 1 // Posted to General Ledger & Immutable
	DocCancelled DocStatus = 2 // Reversed via Contra Entry
)

// Department represents hierarchical organizational department
type Department struct {
	ID        uuid.UUID  `json:"id"`
	CompanyID uuid.UUID  `json:"company_id"`
	Code      string     `json:"code"`
	Name      string     `json:"name"`
	ParentID  *uuid.UUID `json:"parent_id,omitempty"`
	IsActive  bool       `json:"is_active"`
}

// Designation represents job title / position
type Designation struct {
	ID          uuid.UUID `json:"id"`
	CompanyID   uuid.UUID `json:"company_id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

// Employee represents Tryton company.employee bound to universal party.party
type Employee struct {
	ID           uuid.UUID  `json:"id"`
	PartyID      uuid.UUID  `json:"party_id"`
	PartyName    string     `json:"party_name"`
	CompanyID    uuid.UUID  `json:"company_id"`
	CompanyName  string     `json:"company_name"`
	EmployeeCode string     `json:"employee_code"`
	Designation  string     `json:"designation"`
	Department   string     `json:"department"`
	StartDate    string     `json:"start_date"`
	IsActive     bool       `json:"is_active"`
	BaseSalary   decimal.Decimal `json:"base_salary"`
}

// LeaveApplication represents employee leave request
type LeaveApplication struct {
	ID           uuid.UUID `json:"id"`
	CompanyID    uuid.UUID `json:"company_id"`
	EmployeeID   uuid.UUID `json:"employee_id"`
	EmployeeName string    `json:"employee_name"`
	LeaveType    string    `json:"leave_type"`
	FromDate     string    `json:"from_date"`
	ToDate       string    `json:"to_date"`
	TotalDays    int       `json:"total_days"`
	Reason       string    `json:"reason"`
	DocStatus    DocStatus `json:"docstatus"`
	ApprovedBy   *string   `json:"approved_by,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// SalarySlip represents monthly payroll with strict zero-float math and docstatus lifecycle
type SalarySlip struct {
	ID             uuid.UUID       `json:"id"`
	CompanyID      uuid.UUID       `json:"company_id"`
	SlipNumber     string          `json:"slip_number"`
	EmployeeID     uuid.UUID       `json:"employee_id"`
	EmployeeCode   string          `json:"employee_code"`
	EmployeeName   string          `json:"employee_name"`
	Designation    string          `json:"designation"`
	Department     string          `json:"department"`
	PostingDate    string          `json:"posting_date"`
	Month          int16           `json:"month"`
	Year           int             `json:"year"`
	
	// Earnings (Zero-Float)
	BaseSalary     decimal.Decimal `json:"base_salary"`
	HouseRent      decimal.Decimal `json:"house_rent"`
	Medical        decimal.Decimal `json:"medical"`
	Conveyance     decimal.Decimal `json:"conveyance"`
	GrossPay       decimal.Decimal `json:"gross_pay"`
	
	// Deductions (Zero-Float)
	ProvidentFund  decimal.Decimal `json:"provident_fund"`
	TaxDeduction   decimal.Decimal `json:"tax_deduction"`
	TotalDeduction decimal.Decimal `json:"total_deduction"`
	
	// Net Pay
	NetPay         decimal.Decimal `json:"net_pay"`
	
	DocStatus      DocStatus       `json:"docstatus"`
	JournalEntryID *uuid.UUID      `json:"journal_entry_id,omitempty"`
	VoucherNo      string          `json:"voucher_no,omitempty"`
	Remarks        string          `json:"remarks"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

type CreateSalarySlipRequest struct {
	CompanyID     uuid.UUID       `json:"company_id"`
	EmployeeID    uuid.UUID       `json:"employee_id"`
	PostingDate   string          `json:"posting_date"`
	Month         int16           `json:"month"`
	Year          int             `json:"year"`
	BaseSalary    decimal.Decimal `json:"base_salary"`
	HouseRent     decimal.Decimal `json:"house_rent"`
	Medical       decimal.Decimal `json:"medical"`
	Conveyance    decimal.Decimal `json:"conveyance"`
	ProvidentFund decimal.Decimal `json:"provident_fund"`
	TaxDeduction  decimal.Decimal `json:"tax_deduction"`
	Remarks       string          `json:"remarks"`
}

type CreateLeaveRequest struct {
	CompanyID  uuid.UUID `json:"company_id"`
	EmployeeID uuid.UUID `json:"employee_id"`
	LeaveType  string    `json:"leave_type"`
	FromDate   string    `json:"from_date"`
	ToDate     string    `json:"to_date"`
	TotalDays  int       `json:"total_days"`
	Reason     string    `json:"reason"`
}

// AttendanceRecord represents daily check-in/out and overtime (Horilla attendance app)
type AttendanceRecord struct {
	ID             uuid.UUID `json:"id"`
	CompanyID      uuid.UUID `json:"company_id"`
	EmployeeID     uuid.UUID `json:"employee_id"`
	EmployeeName   string    `json:"employee_name"`
	EmployeeCode   string    `json:"employee_code"`
	Department     string    `json:"department"`
	AttendanceDate string    `json:"attendance_date"`
	CheckIn        string    `json:"check_in"`
	CheckOut       *string   `json:"check_out,omitempty"`
	WorkingHours   string    `json:"working_hours"`
	OvertimeHours  string    `json:"overtime_hours"`
	Status         string    `json:"status"` // 'Present', 'Late', 'Half Day', 'Absent'
	Source         string    `json:"source"` // 'Biometric', 'Web', 'GPS'
}

type CheckInRequest struct {
	CompanyID  uuid.UUID `json:"company_id"`
	EmployeeID uuid.UUID `json:"employee_id"`
	Source     string    `json:"source"` // 'Biometric', 'Web', 'GPS'
}

// ReimbursementClaim represents employee expense claims (Horilla payroll.Reimbursement)
type ReimbursementClaim struct {
	ID             uuid.UUID       `json:"id"`
	CompanyID      uuid.UUID       `json:"company_id"`
	ClaimNumber    string          `json:"claim_number"`
	EmployeeID     uuid.UUID       `json:"employee_id"`
	EmployeeName   string          `json:"employee_name"`
	ClaimDate      string          `json:"claim_date"`
	Category       string          `json:"category"` // 'Travel', 'Medical', 'Client Entertainment', 'Office Supplies'
	Amount         decimal.Decimal `json:"amount"`
	Currency       string          `json:"currency"`
	Description    string          `json:"description"`
	ReceiptRef     string          `json:"receipt_ref,omitempty"`
	DocStatus      DocStatus       `json:"docstatus"` // 0=Draft, 1=Approved/GL Posted, 2=Cancelled
	JournalEntryID *uuid.UUID      `json:"journal_entry_id,omitempty"`
	VoucherNo      string          `json:"voucher_no,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}

type CreateClaimRequest struct {
	CompanyID   uuid.UUID       `json:"company_id"`
	EmployeeID  uuid.UUID       `json:"employee_id"`
	Category    string          `json:"category"`
	Amount      decimal.Decimal `json:"amount"`
	Description string          `json:"description"`
	ReceiptRef  string          `json:"receipt_ref"`
}

// ============================================================================
// Horilla Advanced Pillars: Loans, ATS Recruitment, and Assets
// ============================================================================

// Loan represents company loan or salary advance granted to employee
type Loan struct {
	ID                 uuid.UUID       `json:"id"`
	CompanyID          uuid.UUID       `json:"company_id"`
	EmployeeID         uuid.UUID       `json:"employee_id"`
	EmployeeName       string          `json:"employee_name"`
	EmployeeCode       string          `json:"employee_code"`
	Department         string          `json:"department"`
	LoanNumber         string          `json:"loan_number"`
	LoanType           string          `json:"loan_type"` // 'loan', 'advanced_salary', 'fine'
	Title              string          `json:"title"`
	PrincipalAmount    decimal.Decimal `json:"principal_amount"`
	MonthlyInstallment decimal.Decimal `json:"monthly_installment"`
	TotalInstallments  int             `json:"total_installments"`
	PaidInstallments   int             `json:"paid_installments"`
	DisbursementDate   string          `json:"disbursement_date"`
	StartDate          string          `json:"start_date"`
	Settled            bool            `json:"settled"`
	SettledDate        *string         `json:"settled_date,omitempty"`
	DocStatus          DocStatus       `json:"docstatus"` // 0=Draft, 1=Disbursed/Active, 2=Settled, 3=Cancelled
	JournalEntryID     *uuid.UUID      `json:"journal_entry_id,omitempty"`
	VoucherNo          string          `json:"voucher_no,omitempty"`
	Remarks            string          `json:"remarks"`
	CreatedAt          time.Time       `json:"created_at"`
}

type CreateLoanRequest struct {
	CompanyID         uuid.UUID       `json:"company_id"`
	EmployeeID        uuid.UUID       `json:"employee_id"`
	LoanType          string          `json:"loan_type"`
	Title             string          `json:"title"`
	PrincipalAmount   decimal.Decimal `json:"principal_amount"`
	TotalInstallments int             `json:"total_installments"`
	StartDate         string          `json:"start_date"`
	Remarks           string          `json:"remarks"`
}

// JobOpening represents a job requisition in Recruitment ATS
type JobOpening struct {
	ID              uuid.UUID `json:"id"`
	CompanyID       uuid.UUID `json:"company_id"`
	JobTitle        string    `json:"job_title"`
	Department      string    `json:"department"`
	Designation     string    `json:"designation"`
	Vacancies       int       `json:"vacancies"`
	ExperienceYears int       `json:"experience_years"`
	EmploymentType  string    `json:"employment_type"` // 'Full-Time', 'Part-Time', 'Contract'
	TargetDate      string    `json:"target_date"`
	Status          string    `json:"status"` // 'Active', 'Draft', 'Closed'
	Description     string    `json:"description"`
	ApplicantCount  int       `json:"applicant_count"`
	CreatedAt       time.Time `json:"created_at"`
}

type CreateJobOpeningRequest struct {
	CompanyID       uuid.UUID `json:"company_id"`
	JobTitle        string    `json:"job_title"`
	Department      string    `json:"department"`
	Designation     string    `json:"designation"`
	Vacancies       int       `json:"vacancies"`
	ExperienceYears int       `json:"experience_years"`
	EmploymentType  string    `json:"employment_type"`
	TargetDate      string    `json:"target_date"`
	Description     string    `json:"description"`
}

// Candidate represents an applicant in the ATS recruitment pipeline
type Candidate struct {
	ID                  uuid.UUID       `json:"id"`
	CompanyID           uuid.UUID       `json:"company_id"`
	JobOpeningID        uuid.UUID       `json:"job_opening_id"`
	JobTitle            string          `json:"job_title"`
	FullName            string          `json:"full_name"`
	Email               string          `json:"email"`
	Phone               string          `json:"phone"`
	Stage               string          `json:"stage"` // 'Applied', 'Screening', 'Interview', 'Offer', 'Hired', 'Rejected'
	Rating              int             `json:"rating"` // 1 to 5 stars
	ExpectedSalary      decimal.Decimal `json:"expected_salary"`
	CurrentCompany      string          `json:"current_company"`
	ResumeURL           string          `json:"resume_url"`
	Notes               string          `json:"notes"`
	ConvertedEmployeeID *uuid.UUID      `json:"converted_employee_id,omitempty"`
	CreatedAt           time.Time       `json:"created_at"`
}

type CreateCandidateRequest struct {
	CompanyID      uuid.UUID       `json:"company_id"`
	JobOpeningID   uuid.UUID       `json:"job_opening_id"`
	FullName       string          `json:"full_name"`
	Email          string          `json:"email"`
	Phone          string          `json:"phone"`
	ExpectedSalary decimal.Decimal `json:"expected_salary"`
	CurrentCompany string          `json:"current_company"`
	ResumeURL      string          `json:"resume_url"`
	Notes          string          `json:"notes"`
}

type UpdateCandidateStageRequest struct {
	Stage  string `json:"stage"`
	Rating int    `json:"rating"`
	Notes  string `json:"notes"`
}

// Asset represents company equipment (Horilla asset app)
type Asset struct {
	ID                  uuid.UUID       `json:"id"`
	CompanyID           uuid.UUID       `json:"company_id"`
	CategoryName        string          `json:"category_name"` // 'Laptops & Workstations', 'Monitors & Displays', 'Smartphones', 'Vehicles'
	AssetName           string          `json:"asset_name"`
	AssetCode           string          `json:"asset_code"`
	SerialNumber        string          `json:"serial_number"`
	PurchaseDate        string          `json:"purchase_date"`
	PurchaseCost        decimal.Decimal `json:"purchase_cost"`
	Status              string          `json:"status"` // 'Available', 'Allocated', 'Under Maintenance', 'Retired'
	WarrantyExpiry      string          `json:"warranty_expiry"`
	CurrentAssigneeID   *uuid.UUID      `json:"current_assignee_id,omitempty"`
	CurrentAssigneeName string          `json:"current_assignee_name,omitempty"`
}

type CreateAssetRequest struct {
	CompanyID      uuid.UUID       `json:"company_id"`
	CategoryName   string          `json:"category_name"`
	AssetName      string          `json:"asset_name"`
	AssetCode      string          `json:"asset_code"`
	SerialNumber   string          `json:"serial_number"`
	PurchaseDate   string          `json:"purchase_date"`
	PurchaseCost   decimal.Decimal `json:"purchase_cost"`
	WarrantyExpiry string          `json:"warranty_expiry"`
}

// AssetAllocation represents assignment of company equipment to an employee
type AssetAllocation struct {
	ID                uuid.UUID `json:"id"`
	AssetID           uuid.UUID `json:"asset_id"`
	AssetName         string    `json:"asset_name"`
	AssetCode         string    `json:"asset_code"`
	EmployeeID        uuid.UUID `json:"employee_id"`
	EmployeeName      string    `json:"employee_name"`
	EmployeeCode      string    `json:"employee_code"`
	AllocatedDate     string    `json:"allocated_date"`
	ReturnDueDate     string    `json:"return_due_date,omitempty"`
	ReturnedDate      *string   `json:"returned_date,omitempty"`
	ConditionOnAlloc  string    `json:"condition_on_alloc"`
	ConditionOnReturn string    `json:"condition_on_return,omitempty"`
	Status            string    `json:"status"` // 'Active', 'Returned', 'Damaged'
	Remarks           string    `json:"remarks"`
}

type AllocateAssetRequest struct {
	AssetID          uuid.UUID `json:"asset_id"`
	EmployeeID       uuid.UUID `json:"employee_id"`
	AllocatedDate    string    `json:"allocated_date"`
	ReturnDueDate    string    `json:"return_due_date"`
	ConditionOnAlloc string    `json:"condition_on_alloc"`
	Remarks          string    `json:"remarks"`
}

type ReturnAssetRequest struct {
	ReturnedDate      string `json:"returned_date"`
	ConditionOnReturn string `json:"condition_on_return"`
	Remarks           string `json:"remarks"`
}
