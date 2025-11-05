package dsl_manager

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"dsl-ob-poc/internal/store"
)

type DSLManagerCLI struct {
	store *store.Store
	dm    *DSLManager
}

func NewDSLManagerCLI(store *store.Store) *DSLManagerCLI {
	return &DSLManagerCLI{
		store: store,
		dm:    NewDSLManager(*store),
	}
}

func (cli *DSLManagerCLI) Run(args []string) error {
	if len(args) < 1 {
		return cli.printUsage()
	}

	command := args[0]
	switch command {
	case "create-case":
		return cli.createCase(args[1:])
	case "update-case":
		return cli.updateCase(args[1:])
	case "get-case":
		return cli.getCase(args[1:])
	case "list-cases":
		return cli.listCases(args[1:])
	default:
		return cli.printUsage()
	}
}

func (cli *DSLManagerCLI) createCase(args []string) error {
	fs := flag.NewFlagSet("create-case", flag.ExitOnError)
	domain := fs.String("domain", "", "Domain for the case (required)")
	investorName := fs.String("investor-name", "", "Investor name")
	investorType := fs.String("investor-type", "", "Investor type")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *domain == "" {
		return fmt.Errorf("domain is required")
	}

	initialData := map[string]interface{}{
		"investor_name": *investorName,
		"investor_type": *investorType,
	}

	session, err := cli.dm.CreateCase(*domain, initialData)
	if err != nil {
		return err
	}

	// Generate initial DSL fragment based on domain
	var dslFragment string
	switch *domain {
	case "investor":
		dslFragment = fmt.Sprintf(
			`(investor.create
				(investor.name "%s")
				(investor.type "%s")
				(onboarding.id "%s")
			)`,
			*investorName,
			*investorType,
			session.SessionID,
		)
	default:
		dslFragment = fmt.Sprintf(
			`(case.create
				(domain "%s")
				(onboarding.id "%s")
			)`,
			*domain,
			session.SessionID,
		)
	}

	// Accumulate initial DSL
	if accErr := session.AccumulateDSL(dslFragment); accErr != nil {
		return accErr
	}

	// Output result
	output := struct {
		OnboardingID string `json:"onboarding_id"`
		Domain       string `json:"domain"`
		InitialState string `json:"initial_state"`
		DSL          string `json:"dsl"`
	}{
		OnboardingID: session.SessionID,
		Domain:       session.Domain,
		InitialState: session.GetContext().CurrentState,
		DSL:          session.GetDSL(),
	}

	return cli.outputJSON(output)
}

func (cli *DSLManagerCLI) updateCase(args []string) error {
	fs := flag.NewFlagSet("update-case", flag.ExitOnError)
	onboardingID := fs.String("onboarding-id", "", "Onboarding ID (required)")
	stateTransition := fs.String("state", "", "New state for the case (required)")
	dslFragment := fs.String("dsl", "", "DSL fragment to append")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *onboardingID == "" {
		return fmt.Errorf("onboarding ID is required")
	}

	if *stateTransition == "" {
		return fmt.Errorf("state transition is required")
	}

	session, err := cli.dm.UpdateCase(*onboardingID, *dslFragment, *stateTransition)
	if err != nil {
		return err
	}

	// Output result
	output := struct {
		OnboardingID   string `json:"onboarding_id"`
		PreviousState  string `json:"previous_state"`
		CurrentState   string `json:"current_state"`
		AccumulatedDSL string `json:"accumulated_dsl"`
	}{
		OnboardingID:   session.SessionID,
		PreviousState:  session.GetContext().CurrentState,
		CurrentState:   *stateTransition,
		AccumulatedDSL: session.GetDSL(),
	}

	return cli.outputJSON(output)
}

func (cli *DSLManagerCLI) getCase(args []string) error {
	fs := flag.NewFlagSet("get-case", flag.ExitOnError)
	onboardingID := fs.String("onboarding-id", "", "Onboarding ID (required)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *onboardingID == "" {
		return fmt.Errorf("onboarding ID is required")
	}

	session, err := cli.dm.GetCase(*onboardingID)
	if err != nil {
		return err
	}

	// Output result
	output := struct {
		OnboardingID string            `json:"onboarding_id"`
		Domain       string            `json:"domain"`
		CurrentState string            `json:"current_state"`
		Context      map[string]string `json:"context"`
		DSL          string            `json:"dsl"`
	}{
		OnboardingID: session.SessionID,
		Domain:       session.Domain,
		CurrentState: session.GetContext().CurrentState,
		Context: map[string]string{
			"investor_id":   session.GetContext().InvestorID,
			"investor_name": session.GetContext().InvestorName,
			"investor_type": session.GetContext().InvestorType,
		},
		DSL: session.GetDSL(),
	}

	return cli.outputJSON(output)
}

func (cli *DSLManagerCLI) listCases(args []string) error {
	cases := cli.dm.ListCases()

	output := struct {
		TotalCases int      `json:"total_cases"`
		CaseIDs    []string `json:"case_ids"`
	}{
		TotalCases: len(cases),
		CaseIDs:    cases,
	}

	return cli.outputJSON(output)
}

func (cli *DSLManagerCLI) outputJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func (cli *DSLManagerCLI) printUsage() error {
	usage := `DSL Manager CLI

Usage:
  dsl-manager create-case --domain=<domain> [options]
  dsl-manager update-case --onboarding-id=<id> --state=<new_state> [--dsl=<dsl_fragment>]
  dsl-manager get-case --onboarding-id=<id>
  dsl-manager list-cases

Options for create-case:
  --domain        Domain for the case (required)
  --investor-name Optional investor name
  --investor-type Optional investor type

Options for update-case:
  --onboarding-id Onboarding ID (required)
  --state         New state for the case (required)
  --dsl           DSL fragment to append (optional)

Options for get-case:
  --onboarding-id Onboarding ID (required)

Examples:
  dsl-manager create-case --domain=investor --investor-name="John Doe" --investor-type=individual
  dsl-manager update-case --onboarding-id=abc123 --state=KYC_STARTED --dsl='(kyc.start (requirements ...))'
  dsl-manager get-case --onboarding-id=abc123
  dsl-manager list-cases
`
	fmt.Println(usage)
	return nil
}

/*
This CLI implementation provides a comprehensive interface for managing DSL cases with the following key features:

1. `create-case`: Initialize a new case with optional domain-specific details
2. `update-case`: Update an existing case with a new state and optional DSL fragment
3. `get-case`: Retrieve the full details of a specific case
4. `list-cases`: List all active case IDs

Key Design Principles:
- JSON output for machine-readable results
- Flexible domain support
- Clear error handling
- Comprehensive usage instructions
- Support for stateful DSL management

The CLI uses flag parsing to handle arguments, generates domain-specific DSL fragments, and provides a clean, consistent interface for interacting with the DSL Manager.
*/
