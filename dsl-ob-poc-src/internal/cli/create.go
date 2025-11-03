package cli

import (
	"context"
	"flag"
	"fmt"

	"dsl-ob-poc/internal/dsl"
	"dsl-ob-poc/internal/mocks"
	"dsl-ob-poc/internal/store"
)

// RunCreate handles the 'create' command.
func RunCreate(ctx context.Context, s *store.Store, args []string) error {
	fs := flag.NewFlagSet("create", flag.ExitOnError)
	cbuID := fs.String("cbu", "", "The CBU ID for the new case (required)")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	if *cbuID == "" {
		fs.Usage()
		return fmt.Errorf("error: --cbu flag is required")
	}

	mockCBU, err := mocks.GetMockCBU(*cbuID)
	if err != nil {
		return fmt.Errorf("failed to get mock data: %w", err)
	}

	// Generate the initial "CREATE" DSL
	newDSL := dsl.CreateCase(mockCBU.CBUId, mockCBU.NaturePurpose)

	versionID, err := s.InsertDSL(ctx, mockCBU.CBUId, newDSL)
	if err != nil {
		return fmt.Errorf("failed to save new case: %w", err)
	}

	fmt.Printf("Created new case version (v1): %s\n", versionID)
	fmt.Println("---")
	fmt.Println(newDSL)
	fmt.Println("---")

	return nil
}
