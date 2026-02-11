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

// HandleDNSZonesCommand handles all DNS zone-related operations
func (s *Service) HandleDNSZonesCommand(args []string) error {
	cmd := flag.NewFlagSet("dns-zone", flag.ContinueOnError)
	cmd.SetOutput(os.Stderr)
	cmd.Usage = PrintDNSZoneUsage

	// Query flags
	listFlag := cmd.Bool("list", false, "List all DNS zones")
	inspectFlag := cmd.String("inspect", "", "Inspect a DNS zone by ID")
	outputFlag := cmd.String("output", "table", "Output format: table or json")

	// Zone create/update/delete flags
	createFlag := cmd.String("create", "", "Create a new DNS zone (name)")
	updateFlag := cmd.String("update", "", "Update a DNS zone by ID")
	deleteFlag := cmd.String("delete", "", "Delete a DNS zone by ID")

	// Zone properties
	domainFlag := cmd.String("domain", "", "Zone domain")
	groupsFlag := cmd.String("groups", "", "Comma-separated distribution group IDs")
	searchDomainFlag := cmd.Bool("search-domain", false, "Enable search domain")
	enabledFlag := cmd.Bool("enabled", true, "Enable the zone")
	disabledFlag := cmd.Bool("disabled", false, "Disable the zone")

	// Record operations
	listRecordsFlag := cmd.String("list-records", "", "List records in a zone (zone ID)")
	addRecordFlag := cmd.String("add-record", "", "Add a record to a zone (zone ID)")
	inspectRecordFlag := cmd.Bool("inspect-record", false, "Inspect a DNS record")
	updateRecordFlag := cmd.Bool("update-record", false, "Update a DNS record")
	deleteRecordFlag := cmd.Bool("delete-record", false, "Delete a DNS record")

	// Record properties
	zoneIDFlag := cmd.String("zone-id", "", "Zone ID (for record operations)")
	recordIDFlag := cmd.String("record-id", "", "Record ID (for record operations)")
	nameFlag := cmd.String("name", "", "Record name")
	typeFlag := cmd.String("type", "A", "Record type: A, AAAA, CNAME")
	contentFlag := cmd.String("content", "", "Record content (IP or domain)")
	ttlFlag := cmd.Int("ttl", 300, "Record TTL in seconds")

	if len(args) == 1 {
		PrintDNSZoneUsage()
		return nil
	}

	if err := cmd.Parse(args[1:]); err != nil {
		return nil
	}

	// Zone operations
	if *listFlag {
		return s.listDNSZones(*outputFlag)
	}

	if *inspectFlag != "" {
		return s.inspectDNSZone(*inspectFlag, *outputFlag)
	}

	if *createFlag != "" {
		if *domainFlag == "" {
			return fmt.Errorf("--domain is required when creating a zone")
		}
		var groups []string
		if *groupsFlag != "" {
			groups = helpers.SplitCommaList(*groupsFlag)
		} else {
			groups = []string{}
		}
		enabled := *enabledFlag && !*disabledFlag
		return s.createDNSZone(*createFlag, *domainFlag, groups, *searchDomainFlag, enabled)
	}

	if *updateFlag != "" {
		var groups []string
		if *groupsFlag != "" {
			groups = helpers.SplitCommaList(*groupsFlag)
		}
		enabled := *enabledFlag && !*disabledFlag
		return s.updateDNSZone(*updateFlag, *nameFlag, *domainFlag, groups, *searchDomainFlag, enabled)
	}

	if *deleteFlag != "" {
		return s.deleteDNSZone(*deleteFlag)
	}

	// Record operations
	if *listRecordsFlag != "" {
		return s.listDNSRecords(*listRecordsFlag, *outputFlag)
	}

	if *addRecordFlag != "" {
		if *nameFlag == "" || *contentFlag == "" {
			return fmt.Errorf("--name and --content are required when adding a record")
		}
		return s.createDNSRecord(*addRecordFlag, *nameFlag, *typeFlag, *contentFlag, *ttlFlag)
	}

	if *inspectRecordFlag {
		if *zoneIDFlag == "" || *recordIDFlag == "" {
			return fmt.Errorf("--zone-id and --record-id are required for --inspect-record")
		}
		return s.inspectDNSRecord(*zoneIDFlag, *recordIDFlag, *outputFlag)
	}

	if *updateRecordFlag {
		if *zoneIDFlag == "" || *recordIDFlag == "" {
			return fmt.Errorf("--zone-id and --record-id are required for --update-record")
		}
		return s.updateDNSRecord(*zoneIDFlag, *recordIDFlag, *nameFlag, *typeFlag, *contentFlag, *ttlFlag)
	}

	if *deleteRecordFlag {
		if *zoneIDFlag == "" || *recordIDFlag == "" {
			return fmt.Errorf("--zone-id and --record-id are required for --delete-record")
		}
		return s.deleteDNSRecord(*zoneIDFlag, *recordIDFlag)
	}

	PrintDNSZoneUsage()
	return nil
}

// listDNSZones lists all DNS zones
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
		fmt.Println("No DNS zones found")
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
	fmt.Fprintln(w, "ID\tNAME\tDOMAIN\tENABLED\tSEARCH DOMAIN\tRECORDS\tGROUPS")
	fmt.Fprintln(w, "--\t----\t------\t-------\t-------------\t-------\t------")

	for _, zone := range zones {
		enabledStr := "No"
		if zone.Enabled {
			enabledStr = "Yes"
		}
		searchStr := "No"
		if zone.EnableSearchDomain {
			searchStr = "Yes"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%d\t%d\n",
			zone.ID,
			zone.Name,
			zone.Domain,
			enabledStr,
			searchStr,
			len(zone.Records),
			len(zone.DistributionGroups),
		)
	}
	w.Flush()
	return nil
}

// inspectDNSZone retrieves details of a specific DNS zone
func (s *Service) inspectDNSZone(zoneID, outputFormat string) error {
	resp, err := s.Client.MakeRequest("GET", "/dns/zones/"+zoneID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var zone models.DNSZone
	if err := json.NewDecoder(resp.Body).Decode(&zone); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
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
	fmt.Println("---------------------------------")
	fmt.Printf("  Domain:         %s\n", zone.Domain)
	fmt.Printf("  Enabled:        %t\n", zone.Enabled)
	fmt.Printf("  Search Domain:  %t\n", zone.EnableSearchDomain)
	if len(zone.DistributionGroups) > 0 {
		fmt.Printf("  Groups:         %s\n", strings.Join(zone.DistributionGroups, ", "))
	}

	if len(zone.Records) > 0 {
		fmt.Printf("\n  Records (%d):\n", len(zone.Records))
		for _, record := range zone.Records {
			fmt.Printf("    - %s  %s  %s  (TTL: %d, ID: %s)\n",
				record.Name, record.Type, record.Content, record.TTL, record.ID)
		}
	} else {
		fmt.Println("  Records:        None")
	}

	return nil
}

// createDNSZone creates a new DNS zone
func (s *Service) createDNSZone(name, domain string, groups []string, searchDomain, enabled bool) error {
	req := models.DNSZoneRequest{
		Name:               name,
		Domain:             domain,
		Enabled:            enabled,
		EnableSearchDomain: searchDomain,
		DistributionGroups: groups,
	}

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

	fmt.Printf("✓ DNS zone created successfully!\n")
	fmt.Printf("  ID:     %s\n", zone.ID)
	fmt.Printf("  Name:   %s\n", zone.Name)
	fmt.Printf("  Domain: %s\n", zone.Domain)
	return nil
}

// updateDNSZone updates an existing DNS zone
func (s *Service) updateDNSZone(zoneID, name, domain string, groups []string, searchDomain, enabled bool) error {
	// Fetch current zone to preserve values
	resp, err := s.Client.MakeRequest("GET", "/dns/zones/"+zoneID, nil)
	if err != nil {
		return err
	}
	var currentZone models.DNSZone
	if err := json.NewDecoder(resp.Body).Decode(&currentZone); err != nil {
		resp.Body.Close()
		return fmt.Errorf("failed to decode current zone: %v", err)
	}
	resp.Body.Close()

	req := models.DNSZoneRequest{
		Name:               currentZone.Name,
		Domain:             currentZone.Domain,
		Enabled:            enabled,
		EnableSearchDomain: searchDomain,
		DistributionGroups: currentZone.DistributionGroups,
	}

	if name != "" {
		req.Name = name
	}
	if domain != "" {
		req.Domain = domain
	}
	if groups != nil {
		req.DistributionGroups = groups
	}

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err = s.Client.MakeRequest("PUT", "/dns/zones/"+zoneID, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var zone models.DNSZone
	if err := json.NewDecoder(resp.Body).Decode(&zone); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	fmt.Printf("✓ DNS zone updated successfully!\n")
	fmt.Printf("  ID:     %s\n", zone.ID)
	fmt.Printf("  Name:   %s\n", zone.Name)
	fmt.Printf("  Domain: %s\n", zone.Domain)
	return nil
}

// deleteDNSZone deletes a DNS zone
func (s *Service) deleteDNSZone(zoneID string) error {
	// Fetch zone details for confirmation
	resp, err := s.Client.MakeRequest("GET", "/dns/zones/"+zoneID, nil)
	if err != nil {
		return err
	}
	var zone models.DNSZone
	if err := json.NewDecoder(resp.Body).Decode(&zone); err != nil {
		resp.Body.Close()
		return fmt.Errorf("failed to decode zone: %v", err)
	}
	resp.Body.Close()

	details := map[string]string{
		"Domain":  zone.Domain,
		"Enabled": fmt.Sprintf("%t", zone.Enabled),
		"Records": fmt.Sprintf("%d", len(zone.Records)),
	}

	if !helpers.ConfirmSingleDeletion("DNS zone", zone.Name, zone.ID, details) {
		return nil
	}

	resp, err = s.Client.MakeRequest("DELETE", "/dns/zones/"+zoneID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("✓ DNS zone removed successfully: %s\n", zoneID)
	return nil
}

// listDNSRecords lists all records in a DNS zone
func (s *Service) listDNSRecords(zoneID, outputFormat string) error {
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
		fmt.Println("No DNS records found in this zone")
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
			record.ID,
			record.Name,
			record.Type,
			record.Content,
			record.TTL,
		)
	}
	w.Flush()
	return nil
}

// createDNSRecord creates a new DNS record in a zone
func (s *Service) createDNSRecord(zoneID, name, recordType, content string, ttl int) error {
	req := models.DNSRecordRequest{
		Name:    name,
		Type:    recordType,
		Content: content,
		TTL:     ttl,
	}

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

	fmt.Printf("✓ DNS record created successfully!\n")
	fmt.Printf("  ID:      %s\n", record.ID)
	fmt.Printf("  Name:    %s\n", record.Name)
	fmt.Printf("  Type:    %s\n", record.Type)
	fmt.Printf("  Content: %s\n", record.Content)
	fmt.Printf("  TTL:     %d\n", record.TTL)
	return nil
}

// inspectDNSRecord retrieves details of a specific DNS record
func (s *Service) inspectDNSRecord(zoneID, recordID, outputFormat string) error {
	resp, err := s.Client.MakeRequest("GET", "/dns/zones/"+zoneID+"/records/"+recordID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var record models.DNSRecord
	if err := json.NewDecoder(resp.Body).Decode(&record); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
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
	fmt.Println("---------------------------------")
	fmt.Printf("  Type:    %s\n", record.Type)
	fmt.Printf("  Content: %s\n", record.Content)
	fmt.Printf("  TTL:     %d\n", record.TTL)
	return nil
}

// updateDNSRecord updates an existing DNS record
func (s *Service) updateDNSRecord(zoneID, recordID, name, recordType, content string, ttl int) error {
	// Fetch current record
	resp, err := s.Client.MakeRequest("GET", "/dns/zones/"+zoneID+"/records/"+recordID, nil)
	if err != nil {
		return err
	}
	var current models.DNSRecord
	if err := json.NewDecoder(resp.Body).Decode(&current); err != nil {
		resp.Body.Close()
		return fmt.Errorf("failed to decode current record: %v", err)
	}
	resp.Body.Close()

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
		req.Type = recordType
	}
	if content != "" {
		req.Content = content
	}
	if ttl != 300 {
		req.TTL = ttl
	}

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err = s.Client.MakeRequest("PUT", "/dns/zones/"+zoneID+"/records/"+recordID, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var record models.DNSRecord
	if err := json.NewDecoder(resp.Body).Decode(&record); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	fmt.Printf("✓ DNS record updated successfully!\n")
	fmt.Printf("  ID:      %s\n", record.ID)
	fmt.Printf("  Name:    %s\n", record.Name)
	fmt.Printf("  Type:    %s\n", record.Type)
	fmt.Printf("  Content: %s\n", record.Content)
	fmt.Printf("  TTL:     %d\n", record.TTL)
	return nil
}

// deleteDNSRecord deletes a DNS record from a zone
func (s *Service) deleteDNSRecord(zoneID, recordID string) error {
	// Fetch record details for confirmation
	resp, err := s.Client.MakeRequest("GET", "/dns/zones/"+zoneID+"/records/"+recordID, nil)
	if err != nil {
		return err
	}
	var record models.DNSRecord
	if err := json.NewDecoder(resp.Body).Decode(&record); err != nil {
		resp.Body.Close()
		return fmt.Errorf("failed to decode record: %v", err)
	}
	resp.Body.Close()

	details := map[string]string{
		"Type":    record.Type,
		"Content": record.Content,
		"TTL":     fmt.Sprintf("%d", record.TTL),
	}

	if !helpers.ConfirmSingleDeletion("DNS record", record.Name, record.ID, details) {
		return nil
	}

	resp, err = s.Client.MakeRequest("DELETE", "/dns/zones/"+zoneID+"/records/"+recordID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("✓ DNS record removed successfully: %s\n", recordID)
	return nil
}
