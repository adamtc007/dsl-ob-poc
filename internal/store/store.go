package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"dsl-ob-poc/internal/dictionary"
	_ "github.com/lib/pq"
)

// Store represents the database connection and operations.
type Store struct {
	db *sql.DB
}

// CBU represents a Client Business Unit in the catalog.
type CBU struct {
	CBUID         string `json:"cbu_id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	NaturePurpose string `json:"nature_purpose"`
}

// Product represents a product in the catalog.
type Product struct {
	ProductID   string `json:"product_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Service represents a service in the catalog.
type Service struct {
	ServiceID   string `json:"service_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ProdResource represents a resource required by products/services.
type ProdResource struct {
	ResourceID      string `json:"resource_id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	Owner           string `json:"owner"`
	DictionaryGroup string `json:"dictionary_group"`
}

// Attribute represents an attribute in the dictionary (v3 schema).
type Attribute struct {
	AttributeID     string
	Name            string
	LongDescription string
	GroupID         string
	Mask            string
	Domain          string
	Vector          string
	Source          string // JSON string
	Sink            string // JSON string
}

// Role represents a role that entities can play within a CBU.
type Role struct {
	RoleID      string `json:"role_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// EntityType represents the different types of entities.
type EntityType struct {
	EntityTypeID string `json:"entity_type_id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	TableName    string `json:"table_name"`
}

// Entity represents an entity in the central registry.
type Entity struct {
	EntityID     string `json:"entity_id"`
	EntityTypeID string `json:"entity_type_id"`
	ExternalID   string `json:"external_id"`
	Name         string `json:"name"`
}

// CBUEntityRole represents the relationship between CBUs, entities, and roles.
type CBUEntityRole struct {
	CBUEntityRoleID string `json:"cbu_entity_role_id"`
	CBUID           string `json:"cbu_id"`
	EntityID        string `json:"entity_id"`
	RoleID          string `json:"role_id"`
}

// LimitedCompany represents a limited company entity.
type LimitedCompany struct {
	LimitedCompanyID   string     `json:"limited_company_id"`
	CompanyName        string     `json:"company_name"`
	RegistrationNumber string     `json:"registration_number"`
	Jurisdiction       string     `json:"jurisdiction"`
	IncorporationDate  *time.Time `json:"incorporation_date"`
	RegisteredAddress  string     `json:"registered_address"`
	BusinessNature     string     `json:"business_nature"`
}

// Partnership represents a partnership entity.
type Partnership struct {
	PartnershipID            string     `json:"partnership_id"`
	PartnershipName          string     `json:"partnership_name"`
	PartnershipType          string     `json:"partnership_type"`
	Jurisdiction             string     `json:"jurisdiction"`
	FormationDate            *time.Time `json:"formation_date"`
	PrincipalPlaceBusiness   string     `json:"principal_place_business"`
	PartnershipAgreementDate *time.Time `json:"partnership_agreement_date"`
}

// Individual represents an individual (proper person) entity.
type Individual struct {
	IndividualID     string     `json:"individual_id"`
	FirstName        string     `json:"first_name"`
	LastName         string     `json:"last_name"`
	MiddleNames      string     `json:"middle_names"`
	DateOfBirth      *time.Time `json:"date_of_birth"`
	Nationality      string     `json:"nationality"`
	ResidenceAddress string     `json:"residence_address"`
	IDDocumentType   string     `json:"id_document_type"`
	IDDocumentNumber string     `json:"id_document_number"`
}

// NewStore creates a new Store instance and opens a database connection.
func NewStore(connString string) (*Store, error) {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if pingErr := db.Ping(); pingErr != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", pingErr)
	}

	return &Store{db: db}, nil
}

// NewStoreFromDB constructs a Store from an existing *sql.DB. Useful for tests.
func NewStoreFromDB(db *sql.DB) *Store {
	return &Store{db: db}
}

// Close closes the database connection.
func (s *Store) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// InitDB initializes the database schema from the SQL file.
func (s *Store) InitDB(ctx context.Context) error {
	// Read the init.sql file
	sqlFilePath := filepath.Join("sql", "init.sql")
	sqlBytes, err := os.ReadFile(sqlFilePath)
	if err != nil {
		return fmt.Errorf("failed to read SQL file: %w", err)
	}

	// Execute the SQL
	_, err = s.db.ExecContext(ctx, string(sqlBytes))
	if err != nil {
		return fmt.Errorf("failed to execute init SQL: %w", err)
	}

	return nil
}

// SeedCatalog seeds the catalog tables with mock data.
func (s *Store) SeedCatalog(ctx context.Context) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// Insert CBUs
	cbus := []struct {
		name          string
		description   string
		naturePurpose string
	}{
		{"CBU-1234", "Aviva Investors Global Fund", "UCITS equity fund domiciled in LU"},
		{"CBU-5678", "Blackrock US Debt Fund", "Corporate debt fund domiciled in IE"},
		{"CBU-9999", "Test Development Fund", "Mock fund for testing and development"},
	}

	cbuIDs := make(map[string]string)
	for _, c := range cbus {
		var cbuID string
		queryErr := tx.QueryRowContext(ctx,
			`INSERT INTO "dsl-ob-poc".cbus (name, description, nature_purpose)
			 VALUES ($1, $2, $3)
			 ON CONFLICT (name) DO UPDATE SET description = EXCLUDED.description, nature_purpose = EXCLUDED.nature_purpose
			 RETURNING cbu_id`,
			c.name, c.description, c.naturePurpose).Scan(&cbuID)
		if queryErr != nil {
			return fmt.Errorf("failed to insert CBU %s: %w", c.name, queryErr)
		}
		cbuIDs[c.name] = cbuID
	}

	// Insert Products
	products := []struct {
		name        string
		description string
	}{
		{"CUSTODY", "Custody and safekeeping services"},
		{"FUND_ACCOUNTING", "Fund accounting and NAV calculation"},
		{"TRANSFER_AGENCY", "Transfer agency and registry services"},
	}

	productIDs := make(map[string]string)
	for _, p := range products {
		var productID string
		queryErr := tx.QueryRowContext(ctx,
			`INSERT INTO "dsl-ob-poc".products (name, description)
			 VALUES ($1, $2)
			 ON CONFLICT (name) DO UPDATE SET description = EXCLUDED.description
			 RETURNING product_id`,
			p.name, p.description).Scan(&productID)
		if queryErr != nil {
			return fmt.Errorf("failed to insert product %s: %w", p.name, queryErr)
		}
		productIDs[p.name] = productID
	}

	// Insert Services
	services := []struct {
		name        string
		description string
	}{
		{"CustodyService", "Asset custody and safekeeping"},
		{"SettlementService", "Trade settlement processing"},
		{"FundAccountingService", "Daily NAV calculation and reporting"},
		{"TransferAgencyService", "Shareholder registry management"},
	}

	serviceIDs := make(map[string]string)
	for _, srv := range services {
		var serviceID string
		queryErr := tx.QueryRowContext(ctx,
			`INSERT INTO "dsl-ob-poc".services (name, description)
			 VALUES ($1, $2)
			 ON CONFLICT (name) DO UPDATE SET description = EXCLUDED.description
			 RETURNING service_id`,
			srv.name, srv.description).Scan(&serviceID)
		if queryErr != nil {
			return fmt.Errorf("failed to insert service %s: %w", srv.name, queryErr)
		}
		serviceIDs[srv.name] = serviceID
	}

	// Link Products to Services
	productServiceLinks := []struct {
		product string
		service string
	}{
		{"CUSTODY", "CustodyService"},
		{"CUSTODY", "SettlementService"},
		{"FUND_ACCOUNTING", "FundAccountingService"},
		{"TRANSFER_AGENCY", "TransferAgencyService"},
	}

	for _, link := range productServiceLinks {
		_, execErr := tx.ExecContext(ctx,
			`INSERT INTO "dsl-ob-poc".product_services (product_id, service_id)
			 VALUES ($1, $2)
			 ON CONFLICT DO NOTHING`,
			productIDs[link.product], serviceIDs[link.service])
		if execErr != nil {
			return fmt.Errorf("failed to link product %s to service %s: %w", link.product, link.service, execErr)
		}
	}

	// Insert Dictionary Attributes (v3 schema)
	attributes := []struct {
		name            string
		longDescription string
		groupID         string
		mask            string
		domain          string
		sourceJSON      string
		sinkJSON        string
	}{
		{
			"onboard.cbu_id",
			"Client Business Unit identifier for onboarding case tracking and workflow management",
			"Onboarding",
			"string",
			"Onboarding",
			`{"type": "manual", "url": "https://onboarding.example.com/cbu", "required": true, "format": "CBU-[0-9]+"}`,
			`{"type": "database", "url": "postgres://onboarding_db/cases", "table": "onboarding_cases", "field": "cbu_id"}`,
		},
		{
			"entity.legal_name",
			"Legal name of the entity for KYC purposes",
			"KYC",
			"string",
			"KYC",
			`{"type": "manual", "url": "https://kyc.example.com/entity", "required": true}`,
			`{"type": "database", "url": "postgres://kyc_db/entities", "table": "legal_entities", "field": "legal_name"}`,
		},
		{
			"custody.account_number",
			"Custody account identifier for asset safekeeping",
			"CustodyAccount",
			"string",
			"Custody",
			`{"type": "api", "url": "https://custody.example.com/accounts", "method": "GET"}`,
			`{"type": "database", "url": "postgres://custody_db/accounts", "table": "accounts", "field": "account_number"}`,
		},
		{
			"entity.domicile",
			"Domicile jurisdiction of the fund or entity",
			"KYC",
			"string",
			"KYC",
			`{"type": "registry", "url": "https://registry.example.com/jurisdictions", "validated": true}`,
			`{"type": "database", "url": "postgres://kyc_db/entities", "table": "entities", "field": "domicile"}`,
		},
		{
			"security.isin",
			"International Securities Identification Number",
			"Security",
			"string",
			"Trading",
			`{"type": "api", "url": "https://isin-registry.example.com/lookup", "authoritative": true}`,
			`{"type": "database", "url": "postgres://trading_db/securities", "table": "securities", "field": "isin"}`,
		},
		{
			"accounting.nav_value",
			"Net Asset Value calculated daily",
			"FundAccounting",
			"string",
			"Accounting",
			`{"type": "calculated", "formula": "total_assets - total_liabilities", "frequency": "daily"}`,
			`{"type": "database", "url": "postgres://accounting_db/nav", "table": "daily_nav", "field": "nav_value"}`,
		},
	}

	for _, attr := range attributes {
		_, execErr := tx.ExecContext(ctx,
			`INSERT INTO "dsl-ob-poc".dictionary (name, long_description, group_id, mask, domain, source, sink)
			 VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7::jsonb)
			 ON CONFLICT (name) DO UPDATE SET
				long_description = EXCLUDED.long_description,
				group_id = EXCLUDED.group_id,
				mask = EXCLUDED.mask,
				domain = EXCLUDED.domain,
				source = EXCLUDED.source,
				sink = EXCLUDED.sink`,
			attr.name, attr.longDescription, attr.groupID, attr.mask, attr.domain, attr.sourceJSON, attr.sinkJSON)
		if execErr != nil {
			return fmt.Errorf("failed to insert dictionary attribute %s: %w", attr.name, execErr)
		}
	}

	// Insert Resources
	resources := []struct {
		name            string
		description     string
		owner           string
		dictionaryGroup string
	}{
		{"CustodyAccount", "Custody account resource", "CustodyTech", "CustodyAccount"},
		{"FundAccountingRecord", "Fund accounting record resource", "AccountingEng", "FundAccounting"},
		{"ShareholderRegistry", "Shareholder registry resource", "TransferAgencyTeam", "KYC"},
	}

	resourceIDs := make(map[string]string)
	for _, res := range resources {
		var resourceID string
		queryErr := tx.QueryRowContext(ctx,
			`INSERT INTO "dsl-ob-poc".prod_resources (name, description, owner, dictionary_group)
			 VALUES ($1, $2, $3, $4)
			 ON CONFLICT (name) DO UPDATE SET description = EXCLUDED.description
			 RETURNING resource_id`,
			res.name, res.description, res.owner, res.dictionaryGroup).Scan(&resourceID)
		if queryErr != nil {
			return fmt.Errorf("failed to insert resource %s: %w", res.name, queryErr)
		}
		resourceIDs[res.name] = resourceID
	}

	// Link Services to Resources
	serviceResourceLinks := []struct {
		service  string
		resource string
	}{
		{"CustodyService", "CustodyAccount"},
		{"SettlementService", "CustodyAccount"},
		{"FundAccountingService", "FundAccountingRecord"},
		{"TransferAgencyService", "ShareholderRegistry"},
	}

	for _, link := range serviceResourceLinks {
		_, execErr := tx.ExecContext(ctx,
			`INSERT INTO "dsl-ob-poc".service_resources (service_id, resource_id)
			 VALUES ($1, $2)
			 ON CONFLICT DO NOTHING`,
			serviceIDs[link.service], resourceIDs[link.resource])
		if execErr != nil {
			return fmt.Errorf("failed to link service %s to resource %s: %w", link.service, link.resource, execErr)
		}
	}

	return tx.Commit()
}

// InsertDSL inserts a new DSL version and returns its version ID.
func (s *Store) InsertDSL(ctx context.Context, cbuID, dslText string) (string, error) {
	var versionID string
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO "dsl-ob-poc".dsl_ob (cbu_id, dsl_text) VALUES ($1, $2) RETURNING version_id`,
		cbuID, dslText).Scan(&versionID)
	if err != nil {
		return "", fmt.Errorf("failed to insert DSL: %w", err)
	}
	return versionID, nil
}

// GetLatestDSL retrieves the most recent DSL for a given CBU ID.
func (s *Store) GetLatestDSL(ctx context.Context, cbuID string) (string, error) {
	var dslText string
	err := s.db.QueryRowContext(ctx,
		`SELECT dsl_text FROM "dsl-ob-poc".dsl_ob
		 WHERE cbu_id = $1
		 ORDER BY created_at DESC
		 LIMIT 1`,
		cbuID).Scan(&dslText)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("no DSL found for CBU_ID: %s", cbuID)
	}
	if err != nil {
		return "", fmt.Errorf("failed to get latest DSL: %w", err)
	}
	return dslText, nil
}

// DSLVersion represents a single versioned DSL entry.
type DSLVersion struct {
	VersionID string
	CreatedAt time.Time
	DSLText   string
}

// GetDSLHistory returns all DSL versions for a given CBU ID ordered by creation time.
func (s *Store) GetDSLHistory(ctx context.Context, cbuID string) ([]DSLVersion, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT version_id::text, created_at, dsl_text
         FROM "dsl-ob-poc".dsl_ob
         WHERE cbu_id = $1
         ORDER BY created_at ASC`, cbuID)
	if err != nil {
		return nil, fmt.Errorf("failed to query DSL history: %w", err)
	}
	defer rows.Close()

	var history []DSLVersion
	for rows.Next() {
		var v DSLVersion
		if scanErr := rows.Scan(&v.VersionID, &v.CreatedAt, &v.DSLText); scanErr != nil {
			return nil, fmt.Errorf("failed to scan DSL history row: %w", scanErr)
		}
		history = append(history, v)
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, fmt.Errorf("error iterating DSL history: %w", rowsErr)
	}

	return history, nil
}

// GetDictionaryAttributeByName retrieves an attribute from the dictionary by name
func (s *Store) GetDictionaryAttributeByName(ctx context.Context, name string) (*dictionary.Attribute, error) {
	var attr dictionary.Attribute
	var sourceJSON, sinkJSON string

	err := s.db.QueryRowContext(ctx,
		`SELECT attribute_id, name, long_description, group_id, mask, domain,
		        COALESCE(vector, ''), COALESCE(source::text, '{}'), COALESCE(sink::text, '{}')
		 FROM "dsl-ob-poc".dictionary WHERE name = $1`,
		name).Scan(&attr.AttributeID, &attr.Name, &attr.LongDescription, &attr.GroupID,
		&attr.Mask, &attr.Domain, &attr.Vector, &sourceJSON, &sinkJSON)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("attribute '%s' not found in dictionary", name)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get attribute: %w", err)
	}

	// Parse JSON metadata
	if err := json.Unmarshal([]byte(sourceJSON), &attr.Source); err != nil {
		return nil, fmt.Errorf("failed to parse source metadata: %w", err)
	}
	if err := json.Unmarshal([]byte(sinkJSON), &attr.Sink); err != nil {
		return nil, fmt.Errorf("failed to parse sink metadata: %w", err)
	}

	return &attr, nil
}

// GetDictionaryAttributeByID retrieves an attribute from the dictionary by UUID
func (s *Store) GetDictionaryAttributeByID(ctx context.Context, id string) (*dictionary.Attribute, error) {
	var attr dictionary.Attribute
	var sourceJSON, sinkJSON string

	err := s.db.QueryRowContext(ctx,
		`SELECT attribute_id, name, long_description, group_id, mask, domain,
		        COALESCE(vector, ''), COALESCE(source::text, '{}'), COALESCE(sink::text, '{}')
		 FROM "dsl-ob-poc".dictionary WHERE attribute_id = $1`,
		id).Scan(&attr.AttributeID, &attr.Name, &attr.LongDescription, &attr.GroupID,
		&attr.Mask, &attr.Domain, &attr.Vector, &sourceJSON, &sinkJSON)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("attribute with ID '%s' not found in dictionary", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get attribute: %w", err)
	}

	// Parse JSON metadata
	if err := json.Unmarshal([]byte(sourceJSON), &attr.Source); err != nil {
		return nil, fmt.Errorf("failed to parse source metadata: %w", err)
	}
	if err := json.Unmarshal([]byte(sinkJSON), &attr.Sink); err != nil {
		return nil, fmt.Errorf("failed to parse sink metadata: %w", err)
	}

	return &attr, nil
}

// GetCBUByName retrieves a CBU by name from the catalog
func (s *Store) GetCBUByName(ctx context.Context, name string) (*CBU, error) {
	var cbu CBU
	err := s.db.QueryRowContext(ctx,
		`SELECT cbu_id, name, description, nature_purpose FROM "dsl-ob-poc".cbus WHERE name = $1`,
		name).Scan(&cbu.CBUID, &cbu.Name, &cbu.Description, &cbu.NaturePurpose)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("CBU '%s' not found in catalog", name)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get CBU: %w", err)
	}
	return &cbu, nil
}

// ResolveValueFor resolves attribute values using source metadata
func (s *Store) ResolveValueFor(ctx context.Context, cbuID, attributeID string) (json.RawMessage, map[string]any, string, error) {
	a, err := s.GetDictionaryAttributeByID(ctx, attributeID)
	if err != nil {
		return nil, nil, "", err
	}

	// Super simple: if source indicates "cbus" table, fetch by cbuID
	sourceMap := make(map[string]interface{})
	sourceJSON, _ := json.Marshal(a.Source)
	if err := json.Unmarshal(sourceJSON, &sourceMap); err != nil {
		return nil, nil, "", fmt.Errorf("failed to parse source metadata: %w", err)
	}

	if table, ok := sourceMap["table"].(string); ok && table == "cbus" {
		if field, ok := sourceMap["field"].(string); ok && field != "" {
			query := fmt.Sprintf(`SELECT %s FROM "dsl-ob-poc".cbus WHERE cbu_id=$1`, field)
			var val interface{}
			err := s.db.QueryRowContext(ctx, query, cbuID).Scan(&val)
			if err != nil {
				return nil, nil, "", err
			}
			payload, _ := json.Marshal(val)
			prov := map[string]any{"table": "cbus", "field": field}
			return payload, prov, "resolved", nil
		}
	}

	// Unknown source → pending solicit
	return json.RawMessage(`null`), map[string]any{"reason": "no_resolver"}, "pending", nil
}

// UpsertAttributeValue stores or updates an attribute value
func (s *Store) UpsertAttributeValue(ctx context.Context, cbuID string, dslVersion int, attributeID string, value json.RawMessage, state string, source map[string]any) error {
	srcJSON, _ := json.Marshal(source)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO "dsl-ob-poc".attribute_values (cbu_id, dsl_version, attribute_id, value, state, source)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (cbu_id, dsl_version, attribute_id)
		DO UPDATE SET value = EXCLUDED.value, state = EXCLUDED.state, source = EXCLUDED.source, observed_at = (now() at time zone 'utc')`,
		cbuID, dslVersion, attributeID, value, state, string(srcJSON))
	return err
}

// StoreAttributeValue is a simple wrapper for UpsertAttributeValue
func (s *Store) StoreAttributeValue(ctx context.Context, onboardingID, attributeID, value string, sourceInfo map[string]interface{}) error {
	valueJSON, _ := json.Marshal(value)
	// For POC, use dsl_version = 1
	return s.UpsertAttributeValue(ctx, onboardingID, 1, attributeID, valueJSON, "resolved", sourceInfo)
}

// GetProductByName retrieves a product by name from the catalog.
func (s *Store) GetProductByName(ctx context.Context, name string) (*Product, error) {
	var p Product
	err := s.db.QueryRowContext(ctx,
		`SELECT product_id, name, description FROM "dsl-ob-poc".products WHERE name = $1`,
		name).Scan(&p.ProductID, &p.Name, &p.Description)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("product '%s' not found in catalog", name)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	return &p, nil
}

// GetServicesForProduct retrieves all services associated with a product.
func (s *Store) GetServicesForProduct(ctx context.Context, productID string) ([]Service, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT s.service_id, s.name, s.description
         FROM "dsl-ob-poc".services s
         JOIN "dsl-ob-poc".product_services ps ON s.service_id = ps.service_id
		 WHERE ps.product_id = $1`,
		productID)
	if err != nil {
		return nil, fmt.Errorf("failed to query services: %w", err)
	}
	defer rows.Close()

	var services []Service
	for rows.Next() {
		var srv Service
		if scanErr := rows.Scan(&srv.ServiceID, &srv.Name, &srv.Description); scanErr != nil {
			return nil, fmt.Errorf("failed to scan service: %w", scanErr)
		}
		services = append(services, srv)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, fmt.Errorf("error iterating services: %w", rowsErr)
	}

	return services, nil
}

// GetServiceByName retrieves a service by name from the catalog.
func (s *Store) GetServiceByName(ctx context.Context, name string) (*Service, error) {
	var srv Service
	err := s.db.QueryRowContext(ctx,
		`SELECT service_id, name, description FROM "dsl-ob-poc".services WHERE name = $1`,
		name).Scan(&srv.ServiceID, &srv.Name, &srv.Description)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("service '%s' not found in catalog", name)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get service: %w", err)
	}
	return &srv, nil
}

// GetResourcesForService retrieves all resources associated with a service.
func (s *Store) GetResourcesForService(ctx context.Context, serviceID string) ([]ProdResource, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT r.resource_id, r.name, r.description, r.owner, COALESCE(r.dictionary_id::text, '')
         FROM "dsl-ob-poc".prod_resources r
         JOIN "dsl-ob-poc".service_resources sr ON r.resource_id = sr.resource_id
		 WHERE sr.service_id = $1`,
		serviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to query resources: %w", err)
	}
	defer rows.Close()

	var resources []ProdResource
	for rows.Next() {
		var res ProdResource
		if scanErr := rows.Scan(&res.ResourceID, &res.Name, &res.Description, &res.Owner, &res.DictionaryGroup); scanErr != nil {
			return nil, fmt.Errorf("failed to scan resource: %w", scanErr)
		}
		resources = append(resources, res)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, fmt.Errorf("error iterating resources: %w", rowsErr)
	}

	return resources, nil
}

// GetAttributesForDictionaryGroup retrieves all attributes for a given dictionary group.
func (s *Store) GetAttributesForDictionaryGroup(ctx context.Context, groupID string) ([]Attribute, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT attribute_id, name, COALESCE(long_description, ''), group_id,
                COALESCE(mask, 'string'), COALESCE(domain, ''), COALESCE(vector, ''),
                COALESCE(source::text, '{}'), COALESCE(sink::text, '{}')
         FROM "dsl-ob-poc".dictionary
         WHERE group_id = $1`,
		groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to query dictionary attributes: %w", err)
	}
	defer rows.Close()

	var attributes []Attribute
	for rows.Next() {
		var attr Attribute
		if scanErr := rows.Scan(&attr.AttributeID, &attr.Name, &attr.LongDescription,
			&attr.GroupID, &attr.Mask, &attr.Domain, &attr.Vector,
			&attr.Source, &attr.Sink); scanErr != nil {
			return nil, fmt.Errorf("failed to scan attribute: %w", scanErr)
		}
		attributes = append(attributes, attr)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, fmt.Errorf("error iterating attributes: %w", rowsErr)
	}

	return attributes, nil
}

// ============================================================================
// CBU CRUD OPERATIONS
// ============================================================================

// CreateCBU creates a new CBU
func (s *Store) CreateCBU(ctx context.Context, name, description, naturePurpose string) (string, error) {
	query := `INSERT INTO "dsl-ob-poc".cbus (name, description, nature_purpose)
	         VALUES ($1, $2, $3) RETURNING cbu_id`

	var cbuID string
	err := s.db.QueryRowContext(ctx, query, name, description, naturePurpose).Scan(&cbuID)
	if err != nil {
		return "", fmt.Errorf("failed to create CBU: %w", err)
	}

	return cbuID, nil
}

// ListCBUs retrieves all CBUs
func (s *Store) ListCBUs(ctx context.Context) ([]CBU, error) {
	query := `SELECT cbu_id, name, description, nature_purpose
	         FROM "dsl-ob-poc".cbus
	         ORDER BY name`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list CBUs: %w", err)
	}
	defer rows.Close()

	var cbus []CBU
	for rows.Next() {
		var cbu CBU
		if scanErr := rows.Scan(&cbu.CBUID, &cbu.Name, &cbu.Description, &cbu.NaturePurpose); scanErr != nil {
			return nil, fmt.Errorf("failed to scan CBU: %w", scanErr)
		}
		cbus = append(cbus, cbu)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, fmt.Errorf("error iterating CBUs: %w", rowsErr)
	}

	return cbus, nil
}

// GetCBUByID retrieves a CBU by ID
func (s *Store) GetCBUByID(ctx context.Context, cbuID string) (*CBU, error) {
	query := `SELECT cbu_id, name, description, nature_purpose
	         FROM "dsl-ob-poc".cbus
	         WHERE cbu_id = $1`

	var cbu CBU
	err := s.db.QueryRowContext(ctx, query, cbuID).Scan(
		&cbu.CBUID, &cbu.Name, &cbu.Description, &cbu.NaturePurpose)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("CBU not found: %s", cbuID)
		}
		return nil, fmt.Errorf("failed to get CBU: %w", err)
	}

	return &cbu, nil
}

// UpdateCBU updates a CBU
func (s *Store) UpdateCBU(ctx context.Context, cbuID, name, description, naturePurpose string) error {
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if name != "" {
		setParts = append(setParts, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, name)
		argIndex++
	}
	if description != "" {
		setParts = append(setParts, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, description)
		argIndex++
	}
	if naturePurpose != "" {
		setParts = append(setParts, fmt.Sprintf("nature_purpose = $%d", argIndex))
		args = append(args, naturePurpose)
		argIndex++
	}

	if len(setParts) == 0 {
		return fmt.Errorf("no fields to update")
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	args = append(args, cbuID)

	query := fmt.Sprintf(`UPDATE "dsl-ob-poc".cbus SET %s WHERE cbu_id = $%d`,
		strings.Join(setParts, ", "), argIndex)

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update CBU: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("CBU not found: %s", cbuID)
	}

	return nil
}

// DeleteCBU deletes a CBU
func (s *Store) DeleteCBU(ctx context.Context, cbuID string) error {
	query := `DELETE FROM "dsl-ob-poc".cbus WHERE cbu_id = $1`

	result, err := s.db.ExecContext(ctx, query, cbuID)
	if err != nil {
		return fmt.Errorf("failed to delete CBU: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("CBU not found: %s", cbuID)
	}

	return nil
}

// ============================================================================
// ROLE CRUD OPERATIONS
// ============================================================================

// CreateRole creates a new role
func (s *Store) CreateRole(ctx context.Context, name, description string) (string, error) {
	query := `INSERT INTO "dsl-ob-poc".roles (name, description)
	         VALUES ($1, $2) RETURNING role_id`

	var roleID string
	err := s.db.QueryRowContext(ctx, query, name, description).Scan(&roleID)
	if err != nil {
		return "", fmt.Errorf("failed to create role: %w", err)
	}

	return roleID, nil
}

// ListRoles retrieves all roles
func (s *Store) ListRoles(ctx context.Context) ([]Role, error) {
	query := `SELECT role_id, name, description
	         FROM "dsl-ob-poc".roles
	         ORDER BY name`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var role Role
		if scanErr := rows.Scan(&role.RoleID, &role.Name, &role.Description); scanErr != nil {
			return nil, fmt.Errorf("failed to scan role: %w", scanErr)
		}
		roles = append(roles, role)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, fmt.Errorf("error iterating roles: %w", rowsErr)
	}

	return roles, nil
}

// GetRoleByID retrieves a role by ID
func (s *Store) GetRoleByID(ctx context.Context, roleID string) (*Role, error) {
	query := `SELECT role_id, name, description
	         FROM "dsl-ob-poc".roles
	         WHERE role_id = $1`

	var role Role
	err := s.db.QueryRowContext(ctx, query, roleID).Scan(
		&role.RoleID, &role.Name, &role.Description)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("role not found: %s", roleID)
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	return &role, nil
}

// UpdateRole updates a role
func (s *Store) UpdateRole(ctx context.Context, roleID, name, description string) error {
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if name != "" {
		setParts = append(setParts, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, name)
		argIndex++
	}
	if description != "" {
		setParts = append(setParts, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, description)
		argIndex++
	}

	if len(setParts) == 0 {
		return fmt.Errorf("no fields to update")
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	args = append(args, roleID)

	query := fmt.Sprintf(`UPDATE "dsl-ob-poc".roles SET %s WHERE role_id = $%d`,
		strings.Join(setParts, ", "), argIndex)

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("role not found: %s", roleID)
	}

	return nil
}

// DeleteRole deletes a role
func (s *Store) DeleteRole(ctx context.Context, roleID string) error {
	query := `DELETE FROM "dsl-ob-poc".roles WHERE role_id = $1`

	result, err := s.db.ExecContext(ctx, query, roleID)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("role not found: %s", roleID)
	}

	return nil
}
