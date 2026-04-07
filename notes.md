# TUI Development Notes

## Branch: feat/tui

## Implemented

### Phase 1: Shell & Navigation
- `internal/tui/app.go` — Root model (Elm architecture), page switching, focus management
- `internal/tui/nav.go` — Sidebar with 13 sections, cursor nav (j/k), selection (enter)
- `internal/tui/layout.go` — Two-column layout (sidebar + content), status bar, resize handling
- `internal/tui/theme.go` — NetBird-inspired color palette (orange primary, blue secondary)
- `internal/tui/keys.go` — Key binding constants
- `internal/tui/run.go` — Entry point: `tui.Run(client)` launches `tea.Program`
- `internal/tui/placeholder.go` — Temporary page for unimplemented sections
- `cmd/netbird-manage/main.go` — Wired `tui` subcommand

### Phase 2: Infrastructure
- `internal/tui/messages.go` — Typed messages for all API responses, toasts, errors
- `internal/tui/api.go` — Async `tea.Cmd` factories for all entities (fetch + delete)
- `internal/tui/components/detail.go` — Key-value detail renderer
- `internal/tui/components/statusbar.go` — Status bar with key hints

### Phase 3: Peers Page (reference implementation)
- `internal/tui/peers_page.go` — Full peers page:
  - Table list with scrolling, cursor highlight
  - Online/offline status badges (green/red)
  - Detail view with all peer fields
  - Delete via `d` key with API call + auto-refresh
  - Refresh via `r` key
  - Enter for detail, esc/backspace to go back

### Phase 4: Simple Entity Pages
- `internal/tui/groups_page.go` — List + detail with group members, delete
- `internal/tui/setup_keys_page.go` — List with state badges (valid/expired/revoked), detail, delete
- `internal/tui/users_page.go` — List showing email/role/status/service/blocked, detail
- `internal/tui/helpers.go` — Shared JSON decode helper

### Phase 5: Medium Entity Pages
- `internal/tui/routes_page.go` — List with network/type/metric/masq/enabled, detail, delete
- `internal/tui/dns_page.go` — List with server count/domains/primary/enabled, detail, delete
- `internal/tui/accounts_page.go` — Settings display (login exp, DNS domain, JWT, etc.)
- `internal/tui/ingress_page.go` — Cloud-only with graceful 404 message

### Phase 6: Complex Entity Pages
- `internal/tui/networks_page.go` — List + detail with resources/routers sub-tables, delete
- `internal/tui/policies_page.go` — List + detail with full rule breakdown (sources/dest/protocol/ports), delete
- `internal/tui/posture_checks_page.go` — List with auto-detected type, detail with type-specific fields

### Phase 7: Special Pages
- `internal/tui/events_page.go` — Scrollable audit log table with timestamp/activity/initiator/target
- `internal/tui/export_import_page.go` — Info page pointing to CLI commands (TUI export coming later)

### Inline Actions (dashboard parity)
- **Peers**: `/` search (filter by name/IP/hostname), `a` accessible peers view, `d` delete
- **Routes**: `t` toggle enable/disable, `d` delete
- **Policies**: `t` toggle enable/disable, `d` delete
- **DNS**: `t` toggle enable/disable, `d` delete
- **Setup Keys**: `v` revoke/un-revoke, `d` delete
- **Users**: `b` block/unblock
- API helpers in `api.go`: `ToggleRoute`, `TogglePolicy`, `RevokeSetupKey`, `ToggleUserBlock`, `FetchAccessiblePeers`

### Create Forms (huh/v2)
- `internal/tui/forms.go` — All form data types, constructors, submit commands, and helpers
- **Groups**: `c` creates group (name)
- **Setup Keys**: `c` creates key (name, type, expiry duration, usage limit, ephemeral)
- **Routes**: `c` creates route (network ID, CIDR, peer groups, distribution groups, metric, masquerade, enabled)
- **DNS**: `c` creates DNS group (name, nameservers, groups, match domains, search domains, primary)
- **Networks**: `c` creates network (name, description)
- **Policies**: `c` creates policy with initial rule (name, description, protocol, action, ports, sources, destinations)
- **Users**: `c` invites user or creates service user (name, email, role, auto-groups, service flag)
- Forms auto-submit on completion, auto-cancel on esc, auto-refresh list after API success

### Numbered Top Bar + Dashboard (netbird-tui inspired)
- `internal/tui/nav.go` — Numbered hotkeys `[1]Status [2]Peers ... [9]DNS [0]Posture`
  - Hotkeys 1-9 and 0 work globally from both nav and content focus
  - `SectionDashboard` added as first section, all others renumbered
  - Labels shortened: Setup Keys→Keys, Posture Checks→Posture, Export/Import→Export
- `internal/tui/daemon.go` — gRPC daemon client wrapper
  - Connects to `unix:///var/run/netbird.sock` (configurable via `NETBIRD_SOCKET` env)
  - Calls `Status()` and `GetConfig()` on the NetBird daemon
  - Graceful nil return if connection fails (TUI still works in management-only mode)
- `internal/tui/dashboard_page.go` — Two-column dashboard
  - **Left**: Daemon status (management/signal connection, local peer IP/FQDN, kernel, Rosenpass, relays, DNS servers, SSH server, last 3 system events)
  - **Right**: Management API overview (account domain, peer/group/network/policy/route/key/DNS/posture counts)
  - Auto-refresh daemon status every 5 seconds
  - Graceful "Daemon not connected" if netbird isn't running
- `internal/tui/messages.go` — Added `DaemonStatusMsg`, `DaemonTickMsg`, `DashboardCountsMsg`
- `internal/tui/run.go` — Creates daemon client, passes to `NewApp(httpClient, daemon)`
- Dependencies: `github.com/netbirdio/netbird` v0.66.2, `google.golang.org/grpc` v1.79.3

## Still TODO (Phase 8: Polish)
- Help overlay (`?` key)
- Sidebar collapse for narrow terminals
- Count badges on nav items (loaded async on startup)
- Confirmation dialog for destructive actions
- Toast notifications with auto-dismiss
- Edit forms (pre-populate form with existing resource data)
- Search/filter on remaining pages (currently only Peers has it)

## Architecture Decisions

- **All page files in `internal/tui/` package** — avoids import cycles between tui and tui/pages
- **`charm.land/*` import paths** (not `github.com/charmbracelet/*`) — v2 modules moved to charm.land
- **`Init()` returns `tea.Cmd`** (not `(tea.Model, tea.Cmd)`) — bubbletea v2.0.2 API
- **`View()` returns `tea.View`** — use `tea.NewView(s)` with `v.AltScreen = true`
- **`tea.KeyPressMsg`** for key events (not `tea.KeyMsg`)
- **Immutable page updates** — pages map is copied on every update to avoid mutation
- **Sidebar focus model** — tab toggles between sidebar and content focus; q/esc in content returns to sidebar

## Styling Redesign (Phase 8 partial)

Inspired by charmbracelet/soft-serve, dlvhdr/gh-dash, and lipgloss layout examples:

- **lipgloss v2 static tables** (`charm.land/lipgloss/v2/table`) with `RoundedBorder()`, `StyleFunc` for alternating rows, cursor highlights, and semantic column coloring (green/red for status)
- **Adaptive colors** — `lipgloss.HasDarkBackground()` detects terminal theme, all colors adapt
- **Sidebar** — thin `│` right border separator (not a full box), `┃` left bar accent on active/cursor items (soft-serve pattern)
- **Segmented status bar** — colored section badge + help hints (powerline-style)
- **Detail views** — right-aligned labels + left-aligned values, `─` underline on titles, section sub-headers for nested data
- **Status badges** — ONLINE/OFFLINE, VALID/EXPIRED/REVOKED, ENABLED/DISABLED with semantic colors
- **Consistent spacing** — `Padding(0, 1)` on cells, `Padding(1, 2)` on content area

### v2 API Notes (lipgloss)
- `lipgloss.Color()` is a **function** (not a type) returning `color.Color`
- Use `color.Color` as the type for color variables
- `table.New()` returns `*Table`, chain `.Border().Headers().Rows().Width().StyleFunc()`
- `table.HeaderRow` constant for `StyleFunc` row == header detection

## Dependencies Added
- `charm.land/bubbletea/v2` v2.0.2
- `charm.land/bubbles/v2` v2.1.0
- `charm.land/lipgloss/v2` v2.0.2
