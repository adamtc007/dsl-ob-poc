package mocks

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"dsl-ob-poc/internal/store"
)

// JSONDataLoader loads mock data from JSON files
type JSONDataLoader struct {
	basePath string
}

// NewJSONDataLoader creates a new JSON data loader
func NewJSONDataLoader(basePath string) *JSONDataLoader {
	return &JSONDataLoader{basePath: basePath}
}

// loadJSONFile loads a JSON file and unmarshals it into the target
func (j *JSONDataLoader) loadJSONFile(filename string, target interface{}, required bool) error {
	filePath := filepath.Join(j.basePath, filename)
	data, err := os.ReadFile(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) && !required {
			log.Printf("Optional mock data file not found: %s (continuing with empty data)", filePath)
			return nil // Return nil for optional missing files
		}
		return fmt.Errorf("failed to read %s: %w (file path: %s)", filename, err, filePath)
	}

	if len(data) == 0 {
		if !required {
			log.Printf("Mock data file is empty: %s (continuing with empty data)", filePath)
			return nil
		}
		return fmt.Errorf("mock data file is empty: %s", filePath)
	}

	if unmarshalErr := json.Unmarshal(data, target); unmarshalErr != nil {
		return fmt.Errorf("failed to unmarshal %s: %w (file path: %s)", filename, unmarshalErr, filePath)
	}

	return nil
}

// LoadCBUs loads CBU mock data from JSON
func (j *JSONDataLoader) LoadCBUs() ([]store.CBU, error) {
	var cbus []store.CBU
	if err := j.loadJSONFile("cbus.json", &cbus, true); err != nil {
		return nil, err
	}
	return cbus, nil
}

// LoadRoles loads role mock data from JSON
func (j *JSONDataLoader) LoadRoles() ([]store.Role, error) {
	var roles []store.Role
	if err := j.loadJSONFile("roles.json", &roles, true); err != nil {
		return nil, err
	}
	return roles, nil
}

// LoadEntityTypes loads entity type mock data from JSON
func (j *JSONDataLoader) LoadEntityTypes() ([]store.EntityType, error) {
	filePath := filepath.Join(j.basePath, "entity_types.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read entity types file: %w", err)
	}

	var entityTypes []store.EntityType
	if unmarshalErr := json.Unmarshal(data, &entityTypes); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to unmarshal entity types: %w", unmarshalErr)
	}

	return entityTypes, nil
}

// LoadEntities loads entity mock data from JSON
func (j *JSONDataLoader) LoadEntities() ([]store.Entity, error) {
	filePath := filepath.Join(j.basePath, "entities.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read entities file: %w", err)
	}

	var entities []store.Entity
	if unmarshalErr := json.Unmarshal(data, &entities); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to unmarshal entities: %w", unmarshalErr)
	}

	return entities, nil
}

// LoadLimitedCompanies loads limited company mock data from JSON
func (j *JSONDataLoader) LoadLimitedCompanies() ([]store.LimitedCompany, error) {
	filePath := filepath.Join(j.basePath, "entity_limited_companies.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read limited companies file: %w", err)
	}

	var companies []store.LimitedCompany
	if unmarshalErr := json.Unmarshal(data, &companies); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to unmarshal limited companies: %w", unmarshalErr)
	}

	return companies, nil
}

// LoadPartnerships loads partnership mock data from JSON
func (j *JSONDataLoader) LoadPartnerships() ([]store.Partnership, error) {
	filePath := filepath.Join(j.basePath, "entity_partnerships.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read partnerships file: %w", err)
	}

	var partnerships []store.Partnership
	if unmarshalErr := json.Unmarshal(data, &partnerships); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to unmarshal partnerships: %w", unmarshalErr)
	}

	return partnerships, nil
}

// LoadIndividuals loads individual mock data from JSON
func (j *JSONDataLoader) LoadIndividuals() ([]store.Individual, error) {
	var individuals []store.Individual
	if err := j.loadJSONFile("entity_individuals.json", &individuals, false); err != nil {
		return nil, err
	}
	return individuals, nil
}

// LoadCBUEntityRoles loads CBU entity role relationships from JSON
func (j *JSONDataLoader) LoadCBUEntityRoles() ([]store.CBUEntityRole, error) {
	filePath := filepath.Join(j.basePath, "cbu_entity_roles.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CBU entity roles file: %w", err)
	}

	var relationships []store.CBUEntityRole
	if unmarshalErr := json.Unmarshal(data, &relationships); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to unmarshal CBU entity roles: %w", unmarshalErr)
	}

	return relationships, nil
}

// LoadProducts loads product mock data from JSON
func (j *JSONDataLoader) LoadProducts() ([]store.Product, error) {
	var products []store.Product
	if err := j.loadJSONFile("products.json", &products, true); err != nil {
		return nil, err
	}
	return products, nil
}

// LoadServices loads service mock data from JSON
func (j *JSONDataLoader) LoadServices() ([]store.Service, error) {
	var services []store.Service
	if err := j.loadJSONFile("services.json", &services, true); err != nil {
		return nil, err
	}
	return services, nil
}

// LoadProdResources loads production resource mock data from JSON
func (j *JSONDataLoader) LoadProdResources() ([]store.ProdResource, error) {
	filePath := filepath.Join(j.basePath, "prod_resources.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read prod resources file: %w", err)
	}

	var resources []store.ProdResource
	if unmarshalErr := json.Unmarshal(data, &resources); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to unmarshal prod resources: %w", unmarshalErr)
	}

	return resources, nil
}

// ProductServiceRelation represents a product-service relationship
type ProductServiceRelation struct {
	ProductID string `json:"product_id"`
	ServiceID string `json:"service_id"`
}

// LoadProductServices loads product-service relationships from JSON
func (j *JSONDataLoader) LoadProductServices() ([]ProductServiceRelation, error) {
	filePath := filepath.Join(j.basePath, "product_services.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read product services file: %w", err)
	}

	var relations []ProductServiceRelation
	if unmarshalErr := json.Unmarshal(data, &relations); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to unmarshal product services: %w", unmarshalErr)
	}

	return relations, nil
}

// ServiceResourceRelation represents a service-resource relationship
type ServiceResourceRelation struct {
	ServiceID  string `json:"service_id"`
	ResourceID string `json:"resource_id"`
}

// LoadServiceResources loads service-resource relationships from JSON
func (j *JSONDataLoader) LoadServiceResources() ([]ServiceResourceRelation, error) {
	filePath := filepath.Join(j.basePath, "service_resources.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read service resources file: %w", err)
	}

	var relations []ServiceResourceRelation
	if unmarshalErr := json.Unmarshal(data, &relations); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to unmarshal service resources: %w", unmarshalErr)
	}

	return relations, nil
}

// LoadDictionary loads dictionary attribute mock data from JSON
func (j *JSONDataLoader) LoadDictionary() ([]store.Attribute, error) {
	filePath := filepath.Join(j.basePath, "dictionary.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read dictionary file: %w", err)
	}

	var attributes []store.Attribute
	if unmarshalErr := json.Unmarshal(data, &attributes); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to unmarshal dictionary: %w", unmarshalErr)
	}

	return attributes, nil
}

// AttributeValue represents an attribute value record
type AttributeValue struct {
	AVID        string `json:"av_id"`
	CBUID       string `json:"cbu_id"`
	DSLObID     string `json:"dsl_ob_id"`
	DSLVersion  int    `json:"dsl_version"`
	AttributeID string `json:"attribute_id"`
	Value       string `json:"value"`
	State       string `json:"state"`
	Source      string `json:"source"`
	ObservedAt  string `json:"observed_at"`
}

// LoadAttributeValues loads attribute values from JSON
func (j *JSONDataLoader) LoadAttributeValues() ([]AttributeValue, error) {
	var values []AttributeValue
	if err := j.loadJSONFile("attribute_values.json", &values, false); err != nil {
		return nil, err
	}
	return values, nil
}

// DSLRecord represents a DSL record
type DSLRecord struct {
	VersionID string `json:"version_id"`
	CBUID     string `json:"cbu_id"`
	DSLText   string `json:"dsl_text"`
	CreatedAt string `json:"created_at"`
}

// LoadDSLRecords loads DSL records from JSON
func (j *JSONDataLoader) LoadDSLRecords() ([]DSLRecord, error) {
	filePath := filepath.Join(j.basePath, "dsl_ob.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read DSL records file: %w", err)
	}

	var records []DSLRecord
	if unmarshalErr := json.Unmarshal(data, &records); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to unmarshal DSL records: %w", unmarshalErr)
	}

	return records, nil
}
