package inventory

import (
	"time"

	"github.com/enterprise-erp/core/internal/doctype"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type WarehouseType string

const (
	WarehouseInternal WarehouseType = "Internal"
	WarehouseSupplier WarehouseType = "Supplier" // Virtual location for goods arriving from vendor
	WarehouseCustomer WarehouseType = "Customer" // Virtual location for goods delivered to customer
	WarehouseTransit  WarehouseType = "Transit"
	WarehouseWaste    WarehouseType = "Waste"
)

type Warehouse struct {
	ID            uuid.UUID     `json:"id"`
	CompanyID     uuid.UUID     `json:"company_id"`
	Code          string        `json:"code"`
	Name          string        `json:"name"`
	WarehouseType WarehouseType `json:"warehouse_type"`
	AccountID     *uuid.UUID    `json:"account_id,omitempty"` // Default Stock Asset Account (1030)
	IsActive      bool          `json:"is_active"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

type ItemWarehouseBalance struct {
	ID            uuid.UUID       `json:"id"`
	CompanyID     uuid.UUID       `json:"company_id"`
	ItemCode      string          `json:"item_code"`
	WarehouseID   uuid.UUID       `json:"warehouse_id"`
	CurrentQty    decimal.Decimal `json:"current_qty"`
	ValuationRate decimal.Decimal `json:"valuation_rate"`
	StockValue    decimal.Decimal `json:"stock_value"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type StockLedgerEntry struct {
	ID                         uuid.UUID       `json:"id"`
	CompanyID                  uuid.UUID       `json:"company_id"`
	PostingDateTime            time.Time       `json:"posting_datetime"`
	ItemCode                   string          `json:"item_code"`
	ItemName                   string          `json:"item_name"`
	FromWarehouseID            uuid.UUID       `json:"from_warehouse_id"` // Tryton Double-Entry
	ToWarehouseID              uuid.UUID       `json:"to_warehouse_id"`   // Tryton Double-Entry
	VoucherType                string          `json:"voucher_type"`      // 'Delivery Note', 'Purchase Receipt'
	VoucherID                  uuid.UUID       `json:"voucher_id"`
	VoucherNo                  string          `json:"voucher_no"`
	ActualQty                  decimal.Decimal `json:"actual_qty"`
	ValuationRate              decimal.Decimal `json:"valuation_rate"`
	StockValueDifference       decimal.Decimal `json:"stock_value_difference"`
	QtyAfterTransaction        decimal.Decimal `json:"qty_after_transaction"`
	StockValueAfterTransaction decimal.Decimal `json:"stock_value_after_transaction"`
	Remarks                    string          `json:"remarks"`
	IsCancelled                bool            `json:"is_cancelled"`
	CreatedAt                  time.Time       `json:"created_at"`
}

type DeliveryNoteItem struct {
	ID            uuid.UUID       `json:"id"`
	ParentID      uuid.UUID       `json:"parent_id"`
	ItemCode      string          `json:"item_code"`
	ItemName      string          `json:"item_name"`
	Qty           decimal.Decimal `json:"qty"`
	ValuationRate decimal.Decimal `json:"valuation_rate"`
	COGSAmount    decimal.Decimal `json:"cogs_amount"` // Qty * ValuationRate
	CreatedAt     time.Time       `json:"created_at"`
}

type DeliveryNote struct {
	doctype.BaseDocument
	CompanyID           uuid.UUID          `json:"company_id"`
	DeliveryNoteNumber  string             `json:"delivery_note_number"` // DN-YYYY-#####
	CustomerName        string             `json:"customer_name"`
	PostingDate         time.Time          `json:"posting_date"`
	FromWarehouseID     uuid.UUID          `json:"from_warehouse_id"`     // Internal physical warehouse
	ToWarehouseID       uuid.UUID          `json:"to_warehouse_id"`       // Customer virtual warehouse
	COGSAccountID       uuid.UUID          `json:"cogs_account_id"`       // 5010 - Cost of Goods Sold
	StockAssetAccountID uuid.UUID          `json:"stock_asset_account_id"`// 1030 - Stock In Hand
	TotalQty            decimal.Decimal    `json:"total_qty"`
	TotalCOGSAmount     decimal.Decimal    `json:"total_cogs_amount"`
	Items               []DeliveryNoteItem `json:"items"`
	Remarks             string             `json:"remarks"`
	PostedAt            *time.Time         `json:"posted_at,omitempty"`
}

// Moving Average Mathematical Helper Function:
// New Valuation Rate = ((Current Qty * Current Rate) + (Incoming Qty * Incoming Rate)) / (Current Qty + Incoming Qty)
func CalculateMovingAverage(currentQty, currentRate, incomingQty, incomingRate decimal.Decimal) (newQty, newRate, newValue decimal.Decimal) {
	if incomingQty.IsZero() || incomingQty.IsNegative() {
		return currentQty, currentRate, currentQty.Mul(currentRate)
	}

	newQty = currentQty.Add(incomingQty)
	currentValue := currentQty.Mul(currentRate)
	incomingValue := incomingQty.Mul(incomingRate)
	newValue = currentValue.Add(incomingValue)

	if newQty.IsPositive() {
		newRate = newValue.Div(newQty)
	} else {
		newRate = decimal.Zero
	}

	return newQty, newRate, newValue
}
