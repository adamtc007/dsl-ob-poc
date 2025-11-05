package mocks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"dsl-ob-poc/internal/dictionary"
	"dsl-ob-poc/internal/store"
)

// MockStore implements a disconnected store using JSON mock data
type MockStore struct {
	loader *JSONDataLoader

	// Cached data
	cbus             []store.CBU
	roles            []store.Role
	entityTypes      []store.EntityType
	entities         []store.Entity
	limitedCompanies []store.LimitedCompany
	partnerships     []store.Partnership
	individuals      []store.Individual
	cbuEntityRoles   []store.CBUEntityRole
	products         []store.Product
	services         []store.Service
	prodResources    []store.ProdResource
	productServices  []ProductServiceRelation
	serviceResources []ServiceResourceRelation
	dictionary       []store.Attribute
	attributeValues  []AttributeValue
	dslRecords       []DSLRecord

	// In-memory tracking for dynamic DSL versions
	dynamicDSLVersions []store.DSLVersionWithState
	versionCounter     int

	loaded bool
}

// NewMockStore creates a new mock store with JSON data loader
func NewMockStore(mockDataPath string) *MockStore {
	return &MockStore{
		loader: NewJSONDataLoader(mockDataPath),
		loaded: false,
	}
}

// loadData loads all mock data if not already loaded
func (m *MockStore) loadData() error {
	if m.loaded {
		return nil
	}

	var err error

	if m.cbus, err = m.loader.LoadCBUs(); err != nil {
		return fmt.Errorf("failed to load CBUs: %w", err)
	}

	if m.roles, err = m.loader.LoadRoles(); err != nil {
		return fmt.Errorf("failed to load roles: %w", err)
	}

	if m.entityTypes, err = m.loader.LoadEntityTypes(); err != nil {
		return fmt.Errorf("failed to load entity types: %w", err)
	}

	if m.entities, err = m.loader.LoadEntities(); err != nil {
		return fmt.Errorf("failed to load entities: %w", err)
	}

	if m.limitedCompanies, err = m.loader.LoadLimitedCompanies(); err != nil {
		return fmt.Errorf("failed to load limited companies: %w", err)
	}

	if m.partnerships, err = m.loader.LoadPartnerships(); err != nil {
		return fmt.Errorf("failed to load partnerships: %w", err)
	}

	if m.individuals, err = m.loader.LoadIndividuals(); err != nil {
		return fmt.Errorf("failed to load individuals: %w", err)
	}

	if m.cbuEntityRoles, err = m.loader.LoadCBUEntityRoles(); err != nil {
		return fmt.Errorf("failed to load CBU entity roles: %w", err)
	}

	if m.products, err = m.loader.LoadProducts(); err != nil {
		return fmt.Errorf("failed to load products: %w", err)
	}

	if m.services, err = m.loader.LoadServices(); err != nil {
		return fmt.Errorf("failed to load services: %w", err)
	}

	if m.prodResources, err = m.loader.LoadProdResources(); err != nil {
		return fmt.Errorf("failed to load prod resources: %w", err)
	}

	if m.productServices, err = m.loader.LoadProductServices(); err != nil {
		return fmt.Errorf("failed to load product services: %w", err)
	}

	if m.serviceResources, err = m.loader.LoadServiceResources(); err != nil {
		return fmt.Errorf("failed to load service resources: %w", err)
	}

	if m.dictionary, err = m.loader.LoadDictionary(); err != nil {
		return fmt.Errorf("failed to load dictionary: %w", err)
	}

	if m.attributeValues, err = m.loader.LoadAttributeValues(); err != nil {
		return fmt.Errorf("failed to load attribute values: %w", err)
	}

	if m.dslRecords, err = m.loader.LoadDSLRecords(); err != nil {
		return fmt.Errorf("failed to load DSL records: %w", err)
	}

	m.loaded = true
	return nil
}

// Close does nothing for mock store
func (m *MockStore) Close() error {
	return nil
}

// CBU CRUD Operations
func (m *MockStore) ListCBUs(ctx context.Context) ([]store.CBU, error) {
	if err := m.loadData(); err != nil {
		return nil, err
	}
	return m.cbus, nil
}

func (m *MockStore) GetCBUByID(ctx context.Context, cbuID string) (*store.CBU, error) {
	if err := m.loadData(); err != nil {
		return nil, err
	}

	for _, cbu := range m.cbus {
		if cbu.CBUID == cbuID {
			return &cbu, nil
		}
	}
	return nil, fmt.Errorf("CBU not found: %s", cbuID)
}

func (m *MockStore) GetCBUByName(ctx context.Context, name string) (*store.CBU, error) {
	if err := m.loadData(); err != nil {
		return nil, err
	}

	for _, cbu := range m.cbus {
		if cbu.Name == name {
			return &cbu, nil
		}
	}
	return nil, fmt.Errorf("CBU not found: %s", name)
}

// Role CRUD Operations
func (m *MockStore) ListRoles(ctx context.Context) ([]store.Role, error) {
	if err := m.loadData(); err != nil {
		return nil, err
	}
	return m.roles, nil
}

func (m *MockStore) GetRoleByID(ctx context.Context, roleID string) (*store.Role, error) {
	if err := m.loadData(); err != nil {
		return nil, err
	}

	for _, role := range m.roles {
		if role.RoleID == roleID {
			return &role, nil
		}
	}
	return nil, fmt.Errorf("role not found: %s", roleID)
}

// Product Operations
func (m *MockStore) GetProductByName(ctx context.Context, name string) (*store.Product, error) {
	if err := m.loadData(); err != nil {
		return nil, err
	}

	for _, product := range m.products {
		if product.Name == name {
			return &product, nil
		}
	}
	return nil, fmt.Errorf("product not found: %s", name)
}

// Service Operations
func (m *MockStore) GetServicesForProduct(ctx context.Context, productID string) ([]store.Service, error) {
	if err := m.loadData(); err != nil {
		return nil, err
	}

	// Find services for this product
	var serviceIDs []string
	for _, relation := range m.productServices {
		if relation.ProductID == productID {
			serviceIDs = append(serviceIDs, relation.ServiceID)
		}
	}

	// Get service details
	var services []store.Service
	for _, service := range m.services {
		for _, serviceID := range serviceIDs {
			if service.ServiceID == serviceID {
				services = append(services, service)
				break
			}
		}
	}

	return services, nil
}

func (m *MockStore) GetServiceByName(ctx context.Context, name string) (*store.Service, error) {
	if err := m.loadData(); err != nil {
		return nil, err
	}

	for _, service := range m.services {
		if service.Name == name {
			return &service, nil
		}
	}
	return nil, fmt.Errorf("service not found: %s", name)
}

// Resource Operations
func (m *MockStore) GetResourcesForService(ctx context.Context, serviceID string) ([]store.ProdResource, error) {
	if err := m.loadData(); err != nil {
		return nil, err
	}

	// Find resources for this service
	var resourceIDs []string
	for _, relation := range m.serviceResources {
		if relation.ServiceID == serviceID {
			resourceIDs = append(resourceIDs, relation.ResourceID)
		}
	}

	// Get resource details
	var resources []store.ProdResource
	for _, resource := range m.prodResources {
		for _, resourceID := range resourceIDs {
			if resource.ResourceID == resourceID {
				resources = append(resources, resource)
				break
			}
		}
	}

	return resources, nil
}

// Orchestration session methods (mock implementations)
func (m *MockStore) SaveOrchestrationSession(ctx context.Context, session *store.OrchestrationSessionData) error {
	// In mock mode, just return success - sessions are not persisted
	return nil
}

func (m *MockStore) LoadOrchestrationSession(ctx context.Context, sessionID string) (*store.OrchestrationSessionData, error) {
	// In mock mode, return error as sessions are not persisted
	return nil, fmt.Errorf("orchestration session not found: %s", sessionID)
}

func (m *MockStore) ListActiveOrchestrationSessions(ctx context.Context) ([]string, error) {
	// In mock mode, return empty list
	return []string{}, nil
}

func (m *MockStore) DeleteOrchestrationSession(ctx context.Context, sessionID string) error {
	// In mock mode, return error as sessions don't exist
	return fmt.Errorf("session not found: %s", sessionID)
}

func (m *MockStore) CleanupExpiredOrchestrationSessions(ctx context.Context) (int64, error) {
	// In mock mode, return 0 cleaned sessions
	return 0, nil
}

func (m *MockStore) UpdateOrchestrationSessionDSL(ctx context.Context, sessionID, dsl string, version int) error {
	// In mock mode, return error as sessions don't exist
	return fmt.Errorf("session not found: %s", sessionID)
}

// Additional helper methods for mock testing
func (m *MockStore) GetAllProducts(ctx context.Context) ([]store.Product, error) {
	if err := m.loadData(); err != nil {
		return nil, err
	}
	return m.products, nil
}

func (m *MockStore) GetServicesForProducts(ctx context.Context, productNames []string) ([]store.Service, error) {
	if err := m.loadData(); err != nil {
		return nil, err
	}

	var allServices []store.Service
	serviceIDs := make(map[string]bool) // To avoid duplicates

	for _, productName := range productNames {
		// Find product ID
		var productID string
		for _, product := range m.products {
			if product.Name == productName {
				productID = product.ProductID
				break
			}
		}

		if productID != "" {
			// Get services for this product
			services, err := m.GetServicesForProduct(ctx, productID)
			if err != nil {
				continue // Skip if error
			}

			// Add unique services
			for _, service := range services {
				if !serviceIDs[service.ServiceID] {
					allServices = append(allServices, service)
					serviceIDs[service.ServiceID] = true
				}
			}
		}
	}

	return allServices, nil
}

// Dictionary Operations
func (m *MockStore) GetDictionaryAttributeByName(ctx context.Context, name string) (*dictionary.Attribute, error) {
	if err := m.loadData(); err != nil {
		return nil, err
	}

	for _, attr := range m.dictionary {
		if attr.Name == name {
			return &dictionary.Attribute{
				AttributeID:     attr.AttributeID,
				Name:            attr.Name,
				LongDescription: attr.LongDescription,
				GroupID:         attr.GroupID,
				Mask:            attr.Mask,
				Domain:          attr.Domain,
			}, nil
		}
	}
	return nil, fmt.Errorf("attribute not found: %s", name)
}

func (m *MockStore) GetDictionaryAttributeByID(ctx context.Context, id string) (*dictionary.Attribute, error) {
	if err := m.loadData(); err != nil {
		return nil, err
	}

	for _, attr := range m.dictionary {
		if attr.AttributeID == id {
			return &dictionary.Attribute{
				AttributeID:     attr.AttributeID,
				Name:            attr.Name,
				LongDescription: attr.LongDescription,
				GroupID:         attr.GroupID,
				Mask:            attr.Mask,
				Domain:          attr.Domain,
			}, nil
		}
	}
	return nil, fmt.Errorf("attribute not found: %s", id)
}

func (m *MockStore) GetAttributesForDictionaryGroup(ctx context.Context, groupID string) ([]dictionary.Attribute, error) {
	if err := m.loadData(); err != nil {
		return nil, err
	}

	var attributes []dictionary.Attribute
	for _, attr := range m.dictionary {
		if attr.GroupID == groupID {
			attributes = append(attributes, dictionary.Attribute{
				AttributeID:     attr.AttributeID,
				Name:            attr.Name,
				LongDescription: attr.LongDescription,
				GroupID:         attr.GroupID,
				Mask:            attr.Mask,
				Domain:          attr.Domain,
			})
		}
	}

	return attributes, nil
}

// DSL Operations
func (m *MockStore) GetLatestDSL(ctx context.Context, cbuID string) (string, error) {
	if err := m.loadData(); err != nil {
		return "", err
	}

	var latest DSLRecord
	var found bool

	for _, record := range m.dslRecords {
		if record.CBUID == cbuID {
			if !found {
				latest = record
				found = true
			} else {
				// Simple string comparison for mock - in real DB this would use timestamp
				if record.CreatedAt > latest.CreatedAt {
					latest = record
				}
			}
		}
	}

	if !found {
		return "", fmt.Errorf("no DSL found for CBU: %s", cbuID)
	}

	return latest.DSLText, nil
}

func (m *MockStore) InsertDSL(ctx context.Context, cbuID, dslText string) (string, error) {
	// For mock store, we don't actually insert - just return a mock version ID
	return fmt.Sprintf("mock-version-%d", time.Now().Unix()), nil
}

// ResolveValueFor provides mock attribute value resolution
func (m *MockStore) ResolveValueFor(ctx context.Context, cbuID, attributeID string) (json.RawMessage, map[string]any, string, error) {
	if err := m.loadData(); err != nil {
		return nil, nil, "", err
	}

	// Find matching attribute value
	for _, av := range m.attributeValues {
		if av.CBUID == cbuID && av.AttributeID == attributeID {
			value := json.RawMessage(av.Value)

			var source map[string]any
			if err := json.Unmarshal([]byte(av.Source), &source); err != nil {
				source = map[string]any{"type": "mock", "error": err.Error()}
			}

			return value, source, av.State, nil
		}
	}

	// Return pending state with null value if not found
	return json.RawMessage("null"), map[string]any{"reason": "no_resolver", "type": "mock"}, "pending", nil
}

// UpsertAttributeValue is a no-op for mock store
func (m *MockStore) UpsertAttributeValue(ctx context.Context, cbuID string, dslVersion int, attributeID string, value json.RawMessage, state string, source map[string]any) error {
	// For mock store, we don't actually upsert
	return nil
}

// Entity relationship methods would follow the same pattern...
// For brevity, I'm implementing the core ones needed for testing

// GetEntitiesForCBU returns entities associated with a CBU
func (m *MockStore) GetEntitiesForCBU(ctx context.Context, cbuID string) ([]store.Entity, error) {
	if err := m.loadData(); err != nil {
		return nil, err
	}

	var entities []store.Entity
	for _, relation := range m.cbuEntityRoles {
		if relation.CBUID == cbuID {
			for _, entity := range m.entities {
				if entity.EntityID == relation.EntityID {
					entities = append(entities, entity)
					break
				}
			}
		}
	}

	return entities, nil
}

// CBU CRUD Operations
func (m *MockStore) CreateCBU(ctx context.Context, name, description, naturePurpose string) (string, error) {
	// For mock store, we don't actually create - just return a mock CBU ID
	return fmt.Sprintf("mock-cbu-%d", time.Now().Unix()), nil
}

func (m *MockStore) UpdateCBU(ctx context.Context, cbuID, name, description, naturePurpose string) error {
	// For mock store, we don't actually update
	return nil
}

func (m *MockStore) DeleteCBU(ctx context.Context, cbuID string) error {
	// For mock store, we don't actually delete
	return nil
}

// Role CRUD Operations
func (m *MockStore) CreateRole(ctx context.Context, name, description string) (string, error) {
	// For mock store, we don't actually create - just return a mock role ID
	return fmt.Sprintf("mock-role-%d", time.Now().Unix()), nil
}

func (m *MockStore) UpdateRole(ctx context.Context, roleID, name, description string) error {
	// For mock store, we don't actually update
	return nil
}

func (m *MockStore) DeleteRole(ctx context.Context, roleID string) error {
	// For mock store, we don't actually delete
	return nil
}

// DSL History Operation
func (m *MockStore) GetDSLHistory(ctx context.Context, cbuID string) ([]store.DSLVersion, error) {
	if err := m.loadData(); err != nil {
		return nil, err
	}

	var history []store.DSLVersion
	for _, record := range m.dslRecords {
		if record.CBUID == cbuID {
			// Convert mock DSLRecord to store.DSLVersion
			// Parse time string (assuming RFC3339 format)
			createdAt, err := time.Parse(time.RFC3339, record.CreatedAt)
			if err != nil {
				createdAt = time.Now() // fallback
			}

			version := store.DSLVersion{
				VersionID: record.VersionID,
				DSLText:   record.DSLText,
				CreatedAt: createdAt,
			}
			history = append(history, version)
		}
	}

	return history, nil
}

// Enhanced Onboarding State Management for mock adapter
func (m *MockStore) CreateOnboardingSession(ctx context.Context, cbuID string) (*store.OnboardingSession, error) {
	// For mock store, create a mock onboarding session
	session := &store.OnboardingSession{
		OnboardingID:       fmt.Sprintf("mock-onboarding-%d", time.Now().Unix()),
		CBUID:              cbuID,
		CurrentState:       store.StateCreated,
		CurrentVersion:     1,
		LatestDSLVersionID: nil,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
	return session, nil
}

func (m *MockStore) GetOnboardingSession(ctx context.Context, cbuID string) (*store.OnboardingSession, error) {
	// For mock store, return a mock onboarding session
	session := &store.OnboardingSession{
		OnboardingID:       "mock-onboarding-session",
		CBUID:              cbuID,
		CurrentState:       store.StateCreated,
		CurrentVersion:     1,
		LatestDSLVersionID: nil,
		CreatedAt:          time.Now().Add(-24 * time.Hour),
		UpdatedAt:          time.Now(),
	}
	return session, nil
}

func (m *MockStore) UpdateOnboardingState(ctx context.Context, cbuID string, newState store.OnboardingState, dslVersionID string) error {
	// For mock store, we don't actually update - just return success
	return nil
}

func (m *MockStore) InsertDSLWithState(ctx context.Context, cbuID, dslText string, state store.OnboardingState) (string, error) {
	// Increment version counter
	m.versionCounter++

	// Create new DSL version
	versionID := fmt.Sprintf("mock-version-dynamic-%d", m.versionCounter)

	// Create DSL version record
	dslVersion := store.DSLVersionWithState{
		VersionID:       versionID,
		CBUID:           cbuID,
		DSLText:         dslText,
		OnboardingState: state,
		VersionNumber:   m.getNextVersionNumber(cbuID),
		CreatedAt:       time.Now(),
	}

	// Store in memory
	m.dynamicDSLVersions = append(m.dynamicDSLVersions, dslVersion)

	return versionID, nil
}

func (m *MockStore) GetLatestDSLWithState(ctx context.Context, cbuID string) (*store.DSLVersionWithState, error) {
	if err := m.loadData(); err != nil {
		return nil, err
	}

	// Find the latest DSL record for this CBU
	var latest DSLRecord
	var found bool

	for _, record := range m.dslRecords {
		if record.CBUID == cbuID {
			if !found {
				latest = record
				found = true
			} else {
				// Simple string comparison for mock - in real DB this would use timestamp
				if record.CreatedAt > latest.CreatedAt {
					latest = record
				}
			}
		}
	}

	if !found {
		return nil, fmt.Errorf("no DSL found for CBU: %s", cbuID)
	}

	// Parse time string
	createdAt, err := time.Parse(time.RFC3339, latest.CreatedAt)
	if err != nil {
		createdAt = time.Now() // fallback
	}

	// Create DSLVersionWithState from mock data
	dslVersion := &store.DSLVersionWithState{
		VersionID:       latest.VersionID,
		CBUID:           latest.CBUID,
		DSLText:         latest.DSLText,
		OnboardingState: store.StateCreated, // Default state for mock
		VersionNumber:   1,                  // Default version for mock
		CreatedAt:       createdAt,
	}

	return dslVersion, nil
}

func (m *MockStore) GetDSLHistoryWithState(ctx context.Context, cbuID string) ([]store.DSLVersionWithState, error) {
	if err := m.loadData(); err != nil {
		return nil, err
	}

	var history []store.DSLVersionWithState

	// First add static mock data
	version := 1
	for _, record := range m.dslRecords {
		if record.CBUID == cbuID {
			// Parse time string
			createdAt, err := time.Parse(time.RFC3339, record.CreatedAt)
			if err != nil {
				createdAt = time.Now()
			}

			history = append(history, store.DSLVersionWithState{
				VersionID:       record.VersionID,
				CBUID:           record.CBUID,
				DSLText:         record.DSLText,
				OnboardingState: store.StateCreated,
				VersionNumber:   version,
				CreatedAt:       createdAt,
			})
			version++
		}
	}

	// Add dynamic DSL versions created during runtime
	for _, dslVersion := range m.dynamicDSLVersions {
		if dslVersion.CBUID == cbuID {
			history = append(history, dslVersion)
		}
	}

	return history, nil
}

func (m *MockStore) GetDSLByVersion(ctx context.Context, cbuID string, versionNumber int) (*store.DSLVersionWithState, error) {
	if err := m.loadData(); err != nil {
		return nil, err
	}

	// For mock store, just return the first record matching CBU if version exists
	currentVersion := 1
	for _, record := range m.dslRecords {
		if record.CBUID == cbuID {
			if currentVersion == versionNumber {
				// Parse time string
				createdAt, err := time.Parse(time.RFC3339, record.CreatedAt)
				if err != nil {
					createdAt = time.Now() // fallback
				}

				dslVersion := &store.DSLVersionWithState{
					VersionID:       record.VersionID,
					CBUID:           record.CBUID,
					DSLText:         record.DSLText,
					OnboardingState: store.StateCreated, // Default state for mock
					VersionNumber:   versionNumber,
					CreatedAt:       createdAt,
				}
				return dslVersion, nil
			}
			currentVersion++
		}
	}

	return nil, fmt.Errorf("no DSL version %d found for CBU: %s", versionNumber, cbuID)
}

func (m *MockStore) ListOnboardingSessions(ctx context.Context) ([]store.OnboardingSession, error) {
	// For mock store, return a list of mock sessions
	sessions := []store.OnboardingSession{
		{
			OnboardingID:       "mock-session-1",
			CBUID:              "CBU-1234",
			CurrentState:       store.StateCreated,
			CurrentVersion:     1,
			LatestDSLVersionID: nil,
			CreatedAt:          time.Now().Add(-24 * time.Hour),
			UpdatedAt:          time.Now(),
		},
		{
			OnboardingID:       "mock-session-2",
			CBUID:              "CBU-5678",
			CurrentState:       store.StateProductsAdded,
			CurrentVersion:     2,
			LatestDSLVersionID: nil,
			CreatedAt:          time.Now().Add(-12 * time.Hour),
			UpdatedAt:          time.Now().Add(-1 * time.Hour),
		},
	}
	return sessions, nil
}

// getNextVersionNumber calculates the next version number for a CBU
func (m *MockStore) getNextVersionNumber(cbuID string) int {
	maxVersion := 0

	// Check static mock data
	for _, record := range m.dslRecords {
		if record.CBUID == cbuID {
			maxVersion++
		}
	}

	// Check dynamic versions
	for _, dslVersion := range m.dynamicDSLVersions {
		if dslVersion.CBUID == cbuID && dslVersion.VersionNumber > maxVersion {
			maxVersion = dslVersion.VersionNumber
		}
	}

	return maxVersion + 1
}

// ============================================================================
// EXPORT OPERATIONS (for mock data generation and testing)
// ============================================================================

// GetAllServices returns all services from mock data
func (m *MockStore) GetAllServices(ctx context.Context) ([]store.Service, error) {
	if err := m.loadData(); err != nil {
		return nil, err
	}
	return m.services, nil
}

// GetAllDictionaryAttributes returns all dictionary attributes from mock data
func (m *MockStore) GetAllDictionaryAttributes(ctx context.Context) ([]dictionary.Attribute, error) {
	if err := m.loadData(); err != nil {
		return nil, err
	}

	var attributes []dictionary.Attribute
	for _, attr := range m.dictionary {
		// Parse source and sink metadata from JSON strings
		var sourceMetadata dictionary.SourceMetadata
		var sinkMetadata dictionary.SinkMetadata

		if attr.Source != "" {
			_ = json.Unmarshal([]byte(attr.Source), &sourceMetadata)
		}
		if attr.Sink != "" {
			_ = json.Unmarshal([]byte(attr.Sink), &sinkMetadata)
		}

		attributes = append(attributes, dictionary.Attribute{
			AttributeID:     attr.AttributeID,
			Name:            attr.Name,
			LongDescription: attr.LongDescription,
			GroupID:         attr.GroupID,
			Mask:            attr.Mask,
			Domain:          attr.Domain,
			Vector:          attr.Vector,
			Source:          sourceMetadata,
			Sink:            sinkMetadata,
		})
	}
	return attributes, nil
}

// GetAllDSLRecords returns all DSL records with state information from mock data
func (m *MockStore) GetAllDSLRecords(ctx context.Context) ([]store.DSLVersionWithState, error) {
	if err := m.loadData(); err != nil {
		return nil, err
	}

	var records []store.DSLVersionWithState
	for i, record := range m.dslRecords {
		// Parse CreatedAt from string to time.Time
		createdAt, err := time.Parse(time.RFC3339, record.CreatedAt)
		if err != nil {
			// Fallback to current time if parsing fails
			createdAt = time.Now()
		}

		// Map mock DSL record to store format
		dslRecord := store.DSLVersionWithState{
			VersionID:       record.VersionID,
			CBUID:           record.CBUID,
			DSLText:         record.DSLText,
			OnboardingState: parseOnboardingStateFromString("CREATED"), // Default state for mock data
			VersionNumber:   i + 1,                                     // Use index as version number
			CreatedAt:       createdAt,
		}
		records = append(records, dslRecord)
	}
	return records, nil
}

// parseOnboardingStateFromString converts string state to OnboardingState enum
func parseOnboardingStateFromString(stateStr string) store.OnboardingState {
	switch stateStr {
	case "CREATED":
		return store.StateCreated
	case "PRODUCTS_ADDED":
		return store.StateProductsAdded
	case "KYC_DISCOVERED":
		return store.StateKYCDiscovered
	case "SERVICES_DISCOVERED":
		return store.StateServicesDiscovered
	case "RESOURCES_DISCOVERED":
		return store.StateResourcesDiscovered
	case "ATTRIBUTES_POPULATED":
		return store.StateAttributesPopulated
	case "COMPLETED":
		return store.StateCompleted
	default:
		return store.StateCreated
	}
}
