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
	"time"

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
	proxyFlag := eventCmd.Bool("proxy", false, "List reverse proxy access logs")

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

	// Proxy access log filters
	sourceIPFlag := eventCmd.String("source-ip", "", "Filter proxy logs by source IP")
	hostFlag := eventCmd.String("host", "", "Filter proxy logs by host")
	pathFlag := eventCmd.String("path", "", "Filter proxy logs by path")
	methodFlag := eventCmd.String("method", "", "Filter proxy logs by HTTP method")
	statusFlag := eventCmd.String("status", "", "Filter proxy logs by status")
	statusCodeFlag := eventCmd.String("status-code", "", "Filter proxy logs by status code")
	sortByFlag := eventCmd.String("sort-by", "", "Sort proxy logs by field")
	sortOrderFlag := eventCmd.String("sort-order", "", "Sort order: asc or desc")

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
		return err
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

	// List reverse proxy access logs
	if *proxyFlag {
		filters := models.ProxyEventFilters{
			Page:       *pageFlag,
			PageSize:   *pageSizeFlag,
			SortBy:     *sortByFlag,
			SortOrder:  *sortOrderFlag,
			Search:     *searchFlag,
			SourceIP:   *sourceIPFlag,
			Host:       *hostFlag,
			Path:       *pathFlag,
			UserID:     *userIDFlag,
			Method:     *methodFlag,
			Status:     *statusFlag,
			StatusCode: *statusCodeFlag,
			StartDate:  *startDateFlag,
			EndDate:    *endDateFlag,
		}
		return s.listProxyEvents(filters, *outputFlag)
	}

	eventCmd.Usage()
	return nil
}

// listAuditEvents lists all audit events with optional local filters.
// The audit API does not currently document or apply query parameters.
func (s *Service) listAuditEvents(filters models.AuditEventFilters, outputFormat string) error {
	bounds, err := parseAuditFilterBounds(filters)
	if err != nil {
		return err
	}

	resp, err := s.Client.MakeRequest("GET", "/events/audit", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var events []models.AuditEvent
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	filteredEvents := make([]models.AuditEvent, 0, len(events))
	for _, event := range events {
		matches, err := auditEventMatchesFilters(event, filters, bounds)
		if err != nil {
			return err
		}
		if matches {
			filteredEvents = append(filteredEvents, event)
		}
	}
	events = filteredEvents

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

type auditFilterBounds struct {
	start *time.Time
	end   *time.Time
}

func parseAuditFilterBounds(filters models.AuditEventFilters) (auditFilterBounds, error) {
	var bounds auditFilterBounds
	if filters.StartDate != "" {
		parsed, err := time.Parse(time.RFC3339Nano, filters.StartDate)
		if err != nil {
			return auditFilterBounds{}, fmt.Errorf("invalid audit start date %q: expected RFC3339: %w", filters.StartDate, err)
		}
		bounds.start = &parsed
	}
	if filters.EndDate != "" {
		parsed, err := time.Parse(time.RFC3339Nano, filters.EndDate)
		if err != nil {
			return auditFilterBounds{}, fmt.Errorf("invalid audit end date %q: expected RFC3339: %w", filters.EndDate, err)
		}
		bounds.end = &parsed
	}
	if bounds.start != nil && bounds.end != nil && bounds.start.After(*bounds.end) {
		return auditFilterBounds{}, fmt.Errorf("audit start date must not be after end date")
	}
	return bounds, nil
}

func auditEventMatchesFilters(event models.AuditEvent, filters models.AuditEventFilters, bounds auditFilterBounds) (bool, error) {
	if filters.UserID != "" && event.InitiatorID != filters.UserID {
		return false, nil
	}
	if filters.TargetID != "" && event.TargetID != filters.TargetID {
		return false, nil
	}
	if filters.ActivityCode != "" && !strings.EqualFold(event.ActivityCode, filters.ActivityCode) {
		return false, nil
	}
	if bounds.start != nil || bounds.end != nil {
		timestamp, err := time.Parse(time.RFC3339Nano, event.Timestamp)
		if err != nil {
			return false, fmt.Errorf("invalid timestamp %q for audit event %s: %w", event.Timestamp, event.ID, err)
		}
		if bounds.start != nil && timestamp.Before(*bounds.start) {
			return false, nil
		}
		if bounds.end != nil && timestamp.After(*bounds.end) {
			return false, nil
		}
	}
	if filters.Search == "" {
		return true, nil
	}

	meta, _ := json.Marshal(event.Meta)
	haystack := strings.ToLower(strings.Join([]string{
		event.Activity,
		event.ActivityCode,
		event.InitiatorID,
		event.InitiatorName,
		event.InitiatorEmail,
		event.TargetID,
		string(meta),
	}, " "))
	return strings.Contains(haystack, strings.ToLower(filters.Search)), nil
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
	fmt.Fprintln(w, "LAST EVENT\tUSER\tSOURCE\tDESTINATION\tPROTOCOL\tDIRECTION\tTX BYTES\tRX BYTES")
	fmt.Fprintln(w, "----------\t----\t------\t-----------\t--------\t---------\t--------\t--------")
	for _, event := range response.Data {
		// Use the most recent flow event as the display timestamp
		timestamp := "-"
		if len(event.Events) > 0 {
			timestamp = event.Events[len(event.Events)-1].Timestamp
			if len(timestamp) > 19 {
				timestamp = strings.Replace(timestamp[:19], "T", " ", 1)
			}
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

		user := "-"
		if event.User != nil {
			user = event.User.Email
			if user == "" {
				user = event.User.Name
			}
			if len(user) > 20 {
				user = user[:17] + "..."
			}
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%d\t%d\n",
			timestamp,
			user,
			formatTrafficEndpoint(event.Source),
			formatTrafficEndpoint(event.Destination),
			protocol,
			event.Direction,
			event.TxBytes,
			event.RxBytes,
		)
	}
	w.Flush()

	fmt.Printf("\nPage %d of %d | Total events: %d | Page size: %d\n",
		response.Page,
		response.TotalPages,
		response.TotalRecords,
		response.PageSize,
	)

	return nil
}

// formatTrafficEndpoint renders a traffic endpoint as "name (address)" with fallbacks
func formatTrafficEndpoint(endpoint models.TrafficEndpoint) string {
	name := endpoint.Name
	if name == "" {
		name = endpoint.DNSLabel
	}
	if name == "" {
		return endpoint.Address
	}
	if len(name) > 15 {
		name = name[:12] + "..."
	}
	if endpoint.Address == "" {
		return name
	}
	return fmt.Sprintf("%s (%s)", name, endpoint.Address)
}

// listProxyEvents lists reverse proxy access logs with pagination and filters
func (s *Service) listProxyEvents(filters models.ProxyEventFilters, outputFormat string) error {
	// Build query parameters
	params := url.Values{}
	if filters.Page > 0 {
		params.Add("page", strconv.Itoa(filters.Page))
	}
	if filters.PageSize > 0 {
		params.Add("page_size", strconv.Itoa(filters.PageSize))
	}
	stringParams := map[string]string{
		"sort_by":     filters.SortBy,
		"sort_order":  filters.SortOrder,
		"search":      filters.Search,
		"source_ip":   filters.SourceIP,
		"host":        filters.Host,
		"path":        filters.Path,
		"user_id":     filters.UserID,
		"user_email":  filters.UserEmail,
		"user_name":   filters.UserName,
		"method":      filters.Method,
		"status":      filters.Status,
		"status_code": filters.StatusCode,
		"start_date":  filters.StartDate,
		"end_date":    filters.EndDate,
	}
	for key, value := range stringParams {
		if value != "" {
			params.Add(key, value)
		}
	}

	var response models.ProxyEventResponse
	var err error
	if filters.SourceIP != "" {
		response, err = s.fetchProxyEventsBySourceIP(params, filters)
	} else {
		response, err = s.fetchProxyEventPage(params)
	}
	if err != nil {
		return err
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
	fmt.Fprintln(w, "TIMESTAMP\tMETHOD\tHOST\tPATH\tSTATUS\tDURATION\tSOURCE IP\tUSER ID")
	fmt.Fprintln(w, "---------\t------\t----\t----\t------\t--------\t---------\t-------")
	for _, event := range response.Data {
		timestamp := event.Timestamp
		if len(timestamp) > 19 {
			timestamp = strings.Replace(timestamp[:19], "T", " ", 1)
		}
		path := event.Path
		if len(path) > 30 {
			path = path[:27] + "..."
		}
		userID := event.UserID
		if userID == "" {
			userID = "-"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%dms\t%s\t%s\n",
			timestamp,
			event.Method,
			event.Host,
			path,
			event.StatusCode,
			event.DurationMS,
			event.SourceIP,
			userID,
		)
	}
	w.Flush()

	fmt.Printf("\nPage %d of %d | Total records: %d | Page size: %d\n",
		response.Page,
		response.TotalPages,
		response.TotalRecords,
		response.PageSize,
	)

	return nil
}

func (s *Service) fetchProxyEventPage(params url.Values) (models.ProxyEventResponse, error) {
	endpoint := "/events/proxy"
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	resp, err := s.Client.MakeRequest("GET", endpoint, nil)
	if err != nil {
		return models.ProxyEventResponse{}, err
	}
	defer resp.Body.Close()

	var response models.ProxyEventResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return models.ProxyEventResponse{}, fmt.Errorf("failed to decode response: %v", err)
	}
	return response, nil
}

// fetchProxyEventsBySourceIP uses native server pagination when source_ip is
// demonstrably selective. Servers known to ignore the parameter fall back to
// bounded search plus exact local filtering.
func (s *Service) fetchProxyEventsBySourceIP(params url.Values, filters models.ProxyEventFilters) (models.ProxyEventResponse, error) {
	nativeResponse, err := s.fetchProxyEventPage(params)
	if err != nil {
		return models.ProxyEventResponse{}, err
	}
	if nativeResponse.TotalRecords > 0 && proxyEventsMatchSourceIP(nativeResponse.Data, filters.SourceIP) {
		probeParams := cloneURLValues(params)
		probeParams.Del("source_ip")
		unfilteredResponse, err := s.fetchProxyEventPage(probeParams)
		if err != nil {
			return models.ProxyEventResponse{}, err
		}
		if nativeResponse.TotalRecords < unfilteredResponse.TotalRecords {
			return nativeResponse, nil
		}
	}
	return s.fetchAndFilterProxyEventsBySourceIP(params, filters)
}

func proxyEventsMatchSourceIP(events []models.ProxyEvent, sourceIP string) bool {
	if len(events) == 0 {
		return false
	}
	for _, event := range events {
		if event.SourceIP != sourceIP {
			return false
		}
	}
	return true
}

func cloneURLValues(values url.Values) url.Values {
	clone := make(url.Values, len(values))
	for key, entries := range values {
		clone[key] = append([]string(nil), entries...)
	}
	return clone
}

// fetchAndFilterProxyEventsBySourceIP provides compatibility with management
// servers that ignore source_ip by using the API's general search index as a
// bounded fallback, then enforcing an exact source-IP match locally.
func (s *Service) fetchAndFilterProxyEventsBySourceIP(params url.Values, filters models.ProxyEventFilters) (models.ProxyEventResponse, error) {
	const (
		fallbackPageSize   = 100
		maxFallbackPages   = 20
		maxFallbackRecords = fallbackPageSize * maxFallbackPages
	)

	fallbackParams := cloneURLValues(params)
	fallbackParams.Del("source_ip")
	if fallbackParams.Get("search") == "" {
		fallbackParams.Set("search", filters.SourceIP)
	}
	fallbackParams.Set("page", "1")
	fallbackParams.Set("page_size", strconv.Itoa(fallbackPageSize))

	matches := make([]models.ProxyEvent, 0)
	candidateCount := 0
	for page := 1; page <= maxFallbackPages; page++ {
		fallbackParams.Set("page", strconv.Itoa(page))
		pageResponse, err := s.fetchProxyEventPage(fallbackParams)
		if err != nil {
			return models.ProxyEventResponse{}, err
		}
		if len(pageResponse.Data) > fallbackPageSize || pageResponse.TotalPages > maxFallbackPages || pageResponse.TotalRecords > maxFallbackRecords || len(pageResponse.Data) > maxFallbackRecords-candidateCount {
			return models.ProxyEventResponse{}, fmt.Errorf(
				"proxy source-IP fallback matched too many candidates (%d records across %d pages); narrow the query with additional filters",
				pageResponse.TotalRecords,
				pageResponse.TotalPages,
			)
		}
		candidateCount += len(pageResponse.Data)
		for _, event := range pageResponse.Data {
			if event.SourceIP == filters.SourceIP {
				matches = append(matches, event)
			}
		}
		if pageResponse.TotalPages <= page || len(pageResponse.Data) == 0 {
			break
		}
	}

	page := filters.Page
	if page < 1 {
		page = 1
	}
	pageSize := filters.PageSize
	if pageSize < 1 {
		pageSize = 50
	}
	if page > maxFallbackRecords || pageSize > maxFallbackRecords {
		return models.ProxyEventResponse{}, fmt.Errorf(
			"proxy source-IP fallback pagination exceeds the maximum page or page size of %d",
			maxFallbackRecords,
		)
	}
	totalRecords := len(matches)
	totalPages := 0
	if totalRecords > 0 {
		totalPages = (totalRecords + pageSize - 1) / pageSize
	}
	start := totalRecords
	if page-1 <= totalRecords/pageSize {
		start = (page - 1) * pageSize
	}
	if start > totalRecords {
		start = totalRecords
	}
	end := totalRecords
	if pageSize <= totalRecords-start {
		end = start + pageSize
	}

	return models.ProxyEventResponse{
		Data:         matches[start:end],
		Page:         page,
		PageSize:     pageSize,
		TotalRecords: totalRecords,
		TotalPages:   totalPages,
	}, nil
}
