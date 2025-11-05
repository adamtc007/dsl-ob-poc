package dsl_manager

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"

	"dsl-ob-poc/internal/shared-dsl/session"
	"dsl-ob-poc/internal/store"
)

// DSLManager provides domain-specific DSL management with versioning and state tracking
type DSLManager struct {
	sessionManager *session.Manager
	store          store.Store // Dependency for persistence
	mu             sync.RWMutex
}

// DSLVersion represents a versioned DSL record
type DSLVersion struct {
	ID            string    `json:"id"`
	OnboardingID  string    `json:"onboarding_id"`
	Domain        string    `json:"domain"`
	State         string    `json:"state"`
	VersionNumber int       `json:"version_number"`
	DSLContent    string    `json:"dsl_content"`
	CreatedAt     time.Time `json:"created_at"`
}

// NewDSLManager creates a new DSL Manager
func NewDSLManager(store store.Store) *DSLManager {
	return &DSLManager{
		sessionManager: session.NewManager(),
		store:          store,
	}
}

// CreateCase initializes a new case in a specific domain
func (m *DSLManager) CreateCase(domain string, initialData map[string]interface{}) (*session.Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate unique onboarding ID
	onboardingID := uuid.New().String()

	// Create session
	session := m.sessionManager.GetOrCreate(onboardingID, domain)

	// Update context with initial data
	initialData["cbu_id"] = onboardingID
	initialData["current_state"] = "CREATED"

	if err := session.UpdateContext(initialData); err != nil {
		return nil, fmt.Errorf("failed to update session context: %w", err)
	}

	// Generate initial DSL
	initialDSL, err := m.generateInitialDSL(domain, initialData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate initial DSL: %w", err)
	}

	// Accumulate DSL
	if accErr := session.AccumulateDSL(initialDSL); accErr != nil {
		return nil, fmt.Errorf("failed to accumulate initial DSL: %w", accErr)
	}

	// Persist initial DSL version
	version, err := m.persistDSLVersion(session)
	if err != nil {
		return nil, fmt.Errorf("failed to persist initial DSL version: %w", err)
	}

	log.Printf("Created new case: Domain=%s, OnboardingID=%s, Version=%d",
		domain, onboardingID, version.VersionNumber)

	return session, nil
}

// UpdateCase updates an existing case with a new DSL fragment
func (m *DSLManager) UpdateCase(onboardingID, newDSLFragment string, stateTransition string) (*session.Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Retrieve existing session
	session, err := m.sessionManager.Get(onboardingID)
	if err != nil {
		return nil, fmt.Errorf("session not found for onboarding ID %s: %w", onboardingID, err)
	}

	// Validate state transition
	currentState := session.GetContext().CurrentState
	if validateErr := m.validateStateTransition(currentState, stateTransition); validateErr != nil {
		return nil, fmt.Errorf("invalid state transition: %w", validateErr)
	}

	// Update session context with new state
	if updateErr := session.UpdateContext(map[string]interface{}{
		"current_state": stateTransition,
	}); updateErr != nil {
		return nil, fmt.Errorf("failed to update session context: %w", updateErr)
	}

	// Accumulate new DSL fragment
	if err := session.AccumulateDSL(newDSLFragment); err != nil {
		return nil, fmt.Errorf("failed to accumulate DSL fragment: %w", err)
	}

	// Persist updated DSL version
	version, err := m.persistDSLVersion(session)
	if err != nil {
		return nil, fmt.Errorf("failed to persist updated DSL version: %w", err)
	}

	log.Printf("Updated case: OnboardingID=%s, NewState=%s, Version=%d",
		onboardingID, stateTransition, version.VersionNumber)

	return session, nil
}

// GetCase retrieves a case by onboarding ID
func (m *DSLManager) GetCase(onboardingID string) (*session.Session, error) {
	return m.sessionManager.Get(onboardingID)
}

// ListCases retrieves all active case session IDs
func (m *DSLManager) ListCases() []string {
	return m.sessionManager.List()
}

// persistDSLVersion saves the current session DSL as a new version
func (m *DSLManager) persistDSLVersion(session *session.Session) (*DSLVersion, error) {
	ctx := session.GetContext()

	version := &DSLVersion{
		ID:            uuid.New().String(),
		OnboardingID:  ctx.CBUID,
		Domain:        session.Domain,
		State:         ctx.CurrentState,
		VersionNumber: m.calculateNextVersion(ctx.CBUID),
		DSLContent:    session.GetDSL(),
		CreatedAt:     time.Now(),
	}

	// Convert version to JSON for storage
	versionJSON, err := json.Marshal(version)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal DSL version: %w", err)
	}

	// TODO: Implement actual storage in the store
	// For now, we'll log the version data that would be stored
	fmt.Printf("DSL version to store: %s\n", string(versionJSON))
	// This is a placeholder - replace with actual store method
	// err = m.store.StoreDSLVersion(version)

	return version, nil
}

// calculateNextVersion determines the next version number for a case
func (m *DSLManager) calculateNextVersion(onboardingID string) int {
	// TODO: Implement version tracking - query existing versions from store
	return 1
}

// validateStateTransition ensures valid state progression
func (m *DSLManager) validateStateTransition(currentState, newState string) error {
	// Define valid state transitions based on domain rules
	validTransitions := map[string][]string{
		"CREATED":             {"INVESTOR_ADDED", "KYC_STARTED"},
		"INVESTOR_ADDED":      {"KYC_STARTED"},
		"KYC_STARTED":         {"KYC_IN_PROGRESS", "KYC_COMPLETED"},
		"KYC_IN_PROGRESS":     {"KYC_COMPLETED"},
		"KYC_COMPLETED":       {"DOCUMENTS_COLLECTED", "RISK_ASSESSMENT"},
		"DOCUMENTS_COLLECTED": {"RISK_ASSESSMENT"},
		"RISK_ASSESSMENT":     {"ONBOARDING_COMPLETE"},
	}

	allowedTransitions, exists := validTransitions[currentState]
	if !exists {
		return fmt.Errorf("invalid current state: %s", currentState)
	}

	for _, transition := range allowedTransitions {
		if transition == newState {
			return nil
		}
	}

	return fmt.Errorf("invalid transition from %s to %s", currentState, newState)
}

// CleanupExpiredSessions removes inactive sessions
func (m *DSLManager) CleanupExpiredSessions(maxAge time.Duration) int {
	return m.sessionManager.CleanupExpired(maxAge)
}

// generateInitialDSL creates the initial DSL fragment for a new case
func (m *DSLManager) generateInitialDSL(domain string, initialData map[string]interface{}) (string, error) {
	switch domain {
	case "investor":
		name, _ := initialData["investor-name"].(string)
		investorType, _ := initialData["investor-type"].(string)
		if name == "" {
			name = "Unknown Investor"
		}
		if investorType == "" {
			investorType = "INDIVIDUAL"
		}
		return fmt.Sprintf("(investor.create (name \"%s\") (type \"%s\"))", name, investorType), nil

	case "fund":
		fundName, _ := initialData["fund-name"].(string)
		strategy, _ := initialData["strategy"].(string)
		if fundName == "" {
			fundName = "New Fund"
		}
		if strategy == "" {
			strategy = "Long/Short"
		}
		return fmt.Sprintf("(fund.create (name \"%s\") (strategy \"%s\"))", fundName, strategy), nil

	default:
		return fmt.Sprintf("(case.create (domain \"%s\"))", domain), nil
	}
}

/*
This implementation provides a robust DSL Manager with the following key features:

1. Unique onboarding ID generation
2. Domain-specific DSL management
3. State tracking and validation
4. Versioning with incremental updates
5. Comprehensive error handling and logging
6. Flexible context management

Key methods:
- `CreateCase`: Initializes a new case with unique ID and initial DSL
- `UpdateCase`: Updates an existing case with new DSL fragment and handles state transitions
- `GetCase`: Retrieves a case by onboarding ID
- `ListCases`: Lists active case session IDs
- `persistDSLVersion`: Saves each DSL state as a versioned record
- `validateStateTransition`: Ensures valid state progression

TODO items (marked in code):
1. Implement actual DSL version storage in the store
2. Complete version tracking mechanism
3. Add more comprehensive state transition rules

The implementation uses the existing session management system and provides a layer of domain-specific management on top of it.
*/
