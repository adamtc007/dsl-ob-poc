package cli

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"dsl-ob-poc/internal/datastore"
)

// RunHistory handles the 'history' command.
func RunHistory(ctx context.Context, ds datastore.DataStore, args []string) error {
	fs := flag.NewFlagSet("history", flag.ExitOnError)
	cbuID := fs.String("cbu", "", "The CBU ID of the case to view (required)")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	if *cbuID == "" {
		fs.Usage()
		return fmt.Errorf("error: --cbu flag is required")
	}

	log.Printf("Fetching DSL history for CBU: %s", *cbuID)

	// 1. Get all DSL versions from the store
	history, err := ds.GetDSLHistory(ctx, *cbuID)
	if err != nil {
		return err
	}

	// 2. Print the full history
	fmt.Printf("\n--- DSL State Evolution for CBU: %s ---\n", *cbuID)
	fmt.Printf("Found %d versions.\n", len(history))

	for i, version := range history {
		fmt.Printf("\n===========================================\n")
		fmt.Printf("Version %d (State %d)\n", i+1, i+1)
		fmt.Printf("Version ID: %s\n", version.VersionID)
		fmt.Printf("Created At: %s\n", version.CreatedAt.Format(time.RFC3339))
		fmt.Printf("-------------------------------------------\n")
		fmt.Println(version.DSLText)
		fmt.Printf("===========================================\n")
	}

	return nil
}
