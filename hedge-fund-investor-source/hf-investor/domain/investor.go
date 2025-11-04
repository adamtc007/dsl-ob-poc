package domain

import (
	"time"

	"github.com/google/uuid"
)

// HedgeFundInvestor represents a hedge fund investor entity with full lifecycle tracking
type HedgeFundInvestor struct {
	// Core identity
	InvestorID   uuid.UUID `json:"investor_id" db:"investor_id"`
	InvestorCode string    `json:"investor_code" db:"investor_code"`
	Type         string    `json:"type" db:"type"`
	LegalName    string    `json:"legal_name" db:"legal_name"`
	ShortName    *string   `json:"short_name,omitempty" db:"short_name"`

	// Regulatory identifiers
	LEI                *string `json:"lei,omitempty" db:"lei"`
	RegistrationNumber *string `json:"registration_number,omitempty" db:"registration_number"`
	Domicile           string  `json:"domicile" db:"domicile"`

	// Address
	AddressLine1  *string `json:"address_line1,omitempty" db:"address_line1"`
	AddressLine2  *string `json:"address_line2,omitempty" db:"address_line2"`
	AddressLine3  *string `json:"address_line3,omitempty" db:"address_line3"`
	AddressLine4  *string `json:"address_line4,omitempty" db:"address_line4"`
	City          *string `json:"city,omitempty" db:"city"`
	StateProvince *string `json:"state_province,omitempty" db:"state_province"`
	PostalCode    *string `json:"postal_code,omitempty" db:"postal_code"`
	Country       *string `json:"country,omitempty" db:"country"`

	// Contact information
	PrimaryContactName  *string `json:"primary_contact_name,omitempty" db:"primary_contact_name"`
	PrimaryContactEmail *string `json:"primary_contact_email,omitempty" db:"primary_contact_email"`
	PrimaryContactPhone *string `json:"primary_contact_phone,omitempty" db:"primary_contact_phone"`

	// Lifecycle
	Status string  `json:"status" db:"status"`
	Source *string `json:"source,omitempty" db:"source"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// InvestorType constants for hedge fund investor types
const (
	InvestorTypeIndividual  = "INDIVIDUAL"
	InvestorTypeCorporate   = "CORPORATE"
	InvestorTypeTrust       = "TRUST"
	InvestorTypeFOHF        = "FOHF"
	InvestorTypeNominee     = "NOMINEE"
	InvestorTypePensionFund = "PENSION_FUND"
	InvestorTypeInsuranceCo = "INSURANCE_CO"
)

// InvestorStatus constants for hedge fund investor lifecycle states
const (
	InvestorStatusOpportunity      = "OPPORTUNITY"
	InvestorStatusPrechecks        = "PRECHECKS"
	InvestorStatusKYCPending       = "KYC_PENDING"
	InvestorStatusKYCApproved      = "KYC_APPROVED"
	InvestorStatusSubPendingCash   = "SUB_PENDING_CASH"
	InvestorStatusFundedPendingNAV = "FUNDED_PENDING_NAV"
	InvestorStatusIssued           = "ISSUED"
	InvestorStatusActive           = "ACTIVE"
	InvestorStatusRedeemPending    = "REDEEM_PENDING"
	InvestorStatusRedeemed         = "REDEEMED"
	InvestorStatusOffboarded       = "OFFBOARDED"
)

// IsValidInvestorType checks if the provided investor type is valid
func IsValidInvestorType(t string) bool {
	switch t {
	case InvestorTypeIndividual, InvestorTypeCorporate, InvestorTypeTrust,
		InvestorTypeFOHF, InvestorTypeNominee, InvestorTypePensionFund,
		InvestorTypeInsuranceCo:
		return true
	default:
		return false
	}
}

// IsValidInvestorStatus checks if the provided investor status is valid
func IsValidInvestorStatus(s string) bool {
	switch s {
	case InvestorStatusOpportunity, InvestorStatusPrechecks, InvestorStatusKYCPending,
		InvestorStatusKYCApproved, InvestorStatusSubPendingCash, InvestorStatusFundedPendingNAV,
		InvestorStatusIssued, InvestorStatusActive, InvestorStatusRedeemPending,
		InvestorStatusRedeemed, InvestorStatusOffboarded:
		return true
	default:
		return false
	}
}

// CanTransitionTo checks if the investor can transition from current status to target status
func (i *HedgeFundInvestor) CanTransitionTo(targetStatus string) bool {
	validTransitions := map[string][]string{
		InvestorStatusOpportunity:      {InvestorStatusPrechecks},
		InvestorStatusPrechecks:        {InvestorStatusKYCPending},
		InvestorStatusKYCPending:       {InvestorStatusKYCApproved, InvestorStatusOpportunity}, // Can go back
		InvestorStatusKYCApproved:      {InvestorStatusSubPendingCash},
		InvestorStatusSubPendingCash:   {InvestorStatusFundedPendingNAV, InvestorStatusKYCApproved}, // Can go back
		InvestorStatusFundedPendingNAV: {InvestorStatusIssued},
		InvestorStatusIssued:           {InvestorStatusActive},
		InvestorStatusActive:           {InvestorStatusRedeemPending, InvestorStatusSubPendingCash}, // Can add more money
		InvestorStatusRedeemPending:    {InvestorStatusRedeemed, InvestorStatusActive},              // Can cancel redemption
		InvestorStatusRedeemed:         {InvestorStatusOffboarded, InvestorStatusSubPendingCash},    // Can reinvest
		InvestorStatusOffboarded:       {},                                                          // Terminal state
	}

	allowedTargets, exists := validTransitions[i.Status]
	if !exists {
		return false
	}

	for _, allowed := range allowedTargets {
		if allowed == targetStatus {
			return true
		}
	}
	return false
}

// HedgeFundBeneficialOwner represents ultimate beneficial ownership information
type HedgeFundBeneficialOwner struct {
	BOID                uuid.UUID  `json:"bo_id" db:"bo_id"`
	InvestorID          uuid.UUID  `json:"investor_id" db:"investor_id"`
	FullName            string     `json:"full_name" db:"full_name"`
	DateOfBirth         *time.Time `json:"date_of_birth,omitempty" db:"date_of_birth"`
	Nationality         *string    `json:"nationality,omitempty" db:"nationality"`
	OwnershipPercentage float64    `json:"ownership_percentage" db:"ownership_percentage"`
	ControlType         *string    `json:"control_type,omitempty" db:"control_type"`

	// Risk flags
	IsPEP            bool    `json:"is_pep" db:"is_pep"`
	PEPDetails       *string `json:"pep_details,omitempty" db:"pep_details"`
	SanctionsFlag    bool    `json:"sanctions_flag" db:"sanctions_flag"`
	SanctionsDetails *string `json:"sanctions_details,omitempty" db:"sanctions_details"`

	// Address
	AddressLine1 *string `json:"address_line1,omitempty" db:"address_line1"`
	City         *string `json:"city,omitempty" db:"city"`
	Country      *string `json:"country,omitempty" db:"country"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// ControlType constants for beneficial ownership control types
const (
	ControlTypeOwnership              = "OWNERSHIP"
	ControlTypeVoting                 = "VOTING"
	ControlTypeControl                = "CONTROL"
	ControlTypeSeniorManagingOfficial = "SENIOR_MANAGING_OFFICIAL"
)

// HedgeFundKYCProfile represents KYC/KYB profile for hedge fund investors
type HedgeFundKYCProfile struct {
	KYCID              uuid.UUID  `json:"kyc_id" db:"kyc_id"`
	InvestorID         uuid.UUID  `json:"investor_id" db:"investor_id"`
	RiskRating         string     `json:"risk_rating" db:"risk_rating"`
	Status             string     `json:"status" db:"status"`
	KYCTier            *string    `json:"kyc_tier,omitempty" db:"kyc_tier"`
	ScreeningProvider  *string    `json:"screening_provider,omitempty" db:"screening_provider"`
	ScreeningReference *string    `json:"screening_reference,omitempty" db:"screening_reference"`
	ScreeningDate      *time.Time `json:"screening_date,omitempty" db:"screening_date"`
	ScreeningResult    *string    `json:"screening_result,omitempty" db:"screening_result"`

	// Approval
	ApprovedBy       *string    `json:"approved_by,omitempty" db:"approved_by"`
	ApprovedAt       *time.Time `json:"approved_at,omitempty" db:"approved_at"`
	ApprovalComments *string    `json:"approval_comments,omitempty" db:"approval_comments"`

	// Refresh schedule
	RefreshFrequency string     `json:"refresh_frequency" db:"refresh_frequency"`
	RefreshDueAt     *time.Time `json:"refresh_due_at,omitempty" db:"refresh_due_at"`
	LastRefreshedAt  *time.Time `json:"last_refreshed_at,omitempty" db:"last_refreshed_at"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// KYC Risk Rating constants
const (
	KYCRiskRatingLow        = "LOW"
	KYCRiskRatingMedium     = "MEDIUM"
	KYCRiskRatingHigh       = "HIGH"
	KYCRiskRatingProhibited = "PROHIBITED"
)

// KYC Status constants
const (
	KYCStatusPending  = "PENDING"
	KYCStatusInReview = "IN_REVIEW"
	KYCStatusApproved = "APPROVED"
	KYCStatusRejected = "REJECTED"
	KYCStatusExpired  = "EXPIRED"
)

// KYC Tier constants
const (
	KYCTierSimplified = "SIMPLIFIED"
	KYCTierStandard   = "STANDARD"
	KYCTierEnhanced   = "ENHANCED"
)

// Screening Result constants
const (
	ScreeningResultClear          = "CLEAR"
	ScreeningResultPotentialMatch = "POTENTIAL_MATCH"
	ScreeningResultTruePositive   = "TRUE_POSITIVE"
)

// Refresh Frequency constants
const (
	RefreshFrequencyMonthly    = "MONTHLY"
	RefreshFrequencyQuarterly  = "QUARTERLY"
	RefreshFrequencySemiAnnual = "SEMI_ANNUAL"
	RefreshFrequencyAnnual     = "ANNUAL"
	RefreshFrequencyBiennial   = "BIENNIAL"
)

// HedgeFundTaxProfile represents tax classification and withholding information
type HedgeFundTaxProfile struct {
	TaxID      uuid.UUID `json:"tax_id" db:"tax_id"`
	InvestorID uuid.UUID `json:"investor_id" db:"investor_id"`

	// FATCA classification
	FATCAStatus *string `json:"fatca_status,omitempty" db:"fatca_status"`
	FATCAGIIN   *string `json:"fatca_giin,omitempty" db:"fatca_giin"`

	// CRS classification
	CRSClassification *string `json:"crs_classification,omitempty" db:"crs_classification"`
	CRSJurisdiction   *string `json:"crs_jurisdiction,omitempty" db:"crs_jurisdiction"`

	// Tax forms
	FormType       *string    `json:"form_type,omitempty" db:"form_type"`
	FormDate       *time.Time `json:"form_date,omitempty" db:"form_date"`
	FormValidUntil *time.Time `json:"form_valid_until,omitempty" db:"form_valid_until"`

	// Tax rates
	WithholdingRate   float64 `json:"withholding_rate" db:"withholding_rate"`
	BackupWithholding bool    `json:"backup_withholding" db:"backup_withholding"`

	// TIN information
	TINType         *string `json:"tin_type,omitempty" db:"tin_type"`
	TINValue        *string `json:"tin_value,omitempty" db:"tin_value"`
	TINJurisdiction *string `json:"tin_jurisdiction,omitempty" db:"tin_jurisdiction"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// FATCA Status constants
const (
	FATCAStatusUSPerson              = "US_PERSON"
	FATCAStatusNonUSPerson           = "NON_US_PERSON"
	FATCAStatusSpecifiedUSPerson     = "SPECIFIED_US_PERSON"
	FATCAStatusExemptBeneficialOwner = "EXEMPT_BENEFICIAL_OWNER"
)

// CRS Classification constants
const (
	CRSClassificationIndividual           = "INDIVIDUAL"
	CRSClassificationEntity               = "ENTITY"
	CRSClassificationFinancialInstitution = "FINANCIAL_INSTITUTION"
	CRSClassificationInvestmentEntity     = "INVESTMENT_ENTITY"
)

// Form Type constants
const (
	FormTypeW9             = "W9"
	FormTypeW8BEN          = "W8_BEN"
	FormTypeW8BENE         = "W8_BEN_E"
	FormTypeW8ECI          = "W8_ECI"
	FormTypeW8EXP          = "W8_EXP"
	FormTypeW8IMY          = "W8_IMY"
	FormTypeEntitySelfCert = "ENTITY_SELF_CERT"
)

// TIN Type constants
const (
	TINTypeSSN        = "SSN"
	TINTypeITIN       = "ITIN"
	TINTypeEIN        = "EIN"
	TINTypeForeignTIN = "FOREIGN_TIN"
	TINTypeGIIN       = "GIIN"
)
