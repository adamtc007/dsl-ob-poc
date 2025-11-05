// Package optimizer provides dependency analysis and optimization for DSL compilation
package optimizer

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// DependencyOptimizer analyzes and optimizes DSL dependencies
type DependencyOptimizer struct {
	AttributeGraph     *AttributeDependencyGraph
	ResourceTracker    *ResourceDependencyTracker
	ExecutionPlanner   *OptimizedExecutionPlanner
	verbDependencies   map[string][]string // verb -> required preceding verbs
	attributeProducers map[string]string   // attributeID -> verb that produces it
	criticalPaths      []string            // operations on critical path
}

// AttributeDependencyGraph tracks attribute dependencies across domains
type AttributeDependencyGraph struct {
	Attributes map[string]*AttributeNode `json:"attributes"`
	Edges      []AttributeEdge           `json:"edges"`
	Domains    map[string][]string       `json:"domains"` // domain -> attributeIDs
}

// AttributeNode represents an attribute in the dependency graph
type AttributeNode struct {
	AttributeID  string   `json:"attribute_id"`
	Domain       string   `json:"domain"`
	Type         string   `json:"type"` // "INPUT", "COMPUTED", "OUTPUT"
	ProducedBy   []string `json:"produced_by"`
	RequiredBy   []string `json:"required_by"`
	Dependencies []string `json:"dependencies"`
	CriticalPath bool     `json:"critical_path"`
	ComputeCost  int      `json:"compute_cost"` // Relative cost to compute/fetch
}

// AttributeEdge represents a dependency between attributes
type AttributeEdge struct {
	From        string `json:"from"`
	To          string `json:"to"`
	Type        string `json:"type"`     // "REQUIRES", "PRODUCES", "DERIVED_FROM"
	Strength    int    `json:"strength"` // 1=weak, 10=critical
	CrossDomain bool   `json:"cross_domain"`
}

// ResourceDependencyTracker manages resource creation dependencies
type ResourceDependencyTracker struct {
	Resources        map[string]*ResourceDependency `json:"resources"`
	CreationOrder    []string                       `json:"creation_order"`
	Blockers         map[string][]string            `json:"blockers"`      // resource -> blocking conditions
	Prerequisites    map[string][]string            `json:"prerequisites"` // resource -> required attributes
	FailureScenarios map[string]*FailureRecovery    `json:"failure_scenarios"`
}

// ResourceDependency represents dependencies for resource creation
type ResourceDependency struct {
	ResourceType    string   `json:"resource_type"`    // "CUSTODY_ACCOUNT", "SIGNATORY_AUTHORITY"
	DependsOn       []string `json:"depends_on"`       // "@attr{ubo-verification-complete}"
	CreationVerb    string   `json:"creation_verb"`    // "(resources.create-custody-account ...)"
	WaitCondition   string   `json:"wait_condition"`   // "UBO_IDENTITY_VERIFIED"
	FailureHandling string   `json:"failure_handling"` // "ROLLBACK_PARTIAL_RESOURCES"
	EstimatedTime   int      `json:"estimated_time_ms"`
	RetryPolicy     string   `json:"retry_policy"` // "EXPONENTIAL_BACKOFF", "FIXED_DELAY", "NO_RETRY"
	Priority        int      `json:"priority"`     // Higher number = higher priority
}

// FailureRecovery defines how to handle resource creation failures
type FailureRecovery struct {
	Strategy   string   `json:"strategy"` // "ROLLBACK", "CONTINUE", "MANUAL_INTERVENTION"
	Rollback   []string `json:"rollback"` // Operations to execute on rollback
	Notify     []string `json:"notify"`   // Who to notify on failure
	MaxRetries int      `json:"max_retries"`
	Timeout    int      `json:"timeout_ms"`
}

// OptimizedExecutionPlanner creates optimized execution plans
type OptimizedExecutionPlanner struct {
	Phases                []OptimizedPhase `json:"phases"`
	ParallelGroups        map[int][]string `json:"parallel_groups"` // phase -> operation IDs
	SynchronizationPoints []SyncPoint      `json:"sync_points"`
	CriticalPathLength    int              `json:"critical_path_length_ms"`
	TotalEstimatedTime    int              `json:"total_estimated_time_ms"`
}

// OptimizedPhase represents an execution phase with optimization metadata
type OptimizedPhase struct {
	PhaseID           int                `json:"phase_id"`
	Operations        []OptimizedOp      `json:"operations"`
	Dependencies      []string           `json:"dependencies"`
	Parallelizable    bool               `json:"parallelizable"`
	EstimatedTime     int                `json:"estimated_time_ms"`
	ResourcesNeeded   []string           `json:"resources_needed"`
	WaitConditions    []WaitCondition    `json:"wait_conditions"`
	OptimizationHints []OptimizationHint `json:"optimization_hints"`
}

// OptimizedOp represents an operation with optimization metadata
type OptimizedOp struct {
	ID             string   `json:"id"`
	Verb           string   `json:"verb"`
	Domain         string   `json:"domain"`
	Dependencies   []string `json:"dependencies"`
	Produces       []string `json:"produces"`
	EstimatedTime  int      `json:"estimated_time_ms"`
	CanParallelize bool     `json:"can_parallelize"`
	CriticalPath   bool     `json:"critical_path"`
	RetryPolicy    string   `json:"retry_policy"`
	Priority       int      `json:"priority"`
}

// WaitCondition represents a condition that must be met before proceeding
type WaitCondition struct {
	Name        string `json:"name"`
	Type        string `json:"type"`   // "ATTRIBUTE_AVAILABLE", "RESOURCE_CREATED", "EXTERNAL_APPROVAL"
	Target      string `json:"target"` // AttributeID or resource identifier
	Timeout     int    `json:"timeout_ms"`
	OnTimeout   string `json:"on_timeout"` // "FAIL", "CONTINUE", "RETRY"
	Description string `json:"description"`
}

// SyncPoint represents a synchronization point in execution
type SyncPoint struct {
	Name          string   `json:"name"`
	AfterPhase    int      `json:"after_phase"`
	WaitFor       []string `json:"wait_for"`
	TimeoutMs     int      `json:"timeout_ms"`
	OnTimeout     string   `json:"on_timeout"`
	Description   string   `json:"description"`
	CriticalPoint bool     `json:"critical_point"`
}

// OptimizationHint provides guidance for execution optimization
type OptimizationHint struct {
	Type        string `json:"type"` // "PARALLEL_SAFE", "CPU_INTENSIVE", "IO_BOUND", "NETWORK_DEPENDENT"
	Description string `json:"description"`
	Priority    int    `json:"priority"`
}

// NewDependencyOptimizer creates a new dependency optimizer
func NewDependencyOptimizer() *DependencyOptimizer {
	return &DependencyOptimizer{
		AttributeGraph: &AttributeDependencyGraph{
			Attributes: make(map[string]*AttributeNode),
			Edges:      make([]AttributeEdge, 0),
			Domains:    make(map[string][]string),
		},
		ResourceTracker: &ResourceDependencyTracker{
			Resources:        make(map[string]*ResourceDependency),
			CreationOrder:    make([]string, 0),
			Blockers:         make(map[string][]string),
			Prerequisites:    make(map[string][]string),
			FailureScenarios: make(map[string]*FailureRecovery),
		},
		ExecutionPlanner:   &OptimizedExecutionPlanner{},
		verbDependencies:   initializeVerbDependencies(),
		attributeProducers: make(map[string]string),
		criticalPaths:      make([]string, 0),
	}
}

// AnalyzeDependencies performs comprehensive dependency analysis
func (do *DependencyOptimizer) AnalyzeDependencies(dsl string) error {
	// Phase 1: Parse DSL and extract operations
	operations, err := do.parseDSLOperations(dsl)
	if err != nil {
		return fmt.Errorf("failed to parse DSL operations: %w", err)
	}

	// Phase 2: Build attribute dependency graph
	if err := do.buildAttributeGraph(operations); err != nil {
		return fmt.Errorf("failed to build attribute graph: %w", err)
	}

	// Phase 3: Analyze resource dependencies
	if err := do.analyzeResourceDependencies(operations); err != nil {
		return fmt.Errorf("failed to analyze resource dependencies: %w", err)
	}

	// Phase 4: Identify critical paths
	if err := do.identifyCriticalPaths(); err != nil {
		return fmt.Errorf("failed to identify critical paths: %w", err)
	}

	// Phase 5: Detect circular dependencies
	if err := do.detectCircularDependencies(); err != nil {
		return fmt.Errorf("circular dependency detected: %w", err)
	}

	return nil
}

// GenerateOptimalExecutionOrder creates an optimized execution plan
func (do *DependencyOptimizer) GenerateOptimalExecutionOrder(operations []Operation) (*OptimizedExecutionPlanner, error) {
	planner := &OptimizedExecutionPlanner{
		Phases:                make([]OptimizedPhase, 0),
		ParallelGroups:        make(map[int][]string),
		SynchronizationPoints: make([]SyncPoint, 0),
	}

	// Phase 1: Calculate dependency levels using topological sort
	levels := do.calculateDependencyLevels(operations)

	// Phase 2: Group operations by dependency level
	phaseGroups := make(map[int][]Operation)
	for _, op := range operations {
		level := levels[op.ID]
		if phaseGroups[level] == nil {
			phaseGroups[level] = make([]Operation, 0)
		}
		phaseGroups[level] = append(phaseGroups[level], op)
	}

	// Phase 3: Create optimized phases
	maxLevel := 0
	for level := range phaseGroups {
		if level > maxLevel {
			maxLevel = level
		}
	}

	for level := 0; level <= maxLevel; level++ {
		if ops, exists := phaseGroups[level]; exists {
			phase := do.createOptimizedPhase(level, ops)
			planner.Phases = append(planner.Phases, phase)

			// Identify parallel groups
			if phase.Parallelizable && len(phase.Operations) > 1 {
				parallelOps := make([]string, 0)
				for _, op := range phase.Operations {
					if op.CanParallelize {
						parallelOps = append(parallelOps, op.ID)
					}
				}
				planner.ParallelGroups[level] = parallelOps
			}

			// Add synchronization points for critical transitions
			if do.requiresSynchronization(phase) {
				syncPoint := SyncPoint{
					Name:          fmt.Sprintf("sync_after_phase_%d", level),
					AfterPhase:    level,
					WaitFor:       do.extractWaitConditions(phase),
					TimeoutMs:     30000, // 30 second timeout
					OnTimeout:     "FAIL",
					Description:   fmt.Sprintf("Synchronization point after phase %d", level),
					CriticalPoint: do.isOnCriticalPath(phase),
				}
				planner.SynchronizationPoints = append(planner.SynchronizationPoints, syncPoint)
			}
		}
	}

	// Phase 4: Calculate critical path and timing estimates
	do.calculateCriticalPath(planner)
	do.calculateTimingEstimates(planner)

	do.ExecutionPlanner = planner
	return planner, nil
}

// Helper methods for dependency analysis

func (do *DependencyOptimizer) parseDSLOperations(dsl string) ([]Operation, error) {
	operations := make([]Operation, 0)
	lines := strings.Split(dsl, "\n")

	verbPattern := regexp.MustCompile(`^\s*\(([^.]+\.[^.\s]+)`)
	attrPattern := regexp.MustCompile(`@attr\{([^}]+)\}`)

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, ";") {
			continue
		}

		matches := verbPattern.FindStringSubmatch(line)
		if len(matches) > 1 {
			verb := matches[1]
			domain := strings.Split(verb, ".")[0]

			// Extract attribute IDs from the line
			attrMatches := attrPattern.FindAllStringSubmatch(line, -1)
			attributes := make([]string, 0)
			for _, attrMatch := range attrMatches {
				if len(attrMatch) > 1 {
					attributes = append(attributes, attrMatch[1])
				}
			}

			operation := Operation{
				ID:           fmt.Sprintf("op_%d", len(operations)),
				Verb:         verb,
				Domain:       domain,
				Dependencies: make([]string, 0),
				Produces:     do.determineProducedAttributes(verb),
				Attributes:   attributes,
				Line:         i + 1,
			}

			operations = append(operations, operation)
		}
	}

	return operations, nil
}

func (do *DependencyOptimizer) buildAttributeGraph(operations []Operation) error {
	// Build nodes for each attribute
	for _, op := range operations {
		domain := op.Domain

		// Add domain to domains map
		if do.AttributeGraph.Domains[domain] == nil {
			do.AttributeGraph.Domains[domain] = make([]string, 0)
		}

		// Process attributes used by this operation
		for _, attrID := range op.Attributes {
			if do.AttributeGraph.Attributes[attrID] == nil {
				node := &AttributeNode{
					AttributeID:  attrID,
					Domain:       domain,
					Type:         "INPUT", // Default, may be updated
					ProducedBy:   make([]string, 0),
					RequiredBy:   make([]string, 0),
					Dependencies: make([]string, 0),
					ComputeCost:  1,
				}
				do.AttributeGraph.Attributes[attrID] = node
				do.AttributeGraph.Domains[domain] = append(do.AttributeGraph.Domains[domain], attrID)
			}

			// Mark as required by this operation
			do.AttributeGraph.Attributes[attrID].RequiredBy = append(
				do.AttributeGraph.Attributes[attrID].RequiredBy, op.ID)
		}

		// Process attributes produced by this operation
		for _, attrID := range op.Produces {
			if do.AttributeGraph.Attributes[attrID] == nil {
				node := &AttributeNode{
					AttributeID:  attrID,
					Domain:       domain,
					Type:         "COMPUTED",
					ProducedBy:   make([]string, 0),
					RequiredBy:   make([]string, 0),
					Dependencies: make([]string, 0),
					ComputeCost:  5, // Computed attributes have higher cost
				}
				do.AttributeGraph.Attributes[attrID] = node
				do.AttributeGraph.Domains[domain] = append(do.AttributeGraph.Domains[domain], attrID)
			}

			// Mark as produced by this operation
			do.AttributeGraph.Attributes[attrID].ProducedBy = append(
				do.AttributeGraph.Attributes[attrID].ProducedBy, op.ID)
			do.AttributeGraph.Attributes[attrID].Type = "COMPUTED"

			// Record producer
			do.attributeProducers[attrID] = op.ID
		}
	}

	// Build edges based on attribute flow
	for attrID, node := range do.AttributeGraph.Attributes {
		for _, producer := range node.ProducedBy {
			for _, consumer := range node.RequiredBy {
				if producer != consumer {
					edge := AttributeEdge{
						From:        producer,
						To:          consumer,
						Type:        "REQUIRES",
						Strength:    do.calculateDependencyStrength(attrID),
						CrossDomain: do.isCrossDomainDependency(producer, consumer, operations),
					}
					do.AttributeGraph.Edges = append(do.AttributeGraph.Edges, edge)
				}
			}
		}
	}

	return nil
}

func (do *DependencyOptimizer) analyzeResourceDependencies(operations []Operation) error {
	for _, op := range operations {
		if strings.Contains(op.Verb, "resources.create") {
			resourceType := do.extractResourceType(op.Verb)

			dependency := &ResourceDependency{
				ResourceType:    resourceType,
				DependsOn:       do.determineResourcePrerequisites(resourceType, op),
				CreationVerb:    op.Verb,
				WaitCondition:   do.determineWaitCondition(resourceType),
				FailureHandling: "ROLLBACK_PARTIAL_RESOURCES",
				EstimatedTime:   do.estimateResourceCreationTime(resourceType),
				RetryPolicy:     "EXPONENTIAL_BACKOFF",
				Priority:        do.calculateResourcePriority(resourceType),
			}

			do.ResourceTracker.Resources[resourceType] = dependency
			do.ResourceTracker.Prerequisites[resourceType] = dependency.DependsOn

			// Add to creation order based on dependencies
			do.updateCreationOrder(resourceType, dependency)
		}
	}

	return nil
}

func (do *DependencyOptimizer) identifyCriticalPaths() error {
	// Use longest path algorithm to find critical paths
	// This is a simplified implementation - production would use more sophisticated algorithms

	longestPaths := make(map[string]int)

	// Initialize with operations that have no dependencies
	for attrID, node := range do.AttributeGraph.Attributes {
		if len(node.Dependencies) == 0 {
			longestPaths[attrID] = node.ComputeCost
		}
	}

	// Calculate longest paths using dynamic programming
	changed := true
	for changed {
		changed = false
		for attrID, node := range do.AttributeGraph.Attributes {
			if len(node.Dependencies) > 0 {
				maxPath := 0
				for _, depID := range node.Dependencies {
					if pathLength, exists := longestPaths[depID]; exists {
						if pathLength > maxPath {
							maxPath = pathLength
						}
					}
				}
				newPath := maxPath + node.ComputeCost
				if longestPaths[attrID] < newPath {
					longestPaths[attrID] = newPath
					changed = true
				}
			}
		}
	}

	// Mark nodes on critical path
	maxPathLength := 0
	for _, pathLength := range longestPaths {
		if pathLength > maxPathLength {
			maxPathLength = pathLength
		}
	}

	for attrID, pathLength := range longestPaths {
		if pathLength == maxPathLength {
			do.AttributeGraph.Attributes[attrID].CriticalPath = true
			do.criticalPaths = append(do.criticalPaths, attrID)
		}
	}

	return nil
}

func (do *DependencyOptimizer) detectCircularDependencies() error {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var hasCycle func(nodeID string) bool
	hasCycle = func(nodeID string) bool {
		visited[nodeID] = true
		recStack[nodeID] = true

		if node, exists := do.AttributeGraph.Attributes[nodeID]; exists {
			for _, depID := range node.Dependencies {
				if !visited[depID] {
					if hasCycle(depID) {
						return true
					}
				} else if recStack[depID] {
					return true
				}
			}
		}

		recStack[nodeID] = false
		return false
	}

	for nodeID := range do.AttributeGraph.Attributes {
		if !visited[nodeID] {
			if hasCycle(nodeID) {
				return fmt.Errorf("circular dependency detected involving attribute: %s", nodeID)
			}
		}
	}

	return nil
}

// Helper methods

func (do *DependencyOptimizer) calculateDependencyLevels(operations []Operation) map[string]int {
	levels := make(map[string]int)

	// Build operation dependency map
	opDeps := make(map[string][]string)
	for _, op := range operations {
		opDeps[op.ID] = make([]string, 0)

		// Add dependencies based on verb dependencies
		if deps, exists := do.verbDependencies[op.Verb]; exists {
			for _, depVerb := range deps {
				// Find operations with this verb that come before
				for _, prevOp := range operations {
					if prevOp.Verb == depVerb && prevOp.ID != op.ID {
						opDeps[op.ID] = append(opDeps[op.ID], prevOp.ID)
					}
				}
			}
		}
	}

	// Topological sort to calculate levels
	visited := make(map[string]bool)
	var calculateLevel func(opID string) int
	calculateLevel = func(opID string) int {
		if visited[opID] {
			return levels[opID]
		}

		visited[opID] = true
		maxDepLevel := -1

		if deps, exists := opDeps[opID]; exists {
			for _, depID := range deps {
				depLevel := calculateLevel(depID)
				if depLevel > maxDepLevel {
					maxDepLevel = depLevel
				}
			}
		}

		levels[opID] = maxDepLevel + 1
		return levels[opID]
	}

	for _, op := range operations {
		calculateLevel(op.ID)
	}

	return levels
}

func (do *DependencyOptimizer) createOptimizedPhase(level int, operations []Operation) OptimizedPhase {
	phase := OptimizedPhase{
		PhaseID:           level,
		Operations:        make([]OptimizedOp, 0),
		Dependencies:      make([]string, 0),
		Parallelizable:    len(operations) > 1,
		EstimatedTime:     0,
		ResourcesNeeded:   make([]string, 0),
		WaitConditions:    make([]WaitCondition, 0),
		OptimizationHints: make([]OptimizationHint, 0),
	}

	for _, op := range operations {
		optimizedOp := OptimizedOp{
			ID:             op.ID,
			Verb:           op.Verb,
			Domain:         op.Domain,
			Dependencies:   op.Dependencies,
			Produces:       op.Produces,
			EstimatedTime:  do.estimateOperationTime(op.Verb),
			CanParallelize: do.canParallelize(op.Verb),
			CriticalPath:   do.isOperationOnCriticalPath(op.ID),
			RetryPolicy:    do.determineRetryPolicy(op.Verb),
			Priority:       do.calculateOperationPriority(op.Verb),
		}

		phase.Operations = append(phase.Operations, optimizedOp)
		phase.EstimatedTime += optimizedOp.EstimatedTime

		// Add wait conditions for resource-dependent operations
		if strings.Contains(op.Verb, "resources.") {
			waitCondition := WaitCondition{
				Name:        fmt.Sprintf("wait_for_%s", op.ID),
				Type:        "RESOURCE_CREATED",
				Target:      do.extractResourceType(op.Verb),
				Timeout:     10000, // 10 seconds
				OnTimeout:   "RETRY",
				Description: fmt.Sprintf("Wait for resource creation in %s", op.Verb),
			}
			phase.WaitConditions = append(phase.WaitConditions, waitCondition)
		}
	}

	// Sort operations by priority
	sort.Slice(phase.Operations, func(i, j int) bool {
		return phase.Operations[i].Priority > phase.Operations[j].Priority
	})

	// Add optimization hints
	if phase.Parallelizable {
		hint := OptimizationHint{
			Type:        "PARALLEL_SAFE",
			Description: "Operations in this phase can execute in parallel",
			Priority:    5,
		}
		phase.OptimizationHints = append(phase.OptimizationHints, hint)
	}

	return phase
}

// Operation represents a DSL operation for dependency analysis
type Operation struct {
	ID           string   `json:"id"`
	Verb         string   `json:"verb"`
	Domain       string   `json:"domain"`
	Dependencies []string `json:"dependencies"`
	Produces     []string `json:"produces"`
	Attributes   []string `json:"attributes"`
	Line         int      `json:"line"`
}

// Utility functions

func initializeVerbDependencies() map[string][]string {
	return map[string][]string{
		"kyc.verify":         {"kyc.collect", "kyc.start"},
		"kyc.complete":       {"kyc.verify"},
		"ubo.verify":         {"ubo.discover", "ubo.identify"},
		"ubo.complete":       {"ubo.verify"},
		"resources.create":   {"kyc.complete", "ubo.complete"},
		"signatory.create":   {"resources.create"},
		"account.activate":   {"signatory.create"},
		"services.provision": {"account.activate"},
	}
}

func (do *DependencyOptimizer) determineProducedAttributes(verb string) []string {
	switch {
	case strings.Contains(verb, "ubo.verify"):
		return []string{"ubo-verification-complete"}
	case strings.Contains(verb, "kyc.complete"):
		return []string{"kyc-complete"}
	case strings.Contains(verb, "resources.create"):
		return []string{"resource-created"}
	case strings.Contains(verb, "account.activate"):
		return []string{"account-active"}
	default:
		return []string{}
	}
}

func (do *DependencyOptimizer) calculateDependencyStrength(attrID string) int {
	// Critical attributes have higher strength
	if strings.Contains(attrID, "verification-complete") {
		return 10
	}
	if strings.Contains(attrID, "kyc-complete") {
		return 9
	}
	return 5 // Default strength
}

func (do *DependencyOptimizer) isCrossDomainDependency(producer, consumer string, operations []Operation) bool {
	var producerDomain, consumerDomain string

	for _, op := range operations {
		if op.ID == producer {
			producerDomain = op.Domain
		}
		if op.ID == consumer {
			consumerDomain = op.Domain
		}
	}

	return producerDomain != consumerDomain
}

func (do *DependencyOptimizer) extractResourceType(verb string) string {
	if strings.Contains(verb, "custody") {
		return "CUSTODY_ACCOUNT"
	}
	if strings.Contains(verb, "signatory") {
		return "SIGNATORY_AUTHORITY"
	}
	if strings.Contains(verb, "account") {
		return "TRADING_ACCOUNT"
	}
	return "UNKNOWN_RESOURCE"
}

func (do *DependencyOptimizer) determineResourcePrerequisites(resourceType string, op Operation) []string {
	switch resourceType {
	case "CUSTODY_ACCOUNT":
		return []string{"@attr{ubo-verification-complete}", "@attr{kyc-complete}"}
	case "SIGNATORY_AUTHORITY":
		return []string{"@attr{custody-account-created}"}
	case "TRADING_ACCOUNT":
		return []string{"@attr{signatory-authority-created}"}
	default:
		return []string{}
	}
}

func (do *DependencyOptimizer) determineWaitCondition(resourceType string) string {
	switch resourceType {
	case "CUSTODY_ACCOUNT":
		return "UBO_IDENTITY_VERIFIED"
	case "SIGNATORY_AUTHORITY":
		return "CUSTODY_ACCOUNT_ACTIVE"
	default:
		return "PREREQUISITES_MET"
	}
}

func (do *DependencyOptimizer) estimateResourceCreationTime(resourceType string) int {
	switch resourceType {
	case "CUSTODY_ACCOUNT":
		return 5000 // 5 seconds
	case "SIGNATORY_AUTHORITY":
		return 3000 // 3 seconds
	case "TRADING_ACCOUNT":
		return 2000 // 2 seconds
	default:
		return 1000 // 1 second
	}
}

func (do *DependencyOptimizer) calculateResourcePriority(resourceType string) int {
	switch resourceType {
	case "CUSTODY_ACCOUNT":
		return 100
	case "SIGNATORY_AUTHORITY":
		return 90
	case "TRADING_ACCOUNT":
		return 80
	default:
		return 50
	}
}

func (do *DependencyOptimizer) updateCreationOrder(resourceType string, dependency *ResourceDependency) {
	// Simple ordering - in production would use topological sort
	if len(dependency.DependsOn) == 0 {
		// No dependencies, add at beginning
		do.ResourceTracker.CreationOrder = append([]string{resourceType}, do.ResourceTracker.CreationOrder...)
	} else {
		// Has dependencies, add at end
		do.ResourceTracker.CreationOrder = append(do.ResourceTracker.CreationOrder, resourceType)
	}
}

func (do *DependencyOptimizer) requiresSynchronization(phase OptimizedPhase) bool {
	// Phases with resource creation or cross-domain operations require sync
	for _, op := range phase.Operations {
		if strings.Contains(op.Verb, "resources.") || op.CriticalPath {
			return true
		}
	}
	return false
}

func (do *DependencyOptimizer) extractWaitConditions(phase OptimizedPhase) []string {
	conditions := make([]string, 0)
	for _, waitCond := range phase.WaitConditions {
		conditions = append(conditions, waitCond.Name)
	}
	return conditions
}

func (do *DependencyOptimizer) isOnCriticalPath(phase OptimizedPhase) bool {
	for _, op := range phase.Operations {
		if op.CriticalPath {
			return true
		}
	}
	return false
}

func (do *DependencyOptimizer) calculateCriticalPath(planner *OptimizedExecutionPlanner) {
	// Calculate the longest path through the execution plan
	totalTime := 0
	for _, phase := range planner.Phases {
		if phase.Parallelizable {
			// For parallel phases, use max time of any operation
			maxTime := 0
			for _, op := range phase.Operations {
				if op.EstimatedTime > maxTime {
					maxTime = op.EstimatedTime
				}
			}
			totalTime += maxTime
		} else {
			// For sequential phases, sum all operation times
			totalTime += phase.EstimatedTime
		}
	}

	planner.CriticalPathLength = totalTime
}

func (do *DependencyOptimizer) calculateTimingEstimates(planner *OptimizedExecutionPlanner) {
	totalTime := 0
	for _, phase := range planner.Phases {
		totalTime += phase.EstimatedTime
	}
	planner.TotalEstimatedTime = totalTime
}

// estimateOperationTime estimates execution time for an operation
func (do *DependencyOptimizer) estimateOperationTime(verb string) int {
	switch {
	case strings.Contains(verb, "kyc."):
		return 3000 // 3 seconds
	case strings.Contains(verb, "ubo."):
		return 4000 // 4 seconds
	case strings.Contains(verb, "resources."):
		return 5000 // 5 seconds
	default:
		return 2000 // 2 seconds
	}
}

// canParallelize determines if an operation can be parallelized
func (do *DependencyOptimizer) canParallelize(verb string) bool {
	switch {
	case strings.Contains(verb, "resources.create"):
		return false // Resource creation should be sequential
	case strings.Contains(verb, "kyc.collect"):
		return true
	case strings.Contains(verb, "ubo.discover"):
		return true
	default:
		return true
	}
}

// isOperationOnCriticalPath checks if an operation is on the critical path
func (do *DependencyOptimizer) isOperationOnCriticalPath(opID string) bool {
	for _, criticalOpID := range do.criticalPaths {
		if criticalOpID == opID {
			return true
		}
	}
	return false
}

// determineRetryPolicy determines retry policy for an operation
func (do *DependencyOptimizer) determineRetryPolicy(verb string) string {
	switch {
	case strings.Contains(verb, "resources.create"):
		return "NO_RETRY" // Resource creation should not be retried
	case strings.Contains(verb, "kyc."):
		return "EXPONENTIAL_BACKOFF"
	case strings.Contains(verb, "ubo."):
		return "FIXED_DELAY"
	default:
		return "EXPONENTIAL_BACKOFF"
	}
}

// calculateOperationPriority calculates priority for an operation
func (do *DependencyOptimizer) calculateOperationPriority(verb string) int {
	switch {
	case strings.Contains(verb, "kyc."):
		return 100
	case strings.Contains(verb, "ubo."):
		return 90
	case strings.Contains(verb, "resources."):
		return 80
	case strings.Contains(verb, "services."):
		return 70
	default:
		return 50
	}
}
