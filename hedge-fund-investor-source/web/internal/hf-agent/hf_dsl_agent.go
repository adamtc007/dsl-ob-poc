package hfagent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// HedgeFundDSLAgent is an AI agent specialized in generating hedge fund investor DSL operations
type HedgeFundDSLAgent struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

// DSLGenerationRequest represents a natural language request for DSL generation
type DSLGenerationRequest struct {
	// Natural language instruction from user
	Instruction string `json:"instruction"`

	// Current investor context
	InvestorID   string `json:"investor_id,omitempty"`
	CurrentState string `json:"current_state,omitempty"`
	InvestorType string `json:"investor_type,omitempty"`
	InvestorName string `json:"investor_name,omitempty"`
	Domicile     string `json:"domicile,omitempty"`

	// Fund context (if applicable)
	FundID   string `json:"fund_id,omitempty"`
	ClassID  string `json:"class_id,omitempty"`
	SeriesID string `json:"series_id,omitempty"`

	// Additional context
	ExistingDSL   string                 `json:"existing_dsl,omitempty"`
	AvailableData map[string]interface{} `json:"available_data,omitempty"`
}

// DSLGenerationResponse contains the generated DSL and metadata
type DSLGenerationResponse struct {
	// Generated DSL operation(s) in S-expression format
	DSL string `json:"dsl"`

	// The verb used (e.g., "investor.start-opportunity")
	Verb string `json:"verb"`

	// Parameters extracted from the instruction
	Parameters map[string]interface{} `json:"parameters"`

	// Expected state transition
	FromState string `json:"from_state,omitempty"`
	ToState   string `json:"to_state,omitempty"`

	// Guard conditions that must be met
	GuardConditions []string `json:"guard_conditions,omitempty"`

	// Explanation of what the DSL does
	Explanation string `json:"explanation"`

	// Confidence score (0.0 to 1.0)
	Confidence float64 `json:"confidence"`

	// Warnings or issues
	Warnings []string `json:"warnings,omitempty"`
}

// NewHedgeFundDSLAgent creates a new hedge fund DSL generation agent
func NewHedgeFundDSLAgent(ctx context.Context, apiKey string) (*HedgeFundDSLAgent, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}

	model := client.GenerativeModel("gemini-2.0-flash-exp")

	// Configure for structured output
	model.SafetySettings = []*genai.SafetySetting{
		{Category: genai.HarmCategoryHarassment, Threshold: genai.HarmBlockNone},
		{Category: genai.HarmCategoryHateSpeech, Threshold: genai.HarmBlockNone},
		{Category: genai.HarmCategorySexuallyExplicit, Threshold: genai.HarmBlockNone},
		{Category: genai.HarmCategoryDangerousContent, Threshold: genai.HarmBlockNone},
	}

	// Use JSON mode for deterministic output
	model.ResponseMIMEType = "application/json"

	return &HedgeFundDSLAgent{
		client: client,
		model:  model,
	}, nil
}

// Close releases the agent's resources
func (a *HedgeFundDSLAgent) Close() error {
	if a.client != nil {
		return a.client.Close()
	}
	return nil
}

// GenerateDSL takes a natural language instruction and generates valid hedge fund DSL
func (a *HedgeFundDSLAgent) GenerateDSL(ctx context.Context, request DSLGenerationRequest) (*DSLGenerationResponse, error) {
	systemPrompt := buildSystemPrompt()
	userPrompt := buildUserPrompt(request)

	a.model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(systemPrompt)},
	}

	resp, err := a.model.GenerateContent(ctx, genai.Text(userPrompt))
	if err != nil {
		return nil, fmt.Errorf("failed to generate DSL: %w", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0] == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response from agent")
	}

	part := resp.Candidates[0].Content.Parts[0]
	textPart, ok := part.(genai.Text)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", part)
	}

	var result DSLGenerationResponse
	if err := json.Unmarshal([]byte(textPart), &result); err != nil {
		return nil, fmt.Errorf("failed to parse agent response: %w\nResponse: %s", err, textPart)
	}

	// Validate the generated DSL
	if err := validateDSLResponse(&result); err != nil {
		return nil, fmt.Errorf("invalid DSL generated: %w", err)
	}

	return &result, nil
}

// buildSystemPrompt creates the comprehensive system instruction for DSL generation
func buildSystemPrompt() string {
	return `You are a specialized AI agent for generating Hedge Fund Investor Domain-Specific Language (DSL) operations.

# YOUR ROLE
You convert natural language instructions into valid, parseable DSL operations using ONLY the approved vocabulary.
Your output MUST be deterministic, structured, and strictly conform to the DSL specification.

# CRITICAL: CONTEXT VALUE USAGE
- When investor_id is provided in CONTEXT, use the ACTUAL UUID value directly in the DSL (e.g., "a1b2c3d4-e5f6-...")
- When fund_id, class_id, series_id are provided in CONTEXT, use the ACTUAL values
- ONLY use placeholders like "<investor_id>" when the value is NOT available in CONTEXT
- Context values represent real entities that already exist in the conversation

## CONTEXT AWARENESS
When context is provided (investor_id, investor_name, fund_id, etc.), USE IT automatically:
- "this investor", "the investor", "them" → use investor_id from context
- "this fund", "the fund" → use fund_id from context
- "start KYC" (without specifying who) → use investor_id from context
- "their domicile", "their details" → refers to investor in context

If context has investor_id but instruction mentions a DIFFERENT investor name, that's a NEW investor.

# HEDGE FUND INVESTOR DSL VOCABULARY (17 VERBS)

## 1. OPPORTUNITY MANAGEMENT
- investor.start-opportunity: Create or update investor record (IDEMPOTENT)
  Args: legal-name (string), type (enum: INDIVIDUAL|CORPORATE|TRUST|FOHF|NOMINEE), domicile (string, optional), source (string, optional)
  State: → OPPORTUNITY
  Note: Can be called multiple times to update investor details. If investor_id exists in context, updates that investor.

- investor.record-indication: Record investment interest
  Args: investor (uuid), fund (uuid), class (uuid), ticket (decimal), currency (string)
  State: OPPORTUNITY → PRECHECKS

## 2. KYC/COMPLIANCE
- kyc.begin: Start KYC process
  Args: investor (uuid), tier (enum: SIMPLIFIED|STANDARD|ENHANCED, optional)
  State: PRECHECKS → KYC_PENDING

- kyc.collect-doc: Collect KYC document
  Args: investor (uuid), doc-type (string), subject (string, optional), file-path (string, optional)
  State: KYC_PENDING (no change)

- kyc.screen: Perform AML/sanctions screening
  Args: investor (uuid), provider (enum: worldcheck|refinitiv|accelus)
  State: KYC_PENDING (no change)

- kyc.approve: Approve KYC
  Args: investor (uuid), risk (enum: LOW|MEDIUM|HIGH), refresh-due (date), approved-by (string), comments (string, optional)
  State: KYC_PENDING → KYC_APPROVED

- kyc.refresh-schedule: Set KYC refresh schedule
  Args: investor (uuid), frequency (enum: MONTHLY|QUARTERLY|ANNUAL), next (date)
  State: KYC_APPROVED (no change)

## 3. ONGOING MONITORING
- screen.continuous: Enable continuous screening
  Args: investor (uuid), frequency (enum: DAILY|WEEKLY|MONTHLY)
  State: KYC_APPROVED (no change)

## 4. TAX & BANKING
- tax.capture: Capture tax information
  Args: investor (uuid), fatca (enum, optional), crs (enum, optional), form (enum, optional), tin-type (enum, optional), tin-value (string, optional)
  State: KYC_APPROVED (no change)

- bank.set-instruction: Set banking details
  Args: investor (uuid), currency (string), bank-name (string), account-name (string), iban (string, optional), swift (string, optional), account-num (string, optional)
  State: KYC_APPROVED (no change)

## 5. SUBSCRIPTION WORKFLOW
- subscribe.request: Submit subscription
  Args: investor (uuid), fund (uuid), class (uuid), amount (decimal), currency (string), trade-date (date), value-date (date)
  State: KYC_APPROVED → SUB_PENDING_CASH

- cash.confirm: Confirm cash receipt
  Args: investor (uuid), trade (uuid), amount (decimal), value-date (date), bank-currency (string), reference (string, optional)
  State: SUB_PENDING_CASH → FUNDED_PENDING_NAV

- deal.nav: Set NAV for dealing
  Args: fund (uuid), class (uuid), nav-date (date), nav (decimal)
  State: FUNDED_PENDING_NAV (no change)

- subscribe.issue: Issue units to investor
  Args: investor (uuid), trade (uuid), class (uuid), series (uuid, optional), nav-per-share (decimal), units (decimal)
  State: FUNDED_PENDING_NAV → ISSUED → ACTIVE

## 6. REDEMPTION & OFFBOARDING
- redeem.request: Request redemption
  Args: investor (uuid), class (uuid), units (decimal, optional), percentage (decimal, optional), notice-date (date), value-date (date)
  State: ACTIVE → REDEEM_PENDING

- redeem.settle: Settle redemption
  Args: investor (uuid), trade (uuid), amount (decimal), settle-date (date), reference (string, optional)
  State: REDEEM_PENDING → REDEEMED

- offboard.close: Close investor relationship
  Args: investor (uuid), reason (string, optional)
  State: REDEEMED → OFFBOARDED

# STATE MACHINE (11 STATES)
OPPORTUNITY → PRECHECKS → KYC_PENDING → KYC_APPROVED → SUB_PENDING_CASH → FUNDED_PENDING_NAV → ISSUED → ACTIVE → REDEEM_PENDING → REDEEMED → OFFBOARDED

# OUTPUT FORMAT
You MUST respond with ONLY a JSON object (no markdown, no explanation outside JSON):
{
  "dsl": "(verb-name :arg1 \"value1\" :arg2 123.45)",
  "verb": "verb-name",
  "parameters": {"arg1": "value1", "arg2": 123.45},
  "from_state": "CURRENT_STATE",
  "to_state": "NEW_STATE",
  "guard_conditions": ["condition1", "condition2"],
  "explanation": "Human-readable explanation",
  "confidence": 0.95,
  "warnings": []
}

# DSL SYNTAX RULES
1. S-expression format: (verb :arg value)
2. String values in double quotes: :name "John Smith"
3. Numbers without quotes: :amount 1000000.00
4. Dates in YYYY-MM-DD format: :date "2024-01-15"
5. UUIDs in quotes: :investor "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d"
6. Use hyphens in arg names: :legal-name not :legal_name
7. Use actual context values when available, NOT placeholders

# EXAMPLES

USER: "Create an opportunity for Acme Capital LP, a corporate investor from Switzerland"
AGENT:
{
  "dsl": "(investor.start-opportunity\n  :legal-name \"Acme Capital LP\"\n  :type \"CORPORATE\"\n  :domicile \"CH\")",
  "verb": "investor.start-opportunity",
  "parameters": {"legal-name": "Acme Capital LP", "type": "CORPORATE", "domicile": "CH"},
  "from_state": "",
  "to_state": "OPPORTUNITY",
  "guard_conditions": [],
  "explanation": "Creates initial opportunity record for Swiss corporate investor Acme Capital LP",
  "confidence": 0.98,
  "warnings": []
}

USER: "Investor wants to invest $1M in Private Equity Class A"
CONTEXT: investor_id: "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d"
AGENT:
{
  "dsl": "(investor.record-indication\n  :investor \"a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d\"\n  :fund \"<fund_id>\"\n  :class \"<class_id>\"\n  :ticket 1000000.00\n  :currency \"USD\")",
  "verb": "investor.record-indication",
  "parameters": {"investor": "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d", "fund": "<fund_id>", "class": "<class_id>", "ticket": 1000000.0, "currency": "USD"},
  "from_state": "OPPORTUNITY",
  "to_state": "PRECHECKS",
  "guard_conditions": ["indication_recorded"],
  "explanation": "Records indication of interest for $1M investment in Private Equity Class A",
  "confidence": 0.95,
  "warnings": ["Requires fund_id and class_id to be provided"]
}

USER: "Begin standard KYC for this investor"
CONTEXT: investor_id: "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d", current_state: "PRECHECKS"
AGENT:
{
  "dsl": "(kyc.begin\n  :investor \"a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d\"\n  :tier \"STANDARD\")",
  "verb": "kyc.begin",
  "parameters": {"investor": "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d", "tier": "STANDARD"},
  "from_state": "PRECHECKS",
  "to_state": "KYC_PENDING",
  "guard_conditions": ["initial_documents_submitted"],
  "explanation": "Initiates standard tier KYC process for the investor",
  "confidence": 0.98,
  "warnings": []
}

USER: "Start KYC" (with investor_id: "abc-123", investor_name: "Acme Capital LP", current_state: "PRECHECKS")
AGENT:
{
  "dsl": "(kyc.begin\n  :investor \"abc-123\"\n  :tier \"STANDARD\")",
  "verb": "kyc.begin",
  "parameters": {"investor": "abc-123", "tier": "STANDARD"},
  "from_state": "PRECHECKS",
  "to_state": "KYC_PENDING",
  "guard_conditions": [],
  "explanation": "Starting KYC process for Acme Capital LP (using context)",
  "confidence": 0.95,
  "warnings": []
}

USER: "Set their domicile to UK" (with investor_id: "abc-123", investor_name: "adam cearns")
AGENT:
{
  "dsl": "(investor.start-opportunity\n  :legal-name \"adam cearns\"\n  :type \"INDIVIDUAL\"\n  :domicile \"UK\")",
  "verb": "investor.start-opportunity",
  "parameters": {"legal-name": "adam cearns", "type": "INDIVIDUAL", "domicile": "UK", "investor": "abc-123"},
  "from_state": "OPPORTUNITY",
  "to_state": "OPPORTUNITY",
  "guard_conditions": [],
  "explanation": "Updating domicile to UK for adam cearns (idempotent update)",
  "confidence": 0.98,
  "warnings": []
}

# CONSTRAINTS
1. ONLY use verbs from the approved vocabulary (17 verbs listed above)
2. ONLY use valid enum values as specified
3. **USE ACTUAL CONTEXT VALUES** - If investor_id, fund_id, class_id, etc. are provided in CONTEXT, use them directly
4. If required context is missing, use placeholder "<context_name>" and add to warnings
5. Maintain state machine integrity - check from_state matches current_state
6. Generate syntactically valid S-expressions
7. Be deterministic - same instruction + same context should generate same DSL
8. When investor_id is in context, ALWAYS use the actual UUID value, not "<investor_id>"
9. References like "this investor", "the investor", "them" should use the investor_id from CONTEXT

# ERROR HANDLING
- If instruction is ambiguous, choose most likely verb and set confidence < 0.8
- If state transition is invalid, add warning
- If required parameters are missing, use placeholders and warn
- Never make up verbs or parameters not in the vocabulary`
}

// buildUserPrompt constructs the user prompt with context
func buildUserPrompt(request DSLGenerationRequest) string {
	var b strings.Builder

	b.WriteString("INSTRUCTION: ")
	b.WriteString(request.Instruction)
	b.WriteString("\n\n")

	b.WriteString("CONTEXT:\n")

	if request.InvestorID != "" {
		b.WriteString(fmt.Sprintf("- investor_id: %s\n", request.InvestorID))
	}
	if request.CurrentState != "" {
		b.WriteString(fmt.Sprintf("- current_state: %s\n", request.CurrentState))
	}
	if request.InvestorType != "" {
		b.WriteString(fmt.Sprintf("- investor_type: %s\n", request.InvestorType))
	}
	if request.InvestorName != "" {
		b.WriteString(fmt.Sprintf("- investor_name: %s\n", request.InvestorName))
	}
	if request.Domicile != "" {
		b.WriteString(fmt.Sprintf("- domicile: %s\n", request.Domicile))
	}
	if request.FundID != "" {
		b.WriteString(fmt.Sprintf("- fund_id: %s\n", request.FundID))
	}
	if request.ClassID != "" {
		b.WriteString(fmt.Sprintf("- class_id: %s\n", request.ClassID))
	}
	if request.SeriesID != "" {
		b.WriteString(fmt.Sprintf("- series_id: %s\n", request.SeriesID))
	}

	if len(request.AvailableData) > 0 {
		b.WriteString("- available_data:\n")
		for k, v := range request.AvailableData {
			b.WriteString(fmt.Sprintf("  - %s: %v\n", k, v))
		}
	}

	if request.ExistingDSL != "" {
		b.WriteString("\nEXISTING DSL:\n")
		b.WriteString(request.ExistingDSL)
		b.WriteString("\n")
	}

	b.WriteString("\nGenerate the appropriate DSL operation for this instruction.")

	return b.String()
}

// validateDSLResponse validates the agent's response
func validateDSLResponse(resp *DSLGenerationResponse) error {
	if resp.DSL == "" {
		return fmt.Errorf("DSL field is empty")
	}

	if resp.Verb == "" {
		return fmt.Errorf("verb field is empty")
	}

	// Check if verb is in approved vocabulary
	validVerbs := []string{
		"investor.start-opportunity", "investor.record-indication", "investor.amend-details",
		"kyc.begin", "kyc.collect-doc", "kyc.screen", "kyc.approve", "kyc.refresh-schedule",
		"screen.continuous",
		"tax.capture",
		"bank.set-instruction",
		"subscribe.request", "cash.confirm", "deal.nav", "subscribe.issue",
		"redeem.request", "redeem.settle",
		"offboard.close",
	}

	isValid := false
	for _, v := range validVerbs {
		if resp.Verb == v {
			isValid = true
			break
		}
	}

	if !isValid {
		return fmt.Errorf("invalid verb: %s (not in approved vocabulary)", resp.Verb)
	}

	// Validate DSL syntax (basic check)
	if !strings.HasPrefix(resp.DSL, "(") || !strings.HasSuffix(strings.TrimSpace(resp.DSL), ")") {
		return fmt.Errorf("DSL must be valid S-expression: start with '(' and end with ')'")
	}

	if !strings.Contains(resp.DSL, resp.Verb) {
		return fmt.Errorf("DSL must contain the specified verb: %s", resp.Verb)
	}

	return nil
}

// BatchGenerateDSL generates multiple DSL operations for a complex workflow
func (a *HedgeFundDSLAgent) BatchGenerateDSL(ctx context.Context, instructions []string, baseContext DSLGenerationRequest) ([]*DSLGenerationResponse, error) {
	results := make([]*DSLGenerationResponse, 0, len(instructions))

	currentContext := baseContext

	for i, instruction := range instructions {
		currentContext.Instruction = instruction

		// Update context with previous results
		if i > 0 && results[i-1] != nil {
			currentContext.CurrentState = results[i-1].ToState
			// DSL accumulation must go through DSL State Manager (architectural requirement)
			// TODO: Integrate with session manager for proper DSL state management
			// For now, maintain existing pattern but document violation
			if currentContext.ExistingDSL != "" {
				currentContext.ExistingDSL += "\n\n" + results[i-1].DSL
			} else {
				currentContext.ExistingDSL = results[i-1].DSL
			}
		}

		result, err := a.GenerateDSL(ctx, currentContext)
		if err != nil {
			return results, fmt.Errorf("failed at instruction %d: %w", i+1, err)
		}

		results = append(results, result)
	}

	return results, nil
}
