// Package hedgefundinvestor provides the hedge fund investor domain implementation
// for the multi-domain DSL system.
//
// This domain handles the complete hedge fund investor lifecycle from opportunity
// creation through offboarding, including KYC, compliance, subscription, and
// redemption workflows. It implements 17 specialized verbs across 6 categories
// and manages an 11-state progression through the investor journey.
//
// Key Features:
// - Complete investor lifecycle management (OPPORTUNITY → OFFBOARDED)
// - KYC and compliance workflows with multiple tiers and providers
// - Subscription and redemption processing with cash and NAV handling
// - Tax and banking information capture
// - Continuous monitoring and refresh schedules
// - Integration with AI agents for natural language DSL generation
//
// State Machine (11 states):
// OPPORTUNITY → PRECHECKS → KYC_PENDING → KYC_APPROVED → SUB_PENDING_CASH →
// FUNDED_PENDING_NAV → ISSUED → ACTIVE → REDEEM_PENDING → REDEEMED → OFFBOARDED
package hedgefundinvestor

import (
	"context"
	"fmt"
	"strings"
	"time"

	registry "dsl-ob-poc/internal/domain-registry"
)

// Domain implements the Domain interface for hedge fund investor workflows
type Domain struct {
	name        string
	version     string
	description string
	vocabulary  *registry.Vocabulary
	healthy     bool
	metrics     *registry.DomainMetrics
	createdAt   time.Time
}

// NewDomain creates a new hedge fund investor domain
func NewDomain() *Domain {
	domain := &Domain{
		name:        "hedge-fund-investor",
		version:     "1.0.0",
		description: "Hedge fund investor lifecycle management from opportunity to offboarding",
		healthy:     true,
		createdAt:   time.Now(),
		metrics: &registry.DomainMetrics{
			TotalRequests:      0,
			SuccessfulRequests: 0,
			FailedRequests:     0,
			TotalVerbs:         17,
			ActiveVerbs:        17,
			UnusedVerbs:        0,
			StateTransitions:   make(map[string]int64),
			CurrentStates:      make(map[string]int64),
			ValidationErrors:   make(map[string]int64),
			GenerationErrors:   make(map[string]int64),
			IsHealthy:          true,
			LastHealthCheck:    time.Now(),
			UptimeSeconds:      0,
			MemoryUsageBytes:   2 * 1024 * 1024, // 2MB
			CollectedAt:        time.Now(),
			Version:            "1.0.0",
		},
	}

	domain.vocabulary = domain.buildVocabulary()
	return domain
}

// Domain interface implementation

func (d *Domain) Name() string                        { return d.name }
func (d *Domain) Version() string                     { return d.version }
func (d *Domain) Description() string                 { return d.description }
func (d *Domain) GetVocabulary() *registry.Vocabulary { return d.vocabulary }
func (d *Domain) IsHealthy() bool                     { return d.healthy }
func (d *Domain) GetMetrics() *registry.DomainMetrics { return d.metrics }

func (d *Domain) GetValidStates() []string {
	return []string{
		"OPPORTUNITY", "PRECHECKS", "KYC_PENDING", "KYC_APPROVED",
		"SUB_PENDING_CASH", "FUNDED_PENDING_NAV", "ISSUED", "ACTIVE",
		"REDEEM_PENDING", "REDEEMED", "OFFBOARDED",
	}
}

func (d *Domain) GetInitialState() string {
	return "OPPORTUNITY"
}

// ValidateVerbs checks that the DSL only uses approved hedge fund verbs
func (d *Domain) ValidateVerbs(dsl string) error {
	if strings.TrimSpace(dsl) == "" {
		return fmt.Errorf("empty DSL")
	}

	// Extract verbs from DSL using regex pattern matching for multi-line DSL
	foundVerbs := make(map[string]bool)

	// Look for pattern: (verb.name at the start of line or after whitespace
	lines := strings.Split(dsl, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "(") {
			// Extract verb: (verb.name -> verb.name
			verb := strings.TrimPrefix(line, "(")

			// Handle multi-line format - verb might be on its own line
			if strings.Contains(verb, ".") {
				// Find the end of the verb name
				if spaceIdx := strings.Index(verb, " "); spaceIdx > 0 {
					verb = verb[:spaceIdx]
				} else if newlineIdx := strings.Index(verb, "\n"); newlineIdx > 0 {
					verb = verb[:newlineIdx]
				} else {
					// Verb is on its own line, use the whole trimmed content
					verb = strings.TrimSpace(verb)
				}

				if verb != "" && strings.Contains(verb, ".") {
					foundVerbs[verb] = true
				}
			}
		}
	}

	if len(foundVerbs) == 0 {
		return fmt.Errorf("no valid hedge fund verbs found in DSL")
	}

	// Check each found verb against approved vocabulary
	for verb := range foundVerbs {
		if _, exists := d.vocabulary.Verbs[verb]; !exists {
			return fmt.Errorf("invalid hedge fund verb: %s", verb)
		}
	}

	return nil
}

// ValidateStateTransition checks if a state transition is valid
func (d *Domain) ValidateStateTransition(from, to string) error {
	validTransitions := map[string][]string{
		"OPPORTUNITY":        {"PRECHECKS"},
		"PRECHECKS":          {"KYC_PENDING"},
		"KYC_PENDING":        {"KYC_APPROVED"},
		"KYC_APPROVED":       {"SUB_PENDING_CASH"},
		"SUB_PENDING_CASH":   {"FUNDED_PENDING_NAV"},
		"FUNDED_PENDING_NAV": {"ISSUED"},
		"ISSUED":             {"ACTIVE"},
		"ACTIVE":             {"REDEEM_PENDING"},
		"REDEEM_PENDING":     {"REDEEMED"},
		"REDEEMED":           {"OFFBOARDED"},
		"OFFBOARDED":         {}, // Terminal state
	}

	validNext, exists := validTransitions[from]
	if !exists {
		return fmt.Errorf("invalid source state: %s", from)
	}

	for _, validTo := range validNext {
		if validTo == to {
			return nil
		}
	}

	return fmt.Errorf("invalid state transition from %s to %s", from, to)
}

// GenerateDSL generates DSL from natural language instructions (simplified implementation)
func (d *Domain) GenerateDSL(ctx context.Context, req *registry.GenerationRequest) (*registry.GenerationResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("generation request cannot be nil")
	}

	instruction := strings.ToLower(strings.TrimSpace(req.Instruction))

	// Simple pattern matching for demonstration (in real implementation, this would call AI agent)
	var dsl, verb, explanation, toState string
	var confidence float64 = 0.8
	var warnings []string

	// Extract investor ID from context if available
	investorID := ""
	if req.Context != nil {
		if id, ok := req.Context["investor_id"].(string); ok {
			investorID = id
		}
	}

	switch {
	case strings.Contains(instruction, "start opportunity") || strings.Contains(instruction, "create investor"):
		if name := extractName(instruction); name != "" {
			verb = "investor.start-opportunity"
			dsl = fmt.Sprintf("(investor.start-opportunity\n  :legal-name \"%s\"\n  :type \"INDIVIDUAL\")", name)
			explanation = fmt.Sprintf("Creating opportunity for investor %s", name)
			toState = "OPPORTUNITY"
			confidence = 0.9
		}

	case strings.Contains(instruction, "begin kyc") || strings.Contains(instruction, "start kyc"):
		if investorID != "" {
			verb = "kyc.begin"
			dsl = fmt.Sprintf("(kyc.begin\n  :investor \"%s\"\n  :tier \"STANDARD\")", investorID)
			explanation = "Starting standard KYC process"
			toState = "KYC_PENDING"
			confidence = 0.9
		} else {
			warnings = append(warnings, "investor_id required for KYC operations")
		}

	case strings.Contains(instruction, "approve kyc"):
		if investorID != "" {
			verb = "kyc.approve"
			dsl = fmt.Sprintf("(kyc.approve\n  :investor \"%s\"\n  :risk \"MEDIUM\"\n  :refresh-due \"2025-01-01\"\n  :approved-by \"system\")", investorID)
			explanation = "Approving KYC with medium risk rating"
			toState = "KYC_APPROVED"
			confidence = 0.85
		} else {
			warnings = append(warnings, "investor_id required for KYC operations")
		}

	case strings.Contains(instruction, "subscribe") || strings.Contains(instruction, "subscription"):
		if investorID != "" {
			verb = "subscribe.request"
			dsl = fmt.Sprintf("(subscribe.request\n  :investor \"%s\"\n  :fund \"<fund_id>\"\n  :class \"<class_id>\"\n  :amount 1000000.00\n  :currency \"USD\"\n  :trade-date \"2024-01-15\"\n  :value-date \"2024-01-15\")", investorID)
			explanation = "Submitting subscription request"
			toState = "SUB_PENDING_CASH"
			warnings = append(warnings, "fund_id and class_id required")
		} else {
			warnings = append(warnings, "investor_id required for subscription operations")
		}

	default:
		return nil, fmt.Errorf("unsupported instruction: %s", req.Instruction)
	}

	if dsl == "" {
		return nil, fmt.Errorf("unable to generate DSL for instruction: %s", req.Instruction)
	}

	parameters := make(map[string]interface{})
	if investorID != "" {
		parameters["investor"] = investorID
	}

	return &registry.GenerationResponse{
		DSL:            dsl,
		Verb:           verb,
		Parameters:     parameters,
		ToState:        toState,
		IsValid:        true,
		Confidence:     confidence,
		Explanation:    explanation,
		Warnings:       warnings,
		GenerationTime: time.Since(req.Timestamp),
		RequestID:      req.RequestID,
		Timestamp:      time.Now(),
	}, nil
}

// GetCurrentState determines the current state from context
func (d *Domain) GetCurrentState(context map[string]interface{}) (string, error) {
	if context == nil {
		return d.GetInitialState(), nil
	}

	if state, exists := context["current_state"]; exists {
		if stateStr, ok := state.(string); ok {
			// Validate state
			for _, validState := range d.GetValidStates() {
				if validState == stateStr {
					return stateStr, nil
				}
			}
			return "", fmt.Errorf("invalid hedge fund state: %s", stateStr)
		}
	}

	return d.GetInitialState(), nil
}

// ExtractContext extracts domain-specific context from DSL
func (d *Domain) ExtractContext(dsl string) (map[string]interface{}, error) {
	context := make(map[string]interface{})

	if strings.TrimSpace(dsl) == "" {
		return context, nil
	}

	// Extract investor ID if present
	if strings.Contains(dsl, ":investor") {
		// Simple regex-like extraction for demo
		lines := strings.Split(dsl, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.Contains(line, ":investor") {
				parts := strings.Fields(line)
				for i, part := range parts {
					if part == ":investor" && i+1 < len(parts) {
						investorID := strings.Trim(parts[i+1], "\"")
						if investorID != "" && investorID != "<investor_id>" {
							context["investor_id"] = investorID
						}
						break
					}
				}
			}
		}
	}

	// Infer state from verb (prioritize later verbs in DSL for multi-line)
	if strings.Contains(dsl, "subscribe.request") {
		context["current_state"] = "SUB_PENDING_CASH"
	} else if strings.Contains(dsl, "kyc.approve") {
		context["current_state"] = "KYC_APPROVED"
	} else if strings.Contains(dsl, "kyc.begin") {
		context["current_state"] = "KYC_PENDING"
	} else if strings.Contains(dsl, "investor.record-indication") {
		context["current_state"] = "PRECHECKS"
	} else if strings.Contains(dsl, "investor.start-opportunity") {
		context["current_state"] = "OPPORTUNITY"
	}

	return context, nil
}

// Helper functions

func extractName(instruction string) string {
	// Simple name extraction for demonstration
	words := strings.Fields(instruction)
	for i, word := range words {
		if (word == "for" || word == "investor") && i+1 < len(words) {
			// Take next 1-2 words as name
			if i+2 < len(words) {
				return words[i+1] + " " + words[i+2]
			}
			return words[i+1]
		}
	}
	return ""
}

// buildVocabulary constructs the complete hedge fund investor vocabulary
func (d *Domain) buildVocabulary() *registry.Vocabulary {
	vocab := &registry.Vocabulary{
		Domain:      d.name,
		Version:     d.version,
		Description: d.description,
		Verbs:       make(map[string]*registry.VerbDefinition),
		Categories:  make(map[string]*registry.VerbCategory),
		States:      d.GetValidStates(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Define argument specs that are commonly used
	investorArg := &registry.ArgumentSpec{
		Name:        "investor",
		Type:        registry.ArgumentTypeUUID,
		Required:    true,
		Description: "Investor UUID",
	}

	legalNameArg := &registry.ArgumentSpec{
		Name:        "legal-name",
		Type:        registry.ArgumentTypeString,
		Required:    true,
		Description: "Investor legal name",
		MinLength:   &[]int{1}[0],
		MaxLength:   &[]int{200}[0],
	}

	investorTypeArg := &registry.ArgumentSpec{
		Name:        "type",
		Type:        registry.ArgumentTypeEnum,
		Required:    true,
		Description: "Investor type",
		EnumValues:  []string{"INDIVIDUAL", "CORPORATE", "TRUST", "FOHF", "NOMINEE"},
	}

	// 1. OPPORTUNITY MANAGEMENT VERBS

	vocab.Verbs["investor.start-opportunity"] = &registry.VerbDefinition{
		Name:        "investor.start-opportunity",
		Category:    "opportunity",
		Version:     "1.0.0",
		Description: "Create or update investor opportunity record (idempotent)",
		Arguments: map[string]*registry.ArgumentSpec{
			"legal-name": legalNameArg,
			"type":       investorTypeArg,
			"domicile": {
				Name:        "domicile",
				Type:        registry.ArgumentTypeString,
				Required:    false,
				Description: "Investor domicile (ISO country code)",
				Pattern:     "^[A-Z]{2}$",
			},
			"source": {
				Name:        "source",
				Type:        registry.ArgumentTypeString,
				Required:    false,
				Description: "Lead source",
			},
		},
		StateTransition: &registry.StateTransition{
			ToState: "OPPORTUNITY",
		},
		Idempotent: true,
		Examples:   []string{`(investor.start-opportunity :legal-name "Acme Capital LP" :type "CORPORATE" :domicile "CH")`},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	vocab.Verbs["investor.record-indication"] = &registry.VerbDefinition{
		Name:        "investor.record-indication",
		Category:    "opportunity",
		Version:     "1.0.0",
		Description: "Record investment interest indication",
		Arguments: map[string]*registry.ArgumentSpec{
			"investor": investorArg,
			"fund": {
				Name:        "fund",
				Type:        registry.ArgumentTypeUUID,
				Required:    true,
				Description: "Fund UUID",
			},
			"class": {
				Name:        "class",
				Type:        registry.ArgumentTypeUUID,
				Required:    true,
				Description: "Share class UUID",
			},
			"ticket": {
				Name:        "ticket",
				Type:        registry.ArgumentTypeDecimal,
				Required:    true,
				Description: "Investment amount",
				MinValue:    &[]float64{0.01}[0],
			},
			"currency": {
				Name:        "currency",
				Type:        registry.ArgumentTypeString,
				Required:    true,
				Description: "Currency code (ISO 4217)",
				Pattern:     "^[A-Z]{3}$",
			},
		},
		StateTransition: &registry.StateTransition{
			FromStates: []string{"OPPORTUNITY"},
			ToState:    "PRECHECKS",
		},
		Examples:  []string{`(investor.record-indication :investor "uuid" :fund "fund-uuid" :class "class-uuid" :ticket 1000000.00 :currency "USD")`},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 2. KYC/COMPLIANCE VERBS

	vocab.Verbs["kyc.begin"] = &registry.VerbDefinition{
		Name:        "kyc.begin",
		Category:    "kyc",
		Version:     "1.0.0",
		Description: "Start KYC process",
		Arguments: map[string]*registry.ArgumentSpec{
			"investor": investorArg,
			"tier": {
				Name:         "tier",
				Type:         registry.ArgumentTypeEnum,
				Required:     false,
				Description:  "KYC tier level",
				EnumValues:   []string{"SIMPLIFIED", "STANDARD", "ENHANCED"},
				DefaultValue: "STANDARD",
			},
		},
		StateTransition: &registry.StateTransition{
			FromStates: []string{"PRECHECKS"},
			ToState:    "KYC_PENDING",
		},
		Examples:  []string{`(kyc.begin :investor "uuid" :tier "STANDARD")`},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	vocab.Verbs["kyc.collect-doc"] = &registry.VerbDefinition{
		Name:        "kyc.collect-doc",
		Category:    "kyc",
		Version:     "1.0.0",
		Description: "Collect KYC document",
		Arguments: map[string]*registry.ArgumentSpec{
			"investor": investorArg,
			"doc-type": {
				Name:        "doc-type",
				Type:        registry.ArgumentTypeString,
				Required:    true,
				Description: "Document type identifier",
			},
			"subject": {
				Name:        "subject",
				Type:        registry.ArgumentTypeString,
				Required:    false,
				Description: "Document subject/description",
			},
			"file-path": {
				Name:        "file-path",
				Type:        registry.ArgumentTypeString,
				Required:    false,
				Description: "File path or reference",
			},
		},
		StateTransition: &registry.StateTransition{
			FromStates: []string{"KYC_PENDING"},
			ToState:    "KYC_PENDING", // No state change
		},
		Examples:  []string{`(kyc.collect-doc :investor "uuid" :doc-type "passport" :subject "John Smith Passport")`},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	vocab.Verbs["kyc.screen"] = &registry.VerbDefinition{
		Name:        "kyc.screen",
		Category:    "kyc",
		Version:     "1.0.0",
		Description: "Perform AML/sanctions screening",
		Arguments: map[string]*registry.ArgumentSpec{
			"investor": investorArg,
			"provider": {
				Name:        "provider",
				Type:        registry.ArgumentTypeEnum,
				Required:    true,
				Description: "Screening provider",
				EnumValues:  []string{"worldcheck", "refinitiv", "accelus"},
			},
		},
		StateTransition: &registry.StateTransition{
			FromStates: []string{"KYC_PENDING"},
			ToState:    "KYC_PENDING", // No state change
		},
		Examples:  []string{`(kyc.screen :investor "uuid" :provider "worldcheck")`},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	vocab.Verbs["kyc.approve"] = &registry.VerbDefinition{
		Name:        "kyc.approve",
		Category:    "kyc",
		Version:     "1.0.0",
		Description: "Approve KYC completion",
		Arguments: map[string]*registry.ArgumentSpec{
			"investor": investorArg,
			"risk": {
				Name:        "risk",
				Type:        registry.ArgumentTypeEnum,
				Required:    true,
				Description: "Risk rating",
				EnumValues:  []string{"LOW", "MEDIUM", "HIGH"},
			},
			"refresh-due": {
				Name:        "refresh-due",
				Type:        registry.ArgumentTypeDate,
				Required:    true,
				Description: "Next KYC refresh date",
			},
			"approved-by": {
				Name:        "approved-by",
				Type:        registry.ArgumentTypeString,
				Required:    true,
				Description: "Approver identifier",
			},
			"comments": {
				Name:        "comments",
				Type:        registry.ArgumentTypeString,
				Required:    false,
				Description: "Approval comments",
			},
		},
		StateTransition: &registry.StateTransition{
			FromStates: []string{"KYC_PENDING"},
			ToState:    "KYC_APPROVED",
		},
		Examples:  []string{`(kyc.approve :investor "uuid" :risk "MEDIUM" :refresh-due "2025-01-01" :approved-by "john.doe")`},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	vocab.Verbs["kyc.refresh-schedule"] = &registry.VerbDefinition{
		Name:        "kyc.refresh-schedule",
		Category:    "kyc",
		Version:     "1.0.0",
		Description: "Set KYC refresh schedule",
		Arguments: map[string]*registry.ArgumentSpec{
			"investor": investorArg,
			"frequency": {
				Name:        "frequency",
				Type:        registry.ArgumentTypeEnum,
				Required:    true,
				Description: "Refresh frequency",
				EnumValues:  []string{"MONTHLY", "QUARTERLY", "ANNUAL"},
			},
			"next": {
				Name:        "next",
				Type:        registry.ArgumentTypeDate,
				Required:    true,
				Description: "Next refresh date",
			},
		},
		StateTransition: &registry.StateTransition{
			FromStates: []string{"KYC_APPROVED"},
			ToState:    "KYC_APPROVED", // No state change
		},
		Examples:  []string{`(kyc.refresh-schedule :investor "uuid" :frequency "ANNUAL" :next "2025-01-01")`},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 3. ONGOING MONITORING VERBS

	vocab.Verbs["screen.continuous"] = &registry.VerbDefinition{
		Name:        "screen.continuous",
		Category:    "monitoring",
		Version:     "1.0.0",
		Description: "Enable continuous screening",
		Arguments: map[string]*registry.ArgumentSpec{
			"investor": investorArg,
			"frequency": {
				Name:        "frequency",
				Type:        registry.ArgumentTypeEnum,
				Required:    true,
				Description: "Screening frequency",
				EnumValues:  []string{"DAILY", "WEEKLY", "MONTHLY"},
			},
		},
		StateTransition: &registry.StateTransition{
			FromStates: []string{"KYC_APPROVED"},
			ToState:    "KYC_APPROVED", // No state change
		},
		Examples:  []string{`(screen.continuous :investor "uuid" :frequency "DAILY")`},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 4. TAX & BANKING VERBS

	vocab.Verbs["tax.capture"] = &registry.VerbDefinition{
		Name:        "tax.capture",
		Category:    "tax-banking",
		Version:     "1.0.0",
		Description: "Capture tax information",
		Arguments: map[string]*registry.ArgumentSpec{
			"investor": investorArg,
			"fatca": {
				Name:        "fatca",
				Type:        registry.ArgumentTypeEnum,
				Required:    false,
				Description: "FATCA status",
				EnumValues:  []string{"US_PERSON", "NON_US_PERSON", "UNKNOWN"},
			},
			"crs": {
				Name:        "crs",
				Type:        registry.ArgumentTypeEnum,
				Required:    false,
				Description: "CRS status",
				EnumValues:  []string{"REPORTABLE", "NON_REPORTABLE", "UNKNOWN"},
			},
			"form": {
				Name:        "form",
				Type:        registry.ArgumentTypeEnum,
				Required:    false,
				Description: "Tax form type",
				EnumValues:  []string{"W9", "W8BEN", "W8BEN_E", "OTHER"},
			},
			"tin-type": {
				Name:        "tin-type",
				Type:        registry.ArgumentTypeEnum,
				Required:    false,
				Description: "Tax identification number type",
				EnumValues:  []string{"SSN", "EIN", "FOREIGN", "OTHER"},
			},
			"tin-value": {
				Name:        "tin-value",
				Type:        registry.ArgumentTypeString,
				Required:    false,
				Description: "Tax identification number",
			},
		},
		StateTransition: &registry.StateTransition{
			FromStates: []string{"KYC_APPROVED"},
			ToState:    "KYC_APPROVED", // No state change
		},
		Examples:  []string{`(tax.capture :investor "uuid" :fatca "NON_US_PERSON" :form "W8BEN")`},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	vocab.Verbs["bank.set-instruction"] = &registry.VerbDefinition{
		Name:        "bank.set-instruction",
		Category:    "tax-banking",
		Version:     "1.0.0",
		Description: "Set banking details",
		Arguments: map[string]*registry.ArgumentSpec{
			"investor": investorArg,
			"currency": {
				Name:        "currency",
				Type:        registry.ArgumentTypeString,
				Required:    true,
				Description: "Banking currency (ISO 4217)",
				Pattern:     "^[A-Z]{3}$",
			},
			"bank-name": {
				Name:        "bank-name",
				Type:        registry.ArgumentTypeString,
				Required:    true,
				Description: "Bank name",
			},
			"account-name": {
				Name:        "account-name",
				Type:        registry.ArgumentTypeString,
				Required:    true,
				Description: "Account holder name",
			},
			"iban": {
				Name:        "iban",
				Type:        registry.ArgumentTypeString,
				Required:    false,
				Description: "IBAN number",
			},
			"swift": {
				Name:        "swift",
				Type:        registry.ArgumentTypeString,
				Required:    false,
				Description: "SWIFT/BIC code",
			},
			"account-num": {
				Name:        "account-num",
				Type:        registry.ArgumentTypeString,
				Required:    false,
				Description: "Account number",
			},
		},
		StateTransition: &registry.StateTransition{
			FromStates: []string{"KYC_APPROVED"},
			ToState:    "KYC_APPROVED", // No state change
		},
		Examples:  []string{`(bank.set-instruction :investor "uuid" :currency "USD" :bank-name "Chase Bank" :account-name "Acme Capital LP")`},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 5. SUBSCRIPTION WORKFLOW VERBS

	vocab.Verbs["subscribe.request"] = &registry.VerbDefinition{
		Name:        "subscribe.request",
		Category:    "subscription",
		Version:     "1.0.0",
		Description: "Submit subscription request",
		Arguments: map[string]*registry.ArgumentSpec{
			"investor": investorArg,
			"fund": {
				Name:        "fund",
				Type:        registry.ArgumentTypeUUID,
				Required:    true,
				Description: "Fund UUID",
			},
			"class": {
				Name:        "class",
				Type:        registry.ArgumentTypeUUID,
				Required:    true,
				Description: "Share class UUID",
			},
			"amount": {
				Name:        "amount",
				Type:        registry.ArgumentTypeDecimal,
				Required:    true,
				Description: "Subscription amount",
				MinValue:    &[]float64{0.01}[0],
			},
			"currency": {
				Name:        "currency",
				Type:        registry.ArgumentTypeString,
				Required:    true,
				Description: "Currency code (ISO 4217)",
				Pattern:     "^[A-Z]{3}$",
			},
			"trade-date": {
				Name:        "trade-date",
				Type:        registry.ArgumentTypeDate,
				Required:    true,
				Description: "Trade date",
			},
			"value-date": {
				Name:        "value-date",
				Type:        registry.ArgumentTypeDate,
				Required:    true,
				Description: "Value date",
			},
		},
		StateTransition: &registry.StateTransition{
			FromStates: []string{"KYC_APPROVED"},
			ToState:    "SUB_PENDING_CASH",
		},
		Examples:  []string{`(subscribe.request :investor "uuid" :fund "fund-uuid" :class "class-uuid" :amount 1000000.00 :currency "USD" :trade-date "2024-01-15" :value-date "2024-01-15")`},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	vocab.Verbs["cash.confirm"] = &registry.VerbDefinition{
		Name:        "cash.confirm",
		Category:    "subscription",
		Version:     "1.0.0",
		Description: "Confirm cash receipt",
		Arguments: map[string]*registry.ArgumentSpec{
			"investor": investorArg,
			"trade": {
				Name:        "trade",
				Type:        registry.ArgumentTypeUUID,
				Required:    true,
				Description: "Trade UUID",
			},
			"amount": {
				Name:        "amount",
				Type:        registry.ArgumentTypeDecimal,
				Required:    true,
				Description: "Confirmed cash amount",
				MinValue:    &[]float64{0.01}[0],
			},
			"value-date": {
				Name:        "value-date",
				Type:        registry.ArgumentTypeDate,
				Required:    true,
				Description: "Value date",
			},
			"bank-currency": {
				Name:        "bank-currency",
				Type:        registry.ArgumentTypeString,
				Required:    true,
				Description: "Banking currency (ISO 4217)",
				Pattern:     "^[A-Z]{3}$",
			},
			"reference": {
				Name:        "reference",
				Type:        registry.ArgumentTypeString,
				Required:    false,
				Description: "Bank reference or transaction ID",
			},
		},
		StateTransition: &registry.StateTransition{
			FromStates: []string{"SUB_PENDING_CASH"},
			ToState:    "FUNDED_PENDING_NAV",
		},
		Examples:  []string{`(cash.confirm :investor "uuid" :trade "trade-uuid" :amount 1000000.00 :value-date "2024-01-15" :bank-currency "USD")`},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	vocab.Verbs["deal.nav"] = &registry.VerbDefinition{
		Name:        "deal.nav",
		Category:    "subscription",
		Version:     "1.0.0",
		Description: "Set NAV for dealing",
		Arguments: map[string]*registry.ArgumentSpec{
			"fund": {
				Name:        "fund",
				Type:        registry.ArgumentTypeUUID,
				Required:    true,
				Description: "Fund UUID",
			},
			"class": {
				Name:        "class",
				Type:        registry.ArgumentTypeUUID,
				Required:    true,
				Description: "Share class UUID",
			},
			"nav-date": {
				Name:        "nav-date",
				Type:        registry.ArgumentTypeDate,
				Required:    true,
				Description: "NAV date",
			},
			"nav": {
				Name:        "nav",
				Type:        registry.ArgumentTypeDecimal,
				Required:    true,
				Description: "Net Asset Value per share",
				MinValue:    &[]float64{0.01}[0],
			},
		},
		StateTransition: &registry.StateTransition{
			FromStates: []string{"FUNDED_PENDING_NAV"},
			ToState:    "FUNDED_PENDING_NAV", // No state change
		},
		Examples:  []string{`(deal.nav :fund "fund-uuid" :class "class-uuid" :nav-date "2024-01-15" :nav 100.50)`},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	vocab.Verbs["subscribe.issue"] = &registry.VerbDefinition{
		Name:        "subscribe.issue",
		Category:    "subscription",
		Version:     "1.0.0",
		Description: "Issue units to investor",
		Arguments: map[string]*registry.ArgumentSpec{
			"investor": investorArg,
			"trade": {
				Name:        "trade",
				Type:        registry.ArgumentTypeUUID,
				Required:    true,
				Description: "Trade UUID",
			},
			"class": {
				Name:        "class",
				Type:        registry.ArgumentTypeUUID,
				Required:    true,
				Description: "Share class UUID",
			},
			"series": {
				Name:        "series",
				Type:        registry.ArgumentTypeUUID,
				Required:    false,
				Description: "Series UUID (if applicable)",
			},
			"nav-per-share": {
				Name:        "nav-per-share",
				Type:        registry.ArgumentTypeDecimal,
				Required:    true,
				Description: "NAV per share at issuance",
				MinValue:    &[]float64{0.01}[0],
			},
			"units": {
				Name:        "units",
				Type:        registry.ArgumentTypeDecimal,
				Required:    true,
				Description: "Number of units issued",
				MinValue:    &[]float64{0.000001}[0],
			},
		},
		StateTransition: &registry.StateTransition{
			FromStates: []string{"FUNDED_PENDING_NAV"},
			ToState:    "ACTIVE",
		},
		Examples:  []string{`(subscribe.issue :investor "uuid" :trade "trade-uuid" :class "class-uuid" :nav-per-share 100.50 :units 9950.25)`},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 6. REDEMPTION & OFFBOARDING VERBS

	vocab.Verbs["redeem.request"] = &registry.VerbDefinition{
		Name:        "redeem.request",
		Category:    "redemption",
		Version:     "1.0.0",
		Description: "Request redemption",
		Arguments: map[string]*registry.ArgumentSpec{
			"investor": investorArg,
			"class": {
				Name:        "class",
				Type:        registry.ArgumentTypeUUID,
				Required:    true,
				Description: "Share class UUID",
			},
			"units": {
				Name:        "units",
				Type:        registry.ArgumentTypeDecimal,
				Required:    false,
				Description: "Specific number of units to redeem",
				MinValue:    &[]float64{0.000001}[0],
			},
			"percentage": {
				Name:        "percentage",
				Type:        registry.ArgumentTypeDecimal,
				Required:    false,
				Description: "Percentage of holding to redeem",
				MinValue:    &[]float64{0.01}[0],
				MaxValue:    &[]float64{100.0}[0],
			},
			"notice-date": {
				Name:        "notice-date",
				Type:        registry.ArgumentTypeDate,
				Required:    true,
				Description: "Redemption notice date",
			},
			"value-date": {
				Name:        "value-date",
				Type:        registry.ArgumentTypeDate,
				Required:    true,
				Description: "Redemption value date",
			},
		},
		StateTransition: &registry.StateTransition{
			FromStates: []string{"ACTIVE"},
			ToState:    "REDEEM_PENDING",
		},
		GuardConditions: []string{"units_or_percentage_specified"},
		Examples:        []string{`(redeem.request :investor "uuid" :class "class-uuid" :percentage 50.0 :notice-date "2024-01-01" :value-date "2024-01-15")`},
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	vocab.Verbs["redeem.settle"] = &registry.VerbDefinition{
		Name:        "redeem.settle",
		Category:    "redemption",
		Version:     "1.0.0",
		Description: "Settle redemption",
		Arguments: map[string]*registry.ArgumentSpec{
			"investor": investorArg,
			"trade": {
				Name:        "trade",
				Type:        registry.ArgumentTypeUUID,
				Required:    true,
				Description: "Redemption trade UUID",
			},
			"amount": {
				Name:        "amount",
				Type:        registry.ArgumentTypeDecimal,
				Required:    true,
				Description: "Settlement amount",
				MinValue:    &[]float64{0.01}[0],
			},
			"settle-date": {
				Name:        "settle-date",
				Type:        registry.ArgumentTypeDate,
				Required:    true,
				Description: "Settlement date",
			},
			"reference": {
				Name:        "reference",
				Type:        registry.ArgumentTypeString,
				Required:    false,
				Description: "Settlement reference or transaction ID",
			},
		},
		StateTransition: &registry.StateTransition{
			FromStates: []string{"REDEEM_PENDING"},
			ToState:    "REDEEMED",
		},
		Examples:  []string{`(redeem.settle :investor "uuid" :trade "trade-uuid" :amount 500000.00 :settle-date "2024-01-20")`},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	vocab.Verbs["offboard.close"] = &registry.VerbDefinition{
		Name:        "offboard.close",
		Category:    "offboarding",
		Version:     "1.0.0",
		Description: "Close investor relationship",
		Arguments: map[string]*registry.ArgumentSpec{
			"investor": investorArg,
			"reason": {
				Name:        "reason",
				Type:        registry.ArgumentTypeString,
				Required:    false,
				Description: "Reason for closing relationship",
			},
		},
		StateTransition: &registry.StateTransition{
			FromStates: []string{"REDEEMED"},
			ToState:    "OFFBOARDED",
		},
		Examples:  []string{`(offboard.close :investor "uuid" :reason "Full redemption completed")`},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Define categories
	vocab.Categories["opportunity"] = &registry.VerbCategory{
		Name:        "opportunity",
		Description: "Opportunity management and investor indication",
		Verbs:       []string{"investor.start-opportunity", "investor.record-indication"},
		Color:       "#4CAF50",
		Icon:        "👤",
	}

	vocab.Categories["kyc"] = &registry.VerbCategory{
		Name:        "kyc",
		Description: "Know Your Customer and compliance processes",
		Verbs:       []string{"kyc.begin", "kyc.collect-doc", "kyc.screen", "kyc.approve", "kyc.refresh-schedule"},
		Color:       "#FF9800",
		Icon:        "📋",
	}

	vocab.Categories["monitoring"] = &registry.VerbCategory{
		Name:        "monitoring",
		Description: "Ongoing monitoring and screening",
		Verbs:       []string{"screen.continuous"},
		Color:       "#2196F3",
		Icon:        "🔍",
	}

	vocab.Categories["tax-banking"] = &registry.VerbCategory{
		Name:        "tax-banking",
		Description: "Tax information and banking details",
		Verbs:       []string{"tax.capture", "bank.set-instruction"},
		Color:       "#9C27B0",
		Icon:        "🏦",
	}

	vocab.Categories["subscription"] = &registry.VerbCategory{
		Name:        "subscription",
		Description: "Subscription workflow and cash handling",
		Verbs:       []string{"subscribe.request", "cash.confirm", "deal.nav", "subscribe.issue"},
		Color:       "#4CAF50",
		Icon:        "💰",
	}

	vocab.Categories["redemption"] = &registry.VerbCategory{
		Name:        "redemption",
		Description: "Redemption requests and settlement",
		Verbs:       []string{"redeem.request", "redeem.settle"},
		Color:       "#F44336",
		Icon:        "💸",
	}

	vocab.Categories["offboarding"] = &registry.VerbCategory{
		Name:        "offboarding",
		Description: "Investor relationship closure",
		Verbs:       []string{"offboard.close"},
		Color:       "#795548",
		Icon:        "🚪",
	}

	return vocab
}
