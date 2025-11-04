package cli

import (
	"context"
	"flag"
	"fmt"
	"time"

	"dsl-ob-poc/internal/datastore"
	"dsl-ob-poc/internal/hf-investor/domain"
	"dsl-ob-poc/internal/hf-investor/dsl"

	"github.com/google/uuid"
)

// RunHFCreateInvestor handles the 'hf-create-investor' command for hedge fund investor creation
func RunHFCreateInvestor(ctx context.Context, ds datastore.DataStore, args []string) error {
	fs := flag.NewFlagSet("hf-create-investor", flag.ExitOnError)

	// Required flags
	investorCode := fs.String("code", "", "Investor code (e.g., INV-001) (required)")
	legalName := fs.String("legal-name", "", "Legal name of the investor (required)")
	investorType := fs.String("type", "", "Investor type: INDIVIDUAL, CORPORATE, TRUST, FOHF, NOMINEE, PENSION_FUND, INSURANCE_CO (required)")
	domicile := fs.String("domicile", "", "Investor domicile (2-letter country code) (optional)")

	// Optional flags
	shortName := fs.String("short-name", "", "Short name for the investor")
	lei := fs.String("lei", "", "Legal Entity Identifier")
	regNumber := fs.String("reg-number", "", "Registration number")
	source := fs.String("source", "", "Source of the investor lead")

	// Contact information
	contactName := fs.String("contact-name", "", "Primary contact name")
	contactEmail := fs.String("contact-email", "", "Primary contact email")
	contactPhone := fs.String("contact-phone", "", "Primary contact phone")

	// Address information
	addressLine1 := fs.String("address1", "", "Address line 1")
	addressLine2 := fs.String("address2", "", "Address line 2")
	city := fs.String("city", "", "City")
	country := fs.String("country", "", "Country (2-letter code)")
	postalCode := fs.String("postal-code", "", "Postal code")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	// Validate required fields
	if *investorCode == "" || *legalName == "" || *investorType == "" {
		fs.Usage()
		return fmt.Errorf("error: --code, --legal-name, and --type flags are required")
	}

	// Validate investor type
	if !domain.IsValidInvestorType(*investorType) {
		return fmt.Errorf("invalid investor type: %s", *investorType)
	}

	// Create the hedge fund investor entity
	investor := &domain.HedgeFundInvestor{
		InvestorID:   uuid.New(),
		InvestorCode: *investorCode,
		Type:         *investorType,
		LegalName:    *legalName,
		Status:       domain.InvestorStatusOpportunity,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	// Set optional fields
	if *domicile != "" {
		investor.Domicile = *domicile
	}
	if *shortName != "" {
		investor.ShortName = shortName
	}
	if *lei != "" {
		investor.LEI = lei
	}
	if *regNumber != "" {
		investor.RegistrationNumber = regNumber
	}
	if *source != "" {
		investor.Source = source
	}
	if *contactName != "" {
		investor.PrimaryContactName = contactName
	}
	if *contactEmail != "" {
		investor.PrimaryContactEmail = contactEmail
	}
	if *contactPhone != "" {
		investor.PrimaryContactPhone = contactPhone
	}
	if *addressLine1 != "" {
		investor.AddressLine1 = addressLine1
	}
	if *addressLine2 != "" {
		investor.AddressLine2 = addressLine2
	}
	if *city != "" {
		investor.City = city
	}
	if *country != "" {
		investor.Country = country
	}
	if *postalCode != "" {
		investor.PostalCode = postalCode
	}

	// Create DSL operation for investor creation
	operation := &dsl.HedgeFundDSLOperation{
		Verb: "investor.start-opportunity",
		Args: map[string]interface{}{
			"legal-name": *legalName,
			"type":       *investorType,
		},
		Timestamp: time.Now().UTC(),
	}

	if *domicile != "" {
		operation.Args["domicile"] = *domicile
	}
	if *source != "" {
		operation.Args["source"] = *source
	}

	// Validate the DSL operation
	if err := dsl.ValidateHedgeFundDSLOperation(operation); err != nil {
		return fmt.Errorf("invalid DSL operation: %w", err)
	}

	// Generate DSL text
	dslText := dsl.GenerateHedgeFundDSL(operation)

	fmt.Printf("Creating hedge fund investor:\n")
	fmt.Printf("  Investor Code: %s\n", investor.InvestorCode)
	fmt.Printf("  Legal Name: %s\n", investor.LegalName)
	fmt.Printf("  Type: %s\n", investor.Type)
	if investor.Domicile != "" {
		fmt.Printf("  Domicile: %s\n", investor.Domicile)
	}
	fmt.Printf("  Status: %s\n", investor.Status)
	fmt.Printf("\nGenerated DSL:\n%s\n", dslText)

	// TODO: Store investor in database when HF investor store is implemented
	// For now, just show what would be created
	fmt.Printf("\nInvestor ID: %s\n", investor.InvestorID.String())
	fmt.Printf("Created successfully at: %s\n", investor.CreatedAt.Format(time.RFC3339))

	return nil
}

// RunHFRecordIndication handles the 'hf-record-indication' command for recording investment indications
func RunHFRecordIndication(ctx context.Context, ds datastore.DataStore, args []string) error {
	fs := flag.NewFlagSet("hf-record-indication", flag.ExitOnError)

	investorID := fs.String("investor", "", "Investor ID (UUID) (required)")
	fundID := fs.String("fund", "", "Fund ID (UUID) (required)")
	classID := fs.String("class", "", "Share class ID (UUID) (required)")
	ticket := fs.String("ticket", "", "Indicated investment amount (required)")
	currency := fs.String("currency", "", "Investment currency (required)")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	// Validate required fields
	if *investorID == "" || *fundID == "" || *classID == "" || *ticket == "" || *currency == "" {
		fs.Usage()
		return fmt.Errorf("error: --investor, --fund, --class, --ticket, and --currency flags are required")
	}

	// Validate UUIDs
	investorUUID, err := uuid.Parse(*investorID)
	if err != nil {
		return fmt.Errorf("invalid investor ID format: %s", *investorID)
	}

	fundUUID, err := uuid.Parse(*fundID)
	if err != nil {
		return fmt.Errorf("invalid fund ID format: %s", *fundID)
	}

	classUUID, err := uuid.Parse(*classID)
	if err != nil {
		return fmt.Errorf("invalid class ID format: %s", *classID)
	}

	// Create DSL operation for recording indication
	operation := &dsl.HedgeFundDSLOperation{
		Verb: "investor.record-indication",
		Args: map[string]interface{}{
			"investor": investorUUID.String(),
			"fund":     fundUUID.String(),
			"class":    classUUID.String(),
			"ticket":   *ticket,
			"currency": *currency,
		},
		Timestamp: time.Now().UTC(),
	}

	// Validate the DSL operation
	if err := dsl.ValidateHedgeFundDSLOperation(operation); err != nil {
		return fmt.Errorf("invalid DSL operation: %w", err)
	}

	// Generate DSL text
	dslText := dsl.GenerateHedgeFundDSL(operation)

	fmt.Printf("Recording investment indication:\n")
	fmt.Printf("  Investor: %s\n", *investorID)
	fmt.Printf("  Fund: %s\n", *fundID)
	fmt.Printf("  Class: %s\n", *classID)
	fmt.Printf("  Ticket: %s %s\n", *ticket, *currency)
	fmt.Printf("\nGenerated DSL:\n%s\n", dslText)

	// TODO: Execute state transition when HF investor store is implemented
	fmt.Printf("\nIndication recorded successfully\n")

	return nil
}

// RunHFBeginKYC handles the 'hf-begin-kyc' command for starting KYC process
func RunHFBeginKYC(ctx context.Context, ds datastore.DataStore, args []string) error {
	fs := flag.NewFlagSet("hf-begin-kyc", flag.ExitOnError)

	investorID := fs.String("investor", "", "Investor ID (UUID) (required)")
	tier := fs.String("tier", "STANDARD", "KYC tier: SIMPLIFIED, STANDARD, ENHANCED")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	// Validate required fields
	if *investorID == "" {
		fs.Usage()
		return fmt.Errorf("error: --investor flag is required")
	}

	// Validate UUID
	investorUUID, err := uuid.Parse(*investorID)
	if err != nil {
		return fmt.Errorf("invalid investor ID format: %s", *investorID)
	}

	// Create DSL operation for beginning KYC
	operation := &dsl.HedgeFundDSLOperation{
		Verb: "kyc.begin",
		Args: map[string]interface{}{
			"investor": investorUUID.String(),
			"tier":     *tier,
		},
		Timestamp: time.Now().UTC(),
	}

	// Validate the DSL operation
	if err := dsl.ValidateHedgeFundDSLOperation(operation); err != nil {
		return fmt.Errorf("invalid DSL operation: %w", err)
	}

	// Generate DSL text
	dslText := dsl.GenerateHedgeFundDSL(operation)

	fmt.Printf("Beginning KYC process:\n")
	fmt.Printf("  Investor: %s\n", *investorID)
	fmt.Printf("  KYC Tier: %s\n", *tier)
	fmt.Printf("\nGenerated DSL:\n%s\n", dslText)

	// TODO: Execute state transition when HF investor store is implemented
	fmt.Printf("\nKYC process initiated successfully\n")

	return nil
}

// RunHFApproveKYC handles the 'hf-approve-kyc' command for approving KYC
func RunHFApproveKYC(ctx context.Context, ds datastore.DataStore, args []string) error {
	fs := flag.NewFlagSet("hf-approve-kyc", flag.ExitOnError)

	investorID := fs.String("investor", "", "Investor ID (UUID) (required)")
	risk := fs.String("risk", "", "Risk rating: LOW, MEDIUM, HIGH (required)")
	refreshDue := fs.String("refresh-due", "", "Next refresh due date (YYYY-MM-DD) (required)")
	approvedBy := fs.String("approved-by", "", "Approver name (required)")
	comments := fs.String("comments", "", "Approval comments")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	// Validate required fields
	if *investorID == "" || *risk == "" || *refreshDue == "" || *approvedBy == "" {
		fs.Usage()
		return fmt.Errorf("error: --investor, --risk, --refresh-due, and --approved-by flags are required")
	}

	// Validate UUID
	investorUUID, err := uuid.Parse(*investorID)
	if err != nil {
		return fmt.Errorf("invalid investor ID format: %s", *investorID)
	}

	// Validate date format
	if _, err := time.Parse("2006-01-02", *refreshDue); err != nil {
		return fmt.Errorf("invalid refresh due date format (expected YYYY-MM-DD): %s", *refreshDue)
	}

	// Create DSL operation for approving KYC
	operation := &dsl.HedgeFundDSLOperation{
		Verb: "kyc.approve",
		Args: map[string]interface{}{
			"investor":    investorUUID.String(),
			"risk":        *risk,
			"refresh-due": *refreshDue,
			"approved-by": *approvedBy,
		},
		Timestamp: time.Now().UTC(),
	}

	if *comments != "" {
		operation.Args["comments"] = *comments
	}

	// Validate the DSL operation
	if err := dsl.ValidateHedgeFundDSLOperation(operation); err != nil {
		return fmt.Errorf("invalid DSL operation: %w", err)
	}

	// Generate DSL text
	dslText := dsl.GenerateHedgeFundDSL(operation)

	fmt.Printf("Approving KYC:\n")
	fmt.Printf("  Investor: %s\n", *investorID)
	fmt.Printf("  Risk Rating: %s\n", *risk)
	fmt.Printf("  Refresh Due: %s\n", *refreshDue)
	fmt.Printf("  Approved By: %s\n", *approvedBy)
	if *comments != "" {
		fmt.Printf("  Comments: %s\n", *comments)
	}
	fmt.Printf("\nGenerated DSL:\n%s\n", dslText)

	// TODO: Execute state transition when HF investor store is implemented
	fmt.Printf("\nKYC approved successfully\n")

	return nil
}

// RunHFSubscribeRequest handles the 'hf-subscribe-request' command for subscription requests
func RunHFSubscribeRequest(ctx context.Context, ds datastore.DataStore, args []string) error {
	fs := flag.NewFlagSet("hf-subscribe-request", flag.ExitOnError)

	investorID := fs.String("investor", "", "Investor ID (UUID) (required)")
	fundID := fs.String("fund", "", "Fund ID (UUID) (required)")
	classID := fs.String("class", "", "Share class ID (UUID) (required)")
	amount := fs.String("amount", "", "Subscription amount (required)")
	currency := fs.String("currency", "", "Settlement currency (required)")
	tradeDate := fs.String("trade-date", "", "Trade date (YYYY-MM-DD) (required)")
	valueDate := fs.String("value-date", "", "Value date (YYYY-MM-DD) (required)")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	// Validate required fields
	if *investorID == "" || *fundID == "" || *classID == "" || *amount == "" ||
		*currency == "" || *tradeDate == "" || *valueDate == "" {
		fs.Usage()
		return fmt.Errorf("error: all flags are required")
	}

	// Validate UUIDs
	investorUUID, err := uuid.Parse(*investorID)
	if err != nil {
		return fmt.Errorf("invalid investor ID format: %s", *investorID)
	}

	fundUUID, err := uuid.Parse(*fundID)
	if err != nil {
		return fmt.Errorf("invalid fund ID format: %s", *fundID)
	}

	classUUID, err := uuid.Parse(*classID)
	if err != nil {
		return fmt.Errorf("invalid class ID format: %s", *classID)
	}

	// Validate dates
	if _, err := time.Parse("2006-01-02", *tradeDate); err != nil {
		return fmt.Errorf("invalid trade date format (expected YYYY-MM-DD): %s", *tradeDate)
	}

	if _, err := time.Parse("2006-01-02", *valueDate); err != nil {
		return fmt.Errorf("invalid value date format (expected YYYY-MM-DD): %s", *valueDate)
	}

	// Create DSL operation for subscription request
	operation := &dsl.HedgeFundDSLOperation{
		Verb: "subscribe.request",
		Args: map[string]interface{}{
			"investor":   investorUUID.String(),
			"fund":       fundUUID.String(),
			"class":      classUUID.String(),
			"amount":     *amount,
			"currency":   *currency,
			"trade-date": *tradeDate,
			"value-date": *valueDate,
		},
		Timestamp: time.Now().UTC(),
	}

	// Validate the DSL operation
	if err := dsl.ValidateHedgeFundDSLOperation(operation); err != nil {
		return fmt.Errorf("invalid DSL operation: %w", err)
	}

	// Generate DSL text
	dslText := dsl.GenerateHedgeFundDSL(operation)

	fmt.Printf("Creating subscription request:\n")
	fmt.Printf("  Investor: %s\n", *investorID)
	fmt.Printf("  Fund: %s\n", *fundID)
	fmt.Printf("  Class: %s\n", *classID)
	fmt.Printf("  Amount: %s %s\n", *amount, *currency)
	fmt.Printf("  Trade Date: %s\n", *tradeDate)
	fmt.Printf("  Value Date: %s\n", *valueDate)
	fmt.Printf("\nGenerated DSL:\n%s\n", dslText)

	// TODO: Execute state transition when HF investor store is implemented
	fmt.Printf("\nSubscription request created successfully\n")

	return nil
}

// RunHFAmendInvestorDetails handles the 'hf-amend-investor' command for updating investor details
func RunHFAmendInvestorDetails(ctx context.Context, ds datastore.DataStore, args []string) error {
	fs := flag.NewFlagSet("hf-amend-investor", flag.ExitOnError)

	// Require
