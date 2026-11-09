package accounting

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type AccountRootType string

const (
	RootAsset     AccountRootType = "Asset"
	RootLiability AccountRootType = "Liability"
	RootEquity    AccountRootType = "Equity"
	RootIncome    AccountRootType = "Income"
	RootExpense   AccountRootType = "Expense"
)

type DocStatus string

const (
	StatusDraft     DocStatus = "draft"
	StatusPosted    DocStatus = "posted"
	StatusCancelled DocStatus = "cancelled"
)

type Company struct {
	ID                    uuid.UUID `json:"id"`
	Name                  string    `json:"name"`
	BaseCurrency          string    `json:"base_currency"`
	Country               string    `json:"country"`
	FiscalYearStartMonth  int16     `json:"fiscal_year_start_month"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type Account struct {
	ID         uuid.UUID       `json:"id"`
	CompanyID  uuid.UUID       `json:"company_id"`
	Code       string          `json:"code"`
	Name       string          `json:"name"`
	RootType   AccountRootType `json:"root_type"`
	ParentID   *uuid.UUID      `json:"parent_id,omitempty"`
	IsGroup    bool            `json:"is_group"`
	Currency   string          `json:"currency"`
	IsActive   bool            `json:"is_active"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

type JournalEntry struct {
	ID          uuid.UUID       `json:"id"`
	CompanyID   uuid.UUID       `json:"company_id"`
	EntryNumber string          `json:"entry_number"`
	PostingDate time.Time       `json:"posting_date"`
	Status      DocStatus       `json:"status"`
	TotalDebit  decimal.Decimal `json:"total_debit"`
	TotalCredit decimal.Decimal `json:"total_credit"`
	Remarks     string          `json:"remarks"`
	PostedAt    *time.Time      `json:"posted_at,omitempty"`
	PostedBy    *uuid.UUID      `json:"posted_by,omitempty"`
	Lines       []GLEntry       `json:"lines,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// GLEntry represents an append-only row in the general ledger (gl_entries table).
// Strictly uses decimal.Decimal for all monetary values.
type GLEntry struct {
	ID           uuid.UUID       `json:"id"`
	CompanyID    uuid.UUID       `json:"company_id"`
	PostingDate  time.Time       `json:"posting_date"`
	VoucherType  string          `json:"voucher_type"`
	VoucherID    uuid.UUID       `json:"voucher_id"`
	VoucherNo    string          `json:"voucher_no"`
	AccountID    uuid.UUID       `json:"account_id"`
	AccountCode  string          `json:"account_code"`
	PartyType    *string         `json:"party_type,omitempty"`
	PartyID      *uuid.UUID      `json:"party_id,omitempty"`
	Currency     string          `json:"currency"`
	Debit        decimal.Decimal `json:"debit"`
	Credit       decimal.Decimal `json:"credit"`
	ExchangeRate decimal.Decimal `json:"exchange_rate"`
	DebitBase    decimal.Decimal `json:"debit_base"`
	CreditBase   decimal.Decimal `json:"credit_base"`
	Remarks      string          `json:"remarks"`
	IsCancelled  bool            `json:"is_cancelled"`
	CreatedAt    time.Time       `json:"created_at"`
}

// Request DTOs
type CreateCompanyRequest struct {
	Name         string `json:"name"`
	BaseCurrency string `json:"base_currency"`
	Country      string `json:"country"`
}

type CreateAccountRequest struct {
	CompanyID uuid.UUID       `json:"company_id"`
	Code      string          `json:"code"`
	Name      string          `json:"name"`
	RootType  AccountRootType `json:"root_type"`
	ParentID  *uuid.UUID      `json:"parent_id,omitempty"`
	IsGroup   bool            `json:"is_group"`
	Currency  string          `json:"currency"`
}

type JournalLineInput struct {
	AccountID    uuid.UUID       `json:"account_id"`
	Currency     string          `json:"currency"`
	Debit        decimal.Decimal `json:"debit"`
	Credit       decimal.Decimal `json:"credit"`
	ExchangeRate decimal.Decimal `json:"exchange_rate"`
	PartyType    *string         `json:"party_type,omitempty"`
	PartyID      *uuid.UUID      `json:"party_id,omitempty"`
	Remarks      string          `json:"remarks"`
}

type CreateJournalEntryRequest struct {
	CompanyID   uuid.UUID          `json:"company_id"`
	PostingDate string             `json:"posting_date"` // YYYY-MM-DD
	Remarks     string             `json:"remarks"`
	Lines       []JournalLineInput `json:"lines"`
}

// Trial Balance Reporting (Consolidated vs Standalone Branch)
type TrialBalanceRow struct {
	AccountID   uuid.UUID       `json:"account_id"`
	AccountCode string          `json:"account_code"`
	AccountName string          `json:"account_name"`
	RootType    AccountRootType `json:"root_type"`
	TotalDebit  decimal.Decimal `json:"total_debit"`
	TotalCredit decimal.Decimal `json:"total_credit"`
	NetDebit    decimal.Decimal `json:"net_debit"`
	NetCredit   decimal.Decimal `json:"net_credit"`
}

type TrialBalanceReport struct {
	CompanyID        uuid.UUID         `json:"company_id"`
	CompanyName      string            `json:"company_name"`
	IsConsolidated   bool              `json:"is_consolidated"`
	IncludedBranches []string          `json:"included_branches"`
	TotalDebit       decimal.Decimal   `json:"total_debit"`
	TotalCredit      decimal.Decimal   `json:"total_credit"`
	Difference       decimal.Decimal   `json:"difference"`
	IsBalanced       bool              `json:"is_balanced"`
	Lines            []TrialBalanceRow `json:"lines"`
}

// Inter-Branch Transactions (Due To / Due From Clearing Engine)
type InterBranchTransferRequest struct {
	FromCompanyID  uuid.UUID       `json:"from_company_id"`
	ToCompanyID    uuid.UUID       `json:"to_company_id"`
	AssetAccountID uuid.UUID       `json:"asset_account_id"`
	Amount         decimal.Decimal `json:"amount"`
	Currency       string          `json:"currency"`
	Remarks        string          `json:"remarks"`
}

type InterBranchTransferResult struct {
	SourceVoucherNo      string          `json:"source_voucher_no"`
	DestinationVoucherNo string          `json:"destination_voucher_no"`
	PostingDate          time.Time       `json:"posting_date"`
	Amount               decimal.Decimal `json:"amount"`
	Currency             string          `json:"currency"`
	IsBalanced           bool            `json:"is_balanced"`
	Status               string          `json:"status"`
}

// Party Ledger / Sub-Ledger Statement (Customer Aging & Supplier Statement)
type PartyLedgerLine struct {
	PostingDate time.Time       `json:"posting_date"`
	VoucherType string          `json:"voucher_type"`
	VoucherNo   string          `json:"voucher_no"`
	AccountCode string          `json:"account_code"`
	AccountName string          `json:"account_name"`
	BranchName  string          `json:"branch_name"`
	Debit       decimal.Decimal `json:"debit"`
	Credit      decimal.Decimal `json:"credit"`
	Balance     decimal.Decimal `json:"balance"` // Running Cumulative Balance
	Remarks     string          `json:"remarks"`
}

type PartyLedgerReport struct {
	PartyID        uuid.UUID         `json:"party_id"`
	PartyCode      string            `json:"party_code"`
	PartyName      string            `json:"party_name"`
	CompanyID      uuid.UUID         `json:"company_id"`
	CompanyName    string            `json:"company_name"`
	IsConsolidated bool              `json:"is_consolidated"`
	OpeningBalance decimal.Decimal   `json:"opening_balance"`
	TotalDebit     decimal.Decimal   `json:"total_debit"`
	TotalCredit    decimal.Decimal   `json:"total_credit"`
	ClosingBalance decimal.Decimal   `json:"closing_balance"`
	Lines          []PartyLedgerLine `json:"lines"`
}
