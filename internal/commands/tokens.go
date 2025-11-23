// tokens.go
package commands

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"netbird-manage/internal/helpers"
	"netbird-manage/internal/models"
)

// HandleTokensCommand handles all token-related operations
func (s *Service) HandleTokensCommand(args []string) error {
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
		targetUserID, err = s.getCurrentUserID()
		if err != nil {
			return fmt.Errorf("failed to get current user ID: %v", err)
		}
	}

	// Handle commands
	if *listFlag {
		return s.listTokens(targetUserID)
	}

	if *inspectFlag != "" {
		return s.inspectToken(targetUserID, *inspectFlag)
	}

	if *createFlag {
		if *name == "" {
			return fmt.Errorf("--name is required when creating a token")
		}
		if *expiresIn < 1 || *expiresIn > 365 {
			return fmt.Errorf("--expires-in must be between 1 and 365 days")
		}
		return s.createToken(targetUserID, *name, *expiresIn)
	}

	if *revokeFlag != "" {
		return s.revokeToken(targetUserID, *revokeFlag)
	}

	tokenCmd.Usage()
	return nil
}

// getCurrentUserID retrieves the current user's ID
func (s *Service) getCurrentUserID() (string, error) {
	resp, err := s.Client.MakeRequest("GET", "/users/current", nil)
	if err != nil {
		// Check if it's a 403 error (service token)
		if resp != nil && resp.StatusCode == 403 {
			return "", fmt.Errorf("unable to get current user: service user tokens cannot access /users/current. Please provide --user-id flag with your user ID. Use 'user --list' to find your user ID")
		}
		return "", err
	}
	defer resp.Body.Close()

	var user models.User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	return user.ID, nil
}

// listTokens lists all personal access tokens for a user
func (s *Service) listTokens(userID string) error {
	endpoint := fmt.Sprintf("/users/%s/tokens", userID)

	resp, err := s.Client.MakeRequest("GET", endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var tokens []models.PersonalAccessToken
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
func (s *Service) inspectToken(userID, tokenID string) error {
	endpoint := fmt.Sprintf("/users/%s/tokens/%s", userID, tokenID)

	resp, err := s.Client.MakeRequest("GET", endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var token models.PersonalAccessToken
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
func (s *Service) createToken(userID, name string, expiresIn int) error {
	req := models.TokenCreateRequest{
		Name:      name,
		ExpiresIn: expiresIn,
	}

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	endpoint := fmt.Sprintf("/users/%s/tokens", userID)
	resp, err := s.Client.MakeRequest("POST", endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var tokenResp models.TokenCreateResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	// Display the token prominently with warning
	fmt.Println("Token created successfully!")
	fmt.Println()
	fmt.Println("IMPORTANT: Save this token now - it won't be shown again!")
	fmt.Println("============================================================")
	fmt.Printf("Token: %s\n", tokenResp.PlainToken)
	fmt.Println("============================================================")
	fmt.Println()
	fmt.Printf("Token ID:     %s\n", tokenResp.PersonalAccessToken.ID)
	fmt.Printf("Name:         %s\n", tokenResp.PersonalAccessToken.Name)
	fmt.Printf("Expires:      %s\n", tokenResp.PersonalAccessToken.ExpirationDate)
	fmt.Printf("Created By:   %s\n", tokenResp.PersonalAccessToken.CreatedBy)

	return nil
}

// revokeToken deletes/revokes a personal access token
func (s *Service) revokeToken(userID, tokenID string) error {
	endpoint := fmt.Sprintf("/users/%s/tokens/%s", userID, tokenID)

	// Fetch token details first
	resp, err := s.Client.MakeRequest("GET", endpoint, nil)
	if err != nil {
		return err
	}
	var token models.PersonalAccessToken
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		resp.Body.Close()
		return fmt.Errorf("failed to decode token: %v", err)
	}
	resp.Body.Close()

	// Build details map
	details := map[string]string{
		"Created": token.CreatedAt,
		"Expires": token.ExpirationDate,
	}
	if token.LastUsed != "" {
		details["Last Used"] = token.LastUsed
	}

	// Ask for confirmation
	if !helpers.ConfirmSingleDeletion("token", token.Name, tokenID, details) {
		return nil // User cancelled
	}

	resp, err = s.Client.MakeRequest("DELETE", endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("Token revoked successfully: %s\n", tokenID)
	return nil
}
