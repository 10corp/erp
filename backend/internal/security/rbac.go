package security

import (
	"context"
	"errors"
	"net/http"
)

// Permission defines bitmask rights inspired by Hexya's security model
type Permission uint8

const (
	PermRead   Permission = 1 << iota // 1
	PermWrite                         // 2
	PermDelete                        // 4
	PermPost                          // 8
	PermAdmin  Permission = PermRead | PermWrite | PermDelete | PermPost
)

// Role definitions
type Role string

const (
	RoleAdmin     Role = "admin"
	RoleAccountant Role = "accountant"
	RoleAuditor   Role = "auditor"
	RoleManager   Role = "manager"
	RoleStaff     Role = "staff"
)

// Default role permission mapping
var RolePermissions = map[Role]Permission{
	RoleAdmin:      PermAdmin,
	RoleAccountant: PermRead | PermWrite | PermPost,
	RoleManager:    PermRead | PermWrite,
	RoleAuditor:    PermRead,
	RoleStaff:      PermRead,
}

type UserContextKey struct{}

type UserSession struct {
	UserID    string
	CompanyID string
	Role      Role
}

func (u *UserSession) HasPermission(required Permission) bool {
	perms, ok := RolePermissions[u.Role]
	if !ok {
		return false
	}
	return (perms & required) == required
}

var ErrForbidden = errors.New("insufficient permissions for this enterprise operation")

// RequirePermission HTTP middleware ensuring user has the required bitmask permission
func RequirePermission(required Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, ok := r.Context().Value(UserContextKey{}).(*UserSession)
			// In development mode or unauthenticated fallback, allow or reject based on header
			if !ok || session == nil {
				// Check for dev bypass or default accountant session
				role := Role(r.Header.Get("X-User-Role"))
				if role == "" {
					role = RoleAdmin // default to Admin for local dev
				}
				session = &UserSession{
					UserID:    r.Header.Get("X-User-ID"),
					CompanyID: r.Header.Get("X-Company-ID"),
					Role:      role,
				}
				ctx := context.WithValue(r.Context(), UserContextKey{}, session)
				r = r.WithContext(ctx)
			}

			if !session.HasPermission(required) {
				http.Error(w, `{"error":"Forbidden: insufficient permissions"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
