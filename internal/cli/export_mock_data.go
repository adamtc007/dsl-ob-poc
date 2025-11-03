package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"dsl-ob-poc/internal/datastore"
)

// RunExportMockData exports existing database records to JSON files
func RunExportMockData(ctx context.Context, ds datastore.DataStore, args []string) error {
	outputDir := "data/mocks"

	// Parse arguments for custom output directory
	for _, arg := range args {
		if len(arg) > 7 && arg[:7] == "--dir=" {
			outputDir = arg[7:]
		}
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	fmt.Printf("Exporting database records to %s...\n", outputDir)

	// Export CBUs
	if err := exportCBUs(ctx, ds, outputDir); err != nil {
		return fmt.Errorf("failed to export CBUs: %w", err)
	}

	// Export roles
	if err := exportRoles(ctx, ds, outputDir); err != nil {
		return fmt.Errorf("failed to export roles: %w", err)
	}

	// Export products
	if err := exportProducts(ctx, ds, outputDir); err != nil {
		return fmt.Errorf("failed to export products: %w", err)
	}

	// Export services
	if err := exportServices(ctx, ds, outputDir); err != nil {
		return fmt.Errorf("failed to export services: %w", err)
	}

	// Export dictionary
	if err := exportDictionary(ctx, ds, outputDir); err != nil {
		return fmt.Errorf("failed to export dictionary: %w", err)
	}

	// Export DSL records
	if err := exportDSLRecords(ctx, ds, outputDir); err != nil {
		return fmt.Errorf("failed to export DSL records: %w", err)
	}

	fmt.Println("Export completed successfully!")
	return nil
}

func exportCBUs(ctx context.Context, ds datastore.DataStore, outputDir string) error {
	cbus, err := ds.ListCBUs(ctx)
	if err != nil {
		return err
	}

	if len(cbus) == 0 {
		fmt.Println("No CBUs found - keeping existing mock data")
		return nil
	}

	filePath := filepath.Join(outputDir, "cbus.json")
	return writeJSONFile(filePath, cbus, "CBUs")
}

func exportRoles(ctx context.Context, ds datastore.DataStore, outputDir string) error {
	roles, err := ds.ListRoles(ctx)
	if err != nil {
		return err
	}

	if len(roles) == 0 {
		fmt.Println("No roles found - keeping existing mock data")
		return nil
	}

	filePath := filepath.Join(outputDir, "roles.json")
	return writeJSONFile(filePath, roles, "roles")
}

func exportProducts(ctx context.Context, ds datastore.DataStore, outputDir string) error {
	// Note: The real ds doesn't have a GetAllProducts method
	// This would require adding that method or querying directly
	fmt.Println("Products export not implemented - keeping existing mock data")
	return nil
}

func exportServices(ctx context.Context, ds datastore.DataStore, outputDir string) error {
	// Note: We don't have a GetAllServices method, so we'll use a basic query
	// This is a simplified implementation for the export utility
	fmt.Println("Services export not implemented - keeping existing mock data")
	return nil
}

func exportDictionary(ctx context.Context, ds datastore.DataStore, outputDir string) error {
	// Note: We don't have a GetAllDictionaryAttributes method, so we'll skip this
	// This would require adding a method to the ds interface
	fmt.Println("Dictionary export not implemented - keeping existing mock data")
	return nil
}

func exportDSLRecords(ctx context.Context, ds datastore.DataStore, outputDir string) error {
	// Note: We don't have a GetAllDSLRecords method, so we'll skip this
	// This would require adding a method to the ds interface
	fmt.Println("DSL records export not implemented - keeping existing mock data")
	return nil
}

func writeJSONFile(filePath string, data interface{}, dataType string) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal %s: %w", dataType, err)
	}

	if err := os.WriteFile(filePath, jsonData, 0o644); err != nil {
		return fmt.Errorf("failed to write %s file: %w", dataType, err)
	}

	fmt.Printf("Exported %s to %s\n", dataType, filePath)
	return nil
}
