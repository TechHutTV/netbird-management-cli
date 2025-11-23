// users.go
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

// HandleUsersCommand handles all user-related operations
func (s *Service) HandleUsersCommand(args []string) error {
	userCmd := flag.NewFlagSet("user", flag.ContinueOnError)
	userCmd.SetOutput(os.Stderr)

	// Query flags
	listFlag := userCmd.Bool("list", false, "List all users")
	meFlag := userCmd.Bool("me", false, "Get current user information")
	serviceUserFilter := userCmd.Bool("service-users", false, "List only service users")
	regularUserFilter := userCmd.Bool("regular-users", false, "List only regular users")
	outputFlag := userCmd.String("output", "table", "Output format: table or json")

	// Create/Invite flags
	inviteFlag := userCmd.Bool("invite", false, "Invite a new user")
	email := userCmd.String("email", "", "User email address")
	name := userCmd.String("name", "", "User full name")
	role := userCmd.String("role", "user", "User role (admin, user, owner)")
	autoGroups := userCmd.String("auto-groups", "", "Comma-separated group IDs for auto-assignment")
	serviceUser := userCmd.Bool("service-user", false, "Create as service user")

	// Update flags
	updateFlag := userCmd.String("update", "", "Update user by ID")
	blocked := userCmd.Bool("blocked", false, "Block user access (use with --update)")
	unblocked := userCmd.Bool("unblocked", false, "Unblock user access (use with --update)")

	// Delete flags
	removeFlag := userCmd.String("remove", "", "Remove user by ID")

	// Resend invite flag
	resendInviteFlag := userCmd.String("resend-invite", "", "Resend invitation to user by ID")

	if err := userCmd.Parse(args[1:]); err != nil {
		return err
	}

	// Handle commands
	if *meFlag {
		return s.getCurrentUser(*outputFlag)
	}

	if *listFlag || *serviceUserFilter || *regularUserFilter {
		filterType := ""
		if *serviceUserFilter {
			filterType = "service"
		} else if *regularUserFilter {
			filterType = "regular"
		}
		return s.listUsers(filterType, *outputFlag)
	}

	if *inviteFlag {
		if *email == "" {
			return fmt.Errorf("--email is required when inviting a user")
		}

		var groups []string
		if *autoGroups != "" {
			groups = strings.Split(*autoGroups, ",")
			for i := range groups {
				groups[i] = strings.TrimSpace(groups[i])
			}
		}

		return s.inviteUser(*email, *name, *role, groups, *serviceUser)
	}

	if *updateFlag != "" {
		if *blocked && *unblocked {
			return fmt.Errorf("cannot use both --blocked and --unblocked")
		}

		var groups []string
		if *autoGroups != "" {
			groups = strings.Split(*autoGroups, ",")
			for i := range groups {
				groups[i] = strings.TrimSpace(groups[i])
			}
		}

		isBlocked := false
		if *blocked {
			isBlocked = true
		}

		return s.updateUser(*updateFlag, *role, groups, isBlocked)
	}

	if *removeFlag != "" {
		return s.removeUser(*removeFlag)
	}

	if *resendInviteFlag != "" {
		return s.resendUserInvite(*resendInviteFlag)
	}

	userCmd.Usage()
	return nil
}

// listUsers lists all users in the account
func (s *Service) listUsers(filterType string, outputFormat string) error {
	endpoint := "/users"
	if filterType == "service" {
		endpoint += "?service_user=true"
	} else if filterType == "regular" {
		endpoint += "?service_user=false"
	}

	resp, err := s.Client.MakeRequest("GET", endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var users []models.User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	if len(users) == 0 {
		fmt.Println("No users found")
		return nil
	}

	// JSON output
	if outputFormat == "json" {
		output, err := json.MarshalIndent(users, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %v", err)
		}
		fmt.Println(string(output))
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tEMAIL\tNAME\tROLE\tSTATUS\tSERVICE\tBLOCKED\tLAST LOGIN")
	fmt.Fprintln(w, "--\t-----\t----\t----\t------\t-------\t-------\t----------")

	for _, user := range users {
		serviceUserStr := "No"
		if user.IsServiceUser {
			serviceUserStr = "Yes"
		}
		blockedStr := "No"
		if user.IsBlocked {
			blockedStr = "Yes"
		}
		lastLogin := user.LastLogin
		if lastLogin == "" {
			lastLogin = "Never"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			user.ID,
			user.Email,
			user.Name,
			user.Role,
			user.Status,
			serviceUserStr,
			blockedStr,
			lastLogin,
		)
	}

	w.Flush()
	return nil
}

// getCurrentUser retrieves the current authenticated user's information
func (s *Service) getCurrentUser(outputFormat string) error {
	resp, err := s.Client.MakeRequest("GET", "/users/current", nil)
	if err != nil {
		// Check if it's a 403 error (service token)
		if resp != nil && resp.StatusCode == 403 {
			return fmt.Errorf("unable to get current user: this endpoint is not available for service user tokens. Use 'user --list' to see all users instead")
		}
		return err
	}
	defer resp.Body.Close()

	var user models.User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	// JSON output
	if outputFormat == "json" {
		output, err := json.MarshalIndent(user, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %v", err)
		}
		fmt.Println(string(output))
		return nil
	}

	fmt.Printf("Current User Information:\n")
	fmt.Printf("  User ID:        %s\n", user.ID)
	fmt.Printf("  Email:          %s\n", user.Email)
	fmt.Printf("  Name:           %s\n", user.Name)
	fmt.Printf("  Role:           %s\n", user.Role)
	fmt.Printf("  Status:         %s\n", user.Status)
	fmt.Printf("  Service User:   %t\n", user.IsServiceUser)
	fmt.Printf("  Blocked:        %t\n", user.IsBlocked)
	fmt.Printf("  Last Login:     %s\n", user.LastLogin)

	if len(user.AutoGroups) > 0 {
		fmt.Printf("  Auto Groups:    %s\n", strings.Join(user.AutoGroups, ", "))
	}

	return nil
}

// inviteUser creates/invites a new user
func (s *Service) inviteUser(email, name, role string, autoGroups []string, isServiceUser bool) error {
	if autoGroups == nil {
		autoGroups = []string{}
	}

	req := models.UserCreateRequest{
		Email:         email,
		Name:          name,
		Role:          role,
		AutoGroups:    autoGroups,
		IsServiceUser: isServiceUser,
	}

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := s.Client.MakeRequest("POST", "/users", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var user models.User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	userType := "User"
	if isServiceUser {
		userType = "Service user"
	}

	fmt.Printf("✓ %s invited successfully!\n", userType)
	fmt.Printf("  User ID:   %s\n", user.ID)
	fmt.Printf("  Email:     %s\n", user.Email)
	fmt.Printf("  Name:      %s\n", user.Name)
	fmt.Printf("  Role:      %s\n", user.Role)
	fmt.Printf("  Status:    %s\n", user.Status)

	return nil
}

// updateUser updates an existing user's settings
func (s *Service) updateUser(userID, role string, autoGroups []string, isBlocked bool) error {
	if autoGroups == nil {
		autoGroups = []string{}
	}

	req := models.UserUpdateRequest{
		Role:       role,
		AutoGroups: autoGroups,
		IsBlocked:  isBlocked,
	}

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := s.Client.MakeRequest("PUT", "/users/"+userID, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var user models.User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	fmt.Printf("✓ User updated successfully!\n")
	fmt.Printf("  User ID:   %s\n", user.ID)
	fmt.Printf("  Email:     %s\n", user.Email)
	fmt.Printf("  Role:      %s\n", user.Role)
	fmt.Printf("  Blocked:   %t\n", user.IsBlocked)

	return nil
}

// removeUser deletes a user from the account
func (s *Service) removeUser(userID string) error {
	// Fetch user details first
	resp, err := s.Client.MakeRequest("GET", "/users/"+userID, nil)
	if err != nil {
		return err
	}
	var user models.User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		resp.Body.Close()
		return fmt.Errorf("failed to decode user: %v", err)
	}
	resp.Body.Close()

	// Build details map
	details := map[string]string{
		"Email":  user.Email,
		"Role":   user.Role,
		"Status": user.Status,
	}
	if user.IsBlocked {
		details["Blocked"] = "Yes"
	}

	// Ask for confirmation
	if !helpers.ConfirmSingleDeletion("user", user.Name, userID, details) {
		return nil // User cancelled
	}

	resp, err = s.Client.MakeRequest("DELETE", "/users/"+userID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("✓ User removed successfully: %s\n", userID)
	return nil
}

// resendUserInvite resends an invitation to a user
func (s *Service) resendUserInvite(userID string) error {
	resp, err := s.Client.MakeRequest("POST", "/users/"+userID+"/invite", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("✓ Invitation resent successfully to user: %s\n", userID)
	return nil
}
