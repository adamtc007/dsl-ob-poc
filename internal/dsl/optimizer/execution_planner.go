// Package optimizer provides execution planning for optimized DSL compilation
package optimizer

import (
	"fmt"
	"strings"
	"time"
)

// ExecutionPlanner creates optimized execution plans for DSL operations
type ExecutionPlanner struct {
	dependencyOptimizer *DependencyOptimizer
	resourceTracker     *ResourceDependencyTracker
	timingEstimator     *TimingEstimator
	parallelismAnalyzer *ParallelismAnalyzer
}

// ExecutionPlan represents a complete execution plan with optimization
type ExecutionPlan struct {
	PlanID              string                 `json:"plan_id"`
	Phases              []ExecutionPhase       `json:"phases"`
	TotalOperations     int                    `json:"total_operations"`
	ParallelGroups      int                    `json:"parallel_groups"`
	CriticalPath        []string               `json:"critical_path"`
	ResourceOrder       []string               `json:"resource_order"`
	SyncPoints          []SynchronizationPoint `json:"sync_points"`
	EstimatedDuration   int                    `json:"estimated_duration_ms"`
	OptimizationMetrics *OptimizationMetrics   `json:"optimization_metrics"`
	FailureRecovery     *FailureRecoveryPlan   `json:"failure_recovery"`
}

// ExecutionPhase represents a group of operations that execute together
type ExecutionPhase struct {
	PhaseID         int                `json:"phase_id"`
	Name            string             `json:"name"`
	Operations      []PlannedOperation `json:"operations"`
	Dependencies    []string           `json:"dependencies"`
	CanParallelize  bool               `json:"can_parallelize"`
	EstimatedTime   int                `json:"estimated_time_ms"`
	ResourcesNeeded []string           `json:"resources_needed"`
	WaitConditions  []WaitCondition    `json:"wait_conditions"`
	ExecutionHints  []ExecutionHint    `json:"execution_hints"`
	FailureStrategy string             `json:"failure_strategy"`
	MaxRetries      int                `json:"max_retries"`
	Timeout         int                `json:"timeout_ms"`
}

// PlannedOperation represents an operation with execution metadata
type PlannedOperation struct {
	ID               string            `json:"id"`
	Verb             string            `json:"verb"`
	DSLFragment      string            `json:"dsl_fragment"`
	Dependencies     []string          `json:"dependencies"`
	Produces         []string          `json:"produces"`
	Domain           string            `json:"domain"`
	Priority         int               `json:"priority"`
	CanRetry         bool              `json:"can_retry"`
	EstimatedTime    int               `json:"estimated_time_ms"`
	ResourceRequests []ResourceRequest `json:"resource_requests"`
	HealthChecks     []HealthCheck     `json:"health_checks"`
	OnCriticalPath   bool              `json:"on_critical_path"`
	ParallelSafe     bool              `json:"parallel_safe"`
	ExecutionContext *ExecutionContext `json:"execution_context"`
}

// ResourceRequest represents a resource needed by an operation
type ResourceRequest struct {
	ResourceType string            `json:"resource_type"`
	ResourceID   string            `json:"resource_id"`
	AccessType   string            `json:"access_type"` // "READ", "WRITE", "CREATE", "DELETE"
	Required     bool              `json:"required"`
	Timeout      int               `json:"timeout_ms"`
	RetryPolicy  string            `json:"retry_policy"`
	Metadata     map[string]string `json:"metadata"`
}

// HealthCheck represents a health check for an operation
type HealthCheck struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // "ATTRIBUTE_EXISTS", "RESOURCE_AVAILABLE", "SERVICE_HEALTHY"
	Target      string `json:"target"`
	Interval    int    `json:"interval_ms"`
	Timeout     int    `json:"timeout_ms"`
	MaxFailures int    `json:"max_failures"`
	OnFailure   string `json:"on_failure"` // "RETRY", "FAIL", "CONTINUE"
}

// ExecutionContext holds runtime context for an operation
type ExecutionContext struct {
	SessionID    string            `json:"session_id"`
	UserID       string            `json:"user_id,omitempty"`
	TraceID      string            `json:"trace_id"`
	ParentSpanID string            `json:"parent_span_id,omitempty"`
	Attributes   map[string]string `json:"attributes"`
	Environment  string            `json:"environment"`
	Region       string            `json:"region,omitempty"`
}

// ExecutionHint provides guidance for operation execution
type ExecutionHint struct {
	Type        string `json:"type"` // "PARALLEL_SAFE", "CPU_INTENSIVE", "IO_BOUND", "NETWORK_DEPENDENT"
	Description string `json:"description"`
	Priority    int    `json:"priority"`
	Value       string `json:"value,omitempty"`
}

// SynchronizationPoint represents a point where execution must coordinate
type SynchronizationPoint struct {
	Name          string          `json:"name"`
	AfterPhase    int             `json:"after_phase"`
	WaitFor       []string        `json:"wait_for"`
	Timeout       int             `json:"timeout_ms"`
	OnTimeout     string          `json:"on_timeout"`
	Description   string          `json:"description"`
	CriticalPoint bool            `json:"critical_point"`
	Conditions    []SyncCondition `json:"conditions"`
}

// SyncCondition represents a condition that must be met at a sync point
type SyncCondition struct {
	Type        string `json:"type"` // "ATTRIBUTE_READY", "RESOURCE_CREATED", "EXTERNAL_APPROVAL"
	Target      string `json:"target"`
	Expected    string `json:"expected_value,omitempty"`
	Operator    string `json:"operator,omitempty"` // "EQUALS", "EXISTS", "GREATER_THAN"
	Description string `json:"description"`
}

// OptimizationMetrics tracks optimization effectiveness
type OptimizationMetrics struct {
	OriginalDuration     int      `json:"original_duration_ms"`
	OptimizedDuration    int      `json:"optimized_duration_ms"`
	TimeReduction        int      `json:"time_reduction_ms"`
	PercentImprovement   float64  `json:"percent_improvement"`
	ParallelOperations   int      `json:"parallel_operations"`
	SequentialOperations int      `json:"sequential_operations"`
	CriticalPathLength   int      `json:"critical_path_length_ms"`
	OptimizationsApplied []string `json:"optimizations_applied"`
}

// FailureRecoveryPlan defines how to handle execution failures
type FailureRecoveryPlan struct {
	Strategy            string         `json:"strategy"` // "ROLLBACK", "CONTINUE", "RETRY", "MANUAL"
	MaxRetries          int            `json:"max_retries"`
	RetryDelay          int            `json:"retry_delay_ms"`
	RollbackSteps       []RollbackStep `json:"rollback_steps"`
	NotificationTargets []string       `json:"notification_targets"`
	FailureThresholds   map[string]int `json:"failure_thresholds"`
}

// RollbackStep represents a step in rollback execution
type RollbackStep struct {
	StepID      string `json:"step_id"`
	Description string `json:"description"`
	Operation   string `json:"operation"`
	Target      string `json:"target"`
	Condition   string `json:"condition,omitempty"`
}

// TimingEstimator provides execution time estimates
type TimingEstimator struct {
	verbTimings     map[string]int // verb -> estimated time in ms
	resourceTimings map[string]int // resource type -> creation time
	domainOverhead  map[string]int // domain -> additional overhead
}

// ParallelismAnalyzer identifies parallel execution opportunities
type ParallelismAnalyzer struct {
	parallelSafeVerbs map[string]bool
	resourceLocks     map[string]bool // resources that require exclusive access
	domainIsolation   map[string]bool // domains that can't run in parallel
}

// NewExecutionPlanner creates a new execution planner
func NewExecutionPlanner(dependencyOptimizer *DependencyOptimizer) *ExecutionPlanner {
	return &ExecutionPlanner{
		dependencyOptimizer: dependencyOptimizer,
		resourceTracker:     dependencyOptimizer.ResourceTracker,
		timingEstimator:     NewTimingEstimator(),
		parallelismAnalyzer: NewParallelismAnalyzer(),
	}
}

// NewTimingEstimator creates a timing estimator with default values
func NewTimingEstimator() *TimingEstimator {
	return &TimingEstimator{
		verbTimings: map[string]int{
			"kyc.start":          2000, // 2 seconds
			"kyc.collect":        3000, // 3 seconds
			"kyc.verify":         5000, // 5 seconds
			"kyc.complete":       1000, // 1 second
			"ubo.discover":       4000, // 4 seconds
			"ubo.identify":       3000, // 3 seconds
			"ubo.verify":         6000, // 6 seconds
			"ubo.complete":       1000, // 1 second
			"resources.create":   5000, // 5 seconds
			"signatory.create":   3000, // 3 seconds
			"account.activate":   2000, // 2 seconds
			"services.provision": 4000, // 4 seconds
		},
		resourceTimings: map[string]int{
			"CUSTODY_ACCOUNT":     5000, // 5 seconds
			"SIGNATORY_AUTHORITY": 3000, // 3 seconds
			"TRADING_ACCOUNT":     2000, // 2 seconds
			"SERVICE_ENDPOINT":    4000, // 4 seconds
		},
		domainOverhead: map[string]int{
			"kyc":       500,  // 500ms overhead
			"ubo":       800,  // 800ms overhead
			"resources": 1000, // 1 second overhead
			"services":  600,  // 600ms overhead
		},
	}
}

// NewParallelismAnalyzer creates a parallelism analyzer with default rules
func NewParallelismAnalyzer() *ParallelismAnalyzer {
	return &ParallelismAnalyzer{
		parallelSafeVerbs: map[string]bool{
			"kyc.collect":    true,
			"ubo.discover":   true,
			"data.fetch":     true,
			"validation.run": true,
		},
		resourceLocks: map[string]bool{
			"CUSTODY_ACCOUNT":     true, // Exclusive access required
			"SIGNATORY_AUTHORITY": true,
			"TRADING_ACCOUNT":     false, // Can be accessed in parallel
		},
		domainIsolation: map[string]bool{
			"kyc":       false, // KYC can run in parallel with other domains
			"ubo":       false, // UBO can run in parallel
			"resources": true,  // Resources require sequential execution
		},
	}
}

// CreateExecutionPlan generates an optimized execution plan
func (ep *ExecutionPlanner) CreateExecutionPlan(dsl string, sessionID string) (*ExecutionPlan, error) {
	// Phase 1: Parse DSL and analyze dependencies
	if err := ep.dependencyOptimizer.AnalyzeDependencies(dsl); err != nil {
		return nil, fmt.Errorf("dependency analysis failed: %w", err)
	}

	// Phase 2: Extract operations from DSL
	operations, err := ep.extractOperations(dsl)
	if err != nil {
		return nil, fmt.Errorf("operation extraction failed: %w", err)
	}

	// Phase 3: Create execution phases based on dependencies
	phases, err := ep.createExecutionPhases(operations)
	if err != nil {
		return nil, fmt.Errorf("phase creation failed: %w", err)
	}

	// Phase 4: Optimize for parallel execution
	if err := ep.optimizeParallelExecution(phases); err != nil {
		return nil, fmt.Errorf("parallel optimization failed: %w", err)
	}

	// Phase 5: Add synchronization points
	syncPoints := ep.createSynchronizationPoints(phases)

	// Phase 6: Calculate critical path
	criticalPath := ep.calculateCriticalPath(phases)

	// Phase 7: Generate optimization metrics
	metrics := ep.calculateOptimizationMetrics(phases, operations)

	// Phase 8: Create failure recovery plan
	failureRecovery := ep.createFailureRecoveryPlan(phases)

	plan := &ExecutionPlan{
		PlanID:              fmt.Sprintf("plan_%s_%d", sessionID, time.Now().Unix()),
		Phases:              phases,
		TotalOperations:     len(operations),
		ParallelGroups:      ep.countParallelGroups(phases),
		CriticalPath:        criticalPath,
		ResourceOrder:       ep.resourceTracker.CreationOrder,
		SyncPoints:          syncPoints,
		EstimatedDuration:   ep.calculateTotalDuration(phases),
		OptimizationMetrics: metrics,
		FailureRecovery:     failureRecovery,
	}

	return plan, nil
}

// extractOperations parses DSL and extracts operations
func (ep *ExecutionPlanner) extractOperations(dsl string) ([]PlannedOperation, error) {
	operations := make([]PlannedOperation, 0)
	lines := strings.Split(dsl, "\n")

	verbPattern := `^\s*\(([^.]+\.[^.\s]+)`

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, ";") {
			continue
		}

		// Simple regex matching - production would use proper parsing
		if matched, verb := ep.matchVerb(line, verbPattern); matched {
			domain := strings.Split(verb, ".")[0]

			operation := PlannedOperation{
				ID:               fmt.Sprintf("op_%d", len(operations)),
				Verb:             verb,
				DSLFragment:      line,
				Dependencies:     make([]string, 0),
				Produces:         ep.determineProducedAttributes(verb),
				Domain:           domain,
				Priority:         ep.calculateOperationPriority(verb),
				CanRetry:         ep.canRetry(verb),
				EstimatedTime:    ep.timingEstimator.EstimateOperationTime(verb, domain),
				ResourceRequests: ep.extractResourceRequests(verb),
				HealthChecks:     ep.createHealthChecks(verb),
				OnCriticalPath:   false, // Will be updated later
				ParallelSafe:     ep.parallelismAnalyzer.IsParallelSafe(verb),
				ExecutionContext: &ExecutionContext{
					SessionID:   fmt.Sprintf("session_%d", time.Now().Unix()),
					TraceID:     fmt.Sprintf("trace_%d_%d", time.Now().Unix(), i),
					Attributes:  make(map[string]string),
					Environment: "development",
				},
			}

			operations = append(operations, operation)
		}
	}

	return operations, nil
}

// createExecutionPhases groups operations into execution phases
func (ep *ExecutionPlanner) createExecutionPhases(operations []PlannedOperation) ([]ExecutionPhase, error) {
	phases := make([]ExecutionPhase, 0)

	// Group operations by dependency level
	dependencyLevels := ep.calculateOperationLevels(operations)
	levelGroups := make(map[int][]PlannedOperation)

	maxLevel := 0
	for _, op := range operations {
		level := dependencyLevels[op.ID]
		if level > maxLevel {
			maxLevel = level
		}

		if levelGroups[level] == nil {
			levelGroups[level] = make([]PlannedOperation, 0)
		}
		levelGroups[level] = append(levelGroups[level], op)
	}

	// Create phases for each level
	for level := 0; level <= maxLevel; level++ {
		if ops, exists := levelGroups[level]; exists {
			phase := ExecutionPhase{
				PhaseID:         level,
				Name:            fmt.Sprintf("Phase_%d", level),
				Operations:      ops,
				Dependencies:    ep.extractPhaseDependencies(ops),
				CanParallelize:  ep.canParallelizePhase(ops),
				EstimatedTime:   ep.calculatePhaseTime(ops),
				ResourcesNeeded: ep.extractResourcesNeeded(ops),
				WaitConditions:  ep.createWaitConditions(ops),
				ExecutionHints:  ep.generateExecutionHints(ops),
				FailureStrategy: ep.determineFailureStrategy(ops),
				MaxRetries:      3,
				Timeout:         30000, // 30 seconds default timeout
			}

			phases = append(phases, phase)
		}
	}

	return phases, nil
}

// optimizeParallelExecution identifies and optimizes parallel execution opportunities
func (ep *ExecutionPlanner) optimizeParallelExecution(phases []ExecutionPhase) error {
	for i := range phases {
		phase := &phases[i]

		if !phase.CanParallelize {
			continue
		}

		// Group operations that can run in parallel
		parallelGroups := ep.identifyParallelGroups(phase.Operations)

		// Update operations with parallel execution metadata
		for groupID, group := range parallelGroups {
			for j, opID := range group {
				// Find operation and update it
				for k := range phase.Operations {
					if phase.Operations[k].ID == opID {
						phase.Operations[k].ExecutionContext.Attributes["parallel_group"] = fmt.Sprintf("group_%d", groupID)
						phase.Operations[k].ExecutionContext.Attributes["group_position"] = fmt.Sprintf("%d", j)
						break
					}
				}
			}
		}

		// Add parallel execution hints
		if len(parallelGroups) > 1 {
			hint := ExecutionHint{
				Type:        "PARALLEL_EXECUTION",
				Description: fmt.Sprintf("Phase has %d parallel groups", len(parallelGroups)),
				Priority:    8,
				Value:       fmt.Sprintf("%d", len(parallelGroups)),
			}
			phase.ExecutionHints = append(phase.ExecutionHints, hint)
		}
	}

	return nil
}

// createSynchronizationPoints creates sync points between phases
func (ep *ExecutionPlanner) createSynchronizationPoints(phases []ExecutionPhase) []SynchronizationPoint {
	syncPoints := make([]SynchronizationPoint, 0)

	for i, phase := range phases {
		if ep.requiresSynchronization(phase) {
			syncPoint := SynchronizationPoint{
				Name:          fmt.Sprintf("sync_after_phase_%d", i),
				AfterPhase:    i,
				WaitFor:       ep.extractWaitTargets(phase),
				Timeout:       60000, // 60 second timeout
				OnTimeout:     "FAIL",
				Description:   fmt.Sprintf("Synchronization after %s", phase.Name),
				CriticalPoint: ep.isOnCriticalPath(phase),
				Conditions:    ep.createSyncConditions(phase),
			}
			syncPoints = append(syncPoints, syncPoint)
		}
	}

	return syncPoints
}

// calculateCriticalPath identifies the critical path through execution
func (ep *ExecutionPlanner) calculateCriticalPath(phases []ExecutionPhase) []string {
	criticalPath := make([]string, 0)

	// Find the longest path through phases
	maxTime := 0
	criticalPhaseIndex := -1

	for i, phase := range phases {
		if phase.EstimatedTime > maxTime {
			maxTime = phase.EstimatedTime
			criticalPhaseIndex = i
		}
	}

	// Mark operations on critical path
	if criticalPhaseIndex >= 0 {
		phase := phases[criticalPhaseIndex]
		for i := range phase.Operations {
			phases[criticalPhaseIndex].Operations[i].OnCriticalPath = true
			criticalPath = append(criticalPath, phase.Operations[i].ID)
		}
	}

	return criticalPath
}

// Helper methods

func (ep *ExecutionPlanner) matchVerb(line, pattern string) (bool, string) {
	// Simple pattern matching - would use proper regex in production
	if strings.Contains(line, "(") && strings.Contains(line, ".") {
		parts := strings.Split(line, "(")
		if len(parts) > 1 {
			verbPart := strings.Fields(parts[1])
			if len(verbPart) > 0 && strings.Contains(verbPart[0], ".") {
				return true, verbPart[0]
			}
		}
	}
	return false, ""
}

func (ep *ExecutionPlanner) determineProducedAttributes(verb string) []string {
	switch {
	case strings.Contains(verb, "ubo.verify"):
		return []string{"@attr{ubo-verification-complete}"}
	case strings.Contains(verb, "kyc.complete"):
		return []string{"@attr{kyc-complete}"}
	case strings.Contains(verb, "resources.create"):
		return []string{"@attr{resource-created}"}
	case strings.Contains(verb, "account.activate"):
		return []string{"@attr{account-active}"}
	default:
		return []string{}
	}
}

func (ep *ExecutionPlanner) calculateOperationPriority(verb string) int {
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

func (ep *ExecutionPlanner) canRetry(verb string) bool {
	// Resource creation operations generally shouldn't be retried
	return !strings.Contains(verb, "resources.create")
}

func (te *TimingEstimator) EstimateOperationTime(verb, domain string) int {
	baseTime := 1000 // Default 1 second

	if time, exists := te.verbTimings[verb]; exists {
		baseTime = time
	}

	if overhead, exists := te.domainOverhead[domain]; exists {
		baseTime += overhead
	}

	return baseTime
}

func (ep *ExecutionPlanner) extractResourceRequests(verb string) []ResourceRequest {
	requests := make([]ResourceRequest, 0)

	if strings.Contains(verb, "resources.create") {
		resourceType := "CUSTODY_ACCOUNT" // Simplified - would extract from verb
		request := ResourceRequest{
			ResourceType: resourceType,
			ResourceID:   fmt.Sprintf("%s_%d", resourceType, time.Now().Unix()),
			AccessType:   "CREATE",
			Required:     true,
			Timeout:      10000,
			RetryPolicy:  "EXPONENTIAL_BACKOFF",
			Metadata:     map[string]string{"created_by": "execution_planner"},
		}
		requests = append(requests, request)
	}

	return requests
}

func (ep *ExecutionPlanner) createHealthChecks(verb string) []HealthCheck {
	checks := make([]HealthCheck, 0)

	if strings.Contains(verb, "kyc.") || strings.Contains(verb, "ubo.") {
		check := HealthCheck{
			Name:        fmt.Sprintf("health_check_%s", verb),
			Type:        "ATTRIBUTE_EXISTS",
			Target:      "@attr{verification-status}",
			Interval:    5000,  // 5 seconds
			Timeout:     30000, // 30 seconds
			MaxFailures: 3,
			OnFailure:   "RETRY",
		}
		checks = append(checks, check)
	}

	return checks
}

func (pa *ParallelismAnalyzer) IsParallelSafe(verb string) bool {
	return pa.parallelSafeVerbs[verb]
}

func (ep *ExecutionPlanner) calculateOperationLevels(operations []PlannedOperation) map[string]int {
	levels := make(map[string]int)

	// Simple dependency level calculation
	// In production, this would use proper topological sorting
	for i, op := range operations {
		if strings.Contains(op.Verb, "kyc.start") || strings.Contains(op.Verb, "ubo.discover") {
			levels[op.ID] = 0 // Base level
		} else if strings.Contains(op.Verb, "kyc.verify") || strings.Contains(op.Verb, "ubo.verify") {
			levels[op.ID] = 1 // Depends on collection/discovery
		} else if strings.Contains(op.Verb, "resources.") {
			levels[op.ID] = 2 // Depends on verification
		} else {
			levels[op.ID] = i / 3 // Simple level assignment
		}
	}

	return levels
}

func (ep *ExecutionPlanner) extractPhaseDependencies(operations []PlannedOperation) []string {
	deps := make([]string, 0)
	for _, op := range operations {
		deps = append(deps, op.Dependencies...)
	}
	return deps
}

func (ep *ExecutionPlanner) canParallelizePhase(operations []PlannedOperation) bool {
	if len(operations) < 2 {
		return false
	}

	// Check if any operation requires exclusive resources
	for _, op := range operations {
		for _, req := range op.ResourceRequests {
			if ep.parallelismAnalyzer.resourceLocks[req.ResourceType] {
				return false
			}
		}
	}

	return true
}

func (ep *ExecutionPlanner) calculatePhaseTime(operations []PlannedOperation) int {
	maxTime := 0
	totalTime := 0

	for _, op := range operations {
		totalTime += op.EstimatedTime
		if op.EstimatedTime > maxTime {
			maxTime = op.EstimatedTime
		}
	}

	// If operations can parallelize, use max time; otherwise use total
	if len(operations) > 1 && ep.canParallelizePhase(operations) {
		return maxTime
	}
	return totalTime
}

func (ep *ExecutionPlanner) extractResourcesNeeded(operations []PlannedOperation) []string {
	resources := make([]string, 0)
	seen := make(map[string]bool)

	for _, op := range operations {
		for _, req := range op.ResourceRequests {
			if !seen[req.ResourceType] {
				resources = append(resources, req.ResourceType)
				seen[req.ResourceType] = true
			}
		}
	}

	return resources
}

func (ep *ExecutionPlanner) createWaitConditions(operations []PlannedOperation) []WaitCondition {
	conditions := make([]WaitCondition, 0)

	for _, op := range operations {
		for _, check := range op.HealthChecks {
			condition := WaitCondition{
				Name:        fmt.Sprintf("wait_%s", check.Name),
				Type:        check.Type,
				Target:      check.Target,
				Timeout:     check.Timeout,
				OnTimeout:   "RETRY",
				Description: fmt.Sprintf("Wait for %s to be ready", check.Target),
			}
			conditions = append(conditions, condition)
		}
	}

	return conditions
}

func (ep *ExecutionPlanner) generateExecutionHints(operations []PlannedOperation) []ExecutionHint {
	hints := make([]ExecutionHint, 0)

	// Analyze operations to generate hints
	hasNetworkOps := false
	hasCPUOps := false
	hasIOOps := false

	for _, op := range operations {
		if strings.Contains(op.Verb, "fetch") || strings.Contains(op.Verb, "verify") {
			hasNetworkOps = true
		}
		if strings.Contains(op.Verb, "calculate") || strings.Contains(op.Verb, "analyze") {
			hasCPUOps = true
		}
		if strings.Contains(op.Verb, "store") || strings.Contains(op.Verb, "save") {
			hasIOOps = true
		}
	}

	if hasNetworkOps {
		hints = append(hints, ExecutionHint{
			Type:        "NETWORK_DEPENDENT",
			Description: "Phase contains network-dependent operations",
			Priority:    7,
		})
	}

	if hasCPUOps {
		hints = append(hints, ExecutionHint{
			Type:        "CPU_INTENSIVE",
			Description: "Phase contains CPU-intensive operations",
			Priority:    6,
		})
	}

	if hasIOOps {
		hints = append(hints, ExecutionHint{
			Type:        "IO_BOUND",
			Description: "Phase contains I/O-bound operations",
			Priority:    5,
		})
	}

	return hints
}

func (ep *ExecutionPlanner) determineFailureStrategy(operations []PlannedOperation) string {
	// If any operation creates resources, use rollback strategy
	for _, op := range operations {
		if strings.Contains(op.Verb, "resources.create") {
			return "ROLLBACK"
		}
	}

	// For verification operations, continue on failure
	return "CONTINUE"
}

func (ep *ExecutionPlanner) identifyParallelGroups(operations []PlannedOperation) map[int][]string {
	groups := make(map[int][]string)
	groupID := 0

	// Simple grouping by domain - production would be more sophisticated
	domainGroups := make(map[string][]string)
	for _, op := range operations {
		if op.ParallelSafe {
			if domainGroups[op.Domain] == nil {
				domainGroups[op.Domain] = make([]string, 0)
			}
			domainGroups[op.Domain] = append(domainGroups[op.Domain], op.ID)
		}
	}

	for _, opIDs := range domainGroups {
		if len(opIDs) > 1 {
			groups[groupID] = opIDs
			groupID++
		}
	}

	return groups
}

func (ep *ExecutionPlanner) requiresSynchronization(phase ExecutionPhase) bool {
	// Phases with resource creation or critical operations require sync
	for _, op := range phase.Operations {
		if strings.Contains(op.Verb, "resources.") || op.OnCriticalPath {
			return true
		}
	}
	return false
}

func (ep *ExecutionPlanner) extractWaitTargets(phase ExecutionPhase) []string {
	targets := make([]string, 0)
	for _, condition := range phase.WaitConditions {
		targets = append(targets, condition.Target)
	}
	return targets
}

func (ep *ExecutionPlanner) isOnCriticalPath(phase ExecutionPhase) bool {
	for _, op := range phase.Operations {
		if op.OnCriticalPath {
			return true
		}
	}
	return false
}

// calculateOptimizationMetrics calculates optimization effectiveness metrics
func (ep *ExecutionPlanner) calculateOptimizationMetrics(phases []ExecutionPhase, operations []PlannedOperation) *OptimizationMetrics {
	totalTime := 0
	parallelOps := 0
	sequentialOps := 0

	for _, phase := range phases {
		totalTime += phase.EstimatedTime
		if phase.CanParallelize {
			parallelOps += len(phase.Operations)
		} else {
			sequentialOps += len(phase.Operations)
		}
	}

	// Estimate original duration (all sequential)
	originalDuration := len(operations) * 3000 // 3 seconds per operation

	percentImprovement := 0.0
	if originalDuration > 0 {
		percentImprovement = float64(originalDuration-totalTime) / float64(originalDuration) * 100
	}

	return &OptimizationMetrics{
		OriginalDuration:     originalDuration,
		OptimizedDuration:    totalTime,
		TimeReduction:        originalDuration - totalTime,
		PercentImprovement:   percentImprovement,
		ParallelOperations:   parallelOps,
		SequentialOperations: sequentialOps,
		CriticalPathLength:   totalTime,
		OptimizationsApplied: []string{"DEPENDENCY_ORDER", "PARALLEL_EXECUTION"},
	}
}

// createFailureRecoveryPlan creates a failure recovery plan
func (ep *ExecutionPlanner) createFailureRecoveryPlan(phases []ExecutionPhase) *FailureRecoveryPlan {
	rollbackSteps := make([]RollbackStep, 0)

	// Create rollback steps for resource creation phases
	for _, phase := range phases {
		for _, op := range phase.Operations {
			if strings.Contains(op.Verb, "resources.create") {
				rollbackStep := RollbackStep{
					StepID:      fmt.Sprintf("rollback_%s", op.ID),
					Description: fmt.Sprintf("Rollback %s", op.Verb),
					Operation:   "DELETE",
					Target:      op.ID,
					Condition:   "ON_FAILURE",
				}
				rollbackSteps = append(rollbackSteps, rollbackStep)
			}
		}
	}

	return &FailureRecoveryPlan{
		Strategy:            "ROLLBACK",
		MaxRetries:          3,
		RetryDelay:          5000, // 5 seconds
		RollbackSteps:       rollbackSteps,
		NotificationTargets: []string{"admin@company.com"},
		FailureThresholds:   map[string]int{"error_rate": 10, "timeout_count": 5},
	}
}

// countParallelGroups counts the number of parallel groups in phases
func (ep *ExecutionPlanner) countParallelGroups(phases []ExecutionPhase) int {
	count := 0
	for _, phase := range phases {
		if phase.CanParallelize && len(phase.Operations) > 1 {
			count++
		}
	}
	return count
}

// calculateTotalDuration calculates total estimated duration
func (ep *ExecutionPlanner) calculateTotalDuration(phases []ExecutionPhase) int {
	totalTime := 0
	for _, phase := range phases {
		if phase.CanParallelize {
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
	return totalTime
}

// createSyncConditions creates synchronization conditions for a phase
func (ep *ExecutionPlanner) createSyncConditions(phase ExecutionPhase) []SyncCondition {
	conditions := make([]SyncCondition, 0)

	for _, op := range phase.Operations {
		if strings.Contains(op.Verb, "resources.") {
			condition := SyncCondition{
				Type:        "RESOURCE_CREATED",
				Target:      op.ID,
				Expected:    "ACTIVE",
				Operator:    "EQUALS",
				Description: fmt.Sprintf("Wait for %s to be active", op.Verb),
			}
			conditions = append(conditions, condition)
		}
	}

	return conditions
}
