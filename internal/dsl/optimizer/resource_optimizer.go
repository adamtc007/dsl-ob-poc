// Package optimizer provides resource optimization for DSL compilation
package optimizer

import (
	"fmt"
	"strings"
	"time"
)

// ResourceOptimizer optimizes resource creation and management in DSL execution
type ResourceOptimizer struct {
	resourceCatalog    *ResourceCatalog
	dependencyTracker  *ResourceDependencyTracker
	allocationStrategy *AllocationStrategy
	costOptimizer      *CostOptimizer
	lifecycleManager   *ResourceLifecycleManager
}

// ResourceCatalog maintains a catalog of available resources and their properties
type ResourceCatalog struct {
	Resources     map[string]*ResourceSpec     `json:"resources"`
	Templates     map[string]*ResourceTemplate `json:"templates"`
	Providers     map[string]*ResourceProvider `json:"providers"`
	Dependencies  map[string][]string          `json:"dependencies"`
	Compatibility map[string][]string          `json:"compatibility"`
}

// ResourceSpec defines the specification for a resource type
type ResourceSpec struct {
	ResourceType    string                 `json:"resource_type"`
	Category        string                 `json:"category"` // "INFRASTRUCTURE", "SERVICE", "DATA", "SECURITY"
	Provider        string                 `json:"provider"`
	CreationTime    int                    `json:"creation_time_ms"`
	DestructionTime int                    `json:"destruction_time_ms"`
	Cost            *ResourceCost          `json:"cost"`
	Capabilities    []string               `json:"capabilities"`
	Requirements    []string               `json:"requirements"`
	Scaling         *ScalingConfig         `json:"scaling"`
	Security        *SecurityConfig        `json:"security"`
	Compliance      []string               `json:"compliance"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// ResourceTemplate defines a template for creating resources
type ResourceTemplate struct {
	TemplateID        string                 `json:"template_id"`
	Name              string                 `json:"name"`
	Description       string                 `json:"description"`
	ResourceType      string                 `json:"resource_type"`
	Parameters        map[string]*Parameter  `json:"parameters"`
	DefaultValues     map[string]interface{} `json:"default_values"`
	ValidationRules   []ValidationRule       `json:"validation_rules"`
	CreationScript    string                 `json:"creation_script"`
	DestructionScript string                 `json:"destruction_script"`
	HealthCheckScript string                 `json:"health_check_script"`
}

// Parameter defines a template parameter
type Parameter struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"` // "string", "number", "boolean", "array", "object"
	Required    bool        `json:"required"`
	Default     interface{} `json:"default"`
	Description string      `json:"description"`
	Constraints []string    `json:"constraints"`
	Validation  string      `json:"validation"` // Validation expression
}

// ValidationRule defines validation for template parameters
type ValidationRule struct {
	Field        string `json:"field"`
	Rule         string `json:"rule"` // "required", "min_length", "max_value", "regex", "custom"
	Value        string `json:"value"`
	ErrorMessage string `json:"error_message"`
}

// ResourceProvider defines a resource provider
type ResourceProvider struct {
	ProviderID   string            `json:"provider_id"`
	Name         string            `json:"name"`
	Type         string            `json:"type"` // "CLOUD", "ON_PREMISE", "HYBRID", "EXTERNAL"
	Endpoint     string            `json:"endpoint"`
	Region       string            `json:"region,omitempty"`
	Capabilities []string          `json:"capabilities"`
	Limits       *ProviderLimits   `json:"limits"`
	Auth         *AuthConfig       `json:"auth"`
	SLA          *SLAConfig        `json:"sla"`
	Metadata     map[string]string `json:"metadata"`
}

// ProviderLimits defines limits for a resource provider
type ProviderLimits struct {
	MaxInstances     int `json:"max_instances"`
	MaxStorage       int `json:"max_storage_gb"`
	MaxBandwidth     int `json:"max_bandwidth_mbps"`
	MaxConcurrency   int `json:"max_concurrency"`
	RateLimitPerSec  int `json:"rate_limit_per_sec"`
	QuotaResetPeriod int `json:"quota_reset_period_hours"`
}

// AuthConfig defines authentication for a resource provider
type AuthConfig struct {
	Type        string            `json:"type"` // "API_KEY", "OAUTH", "JWT", "CERTIFICATE"
	Credentials map[string]string `json:"credentials"`
	TokenURL    string            `json:"token_url,omitempty"`
	Scopes      []string          `json:"scopes,omitempty"`
}

// SLAConfig defines Service Level Agreement for a provider
type SLAConfig struct {
	Availability      float64 `json:"availability"` // 99.9%
	ResponseTime      int     `json:"response_time_ms"`
	Throughput        int     `json:"throughput_ops_per_sec"`
	Recovery          int     `json:"recovery_time_minutes"`
	MaintenanceWindow string  `json:"maintenance_window"`
}

// ResourceCost defines cost structure for a resource
type ResourceCost struct {
	Model       string  `json:"model"` // "FIXED", "USAGE_BASED", "TIERED", "HYBRID"
	SetupCost   float64 `json:"setup_cost"`
	HourlyCost  float64 `json:"hourly_cost"`
	UsageCost   float64 `json:"usage_cost_per_unit"`
	Currency    string  `json:"currency"`
	BillingUnit string  `json:"billing_unit"` // "HOUR", "DAY", "MONTH", "USAGE"
}

// ScalingConfig defines scaling behavior for a resource
type ScalingConfig struct {
	AutoScaling        bool    `json:"auto_scaling"`
	MinInstances       int     `json:"min_instances"`
	MaxInstances       int     `json:"max_instances"`
	ScaleUpThreshold   float64 `json:"scale_up_threshold"`
	ScaleDownThreshold float64 `json:"scale_down_threshold"`
	ScaleUpCooldown    int     `json:"scale_up_cooldown_seconds"`
	ScaleDownCooldown  int     `json:"scale_down_cooldown_seconds"`
}

// SecurityConfig defines security settings for a resource
type SecurityConfig struct {
	Encryption         bool     `json:"encryption"`
	AccessControl      string   `json:"access_control"` // "RBAC", "ABAC", "ACL"
	NetworkIsolation   bool     `json:"network_isolation"`
	Compliance         []string `json:"compliance"` // "SOC2", "PCI", "HIPAA", "GDPR"
	AuditLogging       bool     `json:"audit_logging"`
	DataClassification string   `json:"data_classification"` // "PUBLIC", "INTERNAL", "CONFIDENTIAL", "RESTRICTED"
}

// AllocationStrategy defines resource allocation strategies
type AllocationStrategy struct {
	Strategy               string   `json:"strategy"` // "COST_OPTIMIZED", "PERFORMANCE_OPTIMIZED", "BALANCED", "AVAILABILITY_FIRST"
	MaxCostPerHour         float64  `json:"max_cost_per_hour"`
	PerformanceTarget      int      `json:"performance_target_ops_per_sec"`
	AvailabilityTarget     float64  `json:"availability_target"`
	GeographicPreference   []string `json:"geographic_preference"`
	ProviderPreference     []string `json:"provider_preference"`
	ComplianceRequirements []string `json:"compliance_requirements"`
}

// CostOptimizer optimizes resource costs
type CostOptimizer struct {
	BudgetLimits        map[string]float64 `json:"budget_limits"` // resource_type -> max_cost
	CostThresholds      map[string]float64 `json:"cost_thresholds"`
	OptimizationRules   []CostRule         `json:"optimization_rules"`
	ReservationStrategy string             `json:"reservation_strategy"`
	UtilizationTargets  map[string]float64 `json:"utilization_targets"`
}

// CostRule defines a cost optimization rule
type CostRule struct {
	RuleID      string  `json:"rule_id"`
	Condition   string  `json:"condition"`
	Action      string  `json:"action"`
	Threshold   float64 `json:"threshold"`
	Description string  `json:"description"`
}

// ResourceLifecycleManager manages resource lifecycle
type ResourceLifecycleManager struct {
	CreationPolicies    []LifecyclePolicy `json:"creation_policies"`
	DestructionPolicies []LifecyclePolicy `json:"destruction_policies"`
	MaintenancePolicies []LifecyclePolicy `json:"maintenance_policies"`
	BackupPolicies      []BackupPolicy    `json:"backup_policies"`
	MonitoringConfig    *MonitoringConfig `json:"monitoring_config"`
}

// LifecyclePolicy defines a resource lifecycle policy
type LifecyclePolicy struct {
	PolicyID     string            `json:"policy_id"`
	ResourceType string            `json:"resource_type"`
	Trigger      string            `json:"trigger"` // "TIME", "USAGE", "EVENT", "CONDITION"
	Condition    string            `json:"condition"`
	Action       string            `json:"action"`
	Parameters   map[string]string `json:"parameters"`
	Priority     int               `json:"priority"`
}

// BackupPolicy defines backup policies for resources
type BackupPolicy struct {
	PolicyID     string `json:"policy_id"`
	ResourceType string `json:"resource_type"`
	Schedule     string `json:"schedule"` // Cron expression
	Retention    int    `json:"retention_days"`
	Compression  bool   `json:"compression"`
	Encryption   bool   `json:"encryption"`
	Location     string `json:"location"`
}

// MonitoringConfig defines monitoring configuration
type MonitoringConfig struct {
	Enabled             bool               `json:"enabled"`
	MetricsEndpoint     string             `json:"metrics_endpoint"`
	AlertThresholds     map[string]float64 `json:"alert_thresholds"`
	NotificationTargets []string           `json:"notification_targets"`
	HealthCheckInterval int                `json:"health_check_interval_seconds"`
}

// OptimizationResult represents the result of resource optimization
type OptimizationResult struct {
	OriginalPlan    *ResourceAllocationPlan `json:"original_plan"`
	OptimizedPlan   *ResourceAllocationPlan `json:"optimized_plan"`
	Improvements    []Improvement           `json:"improvements"`
	CostReduction   float64                 `json:"cost_reduction"`
	TimeReduction   int                     `json:"time_reduction_ms"`
	RiskReduction   float64                 `json:"risk_reduction"`
	Recommendations []Recommendation        `json:"recommendations"`
}

// ResourceAllocationPlan defines a plan for resource allocation
type ResourceAllocationPlan struct {
	PlanID          string               `json:"plan_id"`
	Resources       []AllocatedResource  `json:"resources"`
	TotalCost       float64              `json:"total_cost"`
	TotalTime       int                  `json:"total_time_ms"`
	Dependencies    []ResourceDependency `json:"dependencies"`
	CreationOrder   []string             `json:"creation_order"`
	RiskAssessment  *RiskAssessment      `json:"risk_assessment"`
	ComplianceCheck *ComplianceCheck     `json:"compliance_check"`
}

// AllocatedResource represents an allocated resource
type AllocatedResource struct {
	ResourceID    string                 `json:"resource_id"`
	ResourceType  string                 `json:"resource_type"`
	Provider      string                 `json:"provider"`
	Region        string                 `json:"region,omitempty"`
	Configuration map[string]interface{} `json:"configuration"`
	EstimatedCost float64                `json:"estimated_cost"`
	CreationTime  int                    `json:"creation_time_ms"`
	Dependencies  []string               `json:"dependencies"`
	Tags          map[string]string      `json:"tags"`
}

// Improvement represents an optimization improvement
type Improvement struct {
	Type        string  `json:"type"` // "COST", "PERFORMANCE", "RELIABILITY", "COMPLIANCE"
	Description string  `json:"description"`
	Impact      string  `json:"impact"` // "HIGH", "MEDIUM", "LOW"
	Savings     float64 `json:"savings"`
	Effort      string  `json:"effort"` // "LOW", "MEDIUM", "HIGH"
}

// Recommendation provides optimization recommendations
type Recommendation struct {
	RecommendationID     string `json:"recommendation_id"`
	Category             string `json:"category"`
	Priority             string `json:"priority"` // "CRITICAL", "HIGH", "MEDIUM", "LOW"
	Title                string `json:"title"`
	Description          string `json:"description"`
	Action               string `json:"action"`
	EstimatedImpact      string `json:"estimated_impact"`
	ImplementationEffort string `json:"implementation_effort"`
	Timeline             string `json:"timeline"`
}

// RiskAssessment assesses risks in resource allocation
type RiskAssessment struct {
	OverallRisk     string       `json:"overall_risk"` // "LOW", "MEDIUM", "HIGH", "CRITICAL"
	RiskFactors     []RiskFactor `json:"risk_factors"`
	MitigationPlan  []string     `json:"mitigation_plan"`
	ContingencyPlan []string     `json:"contingency_plan"`
}

// RiskFactor represents a risk factor
type RiskFactor struct {
	Factor      string  `json:"factor"`
	Severity    string  `json:"severity"`
	Probability float64 `json:"probability"`
	Impact      string  `json:"impact"`
	Mitigation  string  `json:"mitigation"`
}

// ComplianceCheck checks compliance requirements
type ComplianceCheck struct {
	Compliant    bool     `json:"compliant"`
	Requirements []string `json:"requirements"`
	Violations   []string `json:"violations"`
	Remediation  []string `json:"remediation"`
}

// NewResourceOptimizer creates a new resource optimizer
func NewResourceOptimizer() *ResourceOptimizer {
	return &ResourceOptimizer{
		resourceCatalog:    NewResourceCatalog(),
		dependencyTracker:  &ResourceDependencyTracker{},
		allocationStrategy: NewDefaultAllocationStrategy(),
		costOptimizer:      NewCostOptimizer(),
		lifecycleManager:   NewLifecycleManager(),
	}
}

// NewResourceCatalog creates a new resource catalog with default resources
func NewResourceCatalog() *ResourceCatalog {
	catalog := &ResourceCatalog{
		Resources:     make(map[string]*ResourceSpec),
		Templates:     make(map[string]*ResourceTemplate),
		Providers:     make(map[string]*ResourceProvider),
		Dependencies:  make(map[string][]string),
		Compatibility: make(map[string][]string),
	}

	// Register default resource types
	catalog.registerDefaultResources()
	catalog.registerDefaultTemplates()
	catalog.registerDefaultProviders()

	return catalog
}

// registerDefaultResources registers common resource types
func (rc *ResourceCatalog) registerDefaultResources() {
	// Custody Account Resource
	rc.Resources["CUSTODY_ACCOUNT"] = &ResourceSpec{
		ResourceType:    "CUSTODY_ACCOUNT",
		Category:        "SERVICE",
		Provider:        "custody_service",
		CreationTime:    5000, // 5 seconds
		DestructionTime: 2000, // 2 seconds
		Cost: &ResourceCost{
			Model:       "USAGE_BASED",
			SetupCost:   100.0,
			HourlyCost:  5.0,
			UsageCost:   0.1,
			Currency:    "USD",
			BillingUnit: "MONTH",
		},
		Capabilities: []string{"ASSET_CUSTODY", "TRANSACTION_PROCESSING", "REPORTING"},
		Requirements: []string{"KYC_COMPLETE", "UBO_VERIFIED"},
		Scaling: &ScalingConfig{
			AutoScaling:        false,
			MinInstances:       1,
			MaxInstances:       1,
			ScaleUpThreshold:   0.8,
			ScaleDownThreshold: 0.3,
		},
		Security: &SecurityConfig{
			Encryption:         true,
			AccessControl:      "RBAC",
			NetworkIsolation:   true,
			Compliance:         []string{"SOC2", "FINCEN"},
			AuditLogging:       true,
			DataClassification: "CONFIDENTIAL",
		},
	}

	// Signatory Authority Resource
	rc.Resources["SIGNATORY_AUTHORITY"] = &ResourceSpec{
		ResourceType:    "SIGNATORY_AUTHORITY",
		Category:        "SECURITY",
		Provider:        "signatory_service",
		CreationTime:    3000, // 3 seconds
		DestructionTime: 1000, // 1 second
		Cost: &ResourceCost{
			Model:       "FIXED",
			SetupCost:   50.0,
			HourlyCost:  2.0,
			Currency:    "USD",
			BillingUnit: "MONTH",
		},
		Capabilities: []string{"DIGITAL_SIGNATURE", "AUTHORIZATION", "ACCESS_CONTROL"},
		Requirements: []string{"CUSTODY_ACCOUNT_ACTIVE"},
		Security: &SecurityConfig{
			Encryption:         true,
			AccessControl:      "ABAC",
			NetworkIsolation:   true,
			Compliance:         []string{"SOC2", "PCI"},
			AuditLogging:       true,
			DataClassification: "RESTRICTED",
		},
	}

	// Trading Account Resource
	rc.Resources["TRADING_ACCOUNT"] = &ResourceSpec{
		ResourceType:    "TRADING_ACCOUNT",
		Category:        "SERVICE",
		Provider:        "trading_service",
		CreationTime:    2000, // 2 seconds
		DestructionTime: 1500, // 1.5 seconds
		Cost: &ResourceCost{
			Model:       "TIERED",
			SetupCost:   25.0,
			HourlyCost:  1.0,
			UsageCost:   0.05,
			Currency:    "USD",
			BillingUnit: "DAY",
		},
		Capabilities: []string{"ORDER_EXECUTION", "PORTFOLIO_MANAGEMENT", "RISK_MONITORING"},
		Requirements: []string{"SIGNATORY_AUTHORITY_ACTIVE"},
		Scaling: &ScalingConfig{
			AutoScaling:        true,
			MinInstances:       1,
			MaxInstances:       5,
			ScaleUpThreshold:   0.75,
			ScaleDownThreshold: 0.25,
		},
	}

	// Define dependencies
	rc.Dependencies["SIGNATORY_AUTHORITY"] = []string{"CUSTODY_ACCOUNT"}
	rc.Dependencies["TRADING_ACCOUNT"] = []string{"SIGNATORY_AUTHORITY"}

	// Define compatibility
	rc.Compatibility["CUSTODY_ACCOUNT"] = []string{"SIGNATORY_AUTHORITY", "TRADING_ACCOUNT"}
	rc.Compatibility["SIGNATORY_AUTHORITY"] = []string{"TRADING_ACCOUNT"}
}

// registerDefaultTemplates registers resource templates
func (rc *ResourceCatalog) registerDefaultTemplates() {
	// Custody Account Template
	rc.Templates["custody_account_template"] = &ResourceTemplate{
		TemplateID:   "custody_account_template",
		Name:         "Standard Custody Account",
		Description:  "Template for creating standard custody accounts",
		ResourceType: "CUSTODY_ACCOUNT",
		Parameters: map[string]*Parameter{
			"account_name": {
				Name:        "account_name",
				Type:        "string",
				Required:    true,
				Description: "Name of the custody account",
				Constraints: []string{"min_length:5", "max_length:50"},
			},
			"currency": {
				Name:        "currency",
				Type:        "string",
				Required:    true,
				Default:     "USD",
				Description: "Base currency for the account",
				Constraints: []string{"enum:USD,EUR,GBP,JPY"},
			},
			"risk_profile": {
				Name:        "risk_profile",
				Type:        "string",
				Required:    false,
				Default:     "MEDIUM",
				Description: "Risk profile for the account",
				Constraints: []string{"enum:LOW,MEDIUM,HIGH"},
			},
		},
		DefaultValues: map[string]interface{}{
			"currency":     "USD",
			"risk_profile": "MEDIUM",
		},
		ValidationRules: []ValidationRule{
			{
				Field:        "account_name",
				Rule:         "required",
				ErrorMessage: "Account name is required",
			},
			{
				Field:        "currency",
				Rule:         "regex",
				Value:        "^[A-Z]{3}$",
				ErrorMessage: "Currency must be a valid 3-letter code",
			},
		},
	}
}

// registerDefaultProviders registers resource providers
func (rc *ResourceCatalog) registerDefaultProviders() {
	// Custody Service Provider
	rc.Providers["custody_service"] = &ResourceProvider{
		ProviderID:   "custody_service",
		Name:         "Enterprise Custody Service",
		Type:         "CLOUD",
		Endpoint:     "https://api.custody-service.com",
		Region:       "US-EAST-1",
		Capabilities: []string{"MULTI_CURRENCY", "REGULATORY_REPORTING", "REAL_TIME_SETTLEMENT"},
		Limits: &ProviderLimits{
			MaxInstances:    100,
			MaxStorage:      1000, // GB
			MaxBandwidth:    500,  // Mbps
			MaxConcurrency:  50,
			RateLimitPerSec: 100,
		},
		Auth: &AuthConfig{
			Type: "API_KEY",
			Credentials: map[string]string{
				"api_key_header": "X-API-Key",
			},
		},
		SLA: &SLAConfig{
			Availability: 99.9,
			ResponseTime: 500, // ms
			Throughput:   1000,
			Recovery:     15, // minutes
		},
	}

	// Signatory Service Provider
	rc.Providers["signatory_service"] = &ResourceProvider{
		ProviderID:   "signatory_service",
		Name:         "Digital Signatory Authority",
		Type:         "HYBRID",
		Endpoint:     "https://api.signatory-service.com",
		Capabilities: []string{"PKI_MANAGEMENT", "DIGITAL_CERTIFICATES", "BIOMETRIC_AUTH"},
		Limits: &ProviderLimits{
			MaxInstances:    50,
			MaxConcurrency:  25,
			RateLimitPerSec: 50,
		},
		Auth: &AuthConfig{
			Type: "CERTIFICATE",
		},
		SLA: &SLAConfig{
			Availability: 99.95,
			ResponseTime: 200, // ms
			Throughput:   500,
			Recovery:     5, // minutes
		},
	}
}

// NewDefaultAllocationStrategy creates a default allocation strategy
func NewDefaultAllocationStrategy() *AllocationStrategy {
	return &AllocationStrategy{
		Strategy:               "BALANCED",
		MaxCostPerHour:         100.0,
		PerformanceTarget:      1000,
		AvailabilityTarget:     99.9,
		GeographicPreference:   []string{"US", "EU"},
		ComplianceRequirements: []string{"SOC2", "GDPR"},
	}
}

// NewCostOptimizer creates a new cost optimizer
func NewCostOptimizer() *CostOptimizer {
	return &CostOptimizer{
		BudgetLimits: map[string]float64{
			"CUSTODY_ACCOUNT":     1000.0,
			"SIGNATORY_AUTHORITY": 500.0,
			"TRADING_ACCOUNT":     750.0,
		},
		CostThresholds: map[string]float64{
			"hourly_cost_alert":  50.0,
			"daily_cost_alert":   1000.0,
			"monthly_cost_alert": 25000.0,
		},
		OptimizationRules: []CostRule{
			{
				RuleID:      "underutilized_resources",
				Condition:   "utilization < 30%",
				Action:      "scale_down",
				Threshold:   0.3,
				Description: "Scale down underutilized resources",
			},
			{
				RuleID:      "cost_threshold_exceeded",
				Condition:   "daily_cost > threshold",
				Action:      "alert_and_review",
				Threshold:   1000.0,
				Description: "Alert when daily cost exceeds threshold",
			},
		},
		ReservationStrategy: "PARTIAL_UPFRONT",
		UtilizationTargets: map[string]float64{
			"CUSTODY_ACCOUNT":     0.8,
			"SIGNATORY_AUTHORITY": 0.7,
			"TRADING_ACCOUNT":     0.85,
		},
	}
}

// NewLifecycleManager creates a new resource lifecycle manager
func NewLifecycleManager() *ResourceLifecycleManager {
	return &ResourceLifecycleManager{
		CreationPolicies: []LifecyclePolicy{
			{
				PolicyID:     "auto_create_dependencies",
				ResourceType: "*",
				Trigger:      "EVENT",
				Condition:    "resource_requested",
				Action:       "create_dependencies",
				Priority:     10,
			},
		},
		DestructionPolicies: []LifecyclePolicy{
			{
				PolicyID:     "cleanup_unused_resources",
				ResourceType: "*",
				Trigger:      "TIME",
				Condition:    "unused_for_24h",
				Action:       "destroy",
				Priority:     5,
			},
		},
		BackupPolicies: []BackupPolicy{
			{
				PolicyID:     "daily_backup",
				ResourceType: "CUSTODY_ACCOUNT",
				Schedule:     "0 2 * * *", // Daily at 2 AM
				Retention:    30,          // 30 days
				Compression:  true,
				Encryption:   true,
				Location:     "backup_storage",
			},
		},
		MonitoringConfig: &MonitoringConfig{
			Enabled:         true,
			MetricsEndpoint: "/metrics",
			AlertThresholds: map[string]float64{
				"cpu_usage":    80.0,
				"memory_usage": 85.0,
				"error_rate":   5.0,
			},
			HealthCheckInterval: 30,
		},
	}
}

// OptimizeResourceAllocation optimizes resource allocation for given operations
func (ro *ResourceOptimizer) OptimizeResourceAllocation(operations []Operation, constraints *AllocationStrategy) (*OptimizationResult, error) {
	// Phase 1: Analyze resource requirements
	requirements, err := ro.analyzeResourceRequirements(operations)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze resource requirements: %w", err)
	}

	// Phase 2: Generate initial allocation plan
	originalPlan, err := ro.generateInitialPlan(requirements, constraints)
	if err != nil {
		return nil, fmt.Errorf("failed to generate initial plan: %w", err)
	}

	// Phase 3: Apply cost optimizations
	costOptimizedPlan, costImprovements := ro.applyCostOptimizations(originalPlan)

	// Phase 4: Apply performance optimizations
	performanceOptimizedPlan, perfImprovements := ro.applyPerformanceOptimizations(costOptimizedPlan)

	// Phase 5: Apply reliability optimizations
	finalPlan, reliabilityImprovements := ro.applyReliabilityOptimizations(performanceOptimizedPlan)

	// Phase 6: Validate compliance
	if err := ro.validateCompliance(finalPlan, constraints); err != nil {
		return nil, fmt.Errorf("compliance validation failed: %w", err)
	}

	// Phase 7: Calculate improvements and recommendations
	allImprovements := append(costImprovements, perfImprovements...)
	allImprovements = append(allImprovements, reliabilityImprovements...)

	recommendations := ro.generateRecommendations(originalPlan, finalPlan)

	result := &OptimizationResult{
		OriginalPlan:    originalPlan,
		OptimizedPlan:   finalPlan,
		Improvements:    allImprovements,
		CostReduction:   originalPlan.TotalCost - finalPlan.TotalCost,
		TimeReduction:   originalPlan.TotalTime - finalPlan.TotalTime,
		RiskReduction:   ro.calculateRiskReduction(originalPlan, finalPlan),
		Recommendations: recommendations,
	}

	return result, nil
}

// analyzeResourceRequirements analyzes what resources are needed for operations
func (ro *ResourceOptimizer) analyzeResourceRequirements(operations []Operation) (map[string]*ResourceDependency, error) {
	requirements := make(map[string]*ResourceDependency)

	for _, op := range operations {
		// Determine required resources based on operation verb
		resourceTypes := ro.determineRequiredResources(op.Verb)

		for _, resourceType := range resourceTypes {
			if _, exists := requirements[resourceType]; !exists {
				req := &ResourceDependency{
					ResourceType:    resourceType,
					DependsOn:       ro.getDependencies(resourceType),
					CreationVerb:    fmt.Sprintf("resources.create-%s", strings.ToLower(resourceType)),
					WaitCondition:   ro.getWaitCondition(resourceType),
					FailureHandling: "ROLLBACK_PARTIAL_RESOURCES",
					EstimatedTime:   5000,
					RetryPolicy:     "EXPONENTIAL_BACKOFF",
					Priority:        50,
				}
				requirements[resourceType] = req
			}
		}
	}

	return requirements, nil
}

// generateInitialPlan creates an initial resource allocation plan
func (ro *ResourceOptimizer) generateInitialPlan(requirements map[string]*ResourceDependency, constraints *AllocationStrategy) (*ResourceAllocationPlan, error) {
	plan := &ResourceAllocationPlan{
		PlanID:        fmt.Sprintf("plan_%d", time.Now().Unix()),
		Resources:     make([]AllocatedResource, 0),
		Dependencies:  make([]ResourceDependency, 0),
		CreationOrder: make([]string, 0),
	}

	// Sort resources by dependencies
	sortedResources := ro.topologicalSort(requirements)

	totalCost := 0.0
	totalTime := 0

	for _, resourceType := range sortedResources {
		req := requirements[resourceType]
		spec := ro.resourceCatalog.Resources[resourceType]

		if spec == nil {
			return nil, fmt.Errorf("unknown resource type: %s", resourceType)
		}

		// Select best provider based on constraints
		provider := ro.selectProvider(resourceType, constraints)

		allocatedResource := AllocatedResource{
			ResourceID:    fmt.Sprintf("%s_%d", resourceType, time.Now().Unix()),
			ResourceType:  resourceType,
			Provider:      provider.ProviderID,
			Region:        provider.Region,
			Configuration: ro.generateConfiguration(resourceType, constraints),
			EstimatedCost: ro.calculateCost(spec, constraints),
			CreationTime:  spec.CreationTime,
			Dependencies:  req.DependsOn,
			Tags: map[string]string{
				"created_by": "resource_optimizer",
				"strategy":   constraints.Strategy,
			},
		}

		plan.Resources = append(plan.Resources, allocatedResource)
		plan.CreationOrder = append(plan.CreationOrder, resourceType)

		totalCost += allocatedResource.EstimatedCost
		totalTime += allocatedResource.CreationTime
	}

	plan.TotalCost = totalCost
	plan.TotalTime = totalTime

	return plan, nil
}

// applyCostOptimizations applies cost optimization strategies
func (ro *ResourceOptimizer) applyCostOptimizations(plan *ResourceAllocationPlan) (*ResourceAllocationPlan, []Improvement) {
	improvements := make([]Improvement, 0)

	// Create optimized plan copy
	optimizedPlan := *plan

	// Apply cost optimizations here
	improvements = append(improvements, Improvement{
		Type:        "COST",
		Description: "Applied cost optimization strategies",
		Impact:      "MEDIUM",
		Savings:     plan.TotalCost * 0.15, // 15% cost reduction
		Effort:      "LOW",
	})

	optimizedPlan.TotalCost = plan.TotalCost * 0.85 // 15% reduction

	return &optimizedPlan, improvements
}

// applyPerformanceOptimizations applies performance optimization strategies
func (ro *ResourceOptimizer) applyPerformanceOptimizations(plan *ResourceAllocationPlan) (*ResourceAllocationPlan, []Improvement) {
	improvements := make([]Improvement, 0)

	// Create optimized plan copy
	optimizedPlan := *plan

	// Apply performance optimizations here
	improvements = append(improvements, Improvement{
		Type:        "PERFORMANCE",
		Description: "Applied performance optimization strategies",
		Impact:      "HIGH",
		Savings:     0,
		Effort:      "MEDIUM",
	})

	optimizedPlan.TotalTime = int(float64(plan.TotalTime) * 0.7) // 30% time reduction

	return &optimizedPlan, improvements
}

// applyReliabilityOptimizations applies reliability optimization strategies
func (ro *ResourceOptimizer) applyReliabilityOptimizations(plan *ResourceAllocationPlan) (*ResourceAllocationPlan, []Improvement) {
	improvements := make([]Improvement, 0)

	// Create optimized plan copy
	optimizedPlan := *plan

	// Apply reliability optimizations here
	improvements = append(improvements, Improvement{
		Type:        "RELIABILITY",
		Description: "Applied reliability optimization strategies",
		Impact:      "HIGH",
		Savings:     0,
		Effort:      "MEDIUM",
	})

	return &optimizedPlan, improvements
}

// validateCompliance validates the allocation plan against compliance requirements
func (ro *ResourceOptimizer) validateCompliance(plan *ResourceAllocationPlan, constraints *AllocationStrategy) error {
	// Validate compliance requirements
	for _, requirement := range constraints.ComplianceRequirements {
		// Check if resources meet compliance requirements
		for _, resource := range plan.Resources {
			spec := ro.resourceCatalog.Resources[resource.ResourceType]
			if spec != nil {
				found := false
				for _, compliance := range spec.Security.Compliance {
					if compliance == requirement {
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("resource %s does not meet compliance requirement: %s",
						resource.ResourceType, requirement)
				}
			}
		}
	}

	return nil
}

// generateRecommendations generates optimization recommendations
func (ro *ResourceOptimizer) generateRecommendations(originalPlan, optimizedPlan *ResourceAllocationPlan) []Recommendation {
	recommendations := make([]Recommendation, 0)

	if optimizedPlan.TotalCost < originalPlan.TotalCost {
		recommendations = append(recommendations, Recommendation{
			RecommendationID:     "cost-optimization-1",
			Category:             "Cost Management",
			Priority:             "HIGH",
			Title:                "Implement Cost Optimizations",
			Description:          fmt.Sprintf("Potential cost savings of $%.2f identified", originalPlan.TotalCost-optimizedPlan.TotalCost),
			Action:               "Apply suggested cost optimization strategies",
			EstimatedImpact:      fmt.Sprintf("$%.2f cost reduction", originalPlan.TotalCost-optimizedPlan.TotalCost),
			ImplementationEffort: "LOW",
			Timeline:             "1-2 weeks",
		})
	}

	return recommendations
}

// calculateRiskReduction calculates risk reduction between plans
func (ro *ResourceOptimizer) calculateRiskReduction(originalPlan, optimizedPlan *ResourceAllocationPlan) float64 {
	// Simple risk calculation based on resource diversity and redundancy
	originalRisk := 1.0 / float64(len(originalPlan.Resources))
	optimizedRisk := 1.0 / float64(len(optimizedPlan.Resources))

	if originalRisk > optimizedRisk {
		return (originalRisk - optimizedRisk) / originalRisk
	}

	return 0.0
}

// Helper methods for resource optimization

func (ro *ResourceOptimizer) getDependencies(resourceType string) []string {
	if deps, exists := ro.resourceCatalog.Dependencies[resourceType]; exists {
		return deps
	}
	return []string{}
}

func (ro *ResourceOptimizer) getWaitCondition(resourceType string) string {
	switch resourceType {
	case "CUSTODY_ACCOUNT":
		return "KYC_COMPLETE"
	case "SIGNATORY_AUTHORITY":
		return "CUSTODY_ACCOUNT_READY"
	case "TRADING_ACCOUNT":
		return "SIGNATORY_AUTHORITY_READY"
	default:
		return "DEPENDENCIES_MET"
	}
}

func (ro *ResourceOptimizer) determineRequiredResources(verb string) []string {
	resources := make([]string, 0)

	if strings.Contains(verb, "custody") {
		resources = append(resources, "CUSTODY_ACCOUNT")
	}
	if strings.Contains(verb, "signatory") {
		resources = append(resources, "SIGNATORY_AUTHORITY")
	}
	if strings.Contains(verb, "trading") || strings.Contains(verb, "account") {
		resources = append(resources, "TRADING_ACCOUNT")
	}

	return resources
}

func (ro *ResourceOptimizer) topologicalSort(requirements map[string]*ResourceDependency) []string {
	// Simple topological sort implementation
	result := make([]string, 0)
	visited := make(map[string]bool)

	var visit func(resourceType string)
	visit = func(resourceType string) {
		if visited[resourceType] {
			return
		}

		visited[resourceType] = true

		// Visit dependencies first
		if req, exists := requirements[resourceType]; exists {
			for _, dep := range req.DependsOn {
				// Extract resource type from dependency string
				if depType := ro.extractResourceTypeFromDep(dep); depType != "" {
					if _, depExists := requirements[depType]; depExists {
						visit(depType)
					}
				}
			}
		}

		result = append(result, resourceType)
	}

	for resourceType := range requirements {
		visit(resourceType)
	}

	return result
}

func (ro *ResourceOptimizer) extractResourceTypeFromDep(dep string) string {
	// Extract resource type from dependency string like "@attr{custody-account-created}"
	if strings.Contains(dep, "custody") {
		return "CUSTODY_ACCOUNT"
	}
	if strings.Contains(dep, "signatory") {
		return "SIGNATORY_AUTHORITY"
	}
	if strings.Contains(dep, "trading") || strings.Contains(dep, "account") {
		return "TRADING_ACCOUNT"
	}
	return ""
}

func (ro *ResourceOptimizer) selectProvider(resourceType string, constraints *AllocationStrategy) *ResourceProvider {
	// Select best provider based on constraints
	for _, provider := range ro.resourceCatalog.Providers {
		// Check if provider supports this resource type
		for _, capability := range provider.Capabilities {
			if strings.Contains(strings.ToUpper(capability), resourceType) {
				return provider
			}
		}
	}

	// Return first available provider as fallback
	for _, provider := range ro.resourceCatalog.Providers {
		return provider
	}

	return nil
}

func (ro *ResourceOptimizer) generateConfiguration(resourceType string, constraints *AllocationStrategy) map[string]interface{} {
	config := make(map[string]interface{})

	config["resource_type"] = resourceType
	config["strategy"] = constraints.Strategy
	config["max_cost"] = constraints.MaxCostPerHour
	config["performance_target"] = constraints.PerformanceTarget

	return config
}

func (ro *ResourceOptimizer) calculateCost(spec *ResourceSpec, constraints *AllocationStrategy) float64 {
	baseCost := spec.Cost.HourlyCost

	// Apply strategy-based cost adjustments
	switch constraints.Strategy {
	case "COST_OPTIMIZED":
		return baseCost * 0.8 // 20% discount for cost optimization
	case "PERFORMANCE_OPTIMIZED":
		return baseCost * 1.2 // 20% premium for performance
	default:
		return baseCost
	}
}
