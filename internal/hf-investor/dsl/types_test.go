package dsl

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestISODate_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "valid date",
			input: `"2024-12-31"`,
			want:  "2024-12-31",
		},
		{
			name:    "invalid format",
			input:   `"31-12-2024"`,
			wantErr: true,
		},
		{
			name:    "invalid date",
			input:   `"2024-13-01"`,
			wantErr: true,
		},
		{
			name:    "not a string",
			input:   `123`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d ISODate
			err := json.Unmarshal([]byte(tt.input), &d)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if got := d.Format("2006-01-02"); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestISODate_MarshalJSON(t *testing.T) {
	d := ISODate{time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)}
	data, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `"2024-12-31"`
	if string(data) != want {
		t.Errorf("got %s, want %s", string(data), want)
	}
}

func TestRunbook_Validate(t *testing.T) {
	tests := []struct {
		name    string
		runbook Runbook
		wantErr string
	}{
		{
			name: "valid runbook",
			runbook: Runbook{
				Steps: []Step{
					{
						Verb:   VerbInvestorStartOpportunity,
						Params: json.RawMessage(`{"legal_name":"Test Corp"}`),
					},
				},
			},
		},
		{
			name:    "empty steps",
			runbook: Runbook{Steps: []Step{}},
			wantErr: "steps: must contain at least one step",
		},
		{
			name: "invalid step",
			runbook: Runbook{
				Steps: []Step{
					{
						Verb:   VerbInvestorStartOpportunity,
						Params: json.RawMessage(`{"legal_name":""}`),
					},
				},
			},
			wantErr: "step[0]: legal_name: required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.runbook.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Error("expected error but got none")
				return
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("got error %q, want to contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestStep_Validate(t *testing.T) {
	tests := []struct {
		name    string
		step    Step
		wantErr string
	}{
		{
			name: "valid investor start opportunity",
			step: Step{
				Verb:   VerbInvestorStartOpportunity,
				Params: json.RawMessage(`{"legal_name":"Test Corp","domicile":"US"}`),
			},
		},
		{
			name: "missing verb",
			step: Step{
				Params: json.RawMessage(`{"legal_name":"Test Corp"}`),
			},
			wantErr: "verb: required",
		},
		{
			name: "missing params",
			step: Step{
				Verb: VerbInvestorStartOpportunity,
			},
			wantErr: "params: required",
		},
		{
			name: "invalid legal name",
			step: Step{
				Verb:   VerbInvestorStartOpportunity,
				Params: json.RawMessage(`{"legal_name":"  "}`),
			},
			wantErr: "legal_name: required",
		},
		{
			name: "valid record indication",
			step: Step{
				Verb:   VerbInvestorRecordIndication,
				Params: json.RawMessage(`{"investor_id":"123e4567-e89b-12d3-a456-426614174000","fund_id":"123e4567-e89b-12d3-a456-426614174001","share_class_id":"123e4567-e89b-12d3-a456-426614174002","indicative_amount":100000,"currency":"USD"}`),
			},
		},
		{
			name: "invalid indicative amount",
			step: Step{
				Verb:   VerbInvestorRecordIndication,
				Params: json.RawMessage(`{"investor_id":"123e4567-e89b-12d3-a456-426614174000","fund_id":"123e4567-e89b-12d3-a456-426614174001","share_class_id":"123e4567-e89b-12d3-a456-426614174002","indicative_amount":0,"currency":"USD"}`),
			},
			wantErr: "indicative_amount: must be > 0",
		},
		{
			name: "invalid currency",
			step: Step{
				Verb:   VerbInvestorRecordIndication,
				Params: json.RawMessage(`{"investor_id":"123e4567-e89b-12d3-a456-426614174000","fund_id":"123e4567-e89b-12d3-a456-426614174001","share_class_id":"123e4567-e89b-12d3-a456-426614174002","indicative_amount":100000,"currency":"INVALID"}`),
			},
			wantErr: "currency: must be ISO-4217 (3 letters)",
		},
		{
			name: "valid kyc begin",
			step: Step{
				Verb:   VerbKycBegin,
				Params: json.RawMessage(`{"investor_id":"123e4567-e89b-12d3-a456-426614174000","jurisdiction":"US"}`),
			},
		},
		{
			name: "invalid jurisdiction",
			step: Step{
				Verb:   VerbKycBegin,
				Params: json.RawMessage(`{"investor_id":"123e4567-e89b-12d3-a456-426614174000","jurisdiction":"USA"}`),
			},
			wantErr: "jurisdiction: must be ISO-3166-1 alpha-2",
		},
		{
			name: "valid kyc collect doc",
			step: Step{
				Verb:   VerbKycCollectDoc,
				Params: json.RawMessage(`{"investor_id":"123e4567-e89b-12d3-a456-426614174000","doc_type":"PASSPORT"}`),
			},
		},
		{
			name: "missing doc type",
			step: Step{
				Verb:   VerbKycCollectDoc,
				Params: json.RawMessage(`{"investor_id":"123e4567-e89b-12d3-a456-426614174000","doc_type":"  "}`),
			},
			wantErr: "doc_type: required",
		},
		{
			name: "valid kyc screen",
			step: Step{
				Verb:   VerbKycScreen,
				Params: json.RawMessage(`{"investor_id":"123e4567-e89b-12d3-a456-426614174000","mode":"initial"}`),
			},
		},
		{
			name: "invalid screen mode",
			step: Step{
				Verb:   VerbKycScreen,
				Params: json.RawMessage(`{"investor_id":"123e4567-e89b-12d3-a456-426614174000","mode":"invalid"}`),
			},
			wantErr: "mode: must be 'initial' or 'refresh'",
		},
		{
			name: "valid kyc approve",
			step: Step{
				Verb:   VerbKycApprove,
				Params: json.RawMessage(`{"investor_id":"123e4567-e89b-12d3-a456-426614174000","decision":"APPROVE"}`),
			},
		},
		{
			name: "invalid decision",
			step: Step{
				Verb:   VerbKycApprove,
				Params: json.RawMessage(`{"investor_id":"123e4567-e89b-12d3-a456-426614174000","decision":"MAYBE"}`),
			},
			wantErr: "decision: must be APPROVE or REJECT",
		},
		{
			name: "valid kyc refresh schedule",
			step: Step{
				Verb:   VerbKycRefreshSchedule,
				Params: json.RawMessage(`{"investor_id":"123e4567-e89b-12d3-a456-426614174000","frequency":"P1Y"}`),
			},
		},
		{
			name: "missing frequency",
			step: Step{
				Verb:   VerbKycRefreshSchedule,
				Params: json.RawMessage(`{"investor_id":"123e4567-e89b-12d3-a456-426614174000","frequency":"  "}`),
			},
			wantErr: "frequency: required",
		},
		{
			name: "valid screen continuous",
			step: Step{
				Verb:   VerbScreenContinuous,
				Params: json.RawMessage(`{"investor_id":"123e4567-e89b-12d3-a456-426614174000","enable":true}`),
			},
		},
		{
			name: "valid tax capture",
			step: Step{
				Verb:   VerbTaxCapture,
				Params: json.RawMessage(`{"investor_id":"123e4567-e89b-12d3-a456-426614174000","form_type":"W8BEN"}`),
			},
		},
		{
			name: "missing form type",
			step: Step{
				Verb:   VerbTaxCapture,
				Params: json.RawMessage(`{"investor_id":"123e4567-e89b-12d3-a456-426614174000","form_type":"  "}`),
			},
			wantErr: "form_type: required",
		},
		{
			name: "valid bank instruction",
			step: Step{
				Verb:   VerbBankSetInstruction,
				Params: json.RawMessage(`{"investor_id":"123e4567-e89b-12d3-a456-426614174000","currency":"USD"}`),
			},
		},
		{
			name: "invalid bank currency",
			step: Step{
				Verb:   VerbBankSetInstruction,
				Params: json.RawMessage(`{"investor_id":"123e4567-e89b-12d3-a456-426614174000","currency":"DOLLAR"}`),
			},
			wantErr: "currency: must be ISO-4217 (3 letters)",
		},
		{
			name: "valid subscribe request",
			step: Step{
				Verb:   VerbSubscribeRequest,
				Params: json.RawMessage(`{"investor_id":"123e4567-e89b-12d3-a456-426614174000","fund_id":"123e4567-e89b-12d3-a456-426614174001","share_class_id":"123e4567-e89b-12d3-a456-426614174002","amount":100000,"currency":"USD"}`),
			},
		},
		{
			name: "invalid subscribe amount",
			step: Step{
				Verb:   VerbSubscribeRequest,
				Params: json.RawMessage(`{"investor_id":"123e4567-e89b-12d3-a456-426614174000","fund_id":"123e4567-e89b-12d3-a456-426614174001","share_class_id":"123e4567-e89b-12d3-a456-426614174002","amount":-1000,"currency":"USD"}`),
			},
			wantErr: "amount: must be > 0",
		},
		{
			name: "valid cash confirm",
			step: Step{
				Verb:   VerbCashConfirm,
				Params: json.RawMessage(`{"investor_id":"123e4567-e89b-12d3-a456-426614174000","fund_id":"123e4567-e89b-12d3-a456-426614174001","amount":100000,"currency":"USD","value_date":"2024-12-31"}`),
			},
		},
		{
			name: "valid deal nav",
			step: Step{
				Verb:   VerbDealNav,
				Params: json.RawMessage(`{"fund_id":"123e4567-e89b-12d3-a456-426614174001","share_class_id":"123e4567-e89b-12d3-a456-426614174002","dealing_date":"2024-12-31","nav_per_unit":100.50}`),
			},
		},
		{
			name: "invalid nav per unit",
			step: Step{
				Verb:   VerbDealNav,
				Params: json.RawMessage(`{"fund_id":"123e4567-e89b-12d3-a456-426614174001","share_class_id":"123e4567-e89b-12d3-a456-426614174002","dealing_date":"2024-12-31","nav_per_unit":0}`),
			},
			wantErr: "nav_per_unit: must be > 0",
		},
		{
			name: "valid subscribe issue",
			step: Step{
				Verb:   VerbSubscribeIssue,
				Params: json.RawMessage(`{"investor_id":"123e4567-e89b-12d3-a456-426614174000","fund_id":"123e4567-e89b-12d3-a456-426614174001","share_class_id":"123e4567-e89b-12d3-a456-426614174002","units":1000,"nav_per_unit":100.50,"value_date":"2024-12-31","event_key":"SUB-2024-001"}`),
			},
		},
		{
			name: "invalid units and nav",
			step: Step{
				Verb:   VerbSubscribeIssue,
				Params: json.RawMessage(`{"investor_id":"123e4567-e89b-12d3-a456-426614174000","fund_id":"123e4567-e89b-12d3-a456-426614174001","share_class_id":"123e4567-e89b-12d3-a456-426614174002","units":0,"nav_per_unit":100.50,"value_date":"2024-12-31","event_key":"SUB-2024-001"}`),
			},
			wantErr: "units/nav_per_unit: must be > 0",
		},
		{
			name: "missing event key",
			step: Step{
				Verb:   VerbSubscribeIssue,
				Params: json.RawMessage(`{"investor_id":"123e4567-e89b-12d3-a456-426614174000","fund_id":"123e4567-e89b-12d3-a456-426614174001","share_class_id":"123e4567-e89b-12d3-a456-426614174002","units":1000,"nav_per_unit":100.50,"value_date":"2024-12-31","event_key":"  "}`),
			},
			wantErr: "event_key: required",
		},
		{
			name: "valid redeem request with units",
			step: Step{
				Verb:   VerbRedeemRequest,
				Params: json.RawMessage(`{"investor_id":"123e4567-e89b-12d3-a456-426614174000","fund_id":"123e4567-e89b-12d3-a456-426614174001","share_class_id":"123e4567-e89b-12d3-a456-426614174002","units":500}`),
			},
		},
		{
			name: "valid redeem request with amount",
			step: Step{
				Verb:   VerbRedeemRequest,
				Params: json.RawMessage(`{"investor_id":"123e4567-e89b-12d3-a456-426614174000","fund_id":"123e4567-e89b-12d3-a456-426614174001","share_class_id":"123e4567-e89b-12d3-a456-426614174002","amount":50000,"currency":"USD"}`),
			},
		},
		{
			name: "invalid redeem request - no units or amount",
			step: Step{
				Verb:   VerbRedeemRequest,
				Params: json.RawMessage(`{"investor_id":"123e4567-e89b-12d3-a456-426614174000","fund_id":"123e4567-e89b-12d3-a456-426614174001","share_class_id":"123e4567-e89b-12d3-a456-426614174002"}`),
			},
			wantErr: "either units>0 or amount>0 is required",
		},
		{
			name: "valid redeem settle",
			step: Step{
				Verb:   VerbRedeemSettle,
				Params: json.RawMessage(`{"investor_id":"123e4567-e89b-12d3-a456-426614174000","fund_id":"123e4567-e89b-12d3-a456-426614174001","share_class_id":"123e4567-e89b-12d3-a456-426614174002","units":500,"nav_per_unit":100.50,"value_date":"2024-12-31","event_key":"RED-2024-001"}`),
			},
		},
		{
			name: "valid offboard close",
			step: Step{
				Verb:   VerbOffboardClose,
				Params: json.RawMessage(`{"investor_id":"123e4567-e89b-12d3-a456-426614174000","reason":"Voluntary closure"}`),
			},
		},
		{
			name: "unsupported verb",
			step: Step{
				Verb:   "unknown.verb",
				Params: json.RawMessage(`{}`),
			},
			wantErr: "unsupported verb: unknown.verb",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.step.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Error("expected error but got none")
				return
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("got error %q, want to contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestVerbConstants(t *testing.T) {
	expectedVerbs := []Verb{
		VerbInvestorStartOpportunity,
		VerbInvestorRecordIndication,
		VerbKycBegin,
		VerbKycCollectDoc,
		VerbKycScreen,
		VerbKycApprove,
		VerbKycRefreshSchedule,
		VerbScreenContinuous,
		VerbTaxCapture,
		VerbBankSetInstruction,
		VerbSubscribeRequest,
		VerbCashConfirm,
		VerbDealNav,
		VerbSubscribeIssue,
		VerbRedeemRequest,
		VerbRedeemSettle,
		VerbOffboardClose,
	}

	// Ensure all verbs are unique
	verbMap := make(map[Verb]bool)
	for _, verb := range expectedVerbs {
		if verbMap[verb] {
			t.Errorf("duplicate verb: %s", verb)
		}
		verbMap[verb] = true
	}

	// Ensure all verbs follow naming convention
	for _, verb := range expectedVerbs {
		verbStr := string(verb)
		if !strings.Contains(verbStr, ".") {
			t.Errorf("verb %q should contain namespace separator '.'", verbStr)
		}
	}
}
