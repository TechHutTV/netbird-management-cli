package tui

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

const (
	peersRefreshInterval      = 5 * time.Second
	eventsRefreshInterval     = 10 * time.Second
	connectionRefreshInterval = 5 * time.Second
	pageRefreshInterval       = 60 * time.Second
)

func peersTickCmd() tea.Cmd {
	return tea.Tick(peersRefreshInterval, func(time.Time) tea.Msg {
		return PeersTickMsg{}
	})
}

func eventsTickCmd() tea.Cmd {
	return tea.Tick(eventsRefreshInterval, func(time.Time) tea.Msg {
		return EventsTickMsg{}
	})
}

func connectionTickCmd() tea.Cmd {
	return tea.Tick(connectionRefreshInterval, func(time.Time) tea.Msg {
		return ConnectionTickMsg{}
	})
}

func pageRefreshTickCmd() tea.Cmd {
	return tea.Tick(pageRefreshInterval, func(time.Time) tea.Msg {
		return PageRefreshTickMsg{}
	})
}
