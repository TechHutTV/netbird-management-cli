// tokens.go
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"
)

// handleTokensCommand handles all token-related operations
func handleTokensCommand(client *Client, args []string) error {
	tokenCmd := flag.NewFlagSet("token", flag.ContinueOnError)
	tokenCmd.SetOutput(os.Stderr)

	// Query flags
	listFlag := tokenCmd.Bool("list", false, "List all personal access tokens")
	inspectFlag := tokenCmd.String("inspect", "", "Inspect token by ID")

	// Create flags
	createFlag := tokenCmd.Bool("create", false, "Create a new personal access token")
	name := tokenCmd.String("name", "", "Token name/description")
	expiresIn := tokenCmd.Int("expires-in", 90, "Expiration in days (1-365)")

	// Delete flags
	revokeFlag := tokenCmd.String("revoke", "", "Revoke/delete token by ID")

	// User ID flag (optional - defaults to current user)
	userID := tokenCmd.String("user-id", "", "User ID (defaults to current user)")

	if err := tokenCmd.Parse(args[1:]); err != nil {
		return err
	}

	// Get user ID (either from flag or current user)
	var targetUserID string
	var err error
	if *userID != "" {
		targetUserID = *userID
	} else {
		targetUserID, err = client.getCurrentUserID()
		if err != nil {
			return fmt.Errorf("failed to get current user ID: %v", err)
		}
	}

	// Handle commands
	if *listFlag {
		return client.listTokens(targetUserID)
	}

	if *inspectFlag != "" {
		return client.inspectToken(targetUserID, *inspectFlag)
	}

	if *createFlag {
		if *name == "" {
			return fmt.Errorf("--name is required when creating a token")
		}
		if *expiresIn < 1 || *expiresIn > 365 {
			return fmt.Errorf("--expires-in must be between 1 and 365 days")
		}
		return client.createToken(targetUserID, *name, *expiresIn)
	}

	if *revokeFlag != "" {
		return client.revokeToken(targetUserID, *revokeFlag)
	}

	tokenCmd.Usage()
	return nil
}

// getCurrentUserID retrieves the current user's ID
func (c *Client) getCurrentUserID() (string, error) {
	resp, err := c.makeRequest("GET", "/users/current", nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	return user.ID, nil
}

// listTokens lists all personal access tokens for a user
func (c *Client) listTokens(userID string) error {
	endpoint := fmt.Sprintf("/users/%s/tokens", userID)

	resp, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var tokens []PersonalAccessToken
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	if len(tokens) == 0 {
		fmt.Println("No tokens found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tCREATED AT\tEXPIRES\tLAST USED\tCREATED BY")
	fmt.Fprintln(w, "--\t----\t----------\t-------\t---------\t----------")

	for _, token := range tokens {
		lastUsed := token.LastUsed
		if lastUsed == "" {
			lastUsed = "Never"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			token.ID,
			token.Name,
			token.CreatedAt,
			token.ExpirationDate,
			lastUsed,
			token.CreatedBy,
		)
	}

	w.Flush()
	return nil
}

// inspectToken shows detailed information about a specific token
func (c *Client) inspectToken(userID, tokenID string) error {
	endpoint := fmt.Sprintf("/users/%s/tokens/%s", userID, tokenID)

	resp, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var token PersonalAccessToken
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	fmt.Printf("Token ID:         %s\n", token.ID)
	fmt.Printf("Name:             %s\n", token.Name)
	fmt.Printf("Created At:       %s\n", token.CreatedAt)
	fmt.Printf("Expiration Date:  %s\n", token.ExpirationDate)
	fmt.Printf("Created By:       %s\n", token.CreatedBy)

	if token.LastUsed != "" {
		fmt.Printf("Last Used:        %s\n", token.LastUsed)
	} else {
		fmt.Printf("Last Used:        Never\n")
	}

	return nil
}

// createToken creates a new personal access token
func (c *Client) createToken(userID, name string, expiresIn int) error {
	req := TokenCreateRequest{
		Name:      name,
		ExpiresIn: expiresIn,
	}

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	endpoint := fmt.Sprintf("/users/%s/tokens", userID)
	resp, err := c.makeRequest("POST", endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var tokenResp TokenCreateResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	// Display the token prominently with warning
	fmt.Println("✓ Token created successfully!")
	fmt.Println()
	fmt.Println("⚠️  IMPORTANT: Save this token now - it won't be shown again!")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Token: %s\n", tokenResp.PlainToken)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Printf("Token ID:     %s\n", tokenResp.PersonalAccessToken.ID)
	fmt.Printf("Name:         %s\n", tokenResp.PersonalAccessToken.Name)
	fmt.Printf("Expires:      %s\n", tokenResp.PersonalAccessToken.ExpirationDate)
	fmt.Printf("Created By:   %s\n", tokenResp.PersonalAccessToken.CreatedBy)

	return nil
}

// revokeToken deletes/revokes a personal access token
func (c *Client) revokeToken(userID, tokenID string) error {
	endpoint := fmt.Sprintf("/users/%s/tokens/%s", userID, tokenID)

	resp, err := c.makeRequest("DELETE", endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("✓ Token revoked successfully: %s\n", tokenID)
	return nil
}

// printTokenUsage provides specific help for the 'token' command
func printTokenUsage() {
	fmt.Println("Usage: netbird-manage token <flag> [arguments]")
	fmt.Println("\nManage personal access tokens.")
	fmt.Println("\nQuery Flags:")
	fmt.Println("  --list                           List all personal access tokens")
	fmt.Println("  --inspect <token-id>             Inspect a specific token")
	fmt.Println()
	fmt.Println("Create Flags:")
	fmt.Println("  --create                         Create a new personal access token")
	fmt.Println("    --name <name>                  Token name/description (required)")
	fmt.Println("    --expires-in <days>            Expiration in days (1-365, default: 90)")
	fmt.Println()
	fmt.Println("Delete Flags:")
	fmt.Println("  --revoke <token-id>              Revoke/delete a token")
	fmt.Println()
	fmt.Println("Optional:")
	fmt.Println("  --user-id <user-id>              User ID (defaults to current user)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  netbird-manage token --list")
	fmt.Println("  netbird-manage token --create --name \"CI/CD Token\" --expires-in 365")
	fmt.Println("  netbird-manage token --inspect <token-id>")
	fmt.Println("  netbird-manage token --revoke <token-id>")
	fmt.Println()
	fmt.Println("Note: The plain token value is only shown once during creation - save it immediately!")
}
