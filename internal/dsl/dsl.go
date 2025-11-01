package dsl

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"dsl-ob-poc/internal/store"
)

// This package is the "Go internal lib" for generating the DSL.

// --- State 1: Create Case ---

func CreateCase(cbuID, naturePurpose string) string {
	var b strings.Builder
	b.WriteString("(case.create\n")
	b.WriteString(fmt.Sprintf("  (cbu.id %q)\n", cbuID))
	b.WriteString(fmt.Sprintf("  (nature-purpose %q)\n", naturePurpose))
	b.WriteString(")")
	return b.String()
}

// --- State 2: Add Products ---

func AddProducts(currentDSL string, products []*store.Product) (string, error) {
	if len(products) == 0 {
		return currentDSL, nil // No change
	}

	productExprs := make([]string, 0, len(products))
	for _, p := range products {
		productExprs = append(productExprs, fmt.Sprintf("%q", p.Name))
	}

	var b strings.Builder
	b.WriteString(currentDSL)
	b.WriteString("\n\n")
	b.WriteString("(products.add ")
	b.WriteString(strings.Join(productExprs, " "))
	b.WriteString(")")

	return b.String(), nil
}

// Simple parser for POC
var productRegex = regexp.MustCompile(`\(products\.add\s+(.*?)\)`)
var naturePurposeRegex = regexp.MustCompile(`\(nature-purpose\s+"(.*?)"\)`)

func ParseProductNames(dsl string) ([]string, error) {
	matches := productRegex.FindStringSubmatch(dsl)
	if len(matches) < 2 {
		return nil, fmt.Errorf("no (products.add ...) block found in DSL")
	}

	namesStr := matches[1] // e.g., "CUSTODY" "FUND_ACCOUNTING"
	namesStr = strings.ReplaceAll(namesStr, "\"", "")
	names := strings.Fields(namesStr) // Use Fields to split on whitespace

	if len(names) == 0 {
		return nil, fmt.Errorf("no product names found in block")
	}
	return names, nil
}

func ParseNaturePurpose(dsl string) (string, error) {
	matches := naturePurposeRegex.FindStringSubmatch(dsl)
	if len(matches) < 2 {
		return "", fmt.Errorf("no (nature-purpose ...) block found in DSL")
	}
	return matches[1], nil
}

// KYCRequirements captures the AI agent output used for DSL generation.
type KYCRequirements struct {
	Documents     []string
	Jurisdictions []string
}

func AddKYCRequirements(currentDSL string, reqs KYCRequirements) (string, error) {
	if len(reqs.Documents) == 0 && len(reqs.Jurisdictions) == 0 {
		return "", fmt.Errorf("no KYC requirements provided")
	}

	var b strings.Builder
	b.WriteString(currentDSL)
	b.WriteString("\n\n")
	b.WriteString("(kyc.start\n")

	if len(reqs.Documents) > 0 {
		docs := append([]string(nil), reqs.Documents...)
		sort.Strings(docs)
		b.WriteString("  (documents\n")
		for _, doc := range docs {
			b.WriteString(fmt.Sprintf("    (document %q)\n", doc))
		}
		b.WriteString("  )\n")
	}

	if len(reqs.Jurisdictions) > 0 {
		jurisdictions := append([]string(nil), reqs.Jurisdictions...)
		sort.Strings(jurisdictions)
		b.WriteString("  (jurisdictions\n")
		for _, jurisdiction := range jurisdictions {
			b.WriteString(fmt.Sprintf("    (jurisdiction %q)\n", jurisdiction))
		}
		b.WriteString("  )\n")
	}

	b.WriteString(")")

	return b.String(), nil
}

// --- State 3: Discover Services ---

// ServiceDiscoveryPlan holds data for the service discovery step
type ServiceDiscoveryPlan struct {
	ProductServices map[string][]store.Service
}

func AddDiscoveredServices(currentDSL string, plan ServiceDiscoveryPlan) (string, error) {
	var b strings.Builder
	b.WriteString(currentDSL)
	b.WriteString("\n\n")

	// Append (services.discover)
	b.WriteString("(services.discover\n")
	for product, services := range plan.ProductServices {
		b.WriteString(fmt.Sprintf("  (for.product %q\n", product))
		// Use a map to de-duplicate service names
		serviceNames := make(map[string]bool)
		for _, service := range services {
			serviceNames[service.Name] = true
		}
		for serviceName := range serviceNames {
			b.WriteString(fmt.Sprintf("    (service %q)\n", serviceName))
		}
		b.WriteString("  )\n")
	}
	b.WriteString(")")

	return b.String(), nil
}

// Simple parser for POC
var serviceRegex = regexp.MustCompile(`\(service\s+"(.*?)"\)`)

func ParseServiceNames(dsl string) ([]string, error) {
	matches := serviceRegex.FindAllStringSubmatch(dsl, -1)
	if len(matches) == 0 {
		return nil, fmt.Errorf("no (service ...) blocks found in DSL")
	}

	serviceNames := make(map[string]bool) // Use map to de-duplicate
	for _, match := range matches {
		if len(match) >= 2 {
			serviceNames[match[1]] = true
		}
	}

	names := make([]string, 0, len(serviceNames))
	for name := range serviceNames {
		names = append(names, name)
	}

	if len(names) == 0 {
		return nil, fmt.Errorf("no service names found")
	}
	return names, nil
}

// --- State 4: Discover Resources ---

// ResourceDiscoveryPlan holds data for the resource discovery step
type ResourceDiscoveryPlan struct {
	ServiceResources   map[string][]store.ProdResource
	ResourceAttributes map[string][]store.Attribute
}

func AddDiscoveredResources(currentDSL string, plan ResourceDiscoveryPlan) (string, error) {
	var b strings.Builder
	b.WriteString(currentDSL)
	b.WriteString("\n\n")

	// Append (resources.plan)
	b.WriteString("(resources.plan\n")

	// Use a map to find all unique resources
	allResources := make(map[string]store.ProdResource)
	for _, resources := range plan.ServiceResources {
		for _, res := range resources {
			allResources[res.ResourceID] = res
		}
	}

	for _, resource := range allResources {
		b.WriteString(fmt.Sprintf("  (resource.create %q\n", resource.Name))
		b.WriteString(fmt.Sprintf("    (owner %q)\n", resource.Owner))

		attributes := plan.ResourceAttributes[resource.DictionaryID]
		for _, attr := range attributes {
			b.WriteString(fmt.Sprintf("    (attr.%q)\n", attr.Name))
		}
		b.WriteString("  )\n")
	}
	b.WriteString(")")

	return b.String(), nil
}
