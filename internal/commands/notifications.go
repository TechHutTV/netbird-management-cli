// notifications.go - Notification channel management (email/webhook on account events)
package commands

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"netbird-manage/internal/helpers"
	"netbird-manage/internal/models"
)

// HandleNotificationsCommand routes notification channel commands
func (s *Service) HandleNotificationsCommand(args []string) error {
	notifCmd := flag.NewFlagSet("notification", flag.ContinueOnError)
	notifCmd.SetOutput(os.Stderr)
	notifCmd.Usage = PrintNotificationUsage

	// Query flags
	listFlag := notifCmd.Bool("list", false, "List all notification channels")
	listTypesFlag := notifCmd.Bool("list-types", false, "List available event types")
	inspectFlag := notifCmd.String("inspect", "", "Inspect a notification channel by ID")

	// Modification flags
	createFlag := notifCmd.Bool("create", false, "Create a notification channel (requires --type and --event-types)")
	updateFlag := notifCmd.String("update", "", "Update a notification channel by ID")
	deleteFlag := notifCmd.String("delete", "", "Delete a notification channel by ID")

	// Channel parameters
	typeFlag := notifCmd.String("type", "", "Channel type: email or webhook")
	emailsFlag := notifCmd.String("emails", "", "Recipient emails (comma-separated, for --type email)")
	urlFlag := notifCmd.String("url", "", "Webhook URL (for --type webhook)")
	eventTypesFlag := notifCmd.String("event-types", "", "Event types to notify on (comma-separated, see --list-types)")
	enabledFlag := notifCmd.String("enabled", "", "Enable/disable the channel (true/false)")

	// Output format
	outputFlag := notifCmd.String("output", "table", "Output format: table or json")

	if len(args) == 1 {
		PrintNotificationUsage()
		return nil
	}

	if err := notifCmd.Parse(args[1:]); err != nil {
		return err
	}

	if *listTypesFlag {
		return s.listNotificationEventTypes(*outputFlag)
	}

	if *listFlag {
		return s.listNotificationChannels(*outputFlag)
	}

	if *inspectFlag != "" {
		return s.inspectNotificationChannel(*inspectFlag, *outputFlag)
	}

	if *createFlag {
		if *typeFlag == "" || *eventTypesFlag == "" {
			return fmt.Errorf("--create requires --type and --event-types")
		}
		req := models.NotificationChannelRequest{
			Type:       *typeFlag,
			EventTypes: helpers.SplitCommaList(*eventTypesFlag),
			Enabled:    true,
		}
		if err := parseBoolFlag(*enabledFlag, "enabled", &req.Enabled); err != nil {
			return err
		}
		target, err := buildNotificationTarget(*typeFlag, *emailsFlag, *urlFlag)
		if err != nil {
			return err
		}
		req.Target = target
		return s.createNotificationChannel(req)
	}

	if *updateFlag != "" {
		return s.updateNotificationChannel(*updateFlag, *typeFlag, *emailsFlag, *urlFlag, *eventTypesFlag, *enabledFlag)
	}

	if *deleteFlag != "" {
		return s.deleteNotificationChannel(*deleteFlag)
	}

	notifCmd.Usage()
	return nil
}

// buildNotificationTarget assembles the type-dependent target object
func buildNotificationTarget(channelType, emails, url string) (*models.NotificationTarget, error) {
	switch channelType {
	case "email":
		if emails == "" {
			return nil, fmt.Errorf("--type email requires --emails")
		}
		return &models.NotificationTarget{Emails: helpers.SplitCommaList(emails)}, nil
	case "webhook":
		if url == "" {
			return nil, fmt.Errorf("--type webhook requires --url")
		}
		return &models.NotificationTarget{URL: url}, nil
	}
	return nil, fmt.Errorf("invalid channel type %q: must be email or webhook", channelType)
}

// formatNotificationTarget renders a channel target for table output
func formatNotificationTarget(channel models.NotificationChannel) string {
	if channel.Target == nil {
		return "-"
	}
	if channel.Type == "email" {
		return strings.Join(channel.Target.Emails, ", ")
	}
	return channel.Target.URL
}

// listNotificationEventTypes lists the event types channels can subscribe to
func (s *Service) listNotificationEventTypes(outputFormat string) error {
	resp, err := s.Client.MakeRequest("GET", "/integrations/notifications/types", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var types map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&types); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	if outputFormat == "json" {
		output, err := json.MarshalIndent(types, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %v", err)
		}
		fmt.Println(string(output))
		return nil
	}

	codes := make([]string, 0, len(types))
	for code := range types {
		codes = append(codes, code)
	}
	sort.Strings(codes)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "EVENT TYPE\tDESCRIPTION")
	fmt.Fprintln(w, "----------\t-----------")
	for _, code := range codes {
		fmt.Fprintf(w, "%s\t%s\n", code, types[code])
	}
	w.Flush()
	return nil
}

// listNotificationChannels lists all notification channels
func (s *Service) listNotificationChannels(outputFormat string) error {
	resp, err := s.Client.MakeRequest("GET", "/integrations/notifications/channels", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var channels []models.NotificationChannel
	if err := json.NewDecoder(resp.Body).Decode(&channels); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	if len(channels) == 0 {
		if outputFormat == "json" {
			fmt.Println("[]")
		} else {
			fmt.Println("No notification channels found")
		}
		return nil
	}

	if outputFormat == "json" {
		output, err := json.MarshalIndent(channels, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %v", err)
		}
		fmt.Println(string(output))
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tTYPE\tTARGET\tEVENT TYPES\tENABLED")
	fmt.Fprintln(w, "--\t----\t------\t-----------\t-------")

	for _, channel := range channels {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%t\n",
			channel.ID,
			channel.Type,
			formatNotificationTarget(channel),
			strings.Join(channel.EventTypes, ", "),
			channel.Enabled,
		)
	}
	w.Flush()
	return nil
}

// getNotificationChannelByID fetches a single notification channel
func (s *Service) getNotificationChannelByID(channelID string) (*models.NotificationChannel, error) {
	resp, err := s.Client.MakeRequest("GET", "/integrations/notifications/channels/"+channelID, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var channel models.NotificationChannel
	if err := json.NewDecoder(resp.Body).Decode(&channel); err != nil {
		return nil, fmt.Errorf("failed to decode channel response: %v", err)
	}
	return &channel, nil
}

// inspectNotificationChannel shows detailed information about a channel
func (s *Service) inspectNotificationChannel(channelID string, outputFormat string) error {
	channel, err := s.getNotificationChannelByID(channelID)
	if err != nil {
		return err
	}

	if outputFormat == "json" {
		output, err := json.MarshalIndent(channel, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %v", err)
		}
		fmt.Println(string(output))
		return nil
	}

	fmt.Printf("Notification Channel: %s\n", channel.ID)
	fmt.Println("--------------------------------------------------")
	fmt.Printf("  Type:        %s\n", channel.Type)
	fmt.Printf("  Target:      %s\n", formatNotificationTarget(*channel))
	fmt.Printf("  Enabled:     %t\n", channel.Enabled)
	fmt.Printf("  Event Types: %s\n", strings.Join(channel.EventTypes, ", "))
	return nil
}

// createNotificationChannel creates a new notification channel
func (s *Service) createNotificationChannel(req models.NotificationChannelRequest) error {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := s.Client.MakeRequest("POST", "/integrations/notifications/channels", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var channel models.NotificationChannel
	if err := json.NewDecoder(resp.Body).Decode(&channel); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	fmt.Printf("Notification channel created successfully\n")
	fmt.Printf("  ID:     %s\n", channel.ID)
	fmt.Printf("  Type:   %s\n", channel.Type)
	fmt.Printf("  Target: %s\n", formatNotificationTarget(channel))
	return nil
}

// updateNotificationChannel updates a channel, preserving unset fields
func (s *Service) updateNotificationChannel(channelID, channelType, emails, url, eventTypes, enabledValue string) error {
	current, err := s.getNotificationChannelByID(channelID)
	if err != nil {
		return err
	}

	req := models.NotificationChannelRequest{
		Type:       current.Type,
		Target:     current.Target,
		EventTypes: current.EventTypes,
		Enabled:    current.Enabled,
	}
	if channelType != "" {
		req.Type = channelType
	}
	if emails != "" || url != "" {
		target, err := buildNotificationTarget(req.Type, emails, url)
		if err != nil {
			return err
		}
		req.Target = target
	}
	if eventTypes != "" {
		req.EventTypes = helpers.SplitCommaList(eventTypes)
	}
	if err := parseBoolFlag(enabledValue, "enabled", &req.Enabled); err != nil {
		return err
	}

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := s.Client.MakeRequest("PUT", "/integrations/notifications/channels/"+channelID, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("Notification channel %s updated successfully\n", channelID)
	return nil
}

// deleteNotificationChannel deletes a notification channel
func (s *Service) deleteNotificationChannel(channelID string) error {
	channel, err := s.getNotificationChannelByID(channelID)
	if err != nil {
		return err
	}

	details := map[string]string{
		"Type":        channel.Type,
		"Target":      formatNotificationTarget(*channel),
		"Event Types": strings.Join(channel.EventTypes, ", "),
	}

	if !helpers.ConfirmSingleDeletion("notification channel", channel.Type, channelID, details) {
		return nil
	}

	resp, err := s.Client.MakeRequest("DELETE", "/integrations/notifications/channels/"+channelID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("Notification channel %s deleted successfully\n", channelID)
	return nil
}
