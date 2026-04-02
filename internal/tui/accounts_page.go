package tui

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"netbird-manage/internal/client"
	"netbird-manage/internal/models"
)

// AccountsPage displays account settings
type AccountsPage struct {
	accounts []models.Account
	loading  bool
	err      error
	focused  bool
}

func NewAccountsPage() *AccountsPage {
	return &AccountsPage{loading: true}
}

func (a *AccountsPage) Title() string { return "Account" }
func (a *AccountsPage) CursorPosition() int { return 0 }
func (a *AccountsPage) SetFocused(focused bool) { a.focused = focused }

func (a *AccountsPage) Init(c *client.Client) tea.Cmd {
	if len(a.accounts) > 0 {
		return nil
	}
	a.loading = true
	a.err = nil
	return FetchAccounts(c)
}

func (a *AccountsPage) Update(msg tea.Msg, c *client.Client) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case AccountsLoadedMsg:
		a.loading = false
		if msg.Err != nil {
			a.err = msg.Err
			return a, nil
		}
		a.accounts = msg.Accounts
		return a, nil

	case APIErrorMsg:
		a.err = msg.Err
		return a, nil

	case tea.KeyPressMsg:
		if msg.String() == "r" {
			a.loading = true
			return a, FetchAccounts(c)
		}
	}

	return a, nil
}

func (a *AccountsPage) View(width, height int) string {
	if a.loading {
		return loadingStyle.Render("  Loading account settings...")
	}
	if a.err != nil {
		return errorStyle.Render("  Error: " + a.err.Error())
	}

	if len(a.accounts) == 0 {
		return dimHintStyle.Render("  No accounts found.")
	}

	var b strings.Builder

	for _, acct := range a.accounts {
		s := acct.Settings

		loginExp := formatDuration(s.PeerLoginExpiration)
		inactivityExp := formatDuration(s.PeerInactivityExpiration)

		b.WriteString(detailTitleStyle.Render("  Account: "+acct.Domain) + "\n\n")

		fields := []struct{ label, value string }{
			{"Account ID", acct.ID},
			{"Domain", acct.Domain},
			{"Login Expiration", loginExp},
			{"Inactivity Exp.", inactivityExp},
			{"DNS Domain", s.DNSDomain},
			{"Network Range", s.NetworkRange},
			{"JWT Groups", fmt.Sprintf("%v", s.JWTGroupsEnabled)},
			{"JWT Claim", s.JWTGroupsClaim},
			{"Groups Propagation", fmt.Sprintf("%v", s.GroupsPropagationEnabled)},
			{"Users View Blocked", fmt.Sprintf("%v", s.RegularUsersViewBlocked)},
			{"Peer Approval", fmt.Sprintf("%v", s.PeerApprovalEnabled)},
			{"Traffic Logging", fmt.Sprintf("%v", s.TrafficLogging)},
		}

		for _, f := range fields {
			label := detailLabelStyle.Render(f.label)
			value := detailValueStyle.Render(f.value)
			b.WriteString(fmt.Sprintf("%s  %s\n", label, value))
		}
	}

	b.WriteString("\n" + dimHintStyle.Render("  r: refresh"))

	return b.String()
}

func formatDuration(seconds int) string {
	if seconds <= 0 {
		return "disabled"
	}
	d := time.Duration(seconds) * time.Second
	if d >= 24*time.Hour {
		days := int(d.Hours()) / 24
		return fmt.Sprintf("%dd", days)
	}
	if d >= time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dm", int(d.Minutes()))
}
