package datastore

import (
	"context"
	"encoding/json"

	"dsl-ob-poc/internal/dictionary"
	"dsl-ob-poc/internal/mocks"
	"dsl-ob-poc/internal/store"
)

// DataStore defines the interface for all data access operations
// This interface can be implemented by both real database store and mock store
type DataStore interface {
	// Lifecycle
	Close() error

	// CBU Operations
	ListCBUs(ctx context.Context) ([]store.CBU, error)
	GetCBUByID(ctx context.Context, cbuID string) (*store.CBU, error)
	GetCBUByName(ctx context.Context, name string) (*store.CBU, error)
	CreateCBU(ctx context.Context, name, description, naturePurpose string) (string, error)
	UpdateCBU(ctx context.Context, cbuID, name, description, naturePurpose string) error
	DeleteCBU(ctx context.Context, cbuID string) error

	// Role Operations
	ListRoles(ctx context.Context) ([]store.Role, error)
	GetRoleByID(ctx context.Context, roleID string) (*store.Role, error)
	CreateRole(ctx context.Context, name, description string) (string, error)
	UpdateRole(ctx context.Context, roleID, name, description string) error
	DeleteRole(ctx context.Context, roleID string) error

	// Product Operations
	GetProductByName(ctx context.Context, name string) (*store.Product, error)

	// Service Operations
	GetServicesForProduct(ctx context.Context, productID string) ([]store.Service, error)
	GetServiceByName(ctx context.Context, name string) (*store.Service, error)

	// Resource Operations
	GetResourcesForService(ctx context.Context, serviceID string) ([]store.ProdResource, error)

	// Dictionary Operations
	GetDictionaryAttributeByName(ctx context.Context, name string) (*dictionary.Attribute, error)
	GetDictionaryAttributeByID(ctx context.Context, id string) (*dictionary.Attribute, error)
	GetAttributesForDictionaryGroup(ctx context.Context, groupID string) ([]dictionary.Attribute, error)

	// DSL Operations
	GetLatestDSL(ctx context.Context, cbuID string) (string, error)
	InsertDSL(ctx context.Context, cbuID, dslText string) (string, error)
	GetDSLHistory(ctx context.Context, cbuID string) ([]store.DSLVersion, error)

	// Attribute Value Operations
	ResolveValueFor(ctx context.Context, cbuID, attributeID string) (json.RawMessage, map[string]any, string, error)
	UpsertAttributeValue(ctx context.Context, cbuID string, dslVersion int, attributeID string, value json.RawMessage, state string, source map[string]any) error

	// Catalog Seeding (for database initialization)
	SeedCatalog(ctx context.Context) error
	InitDB(ctx context.Context) error
}

// DataStoreType represents the type of data store to use
type DataStoreType string

const (
	// PostgreSQLStore uses real PostgreSQL database
	PostgreSQLStore DataStoreType = "postgresql"
	// MockStore uses JSON mock data
	MockStore DataStoreType = "mock"
)

// Config holds configuration for data store creation
type Config struct {
	Type             DataStoreType
	ConnectionString string
	MockDataPath     string
}

// NewDataStore creates a new data store based on configuration
func NewDataStore(config Config) (DataStore, error) {
	switch config.Type {
	case PostgreSQLStore:
		return newPostgreSQLStore(config.ConnectionString)
	case MockStore:
		return newMockStore(config.MockDataPath)
	default:
		return nil, &UnsupportedStoreTypeError{Type: string(config.Type)}
	}
}

// newPostgreSQLStore creates a new PostgreSQL store adapter
func newPostgreSQLStore(connectionString string) (DataStore, error) {
	store, err := store.NewStore(connectionString)
	if err != nil {
		return nil, err
	}
	return &postgresAdapter{store: store}, nil
}

// newMockStore creates a new mock store adapter
func newMockStore(mockDataPath string) (DataStore, error) {
	mockStore := mocks.NewMockStore(mockDataPath)
	return &mockAdapter{store: mockStore}, nil
}

// UnsupportedStoreTypeError is returned when an unsupported store type is requested
type UnsupportedStoreTypeError struct {
	Type string
}

func (e *UnsupportedStoreTypeError) Error() string {
	return "unsupported store type: " + e.Type
}

// postgresAdapter adapts the PostgreSQL store to the DataStore interface
type postgresAdapter struct {
	store *store.Store
}

func (p *postgresAdapter) Close() error {
	return p.store.Close()
}

func (p *postgresAdapter) ListCBUs(ctx context.Context) ([]store.CBU, error) {
	return p.store.ListCBUs(ctx)
}

func (p *postgresAdapter) GetCBUByID(ctx context.Context, cbuID string) (*store.CBU, error) {
	return p.store.GetCBUByID(ctx, cbuID)
}

func (p *postgresAdapter) GetCBUByName(ctx context.Context, name string) (*store.CBU, error) {
	return p.store.GetCBUByName(ctx, name)
}

func (p *postgresAdapter) CreateCBU(ctx context.Context, name, description, naturePurpose string) (string, error) {
	return p.store.CreateCBU(ctx, name, description, naturePurpose)
}

func (p *postgresAdapter) UpdateCBU(ctx context.Context, cbuID, name, description, naturePurpose string) error {
	return p.store.UpdateCBU(ctx, cbuID, name, description, naturePurpose)
}

func (p *postgresAdapter) DeleteCBU(ctx context.Context, cbuID string) error {
	return p.store.DeleteCBU(ctx, cbuID)
}

func (p *postgresAdapter) ListRoles(ctx context.Context) ([]store.Role, error) {
	return p.store.ListRoles(ctx)
}

func (p *postgresAdapter) GetRoleByID(ctx context.Context, roleID string) (*store.Role, error) {
	return p.store.GetRoleByID(ctx, roleID)
}

func (p *postgresAdapter) CreateRole(ctx context.Context, name, description string) (string, error) {
	return p.store.CreateRole(ctx, name, description)
}

func (p *postgresAdapter) UpdateRole(ctx context.Context, roleID, name, description string) error {
	return p.store.UpdateRole(ctx, roleID, name, description)
}

func (p *postgresAdapter) DeleteRole(ctx context.Context, roleID string) error {
	return p.store.DeleteRole(ctx, roleID)
}

func (p *postgresAdapter) GetProductByName(ctx context.Context, name string) (*store.Product, error) {
	return p.store.GetProductByName(ctx, name)
}

func (p *postgresAdapter) GetServicesForProduct(ctx context.Context, productID string) ([]store.Service, error) {
	return p.store.GetServicesForProduct(ctx, productID)
}

func (p *postgresAdapter) GetServiceByName(ctx context.Context, name string) (*store.Service, error) {
	return p.store.GetServiceByName(ctx, name)
}

func (p *postgresAdapter) GetResourcesForService(ctx context.Context, serviceID string) ([]store.ProdResource, error) {
	return p.store.GetResourcesForService(ctx, serviceID)
}

func (p *postgresAdapter) GetDictionaryAttributeByName(ctx context.Context, name string) (*dictionary.Attribute, error) {
	return p.store.GetDictionaryAttributeByName(ctx, name)
}

func (p *postgresAdapter) GetDictionaryAttributeByID(ctx context.Context, id string) (*dictionary.Attribute, error) {
	return p.store.GetDictionaryAttributeByID(ctx, id)
}

func (p *postgresAdapter) GetAttributesForDictionaryGroup(ctx context.Context, groupID string) ([]dictionary.Attribute, error) {
	return p.store.GetAttributesForDictionaryGroup(ctx, groupID)
}

func (p *postgresAdapter) GetLatestDSL(ctx context.Context, cbuID string) (string, error) {
	return p.store.GetLatestDSL(ctx, cbuID)
}

func (p *postgresAdapter) InsertDSL(ctx context.Context, cbuID, dslText string) (string, error) {
	return p.store.InsertDSL(ctx, cbuID, dslText)
}

func (p *postgresAdapter) GetDSLHistory(ctx context.Context, cbuID string) ([]store.DSLVersion, error) {
	return p.store.GetDSLHistory(ctx, cbuID)
}

func (p *postgresAdapter) ResolveValueFor(ctx context.Context, cbuID, attributeID string) (json.RawMessage, map[string]any, string, error) {
	return p.store.ResolveValueFor(ctx, cbuID, attributeID)
}

func (p *postgresAdapter) UpsertAttributeValue(ctx context.Context, cbuID string, dslVersion int, attributeID string, value json.RawMessage, state string, source map[string]any) error {
	return p.store.UpsertAttributeValue(ctx, cbuID, dslVersion, attributeID, value, state, source)
}

func (p *postgresAdapter) SeedCatalog(ctx context.Context) error {
	return p.store.SeedCatalog(ctx)
}

func (p *postgresAdapter) InitDB(ctx context.Context) error {
	return p.store.InitDB(ctx)
}

// mockAdapter adapts the mock store to the DataStore interface
type mockAdapter struct {
	store *mocks.MockStore
}

func (m *mockAdapter) Close() error {
	return m.store.Close()
}

func (m *mockAdapter) ListCBUs(ctx context.Context) ([]store.CBU, error) {
	return m.store.ListCBUs(ctx)
}

func (m *mockAdapter) GetCBUByID(ctx context.Context, cbuID string) (*store.CBU, error) {
	return m.store.GetCBUByID(ctx, cbuID)
}

func (m *mockAdapter) GetCBUByName(ctx context.Context, name string) (*store.CBU, error) {
	return m.store.GetCBUByName(ctx, name)
}

func (m *mockAdapter) CreateCBU(ctx context.Context, name, description, naturePurpose string) (string, error) {
	return m.store.CreateCBU(ctx, name, description, naturePurpose)
}

func (m *mockAdapter) UpdateCBU(ctx context.Context, cbuID, name, description, naturePurpose string) error {
	return m.store.UpdateCBU(ctx, cbuID, name, description, naturePurpose)
}

func (m *mockAdapter) DeleteCBU(ctx context.Context, cbuID string) error {
	return m.store.DeleteCBU(ctx, cbuID)
}

func (m *mockAdapter) ListRoles(ctx context.Context) ([]store.Role, error) {
	return m.store.ListRoles(ctx)
}

func (m *mockAdapter) GetRoleByID(ctx context.Context, roleID string) (*store.Role, error) {
	return m.store.GetRoleByID(ctx, roleID)
}

func (m *mockAdapter) CreateRole(ctx context.Context, name, description string) (string, error) {
	return m.store.CreateRole(ctx, name, description)
}

func (m *mockAdapter) UpdateRole(ctx context.Context, roleID, name, description string) error {
	return m.store.UpdateRole(ctx, roleID, name, description)
}

func (m *mockAdapter) DeleteRole(ctx context.Context, roleID string) error {
	return m.store.DeleteRole(ctx, roleID)
}

func (m *mockAdapter) GetProductByName(ctx context.Context, name string) (*store.Product, error) {
	return m.store.GetProductByName(ctx, name)
}

func (m *mockAdapter) GetServicesForProduct(ctx context.Context, productID string) ([]store.Service, error) {
	return m.store.GetServicesForProduct(ctx, productID)
}

func (m *mockAdapter) GetServiceByName(ctx context.Context, name string) (*store.Service, error) {
	return m.store.GetServiceByName(ctx, name)
}

func (m *mockAdapter) GetResourcesForService(ctx context.Context, serviceID string) ([]store.ProdResource, error) {
	return m.store.GetResourcesForService(ctx, serviceID)
}

func (m *mockAdapter) GetDictionaryAttributeByName(ctx context.Context, name string) (*dictionary.Attribute, error) {
	return m.store.GetDictionaryAttributeByName(ctx, name)
}

func (m *mockAdapter) GetDictionaryAttributeByID(ctx context.Context, id string) (*dictionary.Attribute, error) {
	return m.store.GetDictionaryAttributeByID(ctx, id)
}

func (m *mockAdapter) GetAttributesForDictionaryGroup(ctx context.Context, groupID string) ([]dictionary.Attribute, error) {
	return m.store.GetAttributesForDictionaryGroup(ctx, groupID)
}

func (m *mockAdapter) GetLatestDSL(ctx context.Context, cbuID string) (string, error) {
	return m.store.GetLatestDSL(ctx, cbuID)
}

func (m *mockAdapter) InsertDSL(ctx context.Context, cbuID, dslText string) (string, error) {
	return m.store.InsertDSL(ctx, cbuID, dslText)
}

func (m *mockAdapter) GetDSLHistory(ctx context.Context, cbuID string) ([]store.DSLVersion, error) {
	return m.store.GetDSLHistory(ctx, cbuID)
}

func (m *mockAdapter) ResolveValueFor(ctx context.Context, cbuID, attributeID string) (json.RawMessage, map[string]any, string, error) {
	return m.store.ResolveValueFor(ctx, cbuID, attributeID)
}

func (m *mockAdapter) UpsertAttributeValue(ctx context.Context, cbuID string, dslVersion int, attributeID string, value json.RawMessage, state string, source map[string]any) error {
	return m.store.UpsertAttributeValue(ctx, cbuID, dslVersion, attributeID, value, state, source)
}

func (m *mockAdapter) SeedCatalog(ctx context.Context) error {
	return nil // Mock store doesn't need seeding
}

func (m *mockAdapter) InitDB(ctx context.Context) error {
	return nil // Mock store doesn't need DB initialization
}
