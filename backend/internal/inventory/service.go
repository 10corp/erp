package inventory

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/enterprise-erp/core/internal/accounting"
	"github.com/enterprise-erp/core/internal/database"
	"github.com/enterprise-erp/core/internal/doctype"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

var (
	ErrInsufficientStock     = errors.New("negative stock violation: insufficient quantity in warehouse")
	ErrInvalidWarehouse      = errors.New("source and target double-entry warehouses are required")
	ErrZeroDeliveryQty       = errors.New("delivered quantity must be greater than zero")
	ErrUnbalancedInventoryGL = errors.New("Tryton Invariant: perpetual COGS GL entries are unbalanced")
)

type Service struct {
	db *database.DB
}

func NewService(db *database.DB) *Service {
	return &Service{db: db}
}

type ReceiveStockInput struct {
	CompanyID         uuid.UUID       `json:"company_id"`
	ItemCode          string          `json:"item_code"`
	ItemName          string          `json:"item_name"`
	FromWarehouseID   uuid.UUID       `json:"from_warehouse_id"` // Supplier Virtual Location
	ToWarehouseID     uuid.UUID       `json:"to_warehouse_id"`   // Internal Physical Warehouse
	Qty               decimal.Decimal `json:"qty"`
	Rate              decimal.Decimal `json:"rate"` // Purchase price per unit
	VoucherType       string          `json:"voucher_type"`
	VoucherID         uuid.UUID       `json:"voucher_id"`
	VoucherNo         string          `json:"voucher_no"`
	Remarks           string          `json:"remarks"`
}

// ReceiveStock processes incoming inventory (Purchase Receipt) and updates Moving Average Valuation Rate atomically
func (s *Service) ReceiveStock(ctx context.Context, input ReceiveStockInput) (*StockLedgerEntry, error) {
	if input.Qty.IsZero() || input.Qty.IsNegative() {
		return nil, errors.New("incoming quantity must be greater than zero")
	}

	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Concurrency Guard: Lock item-warehouse balance row
	var currentQty, currentRate, currentValue decimal.Decimal
	queryLock := `
		SELECT current_qty, valuation_rate, stock_value
		FROM item_warehouse_balances
		WHERE company_id = $1 AND item_code = $2 AND warehouse_id = $3
		FOR UPDATE;
	`
	err = tx.QueryRow(ctx, queryLock, input.CompanyID, input.ItemCode, input.ToWarehouseID).Scan(&currentQty, &currentRate, &currentValue)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// First receipt for this item in warehouse
			currentQty = decimal.Zero
			currentRate = decimal.Zero
			currentValue = decimal.Zero

			insertBal := `
				INSERT INTO item_warehouse_balances (id, company_id, item_code, warehouse_id, current_qty, valuation_rate, stock_value, updated_at)
				VALUES ($1, $2, $3, $4, 0, 0, 0, NOW());
			`
			_, err = tx.Exec(ctx, insertBal, uuid.New(), input.CompanyID, input.ItemCode, input.ToWarehouseID)
			if err != nil {
				return nil, fmt.Errorf("failed to initialize item balance: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to lock balance row: %w", err)
		}
	}

	// Calculate New Moving Average Valuation Rate ($O(1)$)
	newQty, newRate, newValue := CalculateMovingAverage(currentQty, currentRate, input.Qty, input.Rate)
	stockValueDiff := input.Qty.Mul(input.Rate)
	now := time.Now().UTC()

	// Update item balance table
	updateBal := `
		UPDATE item_warehouse_balances
		SET current_qty = $1, valuation_rate = $2, stock_value = $3, updated_at = $4
		WHERE company_id = $5 AND item_code = $6 AND warehouse_id = $7;
	`
	_, err = tx.Exec(ctx, updateBal, newQty, newRate, newValue, now, input.CompanyID, input.ItemCode, input.ToWarehouseID)
	if err != nil {
		return nil, fmt.Errorf("failed to update balance: %w", err)
	}

	// Append row to partitioned stock_ledger_entries table
	sleID := uuid.New()
	insertSLE := `
		INSERT INTO stock_ledger_entries (
			id, company_id, posting_datetime, item_code, item_name,
			from_warehouse_id, to_warehouse_id, voucher_type, voucher_id, voucher_no,
			actual_qty, valuation_rate, stock_value_difference,
			qty_after_transaction, stock_value_after_transaction, remarks, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17);
	`
	_, err = tx.Exec(ctx, insertSLE,
		sleID, input.CompanyID, now, input.ItemCode, input.ItemName,
		input.FromWarehouseID, input.ToWarehouseID, input.VoucherType, input.VoucherID, input.VoucherNo,
		input.Qty, newRate, stockValueDiff, newQty, newValue, input.Remarks, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to append stock ledger entry: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit failed: %w", err)
	}

	return &StockLedgerEntry{
		ID:                         sleID,
		CompanyID:                  input.CompanyID,
		PostingDateTime:            now,
		ItemCode:                   input.ItemCode,
		ItemName:                   input.ItemName,
		FromWarehouseID:            input.FromWarehouseID,
		ToWarehouseID:              input.ToWarehouseID,
		VoucherType:                input.VoucherType,
		VoucherID:                  input.VoucherID,
		VoucherNo:                  input.VoucherNo,
		ActualQty:                  input.Qty,
		ValuationRate:              newRate,
		StockValueDifference:       stockValueDiff,
		QtyAfterTransaction:        newQty,
		StockValueAfterTransaction: newValue,
		Remarks:                    input.Remarks,
		CreatedAt:                  now,
	}, nil
}

type DeliverItemInput struct {
	ItemCode string          `json:"item_code"`
	ItemName string          `json:"item_name"`
	Qty      decimal.Decimal `json:"qty"`
}

type DeliverStockRequest struct {
	CompanyID           uuid.UUID          `json:"company_id"`
	CustomerName        string             `json:"customer_name"`
	PostingDate         time.Time          `json:"posting_date"`
	FromWarehouseID     uuid.UUID          `json:"from_warehouse_id"`     // Internal Warehouse
	ToWarehouseID       uuid.UUID          `json:"to_warehouse_id"`       // Customer Virtual Location
	COGSAccountID       uuid.UUID          `json:"cogs_account_id"`       // 5010 - Cost of Goods Sold
	StockAssetAccountID uuid.UUID          `json:"stock_asset_account_id"`// 1030 - Stock In Hand
	COGSAccountCode     string             `json:"cogs_account_code"`     // e.g. "5010"
	StockAssetCode      string             `json:"stock_asset_code"`      // e.g. "1030"
	Items               []DeliverItemInput `json:"items"`
	Remarks             string             `json:"remarks"`
}

// DeliverStock handles delivery note submission, moving average COGS calculation,
// strict negative stock prevention (SELECT FOR UPDATE), and Perpetual COGS posting into gl_entries.
func (s *Service) DeliverStock(ctx context.Context, req DeliverStockRequest) (*DeliveryNote, []accounting.GLEntry, error) {
	if len(req.Items) == 0 {
		return nil, nil, errors.New("delivery note requires at least one item")
	}

	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	deliveryID := uuid.New()
	deliveryNumber := fmt.Sprintf("DN-%s-%s", req.PostingDate.Format("2006"), deliveryID.String()[:8])
	now := time.Now().UTC()

	totalQty := decimal.Zero
	totalCOGS := decimal.Zero

	type ProcessedLine struct {
		ItemCode      string
		ItemName      string
		Qty           decimal.Decimal
		ValuationRate decimal.Decimal
		LineCOGS      decimal.Decimal
		NewQty        decimal.Decimal
		NewValue      decimal.Decimal
	}

	var processedLines []ProcessedLine

	for idx, item := range req.Items {
		if item.Qty.IsZero() || item.Qty.IsNegative() {
			return nil, nil, fmt.Errorf("item %d: %w", idx+1, ErrZeroDeliveryQty)
		}

		// Concurrency Guard & Row-Level Lock on Item-Warehouse balance
		var currentQty, valuationRate, stockValue decimal.Decimal
		queryLock := `
			SELECT current_qty, valuation_rate, stock_value
			FROM item_warehouse_balances
			WHERE company_id = $1 AND item_code = $2 AND warehouse_id = $3
			FOR UPDATE;
		`
		err := tx.QueryRow(ctx, queryLock, req.CompanyID, item.ItemCode, req.FromWarehouseID).Scan(&currentQty, &valuationRate, &stockValue)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, nil, fmt.Errorf("%w for item %s in warehouse %s (0 in stock)", ErrInsufficientStock, item.ItemCode, req.FromWarehouseID)
			}
			return nil, nil, fmt.Errorf("failed to lock balance for item %s: %w", item.ItemCode, err)
		}

		// Negative Stock Prevention
		if currentQty.LessThan(item.Qty) {
			return nil, nil, fmt.Errorf("%w for item %s: available %s, requested %s", ErrInsufficientStock, item.ItemCode, currentQty.String(), item.Qty.String())
		}

		// Outflow at current moving average valuation rate
		lineCOGS := item.Qty.Mul(valuationRate)
		newQty := currentQty.Sub(item.Qty)
		newValue := newQty.Mul(valuationRate)

		totalQty = totalQty.Add(item.Qty)
		totalCOGS = totalCOGS.Add(lineCOGS)

		processedLines = append(processedLines, ProcessedLine{
			ItemCode:      item.ItemCode,
			ItemName:      item.ItemName,
			Qty:           item.Qty,
			ValuationRate: valuationRate,
			LineCOGS:      lineCOGS,
			NewQty:        newQty,
			NewValue:      newValue,
		})
	}

	// 1. Insert Delivery Note Header
	insertDN := `
		INSERT INTO delivery_notes (
			id, company_id, delivery_note_number, customer_name, posting_date,
			from_warehouse_id, to_warehouse_id, cogs_account_id, stock_asset_account_id,
			total_qty, total_cogs_amount, docstatus, remarks, posted_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16);
	`
	_, err = tx.Exec(ctx, insertDN,
		deliveryID, req.CompanyID, deliveryNumber, req.CustomerName, req.PostingDate,
		req.FromWarehouseID, req.ToWarehouseID, req.COGSAccountID, req.StockAssetAccountID,
		totalQty, totalCOGS, doctype.DocSubmitted, req.Remarks, now, now, now,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to insert delivery note: %w", err)
	}

	// 2. Update Balances and Insert Stock Ledger Entries
	insertItem := `
		INSERT INTO delivery_note_items (id, parent_id, item_code, item_name, qty, valuation_rate, cogs_amount, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
	`
	insertSLE := `
		INSERT INTO stock_ledger_entries (
			id, company_id, posting_datetime, item_code, item_name,
			from_warehouse_id, to_warehouse_id, voucher_type, voucher_id, voucher_no,
			actual_qty, valuation_rate, stock_value_difference,
			qty_after_transaction, stock_value_after_transaction, remarks, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17);
	`
	updateBal := `
		UPDATE item_warehouse_balances
		SET current_qty = $1, stock_value = $2, updated_at = $3
		WHERE company_id = $4 AND item_code = $5 AND warehouse_id = $6;
	`

	var deliveryItems []DeliveryNoteItem

	for _, line := range processedLines {
		itemID := uuid.New()
		_, err = tx.Exec(ctx, insertItem, itemID, deliveryID, line.ItemCode, line.ItemName, line.Qty, line.ValuationRate, line.LineCOGS, now)
		if err != nil {
			return nil, nil, err
		}

		deliveryItems = append(deliveryItems, DeliveryNoteItem{
			ID:            itemID,
			ParentID:      deliveryID,
			ItemCode:      line.ItemCode,
			ItemName:      line.ItemName,
			Qty:           line.Qty,
			ValuationRate: line.ValuationRate,
			COGSAmount:    line.LineCOGS,
			CreatedAt:     now,
		})

		// Deduct stock from internal warehouse balance
		_, err = tx.Exec(ctx, updateBal, line.NewQty, line.NewValue, now, req.CompanyID, line.ItemCode, req.FromWarehouseID)
		if err != nil {
			return nil, nil, err
		}

		// Double-entry stock ledger movement: From Internal Warehouse -> Customer Location (Virtual)
		sleID := uuid.New()
		negQty := line.Qty.Neg()
		negValDiff := line.LineCOGS.Neg()
		_, err = tx.Exec(ctx, insertSLE,
			sleID, req.CompanyID, now, line.ItemCode, line.ItemName,
			req.FromWarehouseID, req.ToWarehouseID, "Delivery Note", deliveryID, deliveryNumber,
			negQty, line.ValuationRate, negValDiff, line.NewQty, line.NewValue, req.Remarks, now,
		)
		if err != nil {
			return nil, nil, err
		}
	}

	// 3. Perpetual Inventory: Atomic COGS Auto-Posting to Partitioned gl_entries table
	// Line 1: Debit 5010 - Cost of Goods Sold (COGS)
	// Line 2: Credit 1030 - Stock In Hand (Inventory Asset)
	var glEntries []accounting.GLEntry

	if totalCOGS.IsPositive() {
		// Debit COGS
		glDebitCOGS := accounting.GLEntry{
			ID:           uuid.New(),
			CompanyID:    req.CompanyID,
			PostingDate:  req.PostingDate,
			VoucherType:  "Delivery Note",
			VoucherID:    deliveryID,
			VoucherNo:    deliveryNumber,
			AccountID:    req.COGSAccountID,
			AccountCode:  req.COGSAccountCode,
			PartyType:    &req.CustomerName,
			Currency:     "BDT",
			Debit:        totalCOGS,
			Credit:       decimal.Zero,
			ExchangeRate: decimal.NewFromInt(1),
			DebitBase:    totalCOGS,
			CreditBase:   decimal.Zero,
			Remarks:      fmt.Sprintf("COGS recognition for dispatch %s to %s", deliveryNumber, req.CustomerName),
			IsCancelled:  false,
			CreatedAt:    now,
		}

		// Credit Stock In Hand
		glCreditStock := accounting.GLEntry{
			ID:           uuid.New(),
			CompanyID:    req.CompanyID,
			PostingDate:  req.PostingDate,
			VoucherType:  "Delivery Note",
			VoucherID:    deliveryID,
			VoucherNo:    deliveryNumber,
			AccountID:    req.StockAssetAccountID,
			AccountCode:  req.StockAssetCode,
			PartyType:    nil,
			Currency:     "BDT",
			Debit:        decimal.Zero,
			Credit:       totalCOGS,
			ExchangeRate: decimal.NewFromInt(1),
			DebitBase:    decimal.Zero,
			CreditBase:   totalCOGS,
			Remarks:      fmt.Sprintf("Stock asset reduction for dispatch %s", deliveryNumber),
			IsCancelled:  false,
			CreatedAt:    now,
		}

		glEntries = append(glEntries, glDebitCOGS, glCreditStock)

		// Insert both into Partitioned gl_entries
		insertGL := `
			INSERT INTO gl_entries (
				id, company_id, posting_date, voucher_type, voucher_id, voucher_no,
				account_id, account_code, party_type, currency,
				debit, credit, exchange_rate, debit_base, credit_base, remarks, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17);
		`
		for _, gl := range glEntries {
			_, err = tx.Exec(ctx, insertGL,
				gl.ID, gl.CompanyID, gl.PostingDate, gl.VoucherType, gl.VoucherID, gl.VoucherNo,
				gl.AccountID, gl.AccountCode, gl.PartyType, gl.Currency,
				gl.Debit, gl.Credit, gl.ExchangeRate, gl.DebitBase, gl.CreditBase, gl.Remarks, now,
			)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to post perpetual COGS entry: %w", err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, fmt.Errorf("commit failed: %w", err)
	}

	dn := &DeliveryNote{
		BaseDocument: doctype.BaseDocument{
			ID:        deliveryID,
			DocStatus: doctype.DocSubmitted,
			CreatedAt: now,
			UpdatedAt: now,
		},
		CompanyID:           req.CompanyID,
		DeliveryNoteNumber:  deliveryNumber,
		CustomerName:        req.CustomerName,
		PostingDate:         req.PostingDate,
		FromWarehouseID:     req.FromWarehouseID,
		ToWarehouseID:       req.ToWarehouseID,
		COGSAccountID:       req.COGSAccountID,
		StockAssetAccountID: req.StockAssetAccountID,
		TotalQty:            totalQty,
		TotalCOGSAmount:     totalCOGS,
		Items:               deliveryItems,
		Remarks:             req.Remarks,
		PostedAt:            &now,
	}

	return dn, glEntries, nil
}
