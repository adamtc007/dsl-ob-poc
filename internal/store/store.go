package store

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/lib/pq"
)

// Store represents the database connection and operations.
type Store struct {
	db *sql.DB
}

// Product represents a product in the catalog.
type Product struct {
	ProductID   string
	Name        string
	Description string
}

// Service represents a service in the catalog.
type Service struct {
	ServiceID   string
	Name        string
	Description string
}

// ProdResource represents a resource required by products/services.
type ProdResource struct {
	ResourceID   string
	Name         string
	Description  string
	Owner        string
	DictionaryID string
}

// Attribute represents an attribute in a dictionary.
type Attribute struct {
	AttributeID         string
	Name                string
	DetailedDescription string
	IsPrivate           bool
	PrivateType         *string
	DataType            string
	PrimarySinkURL      string
	PrimarySourceURL    *string
	SecondarySourceURL  *string
	TertiarySourceURL   *string
}

// Dictionary represents a data dictionary.
type Dictionary struct {
	DictionaryID string
	Name         string
	Description  string
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
			`INSERT INTO "kyc-dsl".products (name, description)
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
			`INSERT INTO "kyc-dsl".services (name, description)
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
			`INSERT INTO "kyc-dsl".product_services (product_id, service_id)
			 VALUES ($1, $2)
			 ON CONFLICT DO NOTHING`,
			productIDs[link.product], serviceIDs[link.service])
		if execErr != nil {
			return fmt.Errorf("failed to link product %s to service %s: %w", link.product, link.service, execErr)
		}
	}

	// Insert Attributes
	attributes := []struct {
		name             string
		description      string
		isPrivate        bool
		privateType      *string
		dataType         string
		primarySinkURL   string
		primarySourceURL *string
	}{
		{"account_number", "Custody account identifier", false, nil, "string", "https://custody.example.com/accounts", nil},
		{"domicile", "Fund domicile jurisdiction", false, nil, "string", "https://registry.example.com/domicile", nil},
		{"isin", "International Securities Identification Number", false, nil, "string", "https://registry.example.com/isin", nil},
		{"nav_value", "Net Asset Value", true, strPtr("derived"), "string", "https://accounting.example.com/nav", nil},
	}

	attributeIDs := make(map[string]string)
	for _, attr := range attributes {
		var attributeID string
		queryErr := tx.QueryRowContext(ctx,
			`INSERT INTO "kyc-dsl".attributes (name, detailed_description, is_private, private_type, data_type, primary_sink_url, primary_source_url)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)
			 ON CONFLICT (name) DO UPDATE SET detailed_description = EXCLUDED.detailed_description
			 RETURNING attribute_id`,
			attr.name, attr.description, attr.isPrivate, attr.privateType, attr.dataType, attr.primarySinkURL, attr.primarySourceURL).Scan(&attributeID)
		if queryErr != nil {
			return fmt.Errorf("failed to insert attribute %s: %w", attr.name, queryErr)
		}
		attributeIDs[attr.name] = attributeID
	}

	// Insert Dictionaries
	dictionaries := []struct {
		name        string
		description string
	}{
		{"Custody Account Schema", "Schema for custody account resources"},
		{"Fund Accounting Schema", "Schema for fund accounting resources"},
		{"Transfer Agency Schema", "Schema for transfer agency resources"},
	}

	dictionaryIDs := make(map[string]string)
	for _, dict := range dictionaries {
		var dictionaryID string
		queryErr := tx.QueryRowContext(ctx,
			`INSERT INTO "kyc-dsl".dictionaries (name, description)
			 VALUES ($1, $2)
			 ON CONFLICT (name) DO UPDATE SET description = EXCLUDED.description
			 RETURNING dictionary_id`,
			dict.name, dict.description).Scan(&dictionaryID)
		if queryErr != nil {
			return fmt.Errorf("failed to insert dictionary %s: %w", dict.name, queryErr)
		}
		dictionaryIDs[dict.name] = dictionaryID
	}

	// Link Dictionaries to Attributes
	dictionaryAttributeLinks := []struct {
		dictionary string
		attribute  string
		required   bool
	}{
		{"Custody Account Schema", "account_number", true},
		{"Custody Account Schema", "domicile", true},
		{"Fund Accounting Schema", "nav_value", true},
		{"Fund Accounting Schema", "isin", true},
		{"Transfer Agency Schema", "isin", true},
		{"Transfer Agency Schema", "domicile", false},
	}

	for _, link := range dictionaryAttributeLinks {
		_, execErr := tx.ExecContext(ctx,
			`INSERT INTO "kyc-dsl".dictionary_attributes (dictionary_id, attribute_id, is_required)
			 VALUES ($1, $2, $3)
			 ON CONFLICT DO NOTHING`,
			dictionaryIDs[link.dictionary], attributeIDs[link.attribute], link.required)
		if execErr != nil {
			return fmt.Errorf("failed to link dictionary %s to attribute %s: %w", link.dictionary, link.attribute, execErr)
		}
	}

	// Insert Resources
	resources := []struct {
		name         string
		description  string
		owner        string
		dictionaryID string
	}{
		{"CustodyAccount", "Custody account resource", "CustodyTech", dictionaryIDs["Custody Account Schema"]},
		{"FundAccountingRecord", "Fund accounting record resource", "AccountingEng", dictionaryIDs["Fund Accounting Schema"]},
		{"ShareholderRegistry", "Shareholder registry resource", "TransferAgencyTeam", dictionaryIDs["Transfer Agency Schema"]},
	}

	resourceIDs := make(map[string]string)
	for _, res := range resources {
		var resourceID string
		queryErr := tx.QueryRowContext(ctx,
			`INSERT INTO "kyc-dsl".prod_resources (name, description, owner, dictionary_id)
			 VALUES ($1, $2, $3, $4)
			 ON CONFLICT (name) DO UPDATE SET description = EXCLUDED.description
			 RETURNING resource_id`,
			res.name, res.description, res.owner, res.dictionaryID).Scan(&resourceID)
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
			`INSERT INTO "kyc-dsl".service_resources (service_id, resource_id)
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
		`INSERT INTO "kyc-dsl".dsl_ob (cbu_id, dsl_text) VALUES ($1, $2) RETURNING version_id`,
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
		`SELECT dsl_text FROM "kyc-dsl".dsl_ob
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

// GetProductByName retrieves a product by name from the catalog.
func (s *Store) GetProductByName(ctx context.Context, name string) (*Product, error) {
	var p Product
	err := s.db.QueryRowContext(ctx,
		`SELECT product_id, name, description FROM "kyc-dsl".products WHERE name = $1`,
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
		 FROM "kyc-dsl".services s
		 JOIN "kyc-dsl".product_services ps ON s.service_id = ps.service_id
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
		`SELECT service_id, name, description FROM "kyc-dsl".services WHERE name = $1`,
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
		 FROM "kyc-dsl".prod_resources r
		 JOIN "kyc-dsl".service_resources sr ON r.resource_id = sr.resource_id
		 WHERE sr.service_id = $1`,
		serviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to query resources: %w", err)
	}
	defer rows.Close()

	var resources []ProdResource
	for rows.Next() {
		var res ProdResource
		if scanErr := rows.Scan(&res.ResourceID, &res.Name, &res.Description, &res.Owner, &res.DictionaryID); scanErr != nil {
			return nil, fmt.Errorf("failed to scan resource: %w", scanErr)
		}
		resources = append(resources, res)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, fmt.Errorf("error iterating resources: %w", rowsErr)
	}

	return resources, nil
}

// GetAttributesForDictionary retrieves all attributes for a given dictionary.
func (s *Store) GetAttributesForDictionary(ctx context.Context, dictionaryID string) ([]Attribute, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT a.attribute_id, a.name, COALESCE(a.detailed_description, ''),
		        a.is_private, a.private_type, a.data_type, a.primary_sink_url,
		        a.primary_source_url, a.secondary_source_url, a.tertiary_source_url
		 FROM "kyc-dsl".attributes a
		 JOIN "kyc-dsl".dictionary_attributes da ON a.attribute_id = da.attribute_id
		 WHERE da.dictionary_id = $1`,
		dictionaryID)
	if err != nil {
		return nil, fmt.Errorf("failed to query attributes: %w", err)
	}
	defer rows.Close()

	var attributes []Attribute
	for rows.Next() {
		var attr Attribute
		if scanErr := rows.Scan(&attr.AttributeID, &attr.Name, &attr.DetailedDescription,
			&attr.IsPrivate, &attr.PrivateType, &attr.DataType, &attr.PrimarySinkURL,
			&attr.PrimarySourceURL, &attr.SecondarySourceURL, &attr.TertiarySourceURL); scanErr != nil {
			return nil, fmt.Errorf("failed to scan attribute: %w", scanErr)
		}
		attributes = append(attributes, attr)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, fmt.Errorf("error iterating attributes: %w", rowsErr)
	}

	return attributes, nil
}

// Helper function to create string pointers
func strPtr(s string) *string {
	return &s
}
