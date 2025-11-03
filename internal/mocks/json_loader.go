package mocks

import (
	"encoding/json"
	"fmt"
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

// LoadCBUs loads CBU mock data from JSON
func (j *JSONDataLoader) LoadCBUs() ([]store.CBU, error) {
	filePath := filepath.Join(j.basePath, "cbus.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CBUs file: %w", err)
	}

	var cbus []store.CBU
	if err := json.Unmarshal(data, &cbus); err != nil {
		return nil, fmt.Errorf("failed to unmarshal CBUs: %w", err)
	}

	return cbus, nil
}

// LoadRoles loads role mock data from JSON
func (j *JSONDataLoader) LoadRoles() ([]store.Role, error) {
	filePath := filepath.Join(j.basePath, "roles.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read roles file: %w", err)
	}

	var roles []store.Role
	if err := json.Unmarshal(data, &roles); err != nil {
		return nil, fmt.Errorf("failed to unmarshal roles: %w", err)
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
	if err := json.Unmarshal(data, &entityTypes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal entity types: %w", err)
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
	if err := json.Unmarshal(data, &entities); err != nil {
		return nil, fmt.Errorf("failed to unmarshal entities: %w", err)
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
	if err := json.Unmarshal(data, &companies); err != nil {
		return nil, fmt.Errorf("failed to unmarshal limited companies: %w", err)
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
	if err := json.Unmarshal(data, &partnerships); err != nil {
		return nil, fmt.Errorf("failed to unmarshal partnerships: %w", err)
	}

	return partnerships, nil
}

// LoadIndividuals loads individual mock data from JSON
func (j *JSONDataLoader) LoadIndividuals() ([]store.Individual, error) {
	filePath := filepath.Join(j.basePath, "entity_individuals.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read individuals file: %w", err)
	}

	var individuals []store.Individual
	if err := json.Unmarshal(data, &individuals); err != nil {
		return nil, fmt.Errorf("failed to unmarshal individuals: %w", err)
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
	if err := json.Unmarshal(data, &relationships); err != nil {
		return nil, fmt.Errorf("failed to unmarshal CBU entity roles: %w", err)
	}

	return relationships, nil
}

// LoadProducts loads product mock data from JSON
func (j *JSONDataLoader) LoadProducts() ([]store.Product, error) {
	filePath := filepath.Join(j.basePath, "products.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read products file: %w", err)
	}

	var products []store.Product
	if err := json.Unmarshal(data, &products); err != nil {
		return nil, fmt.Errorf("failed to unmarshal products: %w", err)
	}

	return products, nil
}

// LoadServices loads service mock data from JSON
func (j *JSONDataLoader) LoadServices() ([]store.Service, error) {
	filePath := filepath.Join(j.basePath, "services.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read services file: %w", err)
	}

	var services []store.Service
	if err := json.Unmarshal(data, &services); err != nil {
		return nil, fmt.Errorf("failed to unmarshal services: %w", err)
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
	if err := json.Unmarshal(data, &resources); err != nil {
		return nil, fmt.Errorf("failed to unmarshal prod resources: %w", err)
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
	if err := json.Unmarshal(data, &relations); err != nil {
		return nil, fmt.Errorf("failed to unmarshal product services: %w", err)
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
	if err := json.Unmarshal(data, &relations); err != nil {
		return nil, fmt.Errorf("failed to unmarshal service resources: %w", err)
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
	if err := json.Unmarshal(data, &attributes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal dictionary: %w", err)
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
	filePath := filepath.Join(j.basePath, "attribute_values.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read attribute values file: %w", err)
	}

	var values []AttributeValue
	if err := json.Unmarshal(data, &values); err != nil {
		return nil, fmt.Errorf("failed to unmarshal attribute values: %w", err)
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
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("failed to unmarshal DSL records: %w", err)
	}

	return records, nil
}