package doctype

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrBOMNotDraft             = errors.New("bill of materials is not in draft state")
	ErrWorkOrderNotDraft       = errors.New("work order is not in draft state")
	ErrWorkOrderAlreadyFinished = errors.New("work order is already completed")
)

// BOMItem represents a raw material item in an ERPNext-style BOM
type BOMItem struct {
	ID              uuid.UUID       `json:"id"`
	RawItemCode     string          `json:"raw_item_code"`
	RawItemName     string          `json:"raw_item_name"`
	Qty             Money           `json:"qty"`
	UOM             string          `json:"uom"`
	Rate            Money           `json:"rate"`
	Amount          Money           `json:"amount"`
	ScrapPercentage Money           `json:"scrap_percentage"`
}

// BillOfMaterials (BOM) specifies the recipe for producing a manufactured item
type BillOfMaterials struct {
	BaseDocument
	TenantID             uuid.UUID       `json:"tenant_id"`
	CompanyID            uuid.UUID       `json:"company_id"`
	BOMNumber            string          `json:"bom_number"`
	ItemCode             string          `json:"item_code"`
	ItemName             string          `json:"item_name"`
	Quantity             Money           `json:"quantity"` // Output batch size
	UOM                  string          `json:"uom"`
	TotalRawMaterialCost Money           `json:"total_raw_material_cost"`
	OperatingCost        Money           `json:"operating_cost"`
	TotalCost            Money           `json:"total_cost"`
	IsActive             bool            `json:"is_active"`
	IsDefault            bool            `json:"is_default"`
	Items                []BOMItem       `json:"items"`
}

func init() {
	GetRegistry().Register(&DocTypeDefinition{
		Name:          "Bill of Materials",
		Module:        "Manufacturing & Production",
		IsSubmittable: true,
		TableName:     "manufacturing_boms",
		Fields: map[string]DocField{
			"bom_number": {FieldName: "bom_number", Label: "BOM Number", FieldType: TypeData, Required: true, Unique: true},
			"item_code":  {FieldName: "item_code", Label: "Finished Item Code", FieldType: TypeLink, Options: "Item", Required: true},
			"quantity":   {FieldName: "quantity", Label: "Quantity", FieldType: TypeFloat, Required: true, Default: "1.000000"},
			"total_cost": {FieldName: "total_cost", Label: "Total Unit Cost", FieldType: TypeCurrency, ReadOnly: true},
		},
		FieldOrder: []string{"bom_number", "item_code", "quantity", "total_cost"},
	})
}

// CalculateCosts sums all raw material costs with scrap allowances
func (b *BillOfMaterials) CalculateCosts() {
	totalRM := decimal.Zero
	for i := range b.Items {
		item := &b.Items[i]
		baseAmount := item.Qty.Mul(item.Rate)
		if item.ScrapPercentage.GreaterThan(decimal.Zero) {
			scrapFactor := decimal.NewFromInt(1).Add(item.ScrapPercentage.Div(decimal.NewFromInt(100)))
			baseAmount = baseAmount.Mul(scrapFactor)
		}
		item.Amount = baseAmount
		totalRM = totalRM.Add(baseAmount)
	}

	b.TotalRawMaterialCost = totalRM
	b.TotalCost = totalRM.Add(b.OperatingCost)
}

// Submit locks the BOM definition
func (b *BillOfMaterials) Submit() error {
	if b.DocStatus != DocDraft {
		return ErrBOMNotDraft
	}
	if len(b.Items) == 0 {
		return errors.New("BOM must contain at least one raw material line item")
	}

	b.CalculateCosts()
	b.DocStatus = DocSubmitted
	b.UpdatedAt = time.Now().UTC()
	return nil
}

// WorkOrder coordinates the double-entry manufacturing execution across warehouse locations
type WorkOrder struct {
	BaseDocument
	TenantID           uuid.UUID       `json:"tenant_id"`
	CompanyID          uuid.UUID       `json:"company_id"`
	WorkOrderNo        string          `json:"work_order_no"`
	BOMID              uuid.UUID       `json:"bom_id"`
	ProductionItemCode string          `json:"production_item_code"`
	QtyToProduce       Money           `json:"qty_to_produce"`
	ProducedQty        Money           `json:"produced_qty"`
	
	// Tryton & Odoo Double-Entry Stock Locations
	SourceWarehouseID  uuid.UUID       `json:"source_warehouse_id"` // Raw Material Store
	WIPWarehouseID     uuid.UUID       `json:"wip_warehouse_id"`    // Production Floor
	TargetWarehouseID  uuid.UUID       `json:"target_warehouse_id"` // Finished Goods Store
	
	Status             string          `json:"status"` // 'Draft', 'Submitted', 'In-Progress', 'Completed', 'Cancelled'
}

// Submit starts the work order
func (w *WorkOrder) Submit() error {
	if w.DocStatus != DocDraft {
		return ErrWorkOrderNotDraft
	}
	if w.QtyToProduce.IsZero() || w.QtyToProduce.IsNegative() {
		return errors.New("qty to produce must be greater than zero")
	}

	w.DocStatus = DocSubmitted
	w.Status = "In-Progress"
	w.UpdatedAt = time.Now().UTC()
	return nil
}

// GenerateStockMoves creates double-entry inventory ledger movements
// Move 1: Raw Materials from Source Store -> WIP Floor
// Move 2: Finished Goods from WIP Floor -> Target Store
func (w *WorkOrder) GenerateStockMoves(bom *BillOfMaterials) []map[string]interface{} {
	var moves []map[string]interface{}

	// Move 1: Consume raw materials into WIP
	for _, item := range bom.Items {
		requiredQty := item.Qty.Mul(w.QtyToProduce)
		moves = append(moves, map[string]interface{}{
			"type":              "Raw Material Consumption",
			"from_warehouse_id": w.SourceWarehouseID,
			"to_warehouse_id":   w.WIPWarehouseID,
			"item_code":         item.RawItemCode,
			"qty":               requiredQty,
			"valuation_rate":    item.Rate,
			"voucher_type":      "Work Order",
			"voucher_no":        w.WorkOrderNo,
		})
	}

	// Move 2: Output finished product to Target Store
	moves = append(moves, map[string]interface{}{
		"type":              "Finished Goods Production",
		"from_warehouse_id": w.WIPWarehouseID,
		"to_warehouse_id":   w.TargetWarehouseID,
		"item_code":         w.ProductionItemCode,
		"qty":               w.QtyToProduce,
		"valuation_rate":    bom.TotalCost.Div(bom.Quantity),
		"voucher_type":      "Work Order",
		"voucher_no":        w.WorkOrderNo,
	})

	return moves
}
