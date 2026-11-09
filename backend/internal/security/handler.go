package security

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/enterprise-erp/core/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type UserHandler struct {
	db   *database.DB
	repo *Repository
}

func NewUserHandler(db *database.DB) *UserHandler {
	return &UserHandler{
		db:   db,
		repo: NewRepository(db),
	}
}

var (
	rootCompID   = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	branch1ID    = uuid.MustParse("00000000-0000-0000-0000-000000000002")
	branch2ID    = uuid.MustParse("00000000-0000-0000-0000-000000000003")
	empTuhinID   = uuid.MustParse("00000000-0000-0000-0000-000000000077")
	empRahimID   = uuid.MustParse("00000000-0000-0000-0000-000000000088")
)

// Sample users list populated with Tryton Multi-Branch and Employee bindings
var sampleUsers = []User{
	{
		ID:                  uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		Username:            "admin",
		Email:               "admin@enterprise-erp.local",
		FullName:            "Tuhin (System Administrator & CFO)",
		PasswordHash:        "$2a$10$w857aNfU67v6k2j2q6X.g.Z2yK4/X6s6h3Y2N5u2a1Q6r9K6b9X2W", // "password123"
		IsActive:            true,
		IsSuperuser:         true,
		MainCompanyID:       &rootCompID,
		CurrentCompanyID:    &rootCompID,
		CurrentBranchName:   "Enterprise ERP Global Ltd (HQ)",
		EmployeeID:          &empTuhinID,
		EmployeeCode:        "EMP-EXEC-001",
		BitmaskPermissions:  PermAdmin, // 15
		AllowedCompanies:    []uuid.UUID{rootCompID, branch1ID, branch2ID},
		CreatedAt:           time.Now().UTC().AddDate(0, -6, 0),
		UpdatedAt:           time.Now().UTC(),
	},
	{
		ID:                  uuid.MustParse("00000000-0000-0000-0000-000000000002"),
		Username:            "sales_officer",
		Email:               "sales@enterprise-erp.local",
		FullName:            "Rahim (Senior Sales Officer)",
		PasswordHash:        "$2a$10$w857aNfU67v6k2j2q6X.g.Z2yK4/X6s6h3Y2N5u2a1Q6r9K6b9X2W",
		IsActive:            true,
		IsSuperuser:         false,
		MainCompanyID:       &rootCompID,
		CurrentCompanyID:    &branch1ID,
		CurrentBranchName:   "Dhaka Principal Branch",
		EmployeeID:          &empRahimID,
		EmployeeCode:        "EMP-SALES-042",
		BitmaskPermissions:  PermRead | PermWrite, // 3
		AllowedCompanies:    []uuid.UUID{branch1ID},
		CreatedAt:           time.Now().UTC().AddDate(0, -3, 0),
		UpdatedAt:           time.Now().UTC(),
	},
	{
		ID:                  uuid.MustParse("00000000-0000-0000-0000-000000000003"),
		Username:            "auditor_visitor",
		Email:               "visitor@enterprise-erp.local",
		FullName:            "External Financial Auditor",
		PasswordHash:        "$2a$10$w857aNfU67v6k2j2q6X.g.Z2yK4/X6s6h3Y2N5u2a1Q6r9K6b9X2W",
		IsActive:            true,
		IsSuperuser:         false,
		MainCompanyID:       &rootCompID,
		CurrentCompanyID:    &rootCompID,
		CurrentBranchName:   "Enterprise ERP Global Ltd (HQ)",
		BitmaskPermissions:  PermRead, // 1
		AllowedCompanies:    []uuid.UUID{rootCompID, branch1ID, branch2ID},
		CreatedAt:           time.Now().UTC().AddDate(0, -1, 0),
		UpdatedAt:           time.Now().UTC(),
	},
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token     string `json:"token"`
	User      User   `json:"user"`
	ExpiresIn int64  `json:"expires_in"` // Seconds
}

type SwitchBranchRequest struct {
	CompanyID  uuid.UUID `json:"company_id"`
	BranchName string    `json:"branch_name"`
}

// Login handles POST /api/v1/auth/login
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	for i := range sampleUsers {
		u := &sampleUsers[i]
		if u.Email == req.Email || u.Username == req.Email {
			if u.IsLocked() {
				http.Error(w, ErrUserLockedOut.Error(), http.StatusLocked)
				return
			}

			// In demo mode, accept any non-empty password or valid hash
			if req.Password != "" {
				u.RecordSuccessfulLogin(r.RemoteAddr)

				token, err := GenerateJWT(u, DefaultJWTSecret, 24*time.Hour)
				if err != nil {
					http.Error(w, "failed to issue authentication token", http.StatusInternalServerError)
					return
				}

				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(LoginResponse{
					Token:     token,
					User:      *u,
					ExpiresIn: 86400,
				})
				return
			}
		}
	}

	http.Error(w, ErrInvalidCredentials.Error(), http.StatusUnauthorized)
}

// SwitchBranch handles POST /api/v1/users/switch-branch
func (h *UserHandler) SwitchBranch(w http.ResponseWriter, r *http.Request) {
	var req SwitchBranchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	// In production, user is read from JWT context; here we update active admin session
	adminUser := &sampleUsers[0]
	if err := adminUser.SwitchBranch(req.CompanyID, req.BranchName); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	newToken, err := GenerateJWT(adminUser, DefaultJWTSecret, 24*time.Hour)
	if err != nil {
		http.Error(w, "failed to regenerate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"token":        newToken,
		"active_branch": req.BranchName,
		"company_id":   req.CompanyID,
	})
}

// ListUsers returns all users in the system
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	users, err := h.repo.ListUsers(r.Context())
	if err != nil {
		http.Error(w, "failed to query users: "+err.Error(), http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(users)
}

// ToggleUserStatus activates or deactivates a user
func (h *UserHandler) ToggleUserStatus(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "id")
	targetID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	for i := range sampleUsers {
		if sampleUsers[i].ID == targetID {
			if sampleUsers[i].IsSuperuser {
				http.Error(w, ErrCannotDeactivateSelf.Error(), http.StatusForbidden)
				return
			}
			sampleUsers[i].IsActive = !sampleUsers[i].IsActive
			sampleUsers[i].UpdatedAt = time.Now().UTC()
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(sampleUsers[i])
			return
		}
	}

	http.Error(w, "user not found", http.StatusNotFound)
}
