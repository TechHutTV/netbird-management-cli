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

// HandleIdentityProvidersCommand handles all identity provider-related operations
func (s *Service) HandleIdentityProvidersCommand(args []string) error {
	cmd := flag.NewFlagSet("idp", flag.ContinueOnError)
	cmd.SetOutput(os.Stderr)
	cmd.Usage = PrintIdentityProviderUsage

	// Query flags
	listFlag := cmd.Bool("list", false, "List all identity providers")
	inspectFlag := cmd.String("inspect", "", "Inspect an identity provider by ID")
	outputFlag := cmd.String("output", "table", "Output format: table or json")

	// Create/Update/Delete flags
	createFlag := cmd.String("create", "", "Create a new identity provider (name)")
	updateFlag := cmd.String("update", "", "Update an identity provider by ID")
	deleteFlag := cmd.String("delete", "", "Delete an identity provider by ID")

	// Properties
	typeFlag := cmd.String("type", "oidc", "Provider type: oidc, zitadel, entra, google, okta, pocketid, microsoft")
	issuerFlag := cmd.String("issuer", "", "Issuer URL")
	clientIDFlag := cmd.String("client-id", "", "Client ID")
	clientSecretFlag := cmd.String("client-secret", "", "Client secret")
	nameFlag := cmd.String("name", "", "Provider name (for updates)")

	if len(args) == 1 {
		PrintIdentityProviderUsage()
		return nil
	}

	if err := cmd.Parse(args[1:]); err != nil {
		return nil
	}

	if *listFlag {
		return s.listIdentityProviders(*outputFlag)
	}

	if *inspectFlag != "" {
		return s.inspectIdentityProvider(*inspectFlag, *outputFlag)
	}

	if *createFlag != "" {
		if *issuerFlag == "" || *clientIDFlag == "" || *clientSecretFlag == "" {
			return fmt.Errorf("--issuer, --client-id, and --client-secret are required when creating an identity provider")
		}
		return s.createIdentityProvider(*createFlag, *typeFlag, *issuerFlag, *clientIDFlag, *clientSecretFlag)
	}

	if *updateFlag != "" {
		if *clientSecretFlag == "" {
			return fmt.Errorf("--client-secret is required when updating (the API does not return secrets, so it must be re-provided)")
		}
		// Track which flags were explicitly set
		setFlags := make(map[string]bool)
		cmd.Visit(func(f *flag.Flag) {
			setFlags[f.Name] = true
		})
		typeVal := ""
		if setFlags["type"] {
			typeVal = *typeFlag
		}
		return s.updateIdentityProvider(*updateFlag, *nameFlag, typeVal, *issuerFlag, *clientIDFlag, *clientSecretFlag)
	}

	if *deleteFlag != "" {
		return s.deleteIdentityProvider(*deleteFlag)
	}

	PrintIdentityProviderUsage()
	return nil
}

// listIdentityProviders lists all identity providers
func (s *Service) listIdentityProviders(outputFormat string) error {
	resp, err := s.Client.MakeRequest("GET", "/identity-providers", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var providers []models.IdentityProvider
	if err := json.NewDecoder(resp.Body).Decode(&providers); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	if len(providers) == 0 {
		fmt.Println("No identity providers found")
		return nil
	}

	if outputFormat == "json" {
		output, err := json.MarshalIndent(providers, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %v", err)
		}
		fmt.Println(string(output))
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tTYPE\tISSUER\tCLIENT ID")
	fmt.Fprintln(w, "--\t----\t----\t------\t---------")

	for _, provider := range providers {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			provider.ID,
			provider.Name,
			provider.Type,
			provider.Issuer,
			provider.ClientID,
		)
	}
	w.Flush()
	return nil
}

// inspectIdentityProvider retrieves details of a specific identity provider
func (s *Service) inspectIdentityProvider(idpID, outputFormat string) error {
	resp, err := s.Client.MakeRequest("GET", "/identity-providers/"+idpID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var provider models.IdentityProvider
	if err := json.NewDecoder(resp.Body).Decode(&provider); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	if outputFormat == "json" {
		output, err := json.MarshalIndent(provider, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %v", err)
		}
		fmt.Println(string(output))
		return nil
	}

	fmt.Printf("Identity Provider: %s (%s)\n", provider.Name, provider.ID)
	fmt.Println("---------------------------------")
	fmt.Printf("  Type:      %s\n", provider.Type)
	fmt.Printf("  Issuer:    %s\n", provider.Issuer)
	fmt.Printf("  Client ID: %s\n", provider.ClientID)
	return nil
}

// createIdentityProvider creates a new identity provider
func (s *Service) createIdentityProvider(name, providerType, issuer, clientID, clientSecret string) error {
	req := models.IdentityProviderRequest{
		Name:         name,
		Type:         providerType,
		Issuer:       issuer,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := s.Client.MakeRequest("POST", "/identity-providers", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var provider models.IdentityProvider
	if err := json.NewDecoder(resp.Body).Decode(&provider); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	fmt.Printf("✓ Identity provider created successfully!\n")
	fmt.Printf("  ID:        %s\n", provider.ID)
	fmt.Printf("  Name:      %s\n", provider.Name)
	fmt.Printf("  Type:      %s\n", provider.Type)
	fmt.Printf("  Issuer:    %s\n", provider.Issuer)
	fmt.Printf("  Client ID: %s\n", provider.ClientID)
	return nil
}

// updateIdentityProvider updates an existing identity provider
func (s *Service) updateIdentityProvider(idpID, name, providerType, issuer, clientID, clientSecret string) error {
	// Fetch current provider
	resp, err := s.Client.MakeRequest("GET", "/identity-providers/"+idpID, nil)
	if err != nil {
		return err
	}
	var current models.IdentityProvider
	if err := json.NewDecoder(resp.Body).Decode(&current); err != nil {
		resp.Body.Close()
		return fmt.Errorf("failed to decode current provider: %v", err)
	}
	resp.Body.Close()

	req := models.IdentityProviderRequest{
		Name:     current.Name,
		Type:     current.Type,
		Issuer:   current.Issuer,
		ClientID: current.ClientID,
	}

	if name != "" {
		req.Name = name
	}
	if providerType != "" {
		req.Type = providerType
	}
	if issuer != "" {
		req.Issuer = issuer
	}
	if clientID != "" {
		req.ClientID = clientID
	}
	if clientSecret != "" {
		req.ClientSecret = clientSecret
	}

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err = s.Client.MakeRequest("PUT", "/identity-providers/"+idpID, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var provider models.IdentityProvider
	if err := json.NewDecoder(resp.Body).Decode(&provider); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	fmt.Printf("✓ Identity provider updated successfully!\n")
	fmt.Printf("  ID:        %s\n", provider.ID)
	fmt.Printf("  Name:      %s\n", provider.Name)
	fmt.Printf("  Type:      %s\n", provider.Type)
	fmt.Printf("  Issuer:    %s\n", provider.Issuer)
	fmt.Printf("  Client ID: %s\n", provider.ClientID)
	return nil
}

// deleteIdentityProvider deletes an identity provider
func (s *Service) deleteIdentityProvider(idpID string) error {
	// Fetch provider details for confirmation
	resp, err := s.Client.MakeRequest("GET", "/identity-providers/"+idpID, nil)
	if err != nil {
		return err
	}
	var provider models.IdentityProvider
	if err := json.NewDecoder(resp.Body).Decode(&provider); err != nil {
		resp.Body.Close()
		return fmt.Errorf("failed to decode provider: %v", err)
	}
	resp.Body.Close()

	details := map[string]string{
		"Type":      provider.Type,
		"Issuer":    provider.Issuer,
		"Client ID": provider.ClientID,
	}

	if !helpers.ConfirmSingleDeletion("identity provider", provider.Name, provider.ID, details) {
		return nil
	}

	resp, err = s.Client.MakeRequest("DELETE", "/identity-providers/"+idpID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("✓ Identity provider removed successfully: %s\n", idpID)
	return nil
}
