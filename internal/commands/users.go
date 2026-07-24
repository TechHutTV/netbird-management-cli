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

	// Approval flags (Cloud user approval flow)
	approveFlag := userCmd.String("approve", "", "Approve a pending user by ID")
	rejectFlag := userCmd.String("reject", "", "Reject a pending user by ID")

	// Password flags (embedded IdP)
	passwordFlag := userCmd.String("change-password", "", "Change password for user by ID (requires --old-password and --new-password)")
	oldPasswordFlag := userCmd.String("old-password", "", "Current password (use with --change-password)")
	newPasswordFlag := userCmd.String("new-password", "", "New password, min 8 characters (use with --change-password)")

	// Invite management flags (pre-provisioned invites)
	listInvitesFlag := userCmd.Bool("list-invites", false, "List pending user invites")
	createInviteFlag := userCmd.Bool("create-invite", false, "Create a user invite (requires --email and --name)")
	deleteInviteFlag := userCmd.String("delete-invite", "", "Delete a user invite by ID")
	regenerateInviteFlag := userCmd.String("regenerate-invite", "", "Regenerate a user invite by ID")
	expiresInFlag := userCmd.Int("expires-in", 0, "Invite expiration in seconds (use with --create-invite or --regenerate-invite)")

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

	if *approveFlag != "" {
		return s.approveUser(*approveFlag)
	}

	if *rejectFlag != "" {
		return s.rejectUser(*rejectFlag)
	}

	if *passwordFlag != "" {
		if *oldPasswordFlag == "" || *newPasswordFlag == "" {
			return fmt.Errorf("--change-password requires --old-password and --new-password")
		}
		return s.changeUserPassword(*passwordFlag, *oldPasswordFlag, *newPasswordFlag)
	}

	if *listInvitesFlag {
		return s.listUserInvites(*outputFlag)
	}

	if *createInviteFlag {
		if *email == "" || *name == "" {
			return fmt.Errorf("--create-invite requires --email and --name")
		}
		var groups []string
		if *autoGroups != "" {
			groups = strings.Split(*autoGroups, ",")
			for i := range groups {
				groups[i] = strings.TrimSpace(groups[i])
			}
		}
		return s.createUserInvite(*email, *name, *role, groups, *expiresInFlag)
	}

	if *deleteInviteFlag != "" {
		return s.deleteUserInvite(*deleteInviteFlag)
	}

	if *regenerateInviteFlag != "" {
		return s.regenerateUserInvite(*regenerateInviteFlag, *expiresInFlag)
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
	if user.PendingApproval {
		fmt.Printf("  Pending Approval: true\n")
	}
	if user.Issued != "" {
		fmt.Printf("  Issued:         %s\n", user.Issued)
	}
	fmt.Printf("  Last Login:     %s\n", user.LastLogin)
	fmt.Printf("  Restricted:     %t\n", user.Permissions.IsRestricted)

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

// getUserByID finds a user via the list endpoint (the API has no GET /users/{id})
func (s *Service) getUserByID(userID string) (*models.User, error) {
	resp, err := s.Client.MakeRequest("GET", "/users", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var users []models.User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("failed to decode users response: %v", err)
	}

	for i := range users {
		if users[i].ID == userID {
			return &users[i], nil
		}
	}
	return nil, fmt.Errorf("no user found with ID: %s", userID)
}

// removeUser deletes a user from the account
func (s *Service) removeUser(userID string) error {
	// Fetch user details first (via list; the API has no single-user GET)
	userPtr, err := s.getUserByID(userID)
	if err != nil {
		return err
	}
	user := *userPtr

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

	resp, err := s.Client.MakeRequest("DELETE", "/users/"+userID, nil)
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

// approveUser approves a pending user (Cloud user approval flow)
func (s *Service) approveUser(userID string) error {
	resp, err := s.Client.MakeRequest("POST", "/users/"+userID+"/approve", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("✓ User approved: %s\n", userID)
	return nil
}

// rejectUser rejects a pending user (Cloud user approval flow)
func (s *Service) rejectUser(userID string) error {
	resp, err := s.Client.MakeRequest("DELETE", "/users/"+userID+"/reject", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("✓ User rejected: %s\n", userID)
	return nil
}

// changeUserPassword updates a user's password (embedded IdP)
func (s *Service) changeUserPassword(userID, oldPassword, newPassword string) error {
	req := struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}{OldPassword: oldPassword, NewPassword: newPassword}

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := s.Client.MakeRequest("PUT", "/users/"+userID+"/password", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("✓ Password updated for user: %s\n", userID)
	return nil
}

// listUserInvites lists pending user invites
func (s *Service) listUserInvites(outputFormat string) error {
	resp, err := s.Client.MakeRequest("GET", "/users/invites", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var invites []models.UserInvite
	if err := json.NewDecoder(resp.Body).Decode(&invites); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	if len(invites) == 0 {
		fmt.Println("No pending invites found")
		return nil
	}

	// JSON output
	if outputFormat == "json" {
		output, err := json.MarshalIndent(invites, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %v", err)
		}
		fmt.Println(string(output))
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tEMAIL\tNAME\tROLE\tEXPIRES AT\tEXPIRED")
	fmt.Fprintln(w, "--\t-----\t----\t----\t----------\t-------")

	for _, invite := range invites {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%t\n",
			invite.ID,
			invite.Email,
			invite.Name,
			invite.Role,
			invite.ExpiresAt,
			invite.Expired,
		)
	}
	w.Flush()
	return nil
}

// createUserInvite creates a pre-provisioned user invite
func (s *Service) createUserInvite(email, name, role string, autoGroups []string, expiresIn int) error {
	if autoGroups == nil {
		autoGroups = []string{}
	}

	req := models.UserInviteCreateRequest{
		Email:      email,
		Name:       name,
		Role:       role,
		AutoGroups: autoGroups,
		ExpiresIn:  expiresIn,
	}

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := s.Client.MakeRequest("POST", "/users/invites", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var invite models.UserInvite
	if err := json.NewDecoder(resp.Body).Decode(&invite); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	fmt.Printf("✓ Invite created successfully!\n")
	fmt.Printf("  Invite ID:  %s\n", invite.ID)
	fmt.Printf("  Email:      %s\n", invite.Email)
	fmt.Printf("  Name:       %s\n", invite.Name)
	fmt.Printf("  Role:       %s\n", invite.Role)
	fmt.Printf("  Expires At: %s\n", invite.ExpiresAt)
	if invite.InviteToken != "" {
		fmt.Printf("  Token:      %s\n", invite.InviteToken)
	}

	return nil
}

// deleteUserInvite deletes a pending user invite
func (s *Service) deleteUserInvite(inviteID string) error {
	resp, err := s.Client.MakeRequest("DELETE", "/users/invites/"+inviteID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("✓ Invite deleted: %s\n", inviteID)
	return nil
}

// regenerateUserInvite regenerates a pending user invite (new token/expiry)
func (s *Service) regenerateUserInvite(inviteID string, expiresIn int) error {
	var body *bytes.Reader
	if expiresIn > 0 {
		bodyBytes, err := json.Marshal(struct {
			ExpiresIn int `json:"expires_in"`
		}{ExpiresIn: expiresIn})
		if err != nil {
			return fmt.Errorf("failed to marshal request: %v", err)
		}
		body = bytes.NewReader(bodyBytes)
	} else {
		body = bytes.NewReader([]byte("{}"))
	}

	resp, err := s.Client.MakeRequest("POST", "/users/invites/"+inviteID+"/regenerate", body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var invite models.UserInvite
	if err := json.NewDecoder(resp.Body).Decode(&invite); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	fmt.Printf("✓ Invite regenerated: %s\n", invite.ID)
	fmt.Printf("  Expires At: %s\n", invite.ExpiresAt)
	if invite.InviteToken != "" {
		fmt.Printf("  Token:      %s\n", invite.InviteToken)
	}
	return nil
}
