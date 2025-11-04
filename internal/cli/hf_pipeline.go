package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"sort"
	"strings"
	"time"

	"dsl-ob-poc/internal/datastore"
	"dsl-ob-poc/internal/hf-investor/domain"
	"dsl-ob-poc/internal/hf-investor/store"
)

// RunHFPipeline handles the 'hf-pipeline' command for pipeline funnel analysis
func RunHFPipeline(ctx context.Context, ds datastore.DataStore, args []string) error {
	fs := flag.NewFlagSet("hf-pipeline", flag.ExitOnError)

	investorType := fs.String("type", "", "Filter by investor type (INDIVIDUAL, CORPORATE, TRUST, etc.)")
	domicile := fs.String("domicile", "", "Filter by domicile jurisdiction")
	source := fs.String("source", "", "Filter by investor source")
	dateFrom := fs.String("date-from", "", "Filter investors created from date (YYYY-MM-DD)")
	dateTo := fs.String("date-to", "", "Filter investors created to date (YYYY-MM-DD)")
	output := fs.String("output", "table", "Output format: table, json, csv")
	showPercentages := fs.Bool("percentages", false, "Show percentages in addition to counts")
	sortBy := fs.String("sort", "status", "Sort by: status, count")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	// Get hedge fund store
	hfStore, ok := ds.(store.HedgeFundInvestorStore)
	if !ok {
		return fmt.Errorf("datastore does not implement HedgeFundInvestorStore interface")
	}

	// Build filters
	filters := &store.PipelineFilters{}

	if *investorType != "" {
		if !domain.IsValidInvestorType(*investorType) {
			return fmt.Errorf("invalid investor type: %s", *investorType)
		}
		filters.InvestorType = investorType
	}

	if *domicile != "" {
		filters.Domicile = domicile
	}

	if *source != "" {
		filters.Source = source
	}

	if *dateFrom != "" {
		date, err := time.Parse("2006-01-02", *dateFrom)
		if err != nil {
			return fmt.Errorf("invalid date-from format. Use YYYY-MM-DD: %w", err)
		}
		filters.DateFrom = &date
	}

	if *dateTo != "" {
		date, err := time.Parse("2006-01-02", *dateTo)
		if err != nil {
			return fmt.Errorf("invalid date-to format. Use YYYY-MM-DD: %w", err)
		}
		filters.DateTo = &date
	}

	// Execute query
	pipeline, err := hfStore.GetPipelineFunnel(ctx, filters)
	if err != nil {
		return fmt.Errorf("failed to get pipeline funnel: %w", err)
	}

	// Handle empty results
	if len(pipeline) == 0 {
		fmt.Println("No investors found with the specified filters.")
		return nil
	}

	// Sort results
	switch *sortBy {
	case "count":
		sort.Slice(pipeline, func(i, j int) bool {
			return pipeline[i].Investors > pipeline[j].Investors
		})
	default: // status
		sort.Slice(pipeline, func(i, j int) bool {
			return pipeline[i].Status < pipeline[j].Status
		})
	}

	// Output results
	switch *output {
	case "json":
		return outputPipelineJSON(pipeline, *showPercentages)
	case "csv":
		return outputPipelineCSV(pipeline, *showPercentages)
	default:
		return outputPipelineTable(pipeline, *showPercentages, filters)
	}
}

func outputPipelineJSON(pipeline []store.PipelineStatusCount, showPercentages bool) error {
	total := calculateTotal(pipeline)

	var result interface{}
	if showPercentages {
		type PipelineWithPercentage struct {
			Status     string  `json:"status"`
			Investors  int     `json:"investors"`
			Percentage float64 `json:"percentage"`
		}

		withPercentages := make([]PipelineWithPercentage, len(pipeline))
		for i, p := range pipeline {
			percentage := 0.0
			if total > 0 {
				percentage = float64(p.Investors) / float64(total) * 100
			}
			withPercentages[i] = PipelineWithPercentage{
				Status:     p.Status,
				Investors:  p.Investors,
				Percentage: percentage,
			}
		}
		result = map[string]interface{}{
			"pipeline": withPercentages,
			"total":    total,
		}
	} else {
		result = map[string]interface{}{
			"pipeline": pipeline,
			"total":    total,
		}
	}

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonBytes))
	return nil
}

func outputPipelineCSV(pipeline []store.PipelineStatusCount, showPercentages bool) error {
	total := calculateTotal(pipeline)

	if showPercentages {
		fmt.Println("status,investors,percentage")
		for _, p := range pipeline {
			percentage := 0.0
			if total > 0 {
				percentage = float64(p.Investors) / float64(total) * 100
			}
			fmt.Printf("%s,%d,%.2f\n", p.Status, p.Investors, percentage)
		}
	} else {
		fmt.Println("status,investors")
		for _, p := range pipeline {
			fmt.Printf("%s,%d\n", p.Status, p.Investors)
		}
	}
	return nil
}

func outputPipelineTable(pipeline []store.PipelineStatusCount, showPercentages bool, filters *store.PipelineFilters) error {
	total := calculateTotal(pipeline)

	fmt.Println("\n=== Hedge Fund Investor Pipeline Funnel ===")

	// Show active filters
	var activeFilters []string
	if filters != nil {
		if filters.InvestorType != nil {
			activeFilters = append(activeFilters, fmt.Sprintf("Type: %s", *filters.InvestorType))
		}
		if filters.Domicile != nil {
			activeFilters = append(activeFilters, fmt.Sprintf("Domicile: %s", *filters.Domicile))
		}
		if filters.Source != nil {
			activeFilters = append(activeFilters, fmt.Sprintf("Source: %s", *filters.Source))
		}
		if filters.DateFrom != nil {
			activeFilters = append(activeFilters, fmt.Sprintf("From: %s", filters.DateFrom.Format("2006-01-02")))
		}
		if filters.DateTo != nil {
			activeFilters = append(activeFilters, fmt.Sprintf("To: %s", filters.DateTo.Format("2006-01-02")))
		}
	}

	if len(activeFilters) > 0 {
		fmt.Printf("Filters: %s\n", strings.Join(activeFilters, ", "))
	}
	fmt.Println()

	if len(pipeline) == 0 {
		fmt.Println("No investors found.")
		return nil
	}

	// Calculate column widths
	const (
		statusWidth    = 20
		investorsWidth = 12
		percentWidth   = 12
		barWidth       = 40
	)

	// Header
	if showPercentages {
		fmt.Printf("%-*s %*s %*s %s\n",
			statusWidth, "STATUS",
			investorsWidth, "INVESTORS",
			percentWidth, "PERCENTAGE",
			"DISTRIBUTION")
	} else {
		fmt.Printf("%-*s %*s %s\n",
			statusWidth, "STATUS",
			investorsWidth, "INVESTORS",
			"DISTRIBUTION")
	}

	// Separator
	if showPercentages {
		fmt.Printf("%s %s %s %s\n",
			strings.Repeat("-", statusWidth),
			strings.Repeat("-", investorsWidth),
			strings.Repeat("-", percentWidth),
			strings.Repeat("-", barWidth))
	} else {
		fmt.Printf("%s %s %s\n",
			strings.Repeat("-", statusWidth),
			strings.Repeat("-", investorsWidth),
			strings.Repeat("-", barWidth))
	}

	// Data rows with visual bars
	maxCount := getMaxCount(pipeline)
	for _, p := range pipeline {
		percentage := 0.0
		if total > 0 {
			percentage = float64(p.Investors) / float64(total) * 100
		}

		// Create visual bar
		barLength := int(float64(barWidth) * float64(p.Investors) / float64(maxCount))
		if barLength < 1 && p.Investors > 0 {
			barLength = 1
		}
		bar := strings.Repeat("█", barLength)

		if showPercentages {
			fmt.Printf("%-*s %*d %*s %s\n",
				statusWidth, p.Status,
				investorsWidth, p.Investors,
				percentWidth, fmt.Sprintf("%.1f%%", percentage),
				bar)
		} else {
			fmt.Printf("%-*s %*d %s\n",
				statusWidth, p.Status,
				investorsWidth, p.Investors,
				bar)
		}
	}

	// Summary
	if showPercentages {
		fmt.Printf("\n%s %s %s %s\n",
			strings.Repeat("-", statusWidth),
			strings.Repeat("-", investorsWidth),
			strings.Repeat("-", percentWidth),
			strings.Repeat("-", barWidth))
		fmt.Printf("%-*s %*d %*s\n",
			statusWidth, "TOTAL",
			investorsWidth, total,
			percentWidth, "100.0%")
	} else {
		fmt.Printf("\n%s %s %s\n",
			strings.Repeat("-", statusWidth),
			strings.Repeat("-", investorsWidth),
			strings.Repeat("-", barWidth))
		fmt.Printf("%-*s %*d\n",
			statusWidth, "TOTAL",
			investorsWidth, total)
	}

	fmt.Printf("\nTotal Statuses: %d\n", len(pipeline))
	return nil
}

func calculateTotal(pipeline []store.PipelineStatusCount) int {
	total := 0
	for _, p := range pipeline {
		total += p.Investors
	}
	return total
}

func getMaxCount(pipeline []store.PipelineStatusCount) int {
	max := 0
	for _, p := range pipeline {
		if p.Investors > max {
			max = p.Investors
		}
	}
	return max
}
