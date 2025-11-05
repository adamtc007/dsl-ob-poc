package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"dsl-ob-poc/internal/datastore"
)

// SeedProductRequirementsCommand creates the seed-product-requirements CLI command
func SeedProductRequirementsCommand(ds datastore.DataStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "seed-product-requirements",
		Short: "Seed product requirements into database (Phase 5)",
		Long: `Seed product requirements and entity-product mappings into the database.

This command implements Phase 5 of the Multi-DSL Orchestration Implementation Plan
by populating the database with product-driven workflow customization data.

This replaces hardcoded data mocks with proper database-backed product requirements.

Examples:
  # Seed all product requirements
  ./dsl-poc seed-product-requirements

  # Verify seeding
  ./dsl-poc seed-product-requirements --verify`,
		RunE: func(cmd *cobra.Command, args []string) error {
			verify, _ := cmd.Flags().GetBool("verify")
			ctx := context.Background()
			return RunSeedProductRequirements(ctx, ds, verify)
		},
	}

	cmd.Flags().Bool("verify", false, "Verify seeded data after insertion")
	return cmd
}

// RunSeedProductRequirements executes the seeding of product requirements
func RunSeedProductRequirements(ctx context.Context, ds datastore.DataStore, verify bool) error {
	fmt.Println("🌱 Seeding Product Requirements (Phase 5)...")

	// Use the actual database implementation via SeedProductRequirements
	err := ds.SeedProductRequirements(ctx)
	if err != nil {
		return fmt.Errorf("failed to seed product requirements into database: %w", err)
	}

	fmt.Println("✅ Product requirements successfully seeded into database!")

	if verify {
		fmt.Println("\n🔍 Verifying seeded data...")

		// Verify by listing all product requirements from database
		requirements, err := ds.ListProductRequirements(ctx)
		if err != nil {
			return fmt.Errorf("failed to verify product requirements: %w", err)
		}

		fmt.Printf("📊 Found %d product requirements in database:\n", len(requirements))
		for i, req := range requirements {
			fmt.Printf("   [%d/%d] %s - %d entities, %d DSL verbs, %d compliance rules\n",
				i+1, len(requirements), req.ProductName,
				len(req.EntityTypes), len(req.RequiredDSL), len(req.Compliance))

			if len(req.ConditionalRules) > 0 {
				fmt.Printf("     🔀 %d conditional rules\n", len(req.ConditionalRules))
			}
		}

		fmt.Println("\n✅ Database verification complete!")
	}

	fmt.Println()
	fmt.Println("🎉 Phase 5 Product Requirements Integration Complete!")
	fmt.Println("📋 Benefits:")
	fmt.Println("   ✅ Product requirements stored in database")
	fmt.Println("   ✅ Entity-product compatibility matrix available")
	fmt.Println("   ✅ Dynamic workflow generation ready")
	fmt.Println("   ✅ No more hardcoded product data")
	fmt.Println()
	fmt.Println("🚀 Next: Phase 6 - Compile-Time Optimization & Execution Planning")

	return nil
}

// Note: Seed data is now handled by the Store.SeedProductRequirements() method
// This removes hardcoded data from CLI commands and centralizes it in the database layer
