package hfagent

import (
	"context"
	"fmt"
	"strings"
)

// ExampleScenario represents a complete example with input and expected output
type ExampleScenario struct {
	Name              string
	Description       string
	Request           DSLGenerationRequest
	ExpectedVerb      string
	ExpectedState     string
	ExpectedDSLPrefix string
}

// GetExampleScenarios returns a collection of example scenarios for testing
func GetExampleScenarios() []ExampleScenario {
	return []ExampleScenario{
		{
			Name:        "Create Opportunity - Basic",
			Description: "Create initial investor opportunity with minimal information",
			Request: DSLGenerationRequest{
				Instruction: "Create an opportunity for Acme Capital Partners LP, a corporate investor from Switzerland",
			},
			ExpectedVerb:      "investor.start-opportunity",
			ExpectedState:     "OPPORTUNITY",
			ExpectedDSLPrefix: "(investor.start-opportunity",
		},
		{
			Name:        "Create Opportunity - Detailed",
			Description: "Create opportunity with additional context",
			Request: DSLGenerationRequest{
				Instruction: "Create an opportunity for John Smith, individual investor from United States, sourced from wealth advisor referral",
			},
			ExpectedVerb:      "investor.start-opportunity",
			ExpectedState:     "OPPORTUNITY",
			ExpectedDSLPrefix: "(investor.start-opportunity",
		},
		{
			Name:        "Record Indication",
			Description: "Record investment indication with amount",
			Request: DSLGenerationRequest{
				Instruction:  "Investor wants to invest $5 million in Global Opportunities Fund Class A",
				InvestorID:   "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d",
				CurrentState: "OPPORTUNITY",
				FundID:       "f1a2b3c4-d5e6-4f5a-9b8c-7d6e5f4a3b2c",
				ClassID:      "c1d2e3f4-a5b6-4c7d-8e9f-0a1b2c3d4e5f",
			},
			ExpectedVerb:      "investor.record-indication",
			ExpectedState:     "PRECHECKS",
			ExpectedDSLPrefix: "(investor.record-indication",
		},
		{
			Name:        "Begin KYC - Standard Tier",
			Description: "Start standard KYC process",
			Request: DSLGenerationRequest{
				Instruction:  "Begin standard KYC process for this investor",
				InvestorID:   "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d",
				CurrentState: "PRECHECKS",
			},
			ExpectedVerb:      "kyc.begin",
			ExpectedState:     "KYC_PENDING",
			ExpectedDSLPrefix: "(kyc.begin",
		},
		{
			Name:        "Begin KYC - Enhanced Tier",
			Description: "Start enhanced due diligence for high-risk investor",
			Request: DSLGenerationRequest{
				Instruction:  "Start enhanced KYC due diligence for high net worth individual",
				InvestorID:   "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d",
				CurrentState: "PRECHECKS",
			},
			ExpectedVerb:      "kyc.begin",
			ExpectedState:     "KYC_PENDING",
			ExpectedDSLPrefix: "(kyc.begin",
		},
		{
			Name:        "Collect Document - Passport",
			Description: "Collect passport document",
			Request: DSLGenerationRequest{
				Instruction:  "Collect passport for John Smith",
				InvestorID:   "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d",
				CurrentState: "KYC_PENDING",
			},
			ExpectedVerb:      "kyc.collect-doc",
			ExpectedState:     "KYC_PENDING",
			ExpectedDSLPrefix: "(kyc.collect-doc",
		},
		{
			Name:        "Collect Document - Corporate",
			Description: "Collect certificate of incorporation",
			Request: DSLGenerationRequest{
				Instruction:  "Collect certificate of incorporation document",
				InvestorID:   "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d",
				CurrentState: "KYC_PENDING",
				InvestorType: "CORPORATE",
			},
			ExpectedVerb:      "kyc.collect-doc",
			ExpectedState:     "KYC_PENDING",
			ExpectedDSLPrefix: "(kyc.collect-doc",
		},
		{
			Name:        "Screen Investor",
			Description: "Run AML screening using WorldCheck",
			Request: DSLGenerationRequest{
				Instruction:  "Run AML screening using WorldCheck",
				InvestorID:   "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d",
				CurrentState: "KYC_PENDING",
			},
			ExpectedVerb:      "kyc.screen",
			ExpectedState:     "KYC_PENDING",
			ExpectedDSLPrefix: "(kyc.screen",
		},
		{
			Name:        "Approve KYC - Medium Risk",
			Description: "Approve KYC with medium risk rating",
			Request: DSLGenerationRequest{
				Instruction:  "Approve KYC with medium risk rating, refresh due in 12 months, approved by Sarah Johnson",
				InvestorID:   "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d",
				CurrentState: "KYC_PENDING",
			},
			ExpectedVerb:      "kyc.approve",
			ExpectedState:     "KYC_APPROVED",
			ExpectedDSLPrefix: "(kyc.approve",
		},
		{
			Name:        "Set Banking Instructions",
			Description: "Set USD banking details",
			Request: DSLGenerationRequest{
				Instruction:  "Set banking instructions for USD: JPMorgan Chase, account name Acme Capital LP, SWIFT CHASUS33, account 1234567890",
				InvestorID:   "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d",
				CurrentState: "KYC_APPROVED",
			},
			ExpectedVerb:      "bank.set-instruction",
			ExpectedState:     "KYC_APPROVED",
			ExpectedDSLPrefix: "(bank.set-instruction",
		},
		{
			Name:        "Capture Tax - W8BEN-E",
			Description: "Capture tax information for non-US entity",
			Request: DSLGenerationRequest{
				Instruction:  "Capture tax information: non-US person, entity classification, W8-BEN-E form",
				InvestorID:   "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d",
				CurrentState: "KYC_APPROVED",
			},
			ExpectedVerb:      "tax.capture",
			ExpectedState:     "KYC_APPROVED",
			ExpectedDSLPrefix: "(tax.capture",
		},
		{
			Name:        "Submit Subscription",
			Description: "Submit subscription request",
			Request: DSLGenerationRequest{
				Instruction:  "Submit subscription for $5 million USD, trade date 2024-02-05, settlement 2024-02-10",
				InvestorID:   "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d",
				CurrentState: "KYC_APPROVED",
				FundID:       "f1a2b3c4-d5e6-4f5a-9b8c-7d6e5f4a3b2c",
				ClassID:      "c1d2e3f4-a5b6-4c7d-8e9f-0a1b2c3d4e5f",
			},
			ExpectedVerb:      "subscribe.request",
			ExpectedState:     "SUB_PENDING_CASH",
			ExpectedDSLPrefix: "(subscribe.request",
		},
		{
			Name:        "Confirm Cash Receipt",
			Description: "Confirm subscription cash received",
			Request: DSLGenerationRequest{
				Instruction:  "Confirm cash receipt of $5 million USD on value date 2024-02-10, reference ACME-SUB-001",
				InvestorID:   "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d",
				CurrentState: "SUB_PENDING_CASH",
				AvailableData: map[string]interface{}{
					"trade_id": "t1r2a3-4567-8901-2345-678901234576",
				},
			},
			ExpectedVerb:      "cash.confirm",
			ExpectedState:     "FUNDED_PENDING_NAV",
			ExpectedDSLPrefix: "(cash.confirm",
		},
		{
			Name:        "Strike NAV",
			Description: "Set NAV for dealing date",
			Request: DSLGenerationRequest{
				Instruction:  "Set NAV for 2024-02-10 at 1250.75 per share",
				CurrentState: "FUNDED_PENDING_NAV",
				FundID:       "f1a2b3c4-d5e6-4f5a-9b8c-7d6e5f4a3b2c",
				ClassID:      "c1d2e3f4-a5b6-4c7d-8e9f-0a1b2c3d4e5f",
			},
			ExpectedVerb:      "deal.nav",
			ExpectedState:     "FUNDED_PENDING_NAV",
			ExpectedDSLPrefix: "(deal.nav",
		},
		{
			Name:        "Issue Units",
			Description: "Issue units to investor after NAV strike",
			Request: DSLGenerationRequest{
				Instruction:  "Issue 3997.6 units at NAV 1250.75 per share",
				InvestorID:   "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d",
				CurrentState: "FUNDED_PENDING_NAV",
				ClassID:      "c1d2e3f4-a5b6-4c7d-8e9f-0a1b2c3d4e5f",
				AvailableData: map[string]interface{}{
					"trade_id":  "t1r2a3-4567-8901-2345-678901234576",
					"series_id": "s1e2r3-4567-8901-2345-678901234579",
				},
			},
			ExpectedVerb:      "subscribe.issue",
			ExpectedState:     "ACTIVE",
			ExpectedDSLPrefix: "(subscribe.issue",
		},
		{
			Name:        "Request Redemption - Full",
			Description: "Request full redemption",
			Request: DSLGenerationRequest{
				Instruction:  "Request full redemption, 90 day notice from 2024-10-31, value date 2024-12-31",
				InvestorID:   "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d",
				CurrentState: "ACTIVE",
				ClassID:      "c1d2e3f4-a5b6-4c7d-8e9f-0a1b2c3d4e5f",
			},
			ExpectedVerb:      "redeem.request",
			ExpectedState:     "REDEEM_PENDING",
			ExpectedDSLPrefix: "(redeem.request",
		},
		{
			Name:        "Request Redemption - Partial",
			Description: "Request partial redemption by percentage",
			Request: DSLGenerationRequest{
				Instruction:  "Request redemption of 50% of holdings, notice date today, value date in 90 days",
				InvestorID:   "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d",
				CurrentState: "ACTIVE",
				ClassID:      "c1d2e3f4-a5b6-4c7d-8e9f-0a1b2c3d4e5f",
			},
			ExpectedVerb:      "redeem.request",
			ExpectedState:     "REDEEM_PENDING",
			ExpectedDSLPrefix: "(redeem.request",
		},
		{
			Name:        "Settle Redemption",
			Description: "Settle redemption payment",
			Request: DSLGenerationRequest{
				Instruction:  "Settle redemption payment of $5,604,828.48 on 2025-01-05, reference ACME-RED-001",
				InvestorID:   "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d",
				CurrentState: "REDEEM_PENDING",
				AvailableData: map[string]interface{}{
					"redemption_trade_id": "t2r2d3-4567-8901-2345-678901234582",
				},
			},
			ExpectedVerb:      "redeem.settle",
			ExpectedState:     "REDEEMED",
			ExpectedDSLPrefix: "(redeem.settle",
		},
		{
			Name:        "Offboard Investor",
			Description: "Close investor relationship",
			Request: DSLGenerationRequest{
				Instruction:  "Offboard investor - fully redeemed, relationship closed per client request",
				InvestorID:   "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d",
				CurrentState: "REDEEMED",
			},
			ExpectedVerb:      "offboard.close",
			ExpectedState:     "OFFBOARDED",
			ExpectedDSLPrefix: "(offboard.close",
		},
		{
			Name:        "Enable Continuous Screening",
			Description: "Set up daily continuous screening",
			Request: DSLGenerationRequest{
				Instruction:  "Enable daily continuous screening for this investor",
				InvestorID:   "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d",
				CurrentState: "KYC_APPROVED",
			},
			ExpectedVerb:      "screen.continuous",
			ExpectedState:     "KYC_APPROVED",
			ExpectedDSLPrefix: "(screen.continuous",
		},
		{
			Name:        "Set KYC Refresh Schedule",
			Description: "Schedule annual KYC refresh",
			Request: DSLGenerationRequest{
				Instruction:  "Set annual KYC refresh schedule, next refresh 2025-01-28",
				InvestorID:   "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d",
				CurrentState: "KYC_APPROVED",
			},
			ExpectedVerb:      "kyc.refresh-schedule",
			ExpectedState:     "KYC_APPROVED",
			ExpectedDSLPrefix: "(kyc.refresh-schedule",
		},
	}
}

// RunExampleScenarios executes all example scenarios and returns results
func RunExampleScenarios(ctx context.Context, agent *HedgeFundDSLAgent) ([]ScenarioResult, error) {
	scenarios := GetExampleScenarios()
	results := make([]ScenarioResult, 0, len(scenarios))

	for _, scenario := range scenarios {
		result := ScenarioResult{
			Scenario: scenario,
		}

		response, err := agent.GenerateDSL(ctx, scenario.Request)
		if err != nil {
			result.Error = err
		} else {
			result.Response = response
			result.Success = validateScenarioResult(scenario, response)
		}

		results = append(results, result)
	}

	return results, nil
}

// ScenarioResult represents the result of executing a scenario
type ScenarioResult struct {
	Scenario ExampleScenario
	Response *DSLGenerationResponse
	Success  bool
	Error    error
}

// validateScenarioResult checks if the response matches expectations
func validateScenarioResult(scenario ExampleScenario, response *DSLGenerationResponse) bool {
	if response.Verb != scenario.ExpectedVerb {
		return false
	}

	if scenario.ExpectedState != "" && response.ToState != scenario.ExpectedState {
		return false
	}

	if !strings.HasPrefix(response.DSL, scenario.ExpectedDSLPrefix) {
		return false
	}

	return true
}

// PrintScenarioResults prints formatted results
func PrintScenarioResults(results []ScenarioResult) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("HEDGE FUND DSL AGENT - SCENARIO TEST RESULTS")
	fmt.Println(strings.Repeat("=", 80))

	passed := 0
	failed := 0

	for i, result := range results {
		fmt.Printf("\n[%d] %s\n", i+1, result.Scenario.Name)
		fmt.Printf("    %s\n", result.Scenario.Description)
		fmt.Printf("    Instruction: %s\n", result.Scenario.Request.Instruction)

		if result.Error != nil {
			fmt.Printf("    ❌ ERROR: %v\n", result.Error)
			failed++
		} else if result.Success {
			fmt.Printf("    ✅ PASSED\n")
			fmt.Printf("    Generated Verb: %s\n", result.Response.Verb)
			fmt.Printf("    State Transition: %s → %s\n", result.Response.FromState, result.Response.ToState)
			fmt.Printf("    Confidence: %.2f\n", result.Response.Confidence)
			fmt.Printf("    DSL:\n%s\n", formatDSL(result.Response.DSL))
			passed++
		} else {
			fmt.Printf("    ⚠️  VALIDATION FAILED\n")
			fmt.Printf("    Expected Verb: %s, Got: %s\n", result.Scenario.ExpectedVerb, result.Response.Verb)
			fmt.Printf("    Expected State: %s, Got: %s\n", result.Scenario.ExpectedState, result.Response.ToState)
			failed++
		}

		if result.Response != nil && len(result.Response.Warnings) > 0 {
			fmt.Printf("    Warnings: %v\n", result.Response.Warnings)
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Printf("SUMMARY: %d passed, %d failed (%.1f%% success rate)\n",
		passed, failed, float64(passed)/float64(len(results))*100)
	fmt.Println(strings.Repeat("=", 80) + "\n")
}

// formatDSL adds proper indentation for display
func formatDSL(dsl string) string {
	lines := strings.Split(dsl, "\n")
	var formatted strings.Builder
	for _, line := range lines {
		formatted.WriteString("        ")
		formatted.WriteString(line)
		formatted.WriteString("\n")
	}
	return formatted.String()
}

// GetCompleteLifecycleInstructions returns instructions for a complete investor lifecycle
func GetCompleteLifecycleInstructions() []string {
	return []string{
		"Create an opportunity for Acme Capital Partners LP, a corporate investor from Switzerland",
		"Investor wants to invest $5 million in Global Opportunities Fund Class A USD",
		"Begin standard KYC process for this investor",
		"Collect certificate of incorporation document",
		"Collect partnership agreement",
		"Collect authorized signatory list",
		"Run AML screening using WorldCheck",
		"Approve KYC with medium risk rating, refresh due in 12 months, approved by Sarah Johnson",
		"Set annual KYC refresh schedule, next refresh date one year from today",
		"Enable daily continuous screening",
		"Capture tax information: non-US person, entity classification, W8-BEN-E form",
		"Set banking instructions for USD: JPMorgan Chase, account name Acme Capital LP, SWIFT CHASUS33",
		"Submit subscription for $5 million USD, trade date 2024-02-05, value date 2024-02-10",
		"Confirm cash receipt of $5 million USD on value date 2024-02-10",
		"Set NAV for 2024-02-10 at 1250.75 per share",
		"Issue 3997.6 units at NAV 1250.75 per share",
		"Request full redemption, 90 day notice from 2024-10-31, value date 2024-12-31",
		"Set NAV for 2024-12-31 at 1402.30 per share",
		"Settle redemption payment of $5,604,828.48 on 2025-01-05",
		"Offboard investor - fully redeemed, relationship closed per client request",
	}
}

// ConversationalExample demonstrates natural language interactions
type ConversationalExample struct {
	UserInput          string
	Context            DSLGenerationRequest
	ExpectedOperation  string
	ExpectedConfidence float64
}

// GetConversationalExamples returns examples of natural language inputs
func GetConversationalExamples() []ConversationalExample {
	return []ConversationalExample{
		{
			UserInput:          "I want to onboard a new Swiss investor named Acme Capital",
			ExpectedOperation:  "investor.start-opportunity",
			ExpectedConfidence: 0.9,
		},
		{
			UserInput:          "They're interested in putting in a million dollars",
			ExpectedOperation:  "investor.record-indication",
			ExpectedConfidence: 0.85,
		},
		{
			UserInput:          "Let's start the KYC process",
			ExpectedOperation:  "kyc.begin",
			ExpectedConfidence: 0.95,
		},
		{
			UserInput:          "Can you collect their incorporation certificate?",
			ExpectedOperation:  "kyc.collect-doc",
			ExpectedConfidence: 0.9,
		},
		{
			UserInput:          "Run them through WorldCheck",
			ExpectedOperation:  "kyc.screen",
			ExpectedConfidence: 0.95,
		},
		{
			UserInput:          "Everything looks good, approve them as medium risk",
			ExpectedOperation:  "kyc.approve",
			ExpectedConfidence: 0.9,
		},
		{
			UserInput:          "Set them up for daily screening",
			ExpectedOperation:  "screen.continuous",
			ExpectedConfidence: 0.9,
		},
		{
			UserInput:          "Get their banking details: JPMorgan, USD account",
			ExpectedOperation:  "bank.set-instruction",
			ExpectedConfidence: 0.85,
		},
		{
			UserInput:          "They want to subscribe for 5 million",
			ExpectedOperation:  "subscribe.request",
			ExpectedConfidence: 0.9,
		},
		{
			UserInput:          "Cash is in",
			ExpectedOperation:  "cash.confirm",
			ExpectedConfidence: 0.85,
		},
		{
			UserInput:          "NAV is 1250.75",
			ExpectedOperation:  "deal.nav",
			ExpectedConfidence: 0.9,
		},
		{
			UserInput:          "Allocate the shares",
			ExpectedOperation:  "subscribe.issue",
			ExpectedConfidence: 0.85,
		},
		{
			UserInput:          "They want to redeem everything",
			ExpectedOperation:  "redeem.request",
			ExpectedConfidence: 0.9,
		},
		{
			UserInput:          "Payment sent",
			ExpectedOperation:  "redeem.settle",
			ExpectedConfidence: 0.85,
		},
		{
			UserInput:          "Close their account",
			ExpectedOperation:  "offboard.close",
			ExpectedConfidence: 0.9,
		},
	}
}
