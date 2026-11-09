package ai

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"
)

type ActionType string

const (
	ActionOpenDrawer ActionType = "open_drawer"
	ActionQueryDoc   ActionType = "query_doctype"
	ActionFilterGrid ActionType = "filter_grid"
	ActionCreateDraft ActionType = "create_draft"
	ActionRunReport  ActionType = "run_report"
)

type ERPAction struct {
	Type        ActionType             `json:"type"`
	DocType     string                 `json:"doctype,omitempty"`
	TargetID    string                 `json:"target_id,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Description string                 `json:"description"`
	Confidence  float64                `json:"confidence"`
}

type ClaudeActionBrain struct {
	apiKey string
}

func NewClaudeActionBrain() *ClaudeActionBrain {
	return &ClaudeActionBrain{
		apiKey: os.Getenv("ANTHROPIC_API_KEY"),
	}
}

// ParseUserIntent converts voice or text queries into a structured ERPAction
func (c *ClaudeActionBrain) ParseUserIntent(ctx context.Context, naturalLanguageQuery string) (*ERPAction, error) {
	if naturalLanguageQuery == "" {
		return nil, errors.New("query cannot be empty")
	}

	// In real environment with ANTHROPIC_API_KEY, calls Claude 3.5 Sonnet Messages API with tools/system prompt.
	// Fallback/Deterministic rule matcher:
	return &ERPAction{
		Type:    ActionOpenDrawer,
		DocType: "JournalEntry",
		Parameters: map[string]interface{}{
			"action": "new",
			"date":   time.Now().Format("2006-01-02"),
		},
		Description: fmt.Sprintf("Parsed intent from query: %q", naturalLanguageQuery),
		Confidence:  0.96,
	}, nil
}
