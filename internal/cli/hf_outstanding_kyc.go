package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	"dsl-ob-poc/internal/datastore"
	"dsl-ob-poc/internal/hf-investor/store"

	"github.com/google/uuid"
)

// RunHFOutstandingKYC handles the 'hf-outstanding-kyc' command for outstanding KYC requirements
func RunHFOutstandingKYC(ctx context.Context, ds datastore.DataStore, args []string) error {
	fs := flag.NewFlagSet("hf-outstanding-kyc", flag.ExitOnError)

	investorID := fs.String("investor-id", "", "Filter by investor ID (UUID)")
	docType := fs.String("doc-type", "", "Filter by document type")
	status := fs.String("status", "", "Filter by status (REQUESTED, OVERDUE)")
	priority := fs.String("priority", "", "Filter by priority (LOW, MEDIUM, HIGH, CRITICAL)")
	overdueOnly := fs.Bool("overdue", false, "Show only overdue requirements")
	source := fs.String("source", "", "Filter by requirement source")
	output := fs.String("output", "table", "Output format: table, json, csv")
	sortBy := fs.String("sort", "due_date", "Sort by: due_date, priority, investor, doc_type")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	// Get hedge fund store
	hfStore, ok := ds.(store.HedgeFundInvestorStore)
	if !ok {
		return fmt.Errorf("datastore does not implement HedgeFundInvestorStore interface")
	}

	// Build filters
	filters := &store.KYCRequirementFilters{}

	if *investorID != "" {
		id, err := uuid.Parse(*investorID)
		if err != nil {
			return fmt.Errorf("invalid investor-id format: %w", err)
		}
		filters.InvestorID = &id
	}

	if *docType != "" {
		filters.DocType = docType
	}

	if *status != "" {
		filters.Status = status
	}

	if *priority != "" {
		filters.Priority = priority
	}

	if *overdueOnly {
		filters.OverdueOnly = overdueOnly
	}

	if *source != "" {
		filters.Source = source
	}

	// Execute query
	requirements, err := hfStore.GetOutstandingKYC(ctx, filters)
	if err != nil {
		return fmt.Errorf("failed to get outstanding KYC requirements: %w", err)
	}

	// Handle empty results
	if len(requirements) == 0 {
		fmt.Println("No outstanding KYC requirements found with the specified filters.")
		return nil
	}

	// Sort results
	switch *sortBy {
	case "priority":
		sortByPriority(requirements)
	case "investor":
		sortByInvestor(requirements)
	case "doc_type":
		sortByDocType(requirements)
	default: // due_date
		sortByDueDate(requirements)
	}

	// Output results
	switch *output {
	case "json":
		return outputKYCJSON(requirements)
	case "csv":
		return outputKYCCSV(requirements)
	default:
		return outputKYCTable(requirements, filters)
	}
}

func outputKYCJSON(requirements []store.KYCRequirement) error {
	data := map[string]interface{}{
		"outstanding_kyc": requirements,
		"count":           len(requirements),
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonBytes))
	return nil
}

func outputKYCCSV(requirements []store.KYCRequirement) error {
	fmt.Println("investor_id,investor_code,investor_name,doc_type,status,priority,requested_at,due_at,days_overdue,source")

	for i := range requirements {
		req := requirements[i]
		dueAt := ""
		if req.DueAt != nil {
			dueAt = req.DueAt.Format("2006-01-02")
		}

		daysOverdue := ""
		if req.DaysOverdue != nil {
			daysOverdue = fmt.Sprintf("%d", *req.DaysOverdue)
		}

		source := ""
		if req.Source != nil {
			source = *req.Source
		}

		fmt.Printf("%s,%s,%s,%s,%s,%s,%s,%s,%s,%s\n",
			req.InvestorID.String(),
			req.InvestorCode,
			req.InvestorName,
			req.DocType,
			req.Status,
			req.Priority,
			req.RequestedAt.Format("2006-01-02"),
			dueAt,
			daysOverdue,
			source,
		)
	}
	return nil
}

func outputKYCTable(requirements []store.KYCRequirement, filters *store.KYCRequirementFilters) error {
	fmt.Println("\n=== Outstanding KYC Requirements ===")

	// Show active filters
	var activeFilters []string
	if filters != nil {
		if filters.InvestorID != nil {
			activeFilters = append(activeFilters, fmt.Sprintf("Investor: %s", filters.InvestorID.String()[:8]))
		}
		if filters.DocType != nil {
			activeFilters = append(activeFilters, fmt.Sprintf("Doc Type: %s", *filters.DocType))
		}
		if filters.Status != nil {
			activeFilters = append(activeFilters, fmt.Sprintf("Status: %s", *filters.Status))
		}
		if filters.Priority != nil {
			activeFilters = append(activeFilters, fmt.Sprintf("Priority: %s", *filters.Priority))
		}
		if filters.OverdueOnly != nil && *filters.OverdueOnly {
			activeFilters = append(activeFilters, "Overdue Only")
		}
		if filters.Source != nil {
			activeFilters = append(activeFilters, fmt.Sprintf("Source: %s", *filters.Source))
		}
	}

	if len(activeFilters) > 0 {
		fmt.Printf("Filters: %s\n", strings.Join(activeFilters, ", "))
	}
	fmt.Println()

	if len(requirements) == 0 {
		fmt.Println("No outstanding requirements found.")
		return nil
	}

	// Calculate column widths
	const (
		investorWidth  = 12
		nameWidth      = 20
		docTypeWidth   = 16
		statusWidth    = 10
		priorityWidth  = 8
		requestedWidth = 10
		dueWidth       = 10
		overdueWidth   = 8
		sourceWidth    = 15
	)

	// Header
	fmt.Printf("%-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s\n",
		investorWidth, "INVESTOR",
		nameWidth, "NAME",
		docTypeWidth, "DOCUMENT_TYPE",
		statusWidth, "STATUS",
		priorityWidth, "PRIORITY",
		requestedWidth, "REQUESTED",
		dueWidth, "DUE_DATE",
		overdueWidth, "OVERDUE",
		sourceWidth, "SOURCE")

	// Separator
	fmt.Printf("%s %s %s %s %s %s %s %s %s\n",
		strings.Repeat("-", investorWidth),
		strings.Repeat("-", nameWidth),
		strings.Repeat("-", docTypeWidth),
		strings.Repeat("-", statusWidth),
		strings.Repeat("-", priorityWidth),
		strings.Repeat("-", requestedWidth),
		strings.Repeat("-", dueWidth),
		strings.Repeat("-", overdueWidth),
		strings.Repeat("-", sourceWidth))

	// Data rows
	overdueCount := 0
	for i := range requirements {
		req := requirements[i]
		// Truncate investor code and name for display
		investorCode := req.InvestorCode
		if len(investorCode) > investorWidth {
			investorCode = investorCode[:investorWidth-3] + "..."
		}

		investorName := req.InvestorName
		if len(investorName) > nameWidth {
			investorName = investorName[:nameWidth-3] + "..."
		}

		docType := req.DocType
		if len(docType) > docTypeWidth {
			docType = docType[:docTypeWidth-3] + "..."
		}

		source := ""
		if req.Source != nil {
			source = *req.Source
			if len(source) > sourceWidth {
				source = source[:sourceWidth-3] + "..."
			}
		}

		dueAt := ""
		if req.DueAt != nil {
			dueAt = req.DueAt.Format("2006-01-02")
		}

		daysOverdue := ""
		if req.DaysOverdue != nil {
			daysOverdue = fmt.Sprintf("%dd", *req.DaysOverdue)
			overdueCount++
		}

		// Color coding for priority and status (basic text indicators)
		priorityDisplay := req.Priority
		if req.Priority == "CRITICAL" {
			priorityDisplay = "⚠️ CRIT"
		} else if req.Priority == "HIGH" {
			priorityDisplay = "🔴 HIGH"
		}

		statusDisplay := req.Status
		if req.Status == "OVERDUE" {
			statusDisplay = "🚨 OVER"
		}

		fmt.Printf("%-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s\n",
			investorWidth, investorCode,
			nameWidth, investorName,
			docTypeWidth, docType,
			statusWidth, statusDisplay,
			priorityWidth, priorityDisplay,
			requestedWidth, req.RequestedAt.Format("2006-01-02"),
			dueWidth, dueAt,
			overdueWidth, daysOverdue,
			sourceWidth, source)
	}

	// Summary
	fmt.Printf("\n%s %s %s %s %s %s %s %s %s\n",
		strings.Repeat("-", investorWidth),
		strings.Repeat("-", nameWidth),
		strings.Repeat("-", docTypeWidth),
		strings.Repeat("-", statusWidth),
		strings.Repeat("-", priorityWidth),
		strings.Repeat("-", requestedWidth),
		strings.Repeat("-", dueWidth),
		strings.Repeat("-", overdueWidth),
		strings.Repeat("-", sourceWidth))

	fmt.Printf("Total Requirements: %d | Overdue: %d\n", len(requirements), overdueCount)
	return nil
}

// Sorting functions
func sortByDueDate(requirements []store.KYCRequirement) {
	// Sort by due date (nulls last), then by priority
	for i := 0; i < len(requirements)-1; i++ {
		for j := i + 1; j < len(requirements); j++ {
			iDue := requirements[i].DueAt
			jDue := requirements[j].DueAt

			// Handle null dates
			if iDue == nil && jDue != nil {
				requirements[i], requirements[j] = requirements[j], requirements[i]
			} else if iDue != nil && jDue != nil && iDue.After(*jDue) {
				requirements[i], requirements[j] = requirements[j], requirements[i]
			}
		}
	}
}

func sortByPriority(requirements []store.KYCRequirement) {
	priorityOrder := map[string]int{"CRITICAL": 0, "HIGH": 1, "MEDIUM": 2, "LOW": 3}

	for i := 0; i < len(requirements)-1; i++ {
		for j := i + 1; j < len(requirements); j++ {
			iPriority := priorityOrder[requirements[i].Priority]
			jPriority := priorityOrder[requirements[j].Priority]

			if iPriority > jPriority {
				requirements[i], requirements[j] = requirements[j], requirements[i]
			}
		}
	}
}

func sortByInvestor(requirements []store.KYCRequirement) {
	for i := 0; i < len(requirements)-1; i++ {
		for j := i + 1; j < len(requirements); j++ {
			if requirements[i].InvestorCode > requirements[j].InvestorCode {
				requirements[i], requirements[j] = requirements[j], requirements[i]
			}
		}
	}
}

func sortByDocType(requirements []store.KYCRequirement) {
	for i := 0; i < len(requirements)-1; i++ {
		for j := i + 1; j < len(requirements); j++ {
			if requirements[i].DocType > requirements[j].DocType {
				requirements[i], requirements[j] = requirements[j], requirements[i]
			}
		}
	}
}
