package dsl

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// HedgeFundDSLVocab defines the vocabulary for hedge fund investor DSL operations
type HedgeFundDSLVocab struct {
	Domain  string                      `json:"domain"`
	Version string                      `json:"version"`
	Verbs   map[string]HedgeFundVerbDef `json:"verbs"`
}

// HedgeFundVerbDef defines a hedge fund DSL verb with its parameters and effects
type HedgeFundVerbDef struct {
	Name        string                      `json:"name"`
	Domain      string                      `json:"domain"`
	Category    string                      `json:"category"`
	Args        map[string]HedgeFundArgSpec `json:"args"`
	StateChange *HedgeFundStateTransition   `json:"state_change,omitempty"`
	Description string                      `json:"description"`
}

// HedgeFundArgSpec defines argument specifications for hedge fund DSL verbs
type HedgeFundArgSpec struct {
	Type        string   `json:"type"`
	Required    bool     `json:"required"`
	Description string   `json:"description"`
	Values      []string `json:"values,omitempty"` // For enum types
}

// HedgeFundStateTransition defines state transition rules
type HedgeFundStateTransition struct {
	FromStates []string `json:"from_states"`
	ToState    string   `json:"to_state"`
	Guards     []string `json:"guards,omitempty"`
}

// GetHedgeFundDSLVocabulary returns the complete hedge fund investor DSL vocabulary
func GetHedgeFundDSLVocabulary() *HedgeFundDSLVocab {
	return &HedgeFundDSLVocab{
		Domain:  "hedge-fund-investor",
		Version: "1.0.0",
		Verbs: map[string]HedgeFundVerbDef{
			// Opportunity Management
			"investor.start-opportunity": {
				Name:        "investor.start-opportunity",
				Domain:      "hedge-fund-investor",
				Category:    "opportunity",
				Description: "Create initial investor opportunity record",
				Args: map[string]HedgeFundArgSpec{
					"legal-name": {Type: "string", Required: true, Description: "Legal name of the investor"},
					"type":       {Type: "enum", Required: true, Description: "Investor type", Values: []string{"INDIVIDUAL", "CORPORATE", "TRUST", "FOHF", "NOMINEE"}},
					"domicile":   {Type: "string", Required: true, Description: "Investor domicile jurisdiction"},
					"source":     {Type: "string", Required: false, Description: "Source of the investor lead"},
				},
				StateChange: &HedgeFundStateTransition{
					FromStates: []string{},
					ToState:    "OPPORTUNITY",
				},
			},
			"investor.record-indication": {
				Name:        "investor.record-indication",
				Domain:      "hedge-fund-investor",
				Category:    "opportunity",
				Description: "Record investor's indication of interest",
				Args: map[string]HedgeFundArgSpec{
					"investor": {Type: "uuid", Required: true, Description: "Investor ID"},
					"fund":     {Type: "uuid", Required: true, Description: "Fund ID"},
					"class":    {Type: "uuid", Required: true, Description: "Share class ID"},
					"ticket":   {Type: "decimal", Required: true, Description: "Indicated investment amount"},
					"currency": {Type: "string", Required: true, Description: "Currency of investment"},
				},
				StateChange: &HedgeFundStateTransition{
					FromStates: []string{"OPPORTUNITY"},
					ToState:    "PRECHECKS",
				},
			},

			// KYC/KYB Process
			"kyc.begin": {
				Name:        "kyc.begin",
				Domain:      "hedge-fund-investor",
				Category:    "kyc",
				Description: "Begin KYC/KYB process for investor",
				Args: map[string]HedgeFundArgSpec{
					"investor": {Type: "uuid", Required: true, Description: "Investor ID"},
					"tier":     {Type: "enum", Required: false, Description: "KYC tier", Values: []string{"SIMPLIFIED", "STANDARD", "ENHANCED"}},
				},
				StateChange: &HedgeFundStateTransition{
					FromStates: []string{"PRECHECKS"},
					ToState:    "KYC_PENDING",
				},
			},
			"kyc.collect-doc": {
				Name:        "kyc.collect-doc",
				Domain:      "hedge-fund-investor",
				Category:    "kyc",
				Description: "Collect KYC document from investor",
				Args: map[string]HedgeFundArgSpec{
					"investor":  {Type: "uuid", Required: true, Description: "Investor ID"},
					"doc-type":  {Type: "string", Required: true, Description: "Document type"},
					"subject":   {Type: "string", Required: false, Description: "Document subject (e.g., primary_signatory)"},
					"file-path": {Type: "string", Required: false, Description: "Path to uploaded document"},
				},
			},
			"kyc.screen": {
				Name:        "kyc.screen",
				Domain:      "hedge-fund-investor",
				Category:    "kyc",
				Description: "Perform KYC screening against sanctions/PEP lists",
				Args: map[string]HedgeFundArgSpec{
					"investor": {Type: "uuid", Required: true, Description: "Investor ID"},
					"provider": {Type: "enum", Required: true, Description: "Screening provider", Values: []string{"worldcheck", "refinitiv", "accelus"}},
				},
			},
			"kyc.approve": {
				Name:        "kyc.approve",
				Domain:      "hedge-fund-investor",
				Category:    "kyc",
				Description: "Approve KYC and assign risk rating",
				Args: map[string]HedgeFundArgSpec{
					"investor":    {Type: "uuid", Required: true, Description: "Investor ID"},
					"risk":        {Type: "enum", Required: true, Description: "Risk rating", Values: []string{"LOW", "MEDIUM", "HIGH"}},
					"refresh-due": {Type: "date", Required: true, Description: "Next refresh due date"},
					"approved-by": {Type: "string", Required: true, Description: "Approver name"},
					"comments":    {Type: "string", Required: false, Description: "Approval comments"},
				},
				StateChange: &HedgeFundStateTransition{
					FromStates: []string{"KYC_PENDING"},
					ToState:    "KYC_APPROVED",
				},
			},

			// Tax & Banking Setup
			"tax.capture": {
				Name:        "tax.capture",
				Domain:      "hedge-fund-investor",
				Category:    "tax",
				Description: "Capture tax classification information",
				Args: map[string]HedgeFundArgSpec{
					"investor":  {Type: "uuid", Required: true, Description: "Investor ID"},
					"fatca":     {Type: "enum", Required: false, Description: "FATCA status", Values: []string{"US_PERSON", "NON_US_PERSON", "SPECIFIED_US_PERSON"}},
					"crs":       {Type: "enum", Required: false, Description: "CRS classification", Values: []string{"INDIVIDUAL", "ENTITY", "FINANCIAL_INSTITUTION"}},
					"form":      {Type: "enum", Required: false, Description: "Tax form type", Values: []string{"W9", "W8_BEN", "W8_BEN_E", "ENTITY_SELF_CERT"}},
					"tin-type":  {Type: "enum", Required: false, Description: "TIN type", Values: []string{"SSN", "EIN", "FOREIGN_TIN"}},
					"tin-value": {Type: "string", Required: false, Description: "TIN value"},
				},
			},
			"bank.set-instruction": {
				Name:        "bank.set-instruction",
				Domain:      "hedge-fund-investor",
				Category:    "banking",
				Description: "Set banking instruction for settlement",
				Args: map[string]HedgeFundArgSpec{
					"investor":     {Type: "uuid", Required: true, Description: "Investor ID"},
					"currency":     {Type: "string", Required: true, Description: "Settlement currency"},
					"bank-name":    {Type: "string", Required: true, Description: "Bank name"},
					"account-name": {Type: "string", Required: true, Description: "Account name"},
					"iban":         {Type: "string", Required: false, Description: "IBAN"},
					"swift":        {Type: "string", Required: false, Description: "SWIFT BIC"},
					"account-num":  {Type: "string", Required: false, Description: "Account number"},
				},
			},

			// Subscription Workflow
			"subscribe.request": {
				Name:        "subscribe.request",
				Domain:      "hedge-fund-investor",
				Category:    "trading",
				Description: "Submit subscription request",
				Args: map[string]HedgeFundArgSpec{
					"investor":   {Type: "uuid", Required: true, Description: "Investor ID"},
					"fund":       {Type: "uuid", Required: true, Description: "Fund ID"},
					"class":      {Type: "uuid", Required: true, Description: "Share class ID"},
					"amount":     {Type: "decimal", Required: true, Description: "Subscription amount"},
					"currency":   {Type: "string", Required: true, Description: "Settlement currency"},
					"trade-date": {Type: "date", Required: true, Description: "Trade date"},
					"value-date": {Type: "date", Required: true, Description: "Value date"},
				},
				StateChange: &HedgeFundStateTransition{
					FromStates: []string{"KYC_APPROVED", "ACTIVE"},
					ToState:    "SUB_PENDING_CASH",
				},
			},
			"cash.confirm": {
				Name:        "cash.confirm",
				Domain:      "hedge-fund-investor",
				Category:    "settlement",
				Description: "Confirm cash receipt for subscription",
				Args: map[string]HedgeFundArgSpec{
					"investor":      {Type: "uuid", Required: true, Description: "Investor ID"},
					"trade":         {Type: "uuid", Required: true, Description: "Trade ID"},
					"amount":        {Type: "decimal", Required: true, Description: "Amount received"},
					"value-date":    {Type: "date", Required: true, Description: "Value date"},
					"bank-currency": {Type: "string", Required: true, Description: "Received currency"},
					"reference":     {Type: "string", Required: false, Description: "Bank reference"},
				},
				StateChange: &HedgeFundStateTransition{
					FromStates: []string{"SUB_PENDING_CASH"},
					ToState:    "FUNDED_PENDING_NAV",
				},
			},
			"deal.nav": {
				Name:        "deal.nav",
				Domain:      "hedge-fund-investor",
				Category:    "pricing",
				Description: "Set NAV for dealing date",
				Args: map[string]HedgeFundArgSpec{
					"fund":     {Type: "uuid", Required: true, Description: "Fund ID"},
					"class":    {Type: "uuid", Required: true, Description: "Share class ID"},
					"nav-date": {Type: "date", Required: true, Description: "NAV date"},
					"nav":      {Type: "decimal", Required: true, Description: "NAV per share"},
				},
			},
			"subscribe.issue": {
				Name:        "subscribe.issue",
				Domain:      "hedge-fund-investor",
				Category:    "trading",
				Description: "Issue units to investor",
				Args: map[string]HedgeFundArgSpec{
					"investor":      {Type: "uuid", Required: true, Description: "Investor ID"},
					"trade":         {Type: "uuid", Required: true, Description: "Trade ID"},
					"class":         {Type: "uuid", Required: true, Description: "Share class ID"},
					"series":        {Type: "uuid", Required: false, Description: "Series ID (if applicable)"},
					"nav-per-share": {Type: "decimal", Required: true, Description: "NAV per share"},
					"units":         {Type: "decimal", Required: true, Description: "Units to issue"},
				},
				StateChange: &HedgeFundStateTransition{
					FromStates: []string{"FUNDED_PENDING_NAV"},
					ToState:    "ISSUED",
				},
			},

			// Ongoing Operations
			"kyc.refresh-schedule": {
				Name:        "kyc.refresh-schedule",
				Domain:      "hedge-fund-investor",
				Category:    "kyc",
				Description: "Schedule KYC refresh",
				Args: map[string]HedgeFundArgSpec{
					"investor":  {Type: "uuid", Required: true, Description: "Investor ID"},
					"frequency": {Type: "enum", Required: true, Description: "Refresh frequency", Values: []string{"MONTHLY", "QUARTERLY", "ANNUAL"}},
					"next":      {Type: "date", Required: true, Description: "Next refresh date"},
				},
			},
			"screen.continuous": {
				Name:        "screen.continuous",
				Domain:      "hedge-fund-investor",
				Category:    "compliance",
				Description: "Set up continuous screening",
				Args: map[string]HedgeFundArgSpec{
					"investor":  {Type: "uuid", Required: true, Description: "Investor ID"},
					"frequency": {Type: "enum", Required: true, Description: "Screening frequency", Values: []string{"DAILY", "WEEKLY", "MONTHLY"}},
				},
			},

			// Redemption & Offboarding
			"redeem.request": {
				Name:        "redeem.request",
				Domain:      "hedge-fund-investor",
				Category:    "trading",
				Description: "Submit redemption request",
				Args: map[string]HedgeFundArgSpec{
					"investor":    {Type: "uuid", Required: true, Description: "Investor ID"},
					"class":       {Type: "uuid", Required: true, Description: "Share class ID"},
					"units":       {Type: "decimal", Required: false, Description: "Units to redeem (if partial)"},
					"percentage":  {Type: "decimal", Required: false, Description: "Percentage to redeem (if partial)"},
					"notice-date": {Type: "date", Required: true, Description: "Notice date"},
					"value-date":  {Type: "date", Required: true, Description: "Redemption value date"},
				},
				StateChange: &HedgeFundStateTransition{
					FromStates: []string{"ACTIVE"},
					ToState:    "REDEEM_PENDING",
				},
			},
			"redeem.settle": {
				Name:        "redeem.settle",
				Domain:      "hedge-fund-investor",
				Category:    "settlement",
				Description: "Settle redemption payment",
				Args: map[string]HedgeFundArgSpec{
					"investor":    {Type: "uuid", Required: true, Description: "Investor ID"},
					"trade":       {Type: "uuid", Required: true, Description: "Redemption trade ID"},
					"amount":      {Type: "decimal", Required: true, Description: "Settlement amount"},
					"settle-date": {Type: "date", Required: true, Description: "Settlement date"},
					"reference":   {Type: "string", Required: false, Description: "Payment reference"},
				},
				StateChange: &HedgeFundStateTransition{
					FromStates: []string{"REDEEM_PENDING"},
					ToState:    "REDEEMED",
				},
			},
			"offboard.close": {
				Name:        "offboard.close",
				Domain:      "hedge-fund-investor",
				Category:    "lifecycle",
				Description: "Complete investor offboarding",
				Args: map[string]HedgeFundArgSpec{
					"investor": {Type: "uuid", Required: true, Description: "Investor ID"},
					"reason":   {Type: "string", Required: false, Description: "Offboarding reason"},
				},
				StateChange: &HedgeFundStateTransition{
					FromStates: []string{"REDEEMED"},
					ToState:    "OFFBOARDED",
				},
			},
		},
	}
}

// HedgeFundDSLContext represents the context for DSL execution
type HedgeFundDSLContext struct {
	InvestorID uuid.UUID              `json:"investor_id"`
	FundID     *uuid.UUID             `json:"fund_id,omitempty"`
	ClassID    *uuid.UUID             `json:"class_id,omitempty"`
	SeriesID   *uuid.UUID             `json:"series_id,omitempty"`
	TradeID    *uuid.UUID             `json:"trade_id,omitempty"`
	Variables  map[string]interface{} `json:"variables"`
}

// HedgeFundDSLOperation represents a single DSL operation to be executed
type HedgeFundDSLOperation struct {
	Verb      string                 `json:"verb"`
	Args      map[string]interface{} `json:"args"`
	Context   *HedgeFundDSLContext   `json:"context,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// HedgeFundDSLPlan represents a complete plan of DSL operations
type HedgeFundDSLPlan struct {
	PlanID      uuid.UUID               `json:"plan_id"`
	InvestorID  uuid.UUID               `json:"investor_id"`
	Description string                  `json:"description"`
	Operations  []HedgeFundDSLOperation `json:"operations"`
	Variables   map[string]interface{}  `json:"variables"`
	CreatedAt   time.Time               `json:"created_at"`
	CreatedBy   string                  `json:"created_by"`
}

// Variable patterns for hedge fund DSL
var (
	// Investor variables
	hedgeFundInvestorVarRegex = regexp.MustCompile(`\?INV\.([A-Z_.]+)`)
	// KYC variables
	hedgeFundKYCVarRegex = regexp.MustCompile(`\?KYC\.([A-Z_]+)`)
	// Tax variables
	hedgeFundTaxVarRegex = regexp.MustCompile(`\?TAX\.([A-Z_]+)`)
	// Banking variables (with currency index)
	hedgeFundBankVarRegex = regexp.MustCompile(`\?BANK\[([A-Z]{3})\]\.([A-Z_]+)`)
	// Trading variables
	hedgeFundTradeVarRegex = regexp.MustCompile(`\?TRADE\.([A-Z_]+)`)
	// Date variables
	hedgeFundDateVarRegex = regexp.MustCompile(`\?DATE\.([A-Z_]+)`)
)

// GenerateHedgeFundDSL creates DSL text from a hedge fund operation
func GenerateHedgeFundDSL(operation *HedgeFundDSLOperation) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("(%s", operation.Verb))

	// Add arguments in a consistent order
	for key, value := range operation.Args {
		switch v := value.(type) {
		case string:
			b.WriteString(fmt.Sprintf("\n  :%s %q", key, v))
		case uuid.UUID:
			b.WriteString(fmt.Sprintf("\n  :%s %q", key, v.String()))
		case float64:
			b.WriteString(fmt.Sprintf("\n  :%s %.8f", key, v))
		case time.Time:
			b.WriteString(fmt.Sprintf("\n  :%s %q", key, v.Format("2006-01-02")))
		case bool:
			b.WriteString(fmt.Sprintf("\n  :%s %t", key, v))
		default:
			b.WriteString(fmt.Sprintf("\n  :%s %v", key, v))
		}
	}

	b.WriteString(")")
	return b.String()
}

// GenerateHedgeFundDSLPlan creates DSL text from a complete plan
func GenerateHedgeFundDSLPlan(plan *HedgeFundDSLPlan) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("(hedge-fund.plan\n  :plan-id %q\n  :investor-id %q\n  :description %q\n\n",
		plan.PlanID.String(), plan.InvestorID.String(), plan.Description))

	// Add variables section if any
	if len(plan.Variables) > 0 {
		b.WriteString("  (variables\n")
		for key, value := range plan.Variables {
			switch v := value.(type) {
			case string:
				b.WriteString(fmt.Sprintf("    :%s %q\n", key, v))
			default:
				b.WriteString(fmt.Sprintf("    :%s %v\n", key, v))
			}
		}
		b.WriteString("  )\n\n")
	}

	// Add operations
	b.WriteString("  (operations\n")
	for _, op := range plan.Operations {
		opDSL := GenerateHedgeFundDSL(&op)
		// Indent the operation DSL
		indentedDSL := strings.ReplaceAll(opDSL, "\n", "\n    ")
		b.WriteString(fmt.Sprintf("    %s\n", indentedDSL))
	}
	b.WriteString("  )")

	b.WriteString("\n)")
	return b.String()
}

// ParseHedgeFundInvestorVars extracts investor variables from DSL text
func ParseHedgeFundInvestorVars(dslText string) map[string]string {
	vars := make(map[string]string)

	// Extract investor variables
	matches := hedgeFundInvestorVarRegex.FindAllStringSubmatch(dslText, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			vars[match[0]] = match[1] // Full match -> variable name
		}
	}

	// Extract KYC variables
	matches = hedgeFundKYCVarRegex.FindAllStringSubmatch(dslText, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			vars[match[0]] = match[1]
		}
	}

	// Extract tax variables
	matches = hedgeFundTaxVarRegex.FindAllStringSubmatch(dslText, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			vars[match[0]] = match[1]
		}
	}

	// Extract banking variables
	matches = hedgeFundBankVarRegex.FindAllStringSubmatch(dslText, -1)
	for _, match := range matches {
		if len(match) >= 3 {
			vars[match[0]] = fmt.Sprintf("%s.%s", match[1], match[2]) // Currency.Field
		}
	}

	// Extract trading variables
	matches = hedgeFundTradeVarRegex.FindAllStringSubmatch(dslText, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			vars[match[0]] = match[1]
		}
	}
	// Extract date variables
	matches = hedgeFundDateVarRegex.FindAllStringSubmatch(dslText, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			vars[match[0]] = match[1]
		}
	}

	return vars
}

// ValidateHedgeFundDSLOperation validates a hedge fund DSL operation against the vocabulary
func ValidateHedgeFundDSLOperation(operation *HedgeFundDSLOperation) error {
	vocab := GetHedgeFundDSLVocabulary()

	verbDef, exists := vocab.Verbs[operation.Verb]
	if !exists {
		return fmt.Errorf("unknown hedge fund DSL verb: %s", operation.Verb)
	}

	// Check required arguments
	for argName, argSpec := range verbDef.Args {
		value, provided := operation.Args[argName]

		if argSpec.Required && !provided {
			return fmt.Errorf("required argument '%s' not provided for verb '%s'", argName, operation.Verb)
		}

		if provided {
			// Validate argument type and values
			if err := validateHedgeFundArgument(argName, value, argSpec); err != nil {
				return fmt.Errorf("invalid argument '%s' for verb '%s': %w", argName, operation.Verb, err)
			}
		}
	}

	return nil
}

// validateHedgeFundArgument validates a single argument against its specification
func validateHedgeFundArgument(_ string, value interface{}, spec HedgeFundArgSpec) error {
	switch spec.Type {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case "uuid":
		switch v := value.(type) {
		case string:
			if _, err := uuid.Parse(v); err != nil {
				return fmt.Errorf("invalid UUID format: %s", v)
			}
		case uuid.UUID:
			// Already valid
		default:
			return fmt.Errorf("expected UUID string or uuid.UUID, got %T", value)
		}
	case "decimal":
		switch value.(type) {
		case float64, int, int64:
			// Valid numeric types
		default:
			return fmt.Errorf("expected numeric value, got %T", value)
		}
	case "date":
		switch v := value.(type) {
		case string:
			if _, err := time.Parse("2006-01-02", v); err != nil {
				return fmt.Errorf("invalid date format (expected YYYY-MM-DD): %s", v)
			}
		case time.Time:
			// Already valid
		default:
			return fmt.Errorf("expected date string or time.Time, got %T", value)
		}
	case "enum":
		valueStr := fmt.Sprintf("%v", value)
		for _, allowed := range spec.Values {
			if valueStr == allowed {
				return nil
			}
		}
		return fmt.Errorf("invalid enum value '%v', allowed values: %v", value, spec.Values)
	}

	return nil
}
