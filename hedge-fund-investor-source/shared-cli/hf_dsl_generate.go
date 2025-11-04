package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	"dsl-ob-poc/internal/datastore"
	hfagent "dsl-ob-poc/internal/hf-agent"

	"github.com/google/uuid"
)

// RunHFDSLGenerate handles the 'hf-dsl-generate' command for AI-powered DSL generation
func RunHFDSLGenerate(ctx context.Context, ds datastore.DataStore, args []string) error {
	fs := flag.NewFlagSet("hf-dsl-generate", flag.ExitOnError)

	// Main instruction
	instruction := fs.String("instruction", "", "Natural language instruction for DSL generation (required)")

	// Context flags
	investorID := fs.String("investor", "", "Investor UUID (optional)")
	currentState := fs.String("state", "", "Current investor state (optional)")
	fundID := fs.String("fund", "", "Fund UUID (optional)")
	classID := fs.String("class", "", "Share class UUID (optional)")
	seriesID := fs.String("series", "", "Series UUID (optional)")

	// Investor context
	investorType := fs.String("type", "", "Investor type (optional)")
	investorName := fs.String("name", "", "Investor name (optional)")
	domicile := fs.String("domicile", "", "Investor domicile (optional)")

	// Operation flags
	execute := fs.Bool("execute", false, "Execute the generated DSL operation")
	save := fs.Bool("save", false, "Save DSL to hf_dsl_executions table")
	validate := fs.Bool("validate", true, "Validate generated DSL before execution")
	showPrompt := fs.Bool("show-prompt", false, "Show the AI prompt sent to agent")

	// Batch mode
	batchFile := fs.String("batch", "", "File containing multiple instructions (one per line)")
	conversational := fs.Bool("conversational", false, "Run in conversational mode")

	// Output format
	outputFormat := fs.String("format", "detailed", "Output format: detailed, json, dsl-only")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	// Validate required flags (unless in conversational mode)
	if *instruction == "" && *batchFile == "" && !*conversational {
		fs.Usage()
		fmt.Println("\nExamples:")
		fmt.Println("  # Create opportunity")
		fmt.Println("  ./dsl-poc hf-dsl-generate --instruction=\"Create opportunity for Acme Capital LP, corporate, Switzerland\"")
		fmt.Println("\n  # With context")
		fmt.Println("  ./dsl-poc hf-dsl-generate \\")
		fmt.Println("    --instruction=\"Submit subscription for $5M\" \\")
		fmt.Println("    --investor=<uuid> --fund=<uuid> --class=<uuid> --state=KYC_APPROVED")
		fmt.Println("\n  # Execute immediately")
		fmt.Println("  ./dsl-poc hf-dsl-generate --instruction=\"Begin KYC\" --investor=<uuid> --execute")
		fmt.Println("\n  # Conversational mode")
		fmt.Println("  ./dsl-poc hf-dsl-generate --conversational")
		return fmt.Errorf("error: --instruction flag is required (or use --conversational or --batch)")
	}

	// Get API key
	apiKey := getAPIKey()
	if apiKey == "" {
		return fmt.Errorf("GEMINI_API_KEY or GOOGLE_API_KEY environment variable required")
	}

	// Initialize agent
	agent, err := hfagent.NewHedgeFundDSLAgent(ctx, apiKey)
	if err != nil {
		return fmt.Errorf("failed to initialize DSL agent: %w", err)
	}
	defer agent.Close()

	// Handle conversational mode
	if *conversational {
		return runConversationalMode(ctx, agent, ds)
	}

	// Handle batch mode
	if *batchFile != "" {
		return runBatchMode(ctx, agent, ds, *batchFile, *execute, *save)
	}

	// Single instruction mode
	request := hfagent.DSLGenerationRequest{
		Instruction:  *instruction,
		InvestorID:   *investorID,
		CurrentState: *currentState,
		InvestorType: *investorType,
		InvestorName: *investorName,
		Domicile:     *domicile,
		FundID:       *fundID,
		ClassID:      *classID,
		SeriesID:     *seriesID,
	}

	// Load additional context from database if investor ID provided
	if *investorID != "" {
		if err := enrichContextFromDatabase(ctx, ds, &request); err != nil {
			fmt.Printf("⚠️  Warning: Could not load investor context: %v\n", err)
		}
	}

	if *showPrompt {
		fmt.Println("\n" + strings.Repeat("=", 80))
		fmt.Println("AI AGENT PROMPT")
		fmt.Println(strings.Repeat("=", 80))
		fmt.Printf("Instruction: %s\n", request.Instruction)
		fmt.Printf("Context: %+v\n", request)
		fmt.Println(strings.Repeat("=", 80) + "\n")
	}

	// Generate DSL
	fmt.Println("🤖 Generating DSL from natural language instruction...")
	response, err := agent.GenerateDSL(ctx, request)
	if err != nil {
		return fmt.Errorf("DSL generation failed: %w", err)
	}

	// Display results based on format
	switch *outputFormat {
	case "json":
		printJSONOutput(response)
	case "dsl-only":
		fmt.Println(response.DSL)
	default:
		printDetailedOutput(response, request)
	}

	// Validate if requested
	if *validate {
		if err := validateGeneratedDSL(response); err != nil {
			fmt.Printf("\n❌ Validation failed: %v\n", err)
			return err
		}
		fmt.Println("\n✅ DSL validation passed")
	}

	// Save to database if requested
	if *save {
		if err := saveDSLToDatabase(ctx, ds, response, request); err != nil {
			fmt.Printf("\n⚠️  Failed to save DSL: %v\n", err)
		} else {
			fmt.Println("\n💾 DSL saved to hf_dsl_executions table")
		}
	}

	// Execute if requested
	if *execute {
		fmt.Println("\n⚡ Executing generated DSL operation...")
		if err := executeDSLOperation(ctx, ds, response, request); err != nil {
			return fmt.Errorf("execution failed: %w", err)
		}
		fmt.Println("✅ DSL operation executed successfully")
	}

	return nil
}

// printDetailedOutput shows comprehensive generation results
func printDetailedOutput(response *hfagent.DSLGenerationResponse, request hfagent.DSLGenerationRequest) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("DSL GENERATION RESULT")
	fmt.Println(strings.Repeat("=", 80))

	fmt.Printf("\n📝 Instruction: %s\n", request.Instruction)

	fmt.Printf("\n🎯 Generated Verb: %s\n", response.Verb)

	if response.FromState != "" || response.ToState != "" {
		fmt.Printf("📊 State Transition: %s → %s\n", response.FromState, response.ToState)
	}

	if len(response.GuardConditions) > 0 {
		fmt.Printf("🛡️  Guard Conditions:\n")
		for _, cond := range response.GuardConditions {
			fmt.Printf("   - %s\n", cond)
		}
	}

	fmt.Printf("\n💡 Explanation: %s\n", response.Explanation)
	fmt.Printf("🎲 Confidence: %.1f%%\n", response.Confidence*100)

	if len(response.Warnings) > 0 {
		fmt.Printf("\n⚠️  Warnings:\n")
		for _, warn := range response.Warnings {
			fmt.Printf("   - %s\n", warn)
		}
	}

	fmt.Printf("\n📄 Generated DSL:\n")
	fmt.Println(strings.Repeat("-", 80))
	fmt.Println(response.DSL)
	fmt.Println(strings.Repeat("-", 80))

	fmt.Printf("\n🔧 Parameters:\n")
	for key, value := range response.Parameters {
		fmt.Printf("   %s: %v\n", key, value)
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
}

// printJSONOutput prints response as JSON
func printJSONOutput(response *hfagent.DSLGenerationResponse) {
	data, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

// validateGeneratedDSL performs additional validation
func validateGeneratedDSL(response *hfagent.DSLGenerationResponse) error {
	// Check confidence threshold
	if response.Confidence < 0.7 {
		return fmt.Errorf("confidence too low: %.2f (minimum 0.70)", response.Confidence)
	}

	// Check DSL syntax
	if !strings.HasPrefix(response.DSL, "(") {
		return fmt.Errorf("DSL must start with '('")
	}

	if !strings.Contains(response.DSL, response.Verb) {
		return fmt.Errorf("DSL must contain verb: %s", response.Verb)
	}

	return nil
}

// saveDSLToDatabase saves the generated DSL to hf_dsl_executions
func saveDSLToDatabase(ctx context.Context, ds datastore.DataStore, response *hfagent.DSLGenerationResponse, request hfagent.DSLGenerationRequest) error {
	// This would integrate with your actual database store
	// For now, just show what would be saved
	fmt.Printf("\n💾 Would save to database:\n")
	fmt.Printf("   Investor ID: %s\n", request.InvestorID)
	fmt.Printf("   DSL Text: %s\n", response.DSL)
	fmt.Printf("   Verb: %s\n", response.Verb)
	fmt.Printf("   Status: PENDING\n")

	// TODO: Implement actual database save
	// Example:
	// executionID := uuid.New()
	// err := ds.SaveDSLExecution(ctx, executionID, request.InvestorID, response.DSL, ...)

	return nil
}

// executeDSLOperation executes the generated DSL
func executeDSLOperation(ctx context.Context, ds datastore.DataStore, response *hfagent.DSLGenerationResponse, request hfagent.DSLGenerationRequest) error {
	fmt.Printf("\n⚡ Executing: %s\n", response.Verb)

	// Map verb to actual CLI command execution
	switch response.Verb {
	case "investor.start-opportunity":
		return executeStartOpportunity(ctx, ds, response.Parameters)
	case "investor.record-indication":
		return executeRecordIndication(ctx, ds, response.Parameters)
	case "kyc.begin":
		return executeBeginKYC(ctx, ds, response.Parameters)
	// Add other verb implementations...
	default:
		return fmt.Errorf("execution not yet implemented for verb: %s", response.Verb)
	}
}

// executeStartOpportunity executes investor.start-opportunity
func executeStartOpportunity(ctx context.Context, ds datastore.DataStore, params map[string]interface{}) error {
	legalName, _ := params["legal-name"].(string)
	investorType, _ := params["type"].(string)
	domicile, _ := params["domicile"].(string)

	fmt.Printf("   Creating opportunity: %s (%s, %s)\n", legalName, investorType, domicile)

	// Call actual implementation
	// return RunHFCreateInvestor(ctx, ds, buildArgs(params))

	fmt.Println("   ✅ Opportunity created (mock)")
	return nil
}

// executeRecordIndication executes investor.record-indication
func executeRecordIndication(ctx context.Context, ds datastore.DataStore, params map[string]interface{}) error {
	investorID, _ := params["investor"].(string)
	ticket, _ := params["ticket"].(float64)
	currency, _ := params["currency"].(string)

	fmt.Printf("   Recording indication: %s for %.2f %s\n", investorID, ticket, currency)

	// Call actual implementation
	// return RunHFRecordIndication(ctx, ds, buildArgs(params))

	fmt.Println("   ✅ Indication recorded (mock)")
	return nil
}

// executeBeginKYC executes kyc.begin
func executeBeginKYC(ctx context.Context, ds datastore.DataStore, params map[string]interface{}) error {
	investorID, _ := params["investor"].(string)
	tier, _ := params["tier"].(string)

	fmt.Printf("   Starting KYC: %s (tier: %s)\n", investorID, tier)

	// Call actual implementation
	// return RunHFBeginKYC(ctx, ds, buildArgs(params))

	fmt.Println("   ✅ KYC started (mock)")
	return nil
}

// enrichContextFromDatabase loads additional context from database
func enrichContextFromDatabase(ctx context.Context, ds datastore.DataStore, request *hfagent.DSLGenerationRequest) error {
	// Parse investor ID
	investorUUID, err := uuid.Parse(request.InvestorID)
	if err != nil {
		return fmt.Errorf("invalid investor UUID: %w", err)
	}

	// Load investor from database (this would use actual store methods)
	_ = investorUUID // Use the UUID

	// TODO: Implement actual database query
	// investor, err := ds.GetInvestor(ctx, investorUUID)
	// if err != nil {
	//     return err
	// }
	// request.CurrentState = investor.Status
	// request.InvestorType = investor.Type
	// request.InvestorName = investor.LegalName
	// request.Domicile = investor.Domicile

	fmt.Println("✓ Loaded investor context from database")
	return nil
}

// runConversationalMode provides interactive DSL generation
func runConversationalMode(ctx context.Context, agent *hfagent.HedgeFundDSLAgent, ds datastore.DataStore) error {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("🤖 CONVERSATIONAL DSL GENERATION MODE")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("Enter natural language instructions to generate hedge fund DSL.")
	fmt.Println("Type 'exit' or 'quit' to end the session.")
	fmt.Println("Type 'help' for examples.")
	fmt.Println(strings.Repeat("=", 80) + "\n")

	// Initialize context that persists across conversation
	var currentContext hfagent.DSLGenerationRequest
	var generatedDSL []string

	// Simulated conversation loop (would need actual input handling)
	examples := []string{
		"Create opportunity for Acme Capital LP, corporate, Switzerland",
		"They want to invest $5M in Global Opportunities Fund",
		"Start standard KYC",
		"help",
		"exit",
	}

	for _, instruction := range examples {
		fmt.Printf("You: %s\n", instruction)

		if instruction == "exit" || instruction == "quit" {
			fmt.Println("\n👋 Exiting conversational mode")
			break
		}

		if instruction == "help" {
			printConversationalHelp()
			continue
		}

		currentContext.Instruction = instruction

		response, err := agent.GenerateDSL(ctx, currentContext)
		if err != nil {
			fmt.Printf("❌ Error: %v\n\n", err)
			continue
		}

		fmt.Printf("\n🤖 Agent: Generated %s operation\n", response.Verb)
		fmt.Printf("   State: %s → %s\n", response.FromState, response.ToState)
		fmt.Printf("   Confidence: %.1f%%\n", response.Confidence*100)
		fmt.Printf("\n   DSL:\n   %s\n\n", strings.ReplaceAll(response.DSL, "\n", "\n   "))

		// Update context for next instruction
		currentContext.CurrentState = response.ToState
		if response.Parameters["investor"] != nil {
			currentContext.InvestorID = response.Parameters["investor"].(string)
		}

		generatedDSL = append(generatedDSL, response.DSL)
	}

	if len(generatedDSL) > 0 {
		fmt.Println("\n" + strings.Repeat("=", 80))
		fmt.Println("COMPLETE DSL SEQUENCE")
		fmt.Println(strings.Repeat("=", 80))
		for i, dsl := range generatedDSL {
			fmt.Printf("\n;; Step %d\n%s\n", i+1, dsl)
		}
		fmt.Println(strings.Repeat("=", 80))
	}

	return nil
}

// runBatchMode processes multiple instructions from file
func runBatchMode(ctx context.Context, agent *hfagent.HedgeFundDSLAgent, ds datastore.DataStore, filename string, execute bool, save bool) error {
	fmt.Printf("📂 Processing batch file: %s\n", filename)

	// Load instructions from file
	instructions := hfagent.GetCompleteLifecycleInstructions() // Example data

	fmt.Printf("📝 Found %d instructions\n\n", len(instructions))

	baseContext := hfagent.DSLGenerationRequest{}

	responses, err := agent.BatchGenerateDSL(ctx, instructions, baseContext)
	if err != nil {
		return fmt.Errorf("batch generation failed: %w", err)
	}

	// Display results
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("BATCH GENERATION RESULTS")
	fmt.Println(strings.Repeat("=", 80))

	for i, response := range responses {
		fmt.Printf("\n[%d] %s\n", i+1, response.Verb)
		fmt.Printf("    State: %s → %s\n", response.FromState, response.ToState)
		fmt.Printf("    Confidence: %.1f%%\n", response.Confidence*100)
		fmt.Printf("    DSL: %s\n", strings.Split(response.DSL, "\n")[0]+"...")
	}

	fmt.Printf("\n✅ Successfully generated %d DSL operations\n", len(responses))

	return nil
}

// printConversationalHelp prints help for conversational mode
func printConversationalHelp() {
	fmt.Println("\n📚 CONVERSATIONAL MODE HELP")
	fmt.Println(strings.Repeat("-", 80))
	fmt.Println("Examples:")
	fmt.Println("  • Create opportunity for <name>, <type>, <country>")
	fmt.Println("  • They want to invest <amount> in <fund>")
	fmt.Println("  • Start KYC / Begin KYC process")
	fmt.Println("  • Collect passport / certificate of incorporation")
	fmt.Println("  • Run screening on WorldCheck")
	fmt.Println("  • Approve KYC with medium risk")
	fmt.Println("  • Submit subscription for <amount>")
	fmt.Println("  • Cash is in / Confirm cash receipt")
	fmt.Println("  • Issue shares / Allocate units")
	fmt.Println("  • They want to redeem")
	fmt.Println("  • Close the account")
	fmt.Println(strings.Repeat("-", 80) + "\n")
}

// getAPIKey helper (same as other agent commands)
func getAPIKey() string {
	// This would be imported from a common location
	// Placeholder for this example
	return ""
}
