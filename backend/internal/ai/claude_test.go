package ai

import (
	"context"
	"testing"
)

func TestClaudeActionBrainPaymentReceipt(t *testing.T) {
	brain := NewClaudeActionBrain()
	query := "Show me overdue invoices for Acme and draft receipt"

	action, err := brain.ParseUserIntent(context.Background(), query, nil)
	if err != nil {
		t.Fatalf("ParseUserIntent failed: %v", err)
	}

	if action.Type != ActionOpenDrawer {
		t.Fatalf("Expected ActionOpenDrawer, got %s", action.Type)
	}

	if action.Target != "PaymentEntry" {
		t.Fatalf("Expected target PaymentEntry, got %s", action.Target)
	}

	partyName, ok := action.Parameters["party_name"].(string)
	if !ok || partyName != "Acme" {
		t.Fatalf("Expected party_name Acme, got %v", action.Parameters["party_name"])
	}

	filterOverdue, ok := action.Parameters["filter_overdue_invoices"].(bool)
	if !ok || !filterOverdue {
		t.Fatalf("Expected filter_overdue_invoices to be true")
	}
}

func TestClaudeActionBrainSalesInvoice(t *testing.T) {
	brain := NewClaudeActionBrain()
	query := "Bill Acme for annual cloud software licenses"

	action, err := brain.ParseUserIntent(context.Background(), query, nil)
	if err != nil {
		t.Fatalf("ParseUserIntent failed: %v", err)
	}

	if action.Type != ActionOpenDrawer || action.Target != "SalesInvoice" {
		t.Fatalf("Expected open_drawer for SalesInvoice, got %s -> %s", action.Type, action.Target)
	}

	if action.Parameters["customer_name"] != "Acme" {
		t.Fatalf("Expected customer_name Acme, got %v", action.Parameters["customer_name"])
	}
}

func TestClaudeActionBrainDeliveryNote(t *testing.T) {
	brain := NewClaudeActionBrain()
	query := "Deliver 50 servers to Global Tech Corp"

	action, err := brain.ParseUserIntent(context.Background(), query, nil)
	if err != nil {
		t.Fatalf("ParseUserIntent failed: %v", err)
	}

	if action.Type != ActionOpenDrawer || action.Target != "DeliveryNote" {
		t.Fatalf("Expected open_drawer for DeliveryNote, got %s -> %s", action.Type, action.Target)
	}
}

func TestClaudeActionBrainNavigation(t *testing.T) {
	brain := NewClaudeActionBrain()
	query := "Show me the CEO dashboard and P&L report"

	action, err := brain.ParseUserIntent(context.Background(), query, nil)
	if err != nil {
		t.Fatalf("ParseUserIntent failed: %v", err)
	}

	if action.Type != ActionNavigate || action.Target != "dashboard" {
		t.Fatalf("Expected navigate to dashboard, got %s -> %s", action.Type, action.Target)
	}
}

func TestClaudeActionBrainEmptyQuery(t *testing.T) {
	brain := NewClaudeActionBrain()
	_, err := brain.ParseUserIntent(context.Background(), "   ", nil)
	if err == nil {
		t.Fatalf("Expected error for empty query")
	}
}
