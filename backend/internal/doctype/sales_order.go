package doctype

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrCannotModifySubmitted = errors.New("Tryton Invariant: submitted document is immutable and cannot be updated")
	ErrInvalidOrderItems     = errors.New("sales order must contain at least one valid line item")
	ErrNegativeQuantity      = errors.New("item quantity and rate must be greater than zero")
)

// SalesOrderItem maps directly to ERPNext's `sales_order_item.json`
type SalesOrderItem struct {
	ID          uuid.UUID `json:"id"`
	ParentID    uuid.UUID `json:"parent_id"`
	ItemCode    string    `json:"item_code"`
	ItemName    string    `json:"item_name"`
	Description string    `json:"description"`
	Qty         Money     `json:"qty"`         // shopspring/decimal
	Rate        Money     `json:"rate"`        // Transaction currency rate
	Amount      Money     `json:"amount"`      // Qty * Rate
	BaseRate    Money     `json:"base_rate"`   // Rate * ConversionRate
	BaseAmount  Money     `json:"base_amount"` // Amount * ConversionRate
}

// SalesOrder combines ERPNext's clean DocType schema with Tryton's strict state transitions
type SalesOrder struct {
	BaseDocument
	CompanyID      uuid.UUID        `json:"company_id"`
	OrderNumber    string           `json:"order_number"` // SO-2026-#####
	Customer       string           `json:"customer"`
	CustomerName   string           `json:"customer_name"`
	OrderType      string           `json:"order_type"` // 'Sales', 'Maintenance', 'Shopping Cart'
	TransactionDate time.Time       `json:"transaction_date"`
	DeliveryDate   time.Time        `json:"delivery_date"`
	Currency       string           `json:"currency"`
	ConversionRate Money            `json:"conversion_rate"`
	Items          []SalesOrderItem `json:"items"`
	TotalQty       Money            `json:"total_qty"`
	NetTotal       Money            `json:"net_total"`
	BaseNetTotal   Money            `json:"base_net_total"`
	GrandTotal     Money            `json:"grand_total"`
	BaseGrandTotal Money            `json:"base_grand_total"`
}

func init() {
	// Register SalesOrder DocType metadata in runtime registry
	GetRegistry().Register(&DocTypeDefinition{
		Name:          "SalesOrder",
		Module:        "Selling",
		IsSubmittable: true,
		TableName:     "sales_orders",
		Fields: map[string]DocField{
			"order_number":     {FieldName: "order_number", Label: "Order #", FieldType: TypeData, Unique: true, Required: true},
			"customer":         {FieldName: "customer", Label: "Customer", FieldType: TypeLink, Options: "Customer", Required: true},
			"transaction_date": {FieldName: "transaction_date", Label: "Order Date", FieldType: TypeDate, Required: true},
			"delivery_date":    {FieldName: "delivery_date", Label: "Delivery Date", FieldType: TypeDate, Required: true},
			"currency":         {FieldName: "currency", Label: "Currency", FieldType: TypeLink, Options: "Currency", Required: true},
			"conversion_rate":  {FieldName: "conversion_rate", Label: "Exchange Rate", FieldType: TypeFloat, Required: true},
			"grand_total":      {FieldName: "grand_total", Label: "Grand Total", FieldType: TypeCurrency, ReadOnly: true},
			"items":            {FieldName: "items", Label: "Items", FieldType: TypeTable, Options: "SalesOrderItem", Required: true},
		},
		FieldOrder: []string{"order_number", "customer", "transaction_date", "delivery_date", "currency", "conversion_rate", "grand_total", "items"},
	})
}

// CalculateTotals calculates line item amounts, currency conversion, and grand totals with zero float error
func (so *SalesOrder) CalculateTotals() error {
	if len(so.Items) == 0 {
		return ErrInvalidOrderItems
	}

	if so.ConversionRate.IsZero() || so.ConversionRate.IsNegative() {
		so.ConversionRate = decimal.NewFromInt(1)
	}

	totalQty := decimal.Zero
	netTotal := decimal.Zero
	baseNetTotal := decimal.Zero

	for i := range so.Items {
		item := &so.Items[i]
		if item.Qty.IsNegative() || item.Rate.IsNegative() {
			return ErrNegativeQuantity
		}

		// Fixed-point arithmetic
		item.Amount = item.Qty.Mul(item.Rate)
		item.BaseRate = item.Rate.Mul(so.ConversionRate)
		item.BaseAmount = item.Amount.Mul(so.ConversionRate)

		totalQty = totalQty.Add(item.Qty)
		netTotal = netTotal.Add(item.Amount)
		baseNetTotal = baseNetTotal.Add(item.BaseAmount)
	}

	so.TotalQty = totalQty
	so.NetTotal = netTotal
	so.BaseNetTotal = baseNetTotal
	so.GrandTotal = netTotal
	so.BaseGrandTotal = baseNetTotal
	return nil
}

// Submit transitions the SalesOrder from Draft (0) to Submitted (1) using Tryton state machine logic
func (so *SalesOrder) Submit() error {
	if so.DocStatus != DocDraft {
		return fmt.Errorf("cannot submit order in docstatus %d (must be Draft)", so.DocStatus)
	}

	if err := so.CalculateTotals(); err != nil {
		return err
	}

	// Lock document state as immutable
	so.DocStatus = DocSubmitted
	so.UpdatedAt = time.Now().UTC()
	return nil
}

// Cancel transitions the SalesOrder from Submitted (1) to Cancelled (2)
func (so *SalesOrder) Cancel() error {
	if so.DocStatus != DocSubmitted {
		return fmt.Errorf("cannot cancel order in docstatus %d (must be Submitted)", so.DocStatus)
	}

	so.DocStatus = DocCancelled
	so.UpdatedAt = time.Now().UTC()
	return nil
}
