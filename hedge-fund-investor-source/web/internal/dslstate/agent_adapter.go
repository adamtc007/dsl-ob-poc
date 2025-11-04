package dslstate

import (
	"context"

	hfagent "dsl-ob-poc/hedge-fund-investor-source/web/internal/hf-agent"
)

// HedgeFundAgentAdapter adapts the HedgeFundDSLAgent to implement DSLAgent interface
// This adapter translates between:
// - Old interface: Agent generates DSL text directly
// - New interface: Agent returns operation intent, DSL Manager generates text
type HedgeFundAgentAdapter struct {
	agent *hfagent.HedgeFundDSLAgent
}

// NewHedgeFundAgentAdapter creates an adapter for the hedge fund agent
func NewHedgeFundAgentAdapter(agent *hfagent.HedgeFundDSLAgent) *HedgeFundAgentAdapter {
	return &HedgeFundAgentAdapter{
		agent: agent,
	}
}

// DecideOperation implements DSLAgent interface
// Translates natural language instruction into operation intent
func (a *HedgeFundAgentAdapter) DecideOperation(ctx context.Context, instruction string, currentDSL string, context Context) (*OperationIntent, error) {
	// Build request for the underlying agent
	request := hfagent.DSLGenerationRequest{
		Instruction:   instruction,
		InvestorID:    context.InvestorID,
		CurrentState:  context.CurrentState,
		InvestorType:  context.InvestorType,
		InvestorName:  context.InvestorName,
		Domicile:      context.Domicile,
		FundID:        context.FundID,
		ClassID:       context.ClassID,
		SeriesID:      context.SeriesID,
		ExistingDSL:   currentDSL,
		AvailableData: context.Metadata,
	}

	// Call underlying agent (still generates DSL for now, but we extract intent)
	response, err := a.agent.GenerateDSL(ctx, request)
	if err != nil {
		return nil, err
	}

	// Extract operation intent from response
	// The agent gives us the verb and parameters - that's the intent
	// We ignore the DSL field since DSL Manager will generate it
	intent := &OperationIntent{
		Operation:   mapVerbToOperation(response.Verb),
		Verb:        response.Verb,
		Parameters:  response.Parameters,
		FromState:   response.FromState,
		ToState:     response.ToState,
		Explanation: response.Explanation,
		Confidence:  response.Confidence,
		Warnings:    response.Warnings,
	}

	return intent, nil
}

// Close implements DSLAgent interface
func (a *HedgeFundAgentAdapter) Close() error {
	if a.agent != nil {
		return a.agent.Close()
	}
	return nil
}

// mapVerbToOperation converts DSL verb to high-level operation name
// This provides a more intuitive operation name for logging/debugging
func mapVerbToOperation(verb string) string {
	operationMap := map[string]string{
		"investor.start-opportunity": "create_opportunity",
		"investor.record-indication": "record_indication",
		"investor.amend-details":     "amend_investor",
		"kyc.begin":                  "start_kyc",
		"kyc.collect-doc":            "collect_document",
		"kyc.screen":                 "screen_investor",
		"kyc.approve":                "approve_kyc",
		"kyc.refresh-schedule":       "schedule_kyc_refresh",
		"screen.continuous":          "enable_continuous_screening",
		"tax.capture":                "capture_tax_info",
		"bank.set-instruction":       "set_bank_details",
		"subscribe.request":          "request_subscription",
		"cash.confirm":               "confirm_cash",
		"deal.nav":                   "deal_nav",
		"subscribe.issue":            "issue_shares",
		"redeem.request":             "request_redemption",
		"redeem.settle":              "settle_redemption",
		"offboard.close":             "close_offboarding",
	}

	if op, ok := operationMap[verb]; ok {
		return op
	}
	return verb // Fallback to verb name
}
