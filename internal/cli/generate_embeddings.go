package cli

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"dsl-ob-poc/internal/datastore"
	"dsl-ob-poc/internal/dictionary"
)

// RunGenerateEmbeddings generates vector embeddings for all dictionary attributes
func RunGenerateEmbeddings(ctx context.Context, ds datastore.DataStore, args []string) error {
	fs := flag.NewFlagSet("generate-embeddings", flag.ExitOnError)
	model := fs.String("model", "text-embedding-3-small", "OpenAI embedding model to use")
	domain := fs.String("domain", "", "Generate embeddings only for specific domain (ONBOARDING, KYC, etc.)")
	dryRun := fs.Bool("dry-run", false, "Show what would be done without actually generating embeddings")
	batchSize := fs.Int("batch-size", 10, "Number of attributes to process before progress update")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	// Check for OpenAI API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	log.Println("🔧 Starting embedding generation...")
	log.Printf("📊 Model: %s", *model)

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

	// Count attributes without embeddings
	missingEmbeddings := 0
	for _, attr := range attributes {
		if attr.Vector == "" {
			missingEmbeddings++
		}
	}

	log.Printf("🎯 Attributes without embeddings: %d", missingEmbeddings)
	log.Printf("✅ Attributes with embeddings: %d", len(attributes)-missingEmbeddings)

	if missingEmbeddings == 0 {
		log.Println("✨ All attributes already have embeddings!")
		return nil
	}

	if *dryRun {
		log.Println("🏃 Dry run mode - not generating embeddings")
		log.Println("Attributes that would be processed:")
		for _, attr := range attributes {
			if attr.Vector == "" {
				fmt.Printf("  - %s (%s)\n", attr.Name, attr.Domain)
			}
		}
		return nil
	}

	// Create embedding provider
	provider := dictionary.NewOpenAIEmbeddingProvider(apiKey, *model)

	// Generate embeddings
	log.Printf("🤖 Generating embeddings for %d attributes...", missingEmbeddings)
	log.Println("⏱️  This may take a few minutes depending on API rate limits...")

	startTime := time.Now()
	processed := 0
	updated := 0
	skipped := 0

	for i, attr := range attributes {
		// Skip if already has embedding
		if attr.Vector != "" {
			skipped++
			continue
		}

		// Generate embedding from long description
		text := attr.LongDescription
		if text == "" {
			text = attr.Name // Fallback to name if no description
		}

		log.Printf("🔄 [%d/%d] Processing: %s", i+1, len(attributes), attr.Name)

		embedding, err := provider.GenerateEmbedding(ctx, text)
		if err != nil {
			log.Printf("❌ Failed to generate embedding for %s: %v", attr.Name, err)
			continue
		}

		// Update attribute in database
		err = ds.UpdateAttributeVector(ctx, attr.AttributeID, embedding)
		if err != nil {
			log.Printf("❌ Failed to update vector for %s: %v", attr.Name, err)
			continue
		}

		processed++
		updated++

		// Progress update
		if processed%*batchSize == 0 {
			elapsed := time.Since(startTime)
			rate := float64(processed) / elapsed.Seconds()
			remaining := missingEmbeddings - processed
			eta := time.Duration(float64(remaining)/rate) * time.Second

			log.Printf("📊 Progress: %d/%d processed (%.1f%%) - ETA: %v",
				processed, missingEmbeddings,
				float64(processed)/float64(missingEmbeddings)*100,
				eta.Round(time.Second))
		}

		// Small delay to respect rate limits
		time.Sleep(100 * time.Millisecond)
	}

	elapsed := time.Since(startTime)

	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════════════")
	fmt.Println("🎉 Embedding Generation Complete!")
	fmt.Println("═══════════════════════════════════════════════════════════════════════")
	fmt.Printf("✅ Successfully updated: %d attributes\n", updated)
	fmt.Printf("⏭️  Skipped (already had embeddings): %d attributes\n", skipped)
	fmt.Printf("⏱️  Total time: %v\n", elapsed.Round(time.Second))
	fmt.Printf("⚡ Average rate: %.2f embeddings/second\n", float64(processed)/elapsed.Seconds())
	fmt.Println("═══════════════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("💡 Next steps:")
	fmt.Println("   Try semantic search:")
	fmt.Println("   ./dsl-poc semantic-search --query=\"What tracks someone's wealth?\"")
	fmt.Println("   ./dsl-poc semantic-search --query=\"identity documents for individuals\"")
	fmt.Println("   ./dsl-poc semantic-search --query=\"corporate entity information\"")
	fmt.Println()

	return nil
}
