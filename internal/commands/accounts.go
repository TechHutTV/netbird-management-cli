// accounts.go
package commands

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"netbird-manage/internal/helpers"
	"netbird-manage/internal/models"
)

// HandleAccountsCommand routes account-related commands
func (s *Service) HandleAccountsCommand(args []string) error {
	// Create a new flag set for the 'account' command
	accountCmd := flag.NewFlagSet("account", flag.ContinueOnError)
	accountCmd.SetOutput(os.Stderr)      // Send errors to stderr
	accountCmd.Usage = PrintAccountUsage // Set our custom usage function

	// Query flags
	listFlag := accountCmd.Bool("list", false, "List all accounts")
	inspectFlag := accountCmd.String("inspect", "", "Inspect an account by its ID")

	// Modification flags
	updateFlag := accountCmd.String("update", "", "Update an account by its ID (use with update flags)")
	deleteFlag := accountCmd.String("delete", "", "Delete an account by its ID")

	// Output flags
	outputFlag := accountCmd.String("output", "table", "Output format: table or json")

	// Update flags (use with --update)
	var uf accountUpdateFlags
	accountCmd.StringVar(&uf.peerLoginExp, "peer-login-expiration", "", "Peer login expiration (e.g., 24h, 7d)")
	accountCmd.StringVar(&uf.peerLoginExpEnabled, "peer-login-expiration-enabled", "", "Enable peer login expiration (true/false)")
	accountCmd.StringVar(&uf.peerInactivityExp, "peer-inactivity-expiration", "", "Peer inactivity timeout (e.g., 30d)")
	accountCmd.StringVar(&uf.peerInactivityExpEnabled, "peer-inactivity-expiration-enabled", "", "Enable peer inactivity expiration (true/false)")
	accountCmd.StringVar(&uf.dnsDomain, "dns-domain", "", "Network DNS domain")
	accountCmd.StringVar(&uf.networkRange, "network-range", "", "Network IP range (CIDR, e.g., 100.64.0.0/10)")
	accountCmd.StringVar(&uf.networkRangeV6, "network-range-v6", "", "IPv6 network range (CIDR)")
	accountCmd.StringVar(&uf.routingPeerDNS, "routing-peer-dns-resolution-enabled", "", "Enable DNS resolution on routing peers (true/false)")
	accountCmd.StringVar(&uf.jwtGroupsEnabled, "jwt-groups-enabled", "", "Enable JWT group claims (true/false)")
	accountCmd.StringVar(&uf.jwtGroupsClaim, "jwt-groups-claim", "", "JWT claim name for groups")
	accountCmd.StringVar(&uf.jwtAllowGroups, "jwt-allow-groups", "", "Comma-separated allowed groups")
	accountCmd.StringVar(&uf.groupsPropagation, "groups-propagation-enabled", "", "Enable groups propagation (true/false)")
	accountCmd.StringVar(&uf.regularUsersView, "regular-users-view-blocked", "", "Block regular users view (true/false)")
	accountCmd.StringVar(&uf.lazyConnection, "lazy-connection-enabled", "", "Enable lazy connections (true/false)")
	accountCmd.StringVar(&uf.peerApproval, "peer-approval-enabled", "", "Enable peer approval (true/false, Cloud-only)")
	accountCmd.StringVar(&uf.userApproval, "user-approval-required", "", "Require user approval (true/false, Cloud-only)")
	accountCmd.StringVar(&uf.trafficLogging, "traffic-logging", "", "Enable network traffic logging (true/false, Cloud-only)")

	// If no flags are provided (just 'netbird-manage account'), show usage
	if len(args) == 1 {
		PrintAccountUsage()
		return nil
	}

	// Parse the flags (all args *after* 'account')
	if err := accountCmd.Parse(args[1:]); err != nil {
		// The flag package will print an error, so we just return
		return nil
	}

	// Handle the flags
	if *listFlag {
		return s.listAccounts(*outputFlag)
	}

	if *inspectFlag != "" {
		return s.inspectAccount(*inspectFlag, *outputFlag)
	}

	if *updateFlag != "" {
		// Build update request from flags
		return s.updateAccountFromFlags(*updateFlag, uf)
	}

	if *deleteFlag != "" {
		return s.deleteAccount(*deleteFlag)
	}

	// If no valid flags are provided, show usage
	accountCmd.Usage()
	return nil
}

// listAccounts lists all accounts (returns single account)
func (s *Service) listAccounts(outputFormat string) error {
	resp, err := s.Client.MakeRequest("GET", "/accounts", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var accounts []models.Account
	if err := json.NewDecoder(resp.Body).Decode(&accounts); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	if len(accounts) == 0 {
		fmt.Println("No accounts found")
		return nil
	}

	// JSON output
	if outputFormat == "json" {
		output, err := json.MarshalIndent(accounts, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %v", err)
		}
		fmt.Println(string(output))
		return nil
	}

	// Table output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ACCOUNT ID\tDOMAIN\tNETWORK RANGE\tPEER LOGIN EXP\tDNS DOMAIN")
	fmt.Fprintln(w, "----------\t------\t-------------\t--------------\t----------")

	for _, account := range accounts {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			account.ID,
			account.Domain,
			account.Settings.NetworkRange,
			formatSeconds(account.Settings.PeerLoginExpiration),
			account.Settings.DNSDomain,
		)
	}
	w.Flush()

	// Show detailed settings for each account
	for _, account := range accounts {
		fmt.Println("\nAccount Settings:")
		printAccountSettings(account)
	}

	return nil
}

// printAccountSettings prints the settings and onboarding blocks of an account
func printAccountSettings(account models.Account) {
	settings := account.Settings
	fmt.Printf("  Peer Login Expiration Enabled:      %t\n", settings.PeerLoginExpirationEnabled)
	fmt.Printf("  Peer Login Expiration:              %s\n", formatSeconds(settings.PeerLoginExpiration))
	fmt.Printf("  Peer Inactivity Expiration Enabled: %t\n", settings.PeerInactivityExpirationEnabled)
	fmt.Printf("  Peer Inactivity Expiration:         %s\n", formatSeconds(settings.PeerInactivityExpiration))
	fmt.Printf("  DNS Domain:                         %s\n", settings.DNSDomain)
	fmt.Printf("  Network Range:                      %s\n", settings.NetworkRange)
	if settings.NetworkRangeV6 != "" {
		fmt.Printf("  Network Range (IPv6):               %s\n", settings.NetworkRangeV6)
	}
	fmt.Printf("  Routing Peer DNS Resolution:        %t\n", settings.RoutingPeerDNSResolutionEnabled)
	fmt.Printf("  JWT Groups Enabled:                 %t\n", settings.JWTGroupsEnabled)
	fmt.Printf("  JWT Groups Claim Name:              %s\n", settings.JWTGroupsClaimName)
	fmt.Printf("  Groups Propagation Enabled:         %t\n", settings.GroupsPropagationEnabled)
	fmt.Printf("  Regular Users View Blocked:         %t\n", settings.RegularUsersViewBlocked)
	fmt.Printf("  Peer Expose Enabled:                %t\n", settings.PeerExposeEnabled)
	fmt.Printf("  Lazy Connection Enabled:            %t\n", settings.LazyConnectionEnabled)
	if settings.AutoUpdateVersion != "" {
		fmt.Printf("  Auto Update Version:                %s\n", settings.AutoUpdateVersion)
	}
	fmt.Printf("  Auto Update Always:                 %t\n", settings.AutoUpdateAlways)

	if len(settings.JWTAllowGroups) > 0 {
		fmt.Printf("  JWT Allow Groups:                   %s\n", strings.Join(settings.JWTAllowGroups, ", "))
	}
	if len(settings.PeerExposeGroups) > 0 {
		fmt.Printf("  Peer Expose Groups:                 %s\n", strings.Join(settings.PeerExposeGroups, ", "))
	}
	if len(settings.IPv6EnabledGroups) > 0 {
		fmt.Printf("  IPv6 Enabled Groups:                %s\n", strings.Join(settings.IPv6EnabledGroups, ", "))
	}

	if settings.Extra != nil {
		fmt.Println("\nCloud Settings (extra):")
		fmt.Printf("  Peer Approval Enabled:              %t\n", settings.Extra.PeerApprovalEnabled)
		fmt.Printf("  User Approval Required:             %t\n", settings.Extra.UserApprovalRequired)
		fmt.Printf("  Network Traffic Logs Enabled:       %t\n", settings.Extra.NetworkTrafficLogsEnabled)
		fmt.Printf("  Network Traffic Packet Counter:     %t\n", settings.Extra.NetworkTrafficPacketCounterEnabled)
		if len(settings.Extra.NetworkTrafficLogsGroups) > 0 {
			fmt.Printf("  Network Traffic Logs Groups:        %s\n", strings.Join(settings.Extra.NetworkTrafficLogsGroups, ", "))
		}
	}

	if account.Onboarding != nil {
		fmt.Println("\nOnboarding Status:")
		fmt.Printf("  Signup Form Pending:                %t\n", account.Onboarding.SignupFormPending)
		fmt.Printf("  Onboarding Flow Pending:            %t\n", account.Onboarding.OnboardingFlowPending)
	}
}

// inspectAccount shows detailed information about an account
func (s *Service) inspectAccount(accountID string, outputFormat string) error {
	resp, err := s.Client.MakeRequest("GET", "/accounts/"+accountID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var account models.Account
	if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	// JSON output
	if outputFormat == "json" {
		output, err := json.MarshalIndent(account, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %v", err)
		}
		fmt.Println(string(output))
		return nil
	}

	// Display account details
	fmt.Printf("Account ID:      %s\n", account.ID)
	fmt.Printf("Domain:          %s\n", account.Domain)
	if account.DomainCategory != "" {
		fmt.Printf("Domain Category: %s\n", account.DomainCategory)
	}
	fmt.Printf("Created By:      %s\n", account.CreatedBy)
	fmt.Printf("Created At:      %s\n", account.CreatedAt)

	fmt.Println("\nSettings:")
	printAccountSettings(account)

	return nil
}

// accountUpdateFlags holds the raw string values of all --update sub-flags
type accountUpdateFlags struct {
	peerLoginExp             string
	peerLoginExpEnabled      string
	peerInactivityExp        string
	peerInactivityExpEnabled string
	dnsDomain                string
	networkRange             string
	networkRangeV6           string
	routingPeerDNS           string
	jwtGroupsEnabled         string
	jwtGroupsClaim           string
	jwtAllowGroups           string
	groupsPropagation        string
	regularUsersView         string
	lazyConnection           string
	peerApproval             string
	userApproval             string
	trafficLogging           string
}

// parseBoolFlag parses a true/false string flag into dest, leaving it unchanged when empty
func parseBoolFlag(value, flagName string, dest *bool) error {
	if value == "" {
		return nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fmt.Errorf("invalid value for %s: %v", flagName, err)
	}
	*dest = parsed
	return nil
}

// updateAccountFromFlags updates an account based on provided flags
func (s *Service) updateAccountFromFlags(accountID string, uf accountUpdateFlags) error {
	// First, fetch the current account state
	resp, err := s.Client.MakeRequest("GET", "/accounts/"+accountID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var account models.Account
	if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
		return fmt.Errorf("failed to decode current account: %v", err)
	}

	// Update only the fields that were provided
	if uf.peerLoginExp != "" {
		seconds, err := helpers.ParseDuration(uf.peerLoginExp, nil)
		if err != nil {
			return fmt.Errorf("invalid peer-login-expiration: %v", err)
		}
		account.Settings.PeerLoginExpiration = seconds
	}
	if uf.peerInactivityExp != "" {
		seconds, err := helpers.ParseDuration(uf.peerInactivityExp, nil)
		if err != nil {
			return fmt.Errorf("invalid peer-inactivity-expiration: %v", err)
		}
		account.Settings.PeerInactivityExpiration = seconds
	}
	if uf.dnsDomain != "" {
		account.Settings.DNSDomain = uf.dnsDomain
	}
	if uf.networkRange != "" {
		account.Settings.NetworkRange = uf.networkRange
	}
	if uf.networkRangeV6 != "" {
		account.Settings.NetworkRangeV6 = uf.networkRangeV6
	}
	if uf.jwtGroupsClaim != "" {
		account.Settings.JWTGroupsClaimName = uf.jwtGroupsClaim
	}
	if uf.jwtAllowGroups != "" {
		account.Settings.JWTAllowGroups = strings.Split(uf.jwtAllowGroups, ",")
	}

	boolFlags := []struct {
		value string
		name  string
		dest  *bool
	}{
		{uf.peerLoginExpEnabled, "peer-login-expiration-enabled", &account.Settings.PeerLoginExpirationEnabled},
		{uf.peerInactivityExpEnabled, "peer-inactivity-expiration-enabled", &account.Settings.PeerInactivityExpirationEnabled},
		{uf.routingPeerDNS, "routing-peer-dns-resolution-enabled", &account.Settings.RoutingPeerDNSResolutionEnabled},
		{uf.jwtGroupsEnabled, "jwt-groups-enabled", &account.Settings.JWTGroupsEnabled},
		{uf.groupsPropagation, "groups-propagation-enabled", &account.Settings.GroupsPropagationEnabled},
		{uf.regularUsersView, "regular-users-view-blocked", &account.Settings.RegularUsersViewBlocked},
		{uf.lazyConnection, "lazy-connection-enabled", &account.Settings.LazyConnectionEnabled},
	}
	for _, bf := range boolFlags {
		if err := parseBoolFlag(bf.value, bf.name, bf.dest); err != nil {
			return err
		}
	}

	// Cloud-only settings live in the nested "extra" object
	if uf.peerApproval != "" || uf.userApproval != "" || uf.trafficLogging != "" {
		if account.Settings.Extra == nil {
			account.Settings.Extra = &models.AccountSettingsExtra{}
		}
		extraFlags := []struct {
			value string
			name  string
			dest  *bool
		}{
			{uf.peerApproval, "peer-approval-enabled", &account.Settings.Extra.PeerApprovalEnabled},
			{uf.userApproval, "user-approval-required", &account.Settings.Extra.UserApprovalRequired},
			{uf.trafficLogging, "traffic-logging", &account.Settings.Extra.NetworkTrafficLogsEnabled},
		}
		for _, bf := range extraFlags {
			if err := parseBoolFlag(bf.value, bf.name, bf.dest); err != nil {
				return err
			}
		}
	}

	// Build update request
	updateReq := models.AccountUpdateRequest{
		Settings:   account.Settings,
		Onboarding: account.Onboarding,
	}

	// Send update request
	bodyBytes, err := json.Marshal(updateReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	updateResp, err := s.Client.MakeRequest("PUT", "/accounts/"+accountID, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer updateResp.Body.Close()

	fmt.Printf("Account %s updated successfully\n", accountID)
	return nil
}

// deleteAccount deletes an account and all its resources
func (s *Service) deleteAccount(accountID string) error {
	// Fetch account details first
	resp, err := s.Client.MakeRequest("GET", "/accounts/"+accountID, nil)
	if err != nil {
		return err
	}
	var account models.Account
	if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
		resp.Body.Close()
		return fmt.Errorf("failed to decode account: %v", err)
	}
	resp.Body.Close()

	// Build details map
	details := map[string]string{
		"Domain":     account.Domain,
		"Created By": account.CreatedBy,
		"Created At": account.CreatedAt,
		"WARNING":    "This will delete ALL associated resources!",
	}

	// Ask for confirmation
	if !helpers.ConfirmSingleDeletion("account", "", accountID, details) {
		return nil // User cancelled
	}

	resp, err = s.Client.MakeRequest("DELETE", "/accounts/"+accountID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("Account %s deleted successfully\n", accountID)
	return nil
}

// formatSeconds formats seconds into a human-readable duration string
func formatSeconds(seconds int) string {
	if seconds == 0 {
		return "disabled"
	}

	duration := time.Duration(seconds) * time.Second

	// Format as days, hours, or seconds
	if seconds >= 86400 { // >= 1 day
		days := seconds / 86400
		remainder := seconds % 86400
		if remainder == 0 {
			return fmt.Sprintf("%dd", days)
		}
		hours := remainder / 3600
		if hours > 0 {
			return fmt.Sprintf("%dd%dh", days, hours)
		}
		return fmt.Sprintf("%dd", days)
	}

	if seconds >= 3600 { // >= 1 hour
		hours := seconds / 3600
		remainder := seconds % 3600
		if remainder == 0 {
			return fmt.Sprintf("%dh", hours)
		}
		minutes := remainder / 60
		if minutes > 0 {
			return fmt.Sprintf("%dh%dm", hours, minutes)
		}
		return fmt.Sprintf("%dh", hours)
	}

	if seconds >= 60 { // >= 1 minute
		minutes := seconds / 60
		remainder := seconds % 60
		if remainder == 0 {
			return fmt.Sprintf("%dm", minutes)
		}
		return fmt.Sprintf("%dm%ds", minutes, remainder)
	}

	return duration.String()
}
