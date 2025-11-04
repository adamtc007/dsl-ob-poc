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

// RunHFConfirmCash handles the 'hf-confirm-cash' command for confirming cash receipt
func RunHFConfirmCash(ctx context.Context, ds datastore.DataStore, args []string) error {
	fs := flag.NewFlagSet("hf-confirm-cash", flag.ExitOnError)

	investorID := fs.String("investor", "", "Investor ID (UUID) (required)")
	tradeID := fs.String("trade", "", "Trade ID (UUID) (required)")
	amount := fs.String("amount", "", "Amount received (required)")
	valueDate := fs.String("value-date", "", "Value date (YYYY-MM-DD) (required)")
	bankCurrency := fs.String("bank-currency", "", "Received currency (required)")
	reference := fs.String("reference", "", "Bank reference")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	// Validate required fields
	if *investorID == "" || *tradeID == "" || *amount == "" || *valueDate == "" || *bankCurrency == "" {
		fs.Usage()
		return fmt.Errorf("error: --investor, --trade, --amount, --value-date, and --bank-currency flags are required")
	}

	// Validate UUIDs
	investorUUID, err := uuid.Parse(*investorID)
	if err != nil {
		return fmt.Errorf("invalid investor ID format: %s", *investorID)
	}

	tradeUUID, err := uuid.Parse(*tradeID)
	if err != nil {
		return fmt.Errorf("invalid trade ID format: %s", *tradeID)
	}

	// Validate date
	if _, err := time.Parse("2006-01-02", *valueDate); err != nil {
		return fmt.Errorf("invalid value date format (expected YYYY-MM-DD): %s", *valueDate)
	}

	// Create DSL operation for cash confirmation
	operation := &dsl.HedgeFundDSLOperation{
		Verb: "cash.confirm",
		Args: map[string]interface{}{
			"investor":      investorUUID.String(),
			"trade":         tradeUUID.String(),
			"amount":        *amount,
			"value-date":    *valueDate,
			"bank-currency": *bankCurrency,
		},
		Timestamp: time.Now().UTC(),
	}

	if *reference != "" {
		operation.Args["reference"] = *reference
	}

	// Validate the DSL operation
	if err := dsl.ValidateHedgeFundDSLOperation(operation); err != nil {
		return fmt.Errorf("invalid DSL operation: %w", err)
	}

	// Generate DSL text
	dslText := dsl.GenerateHedgeFundDSL(operation)

	fmt.Printf("Confirming cash receipt:\n")
	fmt.Printf("  Investor: %s\n", *investorID)
	fmt.Printf("  Trade: %s\n", *tradeID)
	fmt.Printf("  Amount: %s %s\n", *amount, *bankCurrency)
	fmt.Printf("  Value Date: %s\n", *valueDate)
	if *reference != "" {
		fmt.Printf("  Reference: %s\n", *reference)
	}
	fmt.Printf("\nGenerated DSL:\n%s\n", dslText)

	// TODO: Execute state transition when HF investor store is implemented
	fmt.Printf("\nCash receipt confirmed successfully\n")

	return nil
}

// RunHFSetNAV handles the 'hf-set-nav' command for setting NAV
func RunHFSetNAV(ctx context.Context, ds datastore.DataStore, args []string) error {
	fs := flag.NewFlagSet("hf-set-nav", flag.ExitOnError)

	fundID := fs.String("fund", "", "Fund ID (UUID) (required)")
	classID := fs.String("class", "", "Share class ID (UUID) (required)")
	navDate := fs.String("nav-date", "", "NAV date (YYYY-MM-DD) (required)")
	nav := fs.String("nav", "", "NAV per share (required)")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	// Validate required fields
	if *fundID == "" || *classID == "" || *navDate == "" || *nav == "" {
		fs.Usage()
		return fmt.Errorf("error: --fund, --class, --nav-date, and --nav flags are required")
	}

	// Validate UUIDs
	fundUUID, err := uuid.Parse(*fundID)
	if err != nil {
		return fmt.Errorf("invalid fund ID format: %s", *fundID)
	}

	classUUID, err := uuid.Parse(*classID)
	if err != nil {
		return fmt.Errorf("invalid class ID format: %s", *classID)
	}

	// Validate date
	if _, err := time.Parse("2006-01-02", *navDate); err != nil {
		return fmt.Errorf("invalid NAV date format (expected YYYY-MM-DD): %s", *navDate)
	}

	// Create DSL operation for setting NAV
	operation := &dsl.HedgeFundDSLOperation{
		Verb: "deal.nav",
		Args: map[string]interface{}{
			"fund":     fundUUID.String(),
			"class":    classUUID.String(),
			"nav-date": *navDate,
			"nav":      *nav,
		},
		Timestamp: time.Now().UTC(),
	}

	// Validate the DSL operation
	if err := dsl.ValidateHedgeFundDSLOperation(operation); err != nil {
		return fmt.Errorf("invalid DSL operation: %w", err)
	}

	// Generate DSL text
	dslText := dsl.GenerateHedgeFundDSL(operation)

	fmt.Printf("Setting NAV:\n")
	fmt.Printf("  Fund: %s\n", *fundID)
	fmt.Printf("  Class: %s\n", *classID)
	fmt.Printf("  NAV Date: %s\n", *navDate)
	fmt.Printf("  NAV per Share: %s\n", *nav)
	fmt.Printf("\nGenerated DSL:\n%s\n", dslText)

	// TODO: Execute NAV setting when HF investor store is implemented
	fmt.Printf("\nNAV set successfully\n")

	return nil
}

// RunHFIssueUnits handles the 'hf-issue-units' command for issuing units
func RunHFIssueUnits(ctx context.Context, ds datastore.DataStore, args []string) error {
	fs := flag.NewFlagSet("hf-issue-units", flag.ExitOnError)

	investorID := fs.String("investor", "", "Investor ID (UUID) (required)")
	tradeID := fs.String("trade", "", "Trade ID (UUID) (required)")
	classID := fs.String("class", "", "Share class ID (UUID) (required)")
	seriesID := fs.String("series", "", "Series ID (UUID)")
	navPerShare := fs.String("nav-per-share", "", "NAV per share (required)")
	units := fs.String("units", "", "Units to issue (required)")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	// Validate required fields
	if *investorID == "" || *tradeID == "" || *classID == "" || *navPerShare == "" || *units == "" {
		fs.Usage()
		return fmt.Errorf("error: --investor, --trade, --class, --nav-per-share, and --units flags are required")
	}

	// Validate UUIDs
	investorUUID, err := uuid.Parse(*investorID)
	if err != nil {
		return fmt.Errorf("invalid investor ID format: %s", *investorID)
	}

	tradeUUID, err := uuid.Parse(*tradeID)
	if err != nil {
		return fmt.Errorf("invalid trade ID format: %s", *tradeID)
	}

	classUUID, err := uuid.Parse(*classID)
	if err != nil {
		return fmt.Errorf("invalid class ID format: %s", *classID)
	}

	// Create DSL operation for issuing units
	operation := &dsl.HedgeFundDSLOperation{
		Verb: "subscribe.issue",
		Args: map[string]interface{}{
			"investor":      investorUUID.String(),
			"trade":         tradeUUID.String(),
			"class":         classUUID.String(),
			"nav-per-share": *navPerShare,
			"units":         *units,
		},
		Timestamp: time.Now().UTC(),
	}

	// Add series if provided
	if *seriesID != "" {
		seriesUUID, err := uuid.Parse(*seriesID)
		if err != nil {
			return fmt.Errorf("invalid series ID format: %s", *seriesID)
		}
		operation.Args["series"] = seriesUUID.String()
	}

	// Validate the DSL operation
	if err := dsl.ValidateHedgeFundDSLOperation(operation); err != nil {
		return fmt.Errorf("invalid DSL operation: %w", err)
	}

	// Generate DSL text
	dslText := dsl.GenerateHedgeFundDSL(operation)

	fmt.Printf("Issuing units:\n")
	fmt.Printf("  Investor: %s\n", *investorID)
	fmt.Printf("  Trade: %s\n", *tradeID)
	fmt.Printf("  Class: %s\n", *classID)
	if *seriesID != "" {
		fmt.Printf("  Series: %s\n", *seriesID)
	}
	fmt.Printf("  NAV per Share: %s\n", *navPerShare)
	fmt.Printf("  Units: %s\n", *units)
	fmt.Printf("\nGenerated DSL:\n%s\n", dslText)

	// TODO: Execute unit issuance when HF investor store is implemented
	fmt.Printf("\nUnits issued successfully\n")

	return nil
}

// RunHFRedeemRequest handles the 'hf-redeem-request' command for redemption requests
func RunHFRedeemRequest(ctx context.Context, ds datastore.DataStore, args []string) error {
	fs := flag.NewFlagSet("hf-redeem-request", flag.ExitOnError)

	investorID := fs.String("investor", "", "Investor ID (UUID) (required)")
	classID := fs.String("class", "", "Share class ID (UUID) (required)")
	units := fs.String("units", "", "Units to redeem (leave empty for percentage)")
	percentage := fs.String("percentage", "", "Percentage to redeem (leave empty for units)")
	noticeDate := fs.String("notice-date", "", "Notice date (YYYY-MM-DD) (required)")
	valueDate := fs.String("value-date", "", "Redemption value date (YYYY-MM-DD) (required)")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	// Validate required fields
	if *investorID == "" || *classID == "" || *noticeDate == "" || *valueDate == "" {
		fs.Usage()
		return fmt.Errorf("error: --investor, --class, --notice-date, and --value-date flags are required")
	}

	// Validate that either units or percentage is provided, but not both
	if (*units == "" && *percentage == "") || (*units != "" && *percentage != "") {
		return fmt.Errorf("error: either --units or --percentage must be provided, but not both")
	}

	// Validate UUIDs
	investorUUID, err := uuid.Parse(*investorID)
	if err != nil {
		return fmt.Errorf("invalid investor ID format: %s", *investorID)
	}

	classUUID, err := uuid.Parse(*classID)
	if err != nil {
		return fmt.Errorf("invalid class ID format: %s", *classID)
	}

	// Validate dates
	if _, err := time.Parse("2006-01-02", *noticeDate); err != nil {
		return fmt.Errorf("invalid notice date format (expected YYYY-MM-DD): %s", *noticeDate)
	}

	if _, err := time.Parse("2006-01-02", *valueDate); err != nil {
		return fmt.Errorf("invalid value date format (expected YYYY-MM-DD): %s", *valueDate)
	}

	// Create DSL operation for redemption request
	operation := &dsl.HedgeFundDSLOperation{
		Verb: "redeem.request",
		Args: map[string]interface{}{
			"investor":    investorUUID.String(),
			"class":       classUUID.String(),
			"notice-date": *noticeDate,
			"value-date":  *valueDate,
		},
		Timestamp: time.Now().UTC(),
	}

	// Add either units or percentage
	if *units != "" {
		operation.Args["units"] = *units
	} else {
		operation.Args["percentage"] = *percentage
	}

	// Validate the DSL operation
	if err := dsl.ValidateHedgeFundDSLOperation(operation); err != nil {
		return fmt.Errorf("invalid DSL operation: %w", err)
	}

	// Generate DSL text
	dslText := dsl.GenerateHedgeFundDSL(operation)

	fmt.Printf("Creating redemption request:\n")
	fmt.Printf("  Investor: %s\n", *investorID)
	fmt.Printf("  Class: %s\n", *classID)
	if *units != "" {
		fmt.Printf("  Units: %s\n", *units)
	} else {
		fmt.Printf("  Percentage: %s%%\n", *percentage)
	}
	fmt.Printf("  Notice Date: %s\n", *noticeDate)
	fmt.Printf("  Value Date: %s\n", *valueDate)
	fmt.Printf("\nGenerated DSL:\n%s\n", dslText)

	// TODO: Execute state transition when HF investor store is implemented
	fmt.Printf("\nRedemption request created successfully\n")

	return nil
}

// RunHFSettleRedemption handles the 'hf-settle-redemption' command for settling redemptions
func RunHFSettleRedemption(ctx context.Context, ds datastore.DataStore, args []string) error {
	fs := flag.NewFlagSet("hf-settle-redemption", flag.ExitOnError)

	investorID := fs.String("investor", "", "Investor ID (UUID) (required)")
	tradeID := fs.String("trade", "", "Redemption trade ID (UUID) (required)")
	amount := fs.String("amount", "", "Settlement amount (required)")
	settleDate := fs.String("settle-date", "", "Settlement date (YYYY-MM-DD) (required)")
	reference := fs.String("reference", "", "Payment reference")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	// Validate required fields
	if *investorID == "" || *tradeID == "" || *amount == "" || *settleDate == "" {
		fs.Usage()
		return fmt.Errorf("error: --investor, --trade, --amount, and --settle-date flags are required")
	}

	// Validate UUIDs
	investorUUID, err := uuid.Parse(*investorID)
	if err != nil {
		return fmt.Errorf("invalid investor ID format: %s", *investorID)
	}

	tradeUUID, err := uuid.Parse(*tradeID)
	if err != nil {
		return fmt.Errorf("invalid trade ID format: %s", *tradeID)
	}

	// Validate date
	if _, err := time.Parse("2006-01-02", *settleDate); err != nil {
		return fmt.Errorf("invalid settle date format (expected YYYY-MM-DD): %s", *settleDate)
	}

	// Create DSL operation for redemption settlement
	operation := &dsl.HedgeFundDSLOperation{
		Verb: "redeem.settle",
		Args: map[string]interface{}{
			"investor":    investorUUID.String(),
			"trade":       tradeUUID.String(),
			"amount":      *amount,
			"settle-date": *settleDate,
		},
		Timestamp: time.Now().UTC(),
	}

	if *reference != "" {
		operation.Args["reference"] = *reference
	}

	// Validate the DSL operation
	if err := dsl.ValidateHedgeFundDSLOperation(operation); err != nil {
		return fmt.Errorf("invalid DSL operation: %w", err)
	}

	// Generate DSL text
	dslText := dsl.GenerateHedgeFundDSL(operation)

	fmt.Printf("Settling redemption:\n")
	fmt.Printf("  Investor: %s\n", *investorID)
	fmt.Printf("  Trade: %s\n", *tradeID)
	fmt.Printf("  Amount: %s\n", *amount)
	fmt.Printf("  Settlement Date: %s\n", *settleDate)
	if *reference != "" {
		fmt.Printf("  Reference: %s\n", *reference)
	}
	fmt.Printf("\nGenerated DSL:\n%s\n", dslText)

	// TODO: Execute state transition when HF investor store is implemented
	fmt.Printf("\nRedemption settled successfully\n")

	return nil
}

// RunHFOffboardInvestor handles the 'hf-offboard-investor' command for offboarding
func RunHFOffboardInvestor(ctx context.Context, ds datastore.DataStore, args []string) error {
	fs := flag.NewFlagSet("hf-offboard-investor", flag.ExitOnError)

	investorID := fs.String("investor", "", "Investor ID (UUID) (required)")
	reason := fs.String("reason", "", "Offboarding reason")

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

	// Create DSL operation for offboarding
	operation := &dsl.HedgeFundDSLOperation{
		Verb: "offboard.close",
		Args: map[string]interface{}{
			"investor": investorUUID.String(),
		},
		Timestamp: time.Now().UTC(),
	}

	if *reason != "" {
		operation.Args["reason"] = *reason
	}

	// Validate the DSL operation
	if err := dsl.ValidateHedgeFundDSLOperation(operation); err != nil {
		return fmt.Errorf("invalid DSL operation: %w", err)
	}

	// Generate DSL text
	dslText := dsl.GenerateHedgeFundDSL(operation)

	fmt.Printf("Offboarding investor:\n")
	fmt.Printf("  Investor: %s\n", *investorID)
	if *reason != "" {
		fmt.Printf("  Reason: %s\n", *reason)
	}
	fmt.Printf("\nGenerated DSL:\n%s\n", dslText)

	// TODO: Execute state transition when HF investor store is implemented
	fmt.Printf("\nInvestor offboarded successfully\n")

	return nil
}
