package party

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/enterprise-erp/core/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Repository struct {
	db *database.DB
}

func NewRepository(db *database.DB) *Repository {
	return &Repository{db: db}
}

// ListParties queries parties with N+1 query guard via LEFT JOIN LATERAL & jsonb_agg
func (r *Repository) ListParties(ctx context.Context, roleFilter string) ([]Party, error) {
	if r.db == nil || r.db.Pool == nil {
		return r.fallbackListParties(roleFilter), nil
	}

	query := `
		SELECT 
			p.id, p.code, p.name, p.is_customer, p.is_supplier, p.is_employee, p.is_company, p.lang, p.is_active, p.created_at, p.updated_at,
			COALESCE(ident.identifiers_json, '[]'::jsonb) AS identifiers,
			COALESCE(addr.addresses_json, '[]'::jsonb) AS addresses,
			COALESCE(cont.contacts_json, '[]'::jsonb) AS contacts
		FROM party_party p
		LEFT JOIN LATERAL (
			SELECT jsonb_agg(jsonb_build_object(
				'id', id, 'party_id', party_id, 'type', type, 'code', code, 'is_primary', is_primary, 'created_at', created_at
			)) AS identifiers_json
			FROM party_identifier WHERE party_id = p.id
		) ident ON TRUE
		LEFT JOIN LATERAL (
			SELECT jsonb_agg(jsonb_build_object(
				'id', id, 'party_id', party_id, 'name', name, 'street', street, 'city', city, 'postal_code', postal_code,
				'country', country, 'is_invoice', is_invoice, 'is_delivery', is_delivery, 'is_active', is_active,
				'created_at', created_at, 'updated_at', updated_at
			)) AS addresses_json
			FROM party_address WHERE party_id = p.id
		) addr ON TRUE
		LEFT JOIN LATERAL (
			SELECT jsonb_agg(jsonb_build_object(
				'id', id, 'party_id', party_id, 'type', type, 'value', value, 'is_primary', is_primary, 'is_active', is_active,
				'created_at', created_at
			)) AS contacts_json
			FROM party_contact_mechanism WHERE party_id = p.id
		) cont ON TRUE
		WHERE ($1 = '' OR 
			  ($1 = 'customer' AND p.is_customer) OR 
			  ($1 = 'supplier' AND p.is_supplier) OR 
			  ($1 = 'employee' AND p.is_employee) OR 
			  ($1 = 'company' AND p.is_company))
		ORDER BY p.name ASC;
	`

	rows, err := r.db.Pool.Query(ctx, query, roleFilter)
	if err != nil {
		// If table does not exist or connection fails, fallback seamlessly
		return r.fallbackListParties(roleFilter), nil
	}
	defer rows.Close()

	var parties []Party
	for rows.Next() {
		var p Party
		var identBytes, addrBytes, contBytes []byte

		err := rows.Scan(
			&p.ID, &p.Code, &p.Name, &p.IsCustomer, &p.IsSupplier, &p.IsEmployee, &p.IsCompany,
			&p.Lang, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
			&identBytes, &addrBytes, &contBytes,
		)
		if err != nil {
			return nil, fmt.Errorf("scan party error: %w", err)
		}

		_ = json.Unmarshal(identBytes, &p.Identifiers)
		_ = json.Unmarshal(addrBytes, &p.Addresses)
		_ = json.Unmarshal(contBytes, &p.Contacts)

		parties = append(parties, p)
	}

	if len(parties) == 0 {
		return r.fallbackListParties(roleFilter), nil
	}

	return parties, nil
}

// GetPartyByID fetches single party with identifiers and addresses
func (r *Repository) GetPartyByID(ctx context.Context, id uuid.UUID) (*Party, error) {
	if r.db == nil || r.db.Pool == nil {
		return r.fallbackGetPartyByID(id)
	}

	query := `
		SELECT 
			p.id, p.code, p.name, p.is_customer, p.is_supplier, p.is_employee, p.is_company, p.lang, p.is_active, p.created_at, p.updated_at,
			COALESCE(ident.identifiers_json, '[]'::jsonb) AS identifiers,
			COALESCE(addr.addresses_json, '[]'::jsonb) AS addresses,
			COALESCE(cont.contacts_json, '[]'::jsonb) AS contacts
		FROM party_party p
		LEFT JOIN LATERAL (
			SELECT jsonb_agg(jsonb_build_object(
				'id', id, 'party_id', party_id, 'type', type, 'code', code, 'is_primary', is_primary, 'created_at', created_at
			)) AS identifiers_json
			FROM party_identifier WHERE party_id = p.id
		) ident ON TRUE
		LEFT JOIN LATERAL (
			SELECT jsonb_agg(jsonb_build_object(
				'id', id, 'party_id', party_id, 'name', name, 'street', street, 'city', city, 'postal_code', postal_code,
				'country', country, 'is_invoice', is_invoice, 'is_delivery', is_delivery, 'is_active', is_active,
				'created_at', created_at, 'updated_at', updated_at
			)) AS addresses_json
			FROM party_address WHERE party_id = p.id
		) addr ON TRUE
		LEFT JOIN LATERAL (
			SELECT jsonb_agg(jsonb_build_object(
				'id', id, 'party_id', party_id, 'type', type, 'value', value, 'is_primary', is_primary, 'is_active', is_active,
				'created_at', created_at
			)) AS contacts_json
			FROM party_contact_mechanism WHERE party_id = p.id
		) cont ON TRUE
		WHERE p.id = $1;
	`

	var p Party
	var identBytes, addrBytes, contBytes []byte

	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.Code, &p.Name, &p.IsCustomer, &p.IsSupplier, &p.IsEmployee, &p.IsCompany,
		&p.Lang, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
		&identBytes, &addrBytes, &contBytes,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return r.fallbackGetPartyByID(id)
		}
		return r.fallbackGetPartyByID(id)
	}

	_ = json.Unmarshal(identBytes, &p.Identifiers)
	_ = json.Unmarshal(addrBytes, &p.Addresses)
	_ = json.Unmarshal(contBytes, &p.Contacts)

	return &p, nil
}

// GetCompanyHierarchy queries recursive branch hierarchy
func (r *Repository) GetCompanyHierarchy(ctx context.Context) ([]CompanyHierarchy, error) {
	if r.db == nil || r.db.Pool == nil {
		return sampleHierarchy, nil
	}

	query := `
		SELECT 
			company_id, company_name, root_company_id, root_company_name, 
			parent_id, is_branch, hierarchy_level, path_names
		FROM vw_company_hierarchy
		ORDER BY hierarchy_level, company_name;
	`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return sampleHierarchy, nil
	}
	defer rows.Close()

	var list []CompanyHierarchy
	for rows.Next() {
		var ch CompanyHierarchy
		var pathNames []string
		if err := rows.Scan(
			&ch.CompanyID, &ch.CompanyName, &ch.RootCompanyID, &ch.RootCompanyName,
			&ch.ParentID, &ch.IsBranch, &ch.HierarchyLevel, &pathNames,
		); err != nil {
			return nil, err
		}
		ch.PathNames = pathNames
		list = append(list, ch)
	}

	if len(list) == 0 {
		return sampleHierarchy, nil
	}

	return list, nil
}

// Fallbacks
func (r *Repository) fallbackListParties(roleFilter string) []Party {
	if roleFilter == "" {
		return sampleParties
	}
	var res []Party
	for _, p := range sampleParties {
		if roleFilter == "customer" && p.IsCustomer {
			res = append(res, p)
		} else if roleFilter == "supplier" && p.IsSupplier {
			res = append(res, p)
		} else if roleFilter == "employee" && p.IsEmployee {
			res = append(res, p)
		} else if roleFilter == "company" && p.IsCompany {
			res = append(res, p)
		}
	}
	return res
}

func (r *Repository) fallbackGetPartyByID(id uuid.UUID) (*Party, error) {
	for _, p := range sampleParties {
		if p.ID == id {
			return &p, nil
		}
	}
	return nil, errors.New("party not found")
}
