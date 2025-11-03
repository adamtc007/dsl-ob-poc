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

	// 2. Get the *current* state of the DSL from the DB
	currentDSL, err := ds.GetLatestDSL(ctx, *cbuID)
	if err != nil {
		return fmt.Errorf("failed to get current case for CBU %s: %w", *cbuID, err)
	}

	// 3. Pass the current DSL and *validated* products to generate the *new* state
	newDSL, err := dsl.AddProducts(currentDSL, validProducts)
	if err != nil {
		return fmt.Errorf("failed to apply state change: %w", err)
	}

	// 4. Save the *new* DSL as a new immutable version
	versionID, err := ds.InsertDSL(ctx, *cbuID, newDSL)
	if err != nil {
		return fmt.Errorf("failed to save updated case: %w", err)
	}

	fmt.Printf("Created new case version (v2): %s\n", versionID)
	fmt.Println("---")
	fmt.Println(newDSL)
	fmt.Println("---")

	return nil
}
