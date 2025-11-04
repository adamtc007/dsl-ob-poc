-- ============================================================================
-- HEDGE FUND INVESTOR DATA DICTIONARY
-- ============================================================================
--
-- Purpose: Comprehensive data dictionary for all hedge fund investor domain attributes
--
-- This dictionary enables:
-- 1. DSL variable validation using attribute UUIDs
-- 2. RAG-powered AI agent with rich semantic context
-- 3. Complete audit trail and data lineage
-- 4. Type-safe attribute resolution in DSL operations
--
-- Schema: "dsl-ob-poc".dictionary
-- Domain: hedge-fund-investor
--
-- Usage in DSL:
--   Instead of: :legal-name "Acme Capital LP"
--   Use: :hf.investor.legal-name @attr{uuid-here} = "Acme Capital LP"
-- ============================================================================

-- Clear existing hedge fund investor attributes (idempotent)
DELETE FROM "dsl-ob-poc".dictionary WHERE domain = 'hedge-fund-investor';

-- ============================================================================
-- INVESTOR IDENTITY ATTRIBUTES
-- Group: hf-investor-identity
-- ============================================================================

INSERT INTO "dsl-ob-poc".dictionary (name, long_description, group_id, mask, domain, vector, source, sink) VALUES
(
  'hf.investor.investor-id',
  'Unique identifier for the hedge fund investor. System-generated UUID that remains constant throughout the investor lifecycle from opportunity through offboarding. This is the primary key for all investor-related operations and state transitions.',
  'hf-investor-identity',
  'uuid',
  'hedge-fund-investor',
  'hedge fund investor unique identifier uuid primary key investor id system generated immutable lifecycle tracking investor register',
  '{"system": "hedge-fund-investor", "table": "hf_investors", "column": "investor_id", "generated": true}',
  '{"table": "hf_investors", "column": "investor_id", "constraints": ["PRIMARY KEY", "NOT NULL"]}'
),
(
  'hf.investor.investor-code',
  'Human-readable investor code used for reporting and external references. Format typically "INV-YYYY-NNN" where YYYY is year and NNN is sequential number. Unique across all investors and used in client communications, reports, and external interfaces. Must be unique and cannot be changed once assigned.',
  'hf-investor-identity',
  'string',
  'hedge-fund-investor',
  'investor code human readable unique identifier reporting external reference client communication investor number account number',
  '{"system": "hedge-fund-investor", "table": "hf_investors", "column": "investor_code", "format": "INV-YYYY-NNN"}',
  '{"table": "hf_investors", "column": "investor_code", "constraints": ["UNIQUE", "NOT NULL"], "max_length": 50}'
),
(
  'hf.investor.legal-name',
  'Official legal name of the investor entity or individual as it appears on incorporation documents, passports, or other legal identification. This is the name used for all legal agreements, subscription documents, tax forms, and regulatory reporting. For individuals: full legal name. For entities: exact name as registered. Critical for KYC/AML compliance and legal documentation.',
  'hf-investor-identity',
  'string',
  'hedge-fund-investor',
  'legal name official name entity name individual name full name investor name legal entity registered name incorporation name kyc compliance',
  '{"system": "hedge-fund-investor", "collector": "kyc-process", "required": true, "verification": "documentation"}',
  '{"table": "hf_investors", "column": "legal_name", "constraints": ["NOT NULL"], "max_length": 500}'
),
(
  'hf.investor.short-name',
  'Abbreviated or commonly used name for the investor. Optional field used for informal communications, reports, and internal references where full legal name is not required. For example: "Acme Capital" instead of "Acme Capital Partners LP". Makes reports and displays more readable while maintaining legal name for official purposes.',
  'hf-investor-identity',
  'string',
  'hedge-fund-investor',
  'short name abbreviated name common name informal name display name investor nickname trading name',
  '{"system": "hedge-fund-investor", "optional": true, "usage": "display"}',
  '{"table": "hf_investors", "column": "short_name", "constraints": [], "max_length": 100}'
),
(
  'hf.investor.type',
  'Classification of the investor entity type. Determines regulatory treatment, KYC requirements, and documentation needs. Valid values: INDIVIDUAL (natural person), CORPORATE (corporation/company), TRUST (trust structure), FOHF (fund of hedge funds), NOMINEE (nominee account), PENSION_FUND (pension/retirement fund), INSURANCE_CO (insurance company). Each type has specific compliance and reporting requirements under different regulations (AIFMD, MiFID II, etc.).',
  'hf-investor-identity',
  'enum',
  'hedge-fund-investor',
  'investor type entity type classification individual corporate trust fund pension insurance nominee regulatory classification entity classification',
  '{"system": "hedge-fund-investor", "required": true, "values": ["INDIVIDUAL", "CORPORATE", "TRUST", "FOHF", "NOMINEE", "PENSION_FUND", "INSURANCE_CO"]}',
  '{"table": "hf_investors", "column": "type", "constraints": ["NOT NULL", "CHECK"], "max_length": 20}'
),
(
  'hf.investor.domicile',
  'Country of domicile or tax residence for the investor. Two-letter ISO 3166-1 alpha-2 country code (e.g., US, UK, CH, LU). Determines tax treaty applicability, withholding rates, CRS reporting obligations, and regulatory framework. Critical for tax form selection (W-8BEN vs W-9), FATCA classification, and cross-border investment compliance. Must align with incorporation jurisdiction for entities.',
  'hf-investor-identity',
  'country-code',
  'hedge-fund-investor',
  'domicile country residence tax residence jurisdiction country code iso tax treaty withholding crs fatca regulatory',
  '{"system": "hedge-fund-investor", "required": true, "format": "ISO-3166-1-alpha-2", "validation": "country-code"}',
  '{"table": "hf_investors", "column": "domicile", "constraints": ["NOT NULL"], "max_length": 5}'
),
(
  'hf.investor.lei',
  'Legal Entity Identifier - 20-character alphanumeric code for entity identification in financial transactions. Required for corporate investors under MiFID II and Dodd-Frank regulations. Format: 20 characters starting with 2-digit LOU code, followed by 2 reserved digits, 16-digit entity identifier, and 2 check digits. Obtained from LEI issuing organizations (GLEIF network). Not applicable for individual investors.',
  'hf-investor-identity',
  'string',
  'hedge-fund-investor',
  'lei legal entity identifier mifid regulatory identifier entity code financial identifier corporate identifier gleif',
  '{"system": "hedge-fund-investor", "optional": true, "format": "20-char-alphanumeric", "validation": "lei-format", "issuer": "GLEIF"}',
  '{"table": "hf_investors", "column": "lei", "constraints": [], "max_length": 20}'
),
(
  'hf.investor.registration-number',
  'Company registration or incorporation number issued by relevant jurisdiction authority. For US: EIN or state registration. For UK: Companies House number. For offshore: varies by jurisdiction. Used for entity verification, KYC documentation, and regulatory reporting. Format and issuing authority vary by domicile.',
  'hf-investor-identity',
  'string',
  'hedge-fund-investor',
  'registration number incorporation number company number ein business number entity number jurisdiction registration',
  '{"system": "hedge-fund-investor", "optional": true, "jurisdiction_dependent": true}',
  '{"table": "hf_investors", "column": "registration_number", "constraints": [], "max_length": 100}'
),
(
  'hf.investor.source',
  'Origin or referral source of the investor lead. Examples: "Wealth Advisor Referral", "Institutional Consultant", "Direct Marketing", "Existing Client Network", "Conference Q1 2024". Used for marketing attribution, relationship tracking, and source analysis. Optional field for business intelligence and relationship management.',
  'hf-investor-identity',
  'string',
  'hedge-fund-investor',
  'source referral origin lead source marketing channel investor origin relationship source attribution',
  '{"system": "hedge-fund-investor", "optional": true, "usage": "analytics"}',
  '{"table": "hf_investors", "column": "source", "constraints": [], "max_length": 255}'
);

-- ============================================================================
-- INVESTOR ADDRESS ATTRIBUTES
-- Group: hf-investor-address
-- ============================================================================

INSERT INTO "dsl-ob-poc".dictionary (name, long_description, group_id, mask, domain, vector, source, sink) VALUES
(
  'hf.investor.address-line1',
  'Primary address line for investor correspondence and legal documentation. Street address, building number, or PO Box. First line of multi-line address format. Required for KYC documentation and regulatory correspondence. Must match address on official documents (incorporation certificate, utility bills, bank statements) for verification purposes.',
  'hf-investor-address',
  'string',
  'hedge-fund-investor',
  'address line1 street address primary address correspondence address legal address physical address mailing address',
  '{"system": "hedge-fund-investor", "collector": "kyc-process", "verification": "documentation"}',
  '{"table": "hf_investors", "column": "address_line1", "constraints": [], "max_length": 255}'
),
(
  'hf.investor.address-line2',
  'Secondary address line for additional address details. Suite number, apartment number, floor, building name, or additional street information. Optional field to accommodate complex address formats across different jurisdictions.',
  'hf-investor-address',
  'string',
  'hedge-fund-investor',
  'address line2 suite apartment floor building unit secondary address additional address',
  '{"system": "hedge-fund-investor", "optional": true}',
  '{"table": "hf_investors", "column": "address_line2", "constraints": [], "max_length": 255}'
),
(
  'hf.investor.city',
  'City or municipality name for investor address. Used in address verification, geographic analysis, and correspondence routing. Should match city name on legal documents for KYC purposes.',
  'hf-investor-address',
  'string',
  'hedge-fund-investor',
  'city municipality town location address city geographic location',
  '{"system": "hedge-fund-investor", "optional": true}',
  '{"table": "hf_investors", "column": "city", "constraints": [], "max_length": 100}'
),
(
  'hf.investor.state-province',
  'State, province, region, or canton name. Format varies by country: US states (2-letter code preferred), Canadian provinces, UK counties, Swiss cantons, etc. Important for US tax purposes (state withholding) and geographic reporting.',
  'hf-investor-address',
  'string',
  'hedge-fund-investor',
  'state province region canton county geographic subdivision administrative division',
  '{"system": "hedge-fund-investor", "optional": true, "format": "jurisdiction-dependent"}',
  '{"table": "hf_investors", "column": "state_province", "constraints": [], "max_length": 100}'
),
(
  'hf.investor.postal-code',
  'Postal or ZIP code for address. Format varies by country: US ZIP (5 or 9 digits), UK postcode, Swiss postal code, etc. Used for address validation and mail routing.',
  'hf-investor-address',
  'string',
  'hedge-fund-investor',
  'postal code zip code postcode mail code address code',
  '{"system": "hedge-fund-investor", "optional": true, "format": "country-dependent"}',
  '{"table": "hf_investors", "column": "postal_code", "constraints": [], "max_length": 20}'
),
(
  'hf.investor.country',
  'Country for investor address. Two-letter ISO 3166-1 alpha-2 country code. May differ from domicile (e.g., Swiss investor with UK correspondence address). Used for mail routing and regulatory reporting.',
  'hf-investor-address',
  'country-code',
  'hedge-fund-investor',
  'country address country correspondence country mail country iso country code',
  '{"system": "hedge-fund-investor", "optional": true, "format": "ISO-3166-1-alpha-2"}',
  '{"table": "hf_investors", "column": "country", "constraints": [], "max_length": 5}'
);

-- ============================================================================
-- INVESTOR CONTACT ATTRIBUTES
-- Group: hf-investor-contact
-- ============================================================================

INSERT INTO "dsl-ob-poc".dictionary (name, long_description, group_id, mask, domain, vector, source, sink) VALUES
(
  'hf.investor.primary-contact-name',
  'Name of primary contact person for investor relations and operational communications. For corporate investors: authorized representative, CFO, investment manager. For individuals: may be the investor themselves or appointed representative. Used for all non-legal correspondence, confirmations, and operational updates.',
  'hf-investor-contact',
  'string',
  'hedge-fund-investor',
  'contact name primary contact representative authorized person contact person investor relations operational contact',
  '{"system": "hedge-fund-investor", "optional": true, "usage": "operations"}',
  '{"table": "hf_investors", "column": "primary_contact_name", "constraints": [], "max_length": 255}'
),
(
  'hf.investor.primary-contact-email',
  'Email address for primary contact. Used for transaction confirmations, statements, NAV updates, and operational communications. Must be validated and active. Critical for digital communication workflow and audit trail.',
  'hf-investor-contact',
  'email',
  'hedge-fund-investor',
  'contact email email address correspondence email primary email investor email operational email communication email',
  '{"system": "hedge-fund-investor", "optional": true, "validation": "email-format", "usage": "operations"}',
  '{"table": "hf_investors", "column": "primary_contact_email", "constraints": [], "max_length": 255}'
),
(
  'hf.investor.primary-contact-phone',
  'Phone number for primary contact. International format preferred (+1-XXX-XXX-XXXX). Used for urgent communications, trade confirmations, and operational queries. Optional but recommended for operational efficiency.',
  'hf-investor-contact',
  'phone',
  'hedge-fund-investor',
  'contact phone phone number telephone contact number primary phone investor phone',
  '{"system": "hedge-fund-investor", "optional": true, "format": "international"}',
  '{"table": "hf_investors", "column": "primary_contact_phone", "constraints": [], "max_length": 50}'
);

-- ============================================================================
-- INVESTOR LIFECYCLE ATTRIBUTES
-- Group: hf-investor-lifecycle
-- ============================================================================

INSERT INTO "dsl-ob-poc".dictionary (name, long_description, group_id, mask, domain, vector, source, sink) VALUES
(
  'hf.investor.status',
  'Current lifecycle state of the investor in the onboarding and investment workflow. 11 states: OPPORTUNITY (initial lead), PRECHECKS (interest confirmed), KYC_PENDING (KYC in progress), KYC_APPROVED (ready to invest), SUB_PENDING_CASH (awaiting subscription funds), FUNDED_PENDING_NAV (cash received, awaiting pricing), ISSUED (units allocated), ACTIVE (holding position), REDEEM_PENDING (redemption requested), REDEEMED (fully exited), OFFBOARDED (relationship closed). Each state has specific guard conditions and valid transitions defined in state machine.',
  'hf-investor-lifecycle',
  'enum',
  'hedge-fund-investor',
  'investor status lifecycle state workflow state onboarding state current state investor state machine state transition',
  '{"system": "hedge-fund-investor", "required": true, "state_machine": true, "values": ["OPPORTUNITY", "PRECHECKS", "KYC_PENDING", "KYC_APPROVED", "SUB_PENDING_CASH", "FUNDED_PENDING_NAV", "ISSUED", "ACTIVE", "REDEEM_PENDING", "REDEEMED", "OFFBOARDED"]}',
  '{"table": "hf_investors", "column": "status", "constraints": ["NOT NULL"], "max_length": 50}'
);

-- ============================================================================
-- KYC PROFILE ATTRIBUTES
-- Group: hf-kyc-profile
-- ============================================================================

INSERT INTO "dsl-ob-poc".dictionary (name, long_description, group_id, mask, domain, vector, source, sink) VALUES
(
  'hf.kyc.risk-rating',
  'AML/CFT risk rating assigned to investor after KYC assessment. Values: LOW (minimal risk, standard monitoring), MEDIUM (normal risk, enhanced monitoring), HIGH (elevated risk, intensive monitoring), PROHIBITED (cannot onboard). Determines monitoring intensity, refresh frequency, and approval authority. Based on factors: jurisdiction risk, investor type, source of funds, beneficial ownership, PEP status, sanctions exposure.',
  'hf-kyc-profile',
  'enum',
  'hedge-fund-investor',
  'risk rating kyc risk aml risk cft risk investor risk compliance risk rating low medium high prohibited risk assessment',
  '{"system": "hedge-fund-investor", "collector": "kyc-approval", "required": true, "values": ["LOW", "MEDIUM", "HIGH", "PROHIBITED"]}',
  '{"table": "hf_kyc_profiles", "column": "risk_rating", "constraints": ["NOT NULL"], "max_length": 20}'
),
(
  'hf.kyc.tier',
  'KYC due diligence tier determining documentation depth and verification requirements. SIMPLIFIED: basic checks for low-risk investors. STANDARD: normal due diligence with full documentation. ENHANCED: intensive due diligence for high-risk investors, PEPs, or complex structures. Determines required documents, verification depth, and approval authority.',
  'hf-kyc-profile',
  'enum',
  'hedge-fund-investor',
  'kyc tier due diligence tier verification level kyc level simplified standard enhanced edd cdd',
  '{"system": "hedge-fund-investor", "required": true, "values": ["SIMPLIFIED", "STANDARD", "ENHANCED"]}',
  '{"table": "hf_kyc_profiles", "column": "kyc_tier", "constraints": [], "max_length": 20}'
),
(
  'hf.kyc.screening-provider',
  'Third-party screening service used for AML/sanctions/PEP checks. Common providers: worldcheck (Refinitiv World-Check), refinitiv (Refinitiv broader suite), accelus (Accelus KYC), dow-jones (Dow Jones Risk & Compliance), comply-advantage. Provider choice affects data sources, coverage, and integration requirements.',
  'hf-kyc-profile',
  'enum',
  'hedge-fund-investor',
  'screening provider aml provider sanctions screening kyc screening worldcheck refinitiv compliance provider',
  '{"system": "hedge-fund-investor", "optional": true, "values": ["worldcheck", "refinitiv", "accelus", "dow-jones", "comply-advantage"]}',
  '{"table": "hf_kyc_profiles", "column": "screening_provider", "constraints": [], "max_length": 50}'
),
(
  'hf.kyc.screening-result',
  'Outcome of AML/sanctions/PEP screening. CLEAR: no adverse findings, proceed. POTENTIAL_MATCH: possible match requiring review. TRUE_POSITIVE: confirmed match, enhanced due diligence or rejection required. Result drives risk rating and approval workflow.',
  'hf-kyc-profile',
  'enum',
  'hedge-fund-investor',
  'screening result aml result sanctions result pep result compliance result clear match true positive',
  '{"system": "hedge-fund-investor", "optional": true, "values": ["CLEAR", "POTENTIAL_MATCH", "TRUE_POSITIVE"]}',
  '{"table": "hf_kyc_profiles", "column": "screening_result", "constraints": [], "max_length": 50}'
),
(
  'hf.kyc.approved-by',
  'Name and title of person who approved KYC. Format: "Name, Title" (e.g., "Sarah Johnson, Head of Compliance"). Required for audit trail and regulatory examinations. Approval authority must match risk rating (e.g., HIGH risk requires senior management approval).',
  'hf-kyc-profile',
  'string',
  'hedge-fund-investor',
  'approved by kyc approver compliance approver approval authority authorizing person audit trail',
  '{"system": "hedge-fund-investor", "required_for": "approval", "audit": true}',
  '{"table": "hf_kyc_profiles", "column": "approved_by", "constraints": [], "max_length": 255}'
),
(
  'hf.kyc.refresh-frequency',
  'How often KYC must be refreshed. MONTHLY: very high risk. QUARTERLY: high risk. SEMI_ANNUAL: medium-high risk. ANNUAL: standard. BIENNIAL: low risk. Determines next_refresh_due date and triggers automated refresh reminders.',
  'hf-kyc-profile',
  'enum',
  'hedge-fund-investor',
  'refresh frequency kyc refresh review frequency periodic review refresh cycle monitoring frequency',
  '{"system": "hedge-fund-investor", "required": true, "values": ["MONTHLY", "QUARTERLY", "SEMI_ANNUAL", "ANNUAL", "BIENNIAL"]}',
  '{"table": "hf_kyc_profiles", "column": "refresh_frequency", "constraints": ["NOT NULL"], "max_length": 20}'
),
(
  'hf.kyc.refresh-due-at',
  'Date when KYC refresh is due. Calculated from approval date plus refresh frequency. Triggers workflow reminders at 90, 60, 30 days before. Overdue KYC may trigger trading restrictions or position freeze depending on policy.',
  'hf-kyc-profile',
  'date',
  'hedge-fund-investor',
  'refresh due date kyc due date review due expiry date refresh date monitoring date',
  '{"system": "hedge-fund-investor", "required": true, "calculated": true, "alerts": true}',
  '{"table": "hf_kyc_profiles", "column": "refresh_due_at", "constraints": []}'
);

-- ============================================================================
-- TAX PROFILE ATTRIBUTES
-- Group: hf-tax-profile
-- ============================================================================

INSERT INTO "dsl-ob-poc".dictionary (name, long_description, group_id, mask, domain, vector, source, sink) VALUES
(
  'hf.tax.fatca-status',
  'FATCA (Foreign Account Tax Compliance Act) classification for US tax purposes. US_PERSON: US citizen/resident, use W-9. NON_US_PERSON: non-US with no US tax nexus, use W-8. SPECIFIED_US_PERSON: US person subject to FATCA reporting. EXEMPT_BENEFICIAL_OWNER: exempt from FATCA (governments, pension funds, etc.). Determines withholding obligations and IRS reporting requirements (Form 1099, 1042, 1042-S).',
  'hf-tax-profile',
  'enum',
  'hedge-fund-investor',
  'fatca status tax status us person withholding classification fatca classification w8 w9 us tax',
  '{"system": "hedge-fund-investor", "required": true, "values": ["US_PERSON", "NON_US_PERSON", "SPECIFIED_US_PERSON", "EXEMPT_BENEFICIAL_OWNER"], "regulatory": "FATCA"}',
  '{"table": "hf_tax_profiles", "column": "fatca_status", "constraints": [], "max_length": 50}'
),
(
  'hf.tax.crs-classification',
  'Common Reporting Standard classification for OECD automatic exchange of information. INDIVIDUAL: reportable individual account holder. ENTITY: legal entity (further classification required). FINANCIAL_INSTITUTION: banks, funds, etc. (exempt from reporting). INVESTMENT_ENTITY: passive NFE or managed investment vehicle. Determines CRS reporting to tax authorities in participating jurisdictions.',
  'hf-tax-profile',
  'enum',
  'hedge-fund-investor',
  'crs classification tax classification oecd reporting entity classification individual entity financial institution',
  '{"system": "hedge-fund-investor", "required": true, "values": ["INDIVIDUAL", "ENTITY", "FINANCIAL_INSTITUTION", "INVESTMENT_ENTITY"], "regulatory": "CRS-OECD"}',
  '{"table": "hf_tax_profiles", "column": "crs_classification", "constraints": [], "max_length": 50}'
),
(
  'hf.tax.form-type',
  'Tax form type submitted by investor. W9: US persons. W8_BEN: non-US individuals claiming treaty benefits. W8_BEN_E: non-US entities claiming treaty benefits or exempt status. W8_ECI: income effectively connected with US trade. W8_EXP: governments, tax-exempt organizations. W8_IMY: intermediaries, flow-through entities. ENTITY_SELF_CERT: CRS self-certification for non-US entities. Form validity: W-8 forms expire after 3 years, W-9 valid until change in circumstances.',
  'hf-tax-profile',
  'enum',
  'hedge-fund-investor',
  'tax form w8 w9 form type tax documentation withholding form certification form',
  '{"system": "hedge-fund-investor", "required": true, "values": ["W9", "W8_BEN", "W8_BEN_E", "W8_ECI", "W8_EXP", "W8_IMY", "ENTITY_SELF_CERT"], "validity": "3-years-for-W8"}',
  '{"table": "hf_tax_profiles", "column": "form_type", "constraints": [], "max_length": 50}'
),
(
  'hf.tax.withholding-rate',
  'Withholding tax rate applied to income distributions. Decimal format (0.30 = 30%). Default US rate: 30% for non-treaty investors. Reduced rates available under tax treaties (0%, 15%, etc.). Zero for US persons. Rate determined by: investor tax status, domicile, treaty applicability, form validity. Applied to dividends, interest, and other distributions.',
  'hf-tax-profile',
  'decimal',
  'hedge-fund-investor',
  'withholding rate tax rate withholding tax treaty rate distribution tax rate income tax',
  '{"system": "hedge-fund-investor", "required": true, "default": 0.30, "range": [0.0, 1.0], "format": "decimal"}',
  '{"table": "hf_tax_profiles", "column": "withholding_rate", "constraints": ["NOT NULL"]}'
),
(
  'hf.tax.tin-type',
  'Type of Tax Identification Number provided. SSN: US Social Security Number (individuals). ITIN: Individual Taxpayer Identification Number (non-citizens). EIN: Employer Identification Number (US entities). FOREIGN_TIN: non-US tax number. GIIN: Global Intermediary Identification Number (FATCA). Type must align with investor type and domicile.',
  'hf-tax-profile',
  'enum',
  'hedge-fund-investor',
  'tin type tax id type ssn ein itin giin foreign tin identification type',
  '{"system": "hedge-fund-investor", "optional": true, "values": ["SSN", "ITIN", "EIN", "FOREIGN_TIN", "GIIN"]}',
  '{"table": "hf_tax_profiles", "column": "tin_type", "constraints": [], "max_length": 20}'
),
(
  'hf.tax.tin-value',
  'Actual Tax Identification Number. Format varies by type: SSN (XXX-XX-XXXX), EIN (XX-XXXXXXX), GIIN (XXXXXX.XXXXX.XX.XXX). Highly sensitive PII - must be encrypted at rest, masked in displays, access-controlled. Required for tax reporting (1099, 1042-S) and validation.',
  'hf-tax-profile',
  'ssn',
  'hedge-fund-investor',
  'tin value tax id number ssn ein itin identification number tax number sensitive pii',
  '{"system": "hedge-fund-investor", "sensitive": true, "encryption": "required", "masking": "required", "pii": true}',
  '{"table": "hf_tax_profiles", "column": "tin_value", "constraints": [], "max_length": 50, "encryption": "AES-256"}'
);

-- ============================================================================
-- FUND STRUCTURE ATTRIBUTES
-- Group: hf-fund-structure
-- ============================================================================

INSERT INTO "dsl-ob-poc".dictionary (name, long_description, group_id, mask, domain, vector, source, sink) VALUES
(
  'hf.fund.fund-id',
  'Unique identifier for the hedge fund. System-generated UUID linking all share classes, series, trades, and NAV records. One fund may have multiple share classes with different fee structures, currencies, and terms.',
  'hf-fund-structure',
  'uuid',
  'hedge-fund-investor',
  'fund id fund identifier hedge fund uuid fund uuid investment fund primary key',
  '{"system": "hedge-fund-investor", "table": "hf_funds", "column": "fund_id", "generated": true}',
  '{"table": "hf_funds", "column": "fund_id", "constraints": ["PRIMARY KEY", "NOT NULL"]}'
),
(
  'hf.fund.fund-name',
  'Official name of the hedge fund. Used in all investor communications, subscription documents, and reports. Examples: "Global Opportunities Hedge Fund", "Market Neutral Strategy Fund". Unique across all funds managed by administrator.',
  'hf-fund-structure',
  'string',
  'hedge-fund-investor',
  'fund name hedge fund name investment fund name fund title official name',
  '{"system": "hedge-fund-investor", "required": true}',
  '{"table": "hf_funds", "column": "fund_name", "constraints": ["NOT NULL", "UNIQUE"], "max_length": 255}'
),
(
  'hf.fund.class-id',
  'Unique identifier for a share class within a fund. Share classes have different: currencies, fee structures, minimum investments, dealing frequencies, lock-up periods. Common class names: A (retail), I (institutional), S (seeded), E (employee). UUID links to class-specific trades and positions.',
  'hf-fund-structure',
  'uuid',
  'hedge-fund-investor',
  'class id share class uuid class identifier class uuid investment class',
  '{"system": "hedge-fund-investor", "table": "hf_share_classes", "column": "class_id", "generated": true}',
  '{"table": "hf_share_classes", "column": "class_id", "constraints": ["PRIMARY KEY", "NOT NULL"]}'
),
(
  'hf.
