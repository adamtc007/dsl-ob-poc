package cli

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"dsl-ob-poc/internal/datastore"
	"dsl-ob-poc/internal/dsl"
	"dsl-ob-poc/internal/store"
)

// RunAddProducts handles the 'add-products' command.
func RunAddProducts(ctx context.Context, ds datastore.DataStore, args []string) error {
	fs := flag.NewFlagSet("add-products", flag.ExitOnError)
	cbuID := fs.String("cbu", "", "The CBU ID of the case to update (required)")
	productsStr := fs.String("products", "", "Comma-separated list of products (required)")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	if *cbuID == "" || *productsStr == "" {
		fs.Usage()
		return fmt.Errorf("error: --cbu and --products flags are required")
	}

	productNames := strings.Split(*productsStr, ",")
	if len(productNames) == 0 {
		return fmt.Errorf("error: no products provided")
	}

	// 1. Validate products against the catalog
	validProducts := make([]*store.Product, 0, len(productNames))
	for _, name := range productNames {
		p, err := ds.GetProductByName(ctx, strings.TrimSpace(name))
		if err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
		validProducts = append(validProducts, p)
	}
	fmt.Printf("Validated %d products against catalog.\n", len(validProducts))

	// 2. Get the current onboarding session
	session, err := ds.GetOnboardingSession(ctx, *cbuID)
	if err != nil {
		return fmt.Errorf("failed to get onboarding session for CBU %s: %w", *cbuID, err)
	}

	// 3. Get the *current* state of the DSL from the DB with state information
	currentDSLState, err := ds.GetLatestDSLWithState(ctx, *cbuID)
	if err != nil {
		return fmt.Errorf("failed to get current case for CBU %s: %w", *cbuID, err)
	}

	// 4. Pass the current DSL and *validated* products to generate the *new* state
	newDSL, err := dsl.AddProducts(currentDSLState.DSLText, validProducts)
	if err != nil {
		return fmt.Errorf("failed to apply state change: %w", err)
	}

	// 5. Save the *new* DSL with PRODUCTS_ADDED state
	versionID, err := ds.InsertDSLWithState(ctx, *cbuID, newDSL, store.StateProductsAdded)
	if err != nil {
		return fmt.Errorf("failed to save updated case: %w", err)
	}

	// 6. Update onboarding session state
	err = ds.UpdateOnboardingState(ctx, *cbuID, store.StateProductsAdded, versionID)
	if err != nil {
		return fmt.Errorf("failed to update onboarding state: %w", err)
	}

	fmt.Printf("🚀 Updated case from %s to %s\n", currentDSLState.OnboardingState, store.StateProductsAdded)
	fmt.Printf("📝 DSL version (v%d): %s\n", session.CurrentVersion+1, versionID)
	fmt.Println("---")
	fmt.Println(newDSL)
	fmt.Println("---")

	return nil
}
