package party

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPartyValidation(t *testing.T) {
	// Missing name
	p1 := &Party{
		Code: "PRTY-001",
	}
	if err := p1.Validate(); err != ErrPartyNameRequired {
		t.Errorf("expected ErrPartyNameRequired, got: %v", err)
	}

	// Missing code
	p2 := &Party{
		Name: "Global Trade Corp",
	}
	if err := p2.Validate(); err != ErrPartyCodeRequired {
		t.Errorf("expected ErrPartyCodeRequired, got: %v", err)
	}

	// Valid party
	p3 := &Party{
		ID:         uuid.New(),
		Code:       "PRTY-003",
		Name:       "Dhaka Distribution Ltd",
		IsCustomer: true,
	}
	if err := p3.Validate(); err != nil {
		t.Errorf("expected valid party to pass, got: %v", err)
	}
}

func TestPartyIdentifiersAndAddresses(t *testing.T) {
	partyID := uuid.New()

	p := &Party{
		ID:         partyID,
		Code:       "PRTY-CORP-99",
		Name:       "Silicon Horizon Tech",
		IsCustomer: true,
		IsSupplier: true,
		Lang:       "en",
		IsActive:   true,
		Identifiers: []PartyIdentifier{
			{
				ID:        uuid.New(),
				PartyID:   partyID,
				Type:      IdentifierBIN,
				Code:      "001827462-0101",
				IsPrimary: true,
			},
			{
				ID:        uuid.New(),
				PartyID:   partyID,
				Type:      IdentifierTIN,
				Code:      "881928471928",
				IsPrimary: false,
			},
		},
		Addresses: []PartyAddress{
			{
				ID:        uuid.New(),
				PartyID:   partyID,
				Street:    "Plot 12, Banani Road 11",
				City:      "Dhaka",
				Country:   "Bangladesh",
				IsInvoice: true,
				IsDelivery: true,
				IsActive:  true,
				CreatedAt: time.Now().UTC(),
			},
		},
	}

	if len(p.Identifiers) != 2 {
		t.Fatalf("expected 2 identifiers, got %d", len(p.Identifiers))
	}
	if p.Identifiers[0].Type != IdentifierBIN || !p.Identifiers[0].IsPrimary {
		t.Errorf("expected primary BIN identifier, got %s", p.Identifiers[0].Type)
	}
	if len(p.Addresses) != 1 || p.Addresses[0].City != "Dhaka" {
		t.Errorf("expected address city 'Dhaka', got %s", p.Addresses[0].City)
	}
}

func TestCompanyMultiBranchHierarchy(t *testing.T) {
	rootID := uuid.New()
	branchID := uuid.New()
	branchCode := "DHK-BR01"

	root := CompanyHierarchy{
		CompanyID:       rootID,
		CompanyName:     "Holding Corp",
		RootCompanyID:   rootID,
		RootCompanyName: "Holding Corp",
		IsBranch:        false,
		HierarchyLevel:  1,
		PathNames:       []string{"Holding Corp"},
	}

	branch := CompanyHierarchy{
		CompanyID:       branchID,
		CompanyName:     "Dhaka Branch",
		RootCompanyID:   rootID,
		RootCompanyName: "Holding Corp",
		ParentID:        &rootID,
		IsBranch:        true,
		BranchCode:      &branchCode,
		HierarchyLevel:  2,
		PathNames:       []string{"Holding Corp", "Dhaka Branch"},
	}

	if branch.HierarchyLevel != 2 {
		t.Errorf("expected hierarchy level 2, got %d", branch.HierarchyLevel)
	}
	if branch.RootCompanyID != root.CompanyID {
		t.Errorf("expected branch root company %s, got %s", root.CompanyID, branch.RootCompanyID)
	}
	if len(branch.PathNames) != 2 || branch.PathNames[1] != "Dhaka Branch" {
		t.Errorf("expected path name 'Dhaka Branch', got %v", branch.PathNames)
	}
}
