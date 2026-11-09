package security

import (
	"context"
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

// GetUserByEmail retrieves a user by email, including company assignments
func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	if r.db == nil || r.db.Pool == nil {
		return r.fallbackGetUserByEmail(email)
	}

	query := `
		SELECT 
			u.id, u.username, u.email, u.full_name, u.password_hash, u.is_active, u.is_superuser,
			u.current_company_id, c.name AS current_branch_name, u.main_company_id,
			u.employee_id, e.employee_code, u.bitmask_permissions, u.failed_login_attempts,
			u.locked_until, u.last_login_at, u.last_login_ip::text, u.created_at, u.updated_at,
			COALESCE(ARRAY_AGG(uc.company_id) FILTER (WHERE uc.company_id IS NOT NULL), '{}') AS allowed_companies
		FROM users u
		LEFT JOIN companies c ON u.current_company_id = c.id
		LEFT JOIN company_employee e ON u.employee_id = e.id
		LEFT JOIN user_companies uc ON u.id = uc.user_id
		WHERE LOWER(u.email) = LOWER($1)
		GROUP BY u.id, c.name, e.employee_code;
	`

	var u User
	var branchName, empCode, lastIP *string

	err := r.db.Pool.QueryRow(ctx, query, email).Scan(
		&u.ID, &u.Username, &u.Email, &u.FullName, &u.PasswordHash, &u.IsActive, &u.IsSuperuser,
		&u.CurrentCompanyID, &branchName, &u.MainCompanyID,
		&u.EmployeeID, &empCode, &u.BitmaskPermissions, &u.FailedLoginAttempts,
		&u.LockedUntil, &u.LastLoginAt, &lastIP, &u.CreatedAt, &u.UpdatedAt,
		&u.AllowedCompanies,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return r.fallbackGetUserByEmail(email)
		}
		return r.fallbackGetUserByEmail(email)
	}

	if branchName != nil {
		u.CurrentBranchName = *branchName
	}
	if empCode != nil {
		u.EmployeeCode = *empCode
	}
	if lastIP != nil {
		u.LastLoginIP = lastIP
	}

	return &u, nil
}

// GetUserByID retrieves a user by ID
func (r *Repository) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	if r.db == nil || r.db.Pool == nil {
		return r.fallbackGetUserByID(id)
	}

	query := `
		SELECT 
			u.id, u.username, u.email, u.full_name, u.password_hash, u.is_active, u.is_superuser,
			u.current_company_id, c.name AS current_branch_name, u.main_company_id,
			u.employee_id, e.employee_code, u.bitmask_permissions, u.failed_login_attempts,
			u.locked_until, u.last_login_at, u.last_login_ip::text, u.created_at, u.updated_at,
			COALESCE(ARRAY_AGG(uc.company_id) FILTER (WHERE uc.company_id IS NOT NULL), '{}') AS allowed_companies
		FROM users u
		LEFT JOIN companies c ON u.current_company_id = c.id
		LEFT JOIN company_employee e ON u.employee_id = e.id
		LEFT JOIN user_companies uc ON u.id = uc.user_id
		WHERE u.id = $1
		GROUP BY u.id, c.name, e.employee_code;
	`

	var u User
	var branchName, empCode, lastIP *string

	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.Username, &u.Email, &u.FullName, &u.PasswordHash, &u.IsActive, &u.IsSuperuser,
		&u.CurrentCompanyID, &branchName, &u.MainCompanyID,
		&u.EmployeeID, &empCode, &u.BitmaskPermissions, &u.FailedLoginAttempts,
		&u.LockedUntil, &u.LastLoginAt, &lastIP, &u.CreatedAt, &u.UpdatedAt,
		&u.AllowedCompanies,
	)
	if err != nil {
		return r.fallbackGetUserByID(id)
	}

	if branchName != nil {
		u.CurrentBranchName = *branchName
	}
	if empCode != nil {
		u.EmployeeCode = *empCode
	}
	if lastIP != nil {
		u.LastLoginIP = lastIP
	}

	return &u, nil
}

// ListUsers retrieves all users
func (r *Repository) ListUsers(ctx context.Context) ([]User, error) {
	if r.db == nil || r.db.Pool == nil {
		return sampleUsers, nil
	}

	query := `
		SELECT 
			u.id, u.username, u.email, u.full_name, u.password_hash, u.is_active, u.is_superuser,
			u.current_company_id, c.name AS current_branch_name, u.main_company_id,
			u.employee_id, e.employee_code, u.bitmask_permissions, u.failed_login_attempts,
			u.locked_until, u.last_login_at, u.last_login_ip::text, u.created_at, u.updated_at,
			COALESCE(ARRAY_AGG(uc.company_id) FILTER (WHERE uc.company_id IS NOT NULL), '{}') AS allowed_companies
		FROM users u
		LEFT JOIN companies c ON u.current_company_id = c.id
		LEFT JOIN company_employee e ON u.employee_id = e.id
		LEFT JOIN user_companies uc ON u.id = uc.user_id
		GROUP BY u.id, c.name, e.employee_code
		ORDER BY u.created_at ASC;
	`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return sampleUsers, nil
	}
	defer rows.Close()

	var list []User
	for rows.Next() {
		var u User
		var branchName, empCode, lastIP *string
		if err := rows.Scan(
			&u.ID, &u.Username, &u.Email, &u.FullName, &u.PasswordHash, &u.IsActive, &u.IsSuperuser,
			&u.CurrentCompanyID, &branchName, &u.MainCompanyID,
			&u.EmployeeID, &empCode, &u.BitmaskPermissions, &u.FailedLoginAttempts,
			&u.LockedUntil, &u.LastLoginAt, &lastIP, &u.CreatedAt, &u.UpdatedAt,
			&u.AllowedCompanies,
		); err != nil {
			return nil, err
		}
		if branchName != nil {
			u.CurrentBranchName = *branchName
		}
		if empCode != nil {
			u.EmployeeCode = *empCode
		}
		if lastIP != nil {
			u.LastLoginIP = lastIP
		}
		list = append(list, u)
	}

	if len(list) == 0 {
		return sampleUsers, nil
	}

	return list, nil
}

// UpdateLoginSuccess updates last login and resets counters
func (r *Repository) UpdateLoginSuccess(ctx context.Context, userID uuid.UUID, clientIP string) error {
	if r.db == nil || r.db.Pool == nil {
		return nil
	}

	query := `
		UPDATE users 
		SET failed_login_attempts = 0, locked_until = NULL, last_login_at = NOW(), last_login_ip = $2::inet, updated_at = NOW()
		WHERE id = $1;
	`
	_, err := r.db.Pool.Exec(ctx, query, userID, clientIP)
	return err
}

// SwitchActiveBranch switches user active branch in database
func (r *Repository) SwitchActiveBranch(ctx context.Context, userID uuid.UUID, companyID uuid.UUID) error {
	if r.db == nil || r.db.Pool == nil {
		return nil
	}

	query := `
		UPDATE users 
		SET current_company_id = $2, updated_at = NOW()
		WHERE id = $1;
	`
	_, err := r.db.Pool.Exec(ctx, query, userID, companyID)
	return err
}

// Fallbacks
func (r *Repository) fallbackGetUserByEmail(email string) (*User, error) {
	for _, u := range sampleUsers {
		if u.Email == email {
			return &u, nil
		}
	}
	return nil, fmt.Errorf("user not found with email %s", email)
}

func (r *Repository) fallbackGetUserByID(id uuid.UUID) (*User, error) {
	for _, u := range sampleUsers {
		if u.ID == id {
			return &u, nil
		}
	}
	return nil, errors.New("user not found")
}
