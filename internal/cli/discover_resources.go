package cli

import (
	"context"
	"flag"
	"fmt"
	"log"

	"dsl-ob-poc/internal/datastore"
	"dsl-ob-poc/internal/dictionary"
	"dsl-ob-poc/internal/dsl"
	"dsl-ob-poc/internal/store"
)

// RunDiscoverResources handles the 'discover-resources' command (Step 5).
func RunDiscoverResources(ctx context.Context, ds datastore.DataStore, args []string) error {
	fs := flag.NewFlagSet("discover-resources", flag.ExitOnError)
	cbuID := fs.String("cbu", "", "The CBU ID of the case to discover (required)")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	if *cbuID == "" {
		fs.Usage()
		return fmt.Errorf("error: --cbu flag is required")
	}

	log.Printf("Starting resource discovery (Step 5) for CBU: %s", *cbuID)

	// 1. Get the latest DSL (should be v4)
	currentDSL, err := ds.GetLatestDSL(ctx, *cbuID)
	if err != nil {
		return err
	}

	// 2. Parse *service* names from the DSL
	serviceNames, err := dsl.ParseServiceNames(currentDSL)
	if err != nil {
		return fmt.Errorf("failed to parse services from DSL: %w. Run 'discover-services' first", err)
	}
	log.Printf("Found %d services in DSL: %v", len(serviceNames), serviceNames)

	// 3. Discover all resources and attributes from the catalog
	serviceResourceMap := make(map[string][]store.ProdResource)
	dictionaryAttributeMap := make(map[string][]dictionary.Attribute)

	allResources := make(map[string]store.ProdResource)

	for _, serviceName := range serviceNames {
		service, getErr := ds.GetServiceByName(ctx, serviceName)
		if getErr != nil {
			return getErr
		}

		resources, getErr := ds.GetResourcesForService(ctx, service.ServiceID)
		if getErr != nil {
			return getErr
		}
		serviceResourceMap[service.Name] = resources

		for _, resource := range resources {
			// Add to unique map
			allResources[resource.ResourceID] = resource

			// If resource has a dictionary group, get its attributes
			if resource.DictionaryGroup != "" {
				// Only fetch if we haven't already
				if _, ok := dictionaryAttributeMap[resource.DictionaryGroup]; !ok {
						// TODO: GetAttributesForDictionaryGroup not implemented in DataStore interface
					// Will implement in next session for proper DSL CRUD
					var storeAttributes []store.Attribute // empty for now
					attrErr := error(nil)
					if attrErr != nil {
						return attrErr
					}
					// Convert store.Attribute to dictionary.Attribute
					dictAttributes := make([]dictionary.Attribute, len(storeAttributes))
					for i, attr := range storeAttributes {
						dictAttributes[i] = dictionary.Attribute{
							AttributeID:     attr.AttributeID,
							Name:            attr.Name,
							LongDescription: attr.LongDescription,
							GroupID:         attr.GroupID,
							Mask:            attr.Mask,
							Domain:          attr.Domain,
							Vector:          attr.Vector,
							// Note: Source and Sink are JSON strings in store.Attribute
							// but SourceMetadata/SinkMetadata structs in dictionary.Attribute
							// For now, we'll leave them as empty structs since this is POC
						}
					}
					dictionaryAttributeMap[resource.DictionaryGroup] = dictAttributes
				}
			}
		}
	}
	log.Printf("Discovery complete: found %d unique resources.", len(allResources))

	// 4. Generate the new DSL with the discovered resources plan
	plan := dsl.ResourceDiscoveryPlan{
		ServiceResources:   serviceResourceMap,
		ResourceAttributes: dictionaryAttributeMap,
	}

	newDSL, err := dsl.AddDiscoveredResources(currentDSL, plan)
	if err != nil {
		return fmt.Errorf("failed to generate new DSL: %w", err)
	}

	// 5. Save the new DSL version (v5)
	versionID, err := ds.InsertDSL(ctx, *cbuID, newDSL)
	if err != nil {
		return fmt.Errorf("failed to save new DSL version: %w", err)
	}

	fmt.Printf("Created new case version (v5): %s\n", versionID)
	fmt.Println("---")
	fmt.Println(newDSL)
	fmt.Println("---")

	return nil
}
