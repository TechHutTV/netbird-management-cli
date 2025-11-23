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
	peerLoginExpFlag := accountCmd.String("peer-login-expiration", "", "Peer login expiration (e.g., 24h, 7d)")
	peerInactivityExpFlag := accountCmd.String("peer-inactivity-expiration", "", "Peer inactivity timeout (e.g., 30d)")
	dnsDomainFlag := accountCmd.String("dns-domain", "", "Network DNS domain")
	networkRangeFlag := accountCmd.String("network-range", "", "Network IP range (CIDR, e.g., 100.64.0.0/10)")
	jwtGroupsEnabledFlag := accountCmd.String("jwt-groups-enabled", "", "Enable JWT group claims (true/false)")
	jwtGroupsClaimFlag := accountCmd.String("jwt-groups-claim", "", "JWT claim name for groups")
	jwtAllowGroupsFlag := accountCmd.String("jwt-allow-groups", "", "Comma-separated allowed groups")
	groupsPropagationFlag := accountCmd.String("groups-propagation-enabled", "", "Enable groups propagation (true/false)")
	regularUsersViewFlag := accountCmd.String("regular-users-view-blocked", "", "Block regular users view (true/false)")
	peerApprovalFlag := accountCmd.String("peer-approval-enabled", "", "Enable peer approval (true/false, Cloud-only)")
	trafficLoggingFlag := accountCmd.String("traffic-logging", "", "Enable traffic logging (true/false, Cloud-only)")

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
		return s.updateAccountFromFlags(*updateFlag,
			*peerLoginExpFlag,
			*peerInactivityExpFlag,
			*dnsDomainFlag,
			*networkRangeFlag,
			*jwtGroupsEnabledFlag,
			*jwtGroupsClaimFlag,
			*jwtAllowGroupsFlag,
			*groupsPropagationFlag,
			*regularUsersViewFlag,
			*peerApprovalFlag,
			*trafficLoggingFlag,
		)
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
		fmt.Printf("  Peer Login Expiration:        %s\n", formatSeconds(account.Settings.PeerLoginExpiration))
		fmt.Printf("  Peer Inactivity Expiration:   %s\n", formatSeconds(account.Settings.PeerInactivityExpiration))
		fmt.Printf("  DNS Domain:                   %s\n", account.Settings.DNSDomain)
		fmt.Printf("  Network Range:                %s\n", account.Settings.NetworkRange)
		fmt.Printf("  JWT Groups Enabled:           %t\n", account.Settings.JWTGroupsEnabled)
		fmt.Printf("  JWT Groups Claim:             %s\n", account.Settings.JWTGroupsClaim)
		fmt.Printf("  Groups Propagation Enabled:   %t\n", account.Settings.GroupsPropagationEnabled)
		fmt.Printf("  Regular Users View Blocked:   %t\n", account.Settings.RegularUsersViewBlocked)
		fmt.Printf("  Peer Approval Enabled:        %t\n", account.Settings.PeerApprovalEnabled)
		fmt.Printf("  Traffic Logging:              %t\n", account.Settings.TrafficLogging)

		if len(account.Settings.JWTAllowGroups) > 0 {
			fmt.Printf("  JWT Allow Groups:             %s\n", strings.Join(account.Settings.JWTAllowGroups, ", "))
		}

		if account.Onboarding != nil {
			fmt.Println("\nOnboarding Status:")
			fmt.Printf("  Signup Form Completed:        %t\n", account.Onboarding.SignupFormCompleted)
			fmt.Printf("  Flow Completed:               %t\n", account.Onboarding.FlowCompleted)
		}
	}

	return nil
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
	fmt.Printf("Account ID:     %s\n", account.ID)
	fmt.Printf("Domain:         %s\n", account.Domain)
	fmt.Printf("Created By:     %s\n", account.CreatedBy)
	fmt.Printf("Created At:     %s\n", account.CreatedAt)

	fmt.Println("\nSettings:")
	fmt.Printf("  Peer Login Expiration:        %s\n", formatSeconds(account.Settings.PeerLoginExpiration))
	fmt.Printf("  Peer Inactivity Expiration:   %s\n", formatSeconds(account.Settings.PeerInactivityExpiration))
	fmt.Printf("  DNS Domain:                   %s\n", account.Settings.DNSDomain)
	fmt.Printf("  Network Range:                %s\n", account.Settings.NetworkRange)
	fmt.Printf("  JWT Groups Enabled:           %t\n", account.Settings.JWTGroupsEnabled)
	fmt.Printf("  JWT Groups Claim:             %s\n", account.Settings.JWTGroupsClaim)
	fmt.Printf("  Groups Propagation Enabled:   %t\n", account.Settings.GroupsPropagationEnabled)
	fmt.Printf("  Regular Users View Blocked:   %t\n", account.Settings.RegularUsersViewBlocked)
	fmt.Printf("  Peer Approval Enabled:        %t\n", account.Settings.PeerApprovalEnabled)
	fmt.Printf("  Traffic Logging:              %t\n", account.Settings.TrafficLogging)

	if len(account.Settings.JWTAllowGroups) > 0 {
		fmt.Printf("  JWT Allow Groups:             %s\n", strings.Join(account.Settings.JWTAllowGroups, ", "))
	}

	if account.Onboarding != nil {
		fmt.Println("\nOnboarding:")
		fmt.Printf("  Signup Form Completed:        %t\n", account.Onboarding.SignupFormCompleted)
		fmt.Printf("  Flow Completed:               %t\n", account.Onboarding.FlowCompleted)
	}

	return nil
}

// updateAccountFromFlags updates an account based on provided flags
func (s *Service) updateAccountFromFlags(accountID string,
	peerLoginExp, peerInactivityExp, dnsDomain, networkRange,
	jwtGroupsEnabled, jwtGroupsClaim, jwtAllowGroups,
	groupsPropagation, regularUsersView, peerApproval, trafficLogging string) error {

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
	if peerLoginExp != "" {
		seconds, err := helpers.ParseDuration(peerLoginExp, nil)
		if err != nil {
			return fmt.Errorf("invalid peer-login-expiration: %v", err)
		}
		account.Settings.PeerLoginExpiration = seconds
	}
	if peerInactivityExp != "" {
		seconds, err := helpers.ParseDuration(peerInactivityExp, nil)
		if err != nil {
			return fmt.Errorf("invalid peer-inactivity-expiration: %v", err)
		}
		account.Settings.PeerInactivityExpiration = seconds
	}
	if dnsDomain != "" {
		account.Settings.DNSDomain = dnsDomain
	}
	if networkRange != "" {
		account.Settings.NetworkRange = networkRange
	}
	if jwtGroupsEnabled != "" {
		enabled, err := strconv.ParseBool(jwtGroupsEnabled)
		if err != nil {
			return fmt.Errorf("invalid value for jwt-groups-enabled: %v", err)
		}
		account.Settings.JWTGroupsEnabled = enabled
	}
	if jwtGroupsClaim != "" {
		account.Settings.JWTGroupsClaim = jwtGroupsClaim
	}
	if jwtAllowGroups != "" {
		account.Settings.JWTAllowGroups = strings.Split(jwtAllowGroups, ",")
	}
	if groupsPropagation != "" {
		enabled, err := strconv.ParseBool(groupsPropagation)
		if err != nil {
			return fmt.Errorf("invalid value for groups-propagation-enabled: %v", err)
		}
		account.Settings.GroupsPropagationEnabled = enabled
	}
	if regularUsersView != "" {
		blocked, err := strconv.ParseBool(regularUsersView)
		if err != nil {
			return fmt.Errorf("invalid value for regular-users-view-blocked: %v", err)
		}
		account.Settings.RegularUsersViewBlocked = blocked
	}
	if peerApproval != "" {
		enabled, err := strconv.ParseBool(peerApproval)
		if err != nil {
			return fmt.Errorf("invalid value for peer-approval-enabled: %v", err)
		}
		account.Settings.PeerApprovalEnabled = enabled
	}
	if trafficLogging != "" {
		enabled, err := strconv.ParseBool(trafficLogging)
		if err != nil {
			return fmt.Errorf("invalid value for traffic-logging: %v", err)
		}
		account.Settings.TrafficLogging = enabled
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
