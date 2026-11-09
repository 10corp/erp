package party

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/enterprise-erp/core/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	db   *database.DB
	repo *Repository
}

func NewHandler(db *database.DB) *Handler {
	return &Handler{
		db:   db,
		repo: NewRepository(db),
	}
}

// Sample in-memory parties for fast fallback when DB container is booting
var sampleParties = []Party{
	{
		ID:         uuid.MustParse("00000000-0000-0000-0000-000000000101"),
		Code:       "PRTY-CUST-001",
		Name:       "Acme Holdings Global Ltd",
		IsCustomer: true,
		IsSupplier: false,
		IsEmployee: false,
		IsCompany:  false,
		Lang:       "en",
		IsActive:   true,
		CreatedAt:  time.Now().UTC().AddDate(0, -6, 0),
		UpdatedAt:  time.Now().UTC(),
		Identifiers: []PartyIdentifier{
			{
				ID:        uuid.New(),
				Type:      IdentifierBIN,
				Code:      "002918274-0101",
				IsPrimary: true,
			},
		},
		Addresses: []PartyAddress{
			{
				ID:        uuid.New(),
				Name:      "HQ Accounts Dept",
				Street:    "Plot 42, Gulshan Avenue, Road 11",
				City:      "Dhaka",
				Country:   "Bangladesh",
				IsInvoice: true,
				IsDelivery: true,
				IsActive:  true,
			},
		},
		Contacts: []PartyContactMechanism{
			{
				ID:        uuid.New(),
				Type:      ContactEmail,
				Value:     "billing@acmeholdings.com",
				IsPrimary: true,
				IsActive:  true,
			},
			{
				ID:        uuid.New(),
				Type:      ContactPhone,
				Value:     "+880 1711 000000",
				IsPrimary: true,
				IsActive:  true,
			},
		},
	},
	{
		ID:         uuid.MustParse("00000000-0000-0000-0000-000000000102"),
		Code:       "PRTY-SUPP-001",
		Name:       "Silicon Valley Chipsets Inc",
		IsCustomer: false,
		IsSupplier: true,
		IsEmployee: false,
		IsCompany:  false,
		Lang:       "en",
		IsActive:   true,
		CreatedAt:  time.Now().UTC().AddDate(0, -8, 0),
		UpdatedAt:  time.Now().UTC(),
		Identifiers: []PartyIdentifier{
			{
				ID:        uuid.New(),
				Type:      IdentifierVAT,
				Code:      "US-948271049",
				IsPrimary: true,
			},
		},
		Addresses: []PartyAddress{
			{
				ID:        uuid.New(),
				Street:    "440 Innovation Way, Suite 800",
				City:      "San Jose",
				Country:   "United States",
				IsInvoice: true,
				IsDelivery: false,
				IsActive:  true,
			},
		},
		Contacts: []PartyContactMechanism{
			{
				ID:        uuid.New(),
				Type:      ContactEmail,
				Value:     "orders@svchipsets.com",
				IsPrimary: true,
				IsActive:  true,
			},
		},
	},
	{
		ID:         uuid.MustParse("00000000-0000-0000-0000-000000000103"),
		Code:       "PRTY-CORP-001",
		Name:       "Enterprise ERP Global Ltd (Holding Company)",
		IsCustomer: false,
		IsSupplier: false,
		IsEmployee: false,
		IsCompany:  true,
		Lang:       "en",
		IsActive:   true,
		CreatedAt:  time.Now().UTC().AddDate(-1, 0, 0),
		UpdatedAt:  time.Now().UTC(),
	},
}

var sampleHierarchy = []CompanyHierarchy{
	{
		CompanyID:       uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		CompanyName:     "Enterprise ERP Global Ltd (HQ)",
		RootCompanyID:   uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		RootCompanyName: "Enterprise ERP Global Ltd (HQ)",
		ParentID:        nil,
		IsBranch:        false,
		HierarchyLevel:  1,
		PathNames:       []string{"Enterprise ERP Global Ltd (HQ)"},
	},
	{
		CompanyID:       uuid.MustParse("00000000-0000-0000-0000-000000000002"),
		CompanyName:     "Dhaka Principal Branch",
		RootCompanyID:   uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		RootCompanyName: "Enterprise ERP Global Ltd (HQ)",
		ParentID:        nil,
		IsBranch:        true,
		HierarchyLevel:  2,
		PathNames:       []string{"Enterprise ERP Global Ltd (HQ)", "Dhaka Principal Branch"},
	},
	{
		CompanyID:       uuid.MustParse("00000000-0000-0000-0000-000000000003"),
		CompanyName:     "Chittagong Port Distribution Branch",
		RootCompanyID:   uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		RootCompanyName: "Enterprise ERP Global Ltd (HQ)",
		ParentID:        nil,
		IsBranch:        true,
		HierarchyLevel:  2,
		PathNames:       []string{"Enterprise ERP Global Ltd (HQ)", "Chittagong Port Distribution Branch"},
	},
}

// ListParties handles GET /api/v1/parties
func (h *Handler) ListParties(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	roleFilter := r.URL.Query().Get("role") // "customer", "supplier", "employee", "company"

	parties, err := h.repo.ListParties(r.Context(), roleFilter)
	if err != nil {
		http.Error(w, "failed to query parties: "+err.Error(), http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(parties)
}

// GetPartyByID handles GET /api/v1/parties/{id}
func (h *Handler) GetPartyByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	targetID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid party id", http.StatusBadRequest)
		return
	}

	party, err := h.repo.GetPartyByID(r.Context(), targetID)
	if err != nil {
		http.Error(w, "party not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(party)
}

// CreateParty handles POST /api/v1/parties
func (h *Handler) CreateParty(w http.ResponseWriter, r *http.Request) {
	var input Party
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := input.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	input.ID = uuid.New()
	input.CreatedAt = time.Now().UTC()
	input.UpdatedAt = time.Now().UTC()
	input.IsActive = true

	sampleParties = append(sampleParties, input)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(input)
}

// GetCompanyHierarchy handles GET /api/v1/companies/hierarchy (Multi-Branch Tree)
func (h *Handler) GetCompanyHierarchy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	hierarchy, err := h.repo.GetCompanyHierarchy(r.Context())
	if err != nil {
		http.Error(w, "failed to fetch company hierarchy: "+err.Error(), http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(hierarchy)
}
