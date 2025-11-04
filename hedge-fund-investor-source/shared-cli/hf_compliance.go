package cli

import (
	"context"
	"flag"
	"fmt"
	"time"

	"dsl-ob-poc/internal/datastore"
	"dsl-ob-poc/internal/hf-investor/dsl"

	"github.com/google/uuid"
)

// RunHFCaptureTax handles the 'hf-capture-tax' command for capturing tax information
func RunHFCaptureTax(ctx context.Context, ds datastore.DataStore, args []string) error {
	fs := flag.NewFlagSet("hf-capture-tax", flag.ExitOnError)

	investorID := fs.String("investor", "", "Investor ID (UUID) (required)")
	fatca := fs.String("fatca", "", "FATCA status: US_PERSON, NON_US_PERSON, SPECIFIED_US_PERSON")
	crs := fs.String("crs", "", "CRS classification: INDIVIDUAL, ENTITY, FINANCIAL_INSTITUTION")
	form := fs.String("form", "", "Tax form type: W9, W8_BEN, W8_BEN_E, ENTITY_SELF_CERT")
	tinType := fs.String("tin-type", "", "TIN type: SSN, EIN, FOREIGN_TIN")
	tinValue := fs.String("tin-value", "", "TIN value")

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

	// Create DSL operation for tax capture
	operation := &dsl.HedgeFundDSLOperation{
		Verb: "tax.capture",
		Args: map[string]interface{}{
			"investor": investorUUID.String(),
		},
		Timestamp: time.Now().UTC(),
	}

	// Add optional tax fields
	if *fatca != "" {
		operation.Args["fatca"] = *fatca
	}
	if *crs != "" {
		operation.Args["crs"] = *crs
	}
	if *form != "" {
		operation.Args["form"] = *form
	}
	if *tinType != "" {
		operation.Args["tin-type"] = *tinType
	}
	if *tinValue != "" {
		operation.Args["tin-value"] = *tinValue
	}

	// Validate the DSL operation
	if err := dsl.ValidateHedgeFundDSLOperation(operation); err != nil {
		return fmt.Errorf("invalid DSL operation: %w", err)
	}

	// Generate DSL text
	dslText := dsl.GenerateHedgeFundDSL(operation)

	fmt.Printf("Capturing tax information:\n")
	fmt.Printf("  Investor: %s\n", *investorID)
	if *fatca != "" {
		fmt.Printf("  FATCA Status: %s\n", *fatca)
	}
	if *crs != "" {
		fmt.Printf("  CRS Classification: %s\n", *crs)
	}
	if *form != "" {
		fmt.Printf("  Tax Form: %s\n", *form)
	}
	if *tinType != "" && *tinValue != "" {
		fmt.Printf("  TIN: %s (%s)\n", *tinValue, *tinType)
	}
	fmt.Printf("\nGenerated DSL:\n%s\n", dslText)

	// TODO: Store tax information when HF investor store is implemented
	fmt.Printf("\nTax information captured successfully\n")

	return nil
}

// RunHFSetBankInstruction handles the 'hf-set-bank-instruction' command for setting banking instructions
func RunHFSetBankInstruction(ctx context.Context, ds datastore.DataStore, args []string) error {
	fs := flag.NewFlagSet("hf-set-bank-instruction", flag.ExitOnError)

	investorID := fs.String("investor", "", "Investor ID (UUID) (required)")
	currency := fs.String("currency", "", "Settlement currency (required)")
	bankName := fs.String("bank-name", "", "Bank name (required)")
	accountName := fs.String("account-name", "", "Account name (required)")
	iban := fs.String("iban", "", "IBAN")
	swift := fs.String("swift", "", "SWIFT BIC")
	accountNum := fs.String("account-num", "", "Account number")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	// Validate required fields
	if *investorID == "" || *currency == "" || *bankName == "" || *accountName == "" {
		fs.Usage()
		return fmt.Errorf("error: --investor, --currency, --bank-name, and --account-name flags are required")
	}

	// Validate UUID
	investorUUID, err := uuid.Parse(*investorID)
	if err != nil {
		return fmt.Errorf("invalid investor ID format: %s", *investorID)
	}

	// Create DSL operation for banking instruction
	operation := &dsl.HedgeFundDSLOperation{
		Verb: "bank.set-instruction",
		Args: map[string]interface{}{
			"investor":     investorUUID.String(),
			"currency":     *currency,
			"bank-name":    *bankName,
			"account-name": *accountName,
		},
		Timestamp: time.Now().UTC(),
	}

	// Add optional banking fields
	if *iban != "" {
		operation.Args["iban"] = *iban
	}
	if *swift != "" {
		operation.Args["swift"] = *swift
	}
	if *accountNum != "" {
		operation.Args["account-num"] = *accountNum
	}

	// Validate the DSL operation
	if err := dsl.ValidateHedgeFundDSLOperation(operation); err != nil {
		return fmt.Errorf("invalid DSL operation: %w", err)
	}

	// Generate DSL text
	dslText := dsl.GenerateHedgeFundDSL(operation)

	fmt.Printf("Setting banking instruction:\n")
	fmt.Printf("  Investor: %s\n", *investorID)
	fmt.Printf("  Currency: %s\n", *currency)
	fmt.Printf("  Bank Name: %s\n", *bankName)
	fmt.Printf("  Account Name: %s\n", *accountName)
	if *iban != "" {
		fmt.Printf("  IBAN: %s\n", *iban)
	}
	if *swift != "" {
		fmt.Printf("  SWIFT: %s\n", *swift)
	}
	if *accountNum != "" {
		fmt.Printf("  Account Number: %s\n", *accountNum)
	}
	fmt.Printf("\nGenerated DSL:\n%s\n", dslText)

	// TODO: Store banking instruction when HF investor store is implemented
	fmt.Printf("\nBanking instruction set successfully\n")

	return nil
}

// RunHFCollectDocument handles the 'hf-collect-document' command for collecting KYC documents
func RunHFCollectDocument(ctx context.Context, ds datastore.DataStore, args []string) error {
	fs := flag.NewFlagSet("hf-collect-document", flag.ExitOnError)

	investorID := fs.String("investor", "", "Investor ID (UUID) (required)")
	docType := fs.String("doc-type", "", "Document type (required)")
	subject := fs.String("subject", "", "Document subject (e.g., primary_signatory)")
	filePath := fs.String("file-path", "", "Path to uploaded document")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	// Validate required fields
	if *investorID == "" || *docType == "" {
		fs.Usage()
		return fmt.Errorf("error: --investor and --doc-type flags are required")
	}

	// Validate UUID
	investorUUID, err := uuid.Parse(*investorID)
	if err != nil {
		return fmt.Errorf("invalid investor ID format: %s", *investorID)
	}

	// Create DSL operation for document collection
	operation := &dsl.HedgeFundDSLOperation{
		Verb: "kyc.collect-doc",
		Args: map[string]interface{}{
			"investor": investorUUID.String(),
			"doc-type": *docType,
		},
		Timestamp: time.Now().UTC(),
	}

	// Add optional fields
	if *subject != "" {
		operation.Args["subject"] = *subject
	}
	if *filePath != "" {
		operation.Args["file-path"] = *filePath
	}

	// Validate the DSL operation
	if err := dsl.ValidateHedgeFundDSLOperation(operation); err != nil {
		return fmt.Errorf("invalid DSL operation: %w", err)
	}

	// Generate DSL text
	dslText := dsl.GenerateHedgeFundDSL(operation)

	fmt.Printf("Collecting KYC document:\n")
	fmt.Printf("  Investor: %s\n", *investorID)
	fmt.Printf("  Document Type: %s\n", *docType)
	if *subject != "" {
		fmt.Printf("  Subject: %s\n", *subject)
	}
	if *filePath != "" {
		fmt.Printf("  File Path: %s\n", *filePath)
	}
	fmt.Printf("\nGenerated DSL:\n%s\n", dslText)

	// TODO: Store document when HF investor store is implemented
	fmt.Printf("\nDocument collected successfully\n")

	return nil
}

// RunHFScreenInvestor handles the 'hf-screen-investor' command for screening investors
func RunHFScreenInvestor(ctx context.Context, ds datastore.DataStore, args []string) error {
	fs := flag.NewFlagSet("hf-screen-investor", flag.ExitOnError)

	investorID := fs.String("investor", "", "Investor ID (UUID) (required)")
	provider := fs.String("provider", "", "Screening provider: worldcheck, refinitiv, accelus (required)")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	// Validate required fields
	if *investorID == "" || *provider == "" {
		fs.Usage()
		return fmt.Errorf("error: --investor and --provider flags are required")
	}

	// Validate UUID
	investorUUID, err := uuid.Parse(*investorID)
	if err != nil {
		return fmt.Errorf("invalid investor ID format: %s", *investorID)
	}

	// Create DSL operation for screening
	operation := &dsl.HedgeFundDSLOperation{
		Verb: "kyc.screen",
		Args: map[string]interface{}{
			"investor": investorUUID.String(),
			"provider": *provider,
		},
		Timestamp: time.Now().UTC(),
	}

	// Validate the DSL operation
	if err := dsl.ValidateHedgeFundDSLOperation(operation); err != nil {
		return fmt.Errorf("invalid DSL operation: %w", err)
	}

	// Generate DSL text
	dslText := dsl.GenerateHedgeFundDSL(operation)

	fmt.Printf("Screening investor:\n")
	fmt.Printf("  Investor: %s\n", *investorID)
	fmt.Printf("  Provider: %s\n", *provider)
	fmt.Printf("\nGenerated DSL:\n%s\n", dslText)

	// TODO: Execute screening when HF investor store is implemented
	fmt.Printf("\nInvestor screening initiated successfully\n")

	return nil
}

// RunHFSetRefreshSchedule handles the 'hf-set-refresh-schedule' command for setting KYC refresh schedules
func RunHFSetRefreshSchedule(ctx context.Context, ds datastore.DataStore, args []string) error {
	fs := flag.NewFlagSet("hf-set-refresh-schedule", flag.ExitOnError)

	investorID := fs.String("investor", "", "Investor ID (UUID) (required)")
	frequency := fs.String("frequency", "", "Refresh frequency: MONTHLY, QUARTERLY, ANNUAL (required)")
	next := fs.String("next", "", "Next refresh date (YYYY-MM-DD) (required)")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	// Validate required fields
	if *investorID == "" || *frequency == "" || *next == "" {
		fs.Usage()
		return fmt.Errorf("error: --investor, --frequency, and --next flags are required")
	}

	// Validate UUID
	investorUUID, err := uuid.Parse(*investorID)
	if err != nil {
		return fmt.Errorf("invalid investor ID format: %s", *investorID)
	}

	// Validate date
	if _, err := time.Parse("2006-01-02", *next); err != nil {
		return fmt.Errorf("invalid next refresh date format (expected YYYY-MM-DD): %s", *next)
	}

	// Create DSL operation for refresh schedule
	operation := &dsl.HedgeFundDSLOperation{
		Verb: "kyc.refresh-schedule",
		Args: map[string]interface{}{
			"investor":  investorUUID.String(),
			"frequency": *frequency,
			"next":      *next,
		},
		Timestamp: time.Now().UTC(),
	}

	// Validate the DSL operation
	if err := dsl.ValidateHedgeFundDSLOperation(operation); err != nil {
		return fmt.Errorf("invalid DSL operation: %w", err)
	}

	// Generate DSL text
	dslText := dsl.GenerateHedgeFundDSL(operation)

	fmt.Printf("Setting KYC refresh schedule:\n")
	fmt.Printf("  Investor: %s\n", *investorID)
	fmt.Printf("  Frequency: %s\n", *frequency)
	fmt.Printf("  Next Refresh: %s\n", *next)
	fmt.Printf("\nGenerated DSL:\n%s\n", dslText)

	// TODO: Set refresh schedule when HF investor store is implemented
	fmt.Printf("\nKYC refresh schedule set successfully\n")

	return nil
}

// RunHFSetContinuousScreening handles the 'hf-set-continuous-screening' command for continuous screening
func RunHFSetContinuousScreening(ctx context.Context, ds datastore.DataStore, args []string) error {
	fs := flag.NewFlagSet("hf-set-continuous-screening", flag.ExitOnError)

	investorID := fs.String("investor", "", "Investor ID (UUID) (required)")
	frequency := fs.String("frequency", "", "Screening frequency: DAILY, WEEKLY, MONTHLY (required)")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	// Validate required fields
	if *investorID == "" || *frequency == "" {
		fs.Usage()
		return fmt.Errorf("error: --investor and --frequency flags are required")
	}

	// Validate UUID
	investorUUID, err := uuid.Parse(*investorID)
	if err != nil {
		return fmt.Errorf("invalid investor ID format: %s", *investorID)
	}

	// Create DSL operation for continuous screening
	operation := &dsl.HedgeFundDSLOperation{
		Verb: "screen.continuous",
		Args: map[string]interface{}{
			"investor":  investorUUID.String(),
			"frequency": *frequency,
		},
		Timestamp: time.Now().UTC(),
	}

	// Validate the DSL operation
	if err := dsl.ValidateHedgeFundDSLOperation(operation); err != nil {
		return fmt.Errorf("invalid DSL operation: %w", err)
	}

	// Generate DSL text
	dslText := dsl.GenerateHedgeFundDSL(operation)

	fmt.Printf("Setting continuous screening:\n")
	fmt.Printf("  Investor: %s\n", *investorID)
	fmt.Printf("  Frequency: %s\n", *frequency)
	fmt.Printf("\nGenerated DSL:\n%s\n", dslText)

	// TODO: Set continuous screening when HF investor store is implemented
	fmt.Printf("\nContinuous screening set successfully\n")

	return nil
}

// RunHFShowRegister handles the 'hf-show-register' command for displaying register of investors
func RunHFShowRegister(ctx context.Context, ds datastore.DataStore, args []string) error {
	fs := flag.NewFlagSet("hf-show-register", flag.ExitOnError)

	fundID := fs.String("fund", "", "Fund ID (UUID) - show specific fund only")
	classID := fs.String("class", "", "Class ID (UUID) - show specific class only")
	status := fs.String("status", "", "Investor status filter")
	format := fs.String("format", "table", "Output format: table, json, csv")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	fmt.Printf("Displaying Register of Investors:\n")
	fmt.Printf("================================\n")

	// Apply filters if provided
	if *fundID != "" {
		fmt.Printf("Fund Filter: %s\n", *fundID)
	}
	if *classID != "" {
		fmt.Printf("Class Filter: %s\n", *classID)
	}
	if *status != "" {
		fmt.Printf("Status Filter: %s\n", *status)
	}
	fmt.Printf("Output Format: %s\n\n", *format)

	// TODO: Query and display register when HF investor store is implemented
	fmt.Printf("Mock Register Data:\n")
	fmt.Printf("%-10s %-30s %-15s %-15s %-15s %-20s\n",
		"Code", "Investor Name", "Type", "Status", "Units", "Last Activity")
	fmt.Printf("%-10s %-30s %-15s %-15s %-15s %-20s\n",
		"----------", "------------------------------", "---------------", "---------------", "---------------", "--------------------")
	fmt.Printf("%-10s %-30s %-15s %-15s %-15.2f %-20s\n",
		"INV-001", "Sample Institutional Investor", "CORPORATE", "ACTIVE", 10000.50, "2024-01-15")
	fmt.Printf("%-10s %-30s %-15s %-15s %-15.2f %-20s\n",
		"INV-002", "John Smith", "INDIVIDUAL", "KYC_PENDING", 0.00, "2024-01-10")

	fmt.Printf("\nRegister displayed successfully\n")
	return nil
}

// RunHFShowKYCDashboard handles the 'hf-show-kyc-dashboard' command for displaying KYC status
func RunHFShowKYCDashboard(ctx context.Context, ds datastore.DataStore, args []string) error {
	fs := flag.NewFlagSet("hf-show-kyc-dashboard", flag.ExitOnError)

	risk := fs.String("risk", "", "Risk rating filter: LOW, MEDIUM, HIGH")
	status := fs.String("status", "", "KYC status filter: PENDING, APPROVED, EXPIRED")
	overdue := fs.Bool("overdue", false, "Show only overdue refreshes")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	fmt.Printf("KYC Dashboard:\n")
	fmt.Printf("==============\n")

	// Apply filters if provided
	if *risk != "" {
		fmt.Printf("Risk Filter: %s\n", *risk)
	}
	if *status != "" {
		fmt.Printf("Status Filter: %s\n", *status)
	}
	if *overdue {
		fmt.Printf("Showing only overdue refreshes\n")
	}
	fmt.Printf("\n")

	// TODO: Query and display KYC dashboard when HF investor store is implemented
	fmt.Printf("Mock KYC Dashboard Data:\n")
	fmt.Printf("%-10s %-30s %-10s %-15s %-15s %-20s\n",
		"Code", "Investor Name", "Risk", "KYC Status", "Refresh Due", "Next Action")
	fmt.Printf("%-10s %-30s %-10s %-15s %-15s %-20s\n",
		"----------", "------------------------------", "----------", "---------------", "---------------", "--------------------")
	fmt.Printf("%-10s %-30s %-10s %-15s %-15s %-20s\n",
		"INV-001", "Sample Institutional Investor", "LOW", "APPROVED", "2024-12-01", "NO_ACTION")
	fmt.Printf("%-10s %-30s %-10s %-15s %-15s %-20s\n",
		"INV-002", "John Smith", "MEDIUM", "PENDING", "N/A", "INITIAL_REVIEW")
	fmt.Printf("%-10s %-30s %-10s %-15s %-15s %-20s\n",
		"INV-003", "ABC Corporation", "HIGH", "APPROVED", "2024-01-01", "REFRESH_OVERDUE")

	fmt.Printf("\nKYC dashboard displayed successfully\n")
	return nil
}
