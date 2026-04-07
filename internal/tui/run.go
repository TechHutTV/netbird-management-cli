package tui

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"netbird-manage/internal/client"
)

// Run starts the TUI application
func Run(c *client.Client) {
	// Connect to local daemon (optional — dashboard degrades gracefully)
	socketPath := os.Getenv("NETBIRD_SOCKET")
	if socketPath == "" {
		socketPath = DefaultDaemonSocket
	}
	daemon := NewDaemonClient(socketPath)

	app := NewApp(c, daemon)

	p := tea.NewProgram(app)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}

	if daemon != nil {
		daemon.Close()
	}
}
