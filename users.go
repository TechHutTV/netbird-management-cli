// users.go
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

// handleUsersCommand handles all user-related operations
func handleUsersCommand(client *Client, args []string) error {
	userCmd := flag.NewFlagSet("user", flag.ContinueOnError)
	userCmd.SetOutput(os.Stderr)

	// Query flags
	listFlag := userCmd.Bool("list", false, "List all users")
	inspectFlag := userCmd.String("inspect", "", "Inspect user by ID")
	meFlag := userCmd.Bool("me", false, "Get current user information")
	serviceUserFilter := userCmd.Bool("service-users", false, "List only service users")
	regularUserFilter := userCmd.Bool("regular-users", false, "List only regular users")

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
		return client.getCurrentUser()
	}

	if *listFlag {
		filterType := ""
		if *serviceUserFilter {
			filterType = "service"
		} else if *regularUserFilter {
			filterType = "regular"
		}
		return client.listUsers(filterType)
	}

	if *inspectFlag != "" {
		return client.inspectUser(*inspectFlag)
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

		return client.inviteUser(*email, *name, *role, groups, *serviceUser)
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

		return client.updateUser(*updateFlag, *role, groups, isBlocked)
	}

	if *removeFlag != "" {
		return client.removeUser(*removeFlag)
	}

	if *resendInviteFlag != "" {
		return client.resendUserInvite(*resendInviteFlag)
	}

	userCmd.Usage()
	return nil
}

// listUsers lists all users in the account
func (c *Client) listUsers(filterType string) error {
	endpoint := "/users"
	if filterType == "service" {
		endpoint += "?service_user=true"
	} else if filterType == "regular" {
		endpoint += "?service_user=false"
	}

	resp, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var users []User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	if len(users) == 0 {
		fmt.Println("No users found")
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

// inspectUser shows detailed information about a specific user
func (c *Client) inspectUser(userID string) error {
	resp, err := c.makeRequest("GET", "/users/"+userID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	fmt.Printf("User ID:          %s\n", user.ID)
	fmt.Printf("Email:            %s\n", user.Email)
	fmt.Printf("Name:             %s\n", user.Name)
	fmt.Printf("Role:             %s\n", user.Role)
	fmt.Printf("Status:           %s\n", user.Status)
	fmt.Printf("Service User:     %t\n", user.IsServiceUser)
	fmt.Printf("Blocked:          %t\n", user.IsBlocked)
	fmt.Printf("Last Login:       %s\n", user.LastLogin)
	fmt.Printf("Dashboard View:   %s\n", user.Permissions.DashboardView)

	if len(user.AutoGroups) > 0 {
		fmt.Printf("Auto Groups:      %s\n", strings.Join(user.AutoGroups, ", "))
	} else {
		fmt.Printf("Auto Groups:      None\n")
	}

	return nil
}

// getCurrentUser retrieves the current authenticated user's information
func (c *Client) getCurrentUser() error {
	resp, err := c.makeRequest("GET", "/users/current", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
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
func (c *Client) inviteUser(email, name, role string, autoGroups []string, isServiceUser bool) error {
	if autoGroups == nil {
		autoGroups = []string{}
	}

	req := UserCreateRequest{
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

	resp, err := c.makeRequest("POST", "/users", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var user User
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
func (c *Client) updateUser(userID, role string, autoGroups []string, isBlocked bool) error {
	if autoGroups == nil {
		autoGroups = []string{}
	}

	req := UserUpdateRequest{
		Role:       role,
		AutoGroups: autoGroups,
		IsBlocked:  isBlocked,
	}

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := c.makeRequest("PUT", "/users/"+userID, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var user User
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
func (c *Client) removeUser(userID string) error {
	resp, err := c.makeRequest("DELETE", "/users/"+userID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("✓ User removed successfully: %s\n", userID)
	return nil
}

// resendUserInvite resends an invitation to a user
func (c *Client) resendUserInvite(userID string) error {
	resp, err := c.makeRequest("POST", "/users/"+userID+"/invite", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("✓ Invitation resent successfully to user: %s\n", userID)
	return nil
}

// printUserUsage provides specific help for the 'user' command
func printUserUsage() {
	fmt.Println("Usage: netbird-manage user <flag> [arguments]")
	fmt.Println("\nManage users and invitations.")
	fmt.Println("\nQuery Flags:")
	fmt.Println("  --list                           List all users")
	fmt.Println("    --service-users                List only service users")
	fmt.Println("    --regular-users                List only regular users")
	fmt.Println("  --inspect <user-id>              Inspect a specific user")
	fmt.Println("  --me                             Get current user information")
	fmt.Println()
	fmt.Println("Invite/Create Flags:")
	fmt.Println("  --invite                         Invite a new user")
	fmt.Println("    --email <email>                User email address (required)")
	fmt.Println("    --name <name>                  User full name (optional)")
	fmt.Println("    --role <role>                  User role: admin, user, owner (default: user)")
	fmt.Println("    --auto-groups <id1,id2,...>    Comma-separated group IDs for auto-assignment")
	fmt.Println("    --service-user                 Create as service user")
	fmt.Println()
	fmt.Println("Modification Flags:")
	fmt.Println("  --update <user-id>               Update user settings")
	fmt.Println("    --role <role>                  New role (optional)")
	fmt.Println("    --auto-groups <id1,id2,...>    New auto-groups (optional)")
	fmt.Println("    --blocked                      Block user access")
	fmt.Println("    --unblocked                    Unblock user access")
	fmt.Println()
	fmt.Println("  --remove <user-id>               Remove a user from the account")
	fmt.Println()
	fmt.Println("  --resend-invite <user-id>        Resend invitation to a user")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  netbird-manage user --list")
	fmt.Println("  netbird-manage user --me")
	fmt.Println("  netbird-manage user --invite --email user@example.com --role admin")
	fmt.Println("  netbird-manage user --update <user-id> --role user --blocked")
}
