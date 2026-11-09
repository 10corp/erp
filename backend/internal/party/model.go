package party

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrPartyNameRequired = errors.New("party name is required")
	ErrPartyCodeRequired = errors.New("party code is required")
)

type IdentifierType string

const (
	IdentifierBIN          IdentifierType = "BIN"
	IdentifierTIN          IdentifierType = "TIN"
	IdentifierVAT          IdentifierType = "VAT"
	IdentifierNID          IdentifierType = "NID"
	IdentifierTradeLicense IdentifierType = "TRADE_LICENSE"
	IdentifierPassport     IdentifierType = "PASSPORT"
)

type ContactType string

const (
	ContactPhone    ContactType = "phone"
	ContactMobile   ContactType = "mobile"
	ContactEmail    ContactType = "email"
	ContactWebsite  ContactType = "website"
	ContactWhatsApp ContactType = "whatsapp"
	ContactOther    ContactType = "other"
)

// PartyIdentifier represents a legal tax/trade identifier (party.identifier)
type PartyIdentifier struct {
	ID        uuid.UUID      `json:"id"`
	PartyID   uuid.UUID      `json:"party_id"`
	Type      IdentifierType `json:"type"`
	Code      string         `json:"code"`
	IsPrimary bool           `json:"is_primary"`
	CreatedAt time.Time      `json:"created_at"`
}

// PartyAddress represents a physical or billing/delivery address (party.address)
type PartyAddress struct {
	ID         uuid.UUID `json:"id"`
	PartyID    uuid.UUID `json:"party_id"`
	Name       string    `json:"name,omitempty"` // Care of
	Street     string    `json:"street"`
	City       string    `json:"city"`
	PostalCode string    `json:"postal_code,omitempty"`
	Country    string    `json:"country"`
	IsInvoice  bool      `json:"is_invoice"`
	IsDelivery bool      `json:"is_delivery"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// PartyContactMechanism represents communications (party.contact_mechanism)
type PartyContactMechanism struct {
	ID        uuid.UUID   `json:"id"`
	PartyID   uuid.UUID   `json:"party_id"`
	Type      ContactType `json:"type"`
	Value     string      `json:"value"`
	IsPrimary bool        `json:"is_primary"`
	IsActive  bool        `json:"is_active"`
	CreatedAt time.Time   `json:"created_at"`
}

// Party represents the Universal Legal Entity (party.party)
type Party struct {
	ID         uuid.UUID `json:"id"`
	Code       string    `json:"code"`
	Name       string    `json:"name"`
	IsCustomer bool      `json:"is_customer"`
	IsSupplier bool      `json:"is_supplier"`
	IsEmployee bool      `json:"is_employee"`
	IsCompany  bool      `json:"is_company"`
	Lang       string    `json:"lang"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// Relational nested data
	Identifiers []PartyIdentifier       `json:"identifiers,omitempty"`
	Addresses   []PartyAddress          `json:"addresses,omitempty"`
	Contacts    []PartyContactMechanism `json:"contacts,omitempty"`
}

// CompanyHierarchy represents Tryton's Multi-Branch & Subsidiary Tree
type CompanyHierarchy struct {
	CompanyID       uuid.UUID  `json:"company_id"`
	CompanyName     string     `json:"company_name"`
	RootCompanyID   uuid.UUID  `json:"root_company_id"`
	RootCompanyName string     `json:"root_company_name"`
	ParentID        *uuid.UUID `json:"parent_id,omitempty"`
	IsBranch        bool       `json:"is_branch"`
	BranchCode      *string    `json:"branch_code,omitempty"`
	HierarchyLevel  int        `json:"hierarchy_level"`
	PathNames       []string   `json:"path_names"`
}

// Validate checks core invariants
func (p *Party) Validate() error {
	if p.Name == "" {
		return ErrPartyNameRequired
	}
	if p.Code == "" {
		return ErrPartyCodeRequired
	}
	return nil
}
