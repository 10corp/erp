package doctype

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestBOMCostingAndWorkOrderStockMoves(t *testing.T) {
	// 1. Setup Bill of Materials (e.g. 1 Unit of Industrial IoT Controller)
	bom := &BillOfMaterials{
		BaseDocument: BaseDocument{
			ID:        uuid.New(),
			DocStatus: DocDraft,
			CreatedAt: time.Now().UTC(),
		},
		BOMNumber:     "BOM-IOT-001",
		ItemCode:      "PROD-IOT-CTRL",
		ItemName:      "Industrial IoT Automation Controller",
		Quantity:      decimal.RequireFromString("1.00"),
		OperatingCost: decimal.RequireFromString("250.00"), // Assembly & Testing Cost
		Items: []BOMItem{
			{
				RawItemCode:     "RAW-PCB-01",
				RawItemName:     "Main Circuit Board",
				Qty:             decimal.RequireFromString("1.00"),
				Rate:            decimal.RequireFromString("500.00"),
				ScrapPercentage: decimal.Zero,
			},
			{
				RawItemCode:     "RAW-CHIP-42",
				RawItemName:     "Microcontroller SoC",
				Qty:             decimal.RequireFromString("2.00"),
				Rate:            decimal.RequireFromString("300.00"), // 2 * 300 = 600
				ScrapPercentage: decimal.RequireFromString("5.00"),  // 5% scrap allowance
			},
			{
				RawItemCode:     "RAW-ENC-09",
				RawItemName:     "Aluminum Chassis",
				Qty:             decimal.RequireFromString("1.00"),
				Rate:            decimal.RequireFromString("150.00"),
				ScrapPercentage: decimal.Zero,
			},
		},
	}

	// Calculate and Submit BOM
	if err := bom.Submit(); err != nil {
		t.Fatalf("expected BOM submission, got: %v", err)
	}

	if bom.DocStatus != DocSubmitted {
		t.Fatalf("expected BOM docstatus=1, got %d", bom.DocStatus)
	}

	// 500 + (600 * 1.05 = 630) + 150 = 1280 Raw Material
	if !bom.TotalRawMaterialCost.Equal(decimal.RequireFromString("1280.00")) {
		t.Fatalf("expected raw material cost 1280.00, got %s", bom.TotalRawMaterialCost)
	}

	// Total Cost = 1280 + 250 operating = 1530.00
	if !bom.TotalCost.Equal(decimal.RequireFromString("1530.00")) {
		t.Fatalf("expected total unit cost 1530.00, got %s", bom.TotalCost)
	}

	// 2. Setup Work Order to produce 10 units
	rawStoreLoc := uuid.New()
	wipFloorLoc := uuid.New()
	finishedStoreLoc := uuid.New()

	wo := &WorkOrder{
		BaseDocument: BaseDocument{
			ID:        uuid.New(),
			DocStatus: DocDraft,
		},
		WorkOrderNo:        "WO-2026-0042",
		BOMID:              bom.ID,
		ProductionItemCode: bom.ItemCode,
		QtyToProduce:       decimal.RequireFromString("10.00"),
		SourceWarehouseID:  rawStoreLoc,
		WIPWarehouseID:     wipFloorLoc,
		TargetWarehouseID:  finishedStoreLoc,
	}

	if err := wo.Submit(); err != nil {
		t.Fatalf("expected WO submission, got: %v", err)
	}

	if wo.DocStatus != DocSubmitted || wo.Status != "In-Progress" {
		t.Fatalf("expected WO docstatus=1 and status=In-Progress, got docstatus=%d, status=%s", wo.DocStatus, wo.Status)
	}

	// 3. Generate Tryton Double-Entry Stock Moves
	moves := wo.GenerateStockMoves(bom)

	// 3 raw material items into WIP + 1 finished good output from WIP = 4 stock movements
	if len(moves) != 4 {
		t.Fatalf("expected 4 stock moves, got %d", len(moves))
	}

	// Move 1: PCB consumption (10 units from rawStore -> wipFloor)
	move1 := moves[0]
	if move1["from_warehouse_id"] != rawStoreLoc || move1["to_warehouse_id"] != wipFloorLoc {
		t.Errorf("move 1 warehouse mismatch: from=%v, to=%v", move1["from_warehouse_id"], move1["to_warehouse_id"])
	}
	if !move1["qty"].(decimal.Decimal).Equal(decimal.RequireFromString("10.00")) {
		t.Errorf("expected 10 PCBs consumed, got %v", move1["qty"])
	}

	// Move 4: Finished product (10 units from wipFloor -> finishedStore)
	finishMove := moves[3]
	if finishMove["from_warehouse_id"] != wipFloorLoc || finishMove["to_warehouse_id"] != finishedStoreLoc {
		t.Errorf("finish move location mismatch: from=%v, to=%v", finishMove["from_warehouse_id"], finishMove["to_warehouse_id"])
	}
	if !finishMove["qty"].(decimal.Decimal).Equal(decimal.RequireFromString("10.00")) {
		t.Errorf("expected 10 controllers produced, got %v", finishMove["qty"])
	}
}
