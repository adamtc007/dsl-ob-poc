package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestInvestorTypeValidation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Valid Individual", InvestorTypeIndividual, true},
		{"Valid Corporate", InvestorTypeCorporate, true},
		{"Valid Trust", InvestorTypeTrust, true},
		{"Valid FOHF", InvestorTypeFOHF, true},
		{"Valid Nominee", InvestorTypeNominee, true},
		{"Valid Pension Fund", InvestorTypePensionFund, true},
		{"Valid Insurance Co", InvestorTypeInsuranceCo, true},
		{"Invalid Type", "INVALID_TYPE", false},
		{"Empty String", "", false},
		{"Lowercase", "individual", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidInvestorType(tt.input)
			if result != tt.expected {
				t.Errorf("IsValidInvestorType(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestInvestorStatusValidation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Valid Opportunity", InvestorStatusOpportunity, true},
		{"Valid Prechecks", InvestorStatusPrechecks, true},
		{"Valid KYC Pending", InvestorStatusKYCPending, true},
		{"Valid KYC Approved", InvestorStatusKYCApproved, true},
		{"Valid Sub Pending Cash", InvestorStatusSubPendingCash, true},
		{"Valid Funded Pending NAV", InvestorStatusFundedPendingNAV, true},
		{"Valid Issued", InvestorStatusIssued, true},
		{"Valid Active", InvestorStatusActive, true},
		{"Valid Redeem Pending", InvestorStatusRedeemPending, true},
		{"Valid Redeemed", InvestorStatusRedeemed, true},
		{"Valid Offboarded", InvestorStatusOffboarded, true},
		{"Invalid Status", "INVALID_STATUS", false},
		{"Empty String", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidInvestorStatus(tt.input)
			if result != tt.expected {
				t.Errorf("IsValidInvestorStatus(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCanTransitionTo(t *testing.T) {
	tests := []struct {
		name         string
		currentState string
		targetState  string
		expected     bool
	}{
		// Valid transitions
		{"Opportunity to Prechecks", InvestorStatusOpportunity, InvestorStatusPrechecks, true},
		{"Prechecks to KYC Pending", InvestorStatusPrechecks, InvestorStatusKYCPending, true},
		{"KYC Pending to KYC Approved", InvestorStatusKYCPending, InvestorStatusKYCApproved, true},
		{"KYC Approved to Sub Pending Cash", InvestorStatusKYCApproved, InvestorStatusSubPendingCash, true},
		{"Sub Pending Cash to Funded Pending NAV", InvestorStatusSubPendingCash, InvestorStatusFundedPendingNAV, true},
		{"Funded Pending NAV to Issued", InvestorStatusFundedPendingNAV, InvestorStatusIssued, true},
		{"Issued to Active", InvestorStatusIssued, InvestorStatusActive, true},
		{"Active to Redeem Pending", InvestorStatusActive, InvestorStatusRedeemPending, true},
		{"Redeem Pending to Redeemed", InvestorStatusRedeemPending, InvestorStatusRedeemed, true},
		{"Redeemed to Offboarded", InvestorStatusRedeemed, InvestorStatusOffboarded, true},

		// Valid backward transitions
		{"KYC Pending back to Prechecks", InvestorStatusKYCPending, InvestorStatusPrechecks, true},
		{"Sub Pending Cash back to KYC Approved", InvestorStatusSubPendingCash, InvestorStatusKYCApproved, true},
		{"Redeem Pending back to Active", InvestorStatusRedeemPending, InvestorStatusActive, true},

		// Valid re-investment transitions
		{"Active to Sub Pending Cash", InvestorStatusActive, InvestorStatusSubPendingCash, true},
		{"Redeemed to Sub Pending Cash", InvestorStatusRedeemed, InvestorStatusSubPendingCash, true},

		// Invalid transitions
		{"Opportunity to KYC Pending", InvestorStatusOpportunity, InvestorStatusKYCPending, false},
		{"Opportunity to Active", InvestorStatusOpportunity, InvestorStatusActive, false},
		{"KYC Approved to Issued", InvestorStatusKYCApproved, InvestorStatusIssued, false},
		{"Issued to Redeemed", InvestorStatusIssued, InvestorStatusRedeemed, false},
		{"Offboarded to any state", InvestorStatusOffboarded, InvestorStatusActive, false},

		// Self-transitions (invalid)
		{"Same state transition", InvestorStatusActive, InvestorStatusActive, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			investor := &HedgeFundInvestor{
				InvestorID:   uuid.New(),
				InvestorCode: "TEST-001",
				Status:       tt.currentState,
				Type:         InvestorTypeCorporate,
				LegalName:    "Test Investor",
				Domicile:     "US",
			}

			result := investor.CanTransitionTo(tt.targetState)
			if result != tt.expected {
				t.Errorf("CanTransitionTo from %s to %s = %v, want %v",
					tt.currentState, tt.targetState, result, tt.expected)
			}
		})
	}
}

func TestKYCRiskRatingConstants(t *testing.T) {
	expectedRatings := []string{
		KYCRiskRatingLow,
		KYCRiskRatingMedium,
		KYCRiskRatingHigh,
		KYCRiskRatingProhibited,
	}

	actualRatings := []string{"LOW", "MEDIUM", "HIGH", "PROHIBITED"}

	for i, expected := range expectedRatings {
		if expected != actualRatings[i] {
			t.Errorf("KYC Risk Rating constant mismatch: got %s, want %s", expected, actualRatings[i])
		}
	}
}

func TestKYCStatusConstants(t *testing.T) {
	expectedStatuses := []string{
		KYCStatusPending,
		KYCStatusInReview,
		KYCStatusApproved,
		KYCStatusRejected,
		KYCStatusExpired,
	}

	actualStatuses := []string{"PENDING", "IN_REVIEW", "APPROVED", "REJECTED", "EXPIRED"}

	for i, expected := range expectedStatuses {
		if expected != actualStatuses[i] {
			t.Errorf("KYC Status constant mismatch: got %s, want %s", expected, actualStatuses[i])
		}
	}
}

func TestFATCAStatusConstants(t *testing.T) {
	expectedStatuses := []string{
		FATCAStatusUSPerson,
		FATCAStatusNonUSPerson,
		FATCAStatusSpecifiedUSPerson,
		FATCAStatusExemptBeneficialOwner,
	}

	actualStatuses := []string{"US_PERSON", "NON_US_PERSON", "SPECIFIED_US_PERSON", "EXEMPT_BENEFICIAL_OWNER"}

	for i, expected := range expectedStatuses {
		if expected != actualStatuses[i] {
			t.Errorf("FATCA Status constant mismatch: got %s, want %s", expected, actualStatuses[i])
		}
	}
}

func TestControlTypeConstants(t *testing.T) {
	expectedTypes := []string{
		ControlTypeOwnership,
		ControlTypeVoting,
		ControlTypeControl,
		ControlTypeSeniorManagingOfficial,
	}

	actualTypes := []string{"OWNERSHIP", "VOTING", "CONTROL", "SENIOR_MANAGING_OFFICIAL"}

	for i, expected := range expectedTypes {
		if expected != actualTypes[i] {
			t.Errorf("Control Type constant mismatch: got %s, want %s", expected, actualTypes[i])
		}
	}
}

func TestFormTypeConstants(t *testing.T) {
	expectedForms := []string{
		FormTypeW9,
		FormTypeW8BEN,
		FormTypeW8BENE,
		FormTypeW8ECI,
		FormTypeW8EXP,
		FormTypeW8IMY,
		FormTypeEntitySelfCert,
	}

	actualForms := []string{"W9", "W8_BEN", "W8_BEN_E", "W8_ECI", "W8_EXP", "W8_IMY", "ENTITY_SELF_CERT"}

	for i, expected := range expectedForms {
		if expected != actualForms[i] {
			t.Errorf("Form Type constant mismatch: got %s, want %s", expected, actualForms[i])
		}
	}
}

func TestInvestorCreation(t *testing.T) {
	investor := &HedgeFundInvestor{
		InvestorID:   uuid.New(),
		InvestorCode: "TEST-001",
		Type:         InvestorTypeCorporate,
		LegalName:    "Test Corporation",
		Domicile:     "US",
		Status:       InvestorStatusOpportunity,
	}

	// Test that investor is created with correct fields
	if investor.InvestorCode != "TEST-001" {
		t.Errorf("Expected investor code TEST-001, got %s", investor.InvestorCode)
	}

	if investor.Type != InvestorTypeCorporate {
		t.Errorf("Expected investor type %s, got %s", InvestorTypeCorporate, investor.Type)
	}

	if investor.Status != InvestorStatusOpportunity {
		t.Errorf("Expected status %s, got %s", InvestorStatusOpportunity, investor.Status)
	}

	// Test that investor can transition to valid next state
	if !investor.CanTransitionTo(InvestorStatusPrechecks) {
		t.Errorf("Expected investor to be able to transition to %s", InvestorStatusPrechecks)
	}

	// Test that investor cannot transition to invalid state
	if investor.CanTransitionTo(InvestorStatusActive) {
		t.Errorf("Expected investor to NOT be able to transition directly to %s", InvestorStatusActive)
	}
}
