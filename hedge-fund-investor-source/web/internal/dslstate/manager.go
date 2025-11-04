package dslstate

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/google/uuid"

	hfagent "dsl-ob-poc/hedge-fund-investor-source/web/internal/hf-agent"
)

// OperationIntent represents what the agent wants to do (high-level intent)
// The agent tells us: "create opportunity with name=X, type=Y"
// NOT the precise DSL S-expression syntax
type OperationIntent struct {
	Operation   string                 // High-level: "create_opportunity", "start_kyc", "collect_document"
	Verb        string                 // DSL verb: "investor.start-opportunity", "kyc.begin", "kyc.collect-doc"
	Parameters  map[string]interface{} // Parameter values
	FromState   string                 // Expected current state
	ToState     string                 // Target state after operation
	Explanation string                 // Human-readable explanation
	Confidence  float64                // Agent confidence in this operation
	Warnings    []string               // Any warnings or issues
}

// DSLAgent is an interface for agents that understand natural language
// and return high-level operation intent (not DSL syntax)
type DSLAgent interface {
	// DecideOperation interprets natural language and returns what operation to perform
	// Returns: operation intent (what to do), not DSL text (how to express it)
	DecideOperation(ctx context.Context, instruction string, currentDSL string, context Context) (*OperationIntent, error)
	Close() error
}

// Manager is the single source of truth for DSL state management.
// It encapsulates all logic for DSL accumulation, placeholder resolution, and context tracking.
type Manager struct {
	// CurrentDSL is the accumulated DSL document (the "state")
	CurrentDSL string

	// Context tracks entities and state across operations
	Context Context

	// Agent generates change requests (intent) from natural language
	agent DSLAgent
}

// Context holds the session context for entity references and state
type Context struct {
	// Entity IDs
	InvestorID string
	FundID     string
	ClassID    string
	SeriesID   string

	// Entity attributes
	InvestorName string
	InvestorType string
	Domicile     string

	// State machine
	CurrentState string

	// Additional metadata
	Metadata map[string]interface{}
}

// NewManager creates a new DSL state manager
func NewManager(agent DSLAgent) *Manager {
	return &Manager{
		CurrentDSL: "",
		Context: Context{
			Metadata: make(map[string]interface{}),
		},
		agent: agent,
	}
}

// AppendOperation is the SINGLE PLACE that updates DSL state.
// Flow:
// 1. Ask agent what operation to perform (agent returns INTENT: "create opportunity with X,Y,Z")
// 2. DSL Manager generates the actual DSL S-expression text from intent
// 3. Extract and track entity references (generate UUIDs as needed)
// 4. Parse and validate the generated DSL
// 5. Accumulate into complete DSL document
// 6. Return updated DSL state
func (m *Manager) AppendOperation(ctx context.Context, instruction string) (completeDSL string, fragment string, response *hfagent.DSLGenerationResponse, err error) {
	log.Printf("[DSLStateManager] Processing instruction: %s", instruction)
	log.Printf("[DSLStateManager] Current state: %s, InvestorID: %s", m.Context.CurrentState, m.Context.InvestorID)

	// 1. Ask agent what operation to perform
	// Agent returns: "I want to create_opportunity with name=X, type=Y" (INTENT)
	// Agent does NOT return DSL S-expression syntax
	intent, err := m.agent.DecideOperation(ctx, instruction, m.CurrentDSL, m.Context)
	if err != nil {
		return "", "", nil, fmt.Errorf("agent decision failed: %w", err)
	}

	log.Printf("[DSLStateManager] Agent intent: %s (verb: %s, params: %v)", intent.Operation, intent.Verb, intent.Parameters)
	log.Printf("[DSLStateManager] State transition: %s → %s", intent.FromState, intent.ToState)

	// Convert intent to response format for backward compatibility
	response = &hfagent.DSLGenerationResponse{
		Verb:        intent.Verb,
		Parameters:  intent.Parameters,
		FromState:   intent.FromState,
		ToState:     intent.ToState,
		Explanation: intent.Explanation,
		Confidence:  intent.Confidence,
		Warnings:    intent.Warnings,
	}

	// 2. Extract and track entity references BEFORE generating DSL
	//    Generate UUIDs for new entities (investor, fund, etc.)
	m.extractAndTrackEntities(intent)

	// 3. Generate the actual DSL S-expression text from intent
	//    DSL Manager owns the syntax, agent owns the semantics
	dslFragment := m.generateDSLFromIntent(intent)
	log.Printf("[DSLStateManager] Generated DSL: %s", dslFragment)

	// 4. Parse and validate the generated DSL
	if err := m.validateDSL(dslFragment, intent.Verb); err != nil {
		return "", "", nil, fmt.Errorf("DSL validation failed: %w", err)
	}
	log.Printf("[DSLStateManager] DSL validated successfully")

	// 5. Update state transition
	if intent.ToState != "" {
		m.Context.CurrentState = intent.ToState
		log.Printf("[DSLStateManager] State updated: → %s", m.Context.CurrentState)
	}

	// 6. Accumulate into complete DSL document (the state)
	if m.CurrentDSL == "" {
		m.CurrentDSL = dslFragment
	} else {
		m.CurrentDSL = m.CurrentDSL + "\n\n" + dslFragment
	}

	log.Printf("[DSLStateManager] Complete DSL: %d chars, %d operations", len(m.CurrentDSL), strings.Count(m.CurrentDSL, "\n\n")+1)

	// 7. Return complete DSL (the state), the fragment, and response metadata
	return m.CurrentDSL, dslFragment, response, nil
}

// generateDSLFromIntent creates the actual DSL S-expression text from operation intent
// THIS IS THE SINGLE PLACE WHERE DSL SYNTAX IS GENERATED
// Input: High-level intent ("create opportunity with name, type, domicile")
// Output: S-expression DSL text "(investor.start-opportunity :name "X" :type "Y")"
func (m *Manager) generateDSLFromIntent(intent *OperationIntent) string {
	var b strings.Builder

	verb := intent.Verb
	params := intent.Parameters

	b.WriteString("(")
	b.WriteString(verb)

	// Generate DSL based on verb type
	switch verb {
	case "investor.start-opportunity":
		if legalName, ok := params["legal-name"].(string); ok {
			b.WriteString("\n  :legal-name \"")
			b.WriteString(legalName)
			b.WriteString("\"")
		}
		if investorType, ok := params["type"].(string); ok {
			b.WriteString("\n  :type \"")
			b.WriteString(investorType)
			b.WriteString("\"")
		}
		if domicile, ok := params["domicile"].(string); ok {
			b.WriteString("\n  :domicile \"")
			b.WriteString(domicile)
			b.WriteString("\"")
		}

	case "kyc.begin":
		// Use actual investor UUID from context
		if m.Context.InvestorID != "" {
			b.WriteString("\n  :investor \"")
			b.WriteString(m.Context.InvestorID)
			b.WriteString("\"")
		}
		if tier, ok := params["tier"].(string); ok {
			b.WriteString("\n  :tier \"")
			b.WriteString(tier)
			b.WriteString("\"")
		}

	case "kyc.collect-doc":
		if m.Context.InvestorID != "" {
			b.WriteString("\n  :investor \"")
			b.WriteString(m.Context.InvestorID)
			b.WriteString("\"")
		}
		if docType, ok := params["doc-type"].(string); ok {
			b.WriteString("\n  :doc-type \"")
			b.WriteString(docType)
			b.WriteString("\"")
		}
		if subject, ok := params["subject"].(string); ok {
			b.WriteString("\n  :subject \"")
			b.WriteString(subject)
			b.WriteString("\"")
		}

	case "kyc.screen":
		if m.Context.InvestorID != "" {
			b.WriteString("\n  :investor \"")
			b.WriteString(m.Context.InvestorID)
			b.WriteString("\"")
		}
		if provider, ok := params["provider"].(string); ok {
			b.WriteString("\n  :provider \"")
			b.WriteString(provider)
			b.WriteString("\"")
		}

	case "kyc.approve":
		if m.Context.InvestorID != "" {
			b.WriteString("\n  :investor \"")
			b.WriteString(m.Context.InvestorID)
			b.WriteString("\"")
		}
		if risk, ok := params["risk"].(string); ok {
			b.WriteString("\n  :risk \"")
			b.WriteString(risk)
			b.WriteString("\"")
		}
		if refreshDue, ok := params["refresh-due"].(string); ok {
			b.WriteString("\n  :refresh-due \"")
			b.WriteString(refreshDue)
			b.WriteString("\"")
		}
		if approvedBy, ok := params["approved-by"].(string); ok {
			b.WriteString("\n  :approved-by \"")
			b.WriteString(approvedBy)
			b.WriteString("\"")
		}

	default:
		// Generic parameter handling for other verbs
		for key, val := range params {
			// Use actual UUIDs from context for known entity references
			if key == "investor" && m.Context.InvestorID != "" {
				b.WriteString("\n  :")
				b.WriteString(key)
				b.WriteString(" \"")
				b.WriteString(m.Context.InvestorID)
				b.WriteString("\"")
			} else if key == "fund" && m.Context.FundID != "" {
				b.WriteString("\n  :")
				b.WriteString(key)
				b.WriteString(" \"")
				b.WriteString(m.Context.FundID)
				b.WriteString("\"")
			} else if valStr, ok := val.(string); ok {
				b.WriteString("\n  :")
				b.WriteString(key)
				b.WriteString(" \"")
				b.WriteString(valStr)
				b.WriteString("\"")
			} else if valNum, ok := val.(float64); ok {
				b.WriteString("\n  :")
				b.WriteString(key)
				b.WriteString(" ")
				b.WriteString(fmt.Sprintf("%.2f", valNum))
			}
		}
	}

	b.WriteString(")")

	return b.String()
}

// extractAndTrackEntities updates context from operation intent
// Generates UUIDs for new entities and tracks them in context
func (m *Manager) extractAndTrackEntities(intent *OperationIntent) {
	// Handle new investor creation
	if intent.Verb == "investor.start-opportunity" {
		// Extract investor attributes from intent parameters
		if legalName, ok := intent.Parameters["legal-name"].(string); ok {
			m.Context.InvestorName = legalName
		}
		if investorType, ok := intent.Parameters["type"].(string); ok {
			m.Context.InvestorType = investorType
		}
		if domicile, ok := intent.Parameters["domicile"].(string); ok {
			m.Context.Domicile = domicile
		}

		// Generate UUID for new investor if not already set
		if m.Context.InvestorID == "" {
			m.Context.InvestorID = uuid.New().String()
			log.Printf("[DSLStateManager] Generated investor UUID: %s for %s", m.Context.InvestorID, m.Context.InvestorName)
		}

		// Store UUID back in intent parameters so DSL generation can use it
		intent.Parameters["investor"] = m.Context.InvestorID
	}

	// Track investor ID from any operation
	if investorID, ok := intent.Parameters["investor"].(string); ok {
		// Only update if it's an actual UUID, not a placeholder
		if !strings.HasPrefix(investorID, "<") {
			m.Context.InvestorID = investorID
		}
	}

	// Track fund/class/series context
	if fundID, ok := intent.Parameters["fund"].(string); ok {
		if !strings.HasPrefix(fundID, "<") {
			m.Context.FundID = fundID
		}
	}
	if classID, ok := intent.Parameters["class"].(string); ok {
		if !strings.HasPrefix(classID, "<") {
			m.Context.ClassID = classID
		}
	}
	if seriesID, ok := intent.Parameters["series"].(string); ok {
		if !strings.HasPrefix(seriesID, "<") {
			m.Context.SeriesID = seriesID
		}
	}
}

// GetCurrentDSL returns the current accumulated DSL (read-only access to state)
func (m *Manager) GetCurrentDSL() string {
	return m.CurrentDSL
}

// GetContext returns the current context (read-only)
func (m *Manager) GetContext() Context {
	return m.Context
}

// Reset clears the DSL state and context (for new session)
func (m *Manager) Reset() {
	m.CurrentDSL = ""
	m.Context = Context{
		Metadata: make(map[string]interface{}),
	}
	log.Printf("[DSLStateManager] State reset")
}

// validateDSL validates the generated DSL S-expression
func (m *Manager) validateDSL(dsl string, expectedVerb string) error {
	// 1. Check basic S-expression structure
	if !strings.HasPrefix(strings.TrimSpace(dsl), "(") {
		return fmt.Errorf("DSL must start with '('")
	}
	if !strings.HasSuffix(strings.TrimSpace(dsl), ")") {
		return fmt.Errorf("DSL must end with ')'")
	}

	// 2. Check that the verb is present
	if !strings.Contains(dsl, expectedVerb) {
		return fmt.Errorf("DSL does not contain expected verb: %s", expectedVerb)
	}

	// 3. Check for unresolved placeholders
	placeholderPattern := regexp.MustCompile(`<[a-z_]+>`)
	if matches := placeholderPattern.FindAllString(dsl, -1); len(matches) > 0 {
		return fmt.Errorf("DSL contains unresolved placeholders: %v", matches)
	}

	// 4. Validate approved verbs only
	approvedVerbs := map[string]bool{
		"investor.start-opportunity": true,
		"investor.record-indication": true,
		"investor.amend-details":     true,
		"kyc.begin":                  true,
		"kyc.collect-doc":            true,
		"kyc.screen":                 true,
		"kyc.approve":                true,
		"kyc.refresh-schedule":       true,
		"screen.continuous":          true,
		"tax.capture":                true,
		"bank.set-instruction":       true,
		"subscribe.request":          true,
		"cash.confirm":               true,
		"deal.nav":                   true,
		"subscribe.issue":            true,
		"redeem.request":             true,
		"redeem.settle":              true,
		"offboard.close":             true,
	}

	if !approvedVerbs[expectedVerb] {
		return fmt.Errorf("verb not in approved vocabulary: %s", expectedVerb)
	}

	// 5. Check balanced parentheses
	openCount := strings.Count(dsl, "(")
	closeCount := strings.Count(dsl, ")")
	if openCount != closeCount {
		return fmt.Errorf("unbalanced parentheses: %d open, %d close", openCount, closeCount)
	}

	return nil
}
