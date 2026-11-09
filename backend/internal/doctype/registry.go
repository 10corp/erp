package doctype

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// FieldType defines the data type of an ERPNext-style DocType field
type FieldType string

const (
	TypeData     FieldType = "Data"
	TypeLink     FieldType = "Link" // Foreign key to another DocType
	TypeCurrency FieldType = "Currency"
	TypeFloat    FieldType = "Float"
	TypeInt      FieldType = "Int"
	TypeDate     FieldType = "Date"
	TypeSelect   FieldType = "Select"
	TypeTable    FieldType = "Table" // Child DocType table
	TypeCheck    FieldType = "Check" // Boolean
	TypeJSON     FieldType = "JSON"
)

// DocField describes a single field in the DocType metadata
type DocField struct {
	FieldName string    `json:"fieldname"`
	Label     string    `json:"label"`
	FieldType FieldType `json:"fieldtype"`
	Options   string    `json:"options,omitempty"` // Target DocType if TypeLink
	Required  bool      `json:"reqd"`
	ReadOnly  bool      `json:"read_only"`
	Unique    bool      `json:"unique"`
	Default   string    `json:"default,omitempty"`
}

// DocTypeDefinition is the runtime representation of an ERPNext-style DocType schema
type DocTypeDefinition struct {
	Name         string              `json:"name"`
	Module       string              `json:"module"`
	IsSubmittable bool               `json:"is_submittable"` // Tryton/ERPNext submission workflow
	TableName    string              `json:"table_name"`
	Fields       map[string]DocField `json:"fields"`
	FieldOrder   []string            `json:"field_order"`
}

// Registry stores all registered DocTypes in memory
type Registry struct {
	mu       sync.RWMutex
	docTypes map[string]*DocTypeDefinition
}

var globalRegistry = &Registry{
	docTypes: make(map[string]*DocTypeDefinition),
}

func GetRegistry() *Registry {
	return globalRegistry
}

func (r *Registry) Register(dt *DocTypeDefinition) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.docTypes[dt.Name] = dt
}

func (r *Registry) Get(name string) (*DocTypeDefinition, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	dt, ok := r.docTypes[name]
	return dt, ok
}

// TrytonDocStatus represents the strict Tryton/ERPNext workflow state machine
type TrytonDocStatus int

const (
	DocDraft     TrytonDocStatus = 0 // Editable
	DocSubmitted TrytonDocStatus = 1 // Posted & Immutable (GAAP/IFRS compliant)
	DocCancelled TrytonDocStatus = 2 // Reversed via contra entry
)

// BaseDocument is embedded in all Go structs to support ERPNext schema + Tryton state machine
type BaseDocument struct {
	ID        uuid.UUID       `json:"id"`
	DocStatus TrytonDocStatus `json:"docstatus"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
}

// Money wraps shopspring/decimal for strict zero-float financial precision
type Money = decimal.Decimal
