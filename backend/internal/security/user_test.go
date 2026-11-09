package security

import (
	"testing"

	"github.com/google/uuid"
)

func TestPasswordHashingAndVerification(t *testing.T) {
	password := "StrongEnterprisePass#2026"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	if !CheckPassword(hash, password) {
		t.Error("expected valid password check to return true")
	}

	if CheckPassword(hash, "WrongPassword") {
		t.Error("expected incorrect password check to return false")
	}
}

func TestBitmaskPermissionEvaluation(t *testing.T) {
	// Standard User with Read + Write (3)
	user := &User{
		Username:           "sales_officer",
		IsActive:           true,
		IsSuperuser:        false,
		BitmaskPermissions: PermRead | PermWrite, // 3
	}

	if !user.HasPermission(PermRead) {
		t.Error("expected user to have PermRead")
	}
	if !user.HasPermission(PermWrite) {
		t.Error("expected user to have PermWrite")
	}
	if user.HasPermission(PermDelete) {
		t.Error("expected user NOT to have PermDelete")
	}
	if user.HasPermission(PermPost) {
		t.Error("expected user NOT to have PermPost")
	}

	// Superuser Administrator (Tryton pattern) - MUST have all permissions
	admin := &User{
		Username:           "cfo_admin",
		IsActive:           true,
		IsSuperuser:        true,
		BitmaskPermissions: PermRead, // Even if raw bitmask is 1, superuser overrides to 15
	}

	if !admin.HasPermission(PermPost) {
		t.Error("expected superuser administrator to have PermPost")
	}
	if !admin.HasPermission(PermDelete) {
		t.Error("expected superuser administrator to have PermDelete")
	}

	// Inactive user MUST NOT have any permission
	inactiveUser := &User{
		Username:           "deactivated",
		IsActive:           false,
		IsSuperuser:        false,
		BitmaskPermissions: PermAdmin,
	}

	if inactiveUser.HasPermission(PermRead) {
		t.Error("expected deactivated user to have NO permissions")
	}
}

func TestBruteForceAccountLockout(t *testing.T) {
	user := &User{
		Username: "employee_01",
		IsActive: true,
	}

	if user.IsLocked() {
		t.Fatal("user should initially not be locked")
	}

	// 4 failed attempts should not lock yet
	for i := 0; i < 4; i++ {
		user.RecordFailedAttempt()
	}
	if user.IsLocked() {
		t.Fatal("user should not be locked after 4 failed attempts")
	}

	// 5th failed attempt MUST trigger lockout
	user.RecordFailedAttempt()
	if !user.IsLocked() {
		t.Fatal("user MUST be locked after 5 failed attempts")
	}

	// Successful login MUST clear lock
	user.RecordSuccessfulLogin("192.168.1.100")
	if user.IsLocked() {
		t.Fatal("successful login must clear account lockout")
	}
	if user.FailedLoginAttempts != 0 {
		t.Errorf("expected failed attempts to reset to 0, got %d", user.FailedLoginAttempts)
	}
	if user.LastLoginIP == nil || *user.LastLoginIP != "192.168.1.100" {
		t.Error("expected LastLoginIP to be recorded")
	}
}

func TestMultiCompanyAccess(t *testing.T) {
	primaryCompany := uuid.New()
	subsidiaryCompany := uuid.New()
	unauthorizedCompany := uuid.New()

	user := &User{
		Username:         "regional_manager",
		CurrentCompanyID: &primaryCompany,
		AllowedCompanies: []uuid.UUID{primaryCompany, subsidiaryCompany},
		IsSuperuser:      false,
	}

	if !user.HasCompanyAccess(primaryCompany) {
		t.Error("expected access to primary company")
	}
	if !user.HasCompanyAccess(subsidiaryCompany) {
		t.Error("expected access to subsidiary company")
	}
	if user.HasCompanyAccess(unauthorizedCompany) {
		t.Error("expected NO access to unauthorized company")
	}
}

func TestSessionTokenHashing(t *testing.T) {
	token := "bearer-token-random-secret-xyz"
	hash1 := HashSessionToken(token)
	hash2 := HashSessionToken(token)

	if hash1 != hash2 {
		t.Error("hash of same token must be deterministic")
	}
	if len(hash1) != 64 {
		t.Errorf("expected 64 char hex string for SHA-256, got %d", len(hash1))
	}
}
