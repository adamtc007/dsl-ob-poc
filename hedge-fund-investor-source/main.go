package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"dsl-ob-poc/internal/agent"
	"dsl-ob-poc/internal/cli"
	"dsl-ob-poc/internal/config"
	"dsl-ob-poc/internal/datastore"
)

// getAPIKey looks for GEMINI_API_KEY first, then falls back to GOOGLE_API_KEY
func getAPIKey() string {
	if apiKey := os.Getenv("GEMINI_API_KEY"); apiKey != "" {
		return apiKey
	}
	if apiKey := os.Getenv("GOOGLE_API_KEY"); apiKey != "" {
		log.Println("ℹ️ Using GOOGLE_API_KEY for Gemini API (consider setting GEMINI_API_KEY)")
		return apiKey
	}
	return ""
}

func main() {
	os.Exit(run())
}

func run() int {
	if len(os.Args) < 2 {
		printUsage()
		return 1
	}

	command := os.Args[1]
	args := os.Args[2:]

	// Handle help command without DB connection
	if command == "help" {
		printUsage()
		return 0
	}

	// All other commands require data store connection
	cfg := config.GetDataStoreConfig()

	dataStore, err := datastore.NewDataStore(cfg)
	if err != nil {
		log.Printf("Failed to initialize data store: %v", err)
		return 1
	}
	defer dataStore.Close()

	// Print mode information for clarity
	if config.IsMockMode() {
		fmt.Printf("Running in MOCK mode (data from: %s)\n", cfg.MockDataPath)
	} else {
		fmt.Println("Running in DATABASE mode")
	}

	ctx := context.Background()

	switch command {
	case "init-db":
		err = dataStore.InitDB(ctx)
		if err != nil {
			log.Printf("Failed to initialize database: %v", err)
			return 1
		}
		fmt.Println("Database initialized successfully.")

	case "seed-catalog":
		err = dataStore.SeedCatalog(ctx)
		if err != nil {
			log.Printf("Failed to seed catalog: %v", err)
			return 1
		}
		fmt.Println("Catalog seeded successfully with mock data.")

	case "create":
		err = cli.RunCreate(ctx, dataStore, args)

	case "add-products":
		err = cli.RunAddProducts(ctx, dataStore, args)

	case "discover-kyc":
		apiKey := getAPIKey()
		aiAgent, agentErr := agent.NewAgent(ctx, apiKey)
		if agentErr != nil {
			log.Printf("Failed to initialize AI agent: %v", agentErr)
			return 1
		}
		if aiAgent == nil {
			log.Println("Error: Neither GEMINI_API_KEY nor GOOGLE_API_KEY environment variable is set.")
			return 1
		}
		defer aiAgent.Close()

		err = cli.RunDiscoverKYC(ctx, dataStore, aiAgent, args)

	case "agent-transform":
		apiKey := getAPIKey()
		aiAgent, agentErr := agent.NewAgent(ctx, apiKey)
		if agentErr != nil {
			log.Printf("Failed to initialize AI agent: %v", agentErr)
			return 1
		}
		if aiAgent == nil {
			log.Println("Error: Neither GEMINI_API_KEY nor GOOGLE_API_KEY environment variable is set.")
			return 1
		}
		defer aiAgent.Close()

		err = cli.RunAgentTransform(ctx, dataStore, aiAgent, args)

	case "agent-validate":
		apiKey := getAPIKey()
		aiAgent, agentErr := agent.NewAgent(ctx, apiKey)
		if agentErr != nil {
			log.Printf("Failed to initialize AI agent: %v", agentErr)
			return 1
		}
		if aiAgent == nil {
			log.Println("Error: Neither GEMINI_API_KEY nor GOOGLE_API_KEY environment variable is set.")
			return 1
		}
		defer aiAgent.Close()

		err = cli.RunAgentValidate(ctx, dataStore, aiAgent, args)

	case "agent-demo":
		err = cli.RunAgentDemo(ctx, dataStore, args)

	case "agent-test":
		err = cli.RunAgentTest(ctx, dataStore, args)

	case "agent-prompt-capture":
		apiKey := getAPIKey()
		aiAgent, agentErr := agent.NewAgent(ctx, apiKey)
		if agentErr != nil {
			log.Printf("Failed to initialize AI agent: %v", agentErr)
		}
		// Allow running with or without API key for prompt capture demonstration
		err = cli.RunAgentPromptCapture(ctx, dataStore, aiAgent, args)

	case "discover-services":
		err = cli.RunDiscoverServices(ctx, dataStore, args)

	case "discover-resources":
		err = cli.RunDiscoverResources(ctx, dataStore, args)

	case "populate-attributes":
		err = cli.RunPopulateAttributes(ctx, dataStore, args)

	case "get-attribute-values":
		err = cli.RunGetAttributeValues(ctx, dataStore, args)

	// NEW COMMAND
	case "history":
		err = cli.RunHistory(ctx, dataStore, args)

	// CBU CRUD COMMANDS
	case "cbu-create":
		err = cli.RunCBUCreate(ctx, dataStore, args)
	case "cbu-list":
		err = cli.RunCBUList(ctx, dataStore, args)
	case "cbu-get":
		err = cli.RunCBUGet(ctx, dataStore, args)
	case "cbu-update":
		err = cli.RunCBUUpdate(ctx, dataStore, args)
	case "cbu-delete":
		err = cli.RunCBUDelete(ctx, dataStore, args)

	// ROLE CRUD COMMANDS
	case "role-create":
		err = cli.RunRoleCreate(ctx, dataStore, args)
	case "role-list":
		err = cli.RunRoleList(ctx, dataStore, args)
	case "role-get":
		err = cli.RunRoleGet(ctx, dataStore, args)
	case "role-update":
		err = cli.RunRoleUpdate(ctx, dataStore, args)
	case "role-delete":
		err = cli.RunRoleDelete(ctx, dataStore, args)

	// MOCK DATA EXPORT
	case "export-mock-data":
		err = cli.RunExportMockData(ctx, dataStore, args)

	// DSL S-EXPRESSION EXECUTION
	case "dsl-execute":
		err = cli.RunDSLExecute(ctx, dataStore, args)

	// HEDGE FUND INVESTOR COMMANDS
	case "hf-create-investor":
		err = cli.RunHFCreateInvestor(ctx, dataStore, args)
	case "hf-record-indication":
		err = cli.RunHFRecordIndication(ctx, dataStore, args)
	case "hf-begin-kyc":
		err = cli.RunHFBeginKYC(ctx, dataStore, args)
	case "hf-approve-kyc":
		err = cli.RunHFApproveKYC(ctx, dataStore, args)
	case "hf-capture-tax":
		err = cli.RunHFCaptureTax(ctx, dataStore, args)
	case "hf-set-bank-instruction":
		err = cli.RunHFSetBankInstruction(ctx, dataStore, args)
	case "hf-collect-document":
		err = cli.RunHFCollectDocument(ctx, dataStore, args)
	case "hf-screen-investor":
		err = cli.RunHFScreenInvestor(ctx, dataStore, args)
	case "hf-subscribe-request":
		err = cli.RunHFSubscribeRequest(ctx, dataStore, args)
	case "hf-confirm-cash":
		err = cli.RunHFConfirmCash(ctx, dataStore, args)
	case "hf-set-nav":
		err = cli.RunHFSetNAV(ctx, dataStore, args)
	case "hf-issue-units":
		err = cli.RunHFIssueUnits(ctx, dataStore, args)
	case "hf-redeem-request":
		err = cli.RunHFRedeemRequest(ctx, dataStore, args)
	case "hf-settle-redemption":
		err = cli.RunHFSettleRedemption(ctx, dataStore, args)
	case "hf-offboard-investor":
		err = cli.RunHFOffboardInvestor(ctx, dataStore, args)
	case "hf-set-refresh-schedule":
		err = cli.RunHFSetRefreshSchedule(ctx, dataStore, args)
	case "hf-set-continuous-screening":
		err = cli.RunHFSetContinuousScreening(ctx, dataStore, args)
	case "hf-show-register":
		err = cli.RunHFShowRegister(ctx, dataStore, args)
	case "hf-show-kyc-dashboard":
		err = cli.RunHFShowKYCDashboard(ctx, dataStore, args)

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		return 1
	}

	if err != nil {
		log.Printf("Command failed: %v", err)
		return 1
	}

	return 0
}

func printUsage() {
	fmt.Println("Onboarding DSL POC CLI (v9: Hedge Fund Investor Register)")
	fmt.Println("Usage: dsl-poc <command> [options]")
	fmt.Println("\nEnvironment Variables:")
	fmt.Println("  DSL_STORE_TYPE         Set to 'mock' for disconnected mode, 'postgresql' for database mode (default)")
	fmt.Println("  DSL_MOCK_DATA_PATH     Path to mock data directory (default: data/mocks)")
	fmt.Println("  DB_CONN_STRING         PostgreSQL connection string (required for database mode)")
	fmt.Println("\nSetup Commands:")
	fmt.Println("  init-db                      (One-time) Initializes the PostgreSQL schema and all tables.")
	fmt.Println("  seed-catalog                 (One-time) Populates catalog tables with mock data.")
	fmt.Println("\nState Machine Commands:")
	fmt.Println("  create --cbu=<cbu-id>        (v1) Creates a new onboarding case.")
	fmt.Println("  add-products --cbu=<cbu-id>  (v2) Adds products to an existing case.")
	fmt.Println("               --products=<p1,p2>")
	fmt.Println("  discover-kyc --cbu=<cbu-id>  (v3) Performs AI-assisted KYC discovery.")
	fmt.Println("  discover-services --cbu=<cbu-id> (v4) Discovers and appends services plan.")
	fmt.Println("  discover-resources --cbu=<cbu-id> (v5) Discovers and appends resources plan.")
	fmt.Println("  populate-attributes --cbu=<cbu-id> (v6) Populates attribute values from runtime sources.")
	fmt.Println("  get-attribute-values --cbu=<cbu-id> (v7) Resolves and binds attribute values deterministically.")
	fmt.Println("\nHedge Fund Investor Commands:")
	fmt.Println("  Investor Lifecycle:")
	fmt.Println("    hf-create-investor --code=<code> --legal-name=<name> --type=<type> --domicile=<country>")
	fmt.Println("                       [--short-name=<name>] [--contact-email=<email>] [--address1=<addr>]")
	fmt.Println("    hf-record-indication --investor=<uuid> --fund=<uuid> --class=<uuid> --ticket=<amount> --currency=<ccy>")
	fmt.Println("  KYC & Compliance:")
	fmt.Println("    hf-begin-kyc --investor=<uuid> [--tier=<SIMPLIFIED|STANDARD|ENHANCED>]")
	fmt.Println("    hf-collect-document --investor=<uuid> --doc-type=<type> [--subject=<subject>] [--file-path=<path>]")
	fmt.Println("    hf-screen-investor --investor=<uuid> --provider=<worldcheck|refinitiv|accelus>")
	fmt.Println("    hf-approve-kyc --investor=<uuid> --risk=<LOW|MEDIUM|HIGH> --refresh-due=<YYYY-MM-DD> --approved-by=<name>")
	fmt.Println("    hf-set-refresh-schedule --investor=<uuid> --frequency=<MONTHLY|QUARTERLY|ANNUAL> --next=<YYYY-MM-DD>")
	fmt.Println("    hf-set-continuous-screening --investor=<uuid> --frequency=<DAILY|WEEKLY|MONTHLY>")
	fmt.Println("  Tax & Banking:")
	fmt.Println("    hf-capture-tax --investor=<uuid> [--fatca=<status>] [--crs=<classification>] [--form=<type>]")
	fmt.Println("    hf-set-bank-instruction --investor=<uuid> --currency=<ccy> --bank-name=<name> --account-name=<name>")
	fmt.Println("                            [--iban=<iban>] [--swift=<bic>] [--account-num=<number>]")
	fmt.Println("  Trading Operations:")
	fmt.Println("    hf-subscribe-request --investor=<uuid> --fund=<uuid> --class=<uuid> --amount=<amount>")
	fmt.Println("                         --currency=<ccy> --trade-date=<YYYY-MM-DD> --value-date=<YYYY-MM-DD>")
	fmt.Println("    hf-confirm-cash --investor=<uuid> --trade=<uuid> --amount=<amount> --value-date=<YYYY-MM-DD>")
	fmt.Println("                    --bank-currency=<ccy> [--reference=<ref>]")
	fmt.Println("    hf-set-nav --fund=<uuid> --class=<uuid> --nav-date=<YYYY-MM-DD> --nav=<amount>")
	fmt.Println("    hf-issue-units --investor=<uuid> --trade=<uuid> --class=<uuid> [--series=<uuid>]")
	fmt.Println("                   --nav-per-share=<amount> --units=<amount>")
	fmt.Println("  Redemption & Offboarding:")
	fmt.Println("    hf-redeem-request --investor=<uuid> --class=<uuid> [--units=<amount>|--percentage=<pct>]")
	fmt.Println("                      --notice-date=<YYYY-MM-DD> --value-date=<YYYY-MM-DD>")
	fmt.Println("    hf-settle-redemption --investor=<uuid> --trade=<uuid> --amount=<amount>")
	fmt.Println("                         --settle-date=<YYYY-MM-DD> [--reference=<ref>]")
	fmt.Println("    hf-offboard-investor --investor=<uuid> [--reason=<reason>]")
	fmt.Println("  Reporting:")
	fmt.Println("    hf-show-register [--fund=<uuid>] [--class=<uuid>] [--status=<status>] [--format=<table|json|csv>]")
	fmt.Println("    hf-show-kyc-dashboard [--risk=<LOW|MEDIUM|HIGH>] [--status=<status>] [--overdue]")
	fmt.Println("\nAI Agent Commands (requires GEMINI_API_KEY):")
	fmt.Println("  agent-transform --cbu=<cbu-id>   AI-powered DSL transformation with natural language instructions")
	fmt.Println("                  --instruction=<text> [--target-state=<state>] [--save]")
	fmt.Println("  agent-validate --cbu=<cbu-id>    AI-powered DSL validation and improvement suggestions")
	fmt.Println("  agent-demo [--cbu=<cbu-id>]      Demonstrates AI agent capabilities (no API key required)")
	fmt.Println("  agent-test [--cbu=<cbu-id>]      Tests AI agents with mock responses (no API key required)")
	fmt.Println("             [--type=<kyc|transform|validate|all>]")
	fmt.Println("  agent-prompt-capture [--cbu=<cbu-id>] [--type=<kyc|transform|validate|all>] [--output=<file>]")
	fmt.Println("                       Captures and displays exact AI prompts and responses for analysis")
	fmt.Println("\nCBU Management Commands:")
	fmt.Println("  cbu-create --name=<name> [--description=<desc>] [--nature-purpose=<purpose>]")
	fmt.Println("  cbu-list                     Lists all CBUs")
	fmt.Println("  cbu-get --id=<cbu-id>        Get CBU details")
	fmt.Println("  cbu-update --id=<cbu-id> [--name=<name>] [--description=<desc>] [--nature-purpose=<purpose>]")
	fmt.Println("  cbu-delete --id=<cbu-id>     Delete CBU")
	fmt.Println("\nRole Management Commands:")
	fmt.Println("  role-create --name=<name> [--description=<desc>]")
	fmt.Println("  role-list                    Lists all roles")
	fmt.Println("  role-get --id=<role-id>      Get role details")
	fmt.Println("  role-update --id=<role-id> [--name=<name>] [--description=<desc>]")
	fmt.Println("  role-delete --id=<role-id>   Delete role")
	fmt.Println("\nUtility Commands:")
	fmt.Println("  history --cbu=<cbu-id>       Views the full, versioned DSL evolution for a case.")
	fmt.Println("  export-mock-data [--dir=<path>] Exports existing database records to JSON mock files")
	fmt.Println("\nDSL Execution Engine:")
	fmt.Println("  dsl-execute [--cbu=<cbu-id>] [--demo] [--file=<path>] [dsl-command]")
	fmt.Println("              Execute S-expression DSL commands with UUID attribute handling")
	fmt.Println("              --demo: Run comprehensive demo workflow")
	fmt.Println("              --file: Execute commands from file")
	fmt.Println("              Interactive mode if no command specified")
}
