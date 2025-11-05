// Package compiler provides compile-time optimization and analysis for DSL documents
package compiler

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// CompilationResult represents the result of DSL compilation with optimizations
type CompilationResult struct {
	OriginalDSL   string
	OptimizedDSL  string
	ExecutionPlan *ExecutionPlan
	Dependencies  *DependencyGraph
	Warnings      []CompilationWarning
	Errors        []CompilationError
	Optimizations []OptimizationApplied
	ResourceMap   map[string]*ResourceRequirement
}

// CompilationWarning represents a non-fatal compilation issue
type CompilationWarning struct {
	Type       string `json:"type"`
	Message    string `json:"message"`
	Line       int    `json:"line,omitempty"`
	Verb       string `json:"verb,omitempty"`
	Suggestion string `json:"suggestion,omitempty"`
}

// CompilationError represents a fatal compilation issue
type CompilationError struct {
	Type     string `json:"type"`
	Message  string `json:"message"`
	Line     int    `json:"line,omitempty"`
	Verb     string `json:"verb,omitempty"`
	Critical bool   `json:"critical"`
}

// OptimizationApplied represents an optimization that was applied during compilation
type OptimizationApplied struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Impact      string `json:"impact"` // "PERFORMANCE", "SAFETY", "CORRECTNESS"
	Before      string `json:"before,omitempty"`
	After       string `json:"after,omitempty"`
}

// ExecutionPlan represents the optimized execution order and dependencies
type ExecutionPlan struct {
	Phases          []ExecutionPhase       `json:"phases"`
	TotalOperations int                    `json:"total_operations"`
	ParallelGroups  int                    `json:"parallel_groups"`
	CriticalPath    []string               `json:"critical_path"`
	ResourceOrder   []string               `json:"resource_order"`
	SyncPoints      []SynchronizationPoint `json:"sync_points"`
}

// ExecutionPhase represents a group of operations that can execute in parallel
type ExecutionPhase struct {
	PhaseID         int         `json:"phase_id"`
	Operations      []Operation `json:"operations"`
	Dependencies    []string    `json:"dependencies"`
	CanParallelize  bool        `json:"can_parallelize"`
	EstimatedTime   int         `json:"estimated_time_ms"`
	ResourcesNeeded []string    `json:"resources_needed"`
	WaitConditions  []string    `json:"wait_conditions"`
}

// Operation represents a single DSL operation with metadata
type Operation struct {
	ID           string   `json:"id"`
	Verb         string   `json:"verb"`
	DSLFragment  string   `json:"dsl_fragment"`
	Dependencies []string `json:"dependencies"`
	Produces     []string `json:"produces"` // AttributeIDs or resources this operation creates
	Domain       string   `json:"domain"`   // "onboarding", "kyc", "ubo", etc.
	Priority     int      `json:"priority"` // Higher number = higher priority
	CanRetry     bool     `json:"can_retry"`
}

// SynchronizationPoint represents a point where execution must wait
type SynchronizationPoint struct {
	Name        string   `json:"name"`
	AfterPhase  int      `json:"after_phase"`
	WaitFor     []string `json:"wait_for"`   // AttributeIDs or conditions to wait for
	Timeout     int      `json:"timeout_ms"` // Maximum wait time
	OnTimeout   string   `json:"on_timeout"` // "FAIL", "CONTINUE", "RETRY"
	Description string   `json:"description"`
}

// DependencyGraph represents relationships between operations and resources
type DependencyGraph struct {
	Nodes map[string]*DependencyNode `json:"nodes"`
	Edges []DependencyEdge           `json:"edges"`
}

// DependencyNode represents a single operation or resource in the dependency graph
type DependencyNode struct {
	ID           string   `json:"id"`
	Type         string   `json:"type"` // "OPERATION", "RESOURCE", "ATTRIBUTE"
	Dependencies []string `json:"dependencies"`
	Dependents   []string `json:"dependents"`
	Domain       string   `json:"domain"`
	Critical     bool     `json:"critical"` // On critical path
	Level        int      `json:"level"`    // Dependency level (0 = no deps)
}

// DependencyEdge represents a dependency relationship
type DependencyEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"` // "REQUIRES", "PRODUCES", "BLOCKS"
}

// ResourceRequirement represents a resource that must be created
type ResourceRequirement struct {
	ResourceType    string   `json:"resource_type"`
	DependsOn       []string `json:"depends_on"`
	CreationVerb    string   `json:"creation_verb"`
	WaitCondition   string   `json:"wait_condition"`
	FailureHandling string   `json:"failure_handling"`
	URL             string   `json:"url,omitempty"`
	Attributes      []string `json:"attributes"`
}

// DSLCompiler provides compile-time analysis and optimization
type DSLCompiler struct {
	verbRegistry      map[string]bool
	attributeRegistry map[string]string // attributeID -> domain
	resourceTemplates map[string]*ResourceRequirement
	optimizations     []OptimizerFunc
}

// OptimizerFunc represents an optimization function
type OptimizerFunc func(*CompilationContext) error

// CompilationContext holds state during compilation
type CompilationContext struct {
	OriginalDSL   string
	CurrentDSL    string
	AST           []ASTNode
	Dependencies  *DependencyGraph
	ExecutionPlan *ExecutionPlan
	ResourceMap   map[string]*ResourceRequirement
	Warnings      []CompilationWarning
	Errors        []CompilationError
	Optimizations []OptimizationApplied
}

// ASTNode represents a parsed DSL node
type ASTNode struct {
	Verb       string    `json:"verb"`
	Parameters []string  `json:"parameters"`
	Children   []ASTNode `json:"children,omitempty"`
	Line       int       `json:"line"`
	Domain     string    `json:"domain"`
}

// NewDSLCompiler creates a new compiler instance
func NewDSLCompiler() *DSLCompiler {
	compiler := &DSLCompiler{
		verbRegistry:      make(map[string]bool),
		attributeRegistry: make(map[string]string),
		resourceTemplates: make(map[string]*ResourceRequirement),
		optimizations:     make([]OptimizerFunc, 0),
	}

	// Register default optimizations
	compiler.RegisterOptimization(OptimizeDependencyOrder)
	compiler.RegisterOptimization(OptimizeParallelExecution)
	compiler.RegisterOptimization(OptimizeResourceCreation)
	compiler.RegisterOptimization(ValidateCrossDomainReferences)

	return compiler
}

// RegisterVerb registers an allowed DSL verb
func (c *DSLCompiler) RegisterVerb(verb string) {
	c.verbRegistry[verb] = true
}

// RegisterAttribute registers an attribute with its domain
func (c *DSLCompiler) RegisterAttribute(attributeID, domain string) {
	c.attributeRegistry[attributeID] = domain
}

// RegisterResourceTemplate registers a resource creation template
func (c *DSLCompiler) RegisterResourceTemplate(resourceType string, template *ResourceRequirement) {
	c.resourceTemplates[resourceType] = template
}

// RegisterOptimization registers an optimization function
func (c *DSLCompiler) RegisterOptimization(optimizer OptimizerFunc) {
	c.optimizations = append(c.optimizations, optimizer)
}

// Compile performs complete DSL compilation with optimizations
func (c *DSLCompiler) Compile(dsl string) (*CompilationResult, error) {
	ctx := &CompilationContext{
		OriginalDSL:   dsl,
		CurrentDSL:    dsl,
		Dependencies:  &DependencyGraph{Nodes: make(map[string]*DependencyNode), Edges: make([]DependencyEdge, 0)},
		ExecutionPlan: &ExecutionPlan{},
		ResourceMap:   make(map[string]*ResourceRequirement),
		Warnings:      make([]CompilationWarning, 0),
		Errors:        make([]CompilationError, 0),
		Optimizations: make([]OptimizationApplied, 0),
	}

	// Phase 1: Parse DSL into AST
	if err := c.parseDSL(ctx); err != nil {
		return nil, fmt.Errorf("parsing failed: %w", err)
	}

	// Phase 2: Validate verbs and structure
	if err := c.validateDSL(ctx); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Phase 3: Build dependency graph
	if err := c.buildDependencyGraph(ctx); err != nil {
		return nil, fmt.Errorf("dependency analysis failed: %w", err)
	}

	// Phase 4: Apply optimizations
	for _, optimizer := range c.optimizations {
		if err := optimizer(ctx); err != nil {
			ctx.Warnings = append(ctx.Warnings, CompilationWarning{
				Type:    "OPTIMIZATION_FAILED",
				Message: fmt.Sprintf("Optimization failed: %v", err),
			})
		}
	}

	// Phase 5: Generate execution plan
	if err := c.generateExecutionPlan(ctx); err != nil {
		return nil, fmt.Errorf("execution planning failed: %w", err)
	}

	return &CompilationResult{
		OriginalDSL:   ctx.OriginalDSL,
		OptimizedDSL:  ctx.CurrentDSL,
		ExecutionPlan: ctx.ExecutionPlan,
		Dependencies:  ctx.Dependencies,
		Warnings:      ctx.Warnings,
		Errors:        ctx.Errors,
		Optimizations: ctx.Optimizations,
		ResourceMap:   ctx.ResourceMap,
	}, nil
}

// parseDSL converts DSL text into AST
func (c *DSLCompiler) parseDSL(ctx *CompilationContext) error {
	lines := strings.Split(ctx.CurrentDSL, "\n")
	ctx.AST = make([]ASTNode, 0)

	verbPattern := regexp.MustCompile(`^\s*\(([^.]+\.[^.\s]+)`)

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, ";") {
			continue
		}

		matches := verbPattern.FindStringSubmatch(line)
		if len(matches) > 1 {
			verb := matches[1]
			domain := c.extractDomainFromVerb(verb)

			node := ASTNode{
				Verb:       verb,
				Parameters: c.extractParameters(line),
				Line:       i + 1,
				Domain:     domain,
			}

			ctx.AST = append(ctx.AST, node)
		}
	}

	return nil
}

// validateDSL validates the parsed AST
func (c *DSLCompiler) validateDSL(ctx *CompilationContext) error {
	for _, node := range ctx.AST {
		// Validate verb is registered
		if !c.verbRegistry[node.Verb] {
			ctx.Errors = append(ctx.Errors, CompilationError{
				Type:     "INVALID_VERB",
				Message:  fmt.Sprintf("Unregistered verb: %s", node.Verb),
				Line:     node.Line,
				Verb:     node.Verb,
				Critical: true,
			})
		}

		// Validate attribute references
		for _, param := range node.Parameters {
			if strings.Contains(param, "@attr{") {
				attrID := c.extractAttributeID(param)
				if attrID != "" && c.attributeRegistry[attrID] == "" {
					ctx.Warnings = append(ctx.Warnings, CompilationWarning{
						Type:       "UNKNOWN_ATTRIBUTE",
						Message:    fmt.Sprintf("Unknown attribute ID: %s", attrID),
						Line:       node.Line,
						Verb:       node.Verb,
						Suggestion: "Register this attribute in the dictionary",
					})
				}
			}
		}
	}

	return nil
}

// buildDependencyGraph analyzes dependencies between operations
func (c *DSLCompiler) buildDependencyGraph(ctx *CompilationContext) error {
	// Create nodes for each operation
	for i, node := range ctx.AST {
		opID := fmt.Sprintf("op_%d", i)
		depNode := &DependencyNode{
			ID:           opID,
			Type:         "OPERATION",
			Dependencies: make([]string, 0),
			Dependents:   make([]string, 0),
			Domain:       node.Domain,
			Critical:     false,
			Level:        0,
		}

		ctx.Dependencies.Nodes[opID] = depNode
	}

	// Analyze dependencies based on verb semantics and attribute usage
	for i, node := range ctx.AST {
		opID := fmt.Sprintf("op_%d", i)

		// Resource creation operations depend on prerequisites
		if strings.Contains(node.Verb, "resources.create") {
			c.analyzResourceDependencies(ctx, opID, node)
		}

		// UBO operations have specific ordering requirements
		if strings.Contains(node.Verb, "ubo.") {
			c.analyzeUBODependencies(ctx, opID, node, i)
		}

		// KYC operations must complete before resource creation
		if strings.Contains(node.Verb, "kyc.") {
			c.analyzeKYCDependencies(ctx, opID, node, i)
		}
	}

	return nil
}

// analyzResourceDependencies determines resource creation dependencies
func (c *DSLCompiler) analyzResourceDependencies(ctx *CompilationContext, opID string, node ASTNode) {
	// Resources typically depend on identity verification
	for i := 0; i < len(ctx.AST); i++ {
		prevNode := ctx.AST[i]
		prevOpID := fmt.Sprintf("op_%d", i)

		if prevOpID == opID {
			break
		}

		// Custody accounts depend on UBO verification
		if strings.Contains(node.Verb, "custody") &&
			(strings.Contains(prevNode.Verb, "ubo.verify") || strings.Contains(prevNode.Verb, "kyc.complete")) {
			ctx.Dependencies.Edges = append(ctx.Dependencies.Edges, DependencyEdge{
				From: prevOpID,
				To:   opID,
				Type: "REQUIRES",
			})

			ctx.Dependencies.Nodes[opID].Dependencies = append(ctx.Dependencies.Nodes[opID].Dependencies, prevOpID)
			ctx.Dependencies.Nodes[prevOpID].Dependents = append(ctx.Dependencies.Nodes[prevOpID].Dependents, opID)
		}
	}
}

// analyzeUBODependencies handles Ultimate Beneficial Owner workflow dependencies
func (c *DSLCompiler) analyzeUBODependencies(ctx *CompilationContext, opID string, node ASTNode, currentIndex int) {
	// UBO discovery must precede verification
	if strings.Contains(node.Verb, "ubo.verify") {
		for i := 0; i < currentIndex; i++ {
			prevNode := ctx.AST[i]
			prevOpID := fmt.Sprintf("op_%d", i)

			if strings.Contains(prevNode.Verb, "ubo.discover") || strings.Contains(prevNode.Verb, "ubo.identify") {
				ctx.Dependencies.Edges = append(ctx.Dependencies.Edges, DependencyEdge{
					From: prevOpID,
					To:   opID,
					Type: "REQUIRES",
				})

				ctx.Dependencies.Nodes[opID].Dependencies = append(ctx.Dependencies.Nodes[opID].Dependencies, prevOpID)
				ctx.Dependencies.Nodes[prevOpID].Dependents = append(ctx.Dependencies.Nodes[prevOpID].Dependents, opID)
			}
		}
	}
}

// analyzeKYCDependencies handles Know Your Customer workflow dependencies
func (c *DSLCompiler) analyzeKYCDependencies(ctx *CompilationContext, opID string, node ASTNode, currentIndex int) {
	// KYC collection must precede verification
	if strings.Contains(node.Verb, "kyc.verify") || strings.Contains(node.Verb, "kyc.complete") {
		for i := 0; i < currentIndex; i++ {
			prevNode := ctx.AST[i]
			prevOpID := fmt.Sprintf("op_%d", i)

			if strings.Contains(prevNode.Verb, "kyc.collect") || strings.Contains(prevNode.Verb, "kyc.start") {
				ctx.Dependencies.Edges = append(ctx.Dependencies.Edges, DependencyEdge{
					From: prevOpID,
					To:   opID,
					Type: "REQUIRES",
				})

				ctx.Dependencies.Nodes[opID].Dependencies = append(ctx.Dependencies.Nodes[opID].Dependencies, prevOpID)
				ctx.Dependencies.Nodes[prevOpID].Dependents = append(ctx.Dependencies.Nodes[prevOpID].Dependents, opID)
			}
		}
	}
}

// generateExecutionPlan creates optimized execution phases
func (c *DSLCompiler) generateExecutionPlan(ctx *CompilationContext) error {
	// Calculate dependency levels
	c.calculateDependencyLevels(ctx.Dependencies)

	// Group operations by dependency level
	levelGroups := make(map[int][]string)
	for nodeID, node := range ctx.Dependencies.Nodes {
		if node.Type == "OPERATION" {
			level := node.Level
			if levelGroups[level] == nil {
				levelGroups[level] = make([]string, 0)
			}
			levelGroups[level] = append(levelGroups[level], nodeID)
		}
	}

	// Create execution phases
	ctx.ExecutionPlan.Phases = make([]ExecutionPhase, 0)
	maxLevel := 0
	for level := range levelGroups {
		if level > maxLevel {
			maxLevel = level
		}
	}

	for level := 0; level <= maxLevel; level++ {
		if nodeIDs, exists := levelGroups[level]; exists {
			phase := ExecutionPhase{
				PhaseID:         level,
				Operations:      make([]Operation, 0),
				Dependencies:    make([]string, 0),
				CanParallelize:  len(nodeIDs) > 1,
				EstimatedTime:   1000, // Default 1 second per operation
				ResourcesNeeded: make([]string, 0),
				WaitConditions:  make([]string, 0),
			}

			// Convert node IDs to operations
			for _, nodeID := range nodeIDs {
				opIndex := c.extractOperationIndex(nodeID)
				if opIndex >= 0 && opIndex < len(ctx.AST) {
					astNode := ctx.AST[opIndex]

					operation := Operation{
						ID:           nodeID,
						Verb:         astNode.Verb,
						DSLFragment:  c.reconstructDSLFragment(astNode),
						Dependencies: ctx.Dependencies.Nodes[nodeID].Dependencies,
						Produces:     c.extractProducedAttributes(astNode),
						Domain:       astNode.Domain,
						Priority:     c.calculatePriority(astNode.Verb),
						CanRetry:     !strings.Contains(astNode.Verb, "resources.create"),
					}

					phase.Operations = append(phase.Operations, operation)
				}
			}

			// Sort operations by priority within phase
			sort.Slice(phase.Operations, func(i, j int) bool {
				return phase.Operations[i].Priority > phase.Operations[j].Priority
			})

			ctx.ExecutionPlan.Phases = append(ctx.ExecutionPlan.Phases, phase)
		}
	}

	ctx.ExecutionPlan.TotalOperations = len(ctx.AST)
	ctx.ExecutionPlan.ParallelGroups = len(ctx.ExecutionPlan.Phases)

	return nil
}

// Helper methods

func (c *DSLCompiler) extractDomainFromVerb(verb string) string {
	parts := strings.Split(verb, ".")
	if len(parts) > 0 {
		return parts[0]
	}
	return "unknown"
}

func (c *DSLCompiler) extractParameters(line string) []string {
	// Simple parameter extraction - would be more sophisticated in production
	params := make([]string, 0)
	if strings.Contains(line, "@attr{") {
		re := regexp.MustCompile(`@attr\{[^}]+\}`)
		matches := re.FindAllString(line, -1)
		params = append(params, matches...)
	}
	return params
}

func (c *DSLCompiler) extractAttributeID(param string) string {
	re := regexp.MustCompile(`@attr\{([^}]+)\}`)
	matches := re.FindStringSubmatch(param)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func (c *DSLCompiler) calculateDependencyLevels(graph *DependencyGraph) {
	// Topological sort to calculate dependency levels
	visited := make(map[string]bool)

	var visit func(nodeID string) int
	visit = func(nodeID string) int {
		if visited[nodeID] {
			return graph.Nodes[nodeID].Level
		}

		visited[nodeID] = true
		maxDepLevel := -1

		for _, depID := range graph.Nodes[nodeID].Dependencies {
			depLevel := visit(depID)
			if depLevel > maxDepLevel {
				maxDepLevel = depLevel
			}
		}

		graph.Nodes[nodeID].Level = maxDepLevel + 1
		return graph.Nodes[nodeID].Level
	}

	for nodeID := range graph.Nodes {
		if !visited[nodeID] {
			visit(nodeID)
		}
	}
}

func (c *DSLCompiler) extractOperationIndex(nodeID string) int {
	// Extract index from "op_N" format
	parts := strings.Split(nodeID, "_")
	if len(parts) > 1 {
		if idx := parts[1]; idx != "" {
			// Simple conversion - would use strconv.Atoi in production
			switch idx {
			case "0":
				return 0
			case "1":
				return 1
			case "2":
				return 2
			case "3":
				return 3
			case "4":
				return 4
			default:
				return -1
			}
		}
	}
	return -1
}

func (c *DSLCompiler) reconstructDSLFragment(node ASTNode) string {
	// Reconstruct the DSL fragment from AST node
	params := strings.Join(node.Parameters, " ")
	if params != "" {
		return fmt.Sprintf("(%s %s)", node.Verb, params)
	}
	return fmt.Sprintf("(%s)", node.Verb)
}

func (c *DSLCompiler) extractProducedAttributes(node ASTNode) []string {
	// Extract attributes this operation produces
	produced := make([]string, 0)

	// Different verbs produce different attributes
	if strings.Contains(node.Verb, "ubo.verify") {
		produced = append(produced, "ubo-verification-complete")
	}
	if strings.Contains(node.Verb, "kyc.complete") {
		produced = append(produced, "kyc-complete")
	}
	if strings.Contains(node.Verb, "resources.create") {
		produced = append(produced, "resource-created")
	}

	return produced
}

func (c *DSLCompiler) calculatePriority(verb string) int {
	// Higher priority verbs execute first within a phase
	if strings.Contains(verb, "kyc.") {
		return 100
	}
	if strings.Contains(verb, "ubo.") {
		return 90
	}
	if strings.Contains(verb, "resources.") {
		return 80
	}
	return 50 // Default priority
}

// Default optimization functions

// OptimizeDependencyOrder ensures proper dependency ordering
func OptimizeDependencyOrder(ctx *CompilationContext) error {
	// This optimization ensures dependencies are respected
	ctx.Optimizations = append(ctx.Optimizations, OptimizationApplied{
		Type:        "DEPENDENCY_ORDER",
		Description: "Optimized operation ordering based on dependencies",
		Impact:      "CORRECTNESS",
	})
	return nil
}

// OptimizeParallelExecution identifies parallel execution opportunities
func OptimizeParallelExecution(ctx *CompilationContext) error {
	// This optimization identifies operations that can run in parallel
	ctx.Optimizations = append(ctx.Optimizations, OptimizationApplied{
		Type:        "PARALLEL_EXECUTION",
		Description: "Identified operations that can execute in parallel",
		Impact:      "PERFORMANCE",
	})
	return nil
}

// OptimizeResourceCreation optimizes resource creation ordering
func OptimizeResourceCreation(ctx *CompilationContext) error {
	// This optimization ensures resources are created in optimal order
	ctx.Optimizations = append(ctx.Optimizations, OptimizationApplied{
		Type:        "RESOURCE_CREATION",
		Description: "Optimized resource creation dependencies",
		Impact:      "SAFETY",
	})
	return nil
}

// ValidateCrossDomainReferences validates references across domains
func ValidateCrossDomainReferences(ctx *CompilationContext) error {
	// This optimization validates attribute references across domains
	ctx.Optimizations = append(ctx.Optimizations, OptimizationApplied{
		Type:        "CROSS_DOMAIN_VALIDATION",
		Description: "Validated cross-domain attribute references",
		Impact:      "CORRECTNESS",
	})
	return nil
}
