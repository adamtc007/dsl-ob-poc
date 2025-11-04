package dsl

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// --- Core types ---

type Verb string

const (
	VerbInvestorStartOpportunity Verb = "investor.start-opportunity"
	VerbInvestorRecordIndication Verb = "investor.record-indication"
	VerbKycBegin                 Verb = "kyc.begin"
	VerbKycCollectDoc            Verb = "kyc.collect-doc"
	VerbKycScreen                Verb = "kyc.screen"
	VerbKycApprove               Verb = "kyc.approve"
	VerbKycRefreshSchedule       Verb = "kyc.refresh-schedule"
	VerbScreenContinuous         Verb = "screen.continuous"
	VerbTaxCapture               Verb = "tax.capture"
	VerbBankSetInstruction       Verb = "bank.set-instruction"
	VerbSubscribeRequest         Verb = "subscribe.request"
	VerbCashConfirm              Verb = "cash.confirm"
	VerbDealNav                  Verb = "deal.nav"
	VerbSubscribeIssue           Verb = "subscribe.issue"
	VerbRedeemRequest            Verb = "redeem.request"
	VerbRedeemSettle             Verb = "redeem.settle"
	VerbOffboardClose            Verb = "offboard.close"
)

type Meta struct {
	IdempotencyKey string `json:"idempotency_key,omitempty"`
	CorrelationID  string `json:"correlation_id,omitempty"`
	CausationID    string `json:"causation_id,omitempty"`
}

type Runbook struct {
	RunbookID string    `json:"runbook_id,omitempty"`
	AsOf      time.Time `json:"as_of,omitempty"`
	Steps     []Step    `json:"steps"`
}

type Step struct {
	ID     string          `json:"id,omitempty"`
	Verb   Verb            `json:"verb"`
	Params json.RawMessage `json:"params"`
	Meta   *Meta           `json:"meta,omitempty"`
	At     *time.Time      `json:"at,omitempty"`
}

// --- Param payloads (mirror the JSON Schema) ---

type InvestorStartOpportunity struct {
	LegalName    string `json:"legal_name"`
	InvestorType string `json:"investor_type,omitempty"`
	Domicile     string `json:"domicile,omitempty"` // ISO-3166-1 alpha-2
	Channel      string `json:"channel,omitempty"`
}

type InvestorRecordIndication struct {
	InvestorID    string   `json:"investor_id"`
	FundID        string   `json:"fund_id"`
	ShareClassID  string   `json:"share_class_id"`
	IndicativeAmt float64  `json:"indicative_amount"`
	Currency      string   `json:"currency"` // ISO-4217
	TargetDate    *ISODate `json:"target_date,omitempty"`
}

type KycBegin struct {
	InvestorID   string   `json:"investor_id"`
	Jurisdiction string   `json:"jurisdiction"` // ISO-3166-1 alpha-2
	ProgramCode  string   `json:"program_code,omitempty"`
	DueDate      *ISODate `json:"due_date,omitempty"`
}

type KycCollectDoc struct {
	InvestorID    string `json:"investor_id"`
	DocType       string `json:"doc_type"`
	DocumentID    string `json:"document_id,omitempty"`
	RequirementID string `json:"requirement_id,omitempty"`
}

type KycScreen struct {
	InvestorID string `json:"investor_id"`
	Vendor     string `json:"vendor,omitempty"`
	Mode       string `json:"mode"` // initial | refresh
	CaseRef    string `json:"case_ref,omitempty"`
}

type KycApprove struct {
	InvestorID string `json:"investor_id"`
	Decision   string `json:"decision"` // APPROVE | REJECT
	Reason     string `json:"reason,omitempty"`
	Approver   string `json:"approver,omitempty"`
}

type KycRefreshSchedule struct {
	InvestorID string   `json:"investor_id"`
	Frequency  string   `json:"frequency"` // ISO 8601 duration e.g. P1Y
	NextDate   *ISODate `json:"next_date,omitempty"`
	RiskRating *int     `json:"risk_rating,omitempty"` // 1..5
}

type ScreenContinuous struct {
	InvestorID string `json:"investor_id"`
	Enable     bool   `json:"enable"`
}

type TaxCapture struct {
	InvestorID    string   `json:"investor_id"`
	FormType      string   `json:"form_type"`
	TIN           string   `json:"tin,omitempty"`
	Country       string   `json:"country,omitempty"` // ISO-3166-1 alpha-2
	EffectiveDate *ISODate `json:"effective_date,omitempty"`
	ExpiryDate    *ISODate `json:"expiry_date,omitempty"`
}

type BankSetInstruction struct {
	InvestorID    string   `json:"investor_id"`
	Currency      string   `json:"currency"` // ISO-4217
	BankName      string   `json:"bank_name,omitempty"`
	Beneficiary   string   `json:"beneficiary_name,omitempty"`
	AccountNumber string   `json:"account_number,omitempty"`
	IBAN          string   `json:"iban,omitempty"`
	SWIFT         string   `json:"swift,omitempty"`
	EffectiveDate *ISODate `json:"effective_date,omitempty"`
}

type SubscribeRequest struct {
	InvestorID   string   `json:"investor_id"`
	FundID       string   `json:"fund_id"`
	ShareClassID string   `json:"share_class_id"`
	Amount       float64  `json:"amount"`
	Currency     string   `json:"currency"`
	DealingDate  *ISODate `json:"dealing_date,omitempty"`
}

type CashConfirm struct {
	InvestorID string  `json:"investor_id"`
	FundID     string  `json:"fund_id"`
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
	ValueDate  ISODate `json:"value_date"`
	Reference  string  `json:"reference,omitempty"`
}

type DealNav struct {
	FundID       string   `json:"fund_id"`
	ShareClassID string   `json:"share_class_id"`
	DealingDate  ISODate  `json:"dealing_date"`
	NavPerUnit   float64  `json:"nav_per_unit"`
	FXRate       *float64 `json:"fx_rate,omitempty"`
}

type SubscribeIssue struct {
	InvestorID   string  `json:"investor_id"`
	FundID       string  `json:"fund_id"`
	ShareClassID string  `json:"share_class_id"`
	SeriesID     string  `json:"series_id,omitempty"`
	Units        float64 `json:"units"`
	NavPerUnit   float64 `json:"nav_per_unit"`
	ValueDate    ISODate `json:"value_date"`
	TradeID      string  `json:"trade_id,omitempty"`
	EventKey     string  `json:"event_key"`
}

type RedeemRequest struct {
	InvestorID   string   `json:"investor_id"`
	FundID       string   `json:"fund_id"`
	ShareClassID string   `json:"share_class_id"`
	Units        *float64 `json:"units,omitempty"`
	Amount       *float64 `json:"amount,omitempty"`
	Currency     *string  `json:"currency,omitempty"`
	DealingDate  *ISODate `json:"dealing_date,omitempty"`
}

type RedeemSettle struct {
	InvestorID   string  `json:"investor_id"`
	FundID       string  `json:"fund_id"`
	ShareClassID string  `json:"share_class_id"`
	Units        float64 `json:"units"`
	NavPerUnit   float64 `json:"nav_per_unit"`
	ValueDate    ISODate `json:"value_date"`
	TradeID      string  `json:"trade_id,omitempty"`
	EventKey     string  `json:"event_key"`
}

type OffboardClose struct {
	InvestorID string `json:"investor_id"`
	Reason     string `json:"reason,omitempty"`
}

// ISO date without time (YYYY-MM-DD)
type ISODate struct{ time.Time }

func (d *ISODate) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return fmt.Errorf("invalid ISO date %q: %w", s, err)
	}
	d.Time = t
	return nil
}

func (d ISODate) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Time.Format("2006-01-02"))
}

// --- Validators ---

func (r *Runbook) Validate() error {
	if len(r.Steps) == 0 {
		return errors.New("steps: must contain at least one step")
	}
	for i := range r.Steps {
		if err := r.Steps[i].Validate(); err != nil {
			return fmt.Errorf("step[%d]: %w", i, err)
		}
	}
	return nil
}

func (s *Step) Validate() error {
	if s.Verb == "" {
		return errors.New("verb: required")
	}
	if len(s.Params) == 0 {
		return errors.New("params: required")
	}
	switch s.Verb {
	case VerbInvestorStartOpportunity:
		var p InvestorStartOpportunity
		if err := json.Unmarshal(s.Params, &p); err != nil {
			return err
		}
		if strings.TrimSpace(p.LegalName) == "" {
			return errors.New("legal_name: required")
		}
	case VerbInvestorRecordIndication:
		var p InvestorRecordIndication
		if err := json.Unmarshal(s.Params, &p); err != nil {
			return err
		}
		if p.IndicativeAmt <= 0 {
			return errors.New("indicative_amount: must be > 0")
		}
		if len(p.Currency) != 3 {
			return errors.New("currency: must be ISO-4217 (3 letters)")
		}
	case VerbKycBegin:
		var p KycBegin
		if err := json.Unmarshal(s.Params, &p); err != nil {
			return err
		}
		if len(p.Jurisdiction) != 2 {
			return errors.New("jurisdiction: must be ISO-3166-1 alpha-2")
		}
	case VerbKycCollectDoc:
		var p KycCollectDoc
		if err := json.Unmarshal(s.Params, &p); err != nil {
			return err
		}
		if strings.TrimSpace(p.DocType) == "" {
			return errors.New("doc_type: required")
		}
	case VerbKycScreen:
		var p KycScreen
		if err := json.Unmarshal(s.Params, &p); err != nil {
			return err
		}
		if p.Mode != "initial" && p.Mode != "refresh" {
			return errors.New("mode: must be 'initial' or 'refresh'")
		}
	case VerbKycApprove:
		var p KycApprove
		if err := json.Unmarshal(s.Params, &p); err != nil {
			return err
		}
		if p.Decision != "APPROVE" && p.Decision != "REJECT" {
			return errors.New("decision: must be APPROVE or REJECT")
		}
	case VerbKycRefreshSchedule:
		var p KycRefreshSchedule
		if err := json.Unmarshal(s.Params, &p); err != nil {
			return err
		}
		if strings.TrimSpace(p.Frequency) == "" {
			return errors.New("frequency: required")
		}
	case VerbScreenContinuous:
		var p ScreenContinuous
		return json.Unmarshal(s.Params, &p)
	case VerbTaxCapture:
		var p TaxCapture
		if err := json.Unmarshal(s.Params, &p); err != nil {
			return err
		}
		if strings.TrimSpace(p.FormType) == "" {
			return errors.New("form_type: required")
		}
	case VerbBankSetInstruction:
		var p BankSetInstruction
		if err := json.Unmarshal(s.Params, &p); err != nil {
			return err
		}
		if len(p.Currency) != 3 {
			return errors.New("currency: must be ISO-4217 (3 letters)")
		}
	case VerbSubscribeRequest:
		var p SubscribeRequest
		if err := json.Unmarshal(s.Params, &p); err != nil {
			return err
		}
		if p.Amount <= 0 {
			return errors.New("amount: must be > 0")
		}
		if len(p.Currency) != 3 {
			return errors.New("currency: must be ISO-4217 (3 letters)")
		}
	case VerbCashConfirm:
		var p CashConfirm
		if err := json.Unmarshal(s.Params, &p); err != nil {
			return err
		}
		if p.Amount <= 0 {
			return errors.New("amount: must be > 0")
		}
		if len(p.Currency) != 3 {
			return errors.New("currency: must be ISO-4217 (3 letters)")
		}
	case VerbDealNav:
		var p DealNav
		if err := json.Unmarshal(s.Params, &p); err != nil {
			return err
		}
		if p.NavPerUnit <= 0 {
			return errors.New("nav_per_unit: must be > 0")
		}
	case VerbSubscribeIssue:
		var p SubscribeIssue
		if err := json.Unmarshal(s.Params, &p); err != nil {
			return err
		}
		if p.Units <= 0 || p.NavPerUnit <= 0 {
			return errors.New("units/nav_per_unit: must be > 0")
		}
		if strings.TrimSpace(p.EventKey) == "" {
			return errors.New("event_key: required")
		}
	case VerbRedeemRequest:
		var p RedeemRequest
		if err := json.Unmarshal(s.Params, &p); err != nil {
			return err
		}
		if (p.Units == nil || *p.Units <= 0) && (p.Amount == nil || *p.Amount <= 0) {
			return errors.New("either units>0 or amount>0 is required")
		}
	case VerbRedeemSettle:
		var p RedeemSettle
		if err := json.Unmarshal(s.Params, &p); err != nil {
			return err
		}
		if p.Units <= 0 || p.NavPerUnit <= 0 {
			return errors.New("units/nav_per_unit: must be > 0")
		}
		if strings.TrimSpace(p.EventKey) == "" {
			return errors.New("event_key: required")
		}
	case VerbOffboardClose:
		var p OffboardClose
		return json.Unmarshal(s.Params, &p)
	default:
		return fmt.Errorf("unsupported verb: %s", s.Verb)
	}
	return nil
}
