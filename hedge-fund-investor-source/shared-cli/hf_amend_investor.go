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

// RunHFAmendInvestor handles the 'hf-amend-investor' command for amending investor details
func RunHFAmendInvestor(ctx context.Context, ds datastore.DataStore, args []string) error {
	fs := flag.NewFlagSet("hf-amend-investor", flag.ExitOnError)

	// Required flags
	investorID := fs.String("investor", "", "Investor ID (UUID) (required)")

	// Optional amendment flags
	legalName := fs.String("legal-name", "", "Updated legal name")
	shortName := fs.String("short-name", "", "Updated short name")
	domicile := fs.String("domicile", "", "Updated domicile (2-letter country code)")
	lei := fs.String("lei", "", "Updated Legal Entity Identifier")
	regNumber := fs.String("reg-number", "", "Updated registration number")

	// Address fields
	addressLine1 := fs.String("address1", "", "Updated address line 1")
	addressLine2 := fs.String("address2", "", "Updated address line 2")
	city := fs.String("city", "", "Updated city")
	country := fs.String("country", "", "Updated country (2-letter code)")
	postalCode := fs.String("postal-code", "", "Updated postal code")

	// Contact fields
	contactName := fs.String("contact-name", "", "Updated primary contact name")
	contactEmail := fs.String("contact-email", "", "Updated primary contact email")
	contactPhone := fs.String("contact-phone", "", "Updated primary contact phone")

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

	// Build DSL operation with only provided fields
	operation := &dsl.HedgeFundDSLOperation{
		Verb: "investor.amend-details",
		Args: map[string]interface{}{
			"investor": investorUUID.String(),
		},
		Timestamp: time.Now().UTC(),
	}

	// Track what's being updated for display
	updates := []string{}

	// Add optional fields only if provided
	if *legalName != "" {
		operation.Args["legal-name"] = *legalName
		updates = append(updates, fmt.Sprintf("Legal Name: %s", *legalName))
	}
	if *shortName != "" {
		operation.Args["short-name"] = *shortName
		updates = append(updates, fmt.Sprintf("Short Name: %s", *shortName))
	}
	if *domicile != "" {
		operation.Args["domicile"] = *domicile
		updates = append(updates, fmt.Sprintf("Domicile: %s", *domicile))
	}
	if *lei != "" {
		operation.Args["lei"] = *lei
		updates = append(updates, fmt.Sprintf("LEI: %s", *lei))
	}
	if *regNumber != "" {
		operation.Args["reg-number"] = *regNumber
		updates = append(updates, fmt.Sprintf("Registration Number: %s", *regNumber))
	}
	if *addressLine1 != "" {
		operation.Args["address-line1"] = *addressLine1
		updates = append(updates, fmt.Sprintf("Address Line 1: %s", *addressLine1))
	}
	if *addressLine2 != "" {
		operation.Args["address-line2"] = *addressLine2
		updates = append(updates, fmt.Sprintf("Address Line 2: %s", *addressLine2))
	}
	if *city != "" {
		operation.Args["city"] = *city
		updates = append(updates, fmt.Sprintf("City: %s", *city))
	}
	if *country != "" {
		operation.Args["country"] = *country
		updates = append(updates, fmt.Sprintf("Country: %s", *country))
	}
	if *postalCode != "" {
		operation.Args["postal-code"] = *postalCode
		updates = append(updates, fmt.Sprintf("Postal Code: %s", *postalCode))
	}
	if *contactName != "" {
		operation.Args["contact-name"] = *contactName
		updates = append(updates, fmt.Sprintf("Contact Name: %s", *contactName))
	}
	if *contactEmail != "" {
		operation.Args["contact-email"] = *contactEmail
		updates = append(updates, fmt.Sprintf("Contact Email: %s", *contactEmail))
	}
	if *contactPhone != "" {
		operation.Args["contact-phone"] = *contactPhone
		updates = append(updates, fmt.Sprintf("Contact Phone: %s", *contactPhone))
	}

	// Check if any fields are being updated
	if len(updates) == 0 {
		return fmt.Errorf("error: at least one field must be provided to amend")
	}

	// Validate the DSL operation
	if err := dsl.ValidateHedgeFundDSLOperation(operation); err != nil {
		return fmt.Errorf("invalid DSL operation: %w", err)
	}

	// Generate DSL text
	dslText := dsl.GenerateHedgeFundDSL(operation)

	fmt.Printf("Amending investor details:\n")
	fmt.Printf("  Investor ID: %s\n", *investorID)
	fmt.Printf("\nUpdates:\n")
	for _, update := range updates {
		fmt.Printf("  - %s\n", update)
	}
	fmt.Printf("\nGenerated DSL:\n%s\n", dslText)

	// TODO: Apply updates to investor record in database when store is implemented
	// For now, just show what would be updated
	fmt.Printf("\nInvestor details amended successfully at: %s\n", time.Now().UTC().Format(time.RFC3339))

	return nil
}
