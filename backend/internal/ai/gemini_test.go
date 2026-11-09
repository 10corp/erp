package ai

import (
	"context"
	"testing"
)

func TestValidateMathGuardSuccess(t *testing.T) {
	data := &ExtractedInvoiceData{
		NetTotal:   "1000.000000",
		TaxTotal:   "150.000000",
		GrandTotal: "1150.000000", // 1000 + 150 = 1150
		Items: []ExtractedInvoiceItem{
			{
				ItemCode:    "ITEM-1",
				Qty:         "10.000000",
				UnitPrice:   "50.000000",
				TotalAmount: "500.000000",
			},
			{
				ItemCode:    "ITEM-2",
				Qty:         "5.000000",
				UnitPrice:   "100.000000",
				TotalAmount: "500.000000",
			},
		},
	}

	ValidateMathGuard(data)

	if data.MathMismatchWarning {
		t.Fatalf("Expected MathMismatchWarning to be false, got true (variance: %s)", data.MathVariance)
	}

	if data.MathVariance != "0.000000" {
		t.Fatalf("Expected 0 variance, got %s", data.MathVariance)
	}
}

func TestValidateMathGuardMismatch(t *testing.T) {
	data := &ExtractedInvoiceData{
		NetTotal:   "1000.000000",
		TaxTotal:   "150.000000",
		GrandTotal: "1199.000000", // Wrong grand total! (49 BDT mismatch)
		Items: []ExtractedInvoiceItem{
			{
				ItemCode:    "ITEM-1",
				Qty:         "10.000000",
				UnitPrice:   "100.000000",
				TotalAmount: "1000.000000",
			},
		},
	}

	ValidateMathGuard(data)

	if !data.MathMismatchWarning {
		t.Fatalf("Expected MathMismatchWarning to be true for 49 BDT discrepancy")
	}

	// Calculated: 1000 + 150 = 1150. Extracted: 1199. Diff = 1150 - 1199 = -49.000000
	expectedDiff := "-49.000000"
	if data.MathVariance != expectedDiff {
		t.Fatalf("Expected variance %s, got %s", expectedDiff, data.MathVariance)
	}
}

func TestExtractInvoiceFromDocument(t *testing.T) {
	service := NewGeminiIDPService()
	dummyBytes := []byte("%PDF-1.4 dummy invoice stream...")

	result, err := service.ExtractInvoiceFromDocument(context.Background(), dummyBytes, "application/pdf")
	if err != nil {
		t.Fatalf("Extraction failed: %v", err)
	}

	if result.InvoiceNumber == "" || result.SupplierName == "" {
		t.Fatalf("Missing required fields in extraction result")
	}

	if result.MathMismatchWarning {
		t.Fatalf("Expected math guard to pass on standard sample")
	}
}
