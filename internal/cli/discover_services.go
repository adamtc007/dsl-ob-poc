package cli

import (
	"context"
	"flag"
	"fmt"
	"log"

	"dsl-ob-poc/internal/datastore"
	"dsl-ob-poc/internal/dsl"
	"dsl-ob-poc/internal/store"
)

// RunDiscoverServices handles the 'discover-services' command (Step 4).
func RunDiscoverServices(ctx context.Context, ds datastore.DataStore, args []string) error {
	fs := flag.NewFlagSet("discover-services", flag.ExitOnError)
	cbuID := fs.String("cbu", "", "The CBU ID of the case to discover (required)")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	if *cbuID == "" {
		fs.Usage()
		return fmt.Errorf("error: --cbu flag is required")
	}

	log.Printf("Starting service discovery (Step 4) for CBU: %s", *cbuID)

	// 1. Get the current onboarding session
	session, err := ds.GetOnboardingSession(ctx, *cbuID)
	if err != nil {
		return fmt.Errorf("failed to get onboarding session for CBU %s: %w", *cbuID, err)
	}

	// 2. Get the latest DSL with state information
	currentDSLState, err := ds.GetLatestDSLWithState(ctx, *cbuID)
	if err != nil {
		return err
	}

	currentDSL := currentDSLState.DSLText

	// 2. Parse product names from DSL (simple parsing for POC)
	productNames, err := dsl.ParseProductNames(currentDSL)
	if err != nil {
		return fmt.Errorf("failed to parse products from DSL: %w", err)
	}
	log.Printf("Found %d products in DSL: %v", len(productNames), productNames)

	// 3. Discover all services from the catalog
	productServicesMap := make(map[string][]store.Service)

	for _, productName := range productNames {
		product, getErr := ds.GetProductByName(ctx, productName)
		if getErr != nil {
			return getErr
		}

		services, getErr := ds.GetServicesForProduct(ctx, product.ProductID)
		if getErr != nil {
			return getErr
		}
		productServicesMap[product.Name] = services
	}
	log.Printf("Discovery complete: found services for %d productds.", len(productServicesMap))

	// 4. Generate the new DSL with the discovered services plan
	plan := dsl.ServiceDiscoveryPlan{
		ProductServices: productServicesMap,
	}

	newDSL, err := dsl.AddDiscoveredServices(currentDSL, plan)
	if err != nil {
		return fmt.Errorf("failed to generate new DSL: %w", err)
	}

	// 5. Save the new DSL with SERVICES_DISCOVERED state
	versionID, err := ds.InsertDSLWithState(ctx, *cbuID, newDSL, store.StateServicesDiscovered)
	if err != nil {
		return fmt.Errorf("failed to save new DSL version: %w", err)
	}

	// 6. Update onboarding session state
	err = ds.UpdateOnboardingState(ctx, *cbuID, store.StateServicesDiscovered, versionID)
	if err != nil {
		return fmt.Errorf("failed to update onboarding state: %w", err)
	}

	fmt.Printf("🛠️ Updated case from %s to %s\n", currentDSLState.OnboardingState, store.StateServicesDiscovered)
	fmt.Printf("📝 DSL version (v%d): %s\n", session.CurrentVersion+1, versionID)
	fmt.Println("---")
	fmt.Println(newDSL)
	fmt.Println("---")

	return nil
}
