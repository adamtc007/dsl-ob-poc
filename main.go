package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"dsl-ob-poc/internal/agent"
	"dsl-ob-poc/internal/cli"
	"dsl-ob-poc/internal/store"
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

	// All other commands require DB connection
	connString := os.Getenv("DB_CONN_STRING")
	if connString == "" {
		log.Println("Error: DB_CONN_STRING environment variable is not set.")
		return 1
	}

	dbStore, err := store.NewStore(connString)
	if err != nil {
		log.Printf("Failed to initialize database store: %v", err)
		return 1
	}
	defer dbStore.Close()

	ctx := context.Background()

	switch command {
	case "init-db":
		err = dbStore.InitDB(ctx)
		if err != nil {
			log.Printf("Failed to initialize database: %v", err)
			return 1
		}
		fmt.Println("Database initialized successfully.")

	case "seed-catalog":
		err = dbStore.SeedCatalog(ctx)
		if err != nil {
			log.Printf("Failed to seed catalog: %v", err)
			return 1
		}
		fmt.Println("Catalog seeded successfully with mock data.")

	case "create":
		err = cli.RunCreate(ctx, dbStore, args)

	case "add-products":
		err = cli.RunAddProducts(ctx, dbStore, args)

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

		err = cli.RunDiscoverKYC(ctx, dbStore, aiAgent, args)

	case "discover-services":
		err = cli.RunDiscoverServices(ctx, dbStore, args)

	case "discover-resources":
		err = cli.RunDiscoverResources(ctx, dbStore, args)

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
	fmt.Println("Onboarding DSL POC CLI (v4: Agent-Aware)")
	fmt.Println("Usage: dsl-poc <command> [options]")
	fmt.Println("\nSetup Commands:")
	fmt.Println("  init-db                      (One-time) Initializes the PostgreSQL schema and all tables.")
	fmt.Println("  seed-catalog                 (One-time) Populates catalog tables with mock data.")
	fmt.Println("\nState Machine Commands:")
	fmt.Println("  create --cbu=<cbu-id>        (v1) Creates a new onboarding case.")
	fmt.Println("  add-products --cbu=<cbu-id>  (v2) Adds products to an existing case.")
	fmt.Println("               --products=<p1,p2>")
	fmt.Println("  discover-kyc --cbu=<cbu-id> (v3) Performs AI-assisted KYC discovery.")
	fmt.Println("  discover-services --cbu=<cbu-id> (v4) Discovers and appends services plan.")
	fmt.Println("  discover-resources --cbu=<cbu-id> (v5) Discovers and appends resources plan.")
}
