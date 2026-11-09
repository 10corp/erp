package accounting

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestZeroFloatPrecision(t *testing.T) {
	// Classic floating point trap: 0.1 + 0.2 == 0.30000000000000004 in float64
	d1 := decimal.RequireFromString("0.1")
	d2 := decimal.RequireFromString("0.2")
	expected := decimal.RequireFromString("0.3")

	result := d1.Add(d2)
	if !result.Equal(expected) {
		t.Fatalf("Decimal precision failed: got %s, expected %s", result.String(), expected.String())
	}
}

func TestMultiCurrencyExchangeRateMath(t *testing.T) {
	// USD 1,250.75 @ Exchange Rate 117.85 BDT/USD
	amountUSD := decimal.RequireFromString("1250.75")
	exchangeRate := decimal.RequireFromString("117.85")

	amountBDT := amountUSD.Mul(exchangeRate)
	// 1250.75 * 117.85 = 147400.8875
	expectedBDT := decimal.RequireFromString("147400.8875")

	if !amountBDT.Equal(expectedBDT) {
		t.Fatalf("Exchange rate math failed: got %s, expected %s", amountBDT.String(), expectedBDT.String())
	}
}

func TestUnbalancedDebitCreditRule(t *testing.T) {
	debitTotal := decimal.RequireFromString("500.00")
	creditTotal := decimal.RequireFromString("499.99")

	if debitTotal.Equal(creditTotal) {
		t.Fatalf("Unbalanced amounts incorrectly reported as equal")
	}

	diff := debitTotal.Sub(creditTotal)
	if !diff.Equal(decimal.RequireFromString("0.01")) {
		t.Fatalf("Discrepancy calculation error: got %s", diff.String())
	}
}
