package doctype

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrPOItemsRequired = errors.New("purchase order requires at least one item")
)

type PurchaseOrderItem struct {
	ID          uuid.UUID `json:"id"`
	ParentID    uuid.UUID `json:"parent_id"`
	ItemCode    string    `json:"item_code"`
	ItemName    string    `json:"item_name"`
	Description string    `json:"description"`
	WarehouseID uuid.UUID `json:"warehouse_id"`
	Qty         Money     `json:"qty"`
	Rate        Money     `json:"rate"`
	Amount      Money     `json:"amount"` // Qty * Rate
	TaxRate     Money     `json:"tax_rate"`
	TaxAmount   Money     `json:"tax_amount"`
	ReceivedQty Money     `json:"received_qty"`
}

type PurchaseOrder struct {
	BaseDocument
	CompanyID       uuid.UUID           `json:"company_id"`
	PONumber        string              `json:"po_number"` // PO-YYYY-#####
	SupplierID      *uuid.UUID          `json:"supplier_id,omitempty"`
	SupplierName    string              `json:"supplier_name"`
	TransactionDate time.Time           `json:"transaction_date"`
	DeliveryDate    *time.Time          `json:"delivery_date,omitempty"`
	Currency        string              `json:"currency"`
	ConversionRate  Money               `json:"conversion_rate"`
	NetTotal        Money               `json:"net_total"`
	TaxTotal        Money               `json:"tax_total"`
	GrandTotal      Money               `json:"grand_total"`
	BaseNetTotal    Money               `json:"base_net_total"`
	BaseTaxTotal    Money               `json:"base_tax_total"`
	BaseGrandTotal  Money               `json:"base_grand_total"`
	Status          string              `json:"status"` // Draft, Submitted, To Receive and Bill, Completed, Cancelled
	Items           []PurchaseOrderItem `json:"items"`
	Remarks         string              `json:"remarks"`
}

func init() {
	GetRegistry().Register(&DocTypeDefinition{
		Name:          "Purchase Order",
		Module:        "Procurement",
		TableName:     "purchase_orders",
		IsSubmittable: true,
	})
}

// CalculateTotals computes line amounts, taxes, and base currency conversions
func (po *PurchaseOrder) CalculateTotals() error {
	if len(po.Items) == 0 {
		return ErrPOItemsRequired
	}

	if po.ConversionRate.IsZero() {
		po.ConversionRate = decimal.NewFromInt(1)
	}

	netTotal := decimal.Zero
	taxTotal := decimal.Zero

	for i := range po.Items {
		item := &po.Items[i]
		item.Amount = item.Qty.Mul(item.Rate).Round(6)

		if !item.TaxRate.IsZero() {
			item.TaxAmount = item.Amount.Mul(item.TaxRate).Div(decimal.NewFromInt(100)).Round(6)
		} else {
			item.TaxAmount = decimal.Zero
		}

		netTotal = netTotal.Add(item.Amount)
		taxTotal = taxTotal.Add(item.TaxAmount)
	}

	po.NetTotal = netTotal
	po.TaxTotal = taxTotal
	po.GrandTotal = netTotal.Add(taxTotal)

	po.BaseNetTotal = po.NetTotal.Mul(po.ConversionRate).Round(6)
	po.BaseTaxTotal = po.TaxTotal.Mul(po.ConversionRate).Round(6)
	po.BaseGrandTotal = po.GrandTotal.Mul(po.ConversionRate).Round(6)

	return nil
}

func (po *PurchaseOrder) Submit() error {
	if po.DocStatus != DocDraft {
		return fmt.Errorf("cannot submit purchase order in docstatus %d (must be Draft)", po.DocStatus)
	}
	if err := po.CalculateTotals(); err != nil {
		return err
	}
	po.DocStatus = DocSubmitted
	po.Status = "To Receive and Bill"
	po.UpdatedAt = time.Now().UTC()
	return nil
}

func (po *PurchaseOrder) Cancel() error {
	if po.DocStatus != DocSubmitted {
		return fmt.Errorf("cannot cancel purchase order in docstatus %d (must be Submitted)", po.DocStatus)
	}
	po.DocStatus = DocCancelled
	po.Status = "Cancelled"
	po.UpdatedAt = time.Now().UTC()
	return nil
}
