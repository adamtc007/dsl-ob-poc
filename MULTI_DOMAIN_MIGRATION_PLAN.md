# Multi-Domain Architecture Migration Plan

## Executive Summary

Migrate the proven hedge fund investor chat interface to support **multiple DSL domains** with shared infrastructure. The onboarding domain becomes the **orchestration layer** that can call into hedge fund investor, KYC, and product onboarding subdomains.

## Core Architectural Principle

**Domain-Specific Vocabularies + Shared Infrastructure**

### Shared Infrastructure (Domain-Agnostic)
1. **Data Dictionary** - Universal attribute definitions (AttributeID-as-Type system)
2. **EBNF Grammar** - S-expression syntax rules
3. **DSL Parser** - Parses any domain's DSL into AST
4. **DSL Syntax Validator** - Validates S-expression structure
5. **Chat Session Management** - Stateful conversation tracking
6. **DSL Accumulation Engine** - Builds up DSL state across operations
7. **UUID Resolution Service** - Resolves placeholders to actual UUIDs
8. **Context Tracking** - Maintains entities across messages

### Domain-Specific Components
1. **Verb Vocabularies** - Each domain defines its approved verbs
   - `onboarding.*` - Orchestration verbs (case.create, products.add, etc.)
   - `investor.*` - Hedge fund investor lifecycle (investor.start-opportunity, etc.)
   - `kyc.*` - KYC/AML workflows (kyc.begin, kyc.collect-doc, etc.)
   - `product.*` - Product-specific onboarding
2. **Domain Agents** - AI agents specialized in generating domain DSL
3. **Verb Validators** - Domain-specific approved verb lists
4. **State Machines** - Domain-specific lifecycle states
5. **Business Logic** - Domain-specific rules and constraints

## Request Flow Architecture

```
User Message
    ↓
Domain Router (determines which domain handles request)
    ↓
Domain Agent (generates domain-specific DSL)
    ↓
Shared DSL Parser (syntax check - EBNF validation)
    ↓
Domain Verb Validator (semantic check - approved verbs)
    ↓
Shared Dictionary Service (resolve AttributeIDs)
    ↓
Domain State Machine (apply state transition)
    ↓
DSL Accumulator (append to session BuiltDSL)
    ↓
Response to User
```

## Migration Plan - 6 Phases

---

## Phase 1: Extract Shared Infrastructure (Week 1)

**Goal**: Create domain-agnostic shared libraries from hedge fund implementation

### 1.1 Create Shared Package Structure

```
internal/
├── shared-dsl/                    # NEW: Shared DSL infrastructure
│   ├── parser/                   # S-expression parser (domain-agnostic)
│   │   ├── parser.go            # Parse DSL into AST
│   │   ├── parser_test.go
│   │   └── ast.go               # Abstract Syntax Tree types
│   ├── validator/                # Syntax validation (EBNF rules)
│   │   ├── syntax.go            # Validate S-expression structure
│   │   ├── ebnf.go              # EBNF grammar definition
│   │   └── validator_test.go
│   ├── dictionary/               # Attribute dictionary service
│   │   ├── service.go           # Query attribute definitions
│   │   ├── types.go             # AttributeID, mask, domain, etc.
│   │   └── mock.go              # Mock dictionary for testing
│   ├── session/                  # Chat session management
│   │   ├── manager.go           # Session lifecycle
│   │   ├── context.go           # Entity context tracking
│   │   └── accumulator.go       # DSL accumulation logic
│   └── resolver/                 # UUID resolution service
│       ├── resolver.go          # Resolve <placeholder> → UUID
│       └── resolver_test.go
│
├── domain-registry/              # NEW: Domain management
│   ├── registry.go              # Register and lookup domains
│   ├── domain.go                # Domain interface definition
│   └── router.go                # Route requests to domains
│
└── domains/                      # NEW: Domain implementations
    ├── onboarding/              # Onboarding orchestration domain
    │   ├── agent.go            # Onboarding AI agent
    │   ├── vocab.go            # Onboarding verb vocabulary
    │   ├── validator.go        # Onboarding verb validator
    │   └── state.go            # Onboarding state machine
    │
    └── hedge-fund-investor/     # Hedge fund domain (migrated)
        ├── agent.go            # HF AI agent
        ├── vocab.go            # HF verb vocabulary
        ├── validator.go        # HF verb validator
        └── state.go            # HF state machine
```

### 1.2 Extract DSL Parser

**Source**: Check if existing parser in `internal/dsl/` or `hedge-fund-investor-source/`

**Action**: Create `internal/shared-dsl/parser/`

```go
package parser

// Parse parses DSL S-expressions into an AST (domain-agnostic)
func Parse(dsl string) (*AST, error)

// AST represents an abstract syntax tree of DSL
type AST struct {
    Root *Node
}

type Node struct {
    Type     NodeType  // Verb, Argument, Value
    Value    string
    Children []*Node
}
```

**Tests**: Parse all existing DSL examples from both onboarding and hedge fund

### 1.3 Extract EBNF Validator

**Action**: Create `internal/shared-dsl/validator/`

```go
package validator

// ValidateSyntax checks S-expression syntax against EBNF grammar
func ValidateSyntax(dsl string) error

// EBNF Grammar (embedded)
const Grammar = `
  dsl         = { expression } ;
  expression  = "(" verb { argument } ")" ;
  verb        = identifier "." identifier ;
  argument    = value | expression ;
  value       = string | number | identifier ;
  ...
`
```

### 1.4 Extract Dictionary Service

**Source**: Existing `dictionary` table in PostgreSQL

**Action**: Create `internal/shared-dsl/dictionary/service.go`

```go
package dictionary

type Service interface {
    GetAttribute(ctx context.Context, attrID uuid.UUID) (*Attribute, error)
    FindAttributeByName(ctx context.Context, name string) (*Attribute, error)
    ListAttributes(ctx context.Context, domain string) ([]*Attribute, error)
}

type Attribute struct {
    AttributeID      uuid.UUID
    Name             string
    LongDescription  string
    GroupID          string
    Mask             string  // Data type
    Domain           string  // Business domain
    Source           SourceMetadata
    Sink             SinkMetadata
}
```

**Tests**: Query attributes used in both domains

### 1.5 Extract Session Management

**Source**: `hedge-fund-investor-source/web/server.go` - `ChatSession` struct

**Action**: Create `internal/shared-dsl/session/manager.go`

```go
package session

type Manager struct {
    sessions map[string]*Session
    mu       sync.RWMutex
}

type Session struct {
    SessionID  string
    Domain     string                 // NEW: Which domain is active
    BuiltDSL   string                 // Accumulated DSL
    Context    map[string]interface{} // Entity tracking (investor_id, fund_id, etc.)
    History    []Message
    CreatedAt  time.Time
    LastUsed   time.Time
}

func (m *Manager) GetOrCreate(sessionID, domain string) *Session
func (m *Manager) AccumulateDSL(sessionID, newDSL string) error
func (m *Manager) UpdateContext(sessionID string, updates map[string]interface{}) error
```

### 1.6 Extract UUID Resolver

**Source**: `hedge-fund-investor-source/web/internal/resolver/`

**Action**: Verify and move to `internal/shared-dsl/resolver/`

```go
package resolver

// Resolve replaces placeholders like <investor_id> with actual UUIDs from context
func Resolve(dsl string, context map[string]interface{}) (string, error)
```

**Tests**: Verify resolution of all placeholder types

### 1.7 Deliverables

- [ ] `internal/shared-dsl/` package with 5 subpackages
- [ ] All packages have comprehensive tests
- [ ] Documentation for each shared service
- [ ] Migration guide for existing code

---

## Phase 2: Create Domain Registry System (Week 2)

**Goal**: Enable multiple domains to coexist with dynamic routing

### 2.1 Define Domain Interface

**Action**: Create `internal/domain-registry/domain.go`

```go
package registry

type Domain interface {
    // Identity
    Name() string        // "onboarding", "hedge-fund-investor", "kyc"
    Version() string     // "1.0.0"
    
    // Vocabulary
    GetVocabulary() *Vocabulary
    ValidateVerbs(dsl string) error
    
    // Agent
    GenerateDSL(ctx context.Context, req *GenerationRequest) (*GenerationResponse, error)
    
    // State Machine
    GetCurrentState(context map[string]interface{}) string
    ValidateTransition(from, to string) error
}

type Vocabulary struct {
    Domain  string
    Version string
    Verbs   map[string]VerbDefinition
}

type VerbDefinition struct {
    Name        string
    Category    string
    Args        map[string]ArgSpec
    StateChange *StateTransition
    Description string
}
```

### 2.2 Create Domain Registry

**Action**: Create `internal/domain-registry/registry.go`

```go
package registry

type Registry struct {
    domains map[string]Domain
    mu      sync.RWMutex
}

func NewRegistry() *Registry

func (r *Registry) Register(domain Domain) error
func (r *Registry) Get(domainName string) (Domain, error)
func (r *Registry) List() []string
func (r *Registry) GetVocabulary(domainName string) (*Vocabulary, error)
```

### 2.3 Create Domain Router

**Action**: Create `internal/domain-registry/router.go`

```go
package registry

type Router struct {
    registry *Registry
}

// Route determines which domain should handle the request
func (r *Router) Route(ctx context.Context, request *RoutingRequest) (Domain, error)

type RoutingRequest struct {
    Message       string
    SessionID     string
    CurrentDomain string
    Context       map[string]interface{}
}

// Routing strategies:
// 1. Explicit domain switch: "switch to hedge fund investor domain"
// 2. Context-based: If investor_id exists, route to hedge-fund-investor
// 3. Verb-based: Parse DSL, check which domain owns the verb
// 4. Default: Use session's current domain
```

### 2.4 Deliverables

- [ ] Domain interface definition with full documentation
- [ ] Registry implementation with thread-safety
- [ ] Router with multiple routing strategies
- [ ] Unit tests for registry and router
- [ ] Example domain implementations (stubs)

---

## Phase 3: Migrate Hedge Fund Domain (Week 3)

**Goal**: Refactor hedge fund implementation to use shared infrastructure

### 3.1 Create Hedge Fund Domain Package

**Action**: Create `internal/domains/hedge-fund-investor/`

**Migrate from**: `hedge-fund-investor-source/hf-investor/` and `hedge-fund-investor-source/web/internal/hf-agent/`

### 3.2 Implement Domain Interface

**Action**: Create `internal/domains/hedge-fund-investor/domain.go`

```go
package hedgefund

import (
    "dsl-ob-poc/internal/domain-registry"
    "dsl-ob-poc/internal/shared-dsl/parser"
    "dsl-ob-poc/internal/shared-dsl/dictionary"
)

type HedgeFundDomain struct {
    agent      *Agent
    vocabulary *registry.Vocabulary
    dictionary dictionary.Service
}

func NewHedgeFundDomain(apiKey string, dict dictionary.Service) (*HedgeFundDomain, error)

// Implement Domain interface
func (d *HedgeFundDomain) Name() string { return "hedge-fund-investor" }
func (d *HedgeFundDomain) Version() string { return "1.0.0" }
func (d *HedgeFundDomain) GetVocabulary() *registry.Vocabulary
func (d *HedgeFundDomain) ValidateVerbs(dsl string) error
func (d *HedgeFundDomain) GenerateDSL(ctx context.Context, req *registry.GenerationRequest) (*registry.GenerationResponse, error)
```

### 3.3 Migrate Hedge Fund Agent

**Action**: Create `internal/domains/hedge-fund-investor/agent.go`

**Source**: `hedge-fund-investor-source/web/internal/hf-agent/hf_dsl_agent.go`

**Changes**:
- Remove domain-specific session management (use shared)
- Use shared dictionary service
- Use shared parser for validation
- Return standardized `GenerationResponse`

### 3.4 Migrate Hedge Fund Vocabulary

**Action**: Create `internal/domains/hedge-fund-investor/vocab.go`

**Source**: `hedge-fund-investor-source/hf-investor/dsl/hedge_fund_dsl.go`

**Changes**:
- Convert to `registry.Vocabulary` format
- Register all 17 hedge fund verbs
- Include state transitions

### 3.5 Create Hedge Fund Verb Validator

**Action**: Create `internal/domains/hedge-fund-investor/validator.go`

```go
package hedgefund

var approvedVerbs = map[string]bool{
    "investor.start-opportunity":  true,
    "investor.amend-details":      true,
    "investor.record-indication":  true,
    "kyc.begin":                   true,
    "kyc.collect-doc":             true,
    "kyc.approve":                 true,
    "subscription.submit":         true,
    "subscription.approve":        true,
    "register.issue":              true,
    "register.record-position":    true,
    "redemption.request":          true,
    "redemption.approve":          true,
    "register.redeem":             true,
    "transfer.initiate":           true,
    "transfer.approve":            true,
    "transfer.complete":           true,
    "investor.terminate":          true,
}

func ValidateVerbs(dsl string) error {
    // Use shared parser to extract verbs
    ast, err := parser.Parse(dsl)
    if err != nil {
        return err
    }
    
    verbs := extractVerbs(ast)
    for _, verb := range verbs {
        if !approvedVerbs[verb] {
            return fmt.Errorf("unapproved verb: %s", verb)
        }
    }
    return nil
}
```

### 3.6 Deliverables

- [ ] Hedge fund domain fully migrated to `internal/domains/hedge-fund-investor/`
- [ ] Uses all shared infrastructure
- [ ] All tests passing
- [ ] Documentation updated
- [ ] Original `hedge-fund-investor-source/` marked deprecated

---

## Phase 4: Create Onboarding Domain (Week 4)

**Goal**: Implement onboarding as an orchestration domain

### 4.1 Create Onboarding Domain Package

**Action**: Create `internal/domains/onboarding/`

### 4.2 Define Onboarding Vocabulary

**Action**: Create `internal/domains/onboarding/vocab.go`

**Source**: `internal/dsl/vocab.go` (existing onboarding verbs)

```go
package onboarding

func GetVocabulary() *registry.Vocabulary {
    return &registry.Vocabulary{
        Domain:  "onboarding",
        Version: "1.0.0",
        Verbs: map[string]registry.VerbDefinition{
            // Case Management (5 verbs)
            "case.create":   {...},
            "case.update":   {...},
            "case.validate": {...},
            "case.approve":  {...},
            "case.close":    {...},
            
            // Entity Identity (5 verbs)
            "entity.register": {...},
            "entity.classify": {...},
            "entity.link":     {...},
            "identity.verify": {...},
            "identity.attest": {...},
            
            // Product Service (5 verbs)
            "products.add":       {...},
            "products.configure": {...},
            "services.discover":  {...},
            "services.provision": {...},
            "services.activate":  {...},
            
            // KYC Compliance (6 verbs)
            "kyc.start":          {...},
            "kyc.collect":        {...},
            "kyc.verify":         {...},
            "kyc.assess":         {...},
            "compliance.screen":  {...},
            "compliance.monitor": {...},
            
            // Resource Infrastructure (5 verbs)
            "resources.plan":      {...},
            "resources.provision": {...},
            "resources.configure": {...},
            "resources.test":      {...},
            "resources.deploy":    {...},
            
            // Attribute Data (5 verbs)
            "attributes.define":  {...},
            "attributes.resolve": {...},
            "values.bind":        {...},
            "values.validate":    {...},
            "values.encrypt":     {...},
            
            // Workflow State (5 verbs)
            "workflow.transition": {...},
            "workflow.gate":       {...},
            "tasks.create":        {...},
            "tasks.assign":        {...},
            "tasks.complete":      {...},
            
            // Notification Communication (4 verbs)
            "notify.send":         {...},
            "communicate.request": {...},
            "escalate.trigger":    {...},
            "audit.log":           {...},
            
            // Integration External (4 verbs)
            "external.query":   {...},
            "external.sync":    {...},
            "api.call":         {...},
            "webhook.register": {...},
            
            // Total: 68 onboarding verbs
        },
    }
}
```

### 4.3 Create Onboarding Agent

**Action**: Create `internal/domains/onboarding/agent.go`

**Source**: `internal/agent/dsl_agent.go` (existing onboarding agent)

```go
package onboarding

type Agent struct {
    client     *genai.Client
    model      *genai.GenerativeModel
    dictionary dictionary.Service
    registry   *registry.Registry  // NEW: Can call other domains
}

func (a *Agent) GenerateDSL(ctx context.Context, req *registry.GenerationRequest) (*registry.GenerationResponse, error) {
    // Generate onboarding DSL
    // Can orchestrate calls to other domains
}
```

### 4.4 Implement Orchestration Logic

**Action**: Create `internal/domains/onboarding/orchestrator.go`

```go
package onboarding

type Orchestrator struct {
    registry *registry.Registry
}

// HandleCrossDomainOperation orchestrates operations across multiple domains
func (o *Orchestrator) HandleCrossDomainOperation(ctx context.Context, op *Operation) (*Result, error) {
    // Example: Onboarding needs to call hedge fund investor domain
    // 1. Generate onboarding DSL
    // 2. Detect need for hedge fund operations
    // 3. Route to hedge fund domain
    // 4. Accumulate both DSLs
    // 5. Return unified result
}
```

**Example**: User says "Onboard Acme Capital as a hedge fund investor"
1. Onboarding domain creates case: `(case.create (cbu.id "CBU-1234") ...)`
2. Detects "hedge fund investor" keyword
3. Routes to hedge fund domain: `(investor.start-opportunity ...)`
4. Both DSLs accumulated in session

### 4.5 Create Onboarding Verb Validator

**Action**: Create `internal/domains/onboarding/validator.go`

**Source**: `internal/agent/dsl_agent.go` - `validateDSLVerbs()`

```go
package onboarding

var approvedVerbs = map[string]bool{
    // All 68 onboarding verbs from vocab.go
    "case.create": true,
    "case.update": true,
    // ... etc
}

func ValidateVerbs(dsl string) error {
    // Similar to hedge fund validator
}
```

### 4.6 Deliverables

- [ ] Onboarding domain implementation in `internal/domains/onboarding/`
- [ ] 68 onboarding verbs registered
- [ ] Orchestration logic for cross-domain calls
- [ ] Onboarding agent using shared infrastructure
- [ ] All tests passing

---

## Phase 5: Update Web Server for Multi-Domain (Week 5)

**Goal**: Refactor web server to support domain routing

### 5.1 Update Server Structure

**Action**: Modify `hedge-fund-investor-source/web/server.go`

**Changes**:
```go
package main

import (
    "dsl-ob-poc/internal/domain-registry"
    "dsl-ob-poc/internal/domains/onboarding"
    "dsl-ob-poc/internal/domains/hedge-fund-investor"
    "dsl-ob-poc/internal/shared-dsl/dictionary"
    "dsl-ob-poc/internal/shared-dsl/session"
)

type Server struct {
    router        *mux.Router
    registry      *registry.Registry     // NEW: Domain registry
    sessionMgr    *session.Manager       // NEW: Shared session manager
    dictionary    dictionary.Service     // NEW: Shared dictionary
    upgrader      websocket.Upgrader
}

func NewServer(dictService dictionary.Service, apiKey string) (*Server, error) {
    // Create domain registry
    reg := registry.NewRegistry()
    
    // Register onboarding domain
    obDomain := onboarding.NewOnboardingDomain(apiKey, dictService, reg)
    reg.Register(obDomain)
    
    // Register hedge fund domain
    hfDomain := hedgefund.NewHedgeFundDomain(apiKey, dictService)
    reg.Register(hfDomain)
    
    return &Server{
        registry:   reg,
        sessionMgr: session.NewManager(),
        dictionary: dictService,
    }
}
```

### 5.2 Update Chat Handler

**Action**: Modify chat endpoint to use domain routing

```go
func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
    var req ChatRequest
    json.NewDecoder(r.Body).Decode(&req)
    
    // Get or create session
    sess := s.sessionMgr.GetOrCreate(req.SessionID, req.Domain)
    
    // Route to appropriate domain
    router := registry.NewRouter(s.registry)
    domain, err := router.Route(ctx, &registry.RoutingRequest{
        Message:       req.Message,
        SessionID:     sess.SessionID,
        CurrentDomain: sess.Domain,
        Context:       sess.Context,
    })
    
    // Generate DSL using domain agent
    resp, err := domain.GenerateDSL(ctx, &registry.GenerationRequest{
        Instruction:   req.Message,
        Context:       sess.Context,
        ExistingDSL:   sess.BuiltDSL,
    })
    
    // Validate verbs
    if err := domain.ValidateVerbs(resp.DSL); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // Accumulate DSL
    s.sessionMgr.AccumulateDSL(sess.SessionID, resp.DSL)
    
    // Update context
    s.sessionMgr.UpdateContext(sess.SessionID, resp.Context)
    
    // Return response
    json.NewEncoder(w).Encode(resp)
}
```

### 5.3 Add Domain Switching Endpoint

**Action**: Create `/api/switch-domain` endpoint

```go
func (s *Server) handleSwitchDomain(w http.ResponseWriter, r *http.Request) {
    var req struct {
        SessionID string `json:"session_id"`
        Domain    string `json:"domain"`
    }
    json.NewDecoder(r.Body).Decode(&req)
    
    // Verify domain exists
    if _, err := s.registry.Get(req.Domain); err != nil {
        http.Error(w, "Domain not found", http.StatusNotFound)
        return
    }
    
    // Update session domain
    sess := s.sessionMgr.Get(req.SessionID)
    sess.Domain = req.Domain
    
    json.NewEncoder(w).Encode(map[string]string{
        "message": fmt.Sprintf("Switched to %s domain", req.Domain),
    })
}
```

### 5.4 Update Frontend

**Action**: Modify `hedge-fund-investor-source/web/frontend/src/App.tsx`

**Changes**:
- Add domain selector dropdown
- Display current domain in UI
- Show domain-specific vocabulary
- Color-code DSL by domain

```typescript
interface ChatState {
  sessionId: string;
  currentDomain: string;  // NEW
  messages: Message[];
  builtDSL: string;
  context: Record<string, any>;
}

// Domain selector component
function DomainSelector({ currentDomain, onSwitch }) {
  return (
    <select value={currentDomain} onChange={(e) => onSwitch(e.target.value)}>
      <option value="onboarding">Onboarding</option>
      <option value="hedge-fund-investor">Hedge Fund Investor</option>
    </select>
  );
}
```

### 5.5 Add Multi-Domain Documentation Endpoint

**Action**: Create `/api/domains` endpoint

```go
func (s *Server) handleGetDomains(w http.ResponseWriter, r *http.Request) {
    domains := s.registry.List()
    
    var response []map[string]interface{}
    for _, name := range domains {
        domain, _ := s.registry.Get(name)
        vocab := domain.GetVocabulary()
        response = append(response, map[string]interface{}{
            "name":    name,
            "version": vocab.Version,
            "verbs":   len(vocab.Verbs),
        })
    }
    
    json.NewEncoder(w).Encode(response)
}
```

### 5.6 Deliverables

- [ ] Web server refactored for multi-domain support
- [ ] Domain routing working end-to-end
- [ ] Frontend updated with domain selector
- [ ] New API endpoints documented
- [ ] All integration tests passing

---

## Phase 6: Testing & Documentation (Week 6)

**Goal**: Comprehensive testing and documentation

### 6.1 Integration Tests

**Action**: Create `internal/integration_test.go`

**Test scenarios**:
1. **Single domain workflow**: Complete onboarding case start to finish
2. **Single domain workflow**: Complete hedge fund investor lifecycle
3. **Cross-domain workflow**: Onboarding orchestrates hedge fund operations
4. **Domain switching**: Switch domains mid-conversation
5. **Shared dictionary**: Both domains reference same attributes
6. **Verb validation**: Each domain validates its own verbs
7. **DSL accumulation**: DSL builds correctly across domains
8. **Context tracking**: Entities tracked across domain switches

### 6.2 Performance Tests

**Test**:
- Domain routing latency (< 10ms)
- Dictionary lookup latency (< 5ms)
- DSL parser performance (< 50ms for large DSL)
- Session manager concurrency (1000+ concurrent sessions)

### 6.3 Documentation

**Create**:
1. **Architecture Guide** - `MULTI_DOMAIN_ARCHITECTURE.md`
   - Shared infrastructure components
   - Domain interface specification
   - Routing strategies
   - Dictionary design

2. **Domain Developer Guide** - `DOMAIN_DEVELOPMENT_GUIDE.md`
   - How to create a new domain
   - Vocabulary definition
   - Agent implementation
   - Verb validation
   - Testing checklist

3. **API Documentation** - `API_REFERENCE.md`
   - All endpoints with examples
   - Domain-specific operations
   - WebSocket protocol
   - Error handling

4. **Migration Guide** - `MIGRATION_GUIDE.md`
   - What changed from single-domain
   - Code examples
   - Breaking changes
   - Rollback procedures

### 6.4 Example Workflows

**Create**: `examples/multi-domain/`

```
examples/multi-domain/
├── onboarding-only.md          # Pure onboarding workflow
├── hedge-fund-only.md          # Pure hedge fund workflow
├── orchestrated-workflow.md    # Onboarding calls hedge fund
├── domain-switching.md         # User switches domains
└── cross-domain-attributes.md  # Shared dictionary usage
```

### 6.5 Deliverables

- [ ] 50+ integration tests covering all scenarios
- [ ] Performance benchmarks documented
- [ ] Complete architecture documentation
- [ ] Domain developer guide with examples
- [ ] API reference with curl examples
- [ ] Migration guide for existing code

---

## Rollback Strategy

### Rollback Points

Each phase has a rollback checkpoint:

1. **Phase 1 Complete**: Shared infrastructure in place, hedge fund still works
2. **Phase 2 Complete**: Domain registry exists, hedge fund still standalone
3. **Phase 3 Complete**: Hedge fund migrated, can rollback to Phase 2
4. **Phase 4 Complete**: Onboarding domain added, can disable via feature flag
5. **Phase 5 Complete**: Web server multi-domain, can fallback to single domain
6. **Phase 6 Complete**: Fully tested, production-ready

### Rollback Procedure

1. **Feature flags**: Each domain has a feature flag in config
2. **Database isolation**: Each domain's data in separate schema
3. **API versioning**: `/api/v1/` (single domain) vs `/api/v2/` (multi-domain)
4. **Graceful degradation**: If domain unavailable, fallback to error message

```go
// Feature flag example
type Config struct {
    EnableMultiDomain bool
    EnabledDomains    []string  // ["onboarding", "hedge-fund-investor"]
}

func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
    if !s.config.EnableMultiDomain {
        // Fallback to single-domain mode
        s.handleChatSingleDomain(w, r)
        return
    }
    
    // Multi-domain logic
    s.handleChatMultiDomain(w, r)
}
```

---

## Success Metrics

### Functional Metrics
- ✅ All onboarding verbs (68) working in onboarding domain
- ✅ All hedge fund verbs (17) working in hedge fund domain
- ✅ Cross-domain orchestration working (onboarding → hedge fund)
- ✅ Shared dictionary queried by both domains
- ✅ DSL parser handles all domain DSLs
- ✅ Domain switching mid-conversation works
- ✅ Context tracked across domain switches

### Performance Metrics
- ✅ Domain routing < 10ms
- ✅ Dictionary lookup < 5ms
- ✅ DSL parsing < 50ms (for 100+ line DSL)
- ✅ Session manager supports 1000+ concurrent sessions
- ✅ No performance degradation vs single-domain

### Quality Metrics
- ✅ 80%+ code coverage across all packages
- ✅ Zero critical security vulnerabilities
- ✅ All linters passing (golangci-lint)
- ✅ Documentation complete for all public APIs

---

## Timeline Summary

| Phase | Week | Deliverable | Dependencies |
|-------|------|-------------|--------------|
| 1 | Week 1 | Shared infrastructure extracted | None |
| 2 | Week 2 | Domain registry system | Phase 1 |
| 3 | Week 3 | Hedge fund domain migrated | Phase 1, 2 |
| 4 | Week 4 | Onboarding domain created | Phase 1, 2 |
| 5 | Week 5 | Web server multi-domain | Phase 1-4 |
| 6 | Week 6 | Testing & documentation | All phases |

**Total Duration**: 6 weeks for complete migration

---

## Quick Start Guide (Post-Migration)

### Starting the Multi-Domain Server

```bash
# Set environment variables
export GEMINI_API_KEY="your-api-key"
export DB_CONN_STRING="postgres://localhost:5432/dsl-ob-poc?sslmode=disable"

# Build the server
cd hedge-fund-investor-source/web
go build -o multi-domain-server server.go

# Run the server
./multi-domain-server
```

### Using Multiple Domains

#### Example 1: Pure Onboarding Workflow

```bash
curl -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Create a new onboarding case for CBU-1234",
    "domain": "onboarding"
  }'

# Response includes DSL:
# (case.create (cbu.id "CBU-1234") (nature-purpose "..."))
```

#### Example 2: Pure Hedge Fund Workflow

```bash
curl -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Create opportunity for Acme Capital LP",
    "domain": "hedge-fund-investor"
  }'

# Response includes DSL:
# (investor.start-opportunity
#   (legal-name "Acme Capital LP")
#   (type "CORPORATE"))
```

#### Example 3: Cross-Domain Orchestration

```bash
# Start in onboarding domain
curl -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Onboard Acme Capital as a hedge fund investor",
    "domain": "onboarding"
  }'

# Onboarding domain generates:
# (case.create (cbu.id "CBU-1234") (nature-purpose "Hedge fund investor onboarding"))
# 
# Then orchestrates to hedge fund domain:
# (investor.start-opportunity (legal-name "Acme Capital LP") ...)
#
# Both DSLs accumulated in session.BuiltDSL
```

#### Example 4: Domain Switching

```bash
# Get session ID from first request
SESSION_ID="abc-123"

# Switch to different domain
curl -X POST http://localhost:8080/api/switch-domain \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "'$SESSION_ID'",
    "domain": "hedge-fund-investor"
  }'

# Continue conversation in new domain
curl -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "'$SESSION_ID'",
    "message": "Start KYC for this investor"
  }'
```

---

## Implementation Examples

### Example 1: Creating a New Domain

Let's say you want to add a "kyc" domain separate from onboarding and hedge fund:

```go
// internal/domains/kyc/domain.go
package kyc

import (
    "context"
    "dsl-ob-poc/internal/domain-registry"
    "dsl-ob-poc/internal/shared-dsl/dictionary"
)

type KYCDomain struct {
    agent      *Agent
    vocabulary *registry.Vocabulary
    dictionary dictionary.Service
}

func NewKYCDomain(apiKey string, dict dictionary.Service) (*KYCDomain, error) {
    agent, err := NewAgent(apiKey, dict)
    if err != nil {
        return nil, err
    }
    
    return &KYCDomain{
        agent:      agent,
        vocabulary: GetVocabulary(),
        dictionary: dict,
    }, nil
}

func (d *KYCDomain) Name() string {
    return "kyc"
}

func (d *KYCDomain) Version() string {
    return "1.0.0"
}

func (d *KYCDomain) GetVocabulary() *registry.Vocabulary {
    return d.vocabulary
}

func (d *KYCDomain) ValidateVerbs(dsl string) error {
    return ValidateVerbs(dsl)
}

func (d *KYCDomain) GenerateDSL(ctx context.Context, req *registry.GenerationRequest) (*registry.GenerationResponse, error) {
    return d.agent.GenerateDSL(ctx, req)
}

func (d *KYCDomain) GetCurrentState(context map[string]interface{}) string {
    if kycState, ok := context["kyc_state"].(string); ok {
        return kycState
    }
    return "INITIAL"
}

func (d *KYCDomain) ValidateTransition(from, to string) error {
    // Implement state machine validation
    return nil
}
```

```go
// internal/domains/kyc/vocab.go
package kyc

import "dsl-ob-poc/internal/domain-registry"

func GetVocabulary() *registry.Vocabulary {
    return &registry.Vocabulary{
        Domain:  "kyc",
        Version: "1.0.0",
        Verbs: map[string]registry.VerbDefinition{
            "kyc.initiate": {
                Name:        "kyc.initiate",
                Category:    "kyc-lifecycle",
                Description: "Initiate KYC process for an entity",
                Args: map[string]registry.ArgSpec{
                    "entity-id": {Type: "uuid", Required: true},
                    "tier":      {Type: "enum", Values: []string{"SIMPLIFIED", "STANDARD", "ENHANCED"}},
                },
                StateChange: &registry.StateTransition{
                    FromStates: []string{"INITIAL"},
                    ToState:    "KYC_PENDING",
                },
            },
            "kyc.collect-document": {
                Name:        "kyc.collect-document",
                Category:    "kyc-documents",
                Description: "Collect a KYC document",
                Args: map[string]registry.ArgSpec{
                    "entity-id":     {Type: "uuid", Required: true},
                    "document-type": {Type: "string", Required: true},
                    "document-id":   {Type: "uuid", Required: true},
                },
            },
            "kyc.verify-document": {
                Name:        "kyc.verify-document",
                Category:    "kyc-verification",
                Description: "Verify a collected document",
                Args: map[string]registry.ArgSpec{
                    "document-id": {Type: "uuid", Required: true},
                    "verifier-id": {Type: "uuid", Required: true},
                    "outcome":     {Type: "enum", Values: []string{"APPROVED", "REJECTED", "NEEDS_INFO"}},
                },
            },
            "kyc.screen-entity": {
                Name:        "kyc.screen-entity",
                Category:    "kyc-screening",
                Description: "Screen entity against sanctions/PEP lists",
                Args: map[string]registry.ArgSpec{
                    "entity-id": {Type: "uuid", Required: true},
                    "lists":     {Type: "array", Required: true},
                },
            },
            "kyc.approve": {
                Name:        "kyc.approve",
                Category:    "kyc-lifecycle",
                Description: "Approve KYC process",
                Args: map[string]registry.ArgSpec{
                    "entity-id":   {Type: "uuid", Required: true},
                    "approver-id": {Type: "uuid", Required: true},
                },
                StateChange: &registry.StateTransition{
                    FromStates: []string{"KYC_PENDING", "KYC_REVIEW"},
                    ToState:    "KYC_APPROVED",
                },
            },
        },
    }
}
```

```go
// Register the new domain in server.go
func NewServer(dictService dictionary.Service, apiKey string) (*Server, error) {
    reg := registry.NewRegistry()
    
    // Register onboarding domain
    obDomain := onboarding.NewOnboardingDomain(apiKey, dictService, reg)
    reg.Register(obDomain)
    
    // Register hedge fund domain
    hfDomain := hedgefund.NewHedgeFundDomain(apiKey, dictService)
    reg.Register(hfDomain)
    
    // Register KYC domain (NEW)
    kycDomain := kyc.NewKYCDomain(apiKey, dictService)
    reg.Register(kycDomain)
    
    return &Server{
        registry:   reg,
        sessionMgr: session.NewManager(),
        dictionary: dictService,
    }
}
```

### Example 2: Shared Dictionary Usage

Both onboarding and hedge fund domains reference the same attributes:

```sql
-- Shared attribute in dictionary table
INSERT INTO "dsl-ob-poc".dictionary (
    attribute_id,
    name,
    long_description,
    group_id,
    mask,
    domain,
    source,
    sink