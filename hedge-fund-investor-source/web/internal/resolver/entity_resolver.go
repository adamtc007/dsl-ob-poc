// Package resolver provides entity resolution with fuzzy matching and user confirmation
package resolver

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// EntityType represents the type of entity being resolved
type EntityType string

const (
	EntityTypeInvestor    EntityType = "investor"
	EntityTypeFund        EntityType = "fund"
	EntityTypeShareClass  EntityType = "share_class"
	EntityTypeBeneficiary EntityType = "beneficial_owner"
)

// EntityResolver provides generalized entity resolution with fuzzy matching
type EntityResolver struct {
	dataStore EntityDataStore
}

// EntityDataStore defines the interface for entity search operations
type EntityDataStore interface {
	// SearchInvestorsByName searches for investors by legal name (fuzzy match)
	SearchInvestorsByName(ctx context.Context, name string) ([]map[string]interface{}, error)

	// SearchFundsByName searches for funds by fund name
	SearchFundsByName(ctx context.Context, name string) ([]map[string]interface{}, error)

	// SearchShareClassesByName searches for share classes by class name or fund+class
	SearchShareClassesByName(ctx context.Context, fundID, className string) ([]map[string]interface{}, error)

	// GetInvestorByID retrieves a single investor by ID
	GetInvestorByID(ctx context.Context, investorID string) (map[string]interface{}, error)

	// GetFundByID retrieves a single fund by ID
	GetFundByID(ctx context.Context, fundID string) (map[string]interface{}, error)

	// GetShareClassByID retrieves a single share class by ID
	GetShareClassByID(ctx context.Context, classID string) (map[string]interface{}, error)
}

// ResolutionResult represents the result of entity resolution
type ResolutionResult struct {
	// Resolution status
	Resolved     bool                   // True if entity was resolved to a single UUID
	RequiresUser bool                   // True if user confirmation needed
	EntityType   EntityType             // Type of entity being resolved
	EntityID     string                 // Resolved entity ID (if Resolved=true)
	Entity       map[string]interface{} // Full entity details (if Resolved=true)

	// For user prompts
	Candidates        []EntityMatch // Matching entities with scores
	PromptMessage     string        // Message to show user
	PendingActionType string        // "select_investor", "select_fund", etc.

	// Metadata
	SearchTerm string  // Original search term
	Confidence float64 // Confidence score (0.0-1.0)
	MatchType  string  // "exact", "fuzzy", "none"
}

// EntityMatch represents a matched entity with similarity score
type EntityMatch struct {
	Entity      map[string]interface{}
	Similarity  float64
	DisplayText string
}

// NewEntityResolver creates a new entity resolver
func NewEntityResolver(dataStore EntityDataStore) *EntityResolver {
	return &EntityResolver{
		dataStore: dataStore,
	}
}

// ResolveInvestor resolves an investor name to UUID
func (r *EntityResolver) ResolveInvestor(ctx context.Context, searchTerm string) (*ResolutionResult, error) {
	if searchTerm == "" {
		return &ResolutionResult{
			Resolved:     false,
			RequiresUser: false,
			EntityType:   EntityTypeInvestor,
			MatchType:    "none",
		}, nil
	}

	// Search database
	results, err := r.dataStore.SearchInvestorsByName(ctx, searchTerm)
	if err != nil {
		return nil, fmt.Errorf("failed to search investors: %w", err)
	}

	return r.processResults(EntityTypeInvestor, searchTerm, results, func(inv map[string]interface{}) string {
		// Format display text for investor
		name := getStringField(inv, "legal_name")
		invType := getStringField(inv, "type")
		status := getStringField(inv, "status")
		domicile := getStringField(inv, "domicile")
		createdAt := getStringField(inv, "created_at")

		var parts []string
		parts = append(parts, fmt.Sprintf("**%s**", name))
		if invType != "" {
			parts = append(parts, fmt.Sprintf("Type: %s", invType))
		}
		if domicile != "" {
			parts = append(parts, fmt.Sprintf("Domicile: %s", domicile))
		}
		if status != "" {
			parts = append(parts, fmt.Sprintf("Status: %s", status))
		}
		if createdAt != "" && len(createdAt) >= 10 {
			parts = append(parts, fmt.Sprintf("Created: %s", createdAt[:10]))
		}

		return strings.Join(parts, " | ")
	})
}

// ResolveFund resolves a fund name to UUID
func (r *EntityResolver) ResolveFund(ctx context.Context, searchTerm string) (*ResolutionResult, error) {
	if searchTerm == "" {
		return &ResolutionResult{
			Resolved:     false,
			RequiresUser: false,
			EntityType:   EntityTypeFund,
			MatchType:    "none",
		}, nil
	}

	// Search database
	results, err := r.dataStore.SearchFundsByName(ctx, searchTerm)
	if err != nil {
		return nil, fmt.Errorf("failed to search funds: %w", err)
	}

	return r.processResults(EntityTypeFund, searchTerm, results, func(fund map[string]interface{}) string {
		// Format display text for fund
		name := getStringField(fund, "fund_name")
		fundType := getStringField(fund, "fund_type")
		domicile := getStringField(fund, "domicile")
		currency := getStringField(fund, "currency")
		status := getStringField(fund, "status")

		var parts []string
		parts = append(parts, fmt.Sprintf("**%s**", name))
		if fundType != "" {
			parts = append(parts, fmt.Sprintf("Type: %s", fundType))
		}
		if domicile != "" {
			parts = append(parts, fmt.Sprintf("Domicile: %s", domicile))
		}
		if currency != "" {
			parts = append(parts, fmt.Sprintf("Currency: %s", currency))
		}
		if status != "" {
			parts = append(parts, fmt.Sprintf("Status: %s", status))
		}

		return strings.Join(parts, " | ")
	})
}

// ResolveShareClass resolves a share class name to UUID
func (r *EntityResolver) ResolveShareClass(ctx context.Context, fundID, className string) (*ResolutionResult, error) {
	if className == "" {
		return &ResolutionResult{
			Resolved:     false,
			RequiresUser: false,
			EntityType:   EntityTypeShareClass,
			MatchType:    "none",
		}, nil
	}

	// Search database
	results, err := r.dataStore.SearchShareClassesByName(ctx, fundID, className)
	if err != nil {
		return nil, fmt.Errorf("failed to search share classes: %w", err)
	}

	return r.processResults(EntityTypeShareClass, className, results, func(class map[string]interface{}) string {
		// Format display text for share class
		className := getStringField(class, "class_name")
		classType := getStringField(class, "class_type")
		currency := getStringField(class, "currency")
		minInvestment := getStringField(class, "min_initial_investment")

		var parts []string
		parts = append(parts, fmt.Sprintf("**Class %s**", className))
		if classType != "" {
			parts = append(parts, fmt.Sprintf("Type: %s", classType))
		}
		if currency != "" {
			parts = append(parts, fmt.Sprintf("Currency: %s", currency))
		}
		if minInvestment != "" {
			parts = append(parts, fmt.Sprintf("Min Investment: %s", minInvestment))
		}

		return strings.Join(parts, " | ")
	})
}

// processResults processes search results with fuzzy matching and scoring
func (r *EntityResolver) processResults(
	entityType EntityType,
	searchTerm string,
	results []map[string]interface{},
	formatFunc func(map[string]interface{}) string,
) (*ResolutionResult, error) {

	if len(results) == 0 {
		// No matches found
		return &ResolutionResult{
			Resolved:          false,
			RequiresUser:      true,
			EntityType:        entityType,
			SearchTerm:        searchTerm,
			MatchType:         "none",
			PromptMessage:     r.formatNoMatchPrompt(entityType, searchTerm),
			PendingActionType: fmt.Sprintf("confirm_create_%s", entityType),
		}, nil
	}

	// Calculate similarity scores
	matches := make([]EntityMatch, 0, len(results))
	for _, entity := range results {
		var nameField string
		switch entityType {
		case EntityTypeInvestor:
			nameField = getStringField(entity, "legal_name")
		case EntityTypeFund:
			nameField = getStringField(entity, "fund_name")
		case EntityTypeShareClass:
			nameField = getStringField(entity, "class_name")
		}

		similarity := calculateSimilarity(searchTerm, nameField)

		// Only include matches above threshold
		if similarity >= similarityThreshold {
			matches = append(matches, EntityMatch{
				Entity:      entity,
				Similarity:  similarity,
				DisplayText: formatFunc(entity),
			})
		}
	}

	if len(matches) == 0 {
		// No matches above threshold
		return &ResolutionResult{
			Resolved:          false,
			RequiresUser:      true,
			EntityType:        entityType,
			SearchTerm:        searchTerm,
			MatchType:         "none",
			PromptMessage:     r.formatNoMatchPrompt(entityType, searchTerm),
			PendingActionType: fmt.Sprintf("confirm_create_%s", entityType),
		}, nil
	}

	// Sort by similarity (highest first)
	for i := 0; i < len(matches); i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[j].Similarity > matches[i].Similarity {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}

	// Check if we can auto-select (high confidence, single match)
	if len(matches) == 1 && matches[0].Similarity >= autoSelectThreshold {
		// Auto-select with high confidence
		entityID := r.getEntityID(entityType, matches[0].Entity)
		return &ResolutionResult{
			Resolved:     true,
			RequiresUser: false,
			EntityType:   entityType,
			EntityID:     entityID,
			Entity:       matches[0].Entity,
			SearchTerm:   searchTerm,
			Confidence:   matches[0].Similarity,
			MatchType:    "exact",
		}, nil
	}

	// Multiple matches or low confidence - require user selection
	return &ResolutionResult{
		Resolved:          false,
		RequiresUser:      true,
		EntityType:        entityType,
		Candidates:        matches,
		SearchTerm:        searchTerm,
		Confidence:        matches[0].Similarity,
		MatchType:         "fuzzy",
		PromptMessage:     r.formatMultipleMatchPrompt(entityType, searchTerm, matches),
		PendingActionType: fmt.Sprintf("select_%s", entityType),
	}, nil
}

// formatMultipleMatchPrompt creates a user prompt for multiple matches
func (r *EntityResolver) formatMultipleMatchPrompt(entityType EntityType, searchTerm string, matches []EntityMatch) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("🔍 Found %d matching %s(s):\n\n", len(matches), r.getEntityDisplayName(entityType)))

	for i, match := range matches {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, match.DisplayText))
		if match.Similarity < 1.0 {
			sb.WriteString(fmt.Sprintf("   Match confidence: %.0f%%\n", match.Similarity*100))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("**Which one?** Reply with the number (1, 2, ...) or 'new' to create a new %s.", r.getEntityDisplayName(entityType)))

	return sb.String()
}

// formatNoMatchPrompt creates a user prompt when no matches found
func (r *EntityResolver) formatNoMatchPrompt(entityType EntityType, searchTerm string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("🔍 No existing %s found matching **\"%s\"**\n\n", r.getEntityDisplayName(entityType), searchTerm))
	sb.WriteString(fmt.Sprintf("**Create new %s?**\n", r.getEntityDisplayName(entityType)))
	sb.WriteString(fmt.Sprintf("- Name: %s\n\n", searchTerm))
	sb.WriteString("Reply 'yes' to create or 'no' to cancel.")

	return sb.String()
}

// getEntityID extracts the ID field from an entity based on type
func (r *EntityResolver) getEntityID(entityType EntityType, entity map[string]interface{}) string {
	switch entityType {
	case EntityTypeInvestor:
		return getStringField(entity, "investor_id")
	case EntityTypeFund:
		return getStringField(entity, "fund_id")
	case EntityTypeShareClass:
		return getStringField(entity, "class_id")
	default:
		return ""
	}
}

// getEntityDisplayName returns a user-friendly name for entity type
func (r *EntityResolver) getEntityDisplayName(entityType EntityType) string {
	switch entityType {
	case EntityTypeInvestor:
		return "investor"
	case EntityTypeFund:
		return "fund"
	case EntityTypeShareClass:
		return "share class"
	case EntityTypeBeneficiary:
		return "beneficial owner"
	default:
		return string(entityType)
	}
}

// Helper function to safely get string field from map
func getStringField(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// Similarity thresholds
const (
	similarityThreshold = 0.6  // Minimum similarity to consider a match
	autoSelectThreshold = 0.95 // Auto-select if similarity above this
)

// calculateSimilarity returns similarity score between 0.0 and 1.0
func calculateSimilarity(s1, s2 string) float64 {
	if s1 == "" || s2 == "" {
		return 0.0
	}

	// Exact match (case-insensitive)
	if strings.EqualFold(s1, s2) {
		return 1.0
	}

	// Calculate Levenshtein distance
	distance := levenshteinDistance(s1, s2)
	maxLen := max(len(s1), len(s2))

	// Convert distance to similarity (1.0 = identical, 0.0 = completely different)
	similarity := 1.0 - float64(distance)/float64(maxLen)

	// Bonus for substring matches
	s1Lower := strings.ToLower(s1)
	s2Lower := strings.ToLower(s2)
	if strings.Contains(s1Lower, s2Lower) || strings.Contains(s2Lower, s1Lower) {
		similarity = similarity * 1.2 // Boost by 20%
		if similarity > 1.0 {
			similarity = 1.0
		}
	}

	return similarity
}

// levenshteinDistance calculates the edit distance between two strings
func levenshteinDistance(s1, s2 string) int {
	s1Lower := strings.ToLower(s1)
	s2Lower := strings.ToLower(s2)

	if len(s1Lower) == 0 {
		return len(s2Lower)
	}
	if len(s2Lower) == 0 {
		return len(s1Lower)
	}

	matrix := make([][]int, len(s1Lower)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2Lower)+1)
		matrix[i][0] = i
	}
	for j := range matrix[0] {
		matrix[0][j] = j
	}

	for i := 1; i <= len(s1Lower); i++ {
		for j := 1; j <= len(s2Lower); j++ {
			cost := 0
			if s1Lower[i-1] != s2Lower[j-1] {
				cost = 1
			}
			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(s1Lower)][len(s2Lower)]
}

// min returns the minimum of three integers
func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// PendingResolution represents an entity resolution waiting for user input
type PendingResolution struct {
	EntityType EntityType
	SearchTerm string
	Result     *ResolutionResult
	CreatedAt  time.Time
}
