package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/google/generative-ai-go/genai"
)

// DSLTransformationRequest represents a request to transform DSL
type DSLTransformationRequest struct {
	CurrentDSL  string                 `json:"current_dsl"`
	Instruction string                 `json:"instruction"`
	TargetState string                 `json:"target_state"`
	Context     map[string]interface{} `json:"context,omitempty"`
}

// DSLTransformationResponse represents the AI agent's response
type DSLTransformationResponse struct {
	NewDSL      string   `json:"new_dsl"`
	Explanation string   `json:"explanation"`
	Changes     []string `json:"changes"`
	Confidence  float64  `json:"confidence"`
}

// CallDSLTransformationAgent handles general DSL transformations using AI
func (a *Agent) CallDSLTransformationAgent(ctx context.Context, request DSLTransformationRequest) (*DSLTransformationResponse, error) {
	if a == nil || a.model == nil {
		return nil, fmt.Errorf("ai agent is not initialized")
	}

	systemPrompt := `You are an expert DSL (Domain Specific Language) architect for financial onboarding workflows.
Your role is to analyze existing DSL and transform it according to user instructions while maintaining correctness and consistency.

RULES:
1. Analyze the current DSL structure and understand its semantic meaning
2. Apply the requested transformation while preserving DSL syntax and structure
3. Ensure all changes are consistent with the target onboarding state
4. Provide clear explanations for all changes made
5. Respond ONLY with a single, well-formed JSON object
6. Do not include markdown, code blocks, or conversational text

DSL SYNTAX GUIDE:
- S-expressions format: (command args...)
- Case creation: (case.create (cbu.id "ID") (nature-purpose "DESC"))
- Products: (products.add "PRODUCT1" "PRODUCT2")
- KYC: (kyc.start (documents (document "DOC")) (jurisdictions (jurisdiction "JUR")))
- Services: (services.discover (for.product "PROD" (service "SVC")))
- Resources: (resources.plan (resource.create "NAME" (owner "OWNER") (var (attr-id "ID"))))
- Values: (values.bind (bind (attr-id "ID") (value "VAL")))

RESPONSE FORMAT:
{
  "new_dsl": "Complete transformed DSL as a string",
  "explanation": "Clear explanation of what was changed and why",
  "changes": ["List of specific changes made"],
  "confidence": 0.95
}

EXAMPLES:
- Adding a product: Transform (products.add "CUSTODY") to (products.add "CUSTODY" "FUND_ACCOUNTING")
- Updating jurisdiction: Change (jurisdiction "US") to (jurisdiction "LU")
- Adding KYC document: Add (document "W8BEN-E") to existing documents list`

	// Format the user prompt with the transformation request
	userPrompt := fmt.Sprintf(`Current DSL:
%s

Instruction: %s
Target State: %s

Additional Context: %s

Please transform the DSL according to the instruction while moving toward the target state.`,
		request.CurrentDSL,
		request.Instruction,
		request.TargetState,
		jsonString(request.Context))

	a.model.SystemInstruction = &genai.Content{Parts: []genai.Part{genai.Text(systemPrompt)}}

	resp, err := a.model.GenerateContent(ctx, genai.Text(userPrompt))
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0] == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response from agent: %v", resp)
	}

	part := resp.Candidates[0].Content.Parts[0]
	textPart, ok := part.(genai.Text)
	if !ok {
		return nil, fmt.Errorf("unexpected response type from agent: %T", part)
	}

	log.Printf("DSL Agent Raw Response: %s", textPart)

	// Clean potential markdown-wrapped JSON using jsonv2's robust parsing
	cleanedJSON := cleanJSONResponse(string(textPart))

	var transformResp DSLTransformationResponse
	if uErr := json.Unmarshal([]byte(cleanedJSON), &transformResp); uErr != nil {
		return nil, fmt.Errorf("failed to parse agent's JSON response: %w (cleaned response was: %s)", uErr, cleanedJSON)
	}

	return &transformResp, nil
}

// CallDSLValidationAgent validates DSL correctness and suggests improvements
func (a *Agent) CallDSLValidationAgent(ctx context.Context, dslToValidate string) (*DSLValidationResponse, error) {
	if a == nil || a.model == nil {
		return nil, fmt.Errorf("ai agent is not initialized")
	}

	systemPrompt := `You are an expert DSL validator for financial onboarding workflows.
Your role is to analyze DSL for correctness, completeness, and best practices.

VALIDATION CRITERIA:
1. Syntax correctness (proper S-expression structure)
2. Semantic correctness (logical flow and consistency)
3. Completeness (required elements for the onboarding state)
4. Best practices (proper naming, structure, etc.)
5. Compliance considerations (regulatory requirements)

RESPONSE FORMAT:
{
  "is_valid": true/false,
  "validation_score": 0.95,
  "errors": ["List of syntax or semantic errors"],
  "warnings": ["List of potential issues"],
  "suggestions": ["List of improvement suggestions"],
  "summary": "Overall assessment of the DSL"
}`

	userPrompt := fmt.Sprintf(`Please validate the following DSL:

%s

Provide a comprehensive validation assessment including errors, warnings, and suggestions for improvement.`, dslToValidate)

	a.model.SystemInstruction = &genai.Content{Parts: []genai.Part{genai.Text(systemPrompt)}}

	resp, err := a.model.GenerateContent(ctx, genai.Text(userPrompt))
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0] == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response from agent: %v", resp)
	}

	part := resp.Candidates[0].Content.Parts[0]
	textPart, ok := part.(genai.Text)
	if !ok {
		return nil, fmt.Errorf("unexpected response type from agent: %T", part)
	}

	log.Printf("DSL Validation Agent Raw Response: %s", textPart)

	// Clean potential markdown-wrapped JSON using jsonv2's robust parsing
	cleanedJSON := cleanJSONResponse(string(textPart))

	var validationResp DSLValidationResponse
	if uErr := json.Unmarshal([]byte(cleanedJSON), &validationResp); uErr != nil {
		return nil, fmt.Errorf("failed to parse agent's JSON response: %w (cleaned response was: %s)", uErr, cleanedJSON)
	}

	return &validationResp, nil
}

// DSLValidationResponse represents validation results
type DSLValidationResponse struct {
	IsValid         bool     `json:"is_valid"`
	ValidationScore float64  `json:"validation_score"`
	Errors          []string `json:"errors"`
	Warnings        []string `json:"warnings"`
	Suggestions     []string `json:"suggestions"`
	Summary         string   `json:"summary"`
}

// Helper function to safely convert context to JSON string
func jsonString(v interface{}) string {
	if v == nil {
		return "{}"
	}

	data, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}

	return string(data)
}

// cleanJSONResponse removes markdown code block wrappers and cleans JSON
// Takes advantage of jsonv2's improved error handling and validation
func cleanJSONResponse(response string) string {
	// Trim whitespace
	cleaned := strings.TrimSpace(response)

	// Remove markdown JSON code blocks (```json ... ```)
	if strings.HasPrefix(cleaned, "```json") {
		// Find the first newline after ```json
		if firstNewline := strings.Index(cleaned, "\n"); firstNewline != -1 {
			cleaned = cleaned[firstNewline+1:]
		}
	}

	// Remove trailing ```
	cleaned = strings.TrimSuffix(cleaned, "```")

	// Remove any other markdown code block markers
	cleaned = strings.TrimPrefix(cleaned, "```")

	// Clean up any remaining whitespace
	cleaned = strings.TrimSpace(cleaned)

	// Validate that we have valid JSON using jsonv2's robust validation
	if err := json.Unmarshal([]byte(cleaned), new(interface{})); err == nil {
		return cleaned
	}

	// If still not valid, try to extract JSON from the response
	// Look for the first { and last } to extract potential JSON object
	firstBrace := strings.Index(cleaned, "{")
	lastBrace := strings.LastIndex(cleaned, "}")

	if firstBrace != -1 && lastBrace != -1 && lastBrace > firstBrace {
		extracted := cleaned[firstBrace : lastBrace+1]
		if err := json.Unmarshal([]byte(extracted), new(interface{})); err == nil {
			return extracted
		}
	}

	// Return original if we can't clean it
	return response
}
