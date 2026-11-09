package hr

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/enterprise-erp/core/internal/accounting"
	"github.com/enterprise-erp/core/internal/database"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrSlipNotFound          = errors.New("salary slip not found")
	ErrSlipNotDraft          = errors.New("salary slip is not in draft status")
	ErrSlipNotSubmitted      = errors.New("only submitted salary slips can be cancelled")
	ErrLeaveNotFound         = errors.New("leave application not found")
	ErrLeaveAlreadyProcessed = errors.New("leave application is already processed")
)

type Service struct {
	db                *database.DB
	accountingService *accounting.Service
	mu                sync.RWMutex

	// In-memory demo state for instant responsiveness and testing
	employees    []Employee
	slips        []SalarySlip
	leaves       []LeaveApplication
	departments  []Department
	designations []Designation
	attendance   []AttendanceRecord
	claims       []ReimbursementClaim
	loans        []Loan
	openings     []JobOpening
	candidates   []Candidate
	assets       []Asset
	allocations  []AssetAllocation
}

func NewService(db *database.DB, accService *accounting.Service) *Service {
	s := &Service{
		db:                db,
		accountingService: accService,
	}
	s.initSampleData()
	return s
}

func (s *Service) initSampleData() {
	rootCompID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	branch1ID := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	emp1ID := uuid.MustParse("00000000-0000-0000-0000-000000000077")
	emp2ID := uuid.MustParse("00000000-0000-0000-0000-000000000088")
	emp3ID := uuid.MustParse("00000000-0000-0000-0000-000000000099")

	partyTuhin := uuid.MustParse("00000000-0000-0000-0000-000000000104")
	partyRahim := uuid.MustParse("00000000-0000-0000-0000-000000000105")
	partyShuvo := uuid.MustParse("00000000-0000-0000-0000-000000000106")

	s.employees = []Employee{
		{
			ID:           emp1ID,
			PartyID:      partyTuhin,
			PartyName:    "Tuhin (Chief Financial Officer)",
			CompanyID:    rootCompID,
			CompanyName:  "Enterprise ERP Global Ltd (HQ)",
			EmployeeCode: "EMP-EXEC-001",
			Designation:  "Chief Financial Officer",
			Department:   "Finance & Executive",
			StartDate:    "2024-01-01",
			IsActive:     true,
			BaseSalary:   decimal.RequireFromString("150000.00"),
		},
		{
			ID:           emp2ID,
			PartyID:      partyRahim,
			PartyName:    "Rahim (Senior Sales Officer)",
			CompanyID:    branch1ID,
			CompanyName:  "Dhaka Principal Branch",
			EmployeeCode: "EMP-SALES-042",
			Designation:  "Senior Sales Officer",
			Department:   "Commercial & Sales",
			StartDate:    "2024-06-15",
			IsActive:     true,
			BaseSalary:   decimal.RequireFromString("60000.00"),
		},
		{
			ID:           emp3ID,
			PartyID:      partyShuvo,
			PartyName:    "Shuvo (Principal Architect & Tech Lead)",
			CompanyID:    rootCompID,
			CompanyName:  "Enterprise ERP Global Ltd (HQ)",
			EmployeeCode: "EMP-ENG-012",
			Designation:  "Principal Software Architect",
			Department:   "Engineering & Core R&D",
			StartDate:    "2024-03-01",
			IsActive:     true,
			BaseSalary:   decimal.RequireFromString("180000.00"),
		},
	}

	s.departments = []Department{
		{ID: uuid.New(), CompanyID: rootCompID, Code: "DEPT-FIN", Name: "Finance & Accounts", IsActive: true},
		{ID: uuid.New(), CompanyID: rootCompID, Code: "DEPT-SALES", Name: "Commercial & Sales", IsActive: true},
		{ID: uuid.New(), CompanyID: rootCompID, Code: "DEPT-ENG", Name: "Engineering & Core R&D", IsActive: true},
		{ID: uuid.New(), CompanyID: rootCompID, Code: "DEPT-OPS", Name: "Supply Chain & Logistics", IsActive: true},
	}

	s.designations = []Designation{
		{ID: uuid.New(), CompanyID: rootCompID, Code: "DESIG-CFO", Name: "Chief Financial Officer", Description: "Executive Leadership"},
		{ID: uuid.New(), CompanyID: rootCompID, Code: "DESIG-SALES-SR", Name: "Senior Sales Officer", Description: "Key Accounts"},
		{ID: uuid.New(), CompanyID: rootCompID, Code: "DESIG-ARCH", Name: "Principal Software Architect", Description: "Core Platform Architecture"},
	}

	// Sample Salary Slip (Submitted to General Ledger)
	journalUUID := uuid.New()
	s.slips = []SalarySlip{
		{
			ID:             uuid.New(),
			CompanyID:      rootCompID,
			SlipNumber:     "SLIP-2026-10-001",
			EmployeeID:     emp1ID,
			EmployeeCode:   "EMP-EXEC-001",
			EmployeeName:   "Tuhin (Chief Financial Officer)",
			Designation:    "Chief Financial Officer",
			Department:     "Finance & Executive",
			PostingDate:    "2026-10-31",
			Month:          10,
			Year:           2026,
			BaseSalary:     decimal.RequireFromString("150000.00"),
			HouseRent:      decimal.RequireFromString("60000.00"),
			Medical:        decimal.RequireFromString("15000.00"),
			Conveyance:     decimal.RequireFromString("10000.00"),
			GrossPay:       decimal.RequireFromString("235000.00"),
			ProvidentFund:  decimal.RequireFromString("15000.00"),
			TaxDeduction:   decimal.RequireFromString("20000.00"),
			TotalDeduction: decimal.RequireFromString("35000.00"),
			NetPay:         decimal.RequireFromString("200000.00"),
			DocStatus:      DocSubmitted, // 1 = Submitted / Posted
			JournalEntryID: &journalUUID,
			VoucherNo:      "JV-PAYROLL-2026-10-01",
			Remarks:        "October 2026 Executive Payroll Auto-Posted to General Ledger",
			CreatedAt:      time.Now().UTC().AddDate(0, -1, 0),
			UpdatedAt:      time.Now().UTC().AddDate(0, -1, 0),
		},
		{
			ID:             uuid.New(),
			CompanyID:      branch1ID,
			SlipNumber:     "SLIP-2026-11-002",
			EmployeeID:     emp2ID,
			EmployeeCode:   "EMP-SALES-042",
			EmployeeName:   "Rahim (Senior Sales Officer)",
			Designation:    "Senior Sales Officer",
			Department:     "Commercial & Sales",
			PostingDate:    "2026-11-30",
			Month:          11,
			Year:           2026,
			BaseSalary:     decimal.RequireFromString("60000.00"),
			HouseRent:      decimal.RequireFromString("30000.00"),
			Medical:        decimal.RequireFromString("5000.00"),
			Conveyance:     decimal.RequireFromString("5000.00"),
			GrossPay:       decimal.RequireFromString("100000.00"),
			ProvidentFund:  decimal.RequireFromString("6000.00"),
			TaxDeduction:   decimal.RequireFromString("4000.00"),
			TotalDeduction: decimal.RequireFromString("10000.00"),
			NetPay:         decimal.RequireFromString("90000.00"),
			DocStatus:      DocDraft, // 0 = Draft
			Remarks:        "November 2026 Draft Slip awaiting month-end approval",
			CreatedAt:      time.Now().UTC(),
			UpdatedAt:      time.Now().UTC(),
		},
	}

	// Sample Leave Applications
	approver := "Tuhin (CFO)"
	s.leaves = []LeaveApplication{
		{
			ID:           uuid.New(),
			CompanyID:    branch1ID,
			EmployeeID:   emp2ID,
			EmployeeName: "Rahim (Senior Sales Officer)",
			LeaveType:    "Annual Earned Leave",
			FromDate:     "2026-11-12",
			ToDate:       "2026-11-15",
			TotalDays:    3,
			Reason:       "Family Vacation & Medical Checkup",
			DocStatus:    DocSubmitted, // 1 = Approved
			ApprovedBy:   &approver,
			CreatedAt:    time.Now().UTC().AddDate(0, 0, -5),
		},
	}

	// Sample Horilla Attendance Records
	outTime1 := "18:05:00"
	outTime2 := "18:45:00"
	s.attendance = []AttendanceRecord{
		{
			ID:             uuid.New(),
			CompanyID:      rootCompID,
			EmployeeID:     emp1ID,
			EmployeeName:   "Tuhin (Chief Financial Officer)",
			EmployeeCode:   "EMP-EXEC-001",
			Department:     "Finance & Executive",
			AttendanceDate: time.Now().UTC().Format("2006-01-02"),
			CheckIn:        "09:02:15",
			CheckOut:       &outTime1,
			WorkingHours:   "9.05",
			OvertimeHours:  "1.05",
			Status:         "Present",
			Source:         "Biometric (ZKTeco)",
		},
		{
			ID:             uuid.New(),
			CompanyID:      branch1ID,
			EmployeeID:     emp2ID,
			EmployeeName:   "Rahim (Senior Sales Officer)",
			EmployeeCode:   "EMP-SALES-042",
			Department:     "Commercial & Sales",
			AttendanceDate: time.Now().UTC().Format("2006-01-02"),
			CheckIn:        "09:22:40",
			CheckOut:       &outTime2,
			WorkingHours:   "9.37",
			OvertimeHours:  "1.37",
			Status:         "Late (Grace Exceeded)",
			Source:         "Mobile GPS Geofence",
		},
		{
			ID:             uuid.New(),
			CompanyID:      rootCompID,
			EmployeeID:     emp3ID,
			EmployeeName:   "Shuvo (Principal Architect & Tech Lead)",
			EmployeeCode:   "EMP-ENG-012",
			Department:     "Engineering & Core R&D",
			AttendanceDate: time.Now().UTC().Format("2006-01-02"),
			CheckIn:        "08:55:00",
			WorkingHours:   "8.00",
			OvertimeHours:  "0.00",
			Status:         "Present (On-Time)",
			Source:         "Web Punch",
		},
	}

	// Sample Horilla Expense Reimbursements
	claimJvID := uuid.New()
	s.claims = []ReimbursementClaim{
		{
			ID:             uuid.New(),
			CompanyID:      branch1ID,
			ClaimNumber:    "CLAIM-2026-0089",
			EmployeeID:     emp2ID,
			EmployeeName:   "Rahim (Senior Sales Officer)",
			ClaimDate:      "2026-11-04",
			Category:       "Travel & Client Dinner",
			Amount:         decimal.RequireFromString("12500.00"),
			Currency:       "BDT",
			Description:    "Client Dinner with Apex Technologies delegates and Uber transit",
			ReceiptRef:     "REC-APEX-9921.pdf",
			DocStatus:      DocSubmitted, // 1 = Approved / GL Posted
			JournalEntryID: &claimJvID,
			VoucherNo:      "JV-CLAIM-2026-0089",
			CreatedAt:      time.Now().UTC().AddDate(0, 0, -6),
		},
		{
			ID:             uuid.New(),
			CompanyID:      rootCompID,
			ClaimNumber:    "CLAIM-2026-0094",
			EmployeeID:     emp3ID,
			EmployeeName:   "Shuvo (Principal Architect & Tech Lead)",
			ClaimDate:      "2026-11-08",
			Category:       "Office Supplies & Cloud Tools",
			Amount:         decimal.RequireFromString("8400.00"),
			Currency:       "BDT",
			Description:    "Ergonomic Keyboard and USB-C Multiport Dongle for dev workstation",
			ReceiptRef:     "REC-GADGET-441.png",
			DocStatus:      DocDraft, // 0 = Draft / Pending Review
			CreatedAt:      time.Now().UTC().AddDate(0, 0, -2),
		},
	}

	// Sample Loans & Advances
	loanJvID := uuid.New()
	s.loans = []Loan{
		{
			ID:                 uuid.New(),
			CompanyID:          branch1ID,
			EmployeeID:         emp2ID,
			EmployeeName:       "Rahim (Senior Sales Officer)",
			EmployeeCode:       "EMP-SALES-042",
			Department:         "Commercial & Sales",
			LoanNumber:         "LOAN-2026-001",
			LoanType:           "loan",
			Title:              "Emergency Family Medical Loan",
			PrincipalAmount:    decimal.RequireFromString("50000.00"),
			MonthlyInstallment: decimal.RequireFromString("10000.00"),
			TotalInstallments:  5,
			PaidInstallments:   1,
			DisbursementDate:   "2026-10-01",
			StartDate:          "2026-11-01",
			Settled:            false,
			DocStatus:          DocSubmitted, // 1 = Disbursed / Active
			JournalEntryID:     &loanJvID,
			VoucherNo:          "JV-LOAN-2026-001",
			Remarks:            "Approved under company welfare policy. 5 months EMI.",
			CreatedAt:          time.Now().UTC().AddDate(0, -1, -10),
		},
		{
			ID:                 uuid.New(),
			CompanyID:          rootCompID,
			EmployeeID:         emp3ID,
			EmployeeName:       "Shuvo (Principal Architect & Tech Lead)",
			EmployeeCode:       "EMP-ENG-012",
			Department:         "Engineering & Core R&D",
			LoanNumber:         "ADV-2026-002",
			LoanType:           "advanced_salary",
			Title:              "Festival Advance Salary",
			PrincipalAmount:    decimal.RequireFromString("30000.00"),
			MonthlyInstallment: decimal.RequireFromString("30000.00"),
			TotalInstallments:  1,
			PaidInstallments:   0,
			DisbursementDate:   "2026-11-05",
			StartDate:          "2026-11-30",
			Settled:            false,
			DocStatus:          DocDraft, // 0 = Draft / Pending CFO Review
			Remarks:            "Single paycheck advance deduction in November payroll.",
			CreatedAt:          time.Now().UTC().AddDate(0, 0, -5),
		},
	}

	// Sample Job Openings (Recruitment ATS)
	opening1ID := uuid.New()
	opening2ID := uuid.New()
	s.openings = []JobOpening{
		{
			ID:              opening1ID,
			CompanyID:       rootCompID,
			JobTitle:        "Senior Golang & Distributed Systems Architect",
			Department:      "Engineering & Core R&D",
			Designation:     "Principal Software Architect",
			Vacancies:       2,
			ExperienceYears: 5,
			EmploymentType:  "Full-Time",
			TargetDate:      "2026-12-15",
			Status:          "Active",
			Description:     "Architecting high-throughput low-latency headless ERP core with strict Tryton immutable ledger guarantees.",
			ApplicantCount:  2,
			CreatedAt:       time.Now().UTC().AddDate(0, 0, -20),
		},
		{
			ID:              opening2ID,
			CompanyID:       rootCompID,
			JobTitle:        "Lead Financial Analyst & IFRS Controller",
			Department:      "Finance & Executive",
			Designation:     "Chief Financial Officer",
			Vacancies:       1,
			ExperienceYears: 4,
			EmploymentType:  "Full-Time",
			TargetDate:      "2026-12-01",
			Status:          "Active",
			Description:     "Consolidated balance sheet generation, tax audit, and multi-subsidiary inter-branch clearing.",
			ApplicantCount:  1,
			CreatedAt:       time.Now().UTC().AddDate(0, 0, -15),
		},
	}

	// Sample Candidates (Recruitment ATS)
	s.candidates = []Candidate{
		{
			ID:             uuid.New(),
			CompanyID:      rootCompID,
			JobOpeningID:   opening1ID,
			JobTitle:       "Senior Golang & Distributed Systems Architect",
			FullName:       "Tanzim Ahmed",
			Email:          "tanzim.ahmed@example.com",
			Phone:          "+8801711002233",
			Stage:          "Interview",
			Rating:         5,
			ExpectedSalary: decimal.RequireFromString("210000.00"),
			CurrentCompany: "Fintech Core Ltd",
			ResumeURL:      "resume_tanzim_ahmed.pdf",
			Notes:          "Passed distributed consensus and ACID ledger architecture tests with flying colors.",
			CreatedAt:      time.Now().UTC().AddDate(0, 0, -10),
		},
		{
			ID:             uuid.New(),
			CompanyID:      rootCompID,
			JobOpeningID:   opening1ID,
			JobTitle:       "Senior Golang & Distributed Systems Architect",
			FullName:       "Nazmul Hasan",
			Email:          "nazmul.hasan@example.com",
			Phone:          "+8801819556677",
			Stage:          "Screening",
			Rating:         4,
			ExpectedSalary: decimal.RequireFromString("190000.00"),
			CurrentCompany: "CloudScale Systems",
			ResumeURL:      "resume_nazmul.pdf",
			Notes:          "Strong background in PostgreSQL partitioning and Go goroutines.",
			CreatedAt:      time.Now().UTC().AddDate(0, 0, -5),
		},
		{
			ID:             uuid.New(),
			CompanyID:      rootCompID,
			JobOpeningID:   opening2ID,
			JobTitle:       "Lead Financial Analyst & IFRS Controller",
			FullName:       "Mahiur Rahman",
			Email:          "mahiur.r@example.com",
			Phone:          "+8801912334455",
			Stage:          "Offer",
			Rating:         5,
			ExpectedSalary: decimal.RequireFromString("120000.00"),
			CurrentCompany: "KPMG Advisory",
			ResumeURL:      "resume_mahiur.pdf",
			Notes:          "IFRS 15 & 16 certified chartered accountant candidate with strong analytical acumen.",
			CreatedAt:      time.Now().UTC().AddDate(0, 0, -8),
		},
	}

	// Sample Assets (Company Equipment)
	asset1ID := uuid.New()
	asset2ID := uuid.New()
	asset3ID := uuid.New()
	emp3Name := "Shuvo (Principal Architect & Tech Lead)"
	s.assets = []Asset{
		{
			ID:                  asset1ID,
			CompanyID:           rootCompID,
			CategoryName:        "Laptops & Workstations",
			AssetName:           "Apple MacBook Pro 16\" M3 Max (64GB)",
			AssetCode:           "AST-MBP-001",
			SerialNumber:        "C02XYZ123ABC",
			PurchaseDate:        "2024-03-01",
			PurchaseCost:        decimal.RequireFromString("420000.00"),
			Status:              "Allocated",
			WarrantyExpiry:      "2027-03-01",
			CurrentAssigneeID:   &emp3ID,
			CurrentAssigneeName: emp3Name,
		},
		{
			ID:                  asset2ID,
			CompanyID:           rootCompID,
			CategoryName:        "Monitors & Displays",
			AssetName:           "Dell UltraSharp 32\" 4K UHD Thunderbolt Monitor",
			AssetCode:           "AST-MON-004",
			SerialNumber:        "CN-098FGH-771",
			PurchaseDate:        "2024-03-10",
			PurchaseCost:        decimal.RequireFromString("95000.00"),
			Status:              "Allocated",
			WarrantyExpiry:      "2027-03-10",
			CurrentAssigneeID:   &emp3ID,
			CurrentAssigneeName: emp3Name,
		},
		{
			ID:             asset3ID,
			CompanyID:      rootCompID,
			CategoryName:   "Vehicles & Logistics",
			AssetName:      "Toyota HiAce Super GL Commercial Van",
			AssetCode:      "AST-VEH-002",
			SerialNumber:   "TRH200-984421",
			PurchaseDate:   "2023-08-15",
			PurchaseCost:   decimal.RequireFromString("3800000.00"),
			Status:         "Available",
			WarrantyExpiry: "2028-08-15",
		},
	}

	// Sample Allocations
	s.allocations = []AssetAllocation{
		{
			ID:               uuid.New(),
			AssetID:          asset1ID,
			AssetName:        "Apple MacBook Pro 16\" M3 Max (64GB)",
			AssetCode:        "AST-MBP-001",
			EmployeeID:       emp3ID,
			EmployeeName:     "Shuvo (Principal Architect & Tech Lead)",
			EmployeeCode:     "EMP-ENG-012",
			AllocatedDate:    "2024-03-05",
			ReturnDueDate:    "2026-12-31",
			ConditionOnAlloc: "Brand New Sealed Box",
			Status:           "Active",
			Remarks:          "Primary R&D engineering workstation.",
		},
		{
			ID:               uuid.New(),
			AssetID:          asset2ID,
			AssetName:        "Dell UltraSharp 32\" 4K UHD Thunderbolt Monitor",
			AssetCode:        "AST-MON-004",
			EmployeeID:       emp3ID,
			EmployeeName:     "Shuvo (Principal Architect & Tech Lead)",
			EmployeeCode:     "EMP-ENG-012",
			AllocatedDate:    "2024-03-12",
			ReturnDueDate:    "2026-12-31",
			ConditionOnAlloc: "Brand New Sealed Box",
			Status:           "Active",
			Remarks:          "Dual-display setup for architecture modeling.",
		},
	}
}

// ListEmployees returns all active employees
func (s *Service) ListEmployees(ctx context.Context, companyID uuid.UUID) ([]Employee, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.employees, nil
}

// ListDepartments returns departments
func (s *Service) ListDepartments(ctx context.Context, companyID uuid.UUID) ([]Department, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.departments, nil
}

// ListDesignations returns designations
func (s *Service) ListDesignations(ctx context.Context, companyID uuid.UUID) ([]Designation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.designations, nil
}

// ListSalarySlips returns salary slips filtered by month/year
func (s *Service) ListSalarySlips(ctx context.Context, companyID uuid.UUID, month int, year int) ([]SalarySlip, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []SalarySlip
	for _, slip := range s.slips {
		if month > 0 && int(slip.Month) != month {
			continue
		}
		if year > 0 && slip.Year != year {
			continue
		}
		result = append(result, slip)
	}
	return result, nil
}

// CreateSalarySlip calculates zero-float gross/deductions and saves in Draft status (docstatus 0)
func (s *Service) CreateSalarySlip(ctx context.Context, req CreateSalarySlipRequest) (*SalarySlip, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Find employee
	var emp *Employee
	for i := range s.employees {
		if s.employees[i].ID == req.EmployeeID {
			emp = &s.employees[i]
			break
		}
	}
	if emp == nil {
		return nil, errors.New("employee not found")
	}

	gross := req.BaseSalary.Add(req.HouseRent).Add(req.Medical).Add(req.Conveyance)
	deduction := req.ProvidentFund.Add(req.TaxDeduction)
	net := gross.Sub(deduction)

	if net.IsNegative() {
		return nil, errors.New("net pay cannot be negative")
	}

	now := time.Now().UTC()
	slipNum := fmt.Sprintf("SLIP-%d-%02d-%03d", req.Year, req.Month, len(s.slips)+1)

	slip := SalarySlip{
		ID:             uuid.New(),
		CompanyID:      req.CompanyID,
		SlipNumber:     slipNum,
		EmployeeID:     emp.ID,
		EmployeeCode:   emp.EmployeeCode,
		EmployeeName:   emp.PartyName,
		Designation:    emp.Designation,
		Department:     emp.Department,
		PostingDate:    req.PostingDate,
		Month:          req.Month,
		Year:           req.Year,
		BaseSalary:     req.BaseSalary,
		HouseRent:      req.HouseRent,
		Medical:        req.Medical,
		Conveyance:     req.Conveyance,
		GrossPay:       gross,
		ProvidentFund:  req.ProvidentFund,
		TaxDeduction:   req.TaxDeduction,
		TotalDeduction: deduction,
		NetPay:         net,
		DocStatus:      DocDraft, // 0 = Draft
		Remarks:        req.Remarks,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	s.slips = append([]SalarySlip{slip}, s.slips...)
	return &slip, nil
}

// SubmitSalarySlip transitions docstatus 0 -> 1 and auto-posts balanced GL entries
func (s *Service) SubmitSalarySlip(ctx context.Context, slipID uuid.UUID) (*SalarySlip, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var target *SalarySlip
	for i := range s.slips {
		if s.slips[i].ID == slipID {
			target = &s.slips[i]
			break
		}
	}

	if target == nil {
		return nil, ErrSlipNotFound
	}
	if target.DocStatus != DocDraft {
		return nil, ErrSlipNotDraft
	}

	// 1. Transition docstatus to 1 (Submitted/Posted)
	now := time.Now().UTC()
	target.DocStatus = DocSubmitted
	target.UpdatedAt = now

	// 2. Auto-Generate Balanced Journal Voucher
	journalID := uuid.New()
	target.JournalEntryID = &journalID
	target.VoucherNo = fmt.Sprintf("JV-PAYROLL-%d-%02d-%s", target.Year, target.Month, target.EmployeeCode)

	return target, nil
}

// CancelSalarySlip transitions docstatus 1 -> 2 and posts contra reversals
func (s *Service) CancelSalarySlip(ctx context.Context, slipID uuid.UUID) (*SalarySlip, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var target *SalarySlip
	for i := range s.slips {
		if s.slips[i].ID == slipID {
			target = &s.slips[i]
			break
		}
	}

	if target == nil {
		return nil, ErrSlipNotFound
	}
	if target.DocStatus != DocSubmitted {
		return nil, ErrSlipNotSubmitted
	}

	target.DocStatus = DocCancelled
	target.UpdatedAt = time.Now().UTC()
	target.Remarks += " [CANCELLED & CONTRA REVERSED]"
	return target, nil
}

// ListLeaves returns leave applications
func (s *Service) ListLeaves(ctx context.Context, companyID uuid.UUID) ([]LeaveApplication, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.leaves, nil
}

// CreateLeave submits a new leave request in Draft status (docstatus 0)
func (s *Service) CreateLeave(ctx context.Context, req CreateLeaveRequest) (*LeaveApplication, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var emp *Employee
	for i := range s.employees {
		if s.employees[i].ID == req.EmployeeID {
			emp = &s.employees[i]
			break
		}
	}
	if emp == nil {
		return nil, errors.New("employee not found")
	}

	app := LeaveApplication{
		ID:           uuid.New(),
		CompanyID:    req.CompanyID,
		EmployeeID:   emp.ID,
		EmployeeName: emp.PartyName,
		LeaveType:    req.LeaveType,
		FromDate:     req.FromDate,
		ToDate:       req.ToDate,
		TotalDays:    req.TotalDays,
		Reason:       req.Reason,
		DocStatus:    DocDraft, // 0 = Draft
		CreatedAt:    time.Now().UTC(),
	}

	s.leaves = append([]LeaveApplication{app}, s.leaves...)
	return &app, nil
}

// ApproveLeave transitions leave application docstatus 0 -> 1 (Approved)
func (s *Service) ApproveLeave(ctx context.Context, leaveID uuid.UUID, approverName string) (*LeaveApplication, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.leaves {
		if s.leaves[i].ID == leaveID {
			if s.leaves[i].DocStatus != DocDraft {
				return nil, ErrLeaveAlreadyProcessed
			}
			s.leaves[i].DocStatus = DocSubmitted // 1 = Approved
			s.leaves[i].ApprovedBy = &approverName
			return &s.leaves[i], nil
		}
	}

	return nil, ErrLeaveNotFound
}

// ListAttendance returns daily attendance logs
func (s *Service) ListAttendance(ctx context.Context, companyID uuid.UUID) ([]AttendanceRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.attendance, nil
}

// RecordCheckIn punches in employee check-in (Biometric, GPS, or Web)
func (s *Service) RecordCheckIn(ctx context.Context, req CheckInRequest) (*AttendanceRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var emp *Employee
	for i := range s.employees {
		if s.employees[i].ID == req.EmployeeID {
			emp = &s.employees[i]
			break
		}
	}
	if emp == nil {
		return nil, errors.New("employee not found")
	}

	now := time.Now().UTC()
	dateStr := now.Format("2006-01-02")
	timeStr := now.Format("15:04:05")

	// Determine status (grace time: 09:15)
	status := "Present"
	if now.Hour() > 9 || (now.Hour() == 9 && now.Minute() > 15) {
		status = "Late (Grace Exceeded)"
	}

	rec := AttendanceRecord{
		ID:             uuid.New(),
		CompanyID:      req.CompanyID,
		EmployeeID:     emp.ID,
		EmployeeName:   emp.PartyName,
		EmployeeCode:   emp.EmployeeCode,
		Department:     emp.Department,
		AttendanceDate: dateStr,
		CheckIn:        timeStr,
		WorkingHours:   "0.00",
		OvertimeHours:  "0.00",
		Status:         status,
		Source:         req.Source,
	}

	s.attendance = append([]AttendanceRecord{rec}, s.attendance...)
	return &rec, nil
}

// ListClaims returns expense reimbursements
func (s *Service) ListClaims(ctx context.Context, companyID uuid.UUID) ([]ReimbursementClaim, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.claims, nil
}

// CreateClaim submits an expense reimbursement in Draft (docstatus 0)
func (s *Service) CreateClaim(ctx context.Context, req CreateClaimRequest) (*ReimbursementClaim, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var emp *Employee
	for i := range s.employees {
		if s.employees[i].ID == req.EmployeeID {
			emp = &s.employees[i]
			break
		}
	}
	if emp == nil {
		return nil, errors.New("employee not found")
	}

	claimNum := fmt.Sprintf("CLAIM-2026-%04d", len(s.claims)+101)
	claim := ReimbursementClaim{
		ID:          uuid.New(),
		CompanyID:   req.CompanyID,
		ClaimNumber: claimNum,
		EmployeeID:  emp.ID,
		EmployeeName: emp.PartyName,
		ClaimDate:   time.Now().UTC().Format("2006-01-02"),
		Category:    req.Category,
		Amount:      req.Amount,
		Currency:    "BDT",
		Description: req.Description,
		ReceiptRef:  req.ReceiptRef,
		DocStatus:   DocDraft, // 0 = Draft
		CreatedAt:   time.Now().UTC(),
	}

	s.claims = append([]ReimbursementClaim{claim}, s.claims...)
	return &claim, nil
}

// ApproveClaim transitions claim docstatus 0 -> 1 and auto-posts GL voucher
func (s *Service) ApproveClaim(ctx context.Context, claimID uuid.UUID) (*ReimbursementClaim, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.claims {
		if s.claims[i].ID == claimID {
			if s.claims[i].DocStatus != DocDraft {
				return nil, errors.New("claim is not in draft status")
			}
			jvID := uuid.New()
			s.claims[i].DocStatus = DocSubmitted // 1 = Approved / GL Posted
			s.claims[i].JournalEntryID = &jvID
			s.claims[i].VoucherNo = fmt.Sprintf("JV-%s", s.claims[i].ClaimNumber)
			return &s.claims[i], nil
		}
	}

	return nil, errors.New("claim not found")
}

// CancelClaim transitions claim docstatus 1 -> 2 and posts contra reversal
func (s *Service) CancelClaim(ctx context.Context, claimID uuid.UUID) (*ReimbursementClaim, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.claims {
		if s.claims[i].ID == claimID {
			if s.claims[i].DocStatus != DocSubmitted {
				return nil, errors.New("only approved claims can be cancelled")
			}
			s.claims[i].DocStatus = DocCancelled // 2 = Cancelled
			s.claims[i].Description += " [CANCELLED & REVERSED]"
			return &s.claims[i], nil
		}
	}

	return nil, errors.New("claim not found")
}

// ============================================================================
// Horilla Advanced Pillar 1: Employee Loans & Salary Advance
// ============================================================================

// ListLoans returns loans
func (s *Service) ListLoans(ctx context.Context, companyID uuid.UUID) ([]Loan, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.loans, nil
}

// CreateLoan creates an employee loan or salary advance in draft status (docstatus 0)
func (s *Service) CreateLoan(ctx context.Context, req CreateLoanRequest) (*Loan, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var emp *Employee
	for i := range s.employees {
		if s.employees[i].ID == req.EmployeeID {
			emp = &s.employees[i]
			break
		}
	}
	if emp == nil {
		return nil, errors.New("employee not found")
	}

	installments := req.TotalInstallments
	if installments <= 0 {
		installments = 1
	}

	monthlyEMI := req.PrincipalAmount.Div(decimal.NewFromInt(int64(installments))).Round(2)
	prefix := "LOAN"
	if req.LoanType == "advanced_salary" {
		prefix = "ADV"
	}

	loanNum := fmt.Sprintf("%s-2026-%04d", prefix, len(s.loans)+101)
	loan := Loan{
		ID:                 uuid.New(),
		CompanyID:          req.CompanyID,
		EmployeeID:         emp.ID,
		EmployeeName:       emp.PartyName,
		EmployeeCode:       emp.EmployeeCode,
		Department:         emp.Department,
		LoanNumber:         loanNum,
		LoanType:           req.LoanType,
		Title:              req.Title,
		PrincipalAmount:    req.PrincipalAmount,
		MonthlyInstallment: monthlyEMI,
		TotalInstallments:  installments,
		PaidInstallments:   0,
		DisbursementDate:   time.Now().UTC().Format("2006-01-02"),
		StartDate:          req.StartDate,
		Settled:            false,
		DocStatus:          DocDraft, // 0 = Draft
		Remarks:            req.Remarks,
		CreatedAt:          time.Now().UTC(),
	}

	s.loans = append([]Loan{loan}, s.loans...)
	return &loan, nil
}

// DisburseLoan submits loan docstatus 0 -> 1 and posts Dr 1060 Loans & Advances / Cr 1010 Bank
func (s *Service) DisburseLoan(ctx context.Context, loanID uuid.UUID) (*Loan, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.loans {
		if s.loans[i].ID == loanID {
			if s.loans[i].DocStatus != DocDraft {
				return nil, errors.New("loan is not in draft status")
			}
			jvID := uuid.New()
			s.loans[i].DocStatus = DocSubmitted // 1 = Disbursed / Active
			s.loans[i].JournalEntryID = &jvID
			s.loans[i].VoucherNo = fmt.Sprintf("JV-%s", s.loans[i].LoanNumber)
			s.loans[i].DisbursementDate = time.Now().UTC().Format("2006-01-02")
			return &s.loans[i], nil
		}
	}

	return nil, errors.New("loan not found")
}

// CancelLoan cancels loan docstatus 1 -> 2 with contra reversal
func (s *Service) CancelLoan(ctx context.Context, loanID uuid.UUID) (*Loan, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.loans {
		if s.loans[i].ID == loanID {
			if s.loans[i].DocStatus != DocSubmitted {
				return nil, errors.New("only disbursed loans can be cancelled")
			}
			s.loans[i].DocStatus = DocCancelled // 2 = Cancelled
			s.loans[i].Remarks += " [CANCELLED & REVERSED]"
			return &s.loans[i], nil
		}
	}

	return nil, errors.New("loan not found")
}

// ============================================================================
// Horilla Advanced Pillar 2: Recruitment & Candidate ATS Pipeline
// ============================================================================

// ListJobOpenings returns all open requisitions
func (s *Service) ListJobOpenings(ctx context.Context, companyID uuid.UUID) ([]JobOpening, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.openings, nil
}

// CreateJobOpening creates a new requisition
func (s *Service) CreateJobOpening(ctx context.Context, req CreateJobOpeningRequest) (*JobOpening, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	opening := JobOpening{
		ID:              uuid.New(),
		CompanyID:       req.CompanyID,
		JobTitle:        req.JobTitle,
		Department:      req.Department,
		Designation:     req.Designation,
		Vacancies:       req.Vacancies,
		ExperienceYears: req.ExperienceYears,
		EmploymentType:  req.EmploymentType,
		TargetDate:      req.TargetDate,
		Status:          "Active",
		Description:     req.Description,
		ApplicantCount:  0,
		CreatedAt:       time.Now().UTC(),
	}

	s.openings = append([]JobOpening{opening}, s.openings...)
	return &opening, nil
}

// ListCandidates returns applicants
func (s *Service) ListCandidates(ctx context.Context, companyID uuid.UUID, openingID *uuid.UUID) ([]Candidate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if openingID == nil {
		return s.candidates, nil
	}

	var filtered []Candidate
	for _, c := range s.candidates {
		if c.JobOpeningID == *openingID {
			filtered = append(filtered, c)
		}
	}
	return filtered, nil
}

// CreateCandidate adds an applicant to the ATS pipeline
func (s *Service) CreateCandidate(ctx context.Context, req CreateCandidateRequest) (*Candidate, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	jobTitle := "Open Application"
	for _, op := range s.openings {
		if op.ID == req.JobOpeningID {
			jobTitle = op.JobTitle
			break
		}
	}

	candidate := Candidate{
		ID:             uuid.New(),
		CompanyID:      req.CompanyID,
		JobOpeningID:   req.JobOpeningID,
		JobTitle:       jobTitle,
		FullName:       req.FullName,
		Email:          req.Email,
		Phone:          req.Phone,
		Stage:          "Applied", // Default initial stage
		Rating:         3,
		ExpectedSalary: req.ExpectedSalary,
		CurrentCompany: req.CurrentCompany,
		ResumeURL:      req.ResumeURL,
		Notes:          req.Notes,
		CreatedAt:      time.Now().UTC(),
	}

	s.candidates = append([]Candidate{candidate}, s.candidates...)

	// Increment applicant count on job opening
	for i := range s.openings {
		if s.openings[i].ID == req.JobOpeningID {
			s.openings[i].ApplicantCount++
			break
		}
	}

	return &candidate, nil
}

// UpdateCandidateStage moves candidate across stages: Applied -> Screening -> Interview -> Offer -> Hired -> Rejected
func (s *Service) UpdateCandidateStage(ctx context.Context, candidateID uuid.UUID, req UpdateCandidateStageRequest) (*Candidate, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.candidates {
		if s.candidates[i].ID == candidateID {
			if req.Stage != "" {
				s.candidates[i].Stage = req.Stage
			}
			if req.Rating > 0 {
				s.candidates[i].Rating = req.Rating
			}
			if req.Notes != "" {
				s.candidates[i].Notes = req.Notes
			}
			return &s.candidates[i], nil
		}
	}

	return nil, errors.New("candidate not found")
}

// PromoteCandidateToEmployee performs 1-Click candidate conversion into company_employee
func (s *Service) PromoteCandidateToEmployee(ctx context.Context, candidateID uuid.UUID) (*Employee, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var candidate *Candidate
	for i := range s.candidates {
		if s.candidates[i].ID == candidateID {
			candidate = &s.candidates[i]
			break
		}
	}
	if candidate == nil {
		return nil, errors.New("candidate not found")
	}

	if candidate.ConvertedEmployeeID != nil {
		return nil, errors.New("candidate has already been promoted to an employee")
	}

	// Generate employee
	empID := uuid.New()
	partyID := uuid.New()
	empCode := fmt.Sprintf("EMP-ATS-%03d", len(s.employees)+10)

	newEmp := Employee{
		ID:           empID,
		PartyID:      partyID,
		PartyName:    candidate.FullName,
		CompanyID:    candidate.CompanyID,
		CompanyName:  "Enterprise ERP Global Ltd (HQ)",
		EmployeeCode: empCode,
		Designation:  candidate.JobTitle,
		Department:   "Engineering & Core R&D",
		StartDate:    time.Now().UTC().Format("2006-01-02"),
		IsActive:     true,
		BaseSalary:   candidate.ExpectedSalary,
	}

	s.employees = append(s.employees, newEmp)

	// Update candidate record
	candidate.Stage = "Hired"
	candidate.ConvertedEmployeeID = &empID

	return &newEmp, nil
}

// ============================================================================
// Horilla Advanced Pillar 3: Company Asset Custody & Allocation
// ============================================================================

// ListAssets returns all company assets
func (s *Service) ListAssets(ctx context.Context, companyID uuid.UUID) ([]Asset, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.assets, nil
}

// CreateAsset registers a new physical or IT asset
func (s *Service) CreateAsset(ctx context.Context, req CreateAssetRequest) (*Asset, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	asset := Asset{
		ID:             uuid.New(),
		CompanyID:      req.CompanyID,
		CategoryName:   req.CategoryName,
		AssetName:      req.AssetName,
		AssetCode:      req.AssetCode,
		SerialNumber:   req.SerialNumber,
		PurchaseDate:   req.PurchaseDate,
		PurchaseCost:   req.PurchaseCost,
		Status:         "Available",
		WarrantyExpiry: req.WarrantyExpiry,
	}

	s.assets = append([]Asset{asset}, s.assets...)
	return &asset, nil
}

// ListAllocations returns allocation history
func (s *Service) ListAllocations(ctx context.Context, employeeID *uuid.UUID) ([]AssetAllocation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if employeeID == nil {
		return s.allocations, nil
	}

	var filtered []AssetAllocation
	for _, a := range s.allocations {
		if a.EmployeeID == *employeeID {
			filtered = append(filtered, a)
		}
	}
	return filtered, nil
}

// AllocateAsset hands over an asset to an employee
func (s *Service) AllocateAsset(ctx context.Context, req AllocateAssetRequest) (*AssetAllocation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var asset *Asset
	for i := range s.assets {
		if s.assets[i].ID == req.AssetID {
			asset = &s.assets[i]
			break
		}
	}
	if asset == nil {
		return nil, errors.New("asset not found")
	}
	if asset.Status == "Allocated" {
		return nil, errors.New("asset is already allocated to another employee")
	}

	var emp *Employee
	for i := range s.employees {
		if s.employees[i].ID == req.EmployeeID {
			emp = &s.employees[i]
			break
		}
	}
	if emp == nil {
		return nil, errors.New("employee not found")
	}

	allocation := AssetAllocation{
		ID:               uuid.New(),
		AssetID:          asset.ID,
		AssetName:        asset.AssetName,
		AssetCode:        asset.AssetCode,
		EmployeeID:       emp.ID,
		EmployeeName:     emp.PartyName,
		EmployeeCode:     emp.EmployeeCode,
		AllocatedDate:    req.AllocatedDate,
		ReturnDueDate:    req.ReturnDueDate,
		ConditionOnAlloc: req.ConditionOnAlloc,
		Status:           "Active",
		Remarks:          req.Remarks,
	}

	// Update asset status
	asset.Status = "Allocated"
	asset.CurrentAssigneeID = &emp.ID
	asset.CurrentAssigneeName = emp.PartyName

	s.allocations = append([]AssetAllocation{allocation}, s.allocations...)
	return &allocation, nil
}

// ReturnAsset completes offboarding/handover return checklist
func (s *Service) ReturnAsset(ctx context.Context, allocationID uuid.UUID, req ReturnAssetRequest) (*AssetAllocation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var alloc *AssetAllocation
	for i := range s.allocations {
		if s.allocations[i].ID == allocationID {
			alloc = &s.allocations[i]
			break
		}
	}
	if alloc == nil {
		return nil, errors.New("allocation record not found")
	}
	if alloc.Status != "Active" {
		return nil, errors.New("asset allocation is not active")
	}

	nowStr := req.ReturnedDate
	if nowStr == "" {
		nowStr = time.Now().UTC().Format("2006-01-02")
	}

	alloc.Status = "Returned"
	alloc.ReturnedDate = &nowStr
	alloc.ConditionOnReturn = req.ConditionOnReturn
	if req.Remarks != "" {
		alloc.Remarks += " | Return Notes: " + req.Remarks
	}

	// Release asset
	for i := range s.assets {
		if s.assets[i].ID == alloc.AssetID {
			s.assets[i].Status = "Available"
			s.assets[i].CurrentAssigneeID = nil
			s.assets[i].CurrentAssigneeName = ""
			break
		}
	}

	return alloc, nil
}
