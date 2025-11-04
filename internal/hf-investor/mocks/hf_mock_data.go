package mocks

import (
	"context"
	"fmt"
	"time"

	"dsl-ob-poc/internal/hf-investor/domain"
	"dsl-ob-poc/internal/hf-investor/store"

	"github.com/google/uuid"
)

// MockHedgeFundData provides mock data for hedge fund investor testing and development
type MockHedgeFundData struct {
	Funds            []domain.HedgeFund
	ShareClasses     []domain.HedgeFundShareClass
	Series           []domain.HedgeFundSeries
	Investors        []domain.HedgeFundInvestor
	KYCProfiles      []domain.HedgeFundKYCProfile
	TaxProfiles      []domain.HedgeFundTaxProfile
	BankInstructions []domain.HedgeFundBankInstruction
	Trades           []domain.HedgeFundTrade
	RegisterLots     []domain.HedgeFundRegisterLot
	Documents        []store.HedgeFundDocument
}

// GetMockHedgeFundData returns comprehensive mock data for hedge fund operations
func GetMockHedgeFundData() *MockHedgeFundData {
	now := time.Now().UTC()

	// Fixed UUIDs for consistent testing
	fundID1 := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	fundID2 := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	classID1A := uuid.MustParse("11111111-aaaa-1111-1111-111111111111")
	classID1B := uuid.MustParse("11111111-bbbb-1111-1111-111111111111")
	classID2A := uuid.MustParse("22222222-aaaa-2222-2222-222222222222")

	seriesID1A1 := uuid.MustParse("11111111-aaaa-1111-0001-111111111111")
	seriesID1A2 := uuid.MustParse("11111111-aaaa-1111-0002-111111111111")

	investorID1 := uuid.MustParse("aaaaaaaa-1111-1111-1111-aaaaaaaaaaaa")
	investorID2 := uuid.MustParse("bbbbbbbb-2222-2222-2222-bbbbbbbbbbbb")
	investorID3 := uuid.MustParse("cccccccc-3333-3333-3333-cccccccccccc")

	// Mock Funds
	funds := []domain.HedgeFund{
		{
			FundID:        fundID1,
			FundName:      "Alpha Global Opportunities Fund",
			LegalName:     "Alpha Global Opportunities Fund Limited",
			LEI:           stringPtr("5493001RKR55V4X61F71"),
			Domicile:      "KY", // Cayman Islands
			FundType:      domain.FundTypeHedgeFund,
			Currency:      "USD",
			InceptionDate: time.Date(2020, 1, 15, 0, 0, 0, 0, time.UTC),
			Status:        domain.FundStatusActive,
			Administrator: stringPtr("Northern Trust International Fund Administration Services"),
			Custodian:     stringPtr("Northern Trust Company"),
			Auditor:       stringPtr("KPMG Cayman Islands"),
			CreatedAt:     now.AddDate(-4, 0, 0),
			UpdatedAt:     now.AddDate(0, -1, 0),
		},
		{
			FundID:        fundID2,
			FundName:      "Beta Credit Strategies Fund",
			LegalName:     "Beta Credit Strategies Fund SPC",
			LEI:           stringPtr("5493001RKR55V4X61F72"),
			Domicile:      "KY",
			FundType:      domain.FundTypeCreditFund,
			Currency:      "USD",
			InceptionDate: time.Date(2021, 6, 1, 0, 0, 0, 0, time.UTC),
			Status:        domain.FundStatusActive,
			Administrator: stringPtr("Citco Fund Services"),
			Custodian:     stringPtr("State Street Bank"),
			Auditor:       stringPtr("Ernst & Young"),
			CreatedAt:     now.AddDate(-3, 0, 0),
			UpdatedAt:     now.AddDate(0, -2, 0),
		},
	}

	// Mock Share Classes
	shareClasses := []domain.HedgeFundShareClass{
		{
			ClassID:                 classID1A,
			FundID:                  fundID1,
			ClassName:               "A",
			ClassType:               domain.ClassTypeInstitutional,
			Currency:                "USD",
			MinInitialInvestment:    1000000.00,
			MinSubsequentInvestment: 100000.00,
			ManagementFeeRate:       0.015, // 1.5%
			PerformanceFeeRate:      0.20,  // 20%
			HighWaterMark:           true,
			DealingFrequency:        domain.DealingFrequencyMonthly,
			NoticePeriodDays:        90,
			GatePercentage:          floatPtr(25.0),
			LockupMonths:            12,
			CreatedAt:               now.AddDate(-4, 0, 0),
			UpdatedAt:               now.AddDate(0, -1, 0),
		},
		{
			ClassID:                 classID1B,
			FundID:                  fundID1,
			ClassName:               "B",
			ClassType:               domain.ClassTypeRetail,
			Currency:                "USD",
			MinInitialInvestment:    250000.00,
			MinSubsequentInvestment: 25000.00,
			ManagementFeeRate:       0.02, // 2.0%
			PerformanceFeeRate:      0.20, // 20%
			HighWaterMark:           true,
			DealingFrequency:        domain.DealingFrequencyMonthly,
			NoticePeriodDays:        60,
			GatePercentage:          floatPtr(25.0),
			LockupMonths:            6,
			CreatedAt:               now.AddDate(-4, 0, 0),
			UpdatedAt:               now.AddDate(0, -1, 0),
		},
		{
			ClassID:                 classID2A,
			FundID:                  fundID2,
			ClassName:               "A",
			ClassType:               domain.ClassTypeInstitutional,
			Currency:                "USD",
			MinInitialInvestment:    5000000.00,
			MinSubsequentInvestment: 500000.00,
			ManagementFeeRate:       0.0125, // 1.25%
			PerformanceFeeRate:      0.15,   // 15%
			HighWaterMark:           true,
			DealingFrequency:        domain.DealingFrequencyQuarterly,
			NoticePeriodDays:        120,
			GatePercentage:          floatPtr(20.0),
			LockupMonths:            24,
			CreatedAt:               now.AddDate(-3, 0, 0),
			UpdatedAt:               now.AddDate(0, -2, 0),
		},
	}

	// Mock Series
	series := []domain.HedgeFundSeries{
		{
			SeriesID:      seriesID1A1,
			ClassID:       classID1A,
			SeriesName:    "2024-01",
			InceptionDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Status:        domain.SeriesStatusActive,
			CreatedAt:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:     now.AddDate(0, -1, 0),
		},
		{
			SeriesID:      seriesID1A2,
			ClassID:       classID1A,
			SeriesName:    "2024-06",
			InceptionDate: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
			Status:        domain.SeriesStatusActive,
			CreatedAt:     time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:     now.AddDate(0, -1, 0),
		},
	}

	// Mock Investors
	investors := []domain.HedgeFundInvestor{
		{
			InvestorID:          investorID1,
			InvestorCode:        "INV-001",
			Type:                domain.InvestorTypeCorporate,
			LegalName:           "Pension Fund Solutions Inc.",
			ShortName:           stringPtr("PFS Inc."),
			LEI:                 stringPtr("5493001RKR55V4X61F73"),
			RegistrationNumber:  stringPtr("12345-DE"),
			Domicile:            "US",
			AddressLine1:        stringPtr("100 Wall Street"),
			AddressLine2:        stringPtr("Suite 2500"),
			City:                stringPtr("New York"),
			StateProvince:       stringPtr("NY"),
			PostalCode:          stringPtr("10005"),
			Country:             stringPtr("US"),
			PrimaryContactName:  stringPtr("John Smith"),
			PrimaryContactEmail: stringPtr("john.smith@pfs.com"),
			PrimaryContactPhone: stringPtr("+1-212-555-0100"),
			Status:              domain.InvestorStatusActive,
			Source:              stringPtr("Referral"),
			CreatedAt:           now.AddDate(-1, 0, 0),
			UpdatedAt:           now.AddDate(0, 0, -7),
		},
		{
			InvestorID:          investorID2,
			InvestorCode:        "INV-002",
			Type:                domain.InvestorTypeIndividual,
			LegalName:           "Dr. Sarah Elizabeth Johnson",
			Domicile:            "GB",
			AddressLine1:        stringPtr("45 Eaton Square"),
			City:                stringPtr("London"),
			PostalCode:          stringPtr("SW1W 9BA"),
			Country:             stringPtr("GB"),
			PrimaryContactName:  stringPtr("Dr. Sarah Johnson"),
			PrimaryContactEmail: stringPtr("sarah.johnson@email.com"),
			Status:              domain.InvestorStatusKYCPending,
			Source:              stringPtr("Website"),
			CreatedAt:           now.AddDate(0, -1, 0),
			UpdatedAt:           now.AddDate(0, 0, -3),
		},
		{
			InvestorID:          investorID3,
			InvestorCode:        "INV-003",
			Type:                domain.InvestorTypeFOHF,
			LegalName:           "Global Multi-Manager Fund of Funds LP",
			ShortName:           stringPtr("GMMF LP"),
			LEI:                 stringPtr("5493001RKR55V4X61F74"),
			RegistrationNumber:  stringPtr("LP-987654"),
			Domicile:            "KY",
			AddressLine1:        stringPtr("Governors Square"),
			AddressLine2:        stringPtr("23 Lime Tree Bay Avenue"),
			City:                stringPtr("George Town"),
			Country:             stringPtr("KY"),
			PrimaryContactName:  stringPtr("Michael Chen"),
			PrimaryContactEmail: stringPtr("michael.chen@gmmf.ky"),
			Status:              domain.InvestorStatusSubPendingCash,
			Source:              stringPtr("Broker"),
			CreatedAt:           now.AddDate(0, -6, 0),
			UpdatedAt:           now.AddDate(0, 0, -1),
		},
	}

	// Mock KYC Profiles
	kycProfiles := []domain.HedgeFundKYCProfile{
		{
			KYCID:              uuid.New(),
			InvestorID:         investorID1,
			RiskRating:         domain.KYCRiskRatingLow,
			Status:             domain.KYCStatusApproved,
			KYCTier:            stringPtr(domain.KYCTierStandard),
			ScreeningProvider:  stringPtr("Refinitiv World-Check"),
			ScreeningReference: stringPtr("WC-2024-001234"),
			ScreeningDate:      timePtr(now.AddDate(0, -11, 0)),
			ScreeningResult:    stringPtr(domain.ScreeningResultClear),
			ApprovedBy:         stringPtr("Jane Doe"),
			ApprovedAt:         timePtr(now.AddDate(0, -11, 0)),
			ApprovalComments:   stringPtr("Standard institutional client, clean screening results"),
			RefreshFrequency:   domain.RefreshFrequencyAnnual,
			RefreshDueAt:       timePtr(now.AddDate(1, -11, 0)),
			LastRefreshedAt:    timePtr(now.AddDate(0, -11, 0)),
			CreatedAt:          now.AddDate(-1, 0, 0),
			UpdatedAt:          now.AddDate(0, -11, 0),
		},
		{
			KYCID:            uuid.New(),
			InvestorID:       investorID2,
			RiskRating:       domain.KYCRiskRatingMedium,
			Status:           domain.KYCStatusPending,
			KYCTier:          stringPtr(domain.KYCTierStandard),
			RefreshFrequency: domain.RefreshFrequencyAnnual,
			CreatedAt:        now.AddDate(0, -1, 0),
			UpdatedAt:        now.AddDate(0, 0, -3),
		},
	}

	// Mock Tax Profiles
	taxProfiles := []domain.HedgeFundTaxProfile{
		{
			TaxID:             uuid.New(),
			InvestorID:        investorID1,
			FATCAStatus:       stringPtr(domain.FATCAStatusNonUSPerson),
			CRSClassification: stringPtr(domain.CRSClassificationEntity),
			CRSJurisdiction:   stringPtr("US"),
			FormType:          stringPtr(domain.FormTypeW8BENE),
			FormDate:          timePtr(now.AddDate(0, -11, 0)),
			FormValidUntil:    timePtr(now.AddDate(2, -11, 0)),
			WithholdingRate:   0.0,
			BackupWithholding: false,
			TINType:           stringPtr(domain.TINTypeEIN),
			TINValue:          stringPtr("12-3456789"),
			TINJurisdiction:   stringPtr("US"),
			CreatedAt:         now.AddDate(-1, 0, 0),
			UpdatedAt:         now.AddDate(0, -11, 0),
		},
		{
			TaxID:             uuid.New(),
			InvestorID:        investorID2,
			FATCAStatus:       stringPtr(domain.FATCAStatusNonUSPerson),
			CRSClassification: stringPtr(domain.CRSClassificationIndividual),
			CRSJurisdiction:   stringPtr("GB"),
			FormType:          stringPtr(domain.FormTypeW8BEN),
			WithholdingRate:   0.0,
			BackupWithholding: false,
			TINType:           stringPtr(domain.TINTypeForeignTIN),
			TINValue:          stringPtr("QQ123456C"),
			TINJurisdiction:   stringPtr("GB"),
			CreatedAt:         now.AddDate(0, -1, 0),
			UpdatedAt:         now.AddDate(0, 0, -3),
		},
	}

	// Mock Bank Instructions
	bankInstructions := []domain.HedgeFundBankInstruction{
		{
			BankID:             uuid.New(),
			InvestorID:         investorID1,
			Currency:           "USD",
			InstructionType:    domain.InstructionTypeSettlement,
			BankName:           "JPMorgan Chase Bank, N.A.",
			SwiftBIC:           stringPtr("CHASUS33"),
			AccountNumber:      stringPtr("123456789012"),
			AccountName:        "Pension Fund Solutions Inc.",
			BankAddressLine1:   stringPtr("270 Park Avenue"),
			BankCity:           stringPtr("New York"),
			BankCountry:        stringPtr("US"),
			Status:             domain.BankInstructionStatusActive,
			VersionNumber:      1,
			VerifiedBy:         stringPtr("Operations Team"),
			VerifiedAt:         timePtr(now.AddDate(-1, 0, 0)),
			VerificationMethod: stringPtr("Manual Review"),
			CreatedAt:          now.AddDate(-1, 0, 0),
			UpdatedAt:          now.AddDate(-1, 0, 0),
		},
		{
			BankID:           uuid.New(),
			InvestorID:       investorID2,
			Currency:         "GBP",
			InstructionType:  domain.InstructionTypeSettlement,
			BankName:         "HSBC UK Bank plc",
			SwiftBIC:         stringPtr("HBUKGB4B"),
			IBAN:             stringPtr("GB82HBUK40187612345678"),
			AccountName:      "Dr. Sarah Elizabeth Johnson",
			BankAddressLine1: stringPtr("8 Canada Square"),
			BankCity:         stringPtr("London"),
			BankCountry:      stringPtr("GB"),
			Status:           domain.BankInstructionStatusActive,
			VersionNumber:    1,
			CreatedAt:        now.AddDate(0, -1, 0),
			UpdatedAt:        now.AddDate(0, -1, 0),
		},
	}

	// Mock Trades
	trades := []domain.HedgeFundTrade{
		{
			TradeID:            uuid.New(),
			TradeReference:     "SUB-001-2024",
			InvestorID:         investorID1,
			FundID:             fundID1,
			ClassID:            classID1A,
			SeriesID:           &seriesID1A1,
			TradeType:          domain.TradeTypeSub,
			Status:             domain.TradeStatusSettled,
			TradeDate:          time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			ValueDate:          time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
			SettlementDate:     timePtr(time.Date(2024, 2, 2, 0, 0, 0, 0, time.UTC)),
			NAVDate:            timePtr(time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)),
			RequestedAmount:    floatPtr(10000000.00),
			GrossAmount:        floatPtr(10000000.00),
			FeesAmount:         0.0,
			NetAmount:          floatPtr(10000000.00),
			NAVPerShare:        floatPtr(100.00),
			Units:              floatPtr(100000.00),
			Currency:           "USD",
			FXRate:             1.0,
			BaseCurrencyAmount: floatPtr(10000000.00),
			IdempotencyKey:     stringPtr("sub-001-2024-idempotent-key"),
			CreatedAt:          time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
			UpdatedAt:          time.Date(2024, 2, 2, 10, 0, 0, 0, time.UTC),
		},
		{
			TradeID:         uuid.New(),
			TradeReference:  "SUB-002-2024",
			InvestorID:      investorID3,
			FundID:          fundID2,
			ClassID:         classID2A,
			TradeType:       domain.TradeTypeSub,
			Status:          domain.TradeStatusPending,
			TradeDate:       now.AddDate(0, 0, -7),
			ValueDate:       now.AddDate(0, 1, 0), // Next month
			RequestedAmount: floatPtr(25000000.00),
			Currency:        "USD",
			FXRate:          1.0,
			IdempotencyKey:  stringPtr("sub-002-2024-idempotent-key"),
			Comments:        stringPtr("Large institutional subscription pending cash confirmation"),
			CreatedAt:       now.AddDate(0, 0, -7),
			UpdatedAt:       now.AddDate(0, 0, -1),
		},
	}

	// Mock Register Lots
	registerLots := []domain.HedgeFundRegisterLot{
		{
			LotID:          uuid.New(),
			InvestorID:     investorID1,
			FundID:         fundID1,
			ClassID:        classID1A,
			SeriesID:       &seriesID1A1,
			Units:          100000.00,
			AverageCost:    floatPtr(100.00),
			TotalCost:      10000000.00,
			FirstTradeDate: timePtr(time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)),
			LastActivityAt: timePtr(time.Date(2024, 2, 2, 0, 0, 0, 0, time.UTC)),
			Status:         domain.RegisterLotStatusActive,
			CreatedAt:      time.Date(2024, 2, 2, 0, 0, 0, 0, time.UTC),
			UpdatedAt:      time.Date(2024, 2, 2, 0, 0, 0, 0, time.UTC),
		},
	}

	// Mock Documents
	documents := []store.HedgeFundDocument{
		{
			DocumentID:      uuid.New(),
			InvestorID:      investorID1,
			DocumentType:    "certificate_of_incorporation",
			DocumentSubject: stringPtr("entity"),
			DocumentTitle:   "Certificate of Incorporation - Pension Fund Solutions Inc.",
			FileName:        stringPtr("pfs_cert_incorporation.pdf"),
			FileSize:        int64Ptr(1024576),
			MimeType:        stringPtr("application/pdf"),
			Status:          store.DocumentStatusApproved,
			ReviewedBy:      stringPtr("Legal Team"),
			ReviewedAt:      timePtr(now.AddDate(-1, 0, 0)),
			ReviewComments:  stringPtr("Valid incorporation certificate"),
			IssuedDate:      timePtr(time.Date(2020, 3, 15, 0, 0, 0, 0, time.UTC)),
			CreatedAt:       now.AddDate(-1, 0, 0),
			UpdatedAt:       now.AddDate(-1, 0, 0),
		},
		{
			DocumentID:      uuid.New(),
			InvestorID:      investorID2,
			DocumentType:    "passport",
			DocumentSubject: stringPtr("primary_signatory"),
			DocumentTitle:   "Passport - Dr. Sarah Elizabeth Johnson",
			FileName:        stringPtr("johnson_passport.pdf"),
			Status:          store.DocumentStatusReceived,
			IssuedDate:      timePtr(time.Date(2019, 8, 12, 0, 0, 0, 0, time.UTC)),
			ExpiryDate:      timePtr(time.Date(2029, 8, 12, 0, 0, 0, 0, time.UTC)),
			CreatedAt:       now.AddDate(0, 0, -3),
			UpdatedAt:       now.AddDate(0, 0, -3),
		},
	}

	return &MockHedgeFundData{
		Funds:            funds,
		ShareClasses:     shareClasses,
		Series:           series,
		Investors:        investors,
		KYCProfiles:      kycProfiles,
		TaxProfiles:      taxProfiles,
		BankInstructions: bankInstructions,
		Trades:           trades,
		RegisterLots:     registerLots,
		Documents:        documents,
	}
}

// GetMockHedgeFundInvestor returns a specific mock investor by code
func GetMockHedgeFundInvestor(investorCode string) (*domain.HedgeFundInvestor, error) {
	data := GetMockHedgeFundData()

	for i := range data.Investors {
		if data.Investors[i].InvestorCode == investorCode {
			return &data.Investors[i], nil
		}
	}

	return nil, fmt.Errorf("mock investor with code %s not found", investorCode)
}

// GetMockHedgeFundFund returns a specific mock fund by name
func GetMockHedgeFundFund(fundName string) (*domain.HedgeFund, error) {
	data := GetMockHedgeFundData()

	for i := range data.Funds {
		if data.Funds[i].FundName == fundName {
			return &data.Funds[i], nil
		}
	}

	return nil, fmt.Errorf("mock fund with name %s not found", fundName)
}

// Helper functions for pointer creation
func stringPtr(s string) *string {
	return &s
}

func floatPtr(f float64) *float64 {
	return &f
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func int64Ptr(i int64) *int64 {
	return &i
}

// SeedMockHedgeFundStore populates a mock store with sample data
func SeedMockHedgeFundStore(store *store.MockHedgeFundInvestorStore) error {
	ctx := context.Background()
	data := GetMockHedgeFundData()

	// Seed funds
	for i := range data.Funds {
		if err := store.CreateFund(ctx, &data.Funds[i]); err != nil {
			return fmt.Errorf("failed to seed fund %s: %w", data.Funds[i].FundName, err)
		}
	}

	// Seed share classes
	for i := range data.ShareClasses {
		if err := store.CreateShareClass(ctx, &data.ShareClasses[i]); err != nil {
			return fmt.Errorf("failed to seed share class %s: %w", data.ShareClasses[i].ClassName, err)
		}
	}

	// Seed investors
	for i := range data.Investors {
		if err := store.CreateInvestor(ctx, &data.Investors[i]); err != nil {
			return fmt.Errorf("failed to seed investor %s: %w", data.Investors[i].InvestorCode, err)
		}
	}

	return nil
}
