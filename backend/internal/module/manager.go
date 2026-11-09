package module

import (
	"fmt"
	"sync"

	"github.com/enterprise-erp/core/internal/doctype"
)

type Status string

const (
	StatusActive   Status = "Active"
	StatusDisabled Status = "Disabled"
	StatusDegraded Status = "Degraded"
)

// ModuleInfo provides metadata and health status for an ERP module
type ModuleInfo struct {
	Name         string   `json:"name"`          // System identifier (e.g., "accounts", "selling", "buying", "stock")
	Label        string   `json:"label"`         // User-facing display name (can be renamed anytime!)
	Aliases      []string `json:"aliases"`       // Renamed historical names for backward compatibility
	Version      string   `json:"version"`
	Status       Status   `json:"status"`        // Active, Disabled, Degraded
	DocTypes     []string `json:"doctypes"`      // DocTypes governed by this module
	Dependencies []string `json:"dependencies"`  // Required modules (e.g., selling requires accounts)
	IsCore       bool     `json:"is_core"`       // Core modules cannot be disabled (e.g. accounts)
}

// Registry manages the lifecycle of all modular pluggable subsystems
type Registry struct {
	mu      sync.RWMutex
	modules map[string]*ModuleInfo
}

var globalModules = &Registry{
	modules: make(map[string]*ModuleInfo),
}

func GetRegistry() *Registry {
	return globalModules
}

func init() {
	// Register standard ERPNext-style default modular structure
	globalModules.Register(&ModuleInfo{
		Name:     "accounts",
		Label:    "Finance & Accounting",
		Aliases:  []string{"finance", "accounting"},
		Version:  "1.0.0",
		Status:   StatusActive,
		IsCore:   true, // Core financial invariant engine
		DocTypes: []string{"Company", "Account", "GL Entry", "Sales Invoice", "Payment Entry", "Purchase Invoice"},
	})

	globalModules.Register(&ModuleInfo{
		Name:     "stock",
		Label:    "Inventory & Warehouses",
		Aliases:  []string{"inventory"},
		Version:  "1.0.0",
		Status:   StatusActive,
		IsCore:   false,
		DocTypes: []string{"Warehouse", "Item Warehouse Balance", "Stock Ledger Entry", "Delivery Note"},
	})

	globalModules.Register(&ModuleInfo{
		Name:     "selling",
		Label:    "Sales & Revenue",
		Aliases:  []string{"sales"},
		Version:  "1.0.0",
		Status:   StatusActive,
		IsCore:   false,
		DocTypes: []string{"Sales Order", "Customer"},
	})

	globalModules.Register(&ModuleInfo{
		Name:     "buying",
		Label:    "Procure-to-Pay (Purchasing)",
		Aliases:  []string{"procurement", "purchasing"},
		Version:  "1.0.0",
		Status:   StatusActive,
		IsCore:   false,
		DocTypes: []string{"Purchase Order", "Supplier"},
	})

	globalModules.Register(&ModuleInfo{
		Name:     "ai",
		Label:    "Multimodal AI Intelligence",
		Aliases:  []string{"intelligence", "copilot"},
		Version:  "1.0.0",
		Status:   StatusActive,
		IsCore:   false,
		DocTypes: []string{"IDP Scan Job", "Action Command"},
	})

	globalModules.Register(&ModuleInfo{
		Name:     "hr",
		Label:    "Human Resources & Payroll",
		Aliases:  []string{"hrms", "payroll", "people"},
		Version:  "1.0.0",
		Status:   StatusActive,
		IsCore:   false,
		DocTypes: []string{"Department", "Designation", "Leave Application", "Salary Structure", "Salary Slip"},
	})

	globalModules.Register(&ModuleInfo{
		Name:     "manufacturing",
		Label:    "Manufacturing & Production",
		Aliases:  []string{"production", "mrp"},
		Version:  "1.0.0",
		Status:   StatusActive,
		IsCore:   false,
		DocTypes: []string{"Bill of Materials", "BOM Item", "Work Order"},
	})
}

// Register adds or updates a module in the registry
func (r *Registry) Register(m *ModuleInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.modules[m.Name] = m
}

// Get finds a module by its primary name or any historical renamed alias
func (r *Registry) Get(nameOrAlias string) (*ModuleInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 1. Direct lookup by primary name
	if m, ok := r.modules[nameOrAlias]; ok {
		return m, true
	}

	// 2. Lookup by alias (if renamed)
	for _, m := range r.modules {
		for _, alias := range m.Aliases {
			if alias == nameOrAlias {
				return m, true
			}
		}
	}

	return nil, false
}

// RenameModule safely changes a module's user-facing label and registers the old name as alias
func (r *Registry) RenameModule(currentName, newLabel, optionalNewAlias string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	m, ok := r.modules[currentName]
	if !ok {
		return fmt.Errorf("module %q not found", currentName)
	}

	m.Label = newLabel
	if optionalNewAlias != "" {
		m.Aliases = append(m.Aliases, optionalNewAlias)
	}
	return nil
}

// SetStatus enables, disables, or marks a module degraded without affecting the core ERP
func (r *Registry) SetStatus(name string, status Status) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	m, ok := r.modules[name]
	if !ok {
		return fmt.Errorf("module %q not found", name)
	}

	if m.IsCore && status == StatusDisabled {
		return fmt.Errorf("core module %q cannot be disabled", name)
	}

	m.Status = status
	return nil
}

// ListActive returns only modules that are currently healthy and active
func (r *Registry) ListActive() []*ModuleInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var active []*ModuleInfo
	for _, m := range r.modules {
		if m.Status == StatusActive {
			active = append(active, m)
		}
	}
	return active
}

// ListAll returns all registered modules
func (r *Registry) ListAll() []*ModuleInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var all []*ModuleInfo
	for _, m := range r.modules {
		all = append(all, m)
	}
	return all
}

// IsDocTypeAvailable checks if the module owning a DocType is currently online
func (r *Registry) IsDocTypeAvailable(docTypeName string) bool {
	dt, ok := doctype.GetRegistry().Get(docTypeName)
	if !ok {
		return false
	}

	mod, ok := r.Get(dt.Module)
	if !ok {
		// If no module constraint explicitly declared, default available
		return true
	}

	return mod.Status == StatusActive
}
