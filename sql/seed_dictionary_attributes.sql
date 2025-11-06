/*
 * Dictionary Attributes Seed Data
 *
 * This file populates the "dsl-ob-poc".dictionary table with comprehensive
 * attribute definitions for:
 * - Onboarding lifecycle (opportunity → offboarding)
 * - KYC - Institutional (corporate entities, trusts, partnerships)
 * - KYC - Proper Person / Retail (individual investors)
 *
 * ARCHITECTURE PRINCIPLE:
 * These attributes support RAG (Retrieval-Augmented Generation) for AI agents.
 * Each attribute has rich, descriptive text optimized for:
 * - AI semantic understanding
 * - Vector database embeddings
 * - Natural language query matching
 * - Context-aware DSL generation
 *
 * The long_description field should be verbose and include:
 * - What the attribute represents
 * - When it's collected in the workflow
 * - How it's used downstream
 * - Regulatory/compliance context if applicable
 */

-- =============================================================================
-- ONBOARDING LIFECYCLE ATTRIBUTES
-- =============================================================================

-- Opportunity Stage
INSERT INTO "dsl-ob-poc".dictionary (attribute_id, name, long_description, group_id, mask, domain, source, sink) VALUES
(gen_random_uuid(), 'opportunity.cbu_id',
 'Client Business Unit identifier assigned at opportunity creation. This is the primary key that links all onboarding activities, documents, and state transitions. Used throughout the entire client lifecycle from initial opportunity through active servicing and eventual offboarding. Critical for audit trail and regulatory reporting.',
 'onboarding_core', 'STRING', 'ONBOARDING',
 '{"primary": "OPPORTUNITY_CREATION", "secondary": "CRM_SYSTEM"}',
 '{"primary": "CBU_REGISTRY", "secondary": "ALL_ONBOARDING_TABLES"}'),

(gen_random_uuid(), 'opportunity.nature_purpose',
 'Free-text description of the client''s business nature and purpose for engaging with the bank. Examples: "UCITS equity fund domiciled in Luxembourg", "US-based hedge fund investing in global equities", "Private wealth client seeking custody and advisory services". This field is parsed by AI agents to determine applicable products, services, and compliance requirements. Critical for product suitability assessment.',
 'onboarding_core', 'TEXT', 'ONBOARDING',
 '{"primary": "CLIENT_APPLICATION_FORM", "secondary": "RELATIONSHIP_MANAGER_INPUT"}',
 '{"primary": "CBU_PROFILE", "secondary": "PRODUCT_SUITABILITY_ASSESSMENT"}'),

(gen_random_uuid(), 'opportunity.source',
 'Origin of the client opportunity. Used for business analytics and relationship management. Influences initial risk assessment and determines appropriate onboarding workflow. Examples: REFERRAL, DIRECT_MARKETING, EXISTING_CLIENT_EXPANSION, STRATEGIC_PARTNERSHIP.',
 'onboarding_metadata', 'ENUM', 'ONBOARDING',
 '{"primary": "CRM_SYSTEM"}',
 '{"primary": "OPPORTUNITY_REGISTRY", "secondary": "BUSINESS_INTELLIGENCE"}'),

(gen_random_uuid(), 'opportunity.created_date',
 'Timestamp when the opportunity was first recorded in the system. Used for SLA tracking, aging reports, and regulatory timekeeping. Critical for measuring onboarding efficiency and identifying bottlenecks.',
 'onboarding_metadata', 'TIMESTAMP', 'ONBOARDING',
 '{"primary": "SYSTEM_GENERATED"}',
 '{"primary": "OPPORTUNITY_REGISTRY", "secondary": "REPORTING_WAREHOUSE"}'),

(gen_random_uuid(), 'opportunity.relationship_manager',
 'Name or identifier of the relationship manager responsible for this client. Used for workflow routing, escalations, and client communication. Critical for accountability and client experience.',
 'onboarding_metadata', 'STRING', 'ONBOARDING',
 '{"primary": "HR_SYSTEM", "secondary": "CRM_ASSIGNMENT"}',
 '{"primary": "CBU_PROFILE", "secondary": "WORKFLOW_ROUTING"}'),

-- Product Selection Stage
(gen_random_uuid(), 'products.selected',
 'Array of product codes selected by the client during onboarding. Products drive the entire onboarding workflow - determining required documents, services to provision, resources to create, and compliance checks to perform. Examples: CUSTODY, FUND_ACCOUNTING, TRANSFER_AGENT, TRADE_EXECUTION, ADVISORY. Each product has associated requirements defined in product_requirements table.',
 'product_selection', 'ARRAY', 'ONBOARDING',
 '{"primary": "PRODUCT_SELECTION_FORM", "secondary": "RELATIONSHIP_MANAGER_INPUT"}',
 '{"primary": "PRODUCT_REGISTRY", "secondary": "SERVICE_ORCHESTRATION"}'),

(gen_random_uuid(), 'products.suitability_confirmed',
 'Boolean flag indicating that product suitability assessment has been completed and documented. Required for regulatory compliance before product provisioning can begin. Typically involves checking client entity type, jurisdiction, regulatory status against product eligibility criteria.',
 'product_selection', 'BOOLEAN', 'ONBOARDING',
 '{"primary": "SUITABILITY_ASSESSMENT_WORKFLOW"}',
 '{"primary": "COMPLIANCE_AUDIT_LOG", "secondary": "PRODUCT_REGISTRY"}'),

-- Service Discovery Stage
(gen_random_uuid(), 'services.discovered',
 'Array of service names identified through catalog lookup based on selected products. Services are the operational capabilities that must be configured to deliver products. Examples: CUSTODY_SETTLEMENT, NAV_CALCULATION, SHAREHOLDER_REGISTRY, ORDER_ROUTING. Each service has associated resources and configuration requirements.',
 'service_discovery', 'ARRAY', 'ONBOARDING',
 '{"primary": "SERVICE_CATALOG_LOOKUP", "secondary": "PRODUCT_SERVICES_JOIN_TABLE"}',
 '{"primary": "SERVICE_REGISTRY", "secondary": "RESOURCE_ORCHESTRATION"}'),

(gen_random_uuid(), 'services.provisioning_status',
 'Current state of service provisioning workflow. Tracks progress through service configuration, testing, and activation. Values: PENDING, IN_PROGRESS, CONFIGURED, TESTED, ACTIVE, FAILED. Used for workflow orchestration and status reporting.',
 'service_discovery', 'ENUM', 'ONBOARDING',
 '{"primary": "SERVICE_PROVISIONING_WORKFLOW"}',
 '{"primary": "SERVICE_REGISTRY", "secondary": "WORKFLOW_STATE_MACHINE"}'),

-- Resource Planning Stage
(gen_random_uuid(), 'resources.planned',
 'Array of resource identifiers that must be created or allocated for this client. Resources are concrete infrastructure elements like custody accounts, fund accounting entities, shareholder registry entries. Each resource is defined by a dictionary group that specifies required attributes.',
 'resource_planning', 'ARRAY', 'ONBOARDING',
 '{"primary": "RESOURCE_CATALOG_LOOKUP", "secondary": "SERVICE_RESOURCES_JOIN_TABLE"}',
 '{"primary": "RESOURCE_REGISTRY", "secondary": "ATTRIBUTE_RESOLUTION"}'),

(gen_random_uuid(), 'resources.provisioning_status',
 'Current state of resource provisioning. Tracks creation, configuration, validation, and activation of each resource. Values: PENDING, CREATING, CONFIGURED, VALIDATING, ACTIVE, FAILED. Critical for ensuring all infrastructure is ready before client go-live.',
 'resource_planning', 'ENUM', 'ONBOARDING',
 '{"primary": "RESOURCE_PROVISIONING_WORKFLOW"}',
 '{"primary": "RESOURCE_REGISTRY", "secondary": "WORKFLOW_STATE_MACHINE"}'),

-- State Tracking
(gen_random_uuid(), 'onboarding.current_state',
 'High-level state of the onboarding workflow. Represents major phase transitions in the client lifecycle. Values: OPPORTUNITY, PRODUCTS_SELECTED, KYC_DISCOVERED, SERVICES_DISCOVERED, RESOURCES_DISCOVERED, READY_FOR_ACTIVATION, ACTIVE, SUSPENDED, OFFBOARDING, OFFBOARDED. Used for executive reporting and SLA monitoring.',
 'onboarding_state', 'ENUM', 'ONBOARDING',
 '{"primary": "ONBOARDING_STATE_MACHINE"}',
 '{"primary": "ONBOARDING_SESSIONS", "secondary": "REPORTING_WAREHOUSE"}'),

(gen_random_uuid(), 'onboarding.version',
 'DSL version number tracking state changes. Each command that modifies the onboarding DSL increments this version. Enables event sourcing, audit trails, and point-in-time state reconstruction. Links to dsl_ob.version_id for complete DSL history.',
 'onboarding_state', 'INTEGER', 'ONBOARDING',
 '{"primary": "DSL_ACCUMULATION_SYSTEM"}',
 '{"primary": "DSL_OB_TABLE", "secondary": "AUDIT_LOG"}'),

(gen_random_uuid(), 'onboarding.go_live_date',
 'Target or actual date when client begins active use of services. Used for project planning, resource scheduling, and client communications. Critical milestone tracked in SLA reports.',
 'onboarding_milestones', 'DATE', 'ONBOARDING',
 '{"primary": "PROJECT_PLAN", "secondary": "RELATIONSHIP_MANAGER_INPUT"}',
 '{"primary": "CBU_PROFILE", "secondary": "REPORTING_WAREHOUSE"}'),

(gen_random_uuid(), 'onboarding.completion_date',
 'Timestamp when all onboarding activities completed and client transitioned to active servicing. Used to calculate total onboarding duration for process improvement analysis.',
 'onboarding_milestones', 'TIMESTAMP', 'ONBOARDING',
 '{"primary": "ONBOARDING_WORKFLOW_COMPLETION"}',
 '{"primary": "CBU_PROFILE", "secondary": "BUSINESS_INTELLIGENCE"}');

-- =============================================================================
-- KYC - INSTITUTIONAL (Corporate Entities, Trusts, Partnerships)
-- =============================================================================

-- Entity Identification
INSERT INTO "dsl-ob-poc".dictionary (attribute_id, name, long_description, group_id, mask, domain, source, sink) VALUES
(gen_random_uuid(), 'kyc.institutional.legal_name',
 'Official registered legal name of the institutional entity exactly as it appears on incorporation or formation documents. Used for identity verification, sanctions screening, and legal contract generation. Must match name on Certificate of Incorporation, Trust Deed, or Partnership Agreement. Critical for regulatory reporting and audit trail.',
 'kyc_institutional_identity', 'STRING', 'KYC',
 '{"primary": "CERTIFICATE_OF_INCORPORATION", "secondary": "CORPORATE_REGISTRY_LOOKUP"}',
 '{"primary": "ENTITY_REGISTRY", "secondary": "SANCTIONS_SCREENING_SYSTEM"}'),

(gen_random_uuid(), 'kyc.institutional.entity_type',
 'Legal structure of the institutional client. Determines applicable KYC procedures, documentation requirements, and UBO identification methodology. Values: CORPORATION (stock corporation, AG, SA), LLC (limited liability company), PARTNERSHIP (general or limited), TRUST (irrevocable, revocable, discretionary), FOUNDATION, FUND (hedge fund, private equity, UCITS). Regulatory treatment varies significantly by entity type.',
 'kyc_institutional_classification', 'ENUM', 'KYC',
 '{"primary": "FORMATION_DOCUMENTS", "secondary": "LEGAL_OPINION"}',
 '{"primary": "ENTITY_REGISTRY", "secondary": "UBO_WORKFLOW_ROUTER"}'),

(gen_random_uuid(), 'kyc.institutional.jurisdiction',
 'Country of incorporation, formation, or establishment (ISO 3166-1 alpha-2 code). Determines applicable regulatory framework, tax treatment, and sanctions risk. High-risk jurisdictions (e.g., sanctioned countries, non-cooperative jurisdictions per FATF) trigger enhanced due diligence. Used in risk rating calculation.',
 'kyc_institutional_identity', 'STRING', 'KYC',
 '{"primary": "CERTIFICATE_OF_INCORPORATION", "secondary": "CORPORATE_REGISTRY"}',
 '{"primary": "ENTITY_REGISTRY", "secondary": "RISK_ASSESSMENT_ENGINE"}'),

(gen_random_uuid(), 'kyc.institutional.registration_number',
 'Official company registration number assigned by jurisdiction''s corporate registry. Used for corporate registry lookups, sanctions screening, and duplicate client detection. Format varies by jurisdiction (e.g., UK: 12345678, Delaware: 123456, BVI: 123456). Essential for identity verification.',
 'kyc_institutional_identity', 'STRING', 'KYC',
 '{"primary": "CERTIFICATE_OF_INCORPORATION", "secondary": "CORPORATE_REGISTRY_API"}',
 '{"primary": "ENTITY_REGISTRY", "secondary": "DUPLICATE_CHECK_SYSTEM"}'),

(gen_random_uuid(), 'kyc.institutional.date_of_incorporation',
 'Date the entity was legally formed or incorporated. Used to assess entity maturity (new entities may trigger enhanced due diligence), verify document authenticity, and calculate corporate anniversaries. Format: YYYY-MM-DD.',
 'kyc_institutional_identity', 'DATE', 'KYC',
 '{"primary": "CERTIFICATE_OF_INCORPORATION"}',
 '{"primary": "ENTITY_REGISTRY", "secondary": "RISK_ASSESSMENT"}'),

(gen_random_uuid(), 'kyc.institutional.registered_address',
 'Official registered office address as filed with corporate registry. Used for correspondence, legal notices, and address verification. If different from operating address, both must be documented. Virtual office or mail forwarding addresses may trigger enhanced due diligence.',
 'kyc_institutional_identity', 'TEXT', 'KYC',
 '{"primary": "CERTIFICATE_OF_INCORPORATION", "secondary": "PROOF_OF_ADDRESS"}',
 '{"primary": "ENTITY_REGISTRY", "secondary": "ADDRESS_VERIFICATION_SERVICE"}'),

(gen_random_uuid(), 'kyc.institutional.business_activities',
 'Detailed description of the entity''s business operations, investment strategy, or purpose. Used to assess product suitability, identify conflicts of interest, and detect suspicious activity patterns. For funds: investment mandate, asset classes, geographic focus. For operating companies: industry sector, products/services, revenue sources. Critical for ongoing transaction monitoring.',
 'kyc_institutional_business', 'TEXT', 'KYC',
 '{"primary": "CLIENT_QUESTIONNAIRE", "secondary": "OFFERING_MEMORANDUM", "tertiary": "WEBSITE_RESEARCH"}',
 '{"primary": "ENTITY_PROFILE", "secondary": "TRANSACTION_MONITORING_SYSTEM"}'),

(gen_random_uuid(), 'kyc.institutional.regulatory_status',
 'Regulatory oversight status of the entity. Indicates if entity is regulated, by which regulator, and registration numbers. Examples: "SEC-registered Investment Adviser (CRD #12345)", "UCITS fund authorized by Luxembourg CSSF", "Unregulated private investment vehicle". Regulated entities may receive expedited KYC treatment.',
 'kyc_institutional_regulatory', 'TEXT', 'KYC',
 '{"primary": "REGULATORY_FILINGS", "secondary": "REGULATOR_DATABASE_LOOKUP"}',
 '{"primary": "ENTITY_PROFILE", "secondary": "RISK_ASSESSMENT"}'),

-- Document Collection (Institutional)
(gen_random_uuid(), 'kyc.institutional.certificate_of_incorporation',
 'Required foundational document proving legal existence. Must be certified copy issued within last 6 months or apostilled. AI agents use this to extract: legal name, jurisdiction, registration number, incorporation date, entity type. Document verification includes authenticity checks against known forgery patterns.',
 'kyc_institutional_documents', 'DOCUMENT', 'KYC',
 '{"primary": "DOCUMENT_UPLOAD_PORTAL", "secondary": "RELATIONSHIP_MANAGER_EMAIL"}',
 '{"primary": "DOCUMENT_REPOSITORY", "secondary": "OCR_EXTRACTION_QUEUE"}'),

(gen_random_uuid(), 'kyc.institutional.articles_memorandum',
 'Articles of Association, Memorandum of Association, or equivalent constitutional documents. Define entity purpose, governance structure, authorized activities. Used to verify client is authorized to engage in proposed banking relationship. AI agents extract permitted activities, share capital structure, director powers.',
 'kyc_institutional_documents', 'DOCUMENT', 'KYC',
 '{"primary": "DOCUMENT_UPLOAD_PORTAL"}',
 '{"primary": "DOCUMENT_REPOSITORY", "secondary": "LEGAL_REVIEW_QUEUE"}'),

(gen_random_uuid(), 'kyc.institutional.shareholder_register',
 'Current list of shareholders/members and their ownership percentages. Critical for UBO identification. Must show direct ownership clearly. If shareholders are entities (not natural persons), triggers recursive UBO analysis. Updated annually or upon material changes. Used to build ownership graph.',
 'kyc_institutional_documents', 'DOCUMENT', 'KYC',
 '{"primary": "CORPORATE_SECRETARY", "secondary": "REGISTRY_AGENT"}',
 '{"primary": "DOCUMENT_REPOSITORY", "secondary": "UBO_CALCULATION_ENGINE"}'),

(gen_random_uuid(), 'kyc.institutional.beneficial_ownership_declaration',
 'Formal declaration identifying natural persons who are Ultimate Beneficial Owners (25%+ ownership or control). Required under 4MLD/5MLD (EU), CDD Rule (US), FATF Recommendations. Must be signed by authorized officer and updated annually. Discrepancies with shareholder register trigger investigation.',
 'kyc_institutional_documents', 'DOCUMENT', 'KYC',
 '{"primary": "DOCUMENT_UPLOAD_PORTAL"}',
 '{"primary": "DOCUMENT_REPOSITORY", "secondary": "UBO_REGISTRY"}'),

(gen_random_uuid(), 'kyc.institutional.proof_of_address',
 'Utility bill, bank statement, or official correspondence confirming registered address. Must be dated within last 3 months. Used to verify registered address is genuine and entity is reachable. Virtual office addresses require additional documentation.',
 'kyc_institutional_documents', 'DOCUMENT', 'KYC',
 '{"primary": "DOCUMENT_UPLOAD_PORTAL"}',
 '{"primary": "DOCUMENT_REPOSITORY", "secondary": "ADDRESS_VERIFICATION_SYSTEM"}'),

(gen_random_uuid(), 'kyc.institutional.financial_statements',
 'Audited or management financial statements for most recent fiscal year. Used for credit assessment, AML risk rating (cash-intensive businesses = higher risk), and detecting financial distress. May be waived for newly formed entities. AI agents extract: total assets, revenue, cash flow patterns, auditor name.',
 'kyc_institutional_documents', 'DOCUMENT', 'KYC',
 '{"primary": "DOCUMENT_UPLOAD_PORTAL", "secondary": "PUBLIC_FILINGS"}',
 '{"primary": "DOCUMENT_REPOSITORY", "secondary": "CREDIT_ASSESSMENT_SYSTEM"}'),

-- Risk Assessment (Institutional)
(gen_random_uuid(), 'kyc.institutional.risk_rating',
 'Overall KYC/AML risk rating for the institutional client. Calculated based on: jurisdiction risk, entity type, business activities, ownership transparency, regulatory status, product usage patterns. Values: LOW, MEDIUM, HIGH, VERY_HIGH. Determines review frequency and transaction monitoring intensity. Reviewed annually or upon trigger events.',
 'kyc_institutional_risk', 'ENUM', 'KYC',
 '{"primary": "RISK_ASSESSMENT_ENGINE", "secondary": "COMPLIANCE_OFFICER_OVERRIDE"}',
 '{"primary": "ENTITY_PROFILE", "secondary": "TRANSACTION_MONITORING_RULES"}'),

(gen_random_uuid(), 'kyc.institutional.pep_exposure',
 'Indicates if entity has connections to Politically Exposed Persons (PEPs). Includes: PEP as UBO, PEP as director, significant business with PEP-controlled entities. Values: NO_PEP_EXPOSURE, INDIRECT_PEP_EXPOSURE, DIRECT_PEP_INVOLVEMENT. PEP exposure triggers enhanced due diligence and senior management approval.',
 'kyc_institutional_risk', 'ENUM', 'KYC',
 '{"primary": "PEP_SCREENING_DATABASE", "secondary": "UBO_SCREENING_RESULTS"}',
 '{"primary": "ENTITY_PROFILE", "secondary": "ENHANCED_DUE_DILIGENCE_WORKFLOW"}'),

(gen_random_uuid(), 'kyc.institutional.sanctions_screening_result',
 'Result of screening entity name, registration number, address, and associated parties against sanctions lists (OFAC SDN, UN, EU, HMT). Values: CLEAR (no hits), POTENTIAL_MATCH_UNDER_REVIEW (requires analyst review), FALSE_POSITIVE_CLEARED, TRUE_POSITIVE_BLOCKED. Must be CLEAR before account opening.',
 'kyc_institutional_risk', 'ENUM', 'KYC',
 '{"primary": "SANCTIONS_SCREENING_SYSTEM"}',
 '{"primary": "COMPLIANCE_AUDIT_LOG", "secondary": "ACCOUNT_OPENING_GATE"}'),

(gen_random_uuid(), 'kyc.institutional.adverse_media_found',
 'Boolean indicating if negative news found during media screening. Includes: fraud allegations, money laundering investigations, sanctions violations, insolvency, regulatory enforcement actions. True value triggers enhanced due diligence and escalation to compliance committee.',
 'kyc_institutional_risk', 'BOOLEAN', 'KYC',
 '{"primary": "ADVERSE_MEDIA_SCREENING", "secondary": "MANUAL_RESEARCH"}',
 '{"primary": "ENTITY_PROFILE", "secondary": "ENHANCED_DUE_DILIGENCE_WORKFLOW"}'),

-- Trust-Specific Attributes
(gen_random_uuid(), 'kyc.trust.trust_type',
 'Classification of trust structure. Determines UBO identification approach per FATF guidance on transparency. Values: DISCRETIONARY (trustee has full discretion over distributions), FIXED_INTEREST (beneficiaries have defined entitlements), UNIT_TRUST (beneficiaries hold units), CHARITABLE (public benefit purpose), BARE_TRUST (beneficiary has immediate right to assets). Each type has different UBO identification rules.',
 'kyc_trust_specific', 'ENUM', 'KYC',
 '{"primary": "TRUST_DEED", "secondary": "TRUSTEE_CERTIFICATION"}',
 '{"primary": "ENTITY_REGISTRY", "secondary": "UBO_WORKFLOW_ROUTER"}'),

(gen_random_uuid(), 'kyc.trust.settlor_identity',
 'Natural person(s) who created the trust and contributed assets. Settlor is always considered a UBO for KYC purposes per FATF recommendations, regardless of whether they retain any interest. Deceased settlors must still be documented. Requires full KYC on each settlor.',
 'kyc_trust_specific', 'TEXT', 'KYC',
 '{"primary": "TRUST_DEED", "secondary": "TRUSTEE_DECLARATION"}',
 '{"primary": "UBO_REGISTRY", "secondary": "INDIVIDUAL_KYC_WORKFLOW"}'),

(gen_random_uuid(), 'kyc.trust.trustee_identity',
 'Entity or natural person(s) holding legal title to trust assets and managing trust affairs. Corporate trustees require full institutional KYC. Individual trustees require personal KYC. Trustees are UBOs if they have discretion over distributions (common in discretionary trusts).',
 'kyc_trust_specific', 'TEXT', 'KYC',
 '{"primary": "TRUST_DEED", "secondary": "TRUSTEE_REGISTER"}',
 '{"primary": "UBO_REGISTRY", "secondary": "KYC_WORKFLOW_ROUTER"}'),

(gen_random_uuid(), 'kyc.trust.beneficiary_class',
 'Description of beneficiaries if not individually named. Examples: "all grandchildren of John Smith", "charitable organizations supporting education in Africa". Class beneficiaries create UBO identification challenges - trustee with discretion over unnamed class members is typically considered UBO. May require ongoing monitoring as class membership changes.',
 'kyc_trust_specific', 'TEXT', 'KYC',
 '{"primary": "TRUST_DEED"}',
 '{"primary": "ENTITY_PROFILE", "secondary": "UBO_MONITORING_SYSTEM"}'),

(gen_random_uuid(), 'kyc.trust.named_beneficiaries',
 'Individually identified beneficiaries with current or future interests in trust assets. Each beneficiary with current right to distribution or 25%+ ultimate interest is a UBO requiring full KYC. Discretionary beneficiaries without vested rights may not be UBOs. Documented in trust deed or letter of wishes.',
 'kyc_trust_specific', 'TEXT', 'KYC',
 '{"primary": "TRUST_DEED", "secondary": "LETTER_OF_WISHES"}',
 '{"primary": "UBO_REGISTRY", "secondary": "INDIVIDUAL_KYC_WORKFLOW"}'),

(gen_random_uuid(), 'kyc.trust.protector_identity',
 'Person or entity with power to appoint/remove trustees, veto distributions, or amend trust terms. Protector with significant control powers is considered a UBO. Common in offshore trust structures. Requires KYC if control powers meet threshold.',
 'kyc_trust_specific', 'TEXT', 'KYC',
 '{"primary": "TRUST_DEED"}',
 '{"primary": "UBO_REGISTRY", "secondary": "KYC_WORKFLOW_ROUTER"}');

-- =============================================================================
-- KYC - PROPER PERSON / RETAIL (Individual Investors)
-- =============================================================================

-- Individual Identity
INSERT INTO "dsl-ob-poc".dictionary (attribute_id, name, long_description, group_id, mask, domain, source, sink) VALUES
(gen_random_uuid(), 'kyc.individual.full_legal_name',
 'Complete name of the individual exactly as it appears on primary identification document (passport or national ID). Format: First name(s), Middle name(s), Last name(s). Used for identity verification, sanctions screening, contract generation, and payment processing. Any discrepancies between documents must be explained and documented. Critical for regulatory reporting.',
 'kyc_individual_identity', 'STRING', 'KYC',
 '{"primary": "PASSPORT", "secondary": "NATIONAL_ID_CARD"}',
 '{"primary": "INDIVIDUAL_REGISTRY", "secondary": "SANCTIONS_SCREENING"}'),

(gen_random_uuid(), 'kyc.individual.date_of_birth',
 'Date of birth in YYYY-MM-DD format. Used for age verification (ensure client meets minimum age requirements), identity verification, PEP screening age correlation, and duplicate detection. Individuals under 18 require guardian/custodial arrangements. Senior individuals (75+) may require additional suitability assessment.',
 'kyc_individual_identity', 'DATE', 'KYC',
 '{"primary": "PASSPORT", "secondary": "NATIONAL_ID_CARD", "tertiary": "BIRTH_CERTIFICATE"}',
 '{"primary": "INDIVIDUAL_REGISTRY", "secondary": "AGE_VERIFICATION_SYSTEM"}'),

(gen_random_uuid(), 'kyc.individual.nationality',
 'Country of citizenship (ISO 3166-1 alpha-2 code). Multiple nationalities must all be documented. Used for tax reporting (FATCA, CRS), sanctions screening, travel document verification, and regulatory classification. US nationals trigger FATCA reporting obligations. Sanctioned country nationals require enhanced due diligence.',
 'kyc_individual_identity', 'STRING', 'KYC',
 '{"primary": "PASSPORT"}',
 '{"primary": "INDIVIDUAL_REGISTRY", "secondary": "TAX_REPORTING_SYSTEM"}'),

(gen_random_uuid(), 'kyc.individual.country_of_birth',
 'Country where individual was born (ISO 3166-1 alpha-2 code). May differ from nationality if naturalized. Used for enhanced PEP screening, sanctions risk assessment, and tax reporting. US place of birth is a FATCA indicator requiring W-9 or W-8BEN documentation.',
 'kyc_individual_identity', 'STRING', 'KYC',
 '{"primary": "PASSPORT", "secondary": "BIRTH_CERTIFICATE"}',
 '{"primary": "INDIVIDUAL_REGISTRY", "secondary": "FATCA_ASSESSMENT"}'),

(gen_random_uuid(), 'kyc.individual.country_of_residence',
 'Country where individual primarily resides (ISO 3166-1 alpha-2 code). Used for tax residency determination, CRS reporting, and address verification. May differ from nationality. Multiple residencies require documentation of primary residence. Changes in residence trigger tax reporting updates.',
 'kyc_individual_identity', 'STRING', 'KYC',
 '{"primary": "PROOF_OF_ADDRESS", "secondary": "TAX_RESIDENCY_CERTIFICATE"}',
 '{"primary": "INDIVIDUAL_REGISTRY", "secondary": "CRS_REPORTING"}'),

(gen_random_uuid(), 'kyc.individual.residential_address',
 'Current permanent residential address (not PO Box). Used for correspondence, regulatory reporting, and address verification. Must be verified with utility bill, bank statement, or government letter dated within 3 months. Virtual addresses, mail forwarding services, or privacy service addresses trigger enhanced due diligence.',
 'kyc_individual_identity', 'TEXT', 'KYC',
 '{"primary": "UTILITY_BILL", "secondary": "BANK_STATEMENT", "tertiary": "GOVERNMENT_LETTER"}',
 '{"primary": "INDIVIDUAL_REGISTRY", "secondary": "ADDRESS_VERIFICATION_SERVICE"}'),

-- Individual Identity Documents
(gen_random_uuid(), 'kyc.individual.passport_number',
 'Passport number from primary identity document. Used for international identity verification, travel document authentication, and duplicate detection. Must be valid (not expired). High-risk countries or known fraudulent passport patterns trigger enhanced verification including MRZ code validation.',
 'kyc_individual_documents', 'STRING', 'KYC',
 '{"primary": "PASSPORT_DOCUMENT"}',
 '{"primary": "DOCUMENT_REGISTRY", "secondary": "IDENTITY_VERIFICATION_SYSTEM"}'),

(gen_random_uuid(), 'kyc.individual.passport_issuing_country',
 'Country that issued the passport (ISO 3166-1 alpha-2 code). Used to validate passport format, assess document fraud risk, and cross-check against nationality claims. Mismatch between nationality and passport issuing country requires explanation. Known fraudulent passport jurisdictions trigger manual review.',
 'kyc_individual_documents', 'STRING', 'KYC',
 '{"primary": "PASSPORT_DOCUMENT"}',
 '{"primary": "DOCUMENT_REGISTRY", "secondary": "FRAUD_RISK_ASSESSMENT"}'),

(gen_random_uuid(), 'kyc.individual.passport_expiry_date',
 'Date passport expires in YYYY-MM-DD format. Expired passports cannot be used for identity verification. System must alert before expiry and request updated document. Some jurisdictions require passport validity extending beyond account opening date.',
 'kyc_individual_documents', 'DATE', 'KYC',
 '{"primary": "PASSPORT_DOCUMENT"}',
 '{"primary": "DOCUMENT_REGISTRY", "secondary": "DOCUMENT_EXPIRY_MONITORING"}'),

(gen_random_uuid(), 'kyc.individual.national_id_number',
 'Government-issued national identity number (if available). Format varies by country: US SSN, UK NI number, etc. Highly sensitive PII - must be encrypted at rest. Used for tax reporting, identity verification, and duplicate detection. Optional if passport provided, but enhances identity confidence.',
 'kyc_individual_documents', 'STRING', 'KYC',
 '{"primary": "NATIONAL_ID_CARD", "secondary": "TAX_FORM"}',
 '{"primary": "INDIVIDUAL_REGISTRY", "secondary": "TAX_REPORTING_SYSTEM"}'),

(gen_random_uuid(), 'kyc.individual.tax_identification_number',
 'Tax identification number (TIN) from country of tax residence. Format and name vary by jurisdiction: US SSN/EIN, UK UTR, etc. Required for FATCA and CRS reporting. Multiple tax residencies require TINs from each jurisdiction. Inability to provide TIN must be explained and documented per CRS rules.',
 'kyc_individual_tax', 'STRING', 'KYC',
 '{"primary": "W9_OR_W8BEN", "secondary": "TAX_RESIDENCY_CERTIFICATE", "tertiary": "NATIONAL_ID"}',
 '{"primary": "TAX_REGISTRY", "secondary": "FATCA_CRS_REPORTING"}'),

-- Individual Financial Profile
(gen_random_uuid(), 'kyc.individual.employment_status',
 'Current employment situation. Values: EMPLOYED (specify employer), SELF_EMPLOYED (specify business), RETIRED, STUDENT, UNEMPLOYED. Used for source of wealth assessment and transaction monitoring. High net worth individuals with unclear employment require enhanced wealth source documentation.',
 'kyc_individual_financial', 'ENUM', 'KYC',
 '{"primary": "CLIENT_QUESTIONNAIRE", "secondary": "EMPLOYMENT_LETTER"}',
 '{"primary": "INDIVIDUAL_PROFILE", "secondary": "SOURCE_OF_WEALTH_ASSESSMENT"}'),

(gen_random_uuid(), 'kyc.individual.occupation',
 'Specific job title or profession. Used for PEP screening (government officials, senior executives), sanctions risk (arms dealers, politically connected businesspeople), and transaction monitoring. High-risk occupations (cash-intensive businesses, gambling, arms dealing) trigger enhanced due diligence.',
 'kyc_individual_financial', 'STRING', 'KYC',
 '{"primary": "CLIENT_QUESTIONNAIRE", "secondary": "EMPLOYMENT_LETTER", "tertiary": "LINKEDIN_VERIFICATION"}',
 '{"primary": "INDIVIDUAL_PROFILE", "secondary": "PEP_SCREENING"}'),

(gen_random_uuid(), 'kyc.individual.employer_name',
 'Name of current employer (if employed). Used for source of income verification, PEP screening (government employer = potential PEP), and reputational risk assessment. Government employers, defense contractors, and politically connected companies may trigger enhanced screening.',
 'kyc_individual_financial', 'STRING', 'KYC',
 '{"primary": "EMPLOYMENT_LETTER", "secondary": "CLIENT_QUESTIONNAIRE"}',
 '{"primary": "INDIVIDUAL_PROFILE", "secondary": "PEP_SCREENING"}'),

(gen_random_uuid(), 'kyc.individual.annual_income',
 'Approximate annual income in reporting currency. Ranges: <50K, 50K-100K, 100K-250K, 250K-500K, 500K-1M, >1M. Used for suitability assessment, wealth verification, and detecting unusual transaction patterns. Large deposits inconsistent with declared income trigger SAR investigation.',
 'kyc_individual_financial', 'ENUM', 'KYC',
 '{"primary": "CLIENT_QUESTIONNAIRE", "secondary": "EMPLOYMENT_LETTER", "tertiary": "TAX_RETURNS"}',
 '{"primary": "INDIVIDUAL_PROFILE", "secondary": "TRANSACTION_MONITORING_THRESHOLDS"}'),

(gen_random_uuid(), 'kyc.individual.net_worth',
 'Approximate total net worth in reporting currency. Ranges: <100K, 100K-500K, 500K-1M, 1M-5M, 5M-10M, >10M. Used for suitability assessment, product eligibility, and transaction monitoring. High net worth individuals require documented source of wealth. Inconsistencies with income level trigger investigation.',
 'kyc_individual_financial', 'ENUM', 'KYC',
 '{"primary": "CLIENT_QUESTIONNAIRE", "secondary": "WEALTH_STATEMENT"}',
 '{"primary": "INDIVIDUAL_PROFILE", "secondary": "PRODUCT_SUITABILITY_ASSESSMENT"}'),

(gen_random_uuid(), 'kyc.individual.source_of_wealth',
 'Origin of individual''s accumulated wealth. Free text but common categories: EMPLOYMENT_INCOME, BUSINESS_OWNERSHIP, INHERITANCE, INVESTMENT_RETURNS, REAL_ESTATE, GIFT, DIVORCE_SETTLEMENT. Used to assess credibility and AML risk. Inherited wealth or gifts require documentation of donor/deceased wealth legitimacy. Unclear wealth sources trigger enhanced due diligence.',
 'kyc_individual_financial', 'TEXT', 'KYC',
 '{"primary": "SOURCE_OF_WEALTH_QUESTIONNAIRE", "secondary": "SUPPORTING_DOCUMENTS"}',
 '{"primary": "INDIVIDUAL_PROFILE", "secondary": "ENHANCED_DUE_DILIGENCE_ASSESSMENT"}'),

(gen_random_uuid(), 'kyc.individual.source_of_funds',
 'Specific origin of funds being invested or deposited. More granular than source of wealth. Examples: "salary from ABC Corp", "proceeds from sale of property at 123 Main St", "inheritance from John Smith estate". Must be verifiable. Large or unusual fund sources require documentation (sale agreements, inheritance letters, business sale documents).',
 'kyc_individual_financial', 'TEXT', 'KYC',
 '{"primary": "SOURCE_OF_FUNDS_DECLARATION", "secondary": "SUPPORTING_DOCUMENTS"}',
 '{"primary": "INDIVIDUAL_PROFILE", "secondary": "TRANSACTION_MONITORING"}'),

-- Individual Risk Assessment
(gen_random_uuid(), 'kyc.individual.pep_status',
 'Politically Exposed Person status. Values: NOT_PEP, DOMESTIC_PEP (senior government official in home country), FOREIGN_PEP (senior official in any country), INTERNATIONAL_ORGANIZATION_PEP (UN, World Bank, etc.), RCA (relative or close associate of PEP), FORMER_PEP (out of office >1 year). PEP status requires enhanced due diligence, source of wealth verification, and senior management approval.',
 'kyc_individual_risk', 'ENUM', 'KYC',
 '{"primary": "PEP_DATABASE_SCREENING", "secondary": "MANUAL_RESEARCH", "tertiary": "CLIENT_DECLARATION"}',
 '{"primary": "INDIVIDUAL_PROFILE", "secondary": "ENHANCED_DUE_DILIGENCE_WORKFLOW"}'),

(gen_random_uuid(), 'kyc.individual.pep_position',
 'Specific government or international organization position held (if PEP). Examples: "Minister of Finance", "Member of Parliament", "Central Bank Governor", "UN Ambassador". Position level determines EDD requirements - heads of state and ministers require strictest oversight. Former positions must be documented with dates.',
 'kyc_individual_risk', 'STRING', 'KYC',
 '{"primary": "PEP_DATABASE", "secondary": "GOVERNMENT_WEBSITES", "tertiary": "CLIENT_DECLARATION"}',
 '{"primary": "INDIVIDUAL_PROFILE", "secondary": "EDD_RISK_RATING"}'),

(gen_random_uuid(), 'kyc.individual.sanctions_screening_result',
 'Result of screening individual name, date of birth, nationality, and ID numbers against global sanctions lists (OFAC SDN, UN, EU, HMT, etc.). Values: CLEAR, POTENTIAL_MATCH_UNDER_REVIEW, FALSE_POSITIVE_CLEARED, TRUE_POSITIVE_BLOCKED. Must be CLEAR before account opening. Ongoing monitoring required.',
 'kyc_individual_risk', 'ENUM', 'KYC',
 '{"primary": "SANCTIONS_SCREENING_SYSTEM"}',
 '{"primary": "COMPLIANCE_AUDIT_LOG", "secondary": "ACCOUNT_OPENING_GATE"}'),

(gen_random_uuid(), 'kyc.individual.risk_rating',
 'Overall KYC/AML risk rating for the individual. Calculated based on: nationality, residence, PEP status, occupation, wealth source, product usage, transaction patterns. Values: LOW, MEDIUM, HIGH, VERY_HIGH. Determines review frequency (Low=3yr, Med=2yr, High=1yr, Very High=6mo) and transaction monitoring sensitivity.',
 'kyc_individual_risk', 'ENUM', 'KYC',
 '{"primary": "RISK_ASSESSMENT_ENGINE", "secondary": "COMPLIANCE_OFFICER_OVERRIDE"}',
 '{"primary": "INDIVIDUAL_PROFILE", "secondary": "TRANSACTION_MONITORING_RULES"}'),

(gen_random_uuid(), 'kyc.individual.adverse_media_found',
 'Boolean indicating negative news coverage discovered during screening. Includes: criminal allegations, fraud, money laundering, sanctions violations, terrorist financing, corruption. True value requires detailed investigation, escalation to compliance, and documented senior management decision. May result in relationship decline.',
 'kyc_individual_risk', 'BOOLEAN', 'KYC',
 '{"primary": "ADVERSE_MEDIA_SCREENING", "secondary": "MANUAL_GOOGLE_SEARCH", "tertiary": "NEWS_DATABASE"}',
 '{"primary": "INDIVIDUAL_PROFILE", "secondary": "ENHANCED_DUE_DILIGENCE_WORKFLOW"}'),

-- Individual Suitability & Investor Profile
(gen_random_uuid(), 'kyc.individual.investment_experience',
 'Self-assessed investment knowledge and experience level. Values: NO_EXPERIENCE, LIMITED (personal savings only), MODERATE (stocks/bonds), EXTENSIVE (complex derivatives, hedge funds). Used for product suitability and investor protection. Inexperienced investors restricted from complex or high-risk products. Required for MiFID II suitability assessment.',
 'kyc_individual_suitability', 'ENUM', 'KYC',
 '{"primary": "INVESTOR_PROFILE_QUESTIONNAIRE"}',
 '{"primary": "INDIVIDUAL_PROFILE", "secondary": "PRODUCT_SUITABILITY_ENGINE"}'),

(gen_random_uuid(), 'kyc.individual.risk_tolerance',
 'Investment risk appetite. Values: CONSERVATIVE (capital preservation), MODERATE (balanced growth), AGGRESSIVE (maximum growth, high volatility accepted). Used for product suitability and portfolio construction. Mismatch between risk tolerance and proposed investments triggers suitability warning.',
 'kyc_individual_suitability', 'ENUM', 'KYC',
 '{"primary": "INVESTOR_PROFILE_QUESTIONNAIRE"}',
 '{"primary": "INDIVIDUAL_PROFILE", "secondary": "PRODUCT_SUITABILITY_ENGINE"}'),

(gen_random_uuid(), 'kyc.individual.investment_objectives',
 'Primary goals for investment. Values: CAPITAL_PRESERVATION, INCOME_GENERATION, BALANCED_GROWTH, CAPITAL_APPRECIATION, SPECULATION. Used for product suitability and advisory recommendations. Objectives inconsistent with product selection require documented override by qualified advisor.',
 'kyc_individual_suitability', 'ENUM', 'KYC',
 '{"primary": "INVESTOR_PROFILE_QUESTIONNAIRE"}',
 '{"primary": "INDIVIDUAL_PROFILE", "secondary": "ADVISORY_RECOMMENDATIONS"}'),

(gen_random_uuid(), 'kyc.individual.investment_time_horizon',
 'Expected investment duration. Values: SHORT_TERM (<3 years), MEDIUM_TERM (3-7 years), LONG_TERM (>7 years). Used for product suitability - illiquid investments (private equity, hedge funds) require long-term horizon. Short horizon clients should not hold long lock-up products.',
 'kyc_individual_suitability', 'ENUM', 'KYC',
 '{"primary": "INVESTOR_PROFILE_QUESTIONNAIRE"}',
 '{"primary": "INDIVIDUAL_PROFILE", "secondary": "PRODUCT_SUITABILITY_ENGINE"}');

-- =============================================================================
-- COMMON KYC WORKFLOW ATTRIBUTES
-- =============================================================================

INSERT INTO "dsl-ob-poc".dictionary (attribute_id, name, long_description, group_id, mask, domain, source, sink) VALUES
(gen_random_uuid(), 'kyc.status',
 'Current state of KYC verification process for this client (individual or institutional). Values: NOT_STARTED, DOCUMENTS_REQUESTED, DOCUMENTS_RECEIVED, UNDER_REVIEW, ADDITIONAL_INFO_REQUIRED, ENHANCED_DUE_DILIGENCE, PENDING_APPROVAL, APPROVED, REJECTED, EXPIRED. Drives workflow routing and account opening gates. Approved status required before account activation.',
 'kyc_workflow', 'ENUM', 'KYC',
 '{"primary": "KYC_WORKFLOW_ENGINE"}',
 '{"primary": "ENTITY_OR_INDIVIDUAL_PROFILE", "secondary": "WORKFLOW_STATE_MACHINE"}'),

(gen_random_uuid(), 'kyc.approval_date',
 'Timestamp when KYC was approved by authorized compliance officer. Used to calculate KYC expiry date (typically approval_date + review_period). Required for audit trail. Must be within last 3 years for continued account activity (regulatory requirement).',
 'kyc_workflow', 'TIMESTAMP', 'KYC',
 '{"primary": "KYC_APPROVAL_WORKFLOW"}',
 '{"primary": "COMPLIANCE_AUDIT_LOG", "secondary": "KYC_EXPIRY_MONITORING"}'),

(gen_random_uuid(), 'kyc.next_review_date',
 'Scheduled date for next periodic KYC review. Calculated based on risk rating: Low=3 years, Medium=2 years, High=1 year, Very High=6 months. System generates alerts 30 days before review date. Account restrictions apply if review not completed by due date.',
 'kyc_workflow', 'DATE', 'KYC',
 '{"primary": "KYC_APPROVAL_WORKFLOW", "secondary": "RISK_RATING_CALCULATION"}',
 '{"primary": "ENTITY_OR_INDIVIDUAL_PROFILE", "secondary": "KYC_REVIEW_MONITORING"}'),

(gen_random_uuid(), 'kyc.approved_by',
 'Name or ID of compliance officer who approved KYC. Used for accountability, audit trail, and quality assurance. High-risk clients require senior compliance officer or MLRO approval. Approval authority levels defined in compliance procedures.',
 'kyc_workflow', 'STRING', 'KYC',
 '{"primary": "KYC_APPROVAL_WORKFLOW", "secondary": "HR_SYSTEM"}',
 '{"primary": "COMPLIANCE_AUDIT_LOG"}'),

(gen_random_uuid(), 'kyc.tier',
 'Level of KYC diligence performed. Values: SIMPLIFIED (low-risk, reduced documentation), STANDARD (normal due diligence), ENHANCED (high-risk, additional documentation and approvals). Tier determined by risk assessment algorithm or compliance officer judgment. Enhanced tier required for PEPs, high-risk jurisdictions, complex structures.',
 'kyc_workflow', 'ENUM', 'KYC',
 '{"primary": "RISK_ASSESSMENT_ENGINE", "secondary": "COMPLIANCE_OFFICER_ASSIGNMENT"}',
 '{"primary": "ENTITY_OR_INDIVIDUAL_PROFILE", "secondary": "DOCUMENT_REQUIREMENTS_ENGINE"}'),

(gen_random_uuid(), 'kyc.documents_outstanding',
 'Array of document types still required to complete KYC. Examples: ["PROOF_OF_ADDRESS", "SOURCE_OF_WEALTH_LETTER", "AUDITED_FINANCIALS"]. Used to track completeness and send reminders. Empty array indicates all required documents received. Drives workflow routing to document collection or review stages.',
 'kyc_workflow', 'ARRAY', 'KYC',
 '{"primary": "DOCUMENT_REQUIREMENTS_ENGINE", "secondary": "DOCUMENT_UPLOAD_TRACKER"}',
 '{"primary": "KYC_WORKFLOW_STATE", "secondary": "CLIENT_PORTAL_ALERTS"}'),

(gen_random_uuid(), 'kyc.edd_required',
 'Boolean flag indicating Enhanced Due Diligence is required. True triggers additional documentation requirements, senior approval requirements, and enhanced transaction monitoring. EDD triggered by: PEP status, high-risk jurisdiction, adverse media, complex ownership, cash-intensive business, prior SAR filing.',
 'kyc_workflow', 'BOOLEAN', 'KYC',
 '{"primary": "RISK_ASSESSMENT_ENGINE"}',
 '{"primary": "ENTITY_OR_INDIVIDUAL_PROFILE", "secondary": "WORKFLOW_ROUTING"}'),

(gen_random_uuid(), 'kyc.edd_completion_date',
 'Date when Enhanced Due Diligence procedures were completed. Required for audit trail when EDD performed. Includes: enhanced source of wealth verification, adverse media deep dive, PEP wealth source documentation, senior management interviews, independent verification of information.',
 'kyc_workflow', 'DATE', 'KYC',
 '{"primary": "EDD_WORKFLOW"}',
 '{"primary": "COMPLIANCE_AUDIT_LOG"}'),

(gen_random_uuid(), 'kyc.refresh_trigger_reason',
 'Reason for KYC refresh outside of scheduled review. Values: MATERIAL_CHANGE (ownership, control, business), ADVERSE_MEDIA, SUSPICIOUS_ACTIVITY, REGULATORY_REQUIREMENT, PRODUCT_CHANGE, JURISDICTION_CHANGE. Trigger events require immediate KYC update before continued account activity.',
 'kyc_workflow', 'ENUM', 'KYC',
 '{"primary": "MONITORING_SYSTEMS", "secondary": "MANUAL_TRIGGER"}',
 '{"primary": "KYC_WORKFLOW_STATE", "secondary": "COMPLIANCE_AUDIT_LOG"}'),

(gen_random_uuid(), 'kyc.comments',
 'Free-text field for compliance officer notes and observations. Used to document: unusual circumstances, explanations for discrepancies, rationale for risk rating, approval conditions, follow-up actions. Critical for audit trail and knowledge transfer between compliance officers.',
 'kyc_workflow', 'TEXT', 'KYC',
 '{"primary": "COMPLIANCE_OFFICER_INPUT"}',
 '{"primary": "COMPLIANCE_AUDIT_LOG", "secondary": "KYC_FILE"}');

COMMIT;
