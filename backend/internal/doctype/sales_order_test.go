package doctype

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestSalesOrderCalculationAndState(t *testing.T) {
	so := &SalesOrder{
		BaseDocument: BaseDocument{
			ID:        uuid.New(),
			DocStatus: DocDraft,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		OrderNumber:    "SO-2026-00001",
		Customer:       "CUST-001",
		CustomerName:   "Apex Global Client",
		Currency:       "USD",
		ConversionRate: decimal.RequireFromString("117.50"),
		Items: []SalesOrderItem{
			{
				ID:       uuid.New(),
				ItemCode: "ITEM-A",
				ItemName: "Enterprise Cloud License",
				Qty:      decimal.RequireFromString("5.0"),
				Rate:     decimal.RequireFromString("200.00"), // 5 * 200 = 1000 USD
			},
			{
				ID:       uuid.New(),
				ItemCode: "ITEM-B",
				ItemName: "Implementation Support Pack",
				Qty:      decimal.RequireFromString("2.0"),
				Rate:     decimal.RequireFromString("350.50"), // 2 * 350.50 = 701 USD
			},
		},
	}

	// 1. Test calculation
	err := so.CalculateTotals()
	if err != nil {
		t.Fatalf("Calculation failed: %v", err)
	}

	expectedGrandTotalUSD := decimal.RequireFromString("1701.00")
	if !so.GrandTotal.Equal(expectedGrandTotalUSD) {
		t.Fatalf("GrandTotal mismatch: got %s, expected %s", so.GrandTotal.String(), expectedGrandTotalUSD.String())
	}

	// 1701 * 117.50 = 199867.50 BDT
	expectedBaseGrandTotalBDT := decimal.RequireFromString("199867.50")
	if !so.BaseGrandTotal.Equal(expectedBaseGrandTotalBDT) {
		t.Fatalf("BaseGrandTotal mismatch: got %s, expected %s", so.BaseGrandTotal.String(), expectedBaseGrandTotalBDT.String())
	}

	// 2. Test Tryton submission transition
	if err := so.Submit(); err != nil {
		t.Fatalf("Submit failed: %v", err)
	}

	if so.DocStatus != DocSubmitted {
		t.Fatalf("Expected DocStatus to be DocSubmitted (1), got %d", so.DocStatus)
	}

	// Re-submitting should fail
	if err := so.Submit(); err == nil {
		t.Fatalf("Expected error re-submitting already submitted order")
	}

	// 3. Test cancellation transition
	if err := so.Cancel(); err != nil {
		t.Fatalf("Cancel failed: %v", err)
	}

	if so.DocStatus != DocCancelled {
		t.Fatalf("Expected DocStatus to be DocCancelled (2), got %d", so.DocStatus)
	}
}
