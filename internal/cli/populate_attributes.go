package cli

import (
	"context"
	"flag"
	"fmt"
	"log"

	"dsl-ob-poc/internal/dsl"
	"dsl-ob-poc/internal/store"
)

// RunPopulateAttributes implements the 6th state of the onboarding DSL:
// Populates attribute values from runtime sources and generates final DSL.
func RunPopulateAttributes(ctx context.Context, s *store.Store, args []string) error {
	fs := flag.NewFlagSet("populate-attributes", flag.ExitOnError)
	cbuID := fs.String("cbu", "", "The CBU ID of the case to populate (required)")

	if parseErr := fs.Parse(args); parseErr != nil {
		return fmt.Errorf("failed to parse flags: %w", parseErr)
	}

	if *cbuID == "" {
		return fmt.Errorf("--cbu flag is required")
	}

	log.Printf("Starting attribute population (Step 6) for CBU: %s", *cbuID)

	// Get the current DSL state
	currentDSL, err := s.GetLatestDSL(ctx, *cbuID)
	if err != nil {
		return fmt.Errorf("failed to get current DSL: %w", err)
	}

	// Parse all attribute references from the DSL
	attributeRefs, err := dsl.ParseAttributeReferences(currentDSL)
	if err != nil {
		return fmt.Errorf("failed to parse attribute references: %w", err)
	}

	log.Printf("Found %d attribute references to populate", len(attributeRefs))

	// Populate attribute values from runtime sources
	populatedValues, err := dsl.PopulateAttributeValues(ctx, s, *cbuID, attributeRefs)
	if err != nil {
		return fmt.Errorf("failed to populate attribute values: %w", err)
	}

	// Generate final DSL with populated values
	finalDSL, err := dsl.AddPopulatedAttributes(currentDSL, populatedValues)
	if err != nil {
		return fmt.Errorf("failed to generate final DSL: %w", err)
	}

	// Save the final DSL version
	versionID, err := s.InsertDSL(ctx, *cbuID, finalDSL)
	if err != nil {
		return fmt.Errorf("failed to save final DSL: %w", err)
	}

	log.Printf("✅ Attribute population completed successfully!")
	log.Printf("📊 Populated %d attributes", len(populatedValues))
	log.Printf("💾 Final DSL saved as version: %s", versionID)

	return nil
}
