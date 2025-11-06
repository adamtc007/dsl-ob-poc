package dictionary

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"
)

// EmbeddingProvider defines the interface for embedding generation
type EmbeddingProvider interface {
	GenerateEmbedding(ctx context.Context, text string) ([]float64, error)
}

// OpenAIEmbeddingProvider implements EmbeddingProvider using OpenAI API
type OpenAIEmbeddingProvider struct {
	APIKey     string
	Model      string // e.g., "text-embedding-3-small" or "text-embedding-3-large"
	HTTPClient *http.Client
}

// OpenAIEmbeddingRequest represents the API request structure
type OpenAIEmbeddingRequest struct {
	Input string `json:"input"`
	Model string `json:"model"`
}

// OpenAIEmbeddingResponse represents the API response structure
type OpenAIEmbeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Embedding []float64 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// NewOpenAIEmbeddingProvider creates a new OpenAI embedding provider
func NewOpenAIEmbeddingProvider(apiKey string, model string) *OpenAIEmbeddingProvider {
	if model == "" {
		model = "text-embedding-3-small" // Default to smaller, faster model
	}

	return &OpenAIEmbeddingProvider{
		APIKey: apiKey,
		Model:  model,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GenerateEmbedding generates an embedding vector for the given text using OpenAI API
func (p *OpenAIEmbeddingProvider) GenerateEmbedding(ctx context.Context, text string) ([]float64, error) {
	if p.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key not configured")
	}

	// Truncate text if too long (OpenAI has token limits)
	if len(text) > 8000 {
		text = text[:8000] + "..."
	}

	reqBody := OpenAIEmbeddingRequest{
		Input: text,
		Model: p.Model,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.APIKey)

	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call OpenAI API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API returned status %d: %s", resp.StatusCode, string(body))
	}

	var embeddingResp OpenAIEmbeddingResponse
	if err := json.Unmarshal(body, &embeddingResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(embeddingResp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data returned")
	}

	return embeddingResp.Data[0].Embedding, nil
}

// CosineSimilarity calculates the cosine similarity between two vectors
// Returns a value between -1 and 1, where 1 means identical, 0 means orthogonal, -1 means opposite
func CosineSimilarity(a, b []float64) (float64, error) {
	if len(a) != len(b) {
		return 0, fmt.Errorf("vectors must have same length: %d vs %d", len(a), len(b))
	}

	if len(a) == 0 {
		return 0, fmt.Errorf("vectors cannot be empty")
	}

	var dotProduct, normA, normB float64
	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	normA = math.Sqrt(normA)
	normB = math.Sqrt(normB)

	if normA == 0 || normB == 0 {
		return 0, fmt.Errorf("cannot compute similarity with zero vector")
	}

	return dotProduct / (normA * normB), nil
}

// AttributeMatch represents a matched attribute with similarity score
type AttributeMatch struct {
	Attribute  Attribute
	Similarity float64
}

// SemanticSearch performs semantic search over attributes using embedding vectors
// Returns attributes ranked by similarity to the query
func SemanticSearch(ctx context.Context, provider EmbeddingProvider, query string, attributes []Attribute, topK int) ([]AttributeMatch, error) {
	if topK <= 0 {
		topK = 10 // Default to top 10 results
	}

	// Generate embedding for the query
	queryEmbedding, err := provider.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Calculate similarity scores for all attributes
	matches := make([]AttributeMatch, 0, len(attributes))
	for _, attr := range attributes {
		// Parse the vector from JSON string
		if attr.Vector == "" {
			continue // Skip attributes without embeddings
		}

		var attrEmbedding []float64
		if err := json.Unmarshal([]byte(attr.Vector), &attrEmbedding); err != nil {
			// Skip attributes with invalid vector data
			continue
		}

		similarity, err := CosineSimilarity(queryEmbedding, attrEmbedding)
		if err != nil {
			// Skip attributes where similarity calculation fails
			continue
		}

		matches = append(matches, AttributeMatch{
			Attribute:  attr,
			Similarity: similarity,
		})
	}

	// Sort by similarity (descending)
	for i := 0; i < len(matches)-1; i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[j].Similarity > matches[i].Similarity {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}

	// Return top K results
	if len(matches) > topK {
		matches = matches[:topK]
	}

	return matches, nil
}

// GenerateAndStoreEmbeddings generates embeddings for attributes and stores them in the Vector field
func GenerateAndStoreEmbeddings(ctx context.Context, provider EmbeddingProvider, attributes []Attribute) ([]Attribute, error) {
	result := make([]Attribute, len(attributes))

	for i, attr := range attributes {
		// Generate embedding from long description
		text := attr.LongDescription
		if text == "" {
			text = attr.Name // Fallback to name if no description
		}

		embedding, err := provider.GenerateEmbedding(ctx, text)
		if err != nil {
			return nil, fmt.Errorf("failed to generate embedding for attribute %s: %w", attr.Name, err)
		}

		// Store embedding as JSON array in Vector field
		embeddingJSON, err := json.Marshal(embedding)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal embedding for attribute %s: %w", attr.Name, err)
		}

		attr.Vector = string(embeddingJSON)
		result[i] = attr
	}

	return result, nil
}
