package module

import (
	"testing"
)

func TestModuleLookupAndAliases(t *testing.T) {
	reg := GetRegistry()

	// 1. Lookup by primary name
	mod, ok := reg.Get("accounts")
	if !ok {
		t.Fatal("expected accounts module to be registered")
	}
	if mod.Status != StatusActive {
		t.Errorf("expected accounts to be active, got %s", mod.Status)
	}

	// 2. Lookup by alias ("finance")
	modByAlias, ok := reg.Get("finance")
	if !ok {
		t.Fatal("expected to find accounts module via alias 'finance'")
	}
	if modByAlias.Name != "accounts" {
		t.Errorf("expected module 'accounts', got %q", modByAlias.Name)
	}

	// 3. Lookup by buying alias ("procurement")
	buyingMod, ok := reg.Get("procurement")
	if !ok {
		t.Fatal("expected to find buying module via alias 'procurement'")
	}
	if buyingMod.Name != "buying" {
		t.Errorf("expected module 'buying', got %q", buyingMod.Name)
	}
}

func TestModuleRenameSafety(t *testing.T) {
	reg := GetRegistry()

	// Rename 'buying' to 'Strategic Sourcing' with alias 'sourcing'
	err := reg.RenameModule("buying", "Strategic Sourcing", "sourcing")
	if err != nil {
		t.Fatalf("expected successful rename, got: %v", err)
	}

	// Verify lookup via new alias
	mod, ok := reg.Get("sourcing")
	if !ok {
		t.Fatal("expected module to be found by newly registered alias 'sourcing'")
	}
	if mod.Label != "Strategic Sourcing" {
		t.Errorf("expected label 'Strategic Sourcing', got %q", mod.Label)
	}

	// Verify lookup via original name still works seamlessly
	modOrig, ok := reg.Get("buying")
	if !ok || modOrig.Name != "buying" {
		t.Fatal("expected original name 'buying' to still resolve")
	}
}

func TestModuleDisablingFaultTolerance(t *testing.T) {
	reg := GetRegistry()

	// Disable 'selling' module (e.g. maintenance or temporary outage)
	err := reg.SetStatus("selling", StatusDisabled)
	if err != nil {
		t.Fatalf("failed to set status: %v", err)
	}

	sellingMod, _ := reg.Get("selling")
	if sellingMod.Status != StatusDisabled {
		t.Errorf("expected selling to be Disabled, got %s", sellingMod.Status)
	}

	// Core accounts module MUST still be active and unaffected!
	accountsMod, ok := reg.Get("accounts")
	if !ok || accountsMod.Status != StatusActive {
		t.Fatal("core accounts module was unexpectedly affected by selling module outage")
	}

	// Core accounts module MUST NOT be allowed to be disabled
	err = reg.SetStatus("accounts", StatusDisabled)
	if err == nil {
		t.Fatal("expected error when attempting to disable core accounts module, got nil")
	}

	// Restore selling module back to active
	_ = reg.SetStatus("selling", StatusActive)
}
