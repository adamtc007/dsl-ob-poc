package cli

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"dsl-ob-poc/internal/datastore"
	"dsl-ob-poc/internal/dictionary"
)

// RunSemanticSearch performs semantic search over dictionary attributes
func RunSemanticSearch(ctx context.Context, ds datastore.DataStore, args []string) error {
	fs := flag.NewFlagSet("semantic-search", flag.ExitOnError)
	query := fs.String("query", "", "Natural language search query (required)")
	topK := fs.Int("top", 10, "Number of results to return")
	domain := fs.String("domain", "", "Filter by domain (ONBOARDING, KYC, UBO, etc.)")
	groupID := fs.String("group", "", "Filter by group_id")
	model := fs.String("model", "text-embedding-3-small", "OpenAI embedding model")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	if *query == "" {
		fs.Usage()
		return fmt.Errorf("error: --query flag is required")
	}

	// Check for OpenAI API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	log.Printf("🔍 Semantic Search Query: %q", *query)
	log.Printf("📊 Model: %s, Top K: %d", *model, *topK)

	// Get all attributes from database
	attributes, err := ds.GetAllAttributes(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch attributes: %w", err)
	}

	log.Printf("📚 Loaded %d attributes from dictionary", len(attributes))

	// Filter by domain if specified
	if *domain != "" {
		filtered := make([]dictionary.Attribute, 0)
		for _, attr := range attributes {
			if attr.Domain == *domain {
				filtered = append(filtered, attr)
			}
		}
		attributes = filtered
		log.Printf("🔎 Filtered to %d attributes in domain %s", len(attributes), *domain)
	}

	// Filter by group_id if specified
	if *groupID != "" {
		filtered := make([]dictionary.Attribute, 0)
		for _, attr := range attributes {
			if attr.GroupID == *groupID {
				filtered = append(filtered, attr)
			}
		}
		attributes = filtered
		log.Printf("🔎 Filtered to %d attributes in group %s", len(attributes), *groupID)
	}

	// Check how many attributes have embeddings
	embeddedCount := 0
	for _, attr := range attributes {
		if attr.Vector != "" {
			embeddedCount++
		}
	}

	if embeddedCount == 0 {
		return fmt.Errorf("no attributes have embeddings yet - run 'generate-embeddings' command first")
	}

	log.Printf("✅ Found %d attributes with embeddings", embeddedCount)

	// Create embedding provider
	provider := dictionary.NewOpenAIEmbeddingProvider(apiKey, *model)

	// Perform semantic search
	log.Println("🤖 Generating query embedding...")
	matches, err := dictionary.SemanticSearch(ctx, provider, *query, attributes, *topK)
	if err != nil {
		return fmt.Errorf("semantic search failed: %w", err)
	}

	// Display results
	fmt.Println()
	fmt.Println("╔═══════════════════════════════════════════════════════════════════════════════╗")
	fmt.Printf("║ %-77s ║\n", fmt.Sprintf("🎯 Top %d Results for: %s", len(matches), *query))
	fmt.Println("╚═══════════════════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	if len(matches) == 0 {
		fmt.Println("❌ No matching attributes found")
		return nil
	}

	for i, match := range matches {
		fmt.Printf("┌─ Result #%d ─────────────────────────────────────────────────────────────────┐\n", i+1)
		fmt.Printf("│ 📝 Name:       %-62s │\n", match.Attribute.Name)
		fmt.Printf("│ 🎲 Similarity: %-62.4f │\n", match.Similarity)
		fmt.Printf("│ 🏷️  Domain:     %-62s │\n", match.Attribute.Domain)
		fmt.Printf("│ 📦 Group:      %-62s │\n", match.Attribute.GroupID)
		fmt.Printf("│ 🔠 Mask:       %-62s │\n", match.Attribute.Mask)
		fmt.Printf("│ 📄 Description:                                                              │\n")

		// Word wrap description
		desc := match.Attribute.LongDescription
		if len(desc) > 500 {
			desc = desc[:500] + "..."
		}
		words := splitWords(desc, 70)
		for _, line := range words {
			fmt.Printf("│    %-74s │\n", line)
		}

		fmt.Printf("│ 🆔 ID:         %-62s │\n", match.Attribute.AttributeID)
		fmt.Println("└──────────────────────────────────────────────────────────────────────────────┘")
		fmt.Println()
	}

	// Show DSL usage example
	if len(matches) > 0 {
		fmt.Println("💡 DSL Usage Example:")
		fmt.Println("────────────────────────────────────────────────────────────────────────────────")
		fmt.Println("(kyc.collect")
		for i, match := range matches {
			if i >= 3 {
				break // Show first 3
			}
			fmt.Printf("  (attr (id \"%s\"))  ; %s\n", match.Attribute.AttributeID, match.Attribute.Name)
		}
		fmt.Println(")")
		fmt.Println("────────────────────────────────────────────────────────────────────────────────")
	}

	return nil
}

// splitWords splits text into lines of maximum width
func splitWords(text string, maxWidth int) []string {
	if text == "" {
		return []string{}
	}

	var lines []string
	var currentLine string

	words := splitBySpace(text)
	for _, word := range words {
		if len(currentLine)+len(word)+1 > maxWidth {
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			currentLine = word
		} else {
			if currentLine == "" {
				currentLine = word
			} else {
				currentLine += " " + word
			}
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

// splitBySpace splits string by spaces
func splitBySpace(s string) []string {
	var words []string
	var current string

	for _, ch := range s {
		if ch == ' ' || ch == '\n' || ch == '\t' {
			if current != "" {
				words = append(words, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}

	if current != "" {
		words = append(words, current)
	}

	return words
}
