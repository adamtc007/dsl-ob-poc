package cli

import (
	"context"
	"flag"
	"fmt"
	"log"

	"dsl-ob-poc/internal/agent"
	"dsl-ob-poc/internal/dsl"
	"dsl-ob-poc/internal/store"
)

// RunDiscoverKYC handles the 'discover-kyc' command (new Step 3 powered by the AI agent).
func RunDiscoverKYC(ctx context.Context, s *store.Store, ai *agent.Agent, args []string) error {
	fs := flag.NewFlagSet("discover-kyc", flag.ExitOnError)
	cbuID := fs.String("cbu", "", "The CBU ID of the case to discover (required)")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	if *cbuID == "" {
		fs.Usage()
		return fmt.Errorf("error: --cbu flag is required")
	}

	if ai == nil {
		return fmt.Errorf("ai agent is not configured; set GEMINI_API_KEY and try again")
	}

	log.Printf("Starting KYC discovery (Agent Step 3) for CBU: %s", *cbuID)

	// 1. Get the latest DSL (should be v2).
	currentDSL, err := s.GetLatestDSL(ctx, *cbuID)
	if err != nil {
		return err
	}

	// 2. Parse the DSL for the inputs needed by the agent.
	naturePurpose, err := dsl.ParseNaturePurpose(currentDSL)
	if err != nil {
		return fmt.Errorf("failed to parse nature-purpose from DSL: %w", err)
	}

	productNames, err := dsl.ParseProductNames(currentDSL)
	if err != nil {
		return fmt.Errorf("failed to parse products from DSL: %w", err)
	}

	log.Printf("Found Nature/Purpose: %q", naturePurpose)
	log.Printf("Found Products: %v", productNames)

	// 3. Call the AI Agent to determine the KYC requirements.
	log.Println("Calling AI Agent (Gemini) to determine KYC requirements...")
	kycReqs, err := ai.CallKYCAgent(ctx, naturePurpose, productNames)
	if err != nil {
		return fmt.Errorf("ai agent failed: %w", err)
	}

	log.Printf("Agent response received. Required docs: %v, Jurisdictions: %v", kycReqs.Documents, kycReqs.Jurisdictions)

	// 4. Generate the new DSL with the agent's response.
	newDSL, err := dsl.AddKYCRequirements(currentDSL, *kycReqs)
	if err != nil {
		return fmt.Errorf("failed to generate new DSL: %w", err)
	}

	// 5. Save the new DSL version (v3).
	versionID, err := s.InsertDSL(ctx, *cbuID, newDSL)
	if err != nil {
		return fmt.Errorf("failed to save new DSL version: %w", err)
	}

	fmt.Printf("Created new case version (v3): %s\n", versionID)
	fmt.Println("---")
	fmt.Println(newDSL)
	fmt.Println("---")

	return nil
}
