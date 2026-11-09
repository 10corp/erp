package security

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserLockedOut        = errors.New("user account is temporarily locked due to excessive failed login attempts")
	ErrInvalidCredentials   = errors.New("invalid username, email or password")
	ErrUserInactive         = errors.New("user account is deactivated")
	ErrCannotDeactivateSelf = errors.New("superuser administrator cannot deactivate themselves")
	ErrUnauthorizedBranch   = errors.New("user is not authorized to access requested company branch")
	ErrInvalidToken         = errors.New("invalid or expired authentication token")
)

const (
	MaxFailedAttempts = 5
	LockoutDuration   = 15 * time.Minute
	DefaultJWTSecret  = "enterprise-bullet-speed-secret-key-2026-change-in-prod"
)

// User represents the core Tryton-style res.user identity
type User struct {
	ID                  uuid.UUID  `json:"id"`
	Username            string     `json:"username"`
	Email               string     `json:"email"`
	FullName            string     `json:"full_name"`
	PasswordHash        string     `json:"-"` // Never expose in JSON
	IsActive            bool       `json:"is_active"`
	IsSuperuser         bool       `json:"is_superuser"` // Tryton administrator (Full Model & Financial Access)
	CurrentCompanyID    *uuid.UUID `json:"current_company_id,omitempty"` // Active Branch/Company
	CurrentBranchName   string     `json:"current_branch_name,omitempty"`
	MainCompanyID       *uuid.UUID `json:"main_company_id,omitempty"`    // Root Holding Company
	EmployeeID          *uuid.UUID `json:"employee_id,omitempty"`        // company.employee -> party.party
	EmployeeCode        string     `json:"employee_code,omitempty"`
	BitmaskPermissions  Permission `json:"bitmask_permissions"`         // Bitmask: 1=Read, 2=Write, 4=Delete, 8=Post
	FailedLoginAttempts int        `json:"failed_login_attempts"`
	LockedUntil         *time.Time `json:"locked_until,omitempty"`
	LastLoginAt         *time.Time `json:"last_login_at,omitempty"`
	LastLoginIP         *string    `json:"last_login_ip,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`

	// Populated from joins
	Roles            []RoleInfo  `json:"roles,omitempty"`
	AllowedCompanies []uuid.UUID `json:"allowed_companies,omitempty"`
}

// RoleInfo represents a hierarchical role (Tryton: res.group)
type RoleInfo struct {
	ID                 uuid.UUID  `json:"id"`
	Code               string     `json:"code"`
	Name               string     `json:"name"`
	Description        string     `json:"description"`
	ParentID           *uuid.UUID `json:"parent_id,omitempty"`
	BitmaskPermissions Permission `json:"bitmask_permissions"`
	IsActive           bool       `json:"is_active"`
}

// Session represents an active authenticated session
type Session struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	TokenHash string     `json:"-"`
	IPAddress *string    `json:"ip_address,omitempty"`
	UserAgent *string    `json:"user_agent,omitempty"`
	ExpiresAt time.Time  `json:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// JWTClaims encapsulates security, active branch, holding company and employee profile
type JWTClaims struct {
	UserID             string `json:"user_id"`
	Username           string `json:"username"`
	CurrentCompanyID   string `json:"current_company_id"`
	MainCompanyID      string `json:"main_company_id"`
	EmployeeID         string `json:"employee_id,omitempty"`
	BitmaskPermissions uint8  `json:"bitmask_permissions"`
	IsSuperuser        bool   `json:"is_superuser"`
	jwt.RegisteredClaims
}

// GenerateJWT creates a cryptographically signed token with multi-branch context
func GenerateJWT(u *User, secret string, duration time.Duration) (string, error) {
	if secret == "" {
		secret = DefaultJWTSecret
	}

	currentCompStr := ""
	if u.CurrentCompanyID != nil {
		currentCompStr = u.CurrentCompanyID.String()
	}

	mainCompStr := ""
	if u.MainCompanyID != nil {
		mainCompStr = u.MainCompanyID.String()
	}

	empStr := ""
	if u.EmployeeID != nil {
		empStr = u.EmployeeID.String()
	}

	claims := JWTClaims{
		UserID:             u.ID.String(),
		Username:           u.Username,
		CurrentCompanyID:   currentCompStr,
		MainCompanyID:      mainCompStr,
		EmployeeID:         empStr,
		BitmaskPermissions: uint8(u.BitmaskPermissions),
		IsSuperuser:        u.IsSuperuser,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			Issuer:    "enterprise-erp-core",
			Subject:   u.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return signedToken, nil
}

// VerifyJWT validates a bearer token and extracts the claims
func VerifyJWT(tokenString, secret string) (*JWTClaims, error) {
	if secret == "" {
		secret = DefaultJWTSecret
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// HashPassword securely hashes plain password with bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}

// CheckPassword verifies password against stored bcrypt hash
func CheckPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// HashSessionToken computes SHA-256 hash for database token lookup
func HashSessionToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// IsLocked checks whether account is currently locked out
func (u *User) IsLocked() bool {
	if u.LockedUntil == nil {
		return false
	}
	return time.Now().UTC().Before(*u.LockedUntil)
}

// RecordFailedAttempt increments failed count and locks account if threshold exceeded
func (u *User) RecordFailedAttempt() {
	u.FailedLoginAttempts++
	if u.FailedLoginAttempts >= MaxFailedAttempts {
		lockUntil := time.Now().UTC().Add(LockoutDuration)
		u.LockedUntil = &lockUntil
	}
	u.UpdatedAt = time.Now().UTC()
}

// RecordSuccessfulLogin clears failed attempts and updates last login metadata
func (u *User) RecordSuccessfulLogin(ip string) {
	now := time.Now().UTC()
	u.FailedLoginAttempts = 0
	u.LockedUntil = nil
	u.LastLoginAt = &now
	if ip != "" {
		u.LastLoginIP = &ip
	}
	u.UpdatedAt = now
}

// HasPermission evaluates bitmask permission with superuser override (sub-microsecond)
func (u *User) HasPermission(required Permission) bool {
	if !u.IsActive {
		return false
	}
	if u.IsSuperuser {
		return true // Superuser holds all permissions (PermAdmin = 15)
	}
	return (u.BitmaskPermissions & required) == required
}

// HasCompanyAccess checks multi-company tenancy permissions (Tryton res.user-company.company)
func (u *User) HasCompanyAccess(companyID uuid.UUID) bool {
	if u.IsSuperuser {
		return true
	}
	if u.CurrentCompanyID != nil && *u.CurrentCompanyID == companyID {
		return true
	}
	for _, id := range u.AllowedCompanies {
		if id == companyID {
			return true
		}
	}
	return false
}

// SwitchBranch updates the current active branch if authorized
func (u *User) SwitchBranch(newCompanyID uuid.UUID, branchName string) error {
	if !u.HasCompanyAccess(newCompanyID) {
		return ErrUnauthorizedBranch
	}
	u.CurrentCompanyID = &newCompanyID
	u.CurrentBranchName = branchName
	u.UpdatedAt = time.Now().UTC()
	return nil
}
