// dns_zones.go - Custom DNS zone and record management
package commands

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"netbird-manage/internal/helpers"
	"netbird-manage/internal/models"
)

// HandleDNSZonesCommand routes DNS zone and record commands
func (s *Service) HandleDNSZonesCommand(args []string) error {
	zoneCmd := flag.NewFlagSet("dns-zone", flag.ContinueOnError)
	zoneCmd.SetOutput(os.Stderr)
	zoneCmd.Usage = PrintDNSZoneUsage

	// Zone query flags
	listFlag := zoneCmd.Bool("list", false, "List all DNS zones")
	inspectFlag := zoneCmd.String("inspect", "", "Inspect a DNS zone by ID")

	// Zone modification flags
	createFlag := zoneCmd.Bool("create", false, "Create a DNS zone (requires --name and --domain)")
	updateFlag := zoneCmd.String("update", "", "Update a DNS zone by ID")
	deleteFlag := zoneCmd.String("delete", "", "Delete a DNS zone by ID")

	// Zone parameters
	nameFlag := zoneCmd.String("name", "", "Zone name")
	domainFlag := zoneCmd.String("domain", "", "Zone domain (e.g., internal.example.com)")
	enabledFlag := zoneCmd.String("enabled", "", "Enable/disable the zone (true/false)")
	searchDomainFlag := zoneCmd.String("search-domain", "", "Add the zone domain as a search domain (true/false)")
	distributionGroupsFlag := zoneCmd.String("distribution-groups", "", "Distribution group IDs (comma-separated)")

	// Record flags
	listRecordsFlag := zoneCmd.String("list-records", "", "List records in the given zone ID")
	inspectRecordFlag := zoneCmd.String("inspect-record", "", "Inspect a record by ID (requires --zone)")
	addRecordFlag := zoneCmd.String("add-record", "", "Add a record to the given zone ID")
	updateRecordFlag := zoneCmd.String("update-record", "", "Update a record by ID (requires --zone)")
	deleteRecordFlag := zoneCmd.String("delete-record", "", "Delete a record by ID (requires --zone)")

	// Record parameters
	zoneFlag := zoneCmd.String("zone", "", "Zone ID (for record operations)")
	recordNameFlag := zoneCmd.String("record-name", "", "Record name (e.g., host.internal.example.com)")
	recordTypeFlag := zoneCmd.String("record-type", "", "Record type: A, AAAA, or CNAME")
	contentFlag := zoneCmd.String("content", "", "Record content (IP address or hostname)")
	ttlFlag := zoneCmd.Int("ttl", 300, "Record TTL in seconds")

	// Output format
	outputFlag := zoneCmd.String("output", "table", "Output format: table or json")

	if len(args) == 1 {
		PrintDNSZoneUsage()
		return nil
	}

	if err := zoneCmd.Parse(args[1:]); err != nil {
		return err
	}

	if *listFlag {
		return s.listDNSZones(*outputFlag)
	}

	if *inspectFlag != "" {
		return s.inspectDNSZone(*inspectFlag, *outputFlag)
	}

	if *createFlag {
		if *nameFlag == "" || *domainFlag == "" {
			return fmt.Errorf("--create requires --name and --domain")
		}
		if *distributionGroupsFlag == "" {
			return fmt.Errorf("--create requires --distribution-groups")
		}
		req := models.DNSZoneRequest{
			Name:               *nameFlag,
			Domain:             *domainFlag,
			Enabled:            true,
			DistributionGroups: helpers.SplitCommaList(*distributionGroupsFlag),
		}
		if err := applyDNSZoneBoolFlags(*enabledFlag, *searchDomainFlag, &req.Enabled, &req.EnableSearchDomain); err != nil {
			return err
		}
		return s.createDNSZone(req)
	}

	if *updateFlag != "" {
		return s.updateDNSZone(*updateFlag, *nameFlag, *domainFlag, *enabledFlag, *searchDomainFlag, *distributionGroupsFlag)
	}

	if *deleteFlag != "" {
		return s.deleteDNSZone(*deleteFlag)
	}

	if *listRecordsFlag != "" {
		return s.listDNSRecords(*listRecordsFlag, *outputFlag)
	}

	if *inspectRecordFlag != "" {
		if *zoneFlag == "" {
			return fmt.Errorf("--inspect-record requires --zone")
		}
		return s.inspectDNSRecord(*zoneFlag, *inspectRecordFlag, *outputFlag)
	}

	if *addRecordFlag != "" {
		if *recordNameFlag == "" || *recordTypeFlag == "" || *contentFlag == "" {
			return fmt.Errorf("--add-record requires --record-name, --record-type, and --content")
		}
		req := models.DNSRecordRequest{
			Name:    *recordNameFlag,
			Type:    strings.ToUpper(*recordTypeFlag),
			Content: *contentFlag,
			TTL:     *ttlFlag,
		}
		if err := validateDNSRecordType(req.Type); err != nil {
			return err
		}
		return s.createDNSRecord(*addRecordFlag, req)
	}

	if *updateRecordFlag != "" {
		if *zoneFlag == "" {
			return fmt.Errorf("--update-record requires --zone")
		}
		return s.updateDNSRecord(*zoneFlag, *updateRecordFlag, *recordNameFlag, *recordTypeFlag, *contentFlag, *ttlFlag)
	}

	if *deleteRecordFlag != "" {
		if *zoneFlag == "" {
			return fmt.Errorf("--delete-record requires --zone")
		}
		return s.deleteDNSRecord(*zoneFlag, *deleteRecordFlag)
	}

	zoneCmd.Usage()
	return nil
}

// applyDNSZoneBoolFlags parses the optional --enabled and --search-domain values
func applyDNSZoneBoolFlags(enabledValue, searchDomainValue string, enabled, searchDomain *bool) error {
	if err := parseBoolFlag(enabledValue, "enabled", enabled); err != nil {
		return err
	}
	return parseBoolFlag(searchDomainValue, "search-domain", searchDomain)
}

// validateDNSRecordType checks a record type against the API enum
func validateDNSRecordType(recordType string) error {
	switch recordType {
	case "A", "AAAA", "CNAME":
		return nil
	}
	return fmt.Errorf("invalid record type %q: must be A, AAAA, or CNAME", recordType)
}

// listDNSZones lists all custom DNS zones
func (s *Service) listDNSZones(outputFormat string) error {
	resp, err := s.Client.MakeRequest("GET", "/dns/zones", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var zones []models.DNSZone
	if err := json.NewDecoder(resp.Body).Decode(&zones); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	if len(zones) == 0 {
		if outputFormat == "json" {
			fmt.Println("[]")
		} else {
			fmt.Println("No DNS zones found")
		}
		return nil
	}

	if outputFormat == "json" {
		output, err := json.MarshalIndent(zones, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %v", err)
		}
		fmt.Println(string(output))
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tDOMAIN\tENABLED\tSEARCH DOMAIN\tRECORDS\tDISTRIBUTION GROUPS")
	fmt.Fprintln(w, "--\t----\t------\t-------\t-------------\t-------\t-------------------")

	for _, zone := range zones {
		fmt.Fprintf(w, "%s\t%s\t%s\t%t\t%t\t%d\t%s\n",
			zone.ID,
			zone.Name,
			zone.Domain,
			zone.Enabled,
			zone.EnableSearchDomain,
			len(zone.Records),
			strings.Join(zone.DistributionGroups, ", "),
		)
	}
	w.Flush()
	return nil
}

// getDNSZoneByID fetches a single DNS zone
func (s *Service) getDNSZoneByID(zoneID string) (*models.DNSZone, error) {
	resp, err := s.Client.MakeRequest("GET", "/dns/zones/"+zoneID, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var zone models.DNSZone
	if err := json.NewDecoder(resp.Body).Decode(&zone); err != nil {
		return nil, fmt.Errorf("failed to decode zone response: %v", err)
	}
	return &zone, nil
}

// inspectDNSZone shows detailed information about a DNS zone
func (s *Service) inspectDNSZone(zoneID string, outputFormat string) error {
	zone, err := s.getDNSZoneByID(zoneID)
	if err != nil {
		return err
	}

	if outputFormat == "json" {
		output, err := json.MarshalIndent(zone, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %v", err)
		}
		fmt.Println(string(output))
		return nil
	}

	fmt.Printf("DNS Zone: %s (%s)\n", zone.Name, zone.ID)
	fmt.Println("--------------------------------------------------")
	fmt.Printf("  Domain:            %s\n", zone.Domain)
	fmt.Printf("  Enabled:           %t\n", zone.Enabled)
	fmt.Printf("  Search Domain:     %t\n", zone.EnableSearchDomain)
	if len(zone.DistributionGroups) > 0 {
		fmt.Printf("  Distribution:      %s\n", strings.Join(zone.DistributionGroups, ", "))
	}

	if len(zone.Records) > 0 {
		fmt.Println("\n  Records:")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "    ID\tNAME\tTYPE\tCONTENT\tTTL")
		fmt.Fprintln(w, "    --\t----\t----\t-------\t---")
		for _, record := range zone.Records {
			fmt.Fprintf(w, "    %s\t%s\t%s\t%s\t%d\n",
				record.ID, record.Name, record.Type, record.Content, record.TTL)
		}
		w.Flush()
	} else {
		fmt.Println("\n  Records:           None")
	}
	return nil
}

// createDNSZone creates a new custom DNS zone
func (s *Service) createDNSZone(req models.DNSZoneRequest) error {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := s.Client.MakeRequest("POST", "/dns/zones", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var zone models.DNSZone
	if err := json.NewDecoder(resp.Body).Decode(&zone); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	fmt.Printf("DNS zone created successfully\n")
	fmt.Printf("  ID:     %s\n", zone.ID)
	fmt.Printf("  Name:   %s\n", zone.Name)
	fmt.Printf("  Domain: %s\n", zone.Domain)
	return nil
}

// updateDNSZone updates a DNS zone, preserving unset fields
func (s *Service) updateDNSZone(zoneID, name, domain, enabledValue, searchDomainValue, distributionGroups string) error {
	current, err := s.getDNSZoneByID(zoneID)
	if err != nil {
		return err
	}

	req := models.DNSZoneRequest{
		Name:               current.Name,
		Domain:             current.Domain,
		Enabled:            current.Enabled,
		EnableSearchDomain: current.EnableSearchDomain,
		DistributionGroups: current.DistributionGroups,
	}
	if name != "" {
		req.Name = name
	}
	if domain != "" {
		req.Domain = domain
	}
	if distributionGroups != "" {
		req.DistributionGroups = helpers.SplitCommaList(distributionGroups)
	}
	if err := applyDNSZoneBoolFlags(enabledValue, searchDomainValue, &req.Enabled, &req.EnableSearchDomain); err != nil {
		return err
	}

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := s.Client.MakeRequest("PUT", "/dns/zones/"+zoneID, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("DNS zone %s updated successfully\n", zoneID)
	return nil
}

// deleteDNSZone deletes a DNS zone
func (s *Service) deleteDNSZone(zoneID string) error {
	zone, err := s.getDNSZoneByID(zoneID)
	if err != nil {
		return err
	}

	details := map[string]string{
		"Domain":  zone.Domain,
		"Records": fmt.Sprintf("%d", len(zone.Records)),
		"Enabled": fmt.Sprintf("%t", zone.Enabled),
	}

	if !helpers.ConfirmSingleDeletion("DNS zone", zone.Name, zoneID, details) {
		return nil
	}

	resp, err := s.Client.MakeRequest("DELETE", "/dns/zones/"+zoneID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("DNS zone %s deleted successfully\n", zoneID)
	return nil
}

// listDNSRecords lists all records in a DNS zone
func (s *Service) listDNSRecords(zoneID string, outputFormat string) error {
	resp, err := s.Client.MakeRequest("GET", "/dns/zones/"+zoneID+"/records", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var records []models.DNSRecord
	if err := json.NewDecoder(resp.Body).Decode(&records); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	if len(records) == 0 {
		if outputFormat == "json" {
			fmt.Println("[]")
		} else {
			fmt.Println("No records found in this zone")
		}
		return nil
	}

	if outputFormat == "json" {
		output, err := json.MarshalIndent(records, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %v", err)
		}
		fmt.Println(string(output))
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tTYPE\tCONTENT\tTTL")
	fmt.Fprintln(w, "--\t----\t----\t-------\t---")
	for _, record := range records {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\n",
			record.ID, record.Name, record.Type, record.Content, record.TTL)
	}
	w.Flush()
	return nil
}

// getDNSRecordByID fetches a single DNS record
func (s *Service) getDNSRecordByID(zoneID, recordID string) (*models.DNSRecord, error) {
	resp, err := s.Client.MakeRequest("GET", "/dns/zones/"+zoneID+"/records/"+recordID, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var record models.DNSRecord
	if err := json.NewDecoder(resp.Body).Decode(&record); err != nil {
		return nil, fmt.Errorf("failed to decode record response: %v", err)
	}
	return &record, nil
}

// inspectDNSRecord shows detailed information about a DNS record
func (s *Service) inspectDNSRecord(zoneID, recordID string, outputFormat string) error {
	record, err := s.getDNSRecordByID(zoneID, recordID)
	if err != nil {
		return err
	}

	if outputFormat == "json" {
		output, err := json.MarshalIndent(record, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %v", err)
		}
		fmt.Println(string(output))
		return nil
	}

	fmt.Printf("DNS Record: %s (%s)\n", record.Name, record.ID)
	fmt.Println("--------------------------------------------------")
	fmt.Printf("  Type:    %s\n", record.Type)
	fmt.Printf("  Content: %s\n", record.Content)
	fmt.Printf("  TTL:     %d\n", record.TTL)
	return nil
}

// createDNSRecord adds a record to a DNS zone
func (s *Service) createDNSRecord(zoneID string, req models.DNSRecordRequest) error {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := s.Client.MakeRequest("POST", "/dns/zones/"+zoneID+"/records", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var record models.DNSRecord
	if err := json.NewDecoder(resp.Body).Decode(&record); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	fmt.Printf("DNS record created successfully\n")
	fmt.Printf("  ID:      %s\n", record.ID)
	fmt.Printf("  Name:    %s\n", record.Name)
	fmt.Printf("  Type:    %s\n", record.Type)
	fmt.Printf("  Content: %s\n", record.Content)
	fmt.Printf("  TTL:     %d\n", record.TTL)
	return nil
}

// updateDNSRecord updates a record in a DNS zone, preserving unset fields
func (s *Service) updateDNSRecord(zoneID, recordID, name, recordType, content string, ttl int) error {
	current, err := s.getDNSRecordByID(zoneID, recordID)
	if err != nil {
		return err
	}

	req := models.DNSRecordRequest{
		Name:    current.Name,
		Type:    current.Type,
		Content: current.Content,
		TTL:     current.TTL,
	}
	if name != "" {
		req.Name = name
	}
	if recordType != "" {
		req.Type = strings.ToUpper(recordType)
		if err := validateDNSRecordType(req.Type); err != nil {
			return err
		}
	}
	if content != "" {
		req.Content = content
	}
	if ttl != 300 { // Only update if not default
		req.TTL = ttl
	}

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := s.Client.MakeRequest("PUT", "/dns/zones/"+zoneID+"/records/"+recordID, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("DNS record %s updated successfully\n", recordID)
	return nil
}

// deleteDNSRecord deletes a record from a DNS zone
func (s *Service) deleteDNSRecord(zoneID, recordID string) error {
	record, err := s.getDNSRecordByID(zoneID, recordID)
	if err != nil {
		return err
	}

	details := map[string]string{
		"Type":    record.Type,
		"Content": record.Content,
		"TTL":     fmt.Sprintf("%d", record.TTL),
	}

	if !helpers.ConfirmSingleDeletion("DNS record", record.Name, recordID, details) {
		return nil
	}

	resp, err := s.Client.MakeRequest("DELETE", "/dns/zones/"+zoneID+"/records/"+recordID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("DNS record %s deleted successfully\n", recordID)
	return nil
}
