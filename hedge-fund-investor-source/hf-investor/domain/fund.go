package domain

import (
	"time"

	"github.com/google/uuid"
)

// HedgeFund represents a hedge fund entity
type HedgeFund struct {
	FundID        uuid.UUID `json:"fund_id" db:"fund_id"`
	FundName      string    `json:"fund_name" db:"fund_name"`
	LegalName     string    `json:"legal_name" db:"legal_name"`
	LEI           *string   `json:"lei,omitempty" db:"lei"`
	Domicile      string    `json:"domicile" db:"domicile"`
	FundType      string    `json:"fund_type" db:"fund_type"`
	Currency      string    `json:"currency" db:"currency"`
	InceptionDate time.Time `json:"inception_date" db:"inception_date"`
	Status        string    `json:"status" db:"status"`

	// Service providers
	Administrator *string `json:"administrator,omitempty" db:"administrator"`
	Custodian     *string `json:"custodian,omitempty" db:"custodian"`
	Auditor       *string `json:"auditor,omitempty" db:"auditor"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Fund Type constants
const (
	FundTypeHedgeFund      = "HEDGE_FUND"
	FundTypePrivateEquity  = "PRIVATE_EQUITY"
	FundTypeCreditFund     = "CREDIT_FUND"
	FundTypeInfrastructure = "INFRASTRUCTURE"
)

// Fund Status constants
const (
	FundStatusSetup       = "SETUP"
	FundStatusActive      = "ACTIVE"
	FundStatusSoftClosed  = "SOFT_CLOSED"
	FundStatusHardClosed  = "HARD_CLOSED"
	FundStatusLiquidating = "LIQUIDATING"
	FundStatusLiquidated  = "LIQUIDATED"
)

// HedgeFundShareClass represents a share class within a hedge fund
type HedgeFundShareClass struct {
	ClassID                 uuid.UUID `json:"class_id" db:"class_id"`
	FundID                  uuid.UUID `json:"fund_id" db:"fund_id"`
	ClassName               string    `json:"class_name" db:"class_name"`
	ClassType               string    `json:"class_type" db:"class_type"`
	Currency                string    `json:"currency" db:"currency"`
	MinInitialInvestment    float64   `json:"min_initial_investment" db:"min_initial_investment"`
	MinSubsequentInvestment float64   `json:"min_subsequent_investment" db:"min_subsequent_investment"`
	ManagementFeeRate       float64   `json:"management_fee_rate" db:"management_fee_rate"`
	PerformanceFeeRate      float64   `json:"performance_fee_rate" db:"performance_fee_rate"`
	HighWaterMark           bool      `json:"high_water_mark" db:"high_water_mark"`
	DealingFrequency        string    `json:"dealing_frequency" db:"dealing_frequency"`
	NoticePeriodDays        int       `json:"notice_period_days" db:"notice_period_days"`
	GatePercentage          *float64  `json:"gate_percentage,omitempty" db:"gate_percentage"`
	LockupMonths            int       `json:"lockup_months" db:"lockup_months"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Class Type constants
const (
	ClassTypeRetail        = "RETAIL"
	ClassTypeInstitutional = "INSTITUTIONAL"
	ClassTypeFounder       = "FOUNDER"
	ClassTypeEmployee      = "EMPLOYEE"
	ClassTypeSeeded        = "SEEDED"
)

// Dealing Frequency constants
const (
	DealingFrequencyDaily      = "DAILY"
	DealingFrequencyWeekly     = "WEEKLY"
	DealingFrequencyMonthly    = "MONTHLY"
	DealingFrequencyQuarterly  = "QUARTERLY"
	DealingFrequencySemiAnnual = "SEMI_ANNUAL"
	DealingFrequencyAnnual     = "ANNUAL"
)

// HedgeFundSeries represents a series within a share class for equalization
type HedgeFundSeries struct {
	SeriesID      uuid.UUID `json:"series_id" db:"series_id"`
	ClassID       uuid.UUID `json:"class_id" db:"class_id"`
	SeriesName    string    `json:"series_name" db:"series_name"`
	InceptionDate time.Time `json:"inception_date" db:"inception_date"`
	Status        string    `json:"status" db:"status"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Series Status constants
const (
	SeriesStatusActive = "ACTIVE"
	SeriesStatusClosed = "CLOSED"
	SeriesStatusMerged = "MERGED"
)

// HedgeFundBankInstruction represents settlement instructions
type HedgeFundBankInstruction struct {
	BankID          uuid.UUID `json:"bank_id" db:"bank_id"`
	InvestorID      uuid.UUID `json:"investor_id" db:"investor_id"`
	Currency        string    `json:"currency" db:"currency"`
	InstructionType string    `json:"instruction_type" db:"instruction_type"`

	// Bank details
	BankName      string  `json:"bank_name" db:"bank_name"`
	SwiftBIC      *string `json:"swift_bic,omitempty" db:"swift_bic"`
	IBAN          *string `json:"iban,omitempty" db:"iban"`
	AccountNumber *string `json:"account_number,omitempty" db:"account_number"`
	AccountName   string  `json:"account_name" db:"account_name"`

	// Bank address
	BankAddressLine1 *string `json:"bank_address_line1,omitempty" db:"bank_address_line1"`
	BankAddressLine2 *string `json:"bank_address_line2,omitempty" db:"bank_address_line2"`
	BankCity         *string `json:"bank_city,omitempty" db:"bank_city"`
	BankCountry      *string `json:"bank_country,omitempty" db:"bank_country"`

	// Intermediary bank
	IntermediarySwift   *string `json:"intermediary_swift,omitempty" db:"intermediary_swift"`
	IntermediaryName    *string `json:"intermediary_name,omitempty" db:"intermediary_name"`
	IntermediaryAccount *string `json:"intermediary_account,omitempty" db:"intermediary_account"`

	// Status and versioning
	Status        string     `json:"status" db:"status"`
	VersionNumber int        `json:"version_number" db:"version_number"`
	SupersededBy  *uuid.UUID `json:"superseded_by,omitempty" db:"superseded_by"`

	// Verification
	VerifiedBy         *string    `json:"verified_by,omitempty" db:"verified_by"`
	VerifiedAt         *time.Time `json:"verified_at,omitempty" db:"verified_at"`
	VerificationMethod *string    `json:"verification_method,omitempty" db:"verification_method"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Bank Instruction Type constants
const (
	InstructionTypeSettlement   = "SETTLEMENT"
	InstructionTypeFeePayment   = "FEE_PAYMENT"
	InstructionTypeDistribution = "DISTRIBUTION"
)

// Bank Instruction Status constants
const (
	BankInstructionStatusActive     = "ACTIVE"
	BankInstructionStatusInactive   = "INACTIVE"
	BankInstructionStatusSuperseded = "SUPERSEDED"
)

// HedgeFundTrade represents subscription, redemption and transfer orders
type HedgeFundTrade struct {
	TradeID        uuid.UUID  `json:"trade_id" db:"trade_id"`
	TradeReference string     `json:"trade_reference" db:"trade_reference"`
	InvestorID     uuid.UUID  `json:"investor_id" db:"investor_id"`
	FundID         uuid.UUID  `json:"fund_id" db:"fund_id"`
	ClassID        uuid.UUID  `json:"class_id" db:"class_id"`
	SeriesID       *uuid.UUID `json:"series_id,omitempty" db:"series_id"`

	// Trade details
	TradeType      string     `json:"trade_type" db:"trade_type"`
	Status         string     `json:"status" db:"status"`
	TradeDate      time.Time  `json:"trade_date" db:"trade_date"`
	ValueDate      time.Time  `json:"value_date" db:"value_date"`
	SettlementDate *time.Time `json:"settlement_date,omitempty" db:"settlement_date"`
	NAVDate        *time.Time `json:"nav_date,omitempty" db:"nav_date"`

	// Amounts and pricing
	RequestedAmount *float64 `json:"requested_amount,omitempty" db:"requested_amount"`
	GrossAmount     *float64 `json:"gross_amount,omitempty" db:"gross_amount"`
	FeesAmount      float64  `json:"fees_amount" db:"fees_amount"`
	NetAmount       *float64 `json:"net_amount,omitempty" db:"net_amount"`
	NAVPerShare     *float64 `json:"nav_per_share,omitempty" db:"nav_per_share"`
	Units           *float64 `json:"units,omitempty" db:"units"`

	// Currency and FX
	Currency           string   `json:"currency" db:"currency"`
	FXRate             float64  `json:"fx_rate" db:"fx_rate"`
	BaseCurrencyAmount *float64 `json:"base_currency_amount,omitempty" db:"base_currency_amount"`

	// Settlement
	BankID              *uuid.UUID `json:"bank_id,omitempty" db:"bank_id"`
	SettlementReference *string    `json:"settlement_reference,omitempty" db:"settlement_reference"`

	// Workflow
	IdempotencyKey *string    `json:"idempotency_key,omitempty" db:"idempotency_key"`
	NoticeDate     *time.Time `json:"notice_date,omitempty" db:"notice_date"`
	Comments       *string    `json:"comments,omitempty" db:"comments"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Trade Type constants
const (
	TradeTypeSub         = "SUB"
	TradeTypeRed         = "RED"
	TradeTypeTransferIn  = "TRANSFER_IN"
	TradeTypeTransferOut = "TRANSFER_OUT"
	TradeTypeSwitchIn    = "SWITCH_IN"
	TradeTypeSwitchOut   = "SWITCH_OUT"
	TradeTypeCorpAction  = "CORP_ACTION"
)

// Trade Status constants
const (
	TradeStatusPending   = "PENDING"
	TradeStatusAllocated = "ALLOCATED"
	TradeStatusSettled   = "SETTLED"
	TradeStatusCanceled  = "CANCELED"
	TradeStatusRejected  = "REJECTED"
)

// HedgeFundRegisterLot represents current unit holdings
type HedgeFundRegisterLot struct {
	LotID      uuid.UUID  `json:"lot_id" db:"lot_id"`
	InvestorID uuid.UUID  `json:"investor_id" db:"investor_id"`
	FundID     uuid.UUID  `json:"fund_id" db:"fund_id"`
	ClassID    uuid.UUID  `json:"class_id" db:"class_id"`
	SeriesID   *uuid.UUID `json:"series_id,omitempty" db:"series_id"`

	// Current position
	Units       float64  `json:"units" db:"units"`
	AverageCost *float64 `json:"average_cost,omitempty" db:"average_cost"`
	TotalCost   float64  `json:"total_cost" db:"total_cost"`

	// Dates
	FirstTradeDate *time.Time `json:"first_trade_date,omitempty" db:"first_trade_date"`
	LastActivityAt *time.Time `json:"last_activity_at,omitempty" db:"last_activity_at"`

	// Status
	Status string `json:"status" db:"status"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Register Lot Status constants
const (
	RegisterLotStatusActive = "ACTIVE"
	RegisterLotStatusClosed = "CLOSED"
)

// HedgeFundRegisterEvent represents immutable unit movement events
type HedgeFundRegisterEvent struct {
	EventID uuid.UUID  `json:"event_id" db:"event_id"`
	LotID   uuid.UUID  `json:"lot_id" db:"lot_id"`
	TradeID *uuid.UUID `json:"trade_id,omitempty" db:"trade_id"`

	// Event details
	EventType      string    `json:"event_type" db:"event_type"`
	EventTimestamp time.Time `json:"event_timestamp" db:"event_timestamp"`
	ValueDate      time.Time `json:"value_date" db:"value_date"`

	// Unit movement
	DeltaUnits     float64 `json:"delta_units" db:"delta_units"`
	RunningBalance float64 `json:"running_balance" db:"running_balance"`

	// Pricing
	NAVPerShare   *float64 `json:"nav_per_share,omitempty" db:"nav_per_share"`
	PricePerShare *float64 `json:"price_per_share,omitempty" db:"price_per_share"`

	// Amounts
	GrossAmount *float64 `json:"gross_amount,omitempty" db:"gross_amount"`
	FeesAmount  float64  `json:"fees_amount" db:"fees_amount"`
	NetAmount   *float64 `json:"net_amount,omitempty" db:"net_amount"`

	// Metadata
	Description       *string `json:"description,omitempty" db:"description"`
	ExternalReference *string `json:"external_reference,omitempty" db:"external_reference"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Register Event Type constants
const (
	RegisterEventTypeIssue       = "ISSUE"
	RegisterEventTypeRedeem      = "REDEEM"
	RegisterEventTypeTransferIn  = "TRANSFER_IN"
	RegisterEventTypeTransferOut = "TRANSFER_OUT"
	RegisterEventTypeCorpAction  = "CORP_ACTION"
	RegisterEventTypeFeeCharge   = "FEE_CHARGE"
	RegisterEventTypeDividend    = "DIVIDEND"
)
