package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"time"

	"dsl-ob-poc/internal/datastore"
	"dsl-ob-poc/internal/hf-investor/store"

	"github.com/google/uuid"
)

func outputPositionsJSON(positions []store.PositionAsOf, showZero bool) error {
	filtered := filterZeroPositions(positions, showZero)
	data := map[string]interface{}{
		"positions": filtered,
		"count":     len(filtered),
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonBytes))
	return nil
}

func outputPositionsCSV(positions []store.PositionAsOf, showZero bool) error {
	filtered := filterZeroPositions(positions, showZero)

	fmt.Println("investor_id,fund_id,class_id,series_id,units")
	for _, pos := range filtered {
		seriesID := ""
		if pos.SeriesID != nil {
			seriesID = pos.SeriesID.String()
		}
		fmt.Printf("%s,%s,%s,%s,%.8f\n",
			pos.InvestorID.String(),
			pos.FundID.String(),
			pos.ClassID.String(),
			seriesID,
			pos.Units,
		)
	}
	return nil
}

func outputPositionsTable(positions []store.PositionAsOf, asOfDate time.Time, showZero bool) error {
	filtered := filterZeroPositions(positions, showZero)

	fmt.Printf("\n=== Hedge Fund Positions as of %s ===\n\n", asOfDate.Format("2006-01-02"))

	if len(filtered) == 0 {
		fmt.Println("No non-zero positions found.")
		return nil
	}

	// Calculate column widths
	const (
		investorWidth = 36
		fundWidth     = 36
		classWidth    = 36
		seriesWidth   = 36
		unitsWidth    = 15
	)

	// Header
	fmt.Printf("%-*s %-*s %-*s %-*s %*s\n",
		investorWidth, "INVESTOR_ID",
		fundWidth, "FUND_ID",
		classWidth, "CLASS_ID",
		seriesWidth, "SERIES_ID",
		unitsWidth, "UNITS")

	fmt.Printf("%s %s %s %s %s\n",
		repeatChar(investorWidth),
		repeatChar(fundWidth),
		repeatChar(classWidth),
		repeatChar(seriesWidth),
		repeatChar(unitsWidth))

	// Data rows
	totalUnits := 0.0
	for _, pos := range filtered {
		seriesID := ""
		if pos.SeriesID != nil {
			seriesID = pos.SeriesID.String()
		}

		fmt.Printf("%-*s %-*s %-*s %-*s %*.6f\n",
			investorWidth, pos.InvestorID.String(),
			fundWidth, pos.FundID.String(),
			classWidth, pos.ClassID.String(),
			seriesWidth, seriesID,
			unitsWidth, pos.Units)

		totalUnits += pos.Units
	}

	// Summary
	fmt.Printf("\n%s %s %s %s %s\n",
		repeatChar(investorWidth),
		repeatChar(fundWidth),
		repeatChar(classWidth),
		repeatChar(seriesWidth),
		repeatChar(unitsWidth))

	fmt.Printf("%-*s %-*s %-*s %-*s %*.6f\n",
		investorWidth, "TOTAL",
		fundWidth, "",
		classWidth, "",
		seriesWidth, "",
		unitsWidth, totalUnits)

	fmt.Printf("\nTotal Positions: %d\n", len(filtered))
	return nil
}

func filterZeroPositions(positions []store.PositionAsOf, showZero bool) []store.PositionAsOf {
	if showZero {
		return positions
	}

	filtered := make([]store.PositionAsOf, 0, len(positions))
	for _, pos := range positions {
		if pos.Units != 0 {
			filtered = append(filtered, pos)
		}
	}
	return filtered
}

func repeatChar(count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += "-"
	}
	return result
}

// RunHFPositions handles the 'hf-positions' command for position as-of queries
func RunHFPositions(ctx context.Context, ds datastore.DataStore, args []string) error {
	fs := flag.NewFlagSet("hf-positions", flag.ExitOnError)

	asOfDate := fs.String("as-of", "", "Date for position query (YYYY-MM-DD format) (required)")
	investorID := fs.String("investor-id", "", "Filter by investor ID (UUID)")
	fundID := fs.String("fund-id", "", "Filter by fund ID (UUID)")
	classID := fs.String("class-id", "", "Filter by share class ID (UUID)")
	seriesID := fs.String("series-id", "", "Filter by series ID (UUID)")
	minUnits := fs.Float64("min-units", 0, "Minimum units threshold")
	output := fs.String("output", "table", "Output format: table, json, csv")
	showZero := fs.Bool("show-zero", false, "Include zero unit positions in output")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	// Validate required fields
	if *asOfDate == "" {
		return fmt.Errorf("--as-of is required (format: YYYY-MM-DD)")
	}

	// Parse as-of date
	parsedAsOfDate, err := time.Parse("2006-01-02", *asOfDate)
	if err != nil {
		return fmt.Errorf("invalid as-of date format. Use YYYY-MM-DD: %w", err)
	}

	// Get hedge fund store
	hfStore, ok := ds.(store.HedgeFundInvestorStore)
	if !ok {
		return fmt.Errorf("datastore does not implement HedgeFundInvestorStore interface")
	}

	// Build filters
	filters := &store.PositionFilters{}

	if *investorID != "" {
		invID, parseErr := uuid.Parse(*investorID)
		if parseErr != nil {
			return fmt.Errorf("invalid investor-id format: %w", parseErr)
		}
		filters.InvestorID = &invID
	}

	if *fundID != "" {
		fID, parseErr := uuid.Parse(*fundID)
		if parseErr != nil {
			return fmt.Errorf("invalid fund-id format: %w", parseErr)
		}
		filters.FundID = &fID
	}

	if *classID != "" {
		cID, parseErr := uuid.Parse(*classID)
		if parseErr != nil {
			return fmt.Errorf("invalid class-id format: %w", parseErr)
		}
		filters.ClassID = &cID
	}

	if *seriesID != "" {
		sID, parseErr := uuid.Parse(*seriesID)
		if parseErr != nil {
			return fmt.Errorf("invalid series-id format: %w", parseErr)
		}
		filters.SeriesID = &sID
	}

	if *minUnits > 0 {
		filters.MinUnits = minUnits
	}

	// Execute query
	positions, err := hfStore.GetPositionsAsOf(ctx, parsedAsOfDate, filters)
	if err != nil {
		return fmt.Errorf("failed to get positions as of %s: %w", parsedAsOfDate.Format("2006-01-02"), err)
	}

	// Handle empty results
	if len(positions) == 0 {
		fmt.Printf("No positions found as of %s with the specified filters.\n", parsedAsOfDate.Format("2006-01-02"))
		return nil
	}

	// Output results
	switch *output {
	case "json":
		return outputPositionsJSON(positions, *showZero)
	case "csv":
		return outputPositionsCSV(positions, *showZero)
	default:
		return outputPositionsTable(positions, parsedAsOfDate, *showZero)
	}
}
