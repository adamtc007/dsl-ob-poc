package state

import (
	"fmt"
	"time"

	"dsl-ob-poc/internal/hf-investor/domain"

	"github.com/google/uuid"
)

// HedgeFundLifecycleState represents a lifecycle state transition
type HedgeFundLifecycleState struct {
	StateID           uuid.UUID              `json:"state_id" db:"state_id"`
	InvestorID        uuid.UUID              `json:"investor_id" db:"investor_id"`
	FromState         *string                `json:"from_state,omitempty" db:"from_state"`
	ToState           string                 `json:"to_state" db:"to_state"`
	TransitionTrigger *string                `json:"transition_trigger,omitempty" db:"transition_trigger"`
	GuardConditions   map[string]interface{} `json:"guard_conditions,omitempty" db:"guard_conditions"`
	Metadata          map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	TransitionedBy    *string                `json:"transitioned_by,omitempty" db:"transitioned_by"`
	TransitionedAt    time.Time              `json:"transitioned_at" db:"transitioned_at"`
}

// HedgeFundStateMachine manages investor lifecycle state transitions
type HedgeFundStateMachine struct {
	transitions map[string][]string
	guards      map[string][]GuardCondition
}

// GuardCondition represents a condition that must be met for state transition
type GuardCondition struct {
	Name        string
	Description string
	Check       func(investor *domain.HedgeFundInvestor, context map[string]interface{}) (bool, error)
}

// NewHedgeFundStateMachine creates a new hedge fund investor state machine
func NewHedgeFundStateMachine() *HedgeFundStateMachine {
	sm := &HedgeFundStateMachine{
		transitions: make(map[string][]string),
		guards:      make(map[string][]GuardCondition),
	}

	// Define valid state transitions
	sm.transitions = map[string][]string{
		domain.InvestorStatusOpportunity: {
			domain.InvestorStatusPrechecks,
		},
		domain.InvestorStatusPrechecks: {
			domain.InvestorStatusKYCPending,
			domain.InvestorStatusOpportunity, // Can go back
		},
		domain.InvestorStatusKYCPending: {
			domain.InvestorStatusKYCApproved,
			domain.InvestorStatusPrechecks, // Can go back
		},
		domain.InvestorStatusKYCApproved: {
			domain.InvestorStatusSubPendingCash,
		},
		domain.InvestorStatusSubPendingCash: {
			domain.InvestorStatusFundedPendingNAV,
			domain.InvestorStatusKYCApproved, // Can go back if subscription canceled
		},
		domain.InvestorStatusFundedPendingNAV: {
			domain.InvestorStatusIssued,
		},
		domain.InvestorStatusIssued: {
			domain.InvestorStatusActive,
		},
		domain.InvestorStatusActive: {
			domain.InvestorStatusRedeemPending,
			domain.InvestorStatusSubPendingCash, // Can add more money
		},
		domain.InvestorStatusRedeemPending: {
			domain.InvestorStatusRedeemed,
			domain.InvestorStatusActive, // Can cancel redemption
		},
		domain.InvestorStatusRedeemed: {
			domain.InvestorStatusOffboarded,
			domain.InvestorStatusSubPendingCash, // Can reinvest
		},
		domain.InvestorStatusOffboarded: {
			// Terminal state - no transitions allowed
		},
	}

	// Define guard conditions for each state transition
	sm.setupGuardConditions()

	return sm
}

// setupGuardConditions defines the guard conditions for state transitions
func (sm *HedgeFundStateMachine) setupGuardConditions() {
	sm.guards = map[string][]GuardCondition{
		fmt.Sprintf("%s->%s", domain.InvestorStatusOpportunity, domain.InvestorStatusPrechecks): {
			{
				Name:        "indication_recorded",
				Description: "Investment indication must be recorded",
				Check: func(investor *domain.HedgeFundInvestor, context map[string]interface{}) (bool, error) {
					// Check if indication has been recorded
					if indication, exists := context["indication"]; exists {
						return indication != nil, nil
					}
					return false, fmt.Errorf("no investment indication found")
				},
			},
		},
		fmt.Sprintf("%s->%s", domain.InvestorStatusPrechecks, domain.InvestorStatusKYCPending): {
			{
				Name:        "initial_docs_submitted",
				Description: "Initial KYC documents must be submitted",
				Check: func(investor *domain.HedgeFundInvestor, context map[string]interface{}) (bool, error) {
					// Check if initial documents have been submitted
					if docs, exists := context["initial_documents"]; exists {
						return docs != nil, nil
					}
					return false, fmt.Errorf("initial KYC documents not submitted")
				},
			},
		},
		fmt.Sprintf("%s->%s", domain.InvestorStatusKYCPending, domain.InvestorStatusKYCApproved): {
			{
				Name:        "kyc_documents_verified",
				Description: "All required KYC documents must be verified",
				Check: func(investor *domain.HedgeFundInvestor, context map[string]interface{}) (bool, error) {
					if verified, exists := context["documents_verified"]; exists {
						if v, ok := verified.(bool); ok {
							return v, nil
						}
					}
					return false, fmt.Errorf("KYC documents not fully verified")
				},
			},
			{
				Name:        "screening_passed",
				Description: "AML/sanctions screening must pass",
				Check: func(investor *domain.HedgeFundInvestor, context map[string]interface{}) (bool, error) {
					if screening, exists := context["screening_result"]; exists {
						if result, ok := screening.(string); ok {
							return result == domain.ScreeningResultClear, nil
						}
					}
					return false, fmt.Errorf("AML/sanctions screening not passed")
				},
			},
			{
				Name:        "risk_rating_assigned",
				Description: "Risk rating must be assigned",
				Check: func(investor *domain.HedgeFundInvestor, context map[string]interface{}) (bool, error) {
					if rating, exists := context["risk_rating"]; exists {
						if r, ok := rating.(string); ok {
							return r != "", nil
						}
					}
					return false, fmt.Errorf("risk rating not assigned")
				},
			},
		},
		fmt.Sprintf("%s->%s", domain.InvestorStatusKYCApproved, domain.InvestorStatusSubPendingCash): {
			{
				Name:        "valid_subscription_order",
				Description: "Valid subscription order must exist",
				Check: func(investor *domain.HedgeFundInvestor, context map[string]interface{}) (bool, error) {
					if order, exists := context["subscription_order"]; exists {
						return order != nil, nil
					}
					return false, fmt.Errorf("no valid subscription order found")
				},
			},
			{
				Name:        "minimum_investment_met",
				Description: "Subscription amount must meet minimum investment threshold",
				Check: func(investor *domain.HedgeFundInvestor, context map[string]interface{}) (bool, error) {
					amount, amountExists := context["subscription_amount"]
					minAmount, minExists := context["minimum_investment"]

					if !amountExists || !minExists {
						return false, fmt.Errorf("subscription amount or minimum investment not specified")
					}

					if amt, ok := amount.(float64); ok {
						if min, minOk := minAmount.(float64); minOk {
							return amt >= min, nil
						}
					}
					return false, fmt.Errorf("invalid subscription amount format")
				},
			},
			{
				Name:        "banking_instructions_set",
				Description: "Banking instructions must be configured",
				Check: func(investor *domain.HedgeFundInvestor, context map[string]interface{}) (bool, error) {
					if banking, exists := context["banking_instructions"]; exists {
						return banking != nil, nil
					}
					return false, fmt.Errorf("banking instructions not configured")
				},
			},
		},
		fmt.Sprintf("%s->%s", domain.InvestorStatusSubPendingCash, domain.InvestorStatusFundedPendingNAV): {
			{
				Name:        "settlement_funds_received",
				Description: "Settlement funds must be received and confirmed",
				Check: func(investor *domain.HedgeFundInvestor, context map[string]interface{}) (bool, error) {
					if funds, exists := context["funds_received"]; exists {
						if confirmed, ok := funds.(bool); ok {
							return confirmed, nil
						}
					}
					return false, fmt.Errorf("settlement funds not received or confirmed")
				},
			},
		},
		fmt.Sprintf("%s->%s", domain.InvestorStatusFundedPendingNAV, domain.InvestorStatusIssued): {
			{
				Name:        "nav_struck",
				Description: "NAV must be struck for the dealing date",
				Check: func(investor *domain.HedgeFundInvestor, context map[string]interface{}) (bool, error) {
					if nav, exists := context["nav_per_share"]; exists {
						if navValue, ok := nav.(float64); ok {
							return navValue > 0, nil
						}
					}
					return false, fmt.Errorf("NAV not struck for dealing date")
				},
			},
			{
				Name:        "units_allocated",
				Description: "Units must be allocated to investor",
				Check: func(investor *domain.HedgeFundInvestor, context map[string]interface{}) (bool, error) {
					if units, exists := context["units_allocated"]; exists {
						if unitsValue, ok := units.(float64); ok {
							return unitsValue > 0, nil
						}
					}
					return false, fmt.Errorf("units not allocated to investor")
				},
			},
		},
		fmt.Sprintf("%s->%s", domain.InvestorStatusActive, domain.InvestorStatusRedeemPending): {
			{
				Name:        "valid_redemption_notice",
				Description: "Valid redemption notice must be submitted",
				Check: func(investor *domain.HedgeFundInvestor, context map[string]interface{}) (bool, error) {
					if notice, exists := context["redemption_notice"]; exists {
						return notice != nil, nil
					}
					return false, fmt.Errorf("no valid redemption notice found")
				},
			},
			{
				Name:        "notice_period_met",
				Description: "Required notice period must be satisfied",
				Check: func(investor *domain.HedgeFundInvestor, context map[string]interface{}) (bool, error) {
					if noticeMet, exists := context["notice_period_satisfied"]; exists {
						if satisfied, ok := noticeMet.(bool); ok {
							return satisfied, nil
						}
					}
					return false, fmt.Errorf("required notice period not satisfied")
				},
			},
		},
		fmt.Sprintf("%s->%s", domain.InvestorStatusRedeemPending, domain.InvestorStatusRedeemed): {
			{
				Name:        "all_units_redeemed",
				Description: "All requested units must be redeemed",
				Check: func(investor *domain.HedgeFundInvestor, context map[string]interface{}) (bool, error) {
					if redeemed, exists := context["units_redeemed"]; exists {
						if units, ok := redeemed.(float64); ok {
							return units > 0, nil
						}
					}
					return false, fmt.Errorf("units not redeemed")
				},
			},
			{
				Name:        "cash_payment_made",
				Description: "Cash payment must be made to investor",
				Check: func(investor *domain.HedgeFundInvestor, context map[string]interface{}) (bool, error) {
					if payment, exists := context["cash_payment_made"]; exists {
						if paid, ok := payment.(bool); ok {
							return paid, nil
						}
					}
					return false, fmt.Errorf("cash payment not made to investor")
				},
			},
		},
		fmt.Sprintf("%s->%s", domain.InvestorStatusRedeemed, domain.InvestorStatusOffboarded): {
			{
				Name:        "final_documentation_complete",
				Description: "Final documentation and cleanup must be complete",
				Check: func(investor *domain.HedgeFundInvestor, context map[string]interface{}) (bool, error) {
					if complete, exists := context["final_docs_complete"]; exists {
						if isComplete, ok := complete.(bool); ok {
							return isComplete, nil
						}
					}
					return false, fmt.Errorf("final documentation not complete")
				},
			},
		},
	}
}

// CanTransition checks if a state transition is valid
func (sm *HedgeFundStateMachine) CanTransition(fromState, toState string) bool {
	allowedStates, exists := sm.transitions[fromState]
	if !exists {
		return false
	}

	for _, allowed := range allowedStates {
		if allowed == toState {
			return true
		}
	}
	return false
}

// ValidateTransition validates a state transition including guard conditions
func (sm *HedgeFundStateMachine) ValidateTransition(
	investor *domain.HedgeFundInvestor,
	toState string,
	context map[string]interface{},
) error {
	// Check if transition is structurally valid
	if !sm.CanTransition(investor.Status, toState) {
		return fmt.Errorf("invalid state transition from %s to %s", investor.Status, toState)
	}

	// Check guard conditions
	transitionKey := fmt.Sprintf("%s->%s", investor.Status, toState)
	guards, hasGuards := sm.guards[transitionKey]

	if hasGuards {
		for _, guard := range guards {
			passed, err := guard.Check(investor, context)
			if err != nil {
				return fmt.Errorf("guard condition '%s' failed: %w", guard.Name, err)
			}
			if !passed {
				return fmt.Errorf("guard condition '%s' not satisfied: %s", guard.Name, guard.Description)
			}
		}
	}

	return nil
}

// TransitionState performs a state transition with validation
func (sm *HedgeFundStateMachine) TransitionState(
	investor *domain.HedgeFundInvestor,
	toState string,
	trigger string,
	context map[string]interface{},
	transitionedBy string,
) (*HedgeFundLifecycleState, error) {
	// Validate the transition
	if err := sm.ValidateTransition(investor, toState, context); err != nil {
		return nil, err
	}

	// Create the lifecycle state record
	fromState := investor.Status
	lifecycleState := &HedgeFundLifecycleState{
		StateID:           uuid.New(),
		InvestorID:        investor.InvestorID,
		FromState:         &fromState,
		ToState:           toState,
		TransitionTrigger: &trigger,
		GuardConditions:   context,
		TransitionedBy:    &transitionedBy,
		TransitionedAt:    time.Now().UTC(),
	}

	// Update investor status
	investor.Status = toState
	investor.UpdatedAt = time.Now().UTC()

	return lifecycleState, nil
}

// GetValidTransitions returns all valid transitions from the current state
func (sm *HedgeFundStateMachine) GetValidTransitions(fromState string) []string {
	if transitions, exists := sm.transitions[fromState]; exists {
		return append([]string(nil), transitions...) // Return a copy
	}
	return []string{}
}

// GetGuardConditions returns the guard conditions for a specific transition
func (sm *HedgeFundStateMachine) GetGuardConditions(fromState, toState string) []GuardCondition {
	transitionKey := fmt.Sprintf("%s->%s", fromState, toState)
	if guards, exists := sm.guards[transitionKey]; exists {
		return append([]GuardCondition(nil), guards...) // Return a copy
	}
	return []GuardCondition{}
}

// IsTerminalState checks if a state is terminal (no outbound transitions)
func (sm *HedgeFundStateMachine) IsTerminalState(state string) bool {
	transitions, exists := sm.transitions[state]
	return !exists || len(transitions) == 0
}

// GetAllStates returns all possible states in the state machine
func (sm *HedgeFundStateMachine) GetAllStates() []string {
	stateSet := make(map[string]bool)

	// Add all "from" states
	for fromState := range sm.transitions {
		stateSet[fromState] = true
	}

	// Add all "to" states
	for _, toStates := range sm.transitions {
		for _, toState := range toStates {
			stateSet[toState] = true
		}
	}

	// Convert to slice
	states := make([]string, 0, len(stateSet))
	for state := range stateSet {
		states = append(states, state)
	}

	return states
}

// StateTransitionPath finds a path between two states
func (sm *HedgeFundStateMachine) StateTransitionPath(fromState, toState string) ([]string, error) {
	if fromState == toState {
		return []string{fromState}, nil
	}

	// Use BFS to find shortest path
	queue := [][]string{{fromState}}
	visited := make(map[string]bool)
	visited[fromState] = true

	for len(queue) > 0 {
		path := queue[0]
		queue = queue[1:]

		currentState := path[len(path)-1]

		// Check all possible transitions from current state
		for _, nextState := range sm.GetValidTransitions(currentState) {
			if nextState == toState {
				return append(path, nextState), nil
			}

			if !visited[nextState] {
				visited[nextState] = true
				queue = append(queue, append(path, nextState))
			}
		}
	}

	return nil, fmt.Errorf("no transition path found from %s to %s", fromState, toState)
}
