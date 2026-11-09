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
