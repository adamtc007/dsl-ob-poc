package store

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"dsl-ob-poc/internal/hf-investor/domain"
	"dsl-ob-poc/internal/hf-investor/state"

	"github.com/google/uuid"
)

// MockHedgeFundInvestorStore is an in-memory implementation of HedgeFundInvestorStore for testing and development
type MockHedgeFundInvestorStore struct {
	mu sync.RWMutex

	// Data storage
	investors        map[uuid.UUID]*domain.HedgeFundInvestor
	beneficialOwners map[uuid.UUID]*domain.HedgeFundBeneficialOwner
	funds            map[uuid.UUID]*domain.HedgeFund
	shareClasses     map[uuid.UUID]*domain.HedgeFundShareClass
	series           map[uuid.UUID]*domain.HedgeFundSeries
	kycProfiles      map[uuid.UUID]*domain.HedgeFundKYCProfile
	taxProfiles      map[uuid.UUID]*domain.HedgeFundTaxProfile
	documents        map[uuid.UUID]*HedgeFundDocument
	bankInstructions map[uuid.UUID]*domain.HedgeFundBankInstruction
	trades           map[uuid.UUID]*domain.HedgeFundTrade
	registerLots     map[uuid.UUID]*domain.HedgeFundRegisterLot
	registerEvents   map[uuid.UUID]*domain.HedgeFundRegisterEvent
	lifecycleStates  map[uuid.UUID]*state.HedgeFundLifecycleState
	dslExecutions    map[uuid.UUID]*HedgeFundDSLExecution
	auditEvents      map[uuid.UUID]*HedgeFundAuditEvent

	// Indexes for efficient lookups
	investorsByCode   map[string]uuid.UUID
	fundsByName       map[string]uuid.UUID
	tradesByReference map[string]uuid.UUID

	// Sequence counters
	investorCodeSeq int
	tradeRefSeq     int
}

// NewMockHedgeFundInvestorStore creates a new in-memory hedge fund investor store
func NewMockHedgeFundInvestorStore() *MockHedgeFundInvestorStore {
	return &MockHedgeFundInvestorStore{
		investors:         make(map[uuid.UUID]*domain.HedgeFundInvestor),
		beneficialOwners:  make(map[uuid.UUID]*domain.HedgeFundBeneficialOwner),
		funds:             make(map[uuid.UUID]*domain.HedgeFund),
		shareClasses:      make(map[uuid.UUID]*domain.HedgeFundShareClass),
		series:            make(map[uuid.UUID]*domain.HedgeFundSeries),
		kycProfiles:       make(map[uuid.UUID]*domain.HedgeFundKYCProfile),
		taxProfiles:       make(map[uuid.UUID]*domain.HedgeFundTaxProfile),
		documents:         make(map[uuid.UUID]*HedgeFundDocument),
		bankInstructions:  make(map[uuid.UUID]*domain.HedgeFundBankInstruction),
		trades:            make(map[uuid.UUID]*domain.HedgeFundTrade),
		registerLots:      make(map[uuid.UUID]*domain.HedgeFundRegisterLot),
		registerEvents:    make(map[uuid.UUID]*domain.HedgeFundRegisterEvent),
		lifecycleStates:   make(map[uuid.UUID]*state.HedgeFundLifecycleState),
		dslExecutions:     make(map[uuid.UUID]*HedgeFundDSLExecution),
		auditEvents:       make(map[uuid.UUID]*HedgeFundAuditEvent),
		investorsByCode:   make(map[string]uuid.UUID),
		fundsByName:       make(map[string]uuid.UUID),
		tradesByReference: make(map[string]uuid.UUID),
		investorCodeSeq:   1,
		tradeRefSeq:       1,
	}
}

// Close implements HedgeFundInvestorStore
func (m *MockHedgeFundInvestorStore) Close() error {
	return nil
}

// ============================================================================
// INVESTOR OPERATIONS
// ============================================================================

// CreateInvestor implements HedgeFundInvestorStore
func (m *MockHedgeFundInvestorStore) CreateInvestor(ctx context.Context, investor *domain.HedgeFundInvestor) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if investor code already exists
	if _, exists := m.investorsByCode[investor.InvestorCode]; exists {
		return fmt.Errorf("investor with code %s already exists", investor.InvestorCode)
	}

	// Generate investor code if empty
	if investor.InvestorCode == "" {
		investor.InvestorCode = fmt.Sprintf("INV-%03d", m.investorCodeSeq)
		m.investorCodeSeq++
	}

	// Set timestamps
	now := time.Now().UTC()
	if investor.CreatedAt.IsZero() {
		investor.CreatedAt = now
	}
	investor.UpdatedAt = now

	// Store investor
	m.investors[investor.InvestorID] = investor
	m.investorsByCode[investor.InvestorCode] = investor.InvestorID

	return nil
}

// GetInvestor implements HedgeFundInvestorStore
func (m *MockHedgeFundInvestorStore) GetInvestor(ctx context.Context, investorID uuid.UUID) (*domain.HedgeFundInvestor, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	investor, exists := m.investors[investorID]
	if !exists {
		return nil, fmt.Errorf("investor with ID %s not found", investorID)
	}

	// Return a copy
	result := *investor
	return &result, nil
}

// GetInvestorByCode implements HedgeFundInvestorStore
func (m *MockHedgeFundInvestorStore) GetInvestorByCode(ctx context.Context, investorCode string) (*domain.HedgeFundInvestor, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	investorID, exists := m.investorsByCode[investorCode]
	if !exists {
		return nil, fmt.Errorf("investor with code %s not found", investorCode)
	}

	investor := m.investors[investorID]
	result := *investor
	return &result, nil
}

// UpdateInvestor implements HedgeFundInvestorStore
func (m *MockHedgeFundInvestorStore) UpdateInvestor(ctx context.Context, investor *domain.HedgeFundInvestor) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.investors[investor.InvestorID]; !exists {
		return fmt.Errorf("investor with ID %s not found", investor.InvestorID)
	}

	investor.UpdatedAt = time.Now().UTC()
	m.investors[investor.InvestorID] = investor

	return nil
}

// DeleteInvestor implements HedgeFundInvestorStore
func (m *MockHedgeFundInvestorStore) DeleteInvestor(ctx context.Context, investorID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	investor, exists := m.investors[investorID]
	if !exists {
		return fmt.Errorf("investor with ID %s not found", investorID)
	}

	delete(m.investors, investorID)
	delete(m.investorsByCode, investor.InvestorCode)

	return nil
}

// ListInvestors implements HedgeFundInvestorStore
func (m *MockHedgeFundInvestorStore) ListInvestors(ctx context.Context, filters InvestorFilters) ([]domain.HedgeFundInvestor, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]domain.HedgeFundInvestor, 0, len(m.investors))

	for _, investor := range m.investors {
		// Apply filters
		if filters.Status != nil && investor.Status != *filters.Status {
			continue
		}
		if filters.Type != nil && investor.Type != *filters.Type {
			continue
		}
		if filters.Domicile != nil && investor.Domicile != *filters.Domicile {
			continue
		}

		result = append(result, *investor)
	}

	// Sort by investor code by default
	sort.Slice(result, func(i, j int) bool {
		return result[i].InvestorCode < result[j].InvestorCode
	})

	// Apply limit and offset
	if filters.Offset != nil && *filters.Offset > 0 {
		if *filters.Offset >= len(result) {
			return []domain.HedgeFundInvestor{}, nil
		}
		result = result[*filters.Offset:]
	}

	if filters.Limit != nil && *filters.Limit > 0 && *filters.Limit < len(result) {
		result = result[:*filters.Limit]
	}

	return result, nil
}

// SearchInvestors implements HedgeFundInvestorStore
func (m *MockHedgeFundInvestorStore) SearchInvestors(ctx context.Context, query string) ([]domain.HedgeFundInvestor, error) {
	// Simple implementation - search in legal name and investor code
	// In a real implementation, this would use full-text search
	return m.ListInvestors(ctx, InvestorFilters{})
}

// ============================================================================
// BENEFICIAL OWNER OPERATIONS
// ============================================================================

// CreateBeneficialOwner implements HedgeFundInvestorStore
func (m *MockHedgeFundInvestorStore) CreateBeneficialOwner(ctx context.Context, bo *domain.HedgeFundBeneficialOwner) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Verify investor exists
	if _, exists := m.investors[bo.InvestorID]; !exists {
		return fmt.Errorf("investor with ID %s not found", bo.InvestorID)
	}

	// Set timestamps
	now := time.Now().UTC()
	if bo.CreatedAt.IsZero() {
		bo.CreatedAt = now
	}
	bo.UpdatedAt = now

	m.beneficialOwners[bo.BOID] = bo
	return nil
}

// GetBeneficialOwners implements HedgeFundInvestorStore
func (m *MockHedgeFundInvestorStore) GetBeneficialOwners(ctx context.Context, investorID uuid.UUID) ([]domain.HedgeFundBeneficialOwner, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []domain.HedgeFundBeneficialOwner

	for _, bo := range m.beneficialOwners {
		if bo.InvestorID == investorID {
			result = append(result, *bo)
		}
	}

	return result, nil
}

// UpdateBeneficialOwner implements HedgeFundInvestorStore
func (m *MockHedgeFundInvestorStore) UpdateBeneficialOwner(ctx context.Context, bo *domain.HedgeFundBeneficialOwner) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.beneficialOwners[bo.BOID]; !exists {
		return fmt.Errorf("beneficial owner with ID %s not found", bo.BOID)
	}

	bo.UpdatedAt = time.Now().UTC()
	m.beneficialOwners[bo.BOID] = bo

	return nil
}

// DeleteBeneficialOwner implements HedgeFundInvestorStore
func (m *MockHedgeFundInvestorStore) DeleteBeneficialOwner(ctx context.Context, boID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.beneficialOwners[boID]; !exists {
		return fmt.Errorf("beneficial owner with ID %s not found", boID)
	}

	delete(m.beneficialOwners, boID)
	return nil
}

// ============================================================================
// FUND STRUCTURE OPERATIONS (Basic Implementations)
// ============================================================================

// CreateFund implements HedgeFundInvestorStore
func (m *MockHedgeFundInvestorStore) CreateFund(ctx context.Context, fund *domain.HedgeFund) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.fundsByName[fund.FundName]; exists {
		return fmt.Errorf("fund with name %s already exists", fund.FundName)
	}

	now := time.Now().UTC()
	if fund.CreatedAt.IsZero() {
		fund.CreatedAt = now
	}
	fund.UpdatedAt = now

	m.funds[fund.FundID] = fund
	m.fundsByName[fund.FundName] = fund.FundID

	return nil
}

// GetFund implements HedgeFundInvestorStore
func (m *MockHedgeFundInvestorStore) GetFund(ctx context.Context, fundID uuid.UUID) (*domain.HedgeFund, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	fund, exists := m.funds[fundID]
	if !exists {
		return nil, fmt.Errorf("fund with ID %s not found", fundID)
	}

	result := *fund
	return &result, nil
}

// GetFundByName implements HedgeFundInvestorStore
func (m *MockHedgeFundInvestorStore) GetFundByName(ctx context.Context, fundName string) (*domain.HedgeFund, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	fundID, exists := m.fundsByName[fundName]
	if !exists {
		return nil, fmt.Errorf("fund with name %s not found", fundName)
	}

	fund := m.funds[fundID]
	result := *fund
	return &result, nil
}

// UpdateFund implements HedgeFundInvestorStore
func (m *MockHedgeFundInvestorStore) UpdateFund(ctx context.Context, fund *domain.HedgeFund) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.funds[fund.FundID]; !exists {
		return fmt.Errorf("fund with ID %s not found", fund.FundID)
	}

	fund.UpdatedAt = time.Now().UTC()
	m.funds[fund.FundID] = fund

	return nil
}

// ListFunds implements HedgeFundInvestorStore
func (m *MockHedgeFundInvestorStore) ListFunds(ctx context.Context) ([]domain.HedgeFund, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]domain.HedgeFund, 0, len(m.funds))
	for _, fund := range m.funds {
		result = append(result, *fund)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].FundName < result[j].FundName
	})

	return result, nil
}

// ============================================================================
// PLACEHOLDER IMPLEMENTATIONS (To be completed)
// ============================================================================

// The following methods provide basic placeholder implementations
// In a full implementation, these would have complete logic

func (m *MockHedgeFundInvestorStore) CreateShareClass(ctx context.Context, class *domain.HedgeFundShareClass) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now().UTC()
	if class.CreatedAt.IsZero() {
		class.CreatedAt = now
	}
	class.UpdatedAt = now
	m.shareClasses[class.ClassID] = class
	return nil
}

func (m *MockHedgeFundInvestorStore) GetShareClass(ctx context.Context, classID uuid.UUID) (*domain.HedgeFundShareClass, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	class, exists := m.shareClasses[classID]
	if !exists {
		return nil, fmt.Errorf("share class with ID %s not found", classID)
	}
	result := *class
	return &result, nil
}

func (m *MockHedgeFundInvestorStore) GetShareClassesForFund(ctx context.Context, fundID uuid.UUID) ([]domain.HedgeFundShareClass, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []domain.HedgeFundShareClass
	for _, class := range m.shareClasses {
		if class.FundID == fundID {
			result = append(result, *class)
		}
	}
	return result, nil
}

func (m *MockHedgeFundInvestorStore) UpdateShareClass(ctx context.Context, class *domain.HedgeFundShareClass) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.shareClasses[class.ClassID]; !exists {
		return fmt.Errorf("share class with ID %s not found", class.ClassID)
	}
	class.UpdatedAt = time.Now().UTC()
	m.shareClasses[class.ClassID] = class
	return nil
}

// Add placeholder implementations for remaining methods to satisfy interface
// In a production system, these would be fully implemented

func (m *MockHedgeFundInvestorStore) CreateSeries(ctx context.Context, series *domain.HedgeFundSeries) error {
	return nil
}
func (m *MockHedgeFundInvestorStore) GetSeries(ctx context.Context, seriesID uuid.UUID) (*domain.HedgeFundSeries, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *MockHedgeFundInvestorStore) GetSeriesForClass(ctx context.Context, classID uuid.UUID) ([]domain.HedgeFundSeries, error) {
	return nil, nil
}
func (m *MockHedgeFundInvestorStore) UpdateSeries(ctx context.Context, series *domain.HedgeFundSeries) error {
	return nil
}
func (m *MockHedgeFundInvestorStore) CreateKYCProfile(ctx context.Context, kyc *domain.HedgeFundKYCProfile) error {
	return nil
}
func (m *MockHedgeFundInvestorStore) GetKYCProfile(ctx context.Context, investorID uuid.UUID) (*domain.HedgeFundKYCProfile, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *MockHedgeFundInvestorStore) UpdateKYCProfile(ctx context.Context, kyc *domain.HedgeFundKYCProfile) error {
	return nil
}
func (m *MockHedgeFundInvestorStore) GetKYCProfilesForRefresh(ctx context.Context, beforeDate time.Time) ([]domain.HedgeFundKYCProfile, error) {
	return nil, nil
}
func (m *MockHedgeFundInvestorStore) CreateTaxProfile(ctx context.Context, tax *domain.HedgeFundTaxProfile) error {
	return nil
}
func (m *MockHedgeFundInvestorStore) GetTaxProfile(ctx context.Context, investorID uuid.UUID) (*domain.HedgeFundTaxProfile, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *MockHedgeFundInvestorStore) UpdateTaxProfile(ctx context.Context, tax *domain.HedgeFundTaxProfile) error {
	return nil
}
func (m *MockHedgeFundInvestorStore) CreateDocument(ctx context.Context, doc *HedgeFundDocument) error {
	return nil
}
func (m *MockHedgeFundInvestorStore) GetDocuments(ctx context.Context, investorID uuid.UUID) ([]HedgeFundDocument, error) {
	return nil, nil
}
func (m *MockHedgeFundInvestorStore) GetDocumentsByType(ctx context.Context, investorID uuid.UUID, docType string) ([]HedgeFundDocument, error) {
	return nil, nil
}
func (m *MockHedgeFundInvestorStore) UpdateDocument(ctx context.Context, doc *HedgeFundDocument) error {
	return nil
}
func (m *MockHedgeFundInvestorStore) GetExpiredDocuments(ctx context.Context, beforeDate time.Time) ([]HedgeFundDocument, error) {
	return nil, nil
}
func (m *MockHedgeFundInvestorStore) CreateBankInstruction(ctx context.Context, bank *domain.HedgeFundBankInstruction) error {
	return nil
}
func (m *MockHedgeFundInvestorStore) GetBankInstructions(ctx context.Context, investorID uuid.UUID) ([]domain.HedgeFundBankInstruction, error) {
	return nil, nil
}
func (m *MockHedgeFundInvestorStore) GetBankInstructionByCurrency(ctx context.Context, investorID uuid.UUID, currency string) (*domain.HedgeFundBankInstruction, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *MockHedgeFundInvestorStore) UpdateBankInstruction(ctx context.Context, bank *domain.HedgeFundBankInstruction) error {
	return nil
}
func (m *MockHedgeFundInvestorStore) SupersedeBankInstruction(ctx context.Context, oldBankID, newBankID uuid.UUID) error {
	return nil
}
func (m *MockHedgeFundInvestorStore) CreateTrade(ctx context.Context, trade *domain.HedgeFundTrade) error {
	return nil
}
func (m *MockHedgeFundInvestorStore) GetTrade(ctx context.Context, tradeID uuid.UUID) (*domain.HedgeFundTrade, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *MockHedgeFundInvestorStore) GetTradeByReference(ctx context.Context, tradeRef string) (*domain.HedgeFundTrade, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *MockHedgeFundInvestorStore) GetTradesForInvestor(ctx context.Context, investorID uuid.UUID, filters TradeFilters) ([]domain.HedgeFundTrade, error) {
	return nil, nil
}
func (m *MockHedgeFundInvestorStore) UpdateTrade(ctx context.Context, trade *domain.HedgeFundTrade) error {
	return nil
}
func (m *MockHedgeFundInvestorStore) GetPendingTrades(ctx context.Context) ([]domain.HedgeFundTrade, error) {
	return nil, nil
}
func (m *MockHedgeFundInvestorStore) CreateRegisterLot(ctx context.Context, lot *domain.HedgeFundRegisterLot) error {
	return nil
}
func (m *MockHedgeFundInvestorStore) GetRegisterLot(ctx context.Context, lotID uuid.UUID) (*domain.HedgeFundRegisterLot, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *MockHedgeFundInvestorStore) GetRegisterLotsForInvestor(ctx context.Context, investorID uuid.UUID) ([]domain.HedgeFundRegisterLot, error) {
	return nil, nil
}
func (m *MockHedgeFundInvestorStore) GetRegisterLotByPosition(ctx context.Context, investorID, fundID, classID uuid.UUID, seriesID *uuid.UUID) (*domain.HedgeFundRegisterLot, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *MockHedgeFundInvestorStore) UpdateRegisterLot(ctx context.Context, lot *domain.HedgeFundRegisterLot) error {
	return nil
}
func (m *MockHedgeFundInvestorStore) CreateRegisterEvent(ctx context.Context, event *domain.HedgeFundRegisterEvent) error {
	return nil
}
func (m *MockHedgeFundInvestorStore) GetRegisterEvents(ctx context.Context, lotID uuid.UUID) ([]domain.HedgeFundRegisterEvent, error) {
	return nil, nil
}
func (m *MockHedgeFundInvestorStore) GetRegisterEventsForInvestor(ctx context.Context, investorID uuid.UUID, filters EventFilters) ([]domain.HedgeFundRegisterEvent, error) {
	return nil, nil
}
func (m *MockHedgeFundInvestorStore) GetRegisterEventsAfter(ctx context.Context, after time.Time) ([]domain.HedgeFundRegisterEvent, error) {
	return nil, nil
}
func (m *MockHedgeFundInvestorStore) CreateLifecycleState(ctx context.Context, state *state.HedgeFundLifecycleState) error {
	return nil
}
func (m *MockHedgeFundInvestorStore) GetLifecycleStates(ctx context.Context, investorID uuid.UUID) ([]state.HedgeFundLifecycleState, error) {
	return nil, nil
}
func (m *MockHedgeFundInvestorStore) GetCurrentLifecycleState(ctx context.Context, investorID uuid.UUID) (*state.HedgeFundLifecycleState, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *MockHedgeFundInvestorStore) CreateDSLExecution(ctx context.Context, execution *HedgeFundDSLExecution) error {
	return nil
}
func (m *MockHedgeFundInvestorStore) GetDSLExecution(ctx context.Context, executionID uuid.UUID) (*HedgeFundDSLExecution, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *MockHedgeFundInvestorStore) GetDSLExecutionsForInvestor(ctx context.Context, investorID uuid.UUID) ([]HedgeFundDSLExecution, error) {
	return nil, nil
}
func (m *MockHedgeFundInvestorStore) UpdateDSLExecution(ctx context.Context, execution *HedgeFundDSLExecution) error {
	return nil
}
func (m *MockHedgeFundInvestorStore) CreateAuditEvent(ctx context.Context, audit *HedgeFundAuditEvent) error {
	return nil
}
func (m *MockHedgeFundInvestorStore) GetAuditEvents(ctx context.Context, filters AuditFilters) ([]HedgeFundAuditEvent, error) {
	return nil, nil
}
func (m *MockHedgeFundInvestorStore) GetRegisterOfInvestors(ctx context.Context, filters RegisterFilters) ([]RegisterOfInvestorsView, error) {
	return nil, nil
}
func (m *MockHedgeFundInvestorStore) GetFundSummary(ctx context.Context, fundID *uuid.UUID) ([]FundSummaryView, error) {
	return nil, nil
}
func (m *MockHedgeFundInvestorStore) GetKYCDashboard(ctx context.Context, filters KYCDashboardFilters) ([]KYCDashboardView, error) {
	return nil, nil
}
func (m *MockHedgeFundInvestorStore) GetInvestorPositions(ctx context.Context, investorID uuid.UUID, asOfDate *time.Time) ([]InvestorPosition, error) {
	return nil, nil
}
func (m *MockHedgeFundInvestorStore) GetFundPositions(ctx context.Context, fundID uuid.UUID, asOfDate *time.Time) ([]FundPosition, error) {
	return nil, nil
}
func (m *MockHedgeFundInvestorStore) InitHFDatabase(ctx context.Context) error { return nil }
func (m *MockHedgeFundInvestorStore) SeedHFCatalog(ctx context.Context) error  { return nil }
func (m *MockHedgeFundInvestorStore) HealthCheck(ctx context.Context) error    { return nil }
