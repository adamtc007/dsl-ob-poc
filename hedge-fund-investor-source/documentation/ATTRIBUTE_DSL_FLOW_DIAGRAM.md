# Attribute-Based DSL Flow - Visual Diagram

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                    ATTRIBUTE-BASED DSL ARCHITECTURE                          │
│                  with RAG-Powered AI DSL Generation                          │
└──────────────────────────────────────────────────────────────────────────────┘

┌─────────────────┐
│  USER           │
│  Natural        │
│  Language       │
└────────┬────────┘
         │
         │ "Create opportunity for
         │  Acme Capital LP, Swiss corporate investor"
         │
         ▼
┌────────────────────────────────────────────────────────────────────────────┐
│  AI AGENT with RAG (Retrieval Augmented Generation)                       │
│  ┌──────────────────────────────────────────────────────────────────┐    │
│  │  1. Parse Instruction → Extract Concepts                         │    │
│  │     "swiss" "corporate" "investor" "opportunity"                 │    │
│  │                                                                   │    │
│  │  2. Semantic Search Dictionary (Vector Similarity)               │    │
│  │     SELECT * FROM dictionary                                     │    │
│  │     WHERE domain = 'hedge-fund-investor'                         │    │
│  │       AND vector @@ 'swiss corporate investor domicile type'    │    │
│  │                                                                   │    │
│  │  3. Retrieved Attributes:                                        │    │
│  │     ┌─────────────────────────────────────────────────────────┐ │    │
│  │     │ UUID: a1b2c3d4-0001-0000-0000-000000000001              │ │    │
│  │     │ Name: hf.investor.legal-name                            │ │    │
│  │     │ Type: string                                            │ │    │
│  │     │ Desc: "Official legal name for all legal agreements..." │ │    │
│  │     ├─────────────────────────────────────────────────────────┤ │    │
│  │     │ UUID: a1b2c3d4-0002-0000-0000-000000000002              │ │    │
│  │     │ Name: hf.investor.type                                  │ │    │
│  │     │ Type: enum                                              │ │    │
│  │     │ Values: [INDIVIDUAL, CORPORATE, TRUST, ...]            │ │    │
│  │     ├─────────────────────────────────────────────────────────┤ │    │
│  │     │ UUID: a1b2c3d4-0003-0000-0000-000000000003              │ │    │
│  │     │ Name: hf.investor.domicile                              │ │    │
│  │     │ Type: country-code (ISO-3166-1-alpha-2)                 │ │    │
│  │     │ Desc: "Country of domicile or tax residence..."         │ │    │
│  │     └─────────────────────────────────────────────────────────┘ │    │
│  │                                                                   │    │
│  │  4. Generate DSL with Attribute UUIDs                             │    │
│  └──────────────────────────────────────────────────────────────────┘    │
└────────┬───────────────────────────────────────────────────────────────────┘
         │
         │ Generated DSL:
         │ (investor.start-opportunity
         │   @attr{a1b2c3d4-0001} = "Acme Capital Partners LP"
         │   @attr{a1b2c3d4-0002} = "CORPORATE"
         │   @attr{a1b2c3d4-0003} = "CH")
         │
         ▼
┌────────────────────────────────────────────────────────────────────────────┐
│  DSL PARSER with Attribute Validation                                     │
│  ┌──────────────────────────────────────────────────────────────────┐    │
│  │  1. Parse S-Expression                                           │    │
│  │     Verb: "investor.start-opportunity"                           │    │
│  │     Attributes: [@attr{uuid-0001}, @attr{uuid-0002}, ...]       │    │
│  │                                                                   │    │
│  │  2. Validate Each Attribute UUID                                 │    │
│  │     ✓ UUID a1b2c3d4-0001 exists in dictionary                   │    │
│  │     ✓ UUID a1b2c3d4-0002 exists in dictionary                   │    │
│  │     ✓ UUID a1b2c3d4-0003 exists in dictionary                   │    │
│  │                                                                   │    │
│  │  3. Type Validation                                              │    │
│  │     ✓ Value "Acme..." matches type 'string'                     │    │
│  │     ✓ Value "CORPORATE" in enum [INDIVIDUAL, CORPORATE, ...]    │    │
│  │     ✓ Value "CH" is valid ISO-3166 country code                 │    │
│  │                                                                   │    │
│  │  4. Resolve to Database Schema                                   │    │
│  │     hf.investor.legal-name → hf_investors.legal_name            │    │
│  │     hf.investor.type → hf_investors.type                        │    │
│  │     hf.investor.domicile → hf_investors.domicile                │    │
│  └──────────────────────────────────────────────────────────────────┘    │
└────────┬───────────────────────────────────────────────────────────────────┘
         │
         │ Validated & Parsed DSL
         │
         ▼
┌────────────────────────────────────────────────────────────────────────────┐
│  DSL EXECUTOR                                                              │
│  ┌──────────────────────────────────────────────────────────────────┐    │
│  │  1. Execute Database Operations                                  │    │
│  │     INSERT INTO hf_investors (                                   │    │
│  │       investor_id, legal_name, type, domicile, status            │    │
│  │     ) VALUES (                                                    │    │
│  │       uuid_generate_v4(),                                        │    │
│  │       'Acme Capital Partners LP',                                │    │
│  │       'CORPORATE',                                               │    │
│  │       'CH',                                                       │    │
│  │       'OPPORTUNITY'                                              │    │
│  │     );                                                            │    │
│  │                                                                   │    │
│  │  2. Store Attribute Values (Audit Trail)                         │    │
│  │     INSERT INTO attribute_values (                               │    │
│  │       cbu_id, attribute_id, value, source                        │    │
│  │     ) VALUES (                                                    │    │
│  │       'new-investor-uuid',                                       │    │
│  │       'a1b2c3d4-0001',                                           │    │
│  │       '{"value": "Acme Capital Partners LP"}',                   │    │
│  │       '{"dsl_operation": "investor.start-opportunity", ...}'     │    │
│  │     );                                                            │    │
│  │                                                                   │    │
│  │  3. Store DSL Execution Record                                   │    │
│  │     INSERT INTO hf_dsl_executions (                              │    │
│  │       investor_id, dsl_text, execution_status, triggered_by      │    │
│  │     ) VALUES (                                                    │    │
│  │       'new-investor-uuid',                                       │    │
│  │       '(investor.start-opportunity @attr{...}...)',              │    │
│  │       'COMPLETED',                                               │    │
│  │       'operations@fundadmin.com'                                 │    │
│  │     );                                                            │    │
│  └──────────────────────────────────────────────────────────────────┘    │
└────────┬───────────────────────────────────────────────────────────────────┘
         │
         │ Result: Investor Created
         │
         ▼
┌────────────────────────────────────────────────────────────────────────────┐
│  DATABASE STATE                                                            │
│  ┌─────────────────────────────────────┬──────────────────────────────┐   │
│  │  hf_investors                       │  attribute_values            │   │
│  ├─────────────────────────────────────┼──────────────────────────────┤   │
│  │  investor_id: <uuid>                │  attribute_id: uuid-0001     │   │
│  │  legal_name: Acme Capital Partners  │  value: "Acme Capital..."    │   │
│  │  type: CORPORATE                    │  source: {dsl_operation...}  │   │
│  │  domicile: CH                       │  observed_at: 2024-01-15     │   │
│  │  status: OPPORTUNITY                │  ─────────────────────────   │   │
│  └─────────────────────────────────────┤  attribute_id: uuid-0002     │   │
│                                         │  value: "CORPORATE"          │   │
│  ┌─────────────────────────────────────┤  source: {dsl_operation...}  │   │
│  │  hf_dsl_executions                  │  ─────────────────────────   │   │
│  ├─────────────────────────────────────┤  attribute_id: uuid-0003     │   │
│  │  execution_id: <uuid>               │  value: "CH"                 │   │
│  │  investor_id: <uuid>                │  source: {dsl_operation...}  │   │
│  │  dsl_text: (investor.start-opport...│                              │   │
│  │  execution_status: COMPLETED        │                              │   │
│  │  triggered_by: operations@fund...  │                              │   │
│  └─────────────────────────────────────┴──────────────────────────────┘   │
└────────────────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────────────────┐
│  AUDIT TRAIL & DATA LINEAGE                                                  │
│  ────────────────────────────────────────────────────────────────────────────│
│  Query: "Where did 'Acme Capital Partners LP' come from?"                    │
│                                                                               │
│  SELECT                                                                       │
│    d.name as attribute,                                                      │
│    av.value,                                                                 │
│    av.source->>'dsl_operation' as operation,                                │
│    de.triggered_by,                                                          │
│    de.created_at                                                             │
│  FROM attribute_values av                                                    │
│  JOIN dictionary d ON d.attribute_id = av.attribute_id                       │
│  JOIN hf_dsl_executions de ON de.investor_id::text = av.cbu_id              │
│  WHERE d.name = 'hf.investor.legal-name'                                     │
│    AND av.value->>'value' = 'Acme Capital Partners LP';                      │
│                                                                               │
│  Result:                                                                      │
│  ┌────────────────────┬───────────────────┬──────────────────────────┐      │
│  │ attribute          │ operation         │ triggered_by             │      │
│  ├────────────────────┼───────────────────┼──────────────────────────┤      │
│  │ hf.investor.legal- │ investor.start-   │ operations@fundadmin.com │      │
│  │ name               │ opportunity       │ 2024-01-15 10:00:00      │      │
│  └────────────────────┴───────────────────┴──────────────────────────┘      │
└──────────────────────────────────────────────────────────────────────────────┘
```

## Key Benefits Illustrated

1. **AI RAG**: Semantic search finds relevant attributes from rich metadata
2. **Type Safety**: Parser validates UUIDs and types before execution
3. **Auditability**: Complete lineage from instruction → DSL → data → storage
4. **Deterministic**: Same instruction always generates same DSL structure
5. **Self-Describing**: Every @attr{uuid} links to full documentation

---

**This architecture makes DSL generation intelligent, execution safe, and data provenance complete.**
