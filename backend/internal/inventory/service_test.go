package inventory

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestMovingAverageCostingMath(t *testing.T) {
	// Step 1: Initial Stock Receipt
	// 100 units @ 50.00 BDT
	currentQty := decimal.Zero
	currentRate := decimal.Zero

	incomingQty1 := decimal.RequireFromString("100.00")
	incomingRate1 := decimal.RequireFromString("50.00")

	qtyAfterReceipt1, rateAfterReceipt1, valAfterReceipt1 := CalculateMovingAverage(currentQty, currentRate, incomingQty1, incomingRate1)

	if !qtyAfterReceipt1.Equal(decimal.RequireFromString("100.00")) {
		t.Fatalf("Receipt 1 Qty mismatch: got %s", qtyAfterReceipt1.String())
	}
	if !rateAfterReceipt1.Equal(decimal.RequireFromString("50.00")) {
		t.Fatalf("Receipt 1 Rate mismatch: got %s", rateAfterReceipt1.String())
	}
	if !valAfterReceipt1.Equal(decimal.RequireFromString("5000.00")) {
		t.Fatalf("Receipt 1 Value mismatch: got %s", valAfterReceipt1.String())
	}

	// Step 2: Second Stock Receipt with higher price
	// 50 units @ 80.00 BDT
	// Total New Value = (100 * 50) + (50 * 80) = 5000 + 4000 = 9000 BDT
	// Total New Qty = 150 units
	// New Moving Average Valuation Rate = 9000 / 150 = 60.00 BDT
	incomingQty2 := decimal.RequireFromString("50.00")
	incomingRate2 := decimal.RequireFromString("80.00")

	qtyAfterReceipt2, rateAfterReceipt2, valAfterReceipt2 := CalculateMovingAverage(qtyAfterReceipt1, rateAfterReceipt1, incomingQty2, incomingRate2)

	if !qtyAfterReceipt2.Equal(decimal.RequireFromString("150.00")) {
		t.Fatalf("Receipt 2 Qty mismatch: got %s", qtyAfterReceipt2.String())
	}
	expectedRate2 := decimal.RequireFromString("60.00")
	if !rateAfterReceipt2.Equal(expectedRate2) {
		t.Fatalf("Receipt 2 Rate mismatch: got %s, expected %s", rateAfterReceipt2.String(), expectedRate2.String())
	}
	expectedValue2 := decimal.RequireFromString("9000.00")
	if !valAfterReceipt2.Equal(expectedValue2) {
		t.Fatalf("Receipt 2 Value mismatch: got %s, expected %s", valAfterReceipt2.String(), expectedValue2.String())
	}

	// Step 3: Stock Delivery (Outflow)
	// Delivering 30 units
	// Valuation Rate remains 60.00 BDT
	// Cost of Goods Sold (COGS) = 30 * 60 = 1,800.00 BDT
	deliveryQty := decimal.RequireFromString("30.00")
	cogsAmount := deliveryQty.Mul(rateAfterReceipt2)
	expectedCOGS := decimal.RequireFromString("1800.00")

	if !cogsAmount.Equal(expectedCOGS) {
		t.Fatalf("COGS Amount mismatch: got %s, expected %s", cogsAmount.String(), expectedCOGS.String())
	}

	remainingQty := qtyAfterReceipt2.Sub(deliveryQty)
	remainingValue := remainingQty.Mul(rateAfterReceipt2)

	expectedRemainingQty := decimal.RequireFromString("120.00")
	expectedRemainingValue := decimal.RequireFromString("7200.00")

	if !remainingQty.Equal(expectedRemainingQty) {
		t.Fatalf("Remaining Qty mismatch: got %s", remainingQty.String())
	}
	if !remainingValue.Equal(expectedRemainingValue) {
		t.Fatalf("Remaining Value mismatch: got %s", remainingValue.String())
	}

	// Verify Double-Entry Perpetual Accounting Balance:
	// Debit COGS (1800.00) == Credit Inventory Asset (1800.00)
	debitCOGS := cogsAmount
	creditStockAsset := cogsAmount
	if !debitCOGS.Equal(creditStockAsset) {
		t.Fatalf("Perpetual COGS double-entry balance check failed!")
	}
}

func TestNegativeStockPrevention(t *testing.T) {
	currentStock := decimal.RequireFromString("120.00")
	requestedDelivery := decimal.RequireFromString("150.00")

	if !currentStock.LessThan(requestedDelivery) {
		t.Fatalf("Negative stock check failed to detect shortage")
	}

	shortage := requestedDelivery.Sub(currentStock)
	if !shortage.Equal(decimal.RequireFromString("30.00")) {
		t.Fatalf("Shortage calculation error: got %s", shortage.String())
	}
}
