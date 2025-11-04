package store

import (
	"context"
	"time"

	"dsl-ob-poc/internal/hf-investor/domain"
	"dsl-ob-poc/internal/hf-investor/state"

	"github.com/google/uuid"
)

// HedgeFundInvestorStore defines the interface for hedge fund investor data operations
type HedgeFundInvestorStore interface {
	// Lifecycle
	Close() error

	// ============================================================================
	// INVESTOR OPERATIONS
	// ============================================================================

	// Investor CRUD
	CreateInvestor(ctx context.Context, investor *domain.HedgeFundInvestor) error
	GetInvestor(ctx context.Context, investorID uuid.UUID) (*domain.HedgeFundInvestor, error)
	GetInvestorByCode(ctx context.Context, investorCode string) (*domain.HedgeFundInvestor, error)
	UpdateInvestor(ctx context.Context, investor *domain.HedgeFundInvestor) error
	DeleteInvestor(ctx context.Context, investorID uuid.UUID) error
	ListInvestors(ctx context.Context, filters InvestorFilters) ([]domain.HedgeFundInvestor, error)
	SearchInvestors(ctx context.Context, query string) ([]domain.HedgeFundInvestor, error)

	// Beneficial Owner Operations
	CreateBeneficialOwner(ctx context.Context, bo *domain.HedgeFundBeneficialOwner) error
	GetBeneficialOwners(ctx context.Context, investorID uuid.UUID) ([]domain.HedgeFundBeneficialOwner, error)
	UpdateBeneficialOwner(ctx context.Context, bo *domain.HedgeFundBeneficialOwner) error
	DeleteBeneficialOwner(ctx context.Context, boID uuid.UUID) error

	// ============================================================================
	// FUND STRUCTURE OPERATIONS
	// ============================================================================

	// Fund Operations
	CreateFund(ctx context.Context, fund *domain.HedgeFund) error
	GetFund(ctx context.Context, fundID uuid.UUID) (*domain.HedgeFund, error)
	GetFundByName(ctx context.Context, fundName string) (*domain.HedgeFund, error)
	UpdateFund(ctx context.Context, fund *domain.HedgeFund) error
	ListFunds(ctx context.Context) ([]domain.HedgeFund, error)

	// Share Class Operations
	CreateShareClass(ctx context.Context, class *domain.HedgeFundShareClass) error
	GetShareClass(ctx context.Context, classID uuid.UUID) (*domain.HedgeFundShareClass, error)
	GetShareClassesForFund(ctx context.Context, fundID uuid.UUID) ([]domain.HedgeFundShareClass, error)
	UpdateShareClass(ctx context.Context, class *domain.HedgeFundShareClass) error

	// Series Operations
	CreateSeries(ctx context.Context, series *domain.HedgeFundSeries) error
	GetSeries(ctx context.Context, seriesID uuid.UUID) (*domain.HedgeFundSeries, error)
	GetSeriesForClass(ctx context.Context, classID uuid.UUID) ([]domain.HedgeFundSeries, error)
	UpdateSeries(ctx context.Context, series *domain.HedgeFundSeries) error

	// ============================================================================
	// COMPLIANCE OPERATIONS
	// ============================================================================

	// KYC Profile Operations
	CreateKYCProfile(ctx context.Context, kyc *domain.HedgeFundKYCProfile) error
	GetKYCProfile(ctx context.Context, investorID uuid.UUID) (*domain.HedgeFundKYCProfile, error)
	UpdateKYCProfile(ctx context.Context, kyc *domain.HedgeFundKYCProfile) error
	GetKYCProfilesForRefresh(ctx context.Context, beforeDate time.Time) ([]domain.HedgeFundKYCProfile, error)

	// Tax Profile Operations
	CreateTaxProfile(ctx context.Context, tax *domain.HedgeFundTaxProfile) error
	GetTaxProfile(ctx context.Context, investorID uuid.UUID) (*domain.HedgeFundTaxProfile, error)
	UpdateTaxProfile(ctx context.Context, tax *domain.HedgeFundTaxProfile) error

	// Document Operations
	CreateDocument(ctx context.Context, doc *HedgeFundDocument) error
	GetDocuments(ctx context.Context, investorID uuid.UUID) ([]HedgeFundDocument, error)
	GetDocumentsByType(ctx context.Context, investorID uuid.UUID, docType string) ([]HedgeFundDocument, error)
	UpdateDocument(ctx context.Context, doc *HedgeFundDocument) error
	GetExpiredDocuments(ctx context.Context, beforeDate time.Time) ([]HedgeFundDocument, error)

	// ============================================================================
	// BANKING OPERATIONS
	// ============================================================================

	// Bank Instruction Operations
	CreateBankInstruction(ctx context.Context, bank *domain.HedgeFundBankInstruction) error
	GetBankInstructions(ctx context.Context, investorID uuid.UUID) ([]domain.HedgeFundBankInstruction, error)
	GetBankInstructionByCurrency(ctx context.Context, investorID uuid.UUID, currency string) (*domain.HedgeFundBankInstruction, error)
	UpdateBankInstruction(ctx context.Context, bank *domain.HedgeFundBankInstruction) error
	SupersedeBankInstruction(ctx context.Context, oldBankID, newBankID uuid.UUID) error

	// ============================================================================
	// TRADING OPERATIONS
	// ============================================================================

	// Trade Operations
	CreateTrade(ctx context.Context, trade *domain.HedgeFundTrade) error
	GetTrade(ctx context.Context, tradeID uuid.UUID) (*domain.HedgeFundTrade, error)
	GetTradeByReference(ctx context.Context, tradeRef string) (*domain.HedgeFundTrade, error)
	GetTradesForInvestor(ctx context.Context, investorID uuid.UUID, filters TradeFilters) ([]domain.HedgeFundTrade, error)
	UpdateTrade(ctx context.Context, trade *domain.HedgeFundTrade) error
	GetPendingTrades(ctx context.Context) ([]domain.HedgeFundTrade, error)

	// Register Lot Operations
	CreateRegisterLot(ctx context.Context, lot *domain.HedgeFundRegisterLot) error
	GetRegisterLot(ctx context.Context, lotID uuid.UUID) (*domain.HedgeFundRegisterLot, error)
	GetRegisterLotsForInvestor(ctx context.Context, investorID uuid.UUID) ([]domain.HedgeFundRegisterLot, error)
	GetRegisterLotByPosition(ctx context.Context, investorID, fundID, classID uuid.UUID, seriesID *uuid.UUID) (*domain.HedgeFundRegisterLot, error)
	UpdateRegisterLot(ctx context.Context, lot *domain.HedgeFundRegisterLot) error

	// Register Event Operations (Event Sourcing)
	CreateRegisterEvent(ctx context.Context, event *domain.HedgeFundRegisterEvent) error
	GetRegisterEvents(ctx context.Context, lotID uuid.UUID) ([]domain.HedgeFundRegisterEvent, error)
	GetRegisterEventsForInvestor(ctx context.Context, investorID uuid.UUID, filters EventFilters) ([]domain.HedgeFundRegisterEvent, error)
	GetRegisterEventsAfter(ctx context.Context, after time.Time) ([]domain.HedgeFundRegisterEvent, error)

	// ============================================================================
	// LIFECYCLE STATE OPERATIONS
	// ============================================================================

	// Lifecycle State Operations
	CreateLifecycleState(ctx context.Context, state *state.HedgeFundLifecycleState) error
	GetLifecycleStates(ctx context.Context, investorID uuid.UUID) ([]state.HedgeFundLifecycleState, error)
	GetCurrentLifecycleState(ctx context.Context, investorID uuid.UUID) (*state.HedgeFundLifecycleState, error)

	// ============================================================================
	// DSL EXECUTION OPERATIONS
	// ============================================================================

	// DSL Execution Operations
	CreateDSLExecution(ctx context.Context, execution *HedgeFundDSLExecution) error
	GetDSLExecution(ctx context.Context, executionID uuid.UUID) (*HedgeFundDSLExecution, error)
	GetDSLExecutionsForInvestor(ctx context.Context, investorID uuid.UUID) ([]HedgeFundDSLExecution, error)
	UpdateDSLExecution(ctx context.Context, execution *HedgeFundDSLExecution) error

	// ============================================================================
	// AUDIT OPERATIONS
	// ============================================================================

	// Audit Event Operations
	CreateAuditEvent(ctx context.Context, audit *HedgeFundAuditEvent) error
	GetAuditEvents(ctx context.Context, filters AuditFilters) ([]HedgeFundAuditEvent, error)

	// ============================================================================
	// REPORTING OPERATIONS
	// ============================================================================

	// Register Reporting
	GetRegisterOfInvestors(ctx context.Context, filters RegisterFilters) ([]RegisterOfInvestorsView, error)
	GetFundSummary(ctx context.Context, fundID *uuid.UUID) ([]FundSummaryView, error)
	GetKYCDashboard(ctx context.Context, filters KYCDashboardFilters) ([]KYCDashboardView, error)

	// Position Reporting
	GetInvestorPositions(ctx context.Context, investorID uuid.UUID, asOfDate *time.Time) ([]InvestorPosition, error)
	GetFundPositions(ctx context.Context, fundID uuid.UUID, asOfDate *time.Time) ([]FundPosition, error)

	// ============================================================================
	// UTILITY OPERATIONS
	// ============================================================================

	// Database Initialization
	InitHFDatabase(ctx context.Context) error
	SeedHFCatalog(ctx context.Context) error

	// Health Check
	HealthCheck(ctx context.Context) error
}

// ============================================================================
// FILTER TYPES
// ============================================================================

// InvestorFilters for filtering investor queries
type InvestorFilters struct {
	Status    *string
	Type      *string
	Domicile  *string
	Limit     *int
	Offset    *int
	SortBy    *string
	SortOrder *string // ASC or DESC
}

// TradeFilters for filtering trade queries
type TradeFilters struct {
	TradeType *string
	Status    *string
	Currency  *string
	FromDate  *time.Time
	ToDate    *time.Time
	Limit     *int
	Offset    *int
}

// EventFilters for filtering register event queries
type EventFilters struct {
	EventType *string
	FromDate  *time.Time
	ToDate    *time.Time
	Limit     *int
	Offset    *int
}

// RegisterFilters for register of investors reporting
type RegisterFilters struct {
	FundID   *uuid.UUID
	ClassID  *uuid.UUID
	SeriesID *uuid.UUID
	Status   *string
	MinUnits *float64
	AsOfDate *time.Time
}

// KYCDashboardFilters for KYC dashboard reporting
type KYCDashboardFilters struct {
	RiskRating  *string
	Status      *string
	RefreshDue  *bool
	OverdueOnly *bool
}

// AuditFilters for audit event queries
type AuditFilters struct {
	InvestorID *uuid.UUID
	EntityType *string
	Action     *string
	UserID     *string
	FromDate   *time.Time
	ToDate     *time.Time
	Limit      *int
	Offset     *int
}

// ============================================================================
// ADDITIONAL TYPES FOR HEDGE FUND OPERATIONS
// ============================================================================

// HedgeFundDocument represents a document in the hedge fund system
type HedgeFundDocument struct {
	DocumentID      uuid.UUID  `json:"document_id" db:"document_id"`
	InvestorID      uuid.UUID  `json:"investor_id" db:"investor_id"`
	DocumentType    string     `json:"document_type" db:"document_type"`
	DocumentSubject *string    `json:"document_subject,omitempty" db:"document_subject"`
	DocumentTitle   string     `json:"document_title" db:"document_title"`
	FileName        *string    `json:"file_name,omitempty" db:"file_name"`
	FileSize        *int64     `json:"file_size,omitempty" db:"file_size"`
	MimeType        *string    `json:"mime_type,omitempty" db:"mime_type"`
	FileHash        *string    `json:"file_hash,omitempty" db:"file_hash"`
	StorageProvider *string    `json:"storage_provider,omitempty" db:"storage_provider"`
	StoragePath     *string    `json:"storage_path,omitempty" db:"storage_path"`
	Status          string     `json:"status" db:"status"`
	ReviewedBy      *string    `json:"reviewed_by,omitempty" db:"reviewed_by"`
	ReviewedAt      *time.Time `json:"reviewed_at,omitempty" db:"reviewed_at"`
	ReviewComments  *string    `json:"review_comments,omitempty" db:"review_comments"`
	IssuedDate      *time.Time `json:"issued_date,omitempty" db:"issued_date"`
	ExpiryDate      *time.Time `json:"expiry_date,omitempty" db:"expiry_date"`
	SupersededBy    *uuid.UUID `json:"superseded_by,omitempty" db:"superseded_by"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// HedgeFundDSLExecution represents a DSL execution log entry
type HedgeFundDSLExecution struct {
	ExecutionID      uuid.UUID              `json:"execution_id" db:"execution_id"`
	InvestorID       uuid.UUID              `json:"investor_id" db:"investor_id"`
	DSLText          string                 `json:"dsl_text" db:"dsl_text"`
	ExecutionStatus  string                 `json:"execution_status" db:"execution_status"`
	IdempotencyKey   *string                `json:"idempotency_key,omitempty" db:"idempotency_key"`
	TriggeredBy      *string                `json:"triggered_by,omitempty" db:"triggered_by"`
	ExecutionEngine  string                 `json:"execution_engine" db:"execution_engine"`
	AffectedEntities map[string]interface{} `json:"affected_entities,omitempty" db:"affected_entities"`
	ErrorDetails     *string                `json:"error_details,omitempty" db:"error_details"`
	ExecutionTimeMs  *int                   `json:"execution_time_ms,omitempty" db:"execution_time_ms"`
	StartedAt        *time.Time             `json:"started_at,omitempty" db:"started_at"`
	CompletedAt      *time.Time             `json:"completed_at,omitempty" db:"completed_at"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
}

// HedgeFundAuditEvent represents an audit log entry
type HedgeFundAuditEvent struct {
	AuditID    uuid.UUID              `json:"audit_id" db:"audit_id"`
	InvestorID *uuid.UUID             `json:"investor_id,omitempty" db:"investor_id"`
	EntityType string                 `json:"entity_type" db:"entity_type"`
	EntityID   uuid.UUID              `json:"entity_id" db:"entity_id"`
	Action     string                 `json:"action" db:"action"`
	Details    map[string]interface{} `json:"details,omitempty" db:"details"`
	UserID     string                 `json:"user_id" db:"user_id"`
	UserIP     *string                `json:"user_ip,omitempty" db:"user_ip"`
	UserAgent  *string                `json:"user_agent,omitempty" db:"user_agent"`
	Timestamp  time.Time              `json:"timestamp" db:"timestamp"`
}

// ============================================================================
// VIEW TYPES FOR REPORTING
// ============================================================================

// RegisterOfInvestorsView represents the register of investors view
type RegisterOfInvestorsView struct {
	InvestorID          uuid.UUID  `json:"investor_id" db:"investor_id"`
	InvestorCode        string     `json:"investor_code" db:"investor_code"`
	InvestorName        string     `json:"investor_name" db:"investor_name"`
	InvestorType        string     `json:"investor_type" db:"investor_type"`
	Domicile            string     `json:"domicile" db:"domicile"`
	InvestorStatus      string     `json:"investor_status" db:"investor_status"`
	FundName            string     `json:"fund_name" db:"fund_name"`
	FundID              uuid.UUID  `json:"fund_id" db:"fund_id"`
	ClassName           string     `json:"class_name" db:"class_name"`
	ClassCurrency       string     `json:"class_currency" db:"class_currency"`
	SeriesName          *string    `json:"series_name,omitempty" db:"series_name"`
	CurrentUnits        float64    `json:"current_units" db:"current_units"`
	TotalInvestment     float64    `json:"total_investment" db:"total_investment"`
	AverageCost         *float64   `json:"average_cost,omitempty" db:"average_cost"`
	FirstTradeDate      *time.Time `json:"first_trade_date,omitempty" db:"first_trade_date"`
	LastActivityAt      *time.Time `json:"last_activity_at,omitempty" db:"last_activity_at"`
	RiskRating          *string    `json:"risk_rating,omitempty" db:"risk_rating"`
	KYCStatus           *string    `json:"kyc_status,omitempty" db:"kyc_status"`
	RefreshDueAt        *time.Time `json:"refresh_due_at,omitempty" db:"refresh_due_at"`
	FATCAStatus         *string    `json:"fatca_status,omitempty" db:"fatca_status"`
	CRSClassification   *string    `json:"crs_classification,omitempty" db:"crs_classification"`
	WithholdingRate     float64    `json:"withholding_rate" db:"withholding_rate"`
	PrimaryContactName  *string    `json:"primary_contact_name,omitempty" db:"primary_contact_name"`
	PrimaryContactEmail *string    `json:"primary_contact_email,omitempty" db:"primary_contact_email"`
	InvestorCreatedAt   time.Time  `json:"investor_created_at" db:"investor_created_at"`
	PositionUpdatedAt   *time.Time `json:"position_updated_at,omitempty" db:"position_updated_at"`
}

// FundSummaryView represents fund summary statistics
type FundSummaryView struct {
	FundID                uuid.UUID  `json:"fund_id" db:"fund_id"`
	FundName              string     `json:"fund_name" db:"fund_name"`
	FundStatus            string     `json:"fund_status" db:"fund_status"`
	ClassID               uuid.UUID  `json:"class_id" db:"class_id"`
	ClassName             string     `json:"class_name" db:"class_name"`
	Currency              string     `json:"currency" db:"currency"`
	TotalInvestors        int        `json:"total_investors" db:"total_investors"`
	ActiveInvestors       int        `json:"active_investors" db:"active_investors"`
	TotalUnitsOutstanding float64    `json:"total_units_outstanding" db:"total_units_outstanding"`
	TotalAUM              float64    `json:"total_aum" db:"total_aum"`
	LastActivityDate      *time.Time `json:"last_activity_date,omitempty" db:"last_activity_date"`
	InceptionDate         time.Time  `json:"inception_date" db:"inception_date"`
	FundCreatedAt         time.Time  `json:"fund_created_at" db:"fund_created_at"`
}

// KYCDashboardView represents KYC dashboard information
type KYCDashboardView struct {
	InvestorID        uuid.UUID  `json:"investor_id" db:"investor_id"`
	InvestorCode      string     `json:"investor_code" db:"investor_code"`
	LegalName         string     `json:"legal_name" db:"legal_name"`
	InvestorStatus    string     `json:"investor_status" db:"investor_status"`
	RiskRating        *string    `json:"risk_rating,omitempty" db:"risk_rating"`
	KYCStatus         *string    `json:"kyc_status,omitempty" db:"kyc_status"`
	RefreshDueAt      *time.Time `json:"refresh_due_at,omitempty" db:"refresh_due_at"`
	LastRefreshedAt   *time.Time `json:"last_refreshed_at,omitempty" db:"last_refreshed_at"`
	TotalDocuments    int        `json:"total_documents" db:"total_documents"`
	ApprovedDocuments int        `json:"approved_documents" db:"approved_documents"`
	PendingDocuments  int        `json:"pending_documents" db:"pending_documents"`
	ExpiredDocuments  int        `json:"expired_documents" db:"expired_documents"`
	HasPEP            bool       `json:"has_pep" db:"has_pep"`
	HasSanctionsFlag  bool       `json:"has_sanctions_flag" db:"has_sanctions_flag"`
	NextAction        string     `json:"next_action" db:"next_action"`
}

// InvestorPosition represents an investor's position summary
type InvestorPosition struct {
	InvestorID     uuid.UUID  `json:"investor_id"`
	InvestorCode   string     `json:"investor_code"`
	FundID         uuid.UUID  `json:"fund_id"`
	FundName       string     `json:"fund_name"`
	ClassID        uuid.UUID  `json:"class_id"`
	ClassName      string     `json:"class_name"`
	SeriesID       *uuid.UUID `json:"series_id,omitempty"`
	SeriesName     *string    `json:"series_name,omitempty"`
	Units          float64    `json:"units"`
	AverageCost    float64    `json:"average_cost"`
	TotalCost      float64    `json:"total_cost"`
	FirstTradeDate time.Time  `json:"first_trade_date"`
	LastActivityAt time.Time  `json:"last_activity_at"`
}

// FundPosition represents a fund's position summary
type FundPosition struct {
	FundID           uuid.UUID `json:"fund_id"`
	FundName         string    `json:"fund_name"`
	ClassID          uuid.UUID `json:"class_id"`
	ClassName        string    `json:"class_name"`
	Currency         string    `json:"currency"`
	TotalUnits       float64   `json:"total_units"`
	TotalInvestors   int       `json:"total_investors"`
	AveragePosition  float64   `json:"average_position"`
	LargestPosition  float64   `json:"largest_position"`
	SmallestPosition float64   `json:"smallest_position"`
}

// ============================================================================
// CONSTANTS
// ============================================================================

// Document Status constants
const (
	DocumentStatusReceived    = "RECEIVED"
	DocumentStatusUnderReview = "UNDER_REVIEW"
	DocumentStatusApproved    = "APPROVED"
	DocumentStatusRejected    = "REJECTED"
	DocumentStatusExpired     = "EXPIRED"
	DocumentStatusSuperseded  = "SUPERSEDED"
)

// DSL Execution Status constants
const (
	DSLExecutionStatusPending   = "PENDING"
	DSLExecutionStatusRunning   = "RUNNING"
	DSLExecutionStatusCompleted = "COMPLETED"
	DSLExecutionStatusFailed    = "FAILED"
	DSLExecutionStatusCanceled  = "CANCELED"
)

// Audit Action constants
const (
	AuditActionCreate     = "CREATE"
	AuditActionUpdate     = "UPDATE"
	AuditActionDelete     = "DELETE"
	AuditActionApprove    = "APPROVE"
	AuditActionReject     = "REJECT"
	AuditActionTransition = "TRANSITION"
)
