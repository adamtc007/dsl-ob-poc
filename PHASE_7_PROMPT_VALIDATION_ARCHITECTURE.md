# Phase 7: Prompt Validation & Cleanup - ARCHITECTURAL DESIGN

## Overview

Phase 7 addresses the user requirement to handle prompts with incorrect or incomplete content using Domain DSL knowledge and rules. This document outlines two complementary architectural approaches:

1. **Command Palette + Macro System** (Primary) - Predictable, fast, no-LLM UI
2. **LLM-based Prompt Repair** (Fallback) - AI correction for complex cases

## 🎯 Core Architectural Principle

> **Prefer deterministic, predictable patterns over AI when possible.**  
> Use LLM only when users provide complex free-form input that can't be handled by structured flows.

---

## Architecture Option A: Command Palette + Macro System ⭐ **RECOMMENDED**

### Design Philosophy

If your verbs/EBNF are tight, you can give non-tech users a "command-palette + snippets" UX that feels like IDE autocomplete, with **no LLM needed**. This is faster, more predictable, and teaches users the domain language.

### 1. Command Palette Flow (UX wired to hotkey ⌘K)

```
1. Choose an intent: Create case • Edit case • Resume case • Generate DSL
2. Pick domain + record: Onboarding • KYC • HF Investor → select or create DSL sheet
3. Choose action macro (autocomplete): "Onboard CBU to products", "Start KYC (STD)", "Subscribe amount to fund/class"
4. Slot fill UI (chips/inputs): auto-suggest from catalogs (CBUs, funds, classes), validate types live
5. Preview: Show Plan IR + S-expr side-by-side (read-only), with "Apply to Sheet"
```

**Result**: Predictable, fast, teaches domain language as users go.

### 2. Macro Specification (Human-Readable Config)

Place in `config/macros.yaml`. Placeholders drive the mini form; `$attr` marks dictionary-bound values.

```yaml
# config/macros.yaml
version: 1
macros:
  - id: onboard_cbu_products
    label: "Onboard CBU to products"
    domain: ONBOARDING
    placeholders:
      - key: CBU_NAME
        label: "CBU legal name"
        type: string
        required: true
        suggest: catalog.cbus         # server-side lookup
      - key: PRODUCTS
        label: "Products"
        type: multiselect
        required: true
        options:
          - CUSTODY
          - FUND_ACCOUNTING
          - ALT_INVESTMENTS
    steps:
      - verb: onboard.cbu.start
        params:
          cbu_name: "${CBU_NAME}"
          channel: "DIRECT"
      - verb: resource.enable
        repeat: PRODUCTS              # fan-out each selected product
        params:
          cbu_id: {"$attr": "cbu.id"}
          product_code: "${ITEM}"     # current product value

  - id: kyc_standard
    label: "Start KYC (Standard)"
    domain: KYC
    placeholders:
      - key: INVESTOR
        label: "Investor"
        type: string
        required: true
        suggest: catalog.investors
      - key: JURIS
        label: "Jurisdiction"
        type: select
        options: [GB, IE, LU, US]     # trim as needed
    steps:
      - verb: kyc.begin
        params:
          investor_id: {"$attr": "investor.id"}
          jurisdiction: "${JURIS}"
          program_code: "STD"
      - verb: kyc.screen
        params:
          investor_id: {"$attr": "investor.id"}
          mode: "initial"
          vendor: {"$attr": "kyc.vendor"}

  - id: hf_subscribe_amount
    label: "Subscribe amount to fund/class"
    domain: HF_INVESTOR
    placeholders:
      - key: INVESTOR
        label: "Investor"
        type: string
        suggest: catalog.investors
        required: true
      - key: FUND
        label: "Fund"
        type: string
        suggest: catalog.funds
        required: true
      - key: CLASS
        label: "Share Class"
        type: select
        options: [CLASS_I, CLASS_A, CLASS_B]
        required: true
      - key: AMOUNT
        label: "Amount"
        type: money
        required: true
      - key: CCY
        label: "Currency"
        type: select
        options: [USD, GBP, EUR]
        required: true
    steps:
      - verb: investor.record-indication
        params:
          investor_id: {"$attr": "investor.id"}
          fund_id: {"$attr": "fund.id"}
          share_class_id: {"$attr": "class.id"}
          indicative_amount: "${AMOUNT}"
          currency: "${CCY}"
          target_date: {"$attr": "class.next_deal_date"}
      - verb: subscribe.request
        params:
          investor_id: {"$attr": "investor.id"}
          fund_id: {"$attr": "fund.id"}
          share_class_id: {"$attr": "class.id"}
          amount: "${AMOUNT}"
          currency: "${CCY}"
          dealing_date: {"$attr": "class.next_deal_date"}
```

**Key Features**:
- `${PLACEHOLDER}` → replaced by user input
- `repeat: PRODUCTS` → fan-out a step per selected item
- `{"$attr": "…"}` → keep as dictionary placeholder for your binder

### 3. Macro Engine Implementation

Create `internal/macros/engine.go`:

```go
package macros

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type PlanIR struct {
	Domain      string   `json:"domain"`
	SourceText  string   `json:"source_text,omitempty"`
	Assumptions []string `json:"assumptions,omitempty"`
	Steps       []Step   `json:"steps"`
	Questions   []PIQ    `json:"questions,omitempty"`
}

type Step struct {
	Verb   string         `json:"verb"`
	Params map[string]any `json:"params"`
}

type PIQ struct {
	ID      string   `json:"id"`
	Text    string   `json:"text"`
	AttrID  string   `json:"attr_id,omitempty"`
	Options []string `json:"options,omitempty"`
}

type Macro struct {
	ID           string         `json:"id"`
	Label        string         `json:"label"`
	Domain       string         `json:"domain"`
	Placeholders []Placeholder  `json:"placeholders"`
	Steps        []MacroStep    `json:"steps"`
}

type Placeholder struct {
	Key      string   `json:"key"`
	Label    string   `json:"label"`
	Type     string   `json:"type"` // string|select|multiselect|money
	Required bool     `json:"required"`
	Options  []string `json:"options,omitempty"`
	Suggest  string   `json:"suggest,omitempty"` // e.g., catalog.cbus
}

type MacroStep struct {
	Verb   string         `json:"verb"`
	Repeat string         `json:"repeat,omitempty"` // key of placeholder to iterate
	Params map[string]any `json:"params"`
}

var reVar = regexp.MustCompile(`\$\{([A-Z0-9_]+)\}`)

type Values map[string]any // user-provided values per placeholder

func Expand(m Macro, vals Values) (PlanIR, []PIQ, error) {
	// gather questions for missing required placeholders
	var questions []PIQ
	for _, p := range m.Placeholders {
		if p.Required && !hasValue(vals, p.Key) {
			q := PIQ{ID: p.Key, Text: p.Label}
			if len(p.Options) > 0 { 
				q.Options = append([]string{}, p.Options...) 
			}
			questions = append(questions, q)
		}
	}
	if len(questions) > 0 {
		return PlanIR{}, questions, nil
	}

	plan := PlanIR{Domain: m.Domain, Steps: []Step{}}

	// helper for variable and $attr handling
	expandVal := func(v any, item any) any {
		switch t := v.(type) {
		case string:
			// attribute placeholder passthrough (JSON object given inline)
			if strings.HasPrefix(strings.TrimSpace(t), "@") {
				// allow "@cbu.id" short-hand → {"$attr":"cbu.id"}
				return map[string]any{"$attr": strings.TrimPrefix(t, "@")}
			}
			out := reVar.ReplaceAllStringFunc(t, func(mm string) string {
				key := reVar.FindStringSubmatch(mm)[1]
				if key == "ITEM" && item != nil {
					return fmt.Sprint(item)
				}
				if vv, ok := vals[key]; ok {
					return fmt.Sprint(vv)
				}
				return "" // missing optional
			})
			return coerce(out)
		case map[string]any, []any, float64, bool, nil:
			return t
		default:
			return t
		}
	}

	for _, st := range m.Steps {
		if st.Repeat != "" {
			items, _ := vals[st.Repeat].([]any)
			if items == nil {
				// tolerate []string too
				if ss, ok := vals[st.Repeat].([]string); ok {
					items = make([]any, len(ss))
					for i := range ss { items[i] = ss[i] }
				}
			}
			for _, it := range items {
				params := map[string]any{}
				for k, v := range st.Params {
					params[k] = expandVal(v, it)
				}
				plan.Steps = append(plan.Steps, Step{Verb: st.Verb, Params: params})
			}
			continue
		}
		params := map[string]any{}
		for k, v := range st.Params {
			params[k] = expandVal(v, nil)
		}
		plan.Steps = append(plan.Steps, Step{Verb: st.Verb, Params: params})
	}
	return plan, nil, nil
}

func hasValue(vals Values, key string) bool {
	_, ok := vals[key]
	return ok
}

func coerce(s string) any {
	// simple numeric coercion for money/ints if you want; keep as string by default
	return s
}

// ToSExpr renders deterministic S-expr for preview
func ToSExpr(p PlanIR) string {
	var b strings.Builder
	for _, s := range p.Steps {
		b.WriteByte('(')
		b.WriteString(s.Verb)
		keys := make([]string, 0, len(s.Params))
		for k := range s.Params { 
			keys = append(keys, k) 
		}
		sort.Strings(keys)
		for _, k := range keys {
			b.WriteString(" :")
			b.WriteString(strings.ReplaceAll(k, "_", "-"))
			b.WriteByte(' ')
			b.WriteString(renderVal(s.Params[k]))
		}
		b.WriteString(")\n")
	}
	return b.String()
}

func renderVal(v any) string {
	switch t := v.(type) {
	case map[string]any:
		if a, ok := t["$attr"]; ok { 
			return "@"+fmt.Sprint(a) 
		}
		b,_ := json.Marshal(t); return string(b)
	case string:
		b,_ := json.Marshal(t); return string(b)
	default:
		b,_ := json.Marshal(t); return string(b)
	}
}
```

### 4. API Endpoints for UI Integration

Add to existing server:

```go
type MacroInvokeReq struct {
	MacroID string         `json:"macro_id"`
	Domain  string         `json:"domain,omitempty"`
	Values  map[string]any `json:"values,omitempty"` // user answers so far
}

type MacroInvokeResp struct {
	Decision  string         `json:"decision"` // QUESTIONS | PLAN
	Questions []macros.PIQ   `json:"questions,omitempty"`
	Plan      *macros.PlanIR `json:"plan,omitempty"`
	SExpr     string         `json:"s_expr,omitempty"`
}

func handleMacroExpand(w http.ResponseWriter, r *http.Request) {
	var req MacroInvokeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, 400, errResp{Error:"bad json: "+err.Error()})
		return
	}
	
	// load macros.yaml once at boot; here we just look up
	m, ok := macroRegistry[req.MacroID]
	if !ok { 
		writeJSON(w,400, errResp{Error:"unknown macro"})
		return 
	}

	plan, qs, err := macros.Expand(m, req.Values)
	if err != nil { 
		writeJSON(w,500, errResp{Error: err.Error()})
		return 
	}

	if len(qs) > 0 {
		writeJSON(w,200, MacroInvokeResp{Decision:"QUESTIONS", Questions: qs})
		return
	}

	// validate against verb/spec validator if you want belt & braces
	// if errs := validatePlanIR(plan); len(errs)>0 { ... }

	sexpr := macros.ToSExpr(plan)
	writeJSON(w,200, MacroInvokeResp{Decision:"PLAN", Plan:&plan, SExpr: sexpr})
}
```

**Usage Flow**:
1. Call once with just `macro_id` → get back questions (mini form)
2. Call again with values filled → get Plan IR + S-expr preview
3. POST to existing `/plan/validate` → `/dsl/emit` or straight to apply

### 5. IDE-Style Completion Extras

**Nice-to-have enhancements**:
- Inline ghost text in command bar ("onboard {CBU} to {products}")
- Fuzzy lookup chips for catalogs (with Jaro-Winkler similarity)
- Live type hints under each field ("expects ISO-4217 currency")
- One-click "Use defaults" to auto-fill optional slots from dictionary
- History panel of last 10 applied macros; "Re-run with changes"

---

## Architecture Option B: LLM-based Prompt Repair (Fallback)

### When to Use LLMs

**Only when**:
- User refuses forms and dumps free prose
- User pastes multi-paragraph brief
- Complex, unstructured input that can't be handled by macros

**Even then**:
- Route through existing Spec IR → Plan IR tool flow
- Use JSON-only output with validator repair loop
- Macro palette remains primary path for speed + predictability

### Implementation Approach

```go
// When macro system can't handle input
func handleComplexPromptRepair(prompt string, domain string) (*DSLValidationResult, error) {
    // 1. Try to parse intent from prompt
    intent, entities := extractIntentAndEntities(prompt)
    
    // 2. Check if we have a matching macro
    if macro := findMatchingMacro(intent, domain); macro != nil {
        // Try to auto-fill macro from extracted entities
        values := mapEntitiesToMacroValues(entities, macro)
        plan, questions, err := macros.Expand(macro, values)
        
        if len(questions) == 0 && err == nil {
            // Success - converted free text to structured macro
            return &DSLValidationResult{
                Valid: true,
                Plan:  plan,
                SExpr: macros.ToSExpr(plan),
            }, nil
        }
    }
    
    // 3. Fall back to LLM-based repair
    return llmPromptRepair(prompt, domain)
}

func llmPromptRepair(prompt string, domain string) (*DSLValidationResult, error) {
    // Use existing domain vocabulary and validation
    vocab := registry.GetVocabulary(domain)
    
    systemPrompt := fmt.Sprintf(`
You are a DSL repair agent. The user provided incorrect or incomplete input.
Available verbs: %s
Domain: %s

Fix the user's prompt and generate valid DSL. Respond only with JSON:
{
  "repaired_prompt": "corrected version",
  "plan": {"domain": "%s", "steps": [...]},
  "s_expr": "generated DSL"
}
`, strings.Join(vocab.VerbNames(), ", "), domain, domain)

    // Call LLM with structured output constraints
    response := callLLMWithValidation(systemPrompt, prompt)
    
    // Validate generated DSL against domain rules
    return validateAndRepairDSL(response, domain)
}
```

### Validation and Repair Loop

```go
func validateAndRepairDSL(response LLMResponse, domain string) (*DSLValidationResult, error) {
    dom, err := registry.Get(domain)
    if err != nil {
        return nil, err
    }
    
    // Validate verbs
    if err := dom.ValidateVerbs(response.SExpr); err != nil {
        // Repair loop: feed error back to LLM
        return repairWithFeedback(response, err, domain)
    }
    
    // Validate state transitions
    if err := validateStateTransitions(response.Plan, domain); err != nil {
        return repairWithFeedback(response, err, domain)
    }
    
    return &DSLValidationResult{
        Valid: true,
        Plan:  response.Plan,
        SExpr: response.SExpr,
        RepairedPrompt: response.RepairedPrompt,
    }, nil
}

func repairWithFeedback(response LLMResponse, validationError error, domain string) (*DSLValidationResult, error) {
    repairPrompt := fmt.Sprintf(`
The generated DSL has validation errors:
Error: %s

Original DSL: %s
Fix the errors and regenerate valid DSL.
`, validationError.Error(), response.SExpr)

    // Recursive repair with limit
    return llmPromptRepairWithContext(repairPrompt, domain, response)
}
```

---

## Integration with DSL State Manager

Both approaches MUST integrate with the DSL State Manager:

```go
func applyMacroToSession(sessionID string, plan macros.PlanIR) error {
    sessionMgr := session.NewManager()
    dslSession := sessionMgr.GetOrCreate(sessionID, plan.Domain)
    
    // Generate DSL from plan
    generatedDSL := macros.ToSExpr(plan)
    
    // Accumulate through state manager (single source of truth)
    return dslSession.AccumulateDSL(generatedDSL)
}

func applyRepairedDSLToSession(sessionID string, result *DSLValidationResult) error {
    sessionMgr := session.NewManager()
    dslSession := sessionMgr.GetOrCreate(sessionID, result.Plan.Domain)
    
    // Accumulate repaired DSL through state manager
    return dslSession.AccumulateDSL(result.SExpr)
}
```

---

## Architecture Decision Matrix

| Aspect | Command Palette + Macros | LLM Prompt Repair |
|--------|---------------------------|-------------------|
| **Speed** | ⚡ Instant | 🐌 2-5 seconds |
| **Predictability** | ✅ Deterministic | ❓ Variable quality |
| **User Learning** | ✅ Teaches domain | ❌ Black box |
| **Maintenance** | ✅ Config-driven | ❌ Prompt engineering |
| **Accuracy** | ✅ 100% | ❓ 85-95% |
| **Cost** | ✅ Free | 💸 LLM API calls |
| **Offline** | ✅ Works offline | ❌ Requires internet |
| **Complex Input** | ❌ Structured only | ✅ Any text |

## Recommended Implementation Strategy

### Phase 7.1: Command Palette System (Primary)
1. Implement macro engine and YAML parser
2. Create initial macro library for common operations
3. Build command palette UI component
4. Integrate with DSL State Manager
5. Add catalog suggestion endpoints

### Phase 7.2: LLM Fallback (Secondary)  
1. Implement intent detection for macro matching
2. Add LLM repair for unstructured input
3. Create validation and repair loop
4. Integrate with existing domain validation

### Phase 7.3: Hybrid UX
1. Command palette as default entry point
2. "Convert to macro" option for free text
3. Seamless fallback to LLM when needed
4. Learning prompts to guide users toward macros

---

## Success Criteria

✅ **Command Palette System**:
- 90%+ of operations handled without LLM
- Sub-100ms response time for macro expansion
- Zero invalid DSL generation
- Complete integration with DSL State Manager

✅ **LLM Fallback**:
- 95%+ accuracy for complex prompt repair
- Automatic validation and repair loop
- Proper error messaging and suggestions
- Graceful degradation when LLM unavailable

✅ **User Experience**:
- Command palette discoverable and intuitive
- Users learn domain vocabulary through usage
- Smooth transition between structured and free-form input
- Complete audit trail through DSL State Manager

---

## Conclusion

The **Command Palette + Macro System** approach provides a superior user experience that is:
- **Faster** than LLM-based approaches
- **More predictable** and reliable  
- **Educational** for domain vocabulary
- **Cost-effective** with no API dependencies
- **Architecturally sound** with full DSL State Manager integration

The **LLM Prompt Repair** serves as a necessary fallback for complex, unstructured input while maintaining the architectural integrity of the system.

This hybrid approach delivers the best of both worlds: IDE-like productivity for structured operations and AI assistance for complex scenarios.