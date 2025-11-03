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
		apiKey := os.Getenv("GEMINI_API_KEY")
		aiAgent, agentErr := agent.NewAgent(ctx, apiKey)
		if agentErr != nil {
			log.Printf("Failed to initialize AI agent: %v", agentErr)
			return 1
		}
		if aiAgent == nil {
			log.Println("Error: GEMINI_API_KEY environment variable is not set.")
			return 1
		}
		defer aiAgent.Close()

		err = cli.RunDiscoverKYC(ctx, dataStore, aiAgent, args)

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
	fmt.Println("Onboarding DSL POC CLI (v8: Entity Relationship Model)")
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
}
