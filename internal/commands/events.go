// events.go
package commands

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"netbird-manage/internal/models"
)

// HandleEventsCommand routes event-related commands
func (s *Service) HandleEventsCommand(args []string) error {
	// Create a new flag set for the 'event' command
	eventCmd := flag.NewFlagSet("event", flag.ContinueOnError)
	eventCmd.SetOutput(os.Stderr)
	eventCmd.Usage = PrintEventUsage

	// Define the flags for the 'event' command
	auditFlag := eventCmd.Bool("audit", false, "List audit events")
	trafficFlag := eventCmd.Bool("traffic", false, "List network traffic events")

	// Audit event filters
	userIDFlag := eventCmd.String("user-id", "", "Filter by user ID")
	targetIDFlag := eventCmd.String("target-id", "", "Filter by target resource ID")
	activityCodeFlag := eventCmd.String("activity-code", "", "Filter by activity code")
	startDateFlag := eventCmd.String("start-date", "", "Start date (ISO 8601)")
	endDateFlag := eventCmd.String("end-date", "", "End date (ISO 8601)")
	searchFlag := eventCmd.String("search", "", "Search in initiator/target names")

	// Traffic event filters
	reporterIDFlag := eventCmd.String("reporter-id", "", "Filter by reporting peer")
	protocolFlag := eventCmd.Int("protocol", 0, "Filter by protocol number (6=TCP, 17=UDP)")
	typeFlag := eventCmd.String("type", "", "Filter by event type")
	connectionTypeFlag := eventCmd.String("connection-type", "", "Filter by connection type")
	directionFlag := eventCmd.String("direction", "", "Filter by traffic direction")

	// Pagination
	pageFlag := eventCmd.Int("page", 1, "Page number")
	pageSizeFlag := eventCmd.Int("page-size", 100, "Items per page")

	// Output
	outputFlag := eventCmd.String("output", "table", "Output format: table or json")

	// If no flags are provided (just 'netbird-manage event'), show usage
	if len(args) == 1 {
		PrintEventUsage()
		return nil
	}

	// Parse the flags (all args *after* 'event')
	if err := eventCmd.Parse(args[1:]); err != nil {
		return nil
	}

	// Handle the flags in priority order

	// List audit events
	if *auditFlag {
		filters := models.AuditEventFilters{
			UserID:       *userIDFlag,
			TargetID:     *targetIDFlag,
			ActivityCode: *activityCodeFlag,
			StartDate:    *startDateFlag,
			EndDate:      *endDateFlag,
			Search:       *searchFlag,
		}
		return s.listAuditEvents(filters, *outputFlag)
	}

	// List traffic events
	if *trafficFlag {
		filters := models.TrafficEventFilters{
			Page:           *pageFlag,
			PageSize:       *pageSizeFlag,
			UserID:         *userIDFlag,
			ReporterID:     *reporterIDFlag,
			Protocol:       *protocolFlag,
			Type:           *typeFlag,
			ConnectionType: *connectionTypeFlag,
			Direction:      *directionFlag,
			Search:         *searchFlag,
			StartDate:      *startDateFlag,
			EndDate:        *endDateFlag,
		}
		return s.listTrafficEvents(filters, *outputFlag)
	}

	eventCmd.Usage()
	return nil
}

// listAuditEvents lists all audit events with optional filters
func (s *Service) listAuditEvents(filters models.AuditEventFilters, outputFormat string) error {
	// Build query parameters
	params := url.Values{}
	if filters.UserID != "" {
		params.Add("user_id", filters.UserID)
	}
	if filters.TargetID != "" {
		params.Add("target_id", filters.TargetID)
	}
	if filters.ActivityCode != "" {
		params.Add("activity_code", filters.ActivityCode)
	}
	if filters.StartDate != "" {
		params.Add("start_date", filters.StartDate)
	}
	if filters.EndDate != "" {
		params.Add("end_date", filters.EndDate)
	}
	if filters.Search != "" {
		params.Add("search", filters.Search)
	}

	endpoint := "/events/audit"
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	resp, err := s.Client.MakeRequest("GET", endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var events []models.AuditEvent
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	// JSON output
	if outputFormat == "json" {
		output, err := json.MarshalIndent(events, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %v", err)
		}
		fmt.Println(string(output))
		return nil
	}

	// Table output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "TIMESTAMP\tACTIVITY\tINITIATOR\tTARGET ID")
	fmt.Fprintln(w, "---------\t--------\t---------\t---------")
	for _, event := range events {
		// Format timestamp (remove milliseconds and timezone for readability)
		timestamp := event.Timestamp
		if len(timestamp) > 19 {
			timestamp = strings.Replace(timestamp[:19], "T", " ", 1)
		}

		initiator := event.InitiatorEmail
		if initiator == "" {
			initiator = event.InitiatorName
		}
		if initiator == "" {
			initiator = event.InitiatorID
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			timestamp,
			event.Activity,
			initiator,
			event.TargetID,
		)
	}
	w.Flush()

	fmt.Printf("\nTotal events: %d\n", len(events))

	return nil
}

// listTrafficEvents lists network traffic events with pagination and filters
func (s *Service) listTrafficEvents(filters models.TrafficEventFilters, outputFormat string) error {
	// Build query parameters
	params := url.Values{}
	if filters.Page > 0 {
		params.Add("page", strconv.Itoa(filters.Page))
	}
	if filters.PageSize > 0 {
		params.Add("page_size", strconv.Itoa(filters.PageSize))
	}
	if filters.UserID != "" {
		params.Add("user_id", filters.UserID)
	}
	if filters.ReporterID != "" {
		params.Add("reporter_id", filters.ReporterID)
	}
	if filters.Protocol > 0 {
		params.Add("protocol", strconv.Itoa(filters.Protocol))
	}
	if filters.Type != "" {
		params.Add("type", filters.Type)
	}
	if filters.ConnectionType != "" {
		params.Add("connection_type", filters.ConnectionType)
	}
	if filters.Direction != "" {
		params.Add("direction", filters.Direction)
	}
	if filters.Search != "" {
		params.Add("search", filters.Search)
	}
	if filters.StartDate != "" {
		params.Add("start_date", filters.StartDate)
	}
	if filters.EndDate != "" {
		params.Add("end_date", filters.EndDate)
	}

	endpoint := "/events/network-traffic"
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	resp, err := s.Client.MakeRequest("GET", endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var response models.TrafficEventResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	// JSON output
	if outputFormat == "json" {
		output, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %v", err)
		}
		fmt.Println(string(output))
		return nil
	}

	// Table output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "TIMESTAMP\tUSER\tREPORTER\tPROTOCOL\tSRC IP\tDST IP\tBYTES OUT\tBYTES IN")
	fmt.Fprintln(w, "---------\t----\t--------\t--------\t------\t------\t---------\t--------")
	for _, event := range response.Data {
		// Format timestamp
		timestamp := event.Timestamp
		if len(timestamp) > 19 {
			timestamp = strings.Replace(timestamp[:19], "T", " ", 1)
		}

		// Format protocol
		protocol := fmt.Sprintf("%d", event.Protocol)
		if event.Protocol == 6 {
			protocol = "TCP"
		} else if event.Protocol == 17 {
			protocol = "UDP"
		} else if event.Protocol == 1 {
			protocol = "ICMP"
		}

		// Truncate email for display
		user := event.UserEmail
		if len(user) > 20 {
			user = user[:17] + "..."
		}

		// Truncate reporter name
		reporter := event.ReporterName
		if len(reporter) > 15 {
			reporter = reporter[:12] + "..."
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%d\t%d\n",
			timestamp,
			user,
			reporter,
			protocol,
			event.SourceIP,
			event.DestinationIP,
			event.BytesSent,
			event.BytesReceived,
		)
	}
	w.Flush()

	fmt.Printf("\nPage %d of %d | Total events: %d | Page size: %d\n",
		response.Page,
		(response.TotalCount+response.PageSize-1)/response.PageSize,
		response.TotalCount,
		response.PageSize,
	)

	return nil
}
