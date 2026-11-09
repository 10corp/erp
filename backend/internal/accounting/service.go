package accounting

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/enterprise-erp/core/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

var (
	ErrUnbalancedEntry       = errors.New("unbalanced transaction: total debit base must equal total credit base")
	ErrEmptyJournalLines     = errors.New("journal entry must have at least two line items")
	ErrInvalidDocStatus      = errors.New("invalid document state transition")
	ErrAccountNotFound       = errors.New("account not found")
	ErrLineDebitCreditInvalid= errors.New("each line must have either debit or credit, not both or neither")
)

type Service struct {
	db *database.DB
}

func NewService(db *database.DB) *Service {
	return &Service{db: db}
}

// CreateAccount adds a new ledger account to the chart of accounts
func (s *Service) CreateAccount(ctx context.Context, req CreateAccountRequest) (*Account, error) {
	if req.Currency == "" {
		req.Currency = "BDT"
	}

	account := &Account{
		ID:        uuid.New(),
		CompanyID: req.CompanyID,
		Code:      req.Code,
		Name:      req.Name,
		RootType:  req.RootType,
		ParentID:  req.ParentID,
		IsGroup:   req.IsGroup,
		Currency:  req.Currency,
		IsActive:  true,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	query := `
		INSERT INTO accounts (id, company_id, code, name, root_type, parent_id, is_group, currency, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, company_id, code, name, root_type, parent_id, is_group, currency, is_active, created_at, updated_at;
	`

	err := s.db.Pool.QueryRow(ctx, query,
		account.ID, account.CompanyID, account.Code, account.Name,
		account.RootType, account.ParentID, account.IsGroup,
		account.Currency, account.IsActive, account.CreatedAt, account.UpdatedAt,
	).Scan(
		&account.ID, &account.CompanyID, &account.Code, &account.Name,
		&account.RootType, &account.ParentID, &account.IsGroup,
		&account.Currency, &account.IsActive, &account.CreatedAt, &account.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert account: %w", err)
	}

	return account, nil
}

// ListAccounts lists all accounts for a given company
func (s *Service) ListAccounts(ctx context.Context, companyID uuid.UUID) ([]Account, error) {
	query := `
		SELECT id, company_id, code, name, root_type, parent_id, is_group, currency, is_active, created_at, updated_at
		FROM accounts
		WHERE company_id = $1 AND is_active = true
		ORDER BY code ASC;
	`
	rows, err := s.db.Pool.Query(ctx, query, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to query accounts: %w", err)
	}
	defer rows.Close()

	var accounts []Account
	for rows.Next() {
		var a Account
		err := rows.Scan(
			&a.ID, &a.CompanyID, &a.Code, &a.Name,
			&a.RootType, &a.ParentID, &a.IsGroup,
			&a.Currency, &a.IsActive, &a.CreatedAt, &a.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, nil
}

// PostJournalEntry performs strict Tryton-style financial consistency checks
// and appends entries into the partitioned gl_entries table inside a single atomic transaction.
func (s *Service) PostJournalEntry(ctx context.Context, req CreateJournalEntryRequest) (*JournalEntry, error) {
	if len(req.Lines) < 2 {
		return nil, ErrEmptyJournalLines
	}

	postingDate, err := time.Parse("2006-01-02", req.PostingDate)
	if err != nil {
		return nil, fmt.Errorf("invalid posting date format (expected YYYY-MM-DD): %w", err)
	}

	// Begin atomic PostgreSQL transaction
	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Fetch company base currency
	var baseCurrency string
	err = tx.QueryRow(ctx, "SELECT base_currency FROM companies WHERE id = $1", req.CompanyID).Scan(&baseCurrency)
	if err != nil {
		return nil, fmt.Errorf("company not found: %w", err)
	}

	entryID := uuid.New()
	entryNumber := fmt.Sprintf("JV-%s-%s", postingDate.Format("20060102"), entryID.String()[:8])

	totalDebitBase := decimal.Zero
	totalCreditBase := decimal.Zero

	type PreparedGLLine struct {
		AccountID    uuid.UUID
		AccountCode  string
		Currency     string
		Debit        decimal.Decimal
		Credit       decimal.Decimal
		ExchangeRate decimal.Decimal
		DebitBase    decimal.Decimal
		CreditBase   decimal.Decimal
		PartyType    *string
		PartyID      *uuid.UUID
		Remarks      string
	}

	var preparedLines []PreparedGLLine

	for idx, line := range req.Lines {
		if (line.Debit.IsPositive() && line.Credit.IsPositive()) ||
			(line.Debit.IsZero() && line.Credit.IsZero()) {
			return nil, fmt.Errorf("line %d: %w", idx+1, ErrLineDebitCreditInvalid)
		}

		rate := line.ExchangeRate
		if rate.IsZero() || rate.IsNegative() {
			rate = decimal.NewFromInt(1)
		}

		// Calculate base amounts strictly with decimal math
		debitBase := line.Debit.Mul(rate)
		creditBase := line.Credit.Mul(rate)

		totalDebitBase = totalDebitBase.Add(debitBase)
		totalCreditBase = totalCreditBase.Add(creditBase)

		// Fetch account code
		var accountCode string
		err := tx.QueryRow(ctx, "SELECT code FROM accounts WHERE id = $1 AND company_id = $2", line.AccountID, req.CompanyID).Scan(&accountCode)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, fmt.Errorf("line %d: account %s not found in company", idx+1, line.AccountID)
			}
			return nil, err
		}

		preparedLines = append(preparedLines, PreparedGLLine{
			AccountID:    line.AccountID,
			AccountCode:  accountCode,
			Currency:     line.Currency,
			Debit:        line.Debit,
			Credit:       line.Credit,
			ExchangeRate: rate,
			DebitBase:    debitBase,
			CreditBase:   creditBase,
			PartyType:    line.PartyType,
			PartyID:      line.PartyID,
			Remarks:      line.Remarks,
		})
	}

	// Double-entry validation: SUM(Debit Base) == SUM(Credit Base)
	if !totalDebitBase.Equal(totalCreditBase) {
		return nil, fmt.Errorf("%w: debit=%s, credit=%s, diff=%s",
			ErrUnbalancedEntry, totalDebitBase.String(), totalCreditBase.String(), totalDebitBase.Sub(totalCreditBase).String())
	}

	now := time.Now().UTC()

	// 1. Insert Journal Entry Header
	insertJEQuery := `
		INSERT INTO journal_entries (id, company_id, entry_number, posting_date, status, total_debit, total_credit, remarks, posted_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);
	`
	_, err = tx.Exec(ctx, insertJEQuery,
		entryID, req.CompanyID, entryNumber, postingDate, StatusPosted,
		totalDebitBase, totalCreditBase, req.Remarks, now, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert journal entry header: %w", err)
	}

	// 2. Append lines to Partitioned gl_entries table
	insertGLQuery := `
		INSERT INTO gl_entries (
			id, company_id, posting_date, voucher_type, voucher_id, voucher_no,
			account_id, account_code, party_type, party_id, currency,
			debit, credit, exchange_rate, debit_base, credit_base, remarks, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18);
	`

	var insertedGLLines []GLEntry
	for _, l := range preparedLines {
		glID := uuid.New()
		_, err = tx.Exec(ctx, insertGLQuery,
			glID, req.CompanyID, postingDate, "Journal Entry", entryID, entryNumber,
			l.AccountID, l.AccountCode, l.PartyType, l.PartyID, l.Currency,
			l.Debit, l.Credit, l.ExchangeRate, l.DebitBase, l.CreditBase, l.Remarks, now,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to insert gl_entry: %w", err)
		}

		insertedGLLines = append(insertedGLLines, GLEntry{
			ID:           glID,
			CompanyID:    req.CompanyID,
			PostingDate:  postingDate,
			VoucherType:  "Journal Entry",
			VoucherID:    entryID,
			VoucherNo:    entryNumber,
			AccountID:    l.AccountID,
			AccountCode:  l.AccountCode,
			PartyType:    l.PartyType,
			PartyID:      l.PartyID,
			Currency:     l.Currency,
			Debit:        l.Debit,
			Credit:       l.Credit,
			ExchangeRate: l.ExchangeRate,
			DebitBase:    l.DebitBase,
			CreditBase:   l.CreditBase,
			Remarks:      l.Remarks,
			CreatedAt:    now,
		})
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &JournalEntry{
		ID:          entryID,
		CompanyID:   req.CompanyID,
		EntryNumber: entryNumber,
		PostingDate: postingDate,
		Status:      StatusPosted,
		TotalDebit:  totalDebitBase,
		TotalCredit: totalCreditBase,
		Remarks:     req.Remarks,
		PostedAt:    &now,
		Lines:       insertedGLLines,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// GetLedgerEntries retrieves paginated ledger entries from the partitioned gl_entries table
func (s *Service) GetLedgerEntries(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]GLEntry, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	query := `
		SELECT id, company_id, posting_date, voucher_type, voucher_id, voucher_no,
		       account_id, account_code, party_type, party_id, currency,
		       debit, credit, exchange_rate, debit_base, credit_base, remarks, is_cancelled, created_at
		FROM gl_entries
		WHERE company_id = $1
		ORDER BY posting_date DESC, created_at DESC
		LIMIT $2 OFFSET $3;
	`

	rows, err := s.db.Pool.Query(ctx, query, companyID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query gl_entries: %w", err)
	}
	defer rows.Close()

	var entries []GLEntry
	for rows.Next() {
		var g GLEntry
		err := rows.Scan(
			&g.ID, &g.CompanyID, &g.PostingDate, &g.VoucherType, &g.VoucherID, &g.VoucherNo,
			&g.AccountID, &g.AccountCode, &g.PartyType, &g.PartyID, &g.Currency,
			&g.Debit, &g.Credit, &g.ExchangeRate, &g.DebitBase, &g.CreditBase,
			&g.Remarks, &g.IsCancelled, &g.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		entries = append(entries, g)
	}
	return entries, nil
}
