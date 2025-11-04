package datastore

import (
	"encoding/json"
)

// DataStore interface for web server operations
type DataStore interface {
	Close() error
}

// MockDataStore is a minimal in-memory implementation
type MockDataStore struct{}

// NewMockDataStore creates a new mock datastore
func NewMockDataStore() *MockDataStore {
	return &MockDataStore{}
}

// Close implements DataStore
func (m *MockDataStore) Close() error {
	return nil
}

// DSLState represents a DSL execution state (for future use)
type DSLState struct {
	ID            string          `json:"id"`
	InvestorID    string          `json:"investor_id"`
	DSLText       string          `json:"dsl_text"`
	State         string          `json:"state"`
	VersionNumber int             `json:"version_number"`
	CreatedAt     string          `json:"created_at"`
	Parameters    json.RawMessage `json:"parameters,omitempty"`
}
