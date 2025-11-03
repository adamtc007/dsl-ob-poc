# 🤖 AI Agent Integration Analysis

## Overview

This document details the AI Agent integration architecture, showing how different domains (KYC vs DSL Onboarding) utilize the Gemini API with different inputs, prompts, and outputs.

## 🏗️ Agent Architecture

### Core Components
- **Agent Client**: Google Gemini 2.5 Flash model integration
- **System Prompting**: Domain-specific instructions with strict response formatting
- **Structured Responses**: JSON-based outputs with strong typing
- **Context Integration**: Awareness of DSL state and onboarding progression

## 🔍 KYC Domain Analysis

### Input Data Structure
```go
type KYCInput struct {
    NaturePurpose string   // "UCITS equity fund domiciled in LU"
    Products      []string // ["CUSTODY", "FUND_ACCOUNTING"]
}
```

### System Prompt Strategy
```
You are an expert KYC/AML Compliance Officer for a major global bank.
Your job is to analyze a new client's "nature and purpose" and their "requested products"
to determine the *minimum* required KYC documents and all relevant jurisdictions.

RULES:
1. Analyze the "nature and purpose" for entity type and domicile
2. Analyze the products for regulatory impact
3. Respond ONLY with a single, minified JSON object
4. The JSON format MUST be: {"required_documents": [...], "jurisdictions": [...]}
```

### Example Request/Response Flow

**Input:**
- Nature: "UCITS equity fund domiciled in LU"
- Products: ["CUSTODY", "FUND_ACCOUNTING"]

**AI Analysis:**
- Entity Type: UCITS fund (regulated EU investment vehicle)
- Domicile: Luxembourg (LU)
- Products: CUSTODY (requires custody agreement), FUND_ACCOUNTING (requires accounting policy)

**Output:**
```json
{
  "required_documents": [
    "CertificateOfIncorporation",
    "ArticlesOfAssociation",
    "W8BEN-E",
    "CustodyAgreement",
    "AccountingPolicy"
  ],
  "jurisdictions": ["LU"]
}
```

### KYC Context Variations

| Entity Type | Example Input | Key Documents | Jurisdictions |
|-------------|---------------|---------------|---------------|
| UCITS Fund | "UCITS equity fund domiciled in LU" | Certificate, Articles, W8BEN-E | LU |
| US Hedge Fund | "US-based hedge fund" | Partnership Agreement, W9, AML Policy | US |
| Delaware Corp | "Delaware corporation" | Certificate, Articles | US |

## 🔄 DSL Transformation Domain Analysis

### Input Data Structure
```go
type DSLTransformationRequest struct {
    CurrentDSL    string                 // Full S-expression DSL
    Instruction   string                 // Natural language instruction
    TargetState   string                 // Desired onboarding state
    Context       map[string]interface{} // Additional context data
}
```

### System Prompt Strategy
```
You are an expert DSL (Domain Specific Language) architect for financial onboarding workflows.
Your role is to analyze existing DSL and transform it according to user instructions
while maintaining correctness and consistency.

RULES:
1. Analyze the current DSL structure and understand its semantic meaning
2. Apply the requested transformation while preserving DSL syntax
3. Ensure all changes are consistent with the target onboarding state
4. Provide clear explanations for all changes made
5. Respond ONLY with a single, well-formed JSON object

DSL SYNTAX GUIDE:
- S-expressions format: (command args...)
- Case creation: (case.create (cbu.id "ID") (nature-purpose "DESC"))
- Products: (products.add "PRODUCT1" "PRODUCT2")
- KYC: (kyc.start (documents (document "DOC")) (jurisdictions (jurisdiction "JUR")))
```

### Example Request/Response Flow

**Input:**
```json
{
  "current_dsl": "(case.create\n  (cbu.id \"CBU-1234\")\n  (nature-purpose \"UCITS equity fund\")\n)\n\n(products.add \"CUSTODY\")",
  "instruction": "Add FUND_ACCOUNTING to the products list",
  "target_state": "PRODUCTS_ADDED",
  "context": {
    "current_state": "PRODUCTS_ADDED",
    "version_number": 2,
    "onboarding_id": "session-123"
  }
}
```

**AI Analysis:**
- Parses existing S-expression DSL structure
- Identifies products.add block with single product "CUSTODY"
- Understands instruction to add "FUND_ACCOUNTING"
- Maintains DSL syntax and structure

**Output:**
```json
{
  "new_dsl": "(case.create\n  (cbu.id \"CBU-1234\")\n  (nature-purpose \"UCITS equity fund\")\n)\n\n(products.add \"CUSTODY\" \"FUND_ACCOUNTING\")",
  "explanation": "Added FUND_ACCOUNTING to the products list as requested",
  "changes": [
    "Modified products.add block to include FUND_ACCOUNTING",
    "Maintained existing CUSTODY product"
  ],
  "confidence": 0.95
}
```

### DSL Transformation Capabilities

| Instruction Type | Example | AI Action | Confidence |
|------------------|---------|-----------|------------|
| Add Product | "Add TRANSFER_AGENT to products" | Modifies (products.add) block | High (0.9+) |
| Add Document | "Add W8BEN-E to KYC documents" | Modifies/creates (documents) block | High (0.9+) |
| Change Nature | "Update to hedge fund" | Modifies (nature-purpose) field | High (0.95+) |
| Add Jurisdiction | "Add LU jurisdiction" | Modifies/creates (jurisdictions) block | Medium (0.85+) |

## ✅ DSL Validation Domain Analysis

### Input Data Structure
```go
type DSLValidationInput struct {
    DSLText string // Complete DSL to validate
}
```

### System Prompt Strategy
```
You are an expert DSL validator for financial onboarding workflows.
Your role is to analyze DSL for correctness, completeness, and best practices.

VALIDATION CRITERIA:
1. Syntax correctness (proper S-expression structure)
2. Semantic correctness (logical flow and consistency)
3. Completeness (required elements for the onboarding state)
4. Best practices (proper naming, structure, etc.)
5. Compliance considerations (regulatory requirements)
```

### Example Request/Response Flow

**Input:**
```
(case.create
  (cbu.id "CBU-1234")
  (nature-purpose "UCITS equity fund domiciled in LU")
)

(products.add "CUSTODY" "FUND_ACCOUNTING")

(services.discover
  (for.product "CUSTODY"
    (service "CustodyService")
  )
)
```

**AI Analysis:**
- Validates S-expression syntax ✅
- Checks required elements (case.create, cbu.id) ✅
- Identifies missing KYC requirements ⚠️
- Suggests improvements 💡

**Output:**
```json
{
  "is_valid": true,
  "validation_score": 0.80,
  "errors": [],
  "warnings": ["No KYC requirements defined"],
  "suggestions": [
    "Consider running 'discover-kyc' to generate KYC requirements",
    "Ensure nature-purpose accurately reflects the entity type and domicile"
  ],
  "summary": "DSL is valid but could be improved with additional onboarding steps"
}
```

## 🎯 Context Customization & RAG Integration

### Request Customization Options

#### 1. KYC Domain Customization
```go
// Custom business context
context := map[string]interface{}{
    "regulatory_framework": "MiFID II",
    "risk_profile": "High",
    "client_segment": "Professional",
    "existing_relationships": ["Parent Company XYZ"],
}

// Custom instruction examples
instructions := []string{
    "Focus on EU regulatory requirements",
    "Include enhanced due diligence for high-risk",
    "Consider existing client relationships",
    "Apply professional client standards",
}
```

#### 2. DSL Transformation Customization
```go
// Rich context for transformation decisions
context := map[string]interface{}{
    "current_state": "SERVICES_DISCOVERED",
    "version_number": 3,
    "onboarding_id": "session-abc-123",
    "compliance_requirements": ["GDPR", "MiFID II"],
    "business_rules": {
        "max_products": 5,
        "required_jurisdictions": ["LU", "DE"],
    },
    "user_preferences": {
        "auto_validate": true,
        "verbose_explanations": true,
    }
}
```

### RAG Integration Opportunities

#### 1. Document Context Enhancement
```go
// Add regulatory document context
ragContext := map[string]interface{}{
    "relevant_regulations": [
        {
            "name": "MiFID II Article 25",
            "text": "Investment firms shall obtain information...",
            "relevance_score": 0.95
        }
    ],
    "compliance_precedents": [
        {
            "case": "Similar UCITS fund CBU-5678",
            "documents_required": ["W8BEN-E", "FATCA"],
            "outcome": "Approved"
        }
    ]
}
```

#### 2. Best Practices Context
```go
// Add institutional knowledge
bestPractices := map[string]interface{}{
    "industry_standards": {
        "UCITS_funds": {
            "typical_documents": ["KIID", "Prospectus", "W8BEN-E"],
            "common_jurisdictions": ["LU", "IE", "DE"],
            "processing_time": "5-10 business days"
        }
    },
    "internal_policies": {
        "document_retention": "7 years",
        "review_frequency": "Annual",
        "escalation_criteria": ["AML flags", "Sanctions screening"]
    }
}
```

## 🚀 Advanced Agent Capabilities

### 1. Multi-Step Reasoning
The AI can perform complex multi-step analysis:
- Parse business context → Identify entity type → Map to regulations → Generate requirements
- Analyze current DSL → Understand state → Apply transformation → Validate result

### 2. Contextual Awareness
Agents understand:
- Current onboarding state and progression
- Business context and regulatory environment
- DSL syntax and semantic meaning
- Historical patterns and best practices

### 3. Intelligent Suggestions
Beyond basic transformations:
- Proactive compliance recommendations
- State transition guidance
- Optimization suggestions
- Risk mitigation advice

## 🔧 Integration Patterns

### 1. Workflow Automation
```go
// Automated onboarding progression
workflow := []Step{
    {Agent: "KYC", Action: "discover_requirements"},
    {Agent: "DSL", Action: "transform_add_kyc"},
    {Agent: "Validation", Action: "validate_completeness"},
    {Agent: "DSL", Action: "transform_next_state"},
}
```

### 2. Human-in-the-Loop
```go
// AI assistance with human oversight
review := AIReview{
    Confidence: 0.85,
    RequiresHumanReview: confidence < 0.9,
    Suggestions: agent.GetSuggestions(),
    ApprovalRequired: true,
}
```

### 3. Continuous Learning
```go
// Feedback integration for model improvement
feedback := AgentFeedback{
    RequestID: "req-123",
    UserApproved: true,
    Modifications: []string{"Added additional document"},
    OutcomeRating: 4.5,
}
```

This architecture provides a robust foundation for AI-powered onboarding automation while maintaining human oversight and regulatory compliance.