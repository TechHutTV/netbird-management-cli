# TUI New Features Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add search filtering, edit forms, connection detail, live monitoring, and bulk group assignment to the NetBird TUI.

**Architecture:** Extends the existing Elm-architecture TUI (bubbletea v2) with five features implemented in dependency order. Each feature builds on existing patterns — search extends the Peers search pattern, edit forms extend the create form infrastructure, connection detail extends the daemon client, live monitoring adds tick commands, and bulk group assignment adds a new multi-select interaction.

**Tech Stack:** Go 1.25+, charm.land/bubbletea/v2, charm.land/huh/v2, charm.land/lipgloss/v2, google.golang.org/grpc

**Spec:** `docs/superpowers/specs/2026-04-02-tui-features-design.md`

---

## File Structure

### New Files
- `internal/tui/search.go` — Reusable search state and filter helpers (~60 lines)
- `internal/tui/search_test.go` — Table-driven tests for search matching
- `internal/tui/bulk.go` — Bulk select state machine for peers (~150 lines)
- `internal/tui/live.go` — Auto-refresh tick helpers and message types (~40 lines)

### Modified Files
- `internal/tui/messages.go` — New message types for edit, bulk, and tick
- `internal/tui/api.go` — Update/PUT command factories for all editable resources
- `internal/tui/forms.go` — Edit form constructors (pre-populated variants)
- `internal/tui/keys.go` — New key constants (edit, bulk, pause)
- `internal/tui/daemon.go` — Extended peer connection detail from daemon
- `internal/tui/helpers.go` — Byte formatting helper
- `internal/tui/groups_page.go` — Search + edit
- `internal/tui/networks_page.go` — Search + edit
- `internal/tui/policies_page.go` — Search + edit
- `internal/tui/routes_page.go` — Search + edit
- `internal/tui/setup_keys_page.go` — Search (no edit — immutable after creation)
- `internal/tui/users_page.go` — Search + edit
- `internal/tui/service_users_page.go` — Search
- `internal/tui/posture_checks_page.go` — Search + edit + create + delete
- `internal/tui/events_page.go` — Search + live auto-refresh
- `internal/tui/ingress_page.go` — Search
- `internal/tui/peers_page.go` — Connection detail, live refresh, bulk select
- `internal/tui/app.go` — Pass daemon to page updates for connection detail

---

## Phase 1: Search on All List Pages

### Task 1: Extract reusable search infrastructure

**Files:**
- Create: `internal/tui/search.go`
- Create: `internal/tui/search_test.go`

- [ ] **Step 1: Write failing tests for search matching**

```go
// internal/tui/search_test.go
package tui

import "testing"

func TestMatchesQuery(t *testing.T) {
	tests := []struct {
		name   string
		query  string
		fields []string
		want   bool
	}{
		{
			name:   "empty query matches everything",
			query:  "",
			fields: []string{"foo"},
			want:   true,
		},
		{
			name:   "exact match",
			query:  "prod",
			fields: []string{"production"},
			want:   true,
		},
		{
			name:   "case insensitive",
			query:  "PROD",
			fields: []string{"production"},
			want:   true,
		},
		{
			name:   "no match",
			query:  "staging",
			fields: []string{"production", "main"},
			want:   false,
		},
		{
			name:   "matches any field",
			query:  "admin",
			fields: []string{"web-server", "admin@co.com"},
			want:   true,
		},
		{
			name:   "partial match",
			query:  "serv",
			fields: []string{"web-server"},
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesQuery(tt.query, tt.fields...)
			if got != tt.want {
				t.Errorf("matchesQuery(%q, %v) = %v, want %v", tt.query, tt.fields, got, tt.want)
			}
		})
	}
}

func TestClampCursor(t *testing.T) {
	tests := []struct {
		name   string
		cursor int
		length int
		want   int
	}{
		{"within range", 3, 10, 3},
		{"at end", 9, 10, 9},
		{"past end", 15, 10, 9},
		{"empty list", 5, 0, 0},
		{"negative", -1, 10, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clampCursor(tt.cursor, tt.length)
			if got != tt.want {
				t.Errorf("clampCursor(%d, %d) = %d, want %d", tt.cursor, tt.length, got, tt.want)
			}
		})
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/jack/Development/netbird-management-cli && go test -race ./internal/tui/ -run "TestMatchesQuery|TestClampCursor" -v`
Expected: FAIL — `matchesQuery` and `clampCursor` undefined

- [ ] **Step 3: Implement search helpers**

```go
// internal/tui/search.go
package tui

import "strings"

// matchesQuery returns true if any field contains the query (case-insensitive).
// An empty query matches everything.
func matchesQuery(query string, fields ...string) bool {
	if query == "" {
		return true
	}
	q := strings.ToLower(query)
	for _, f := range fields {
		if strings.Contains(strings.ToLower(f), q) {
			return true
		}
	}
	return false
}

// clampCursor ensures cursor stays within [0, length-1].
// Returns 0 for empty lists or negative cursors.
func clampCursor(cursor, length int) int {
	if length == 0 || cursor < 0 {
		return 0
	}
	if cursor >= length {
		return length - 1
	}
	return cursor
}

// handleSearchKey processes a key press during active search input.
// Returns the updated search string, whether search is still active, and whether the filter changed.
func handleSearchKey(key string, search string) (newSearch string, stillSearching bool, changed bool) {
	switch key {
	case "esc":
		return "", false, search != ""
	case "enter":
		return search, false, false
	case "backspace":
		if len(search) > 0 {
			return search[:len(search)-1], true, true
		}
		return search, true, false
	default:
		if len(key) == 1 {
			return search + key, true, true
		}
		return search, true, false
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/jack/Development/netbird-management-cli && go test -race ./internal/tui/ -run "TestMatchesQuery|TestClampCursor" -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/tui/search.go internal/tui/search_test.go
git commit -m "feat(tui): extract reusable search helpers"
```

---

### Task 2: Add search to Groups page

**Files:**
- Modify: `internal/tui/groups_page.go`

This task establishes the pattern that all subsequent pages will follow.

- [ ] **Step 1: Add search fields to GroupsPage struct**

Add `search`, `searching`, and `filtered` fields to `GroupsPage`:

```go
// In GroupsPage struct, add these fields:
search    string
searching bool
```

Also add `filtered []models.PolicyGroup` if not already present.

- [ ] **Step 2: Add applyFilter method**

```go
func (g *GroupsPage) applyFilter() {
	if g.search == "" {
		g.filtered = g.groups
	} else {
		filtered := make([]models.PolicyGroup, 0)
		for _, group := range g.groups {
			if matchesQuery(g.search, group.Name) {
				filtered = append(filtered, group)
			}
		}
		g.filtered = filtered
	}
	g.cursor = clampCursor(g.cursor, len(g.filtered))
}
```

- [ ] **Step 3: Handle search keys in Update**

In `handleKey`, add search input handling at the top (before other key handling), and add `/` key to enter search mode:

```go
// At top of handleKey, before existing key handling:
if g.searching {
	newSearch, still, changed := handleSearchKey(key, g.search)
	g.search = newSearch
	g.searching = still
	if changed || !still {
		g.applyFilter()
	}
	return g, nil
}

// In the list view switch statement, add:
case "/":
	g.searching = true
	g.search = ""
```

Also update the `esc` case in the list view to clear search:
```go
case "esc":
	if g.search != "" {
		g.search = ""
		g.cursor = 0
		g.applyFilter()
	}
```

- [ ] **Step 4: Update GroupsLoadedMsg handler to apply filter**

In the `GroupsLoadedMsg` case, after setting `g.groups`, call `g.applyFilter()`:

```go
case GroupsLoadedMsg:
	g.loading = false
	if msg.Err != nil {
		g.err = msg.Err
		return g, nil
	}
	g.groups = msg.Groups
	g.applyFilter()
	return g, nil
```

- [ ] **Step 5: Add search bar rendering in View**

In the list view rendering method, add the search bar display above the table (same pattern as peers):

```go
if g.searching {
	b.WriteString(sectionHeaderStyle.Render("  / ") + detailValueStyle.Render(g.search) + "\u2588\n")
} else if g.search != "" {
	b.WriteString(dimHintStyle.Render(fmt.Sprintf("  filter: %s", g.search)) + "\n")
}
```

- [ ] **Step 6: Build and manually test**

Run: `cd /Users/jack/Development/netbird-management-cli && go build ./cmd/netbird-manage/`
Expected: Builds successfully. Test by running `./netbird-manage tui`, navigate to Groups, press `/`, type a query, verify filtering works.

- [ ] **Step 7: Commit**

```bash
git add internal/tui/groups_page.go
git commit -m "feat(tui): add search filtering to groups page"
```

---

### Task 3: Add search to Networks, Policies, Routes pages

**Files:**
- Modify: `internal/tui/networks_page.go`
- Modify: `internal/tui/policies_page.go`
- Modify: `internal/tui/routes_page.go`

Apply the same pattern from Task 2 to each page. For each page:
1. Add `search string`, `searching bool` fields to the struct
2. Add `applyFilter()` method with page-specific `matchesQuery` fields
3. Add search key handling at top of `handleKey`
4. Update the `*LoadedMsg` handler to call `applyFilter()`
5. Add search bar rendering in the list view

- [ ] **Step 1: Add search to NetworksPage**

NetworksPage filter uses `matchesQuery(n.search, net.Name, net.Description)`.

Add `search string` and `searching bool` to the struct. Add `applyFilter()`:

```go
func (n *NetworksPage) applyFilter() {
	if n.search == "" {
		n.filtered = n.networks
	} else {
		filtered := make([]models.Network, 0)
		for _, net := range n.networks {
			if matchesQuery(n.search, net.Name, net.Description) {
				filtered = append(filtered, net)
			}
		}
		n.filtered = filtered
	}
	n.cursor = clampCursor(n.cursor, len(n.filtered))
}
```

Add search key handling and rendering following the Groups pattern from Task 2.

- [ ] **Step 2: Add search to PoliciesPage**

PoliciesPage filter uses `matchesQuery(p.search, pol.Name)`.

Same pattern — add fields, `applyFilter()`, key handling, rendering.

```go
func (p *PoliciesPage) applyFilter() {
	if p.search == "" {
		p.filtered = p.policies
	} else {
		filtered := make([]models.Policy, 0)
		for _, pol := range p.policies {
			if matchesQuery(p.search, pol.Name) {
				filtered = append(filtered, pol)
			}
		}
		p.filtered = filtered
	}
	p.cursor = clampCursor(p.cursor, len(p.filtered))
}
```

- [ ] **Step 3: Add search to RoutesPage**

RoutesPage filter uses `matchesQuery(r.search, route.NetworkID, route.Description)`.

Same pattern.

```go
func (r *RoutesPage) applyFilter() {
	if r.search == "" {
		r.filtered = r.routes
	} else {
		filtered := make([]models.Route, 0)
		for _, route := range r.routes {
			if matchesQuery(r.search, route.NetworkID, route.Description) {
				filtered = append(filtered, route)
			}
		}
		r.filtered = filtered
	}
	r.cursor = clampCursor(r.cursor, len(r.filtered))
}
```

- [ ] **Step 4: Build and test**

Run: `cd /Users/jack/Development/netbird-management-cli && go build ./cmd/netbird-manage/`
Expected: Builds successfully.

- [ ] **Step 5: Commit**

```bash
git add internal/tui/networks_page.go internal/tui/policies_page.go internal/tui/routes_page.go
git commit -m "feat(tui): add search to networks, policies, and routes pages"
```

---

### Task 4: Add search to Setup Keys, Users, Service Users pages

**Files:**
- Modify: `internal/tui/setup_keys_page.go`
- Modify: `internal/tui/users_page.go`
- Modify: `internal/tui/service_users_page.go`

- [ ] **Step 1: Add search to SetupKeysPage**

Filter uses `matchesQuery(s.search, key.Name, key.State)`.

Same pattern as Task 2 — add fields, `applyFilter()`, key handling, rendering.

```go
func (s *SetupKeysPage) applyFilter() {
	if s.search == "" {
		s.filtered = s.keys
	} else {
		filtered := make([]models.SetupKey, 0)
		for _, key := range s.keys {
			if matchesQuery(s.search, key.Name, key.State) {
				filtered = append(filtered, key)
			}
		}
		s.filtered = filtered
	}
	s.cursor = clampCursor(s.cursor, len(s.filtered))
}
```

- [ ] **Step 2: Add search to UsersPage**

Filter uses `matchesQuery(u.search, user.Name, user.Email, user.Role)`.

```go
func (u *UsersPage) applyFilter() {
	if u.search == "" {
		u.filtered = u.users
	} else {
		filtered := make([]models.User, 0)
		for _, user := range u.users {
			if matchesQuery(u.search, user.Name, user.Email, user.Role) {
				filtered = append(filtered, user)
			}
		}
		u.filtered = filtered
	}
	u.cursor = clampCursor(u.cursor, len(u.filtered))
}
```

- [ ] **Step 3: Add search to ServiceUsersPage**

Filter uses `matchesQuery(s.search, user.Name)`.

Same pattern.

- [ ] **Step 4: Build and test**

Run: `cd /Users/jack/Development/netbird-management-cli && go build ./cmd/netbird-manage/`

- [ ] **Step 5: Commit**

```bash
git add internal/tui/setup_keys_page.go internal/tui/users_page.go internal/tui/service_users_page.go
git commit -m "feat(tui): add search to setup keys, users, and service users pages"
```

---

### Task 5: Add search to Posture Checks, Events, Ingress pages

**Files:**
- Modify: `internal/tui/posture_checks_page.go`
- Modify: `internal/tui/events_page.go`
- Modify: `internal/tui/ingress_page.go`

- [ ] **Step 1: Add search to PostureChecksPage**

Filter uses `matchesQuery(p.search, check.Name)`.

Same pattern as Task 2.

- [ ] **Step 2: Add search to EventsPage**

Filter uses `matchesQuery(e.search, event.Activity, event.InitiatorName, event.TargetID)`.

Events page is read-only with no detail view, but search works the same way on the list.

```go
func (e *EventsPage) applyFilter() {
	if e.search == "" {
		e.filtered = e.events
	} else {
		filtered := make([]models.AuditEvent, 0)
		for _, event := range e.events {
			if matchesQuery(e.search, event.Activity, event.InitiatorName, event.TargetID) {
				filtered = append(filtered, event)
			}
		}
		e.filtered = filtered
	}
	e.cursor = clampCursor(e.cursor, len(e.filtered))
}
```

- [ ] **Step 3: Add search to IngressPage**

Filter on peer name. Same pattern.

- [ ] **Step 4: Build and test**

Run: `cd /Users/jack/Development/netbird-management-cli && go build ./cmd/netbird-manage/`

- [ ] **Step 5: Commit**

```bash
git add internal/tui/posture_checks_page.go internal/tui/events_page.go internal/tui/ingress_page.go
git commit -m "feat(tui): add search to posture checks, events, and ingress pages"
```

---

### Task 6: Refactor Peers page to use shared search helpers

**Files:**
- Modify: `internal/tui/peers_page.go`

The Peers page already has search but uses inline logic. Refactor it to use the shared `matchesQuery`, `clampCursor`, and `handleSearchKey` helpers for consistency.

- [ ] **Step 1: Replace inline search logic with shared helpers**

In `applyFilter()`, replace the inline `strings.Contains`/`strings.ToLower` logic with `matchesQuery`:

```go
func (p *PeersPage) applyFilter() {
	if p.search == "" {
		p.filtered = p.peers
	} else {
		filtered := make([]models.Peer, 0)
		for _, peer := range p.peers {
			if matchesQuery(p.search, peer.Name, peer.IP, peer.Hostname) {
				filtered = append(filtered, peer)
			}
		}
		p.filtered = filtered
	}
	p.cursor = clampCursor(p.cursor, len(p.filtered))
}
```

In the search key handling, replace the inline switch with `handleSearchKey`:

```go
if p.searching {
	newSearch, still, changed := handleSearchKey(key, p.search)
	p.search = newSearch
	p.searching = still
	if changed || !still {
		p.applyFilter()
	}
	return p, nil
}
```

- [ ] **Step 2: Build and test**

Run: `cd /Users/jack/Development/netbird-management-cli && go build ./cmd/netbird-manage/`
Verify peers search still works identically.

- [ ] **Step 3: Commit**

```bash
git add internal/tui/peers_page.go
git commit -m "refactor(tui): use shared search helpers in peers page"
```

---

## Phase 2: Edit Forms

### Task 7: Add edit message types and API update commands

**Files:**
- Modify: `internal/tui/messages.go`
- Modify: `internal/tui/api.go`
- Modify: `internal/tui/keys.go`

- [ ] **Step 1: Add new key constants**

In `internal/tui/keys.go`, add:

```go
const (
	keyBulkSelect = "g"
	keyPause      = "p"
)
```

Note: `keyEdit = "e"` already exists in keys.go.

- [ ] **Step 2: Add edit result message types**

In `internal/tui/messages.go`, add:

```go
// Edit operation results
type PeerUpdatedMsg struct {
	Peer models.Peer
	Err  error
}

type GroupUpdatedMsg struct {
	Err error
}

type PolicyUpdatedMsg struct {
	Err error
}

type RouteUpdatedMsg struct {
	Err error
}

type DNSUpdatedMsg struct {
	Err error
}

type NetworkUpdatedMsg struct {
	Err error
}

type UserUpdatedMsg struct {
	Err error
}

type PostureCheckUpdatedMsg struct {
	Err error
}

type PostureCheckDeletedMsg struct {
	Err error
}
```

- [ ] **Step 3: Add API update commands for Peers**

In `internal/tui/api.go`, add:

```go
// UpdatePeer updates a peer via PUT
func UpdatePeer(c *client.Client, peerID string, req models.PeerUpdateRequest) tea.Cmd {
	return func() tea.Msg {
		body, err := json.Marshal(req)
		if err != nil {
			return PeerUpdatedMsg{Err: fmt.Errorf("marshal request: %w", err)}
		}
		resp, err := c.MakeRequest("PUT", "/peers/"+url.PathEscape(peerID), bytes.NewReader(body))
		if err != nil {
			return PeerUpdatedMsg{Err: err}
		}
		defer resp.Body.Close()

		var peer models.Peer
		if err := json.NewDecoder(resp.Body).Decode(&peer); err != nil {
			return PeerUpdatedMsg{Err: fmt.Errorf("decode peer: %w", err)}
		}
		return PeerUpdatedMsg{Peer: peer}
	}
}
```

- [ ] **Step 4: Add API update commands for Groups**

```go
// UpdateGroup updates a group name via PUT
func UpdateGroup(c *client.Client, groupID string, name string) tea.Cmd {
	return func() tea.Msg {
		req := struct {
			Name string `json:"name"`
		}{Name: name}
		body, err := json.Marshal(req)
		if err != nil {
			return GroupUpdatedMsg{Err: fmt.Errorf("marshal request: %w", err)}
		}
		resp, err := c.MakeRequest("PUT", "/groups/"+url.PathEscape(groupID), bytes.NewReader(body))
		if err != nil {
			return GroupUpdatedMsg{Err: err}
		}
		defer resp.Body.Close()
		return GroupUpdatedMsg{}
	}
}
```

- [ ] **Step 5: Add API update commands for Policies, Routes, DNS, Networks, Users**

```go
// UpdatePolicy updates a policy via PUT
func UpdatePolicy(c *client.Client, policyID string, req models.PolicyUpdateRequest) tea.Cmd {
	return func() tea.Msg {
		body, err := json.Marshal(req)
		if err != nil {
			return PolicyUpdatedMsg{Err: fmt.Errorf("marshal request: %w", err)}
		}
		resp, err := c.MakeRequest("PUT", "/policies/"+url.PathEscape(policyID), bytes.NewReader(body))
		if err != nil {
			return PolicyUpdatedMsg{Err: err}
		}
		defer resp.Body.Close()
		return PolicyUpdatedMsg{}
	}
}

// UpdateRoute updates a route via PUT
func UpdateRoute(c *client.Client, routeID string, req models.RouteRequest) tea.Cmd {
	return func() tea.Msg {
		body, err := json.Marshal(req)
		if err != nil {
			return RouteUpdatedMsg{Err: fmt.Errorf("marshal request: %w", err)}
		}
		resp, err := c.MakeRequest("PUT", "/routes/"+url.PathEscape(routeID), bytes.NewReader(body))
		if err != nil {
			return RouteUpdatedMsg{Err: err}
		}
		defer resp.Body.Close()
		return RouteUpdatedMsg{}
	}
}

// UpdateDNSGroup updates a DNS nameserver group via PUT
func UpdateDNSGroup(c *client.Client, groupID string, req models.DNSNameserverGroupRequest) tea.Cmd {
	return func() tea.Msg {
		body, err := json.Marshal(req)
		if err != nil {
			return DNSUpdatedMsg{Err: fmt.Errorf("marshal request: %w", err)}
		}
		resp, err := c.MakeRequest("PUT", "/dns/nameservers/"+url.PathEscape(groupID), bytes.NewReader(body))
		if err != nil {
			return DNSUpdatedMsg{Err: err}
		}
		defer resp.Body.Close()
		return DNSUpdatedMsg{}
	}
}

// UpdateNetwork updates a network via PUT
func UpdateNetwork(c *client.Client, networkID string, req models.NetworkUpdateRequest) tea.Cmd {
	return func() tea.Msg {
		body, err := json.Marshal(req)
		if err != nil {
			return NetworkUpdatedMsg{Err: fmt.Errorf("marshal request: %w", err)}
		}
		resp, err := c.MakeRequest("PUT", "/networks/"+url.PathEscape(networkID), bytes.NewReader(body))
		if err != nil {
			return NetworkUpdatedMsg{Err: err}
		}
		defer resp.Body.Close()
		return NetworkUpdatedMsg{}
	}
}

// UpdateUser updates a user via PUT
func UpdateUser(c *client.Client, userID string, req models.UserUpdateRequest) tea.Cmd {
	return func() tea.Msg {
		body, err := json.Marshal(req)
		if err != nil {
			return UserUpdatedMsg{Err: fmt.Errorf("marshal request: %w", err)}
		}
		resp, err := c.MakeRequest("PUT", "/users/"+url.PathEscape(userID), bytes.NewReader(body))
		if err != nil {
			return UserUpdatedMsg{Err: err}
		}
		defer resp.Body.Close()
		return UserUpdatedMsg{}
	}
}

// UpdatePostureCheck updates a posture check via PUT
func UpdatePostureCheck(c *client.Client, checkID string, req models.PostureCheckRequest) tea.Cmd {
	return func() tea.Msg {
		body, err := json.Marshal(req)
		if err != nil {
			return PostureCheckUpdatedMsg{Err: fmt.Errorf("marshal request: %w", err)}
		}
		resp, err := c.MakeRequest("PUT", "/posture-checks/"+url.PathEscape(checkID), bytes.NewReader(body))
		if err != nil {
			return PostureCheckUpdatedMsg{Err: err}
		}
		defer resp.Body.Close()
		return PostureCheckUpdatedMsg{}
	}
}

// DeletePostureCheck deletes a posture check
func DeletePostureCheck(c *client.Client, checkID string) tea.Cmd {
	return func() tea.Msg {
		resp, err := c.MakeRequest("DELETE", "/posture-checks/"+url.PathEscape(checkID), nil)
		if err != nil {
			return PostureCheckDeletedMsg{Err: err}
		}
		defer resp.Body.Close()
		return PostureCheckDeletedMsg{}
	}
}
```

- [ ] **Step 6: Build**

Run: `cd /Users/jack/Development/netbird-management-cli && go build ./cmd/netbird-manage/`

- [ ] **Step 7: Commit**

```bash
git add internal/tui/messages.go internal/tui/api.go internal/tui/keys.go
git commit -m "feat(tui): add edit message types and API update commands"
```

---

### Task 8: Add edit forms for simple resources (Peers, Groups, Users)

**Files:**
- Modify: `internal/tui/forms.go`
- Modify: `internal/tui/peers_page.go`
- Modify: `internal/tui/groups_page.go`
- Modify: `internal/tui/users_page.go`

- [ ] **Step 1: Add peer edit form constructor**

In `internal/tui/forms.go`, add:

```go
func newPeerEditForm(data *peerEditFormData) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Name").
				Value(&data.name),
			huh.NewConfirm().
				Title("SSH Enabled").
				Value(&data.sshEnabled),
			huh.NewConfirm().
				Title("Login Expiration Enabled").
				Value(&data.loginExpirationEnabled),
			huh.NewConfirm().
				Title("Inactivity Expiration Enabled").
				Value(&data.inactivityExpirationEnabled),
		),
	).WithTheme(formTheme())
}
```

Add the form data struct:

```go
type peerEditFormData struct {
	name                        string
	sshEnabled                  bool
	loginExpirationEnabled      bool
	inactivityExpirationEnabled bool
}
```

- [ ] **Step 2: Add group edit and user edit form constructors**

```go
// Group edit reuses the same form as create — just a name field
func newGroupEditForm(data *groupFormData) *huh.Form {
	return newGroupCreateForm(data)
}

type userEditFormData struct {
	role       string
	autoGroups string
}

func newUserEditForm(data *userEditFormData) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Role").
				Options(
					huh.NewOption("Admin", "admin"),
					huh.NewOption("User", "user"),
					huh.NewOption("Owner", "owner"),
				).
				Value(&data.role),
			huh.NewInput().
				Title("Auto Groups (comma-separated IDs)").
				Value(&data.autoGroups),
		),
	).WithTheme(formTheme())
}
```

- [ ] **Step 3: Add submit functions for peer, group, user edits**

In `internal/tui/forms.go`, add:

```go
func submitPeerEdit(c *client.Client, peerID string, data peerEditFormData) tea.Cmd {
	req := models.PeerUpdateRequest{
		Name:                        data.name,
		SSHEnabled:                  data.sshEnabled,
		LoginExpirationEnabled:      data.loginExpirationEnabled,
		InactivityExpirationEnabled: data.inactivityExpirationEnabled,
	}
	return UpdatePeer(c, peerID, req)
}

func submitGroupEdit(c *client.Client, groupID string, data groupFormData) tea.Cmd {
	return UpdateGroup(c, groupID, data.name)
}

func submitUserEdit(c *client.Client, userID string, data userEditFormData) tea.Cmd {
	req := models.UserUpdateRequest{
		Role:       data.role,
		AutoGroups: splitTrim(data.autoGroups),
	}
	return UpdateUser(c, userID, req)
}
```

- [ ] **Step 4: Add edit state and `e` key handling to PeersPage**

Add `peersViewEdit` and `peersViewEditConfirm` to the view state enum. Add `editForm *huh.Form`, `editData peerEditFormData`, `editPeerID string` fields to `PeersPage`.

In the detail view key handling, add:

```go
case "e":
	if len(p.filtered) > 0 {
		peer := p.filtered[p.cursor]
		p.editData = peerEditFormData{
			name:                        peer.Name,
			sshEnabled:                  peer.SSHEnabled,
			loginExpirationEnabled:      peer.LoginExpirationEnabled,
			inactivityExpirationEnabled: peer.InactivityExpirationEnabled,
		}
		p.editForm = newPeerEditForm(&p.editData)
		p.editPeerID = peer.ID
		p.state = peersViewEdit
	}
```

Add form delegation in Update (same pattern as create forms — delegate to `p.editForm.Update(msg)`, check for `StateCompleted` → transition to `peersViewEditConfirm`, check for `StateAborted` → back to detail).

Add confirm handling: `y` → `submitPeerEdit(c, p.editPeerID, p.editData)`, `n`/`esc` → back to detail.

Add `PeerUpdatedMsg` handler: refresh the page via `p.Init(c)`, return to list view.

Add edit form and confirm views in `View()`.

- [ ] **Step 5: Add edit state and `e` key handling to GroupsPage**

Same pattern. Add `groupsViewEdit`, `groupsViewEditConfirm` states. In detail view, `e` pre-populates `groupFormData{name: g.detail.Name}` and opens the form. On confirm, call `submitGroupEdit`. Handle `GroupUpdatedMsg` to refresh.

- [ ] **Step 6: Add edit state and `e` key handling to UsersPage**

Same pattern. Add edit states. In detail view, `e` pre-populates `userEditFormData{role: user.Role, autoGroups: strings.Join(user.AutoGroups, ", ")}`. On confirm, call `submitUserEdit`. Handle `UserUpdatedMsg` to refresh.

- [ ] **Step 7: Build and test**

Run: `cd /Users/jack/Development/netbird-management-cli && go build ./cmd/netbird-manage/`
Test: Navigate to a peer/group/user detail, press `e`, verify form opens pre-populated, submit edit.

- [ ] **Step 8: Commit**

```bash
git add internal/tui/forms.go internal/tui/peers_page.go internal/tui/groups_page.go internal/tui/users_page.go
git commit -m "feat(tui): add edit forms for peers, groups, and users"
```

---

### Task 9: Add edit forms for Policies, Routes, DNS

**Files:**
- Modify: `internal/tui/forms.go`
- Modify: `internal/tui/policies_page.go`
- Modify: `internal/tui/routes_page.go`
- Modify: `internal/tui/dns_page.go`

- [ ] **Step 1: Add policy edit form**

In `forms.go`, add edit form constructor that reuses `policyFormData` pre-populated from the existing policy:

```go
func newPolicyEditForm(data *policyFormData, availableGroups map[string]string) *huh.Form {
	return newPolicyCreateForm(data, availableGroups)
}

func submitPolicyEdit(c *client.Client, policyID string, data policyFormData) tea.Cmd {
	rules := []models.PolicyRuleForWrite{{
		Name:          data.name + "-rule",
		Enabled:       true,
		Action:        data.action,
		Bidirectional: true,
		Protocol:      data.protocol,
		Ports:         splitTrim(data.ports),
		Sources:       data.selectedSrc,
		Destinations:  data.selectedDst,
	}}
	req := models.PolicyUpdateRequest{
		Name:        data.name,
		Description: data.description,
		Enabled:     true,
		Rules:       rules,
	}
	return UpdatePolicy(c, policyID, req)
}
```

- [ ] **Step 2: Add route edit form**

```go
func newRouteEditForm(data *routeFormData, availableGroups map[string]string) *huh.Form {
	return newRouteCreateForm(data, availableGroups)
}

func submitRouteEdit(c *client.Client, route models.Route, data routeFormData) tea.Cmd {
	metric := 9999
	if m, err := strconv.Atoi(data.metric); err == nil {
		metric = m
	}
	req := models.RouteRequest{
		Description:         data.description,
		NetworkID:           data.networkID,
		Network:             data.network,
		Domains:             route.Domains,
		Peer:                route.Peer,
		PeerGroups:          data.selectedPeerGrps,
		Metric:              metric,
		Masquerade:          data.masquerade,
		Enabled:             data.enabled,
		Groups:              data.selectedDistGrps,
		AccessControlGroups: route.AccessControlGroups,
		KeepRoute:           route.KeepRoute,
	}
	return UpdateRoute(c, route.ID, req)
}
```

- [ ] **Step 3: Add DNS edit form**

```go
func newDNSEditForm(data *dnsFormData, availableGroups map[string]string) *huh.Form {
	return newDNSCreateForm(data, availableGroups)
}

func submitDNSEdit(c *client.Client, groupID string, data dnsFormData) tea.Cmd {
	req := models.DNSNameserverGroupRequest{
		Name:                 data.name,
		Description:          data.description,
		Nameservers:          buildNameservers(data.nameservers),
		Groups:               data.selectedGroups,
		Domains:              splitTrim(data.domains),
		SearchDomainsEnabled: data.searchDomain,
		Primary:              data.primary,
		Enabled:              true,
	}
	return UpdateDNSGroup(c, groupID, req)
}
```

- [ ] **Step 4: Wire edit into PoliciesPage**

Add edit view states. In detail view, `e` pre-populates `policyFormData` from the current policy (extract first rule's sources/destinations as group IDs). The form needs `availableGroups` — fetch groups if not cached, or use groups already loaded by the page.

Handle `PolicyUpdatedMsg` to refresh.

- [ ] **Step 5: Wire edit into RoutesPage**

Same pattern. Pre-populate `routeFormData` from the current route. Needs `availableGroups` for peer groups and distribution groups multiselects.

Handle `RouteUpdatedMsg` to refresh.

- [ ] **Step 6: Wire edit into DNSPage**

Same pattern. Pre-populate `dnsFormData` from the current DNS group. Needs `availableGroups`.

Handle `DNSUpdatedMsg` to refresh.

- [ ] **Step 7: Build and test**

Run: `cd /Users/jack/Development/netbird-management-cli && go build ./cmd/netbird-manage/`

- [ ] **Step 8: Commit**

```bash
git add internal/tui/forms.go internal/tui/policies_page.go internal/tui/routes_page.go internal/tui/dns_page.go
git commit -m "feat(tui): add edit forms for policies, routes, and DNS"
```

---

### Task 10: Add edit form for Networks

**Files:**
- Modify: `internal/tui/forms.go`
- Modify: `internal/tui/networks_page.go`

- [ ] **Step 1: Add network edit form and submit**

```go
func newNetworkEditForm(data *networkFormData) *huh.Form {
	return newNetworkCreateForm(data)
}

func submitNetworkEdit(c *client.Client, networkID string, data networkFormData) tea.Cmd {
	req := models.NetworkUpdateRequest{
		Name:        data.name,
		Description: data.description,
	}
	return UpdateNetwork(c, networkID, req)
}
```

- [ ] **Step 2: Wire edit into NetworksPage**

Add edit view states. In detail view, `e` pre-populates `networkFormData{name: net.Name, description: net.Description}`. Handle `NetworkUpdatedMsg` to refresh.

Note: Editing resources and routers within a network uses the existing sub-form pattern (the network detail already shows resources/routers). For now, network edit covers name and description. Resource/router editing follows the same sub-form pattern as creation but is deferred to a follow-up if the scope is too large.

- [ ] **Step 3: Build and test**

Run: `cd /Users/jack/Development/netbird-management-cli && go build ./cmd/netbird-manage/`

- [ ] **Step 4: Commit**

```bash
git add internal/tui/forms.go internal/tui/networks_page.go
git commit -m "feat(tui): add edit form for networks"
```

---

### Task 11: Add edit form, create, and delete for Posture Checks

**Files:**
- Modify: `internal/tui/forms.go`
- Modify: `internal/tui/posture_checks_page.go`

This is the most complex edit form due to type-specific fields.

- [ ] **Step 1: Add posture check form data structs**

In `forms.go`, add:

```go
type postureCheckFormData struct {
	name        string
	description string
	checkType   string // "nb_version", "os_version", "geo_location", "peer_network_range", "process"
	// NB Version
	nbMinVersion string
	// OS Version
	androidMinVersion      string
	darwinMinVersion       string
	iosMinVersion          string
	linuxMinKernelVersion  string
	windowsMinKernelVersion string
	// Geo Location
	geoLocations string // comma-separated "US,GB,DE"
	geoAction    string // "allow" or "deny"
	// Peer Network Range
	networkRanges      string // comma-separated CIDRs
	networkRangeAction string // "allow" or "deny"
	// Process
	processes string // comma-separated process paths
}
```

- [ ] **Step 2: Add posture check create form**

```go
func newPostureCheckCreateForm(data *postureCheckFormData) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Name").Value(&data.name),
			huh.NewInput().Title("Description").Value(&data.description),
			huh.NewSelect[string]().
				Title("Check Type").
				Options(
					huh.NewOption("NetBird Version", "nb_version"),
					huh.NewOption("OS Version", "os_version"),
					huh.NewOption("Geo Location", "geo_location"),
					huh.NewOption("Peer Network Range", "peer_network_range"),
					huh.NewOption("Process Check", "process"),
				).
				Value(&data.checkType),
		),
		// NB Version group
		huh.NewGroup(
			huh.NewInput().Title("Minimum NetBird Version (e.g. 0.25.0)").Value(&data.nbMinVersion),
		).WithHideFunc(func() bool { return data.checkType != "nb_version" }),
		// OS Version group
		huh.NewGroup(
			huh.NewInput().Title("Android Min Version").Value(&data.androidMinVersion),
			huh.NewInput().Title("macOS Min Version").Value(&data.darwinMinVersion),
			huh.NewInput().Title("iOS Min Version").Value(&data.iosMinVersion),
			huh.NewInput().Title("Linux Min Kernel Version").Value(&data.linuxMinKernelVersion),
			huh.NewInput().Title("Windows Min Kernel Version").Value(&data.windowsMinKernelVersion),
		).WithHideFunc(func() bool { return data.checkType != "os_version" }),
		// Geo Location group
		huh.NewGroup(
			huh.NewInput().Title("Country Codes (comma-separated, e.g. US,GB,DE)").Value(&data.geoLocations),
			huh.NewSelect[string]().
				Title("Action").
				Options(
					huh.NewOption("Allow", "allow"),
					huh.NewOption("Deny", "deny"),
				).
				Value(&data.geoAction),
		).WithHideFunc(func() bool { return data.checkType != "geo_location" }),
		// Peer Network Range group
		huh.NewGroup(
			huh.NewInput().Title("Network Ranges (comma-separated CIDRs)").Value(&data.networkRanges),
			huh.NewSelect[string]().
				Title("Action").
				Options(
					huh.NewOption("Allow", "allow"),
					huh.NewOption("Deny", "deny"),
				).
				Value(&data.networkRangeAction),
		).WithHideFunc(func() bool { return data.checkType != "peer_network_range" }),
		// Process group
		huh.NewGroup(
			huh.NewInput().Title("Process Paths (comma-separated)").Value(&data.processes),
		).WithHideFunc(func() bool { return data.checkType != "process" }),
	).WithTheme(formTheme())
}
```

- [ ] **Step 3: Add posture check edit form (reuses create form)**

```go
func newPostureCheckEditForm(data *postureCheckFormData) *huh.Form {
	return newPostureCheckCreateForm(data)
}
```

- [ ] **Step 4: Add submit functions for posture check create and edit**

```go
func buildPostureCheckRequest(data postureCheckFormData) models.PostureCheckRequest {
	req := models.PostureCheckRequest{
		Name:        data.name,
		Description: data.description,
	}
	switch data.checkType {
	case "nb_version":
		req.Checks.NBVersionCheck = &models.NBVersionCheck{
			MinVersion: data.nbMinVersion,
		}
	case "os_version":
		check := &models.OSVersionCheck{}
		if data.androidMinVersion != "" {
			check.Android = &models.MinVersionConfig{MinVersion: data.androidMinVersion}
		}
		if data.darwinMinVersion != "" {
			check.Darwin = &models.MinVersionConfig{MinVersion: data.darwinMinVersion}
		}
		if data.iosMinVersion != "" {
			check.IOS = &models.MinVersionConfig{MinVersion: data.iosMinVersion}
		}
		if data.linuxMinKernelVersion != "" {
			check.Linux = &models.MinKernelVersionConfig{MinKernelVersion: data.linuxMinKernelVersion}
		}
		if data.windowsMinKernelVersion != "" {
			check.Windows = &models.MinKernelVersionConfig{MinKernelVersion: data.windowsMinKernelVersion}
		}
		req.Checks.OSVersionCheck = check
	case "geo_location":
		locations := make([]models.Location, 0)
		for _, code := range splitTrim(data.geoLocations) {
			locations = append(locations, models.Location{CountryCode: code})
		}
		req.Checks.GeoLocationCheck = &models.GeoLocationCheck{
			Locations: locations,
			Action:    data.geoAction,
		}
	case "peer_network_range":
		req.Checks.PeerNetworkRangeCheck = &models.PeerNetworkRangeCheck{
			Ranges: splitTrim(data.networkRanges),
			Action: data.networkRangeAction,
		}
	case "process":
		processes := make([]models.Process, 0)
		for _, path := range splitTrim(data.processes) {
			processes = append(processes, models.Process{LinuxPath: path, MacPath: path, WindowsPath: path})
		}
		req.Checks.ProcessCheck = &models.ProcessCheck{Processes: processes}
	}
	return req
}

func submitPostureCheckCreate(c *client.Client, data postureCheckFormData) tea.Cmd {
	return func() tea.Msg {
		req := buildPostureCheckRequest(data)
		body, err := json.Marshal(req)
		if err != nil {
			return APIErrorMsg{Err: fmt.Errorf("marshal request: %w", err), Context: "create posture check"}
		}
		resp, err := c.MakeRequest("POST", "/posture-checks", bytes.NewReader(body))
		if err != nil {
			return APIErrorMsg{Err: err, Context: "create posture check"}
		}
		defer resp.Body.Close()
		return formCompleteMsg{message: "Posture check created"}
	}
}

func submitPostureCheckEdit(c *client.Client, checkID string, data postureCheckFormData) tea.Cmd {
	req := buildPostureCheckRequest(data)
	return UpdatePostureCheck(c, checkID, req)
}
```

- [ ] **Step 5: Add pre-population helper for posture check edit**

```go
func postureCheckToFormData(check models.PostureCheck) postureCheckFormData {
	data := postureCheckFormData{
		name:        check.Name,
		description: check.Description,
	}
	switch {
	case check.Checks.NBVersionCheck != nil:
		data.checkType = "nb_version"
		data.nbMinVersion = check.Checks.NBVersionCheck.MinVersion
	case check.Checks.OSVersionCheck != nil:
		data.checkType = "os_version"
		os := check.Checks.OSVersionCheck
		if os.Android != nil {
			data.androidMinVersion = os.Android.MinVersion
		}
		if os.Darwin != nil {
			data.darwinMinVersion = os.Darwin.MinVersion
		}
		if os.IOS != nil {
			data.iosMinVersion = os.IOS.MinVersion
		}
		if os.Linux != nil {
			data.linuxMinKernelVersion = os.Linux.MinKernelVersion
		}
		if os.Windows != nil {
			data.windowsMinKernelVersion = os.Windows.MinKernelVersion
		}
	case check.Checks.GeoLocationCheck != nil:
		data.checkType = "geo_location"
		codes := make([]string, len(check.Checks.GeoLocationCheck.Locations))
		for i, loc := range check.Checks.GeoLocationCheck.Locations {
			codes[i] = loc.CountryCode
		}
		data.geoLocations = strings.Join(codes, ", ")
		data.geoAction = check.Checks.GeoLocationCheck.Action
	case check.Checks.PeerNetworkRangeCheck != nil:
		data.checkType = "peer_network_range"
		data.networkRanges = strings.Join(check.Checks.PeerNetworkRangeCheck.Ranges, ", ")
		data.networkRangeAction = check.Checks.PeerNetworkRangeCheck.Action
	case check.Checks.ProcessCheck != nil:
		data.checkType = "process"
		paths := make([]string, len(check.Checks.ProcessCheck.Processes))
		for i, p := range check.Checks.ProcessCheck.Processes {
			if p.LinuxPath != "" {
				paths[i] = p.LinuxPath
			} else if p.MacPath != "" {
				paths[i] = p.MacPath
			} else {
				paths[i] = p.WindowsPath
			}
		}
		data.processes = strings.Join(paths, ", ")
	}
	return data
}
```

- [ ] **Step 6: Wire create, edit, delete into PostureChecksPage**

Add view states: `postureChecksViewForm`, `postureChecksViewConfirm`, `postureChecksViewEdit`, `postureChecksViewEditConfirm`.

Add fields: `form *huh.Form`, `formData postureCheckFormData`, `editCheckID string`.

Add key handling in list view:
- `c` → create (new form with empty data)
- `d` → delete (call `DeletePostureCheck`)

Add key handling in detail view:
- `e` → edit (pre-populate form data via `postureCheckToFormData`)

Handle `PostureCheckUpdatedMsg`, `PostureCheckDeletedMsg`, `formCompleteMsg` → refresh via `Init(c)`.

Add form delegation, confirm screen rendering (same patterns as groups page).

- [ ] **Step 7: Build and test**

Run: `cd /Users/jack/Development/netbird-management-cli && go build ./cmd/netbird-manage/`

- [ ] **Step 8: Commit**

```bash
git add internal/tui/forms.go internal/tui/posture_checks_page.go
git commit -m "feat(tui): add create, edit, and delete for posture checks"
```

---

## Phase 3: Connection Detail Per Peer

### Task 12: Extend daemon client for per-peer connection info

**Files:**
- Modify: `internal/tui/daemon.go`
- Modify: `internal/tui/helpers.go`
- Modify: `internal/tui/messages.go`

- [ ] **Step 1: Add PeerConnectionInfo struct and message type**

In `messages.go`, add:

```go
type PeerConnectionInfo struct {
	ConnType        string // "P2P" or "Relayed"
	RemoteEndpoint  string
	LocalICEType    string
	RemoteICEType   string
	Latency         string
	BytesSent       int64
	BytesReceived   int64
	LastHandshake   string
	RelayAddress    string
}

type PeerConnectionMsg struct {
	Connections map[string]PeerConnectionInfo // keyed by NetBird IP
	Err         error
}
```

- [ ] **Step 2: Add daemon method to fetch peer connections**

In `daemon.go`, add:

```go
func (d *DaemonClient) PeerConnections() (map[string]PeerConnectionInfo, error) {
	if d == nil {
		return nil, fmt.Errorf("daemon not connected")
	}
	ctx, cancel := context.WithTimeout(context.Background(), daemonTimeout)
	defer cancel()

	status, err := d.daemon.Status(ctx, &proto.StatusRequest{GetFullPeerStatus: true})
	if err != nil {
		return nil, fmt.Errorf("daemon status: %w", err)
	}

	conns := make(map[string]PeerConnectionInfo)
	if status.GetFullStatus() != nil {
		for _, peer := range status.GetFullStatus().GetPeers() {
			info := PeerConnectionInfo{
				ConnType:       peer.GetConnStatusType().String(),
				RemoteEndpoint: peer.GetRemoteAddress(),
				LocalICEType:   peer.GetLocalIceCandidateType(),
				RemoteICEType:  peer.GetRemoteIceCandidateType(),
				Latency:        peer.GetLatency().String(),
				BytesSent:      peer.GetBytesTx(),
				BytesReceived:  peer.GetBytesRx(),
				LastHandshake:  peer.GetLastWireguardHandshake().String(),
				RelayAddress:   peer.GetRelayAddress(),
			}
			if peer.GetConnStatusType().String() == "Connected" {
				if peer.GetRelayed() {
					info.ConnType = "Relayed"
				} else {
					info.ConnType = "P2P"
				}
			} else {
				info.ConnType = "Disconnected"
			}
			conns[peer.GetIp()] = info
		}
	}
	return conns, nil
}
```

- [ ] **Step 3: Add byte formatting helper**

In `helpers.go`, add:

```go
func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
```

- [ ] **Step 4: Add FetchPeerConnections command**

In `api.go` (or a new section at the bottom), add:

```go
// FetchPeerConnections fetches connection info from the daemon
func FetchPeerConnections(daemon *DaemonClient) tea.Cmd {
	return func() tea.Msg {
		if daemon == nil {
			return PeerConnectionMsg{Err: fmt.Errorf("daemon not connected")}
		}
		conns, err := daemon.PeerConnections()
		return PeerConnectionMsg{Connections: conns, Err: err}
	}
}
```

- [ ] **Step 5: Build**

Run: `cd /Users/jack/Development/netbird-management-cli && go build ./cmd/netbird-manage/`

- [ ] **Step 6: Commit**

```bash
git add internal/tui/daemon.go internal/tui/helpers.go internal/tui/messages.go internal/tui/api.go
git commit -m "feat(tui): add daemon peer connection info infrastructure"
```

---

### Task 13: Add connection detail section to peer detail view

**Files:**
- Modify: `internal/tui/peers_page.go`
- Modify: `internal/tui/app.go`

- [ ] **Step 1: Add connection info state to PeersPage**

Add fields to `PeersPage`:

```go
connections map[string]PeerConnectionInfo // keyed by IP
```

- [ ] **Step 2: Pass daemon reference to pages**

The `Page.Update` currently receives `*client.Client`. To access the daemon, either:
- Add a `daemon *DaemonClient` field to `PeersPage` set during construction, or
- Pass the daemon through the App when creating pages

In `app.go`, when creating `PeersPage`, pass the daemon:

```go
// In NewApp or wherever pages are initialized:
peersPage := NewPeersPage()
peersPage.daemon = daemon
```

Add `daemon *DaemonClient` field to `PeersPage`.

- [ ] **Step 3: Fetch connections when entering detail view**

When the user presses `enter` to view peer detail, also fetch connection info:

```go
case "enter":
	if len(p.filtered) > 0 {
		p.state = peersViewDetail
		if p.daemon != nil {
			return p, FetchPeerConnections(p.daemon)
		}
	}
```

Handle `PeerConnectionMsg`:

```go
case PeerConnectionMsg:
	if msg.Err == nil {
		p.connections = msg.Connections
	}
	return p, nil
```

- [ ] **Step 4: Render connection detail in peer detail view**

In the detail view rendering, after existing fields, add a "Connection Detail" section:

```go
// After existing detail fields...
peer := p.filtered[p.cursor]
if conn, ok := p.connections[peer.IP]; ok {
	b.WriteString("\n" + sectionHeaderStyle.Render("  Connection Detail") + "\n")
	b.WriteString(padLabel("  Type:", labelWidth) + detailValueStyle.Render(conn.ConnType) + "\n")
	b.WriteString(padLabel("  Remote Endpoint:", labelWidth) + detailValueStyle.Render(conn.RemoteEndpoint) + "\n")
	b.WriteString(padLabel("  Local ICE:", labelWidth) + detailValueStyle.Render(conn.LocalICEType) + "\n")
	b.WriteString(padLabel("  Remote ICE:", labelWidth) + detailValueStyle.Render(conn.RemoteICEType) + "\n")
	b.WriteString(padLabel("  Latency:", labelWidth) + detailValueStyle.Render(conn.Latency) + "\n")
	b.WriteString(padLabel("  Sent:", labelWidth) + detailValueStyle.Render(formatBytes(conn.BytesSent)) + "\n")
	b.WriteString(padLabel("  Received:", labelWidth) + detailValueStyle.Render(formatBytes(conn.BytesReceived)) + "\n")
	b.WriteString(padLabel("  Last Handshake:", labelWidth) + detailValueStyle.Render(conn.LastHandshake) + "\n")
	if conn.RelayAddress != "" {
		b.WriteString(padLabel("  Relay:", labelWidth) + detailValueStyle.Render(conn.RelayAddress) + "\n")
	}
} else if p.daemon != nil {
	b.WriteString("\n" + dimHintStyle.Render("  Peer not connected locally") + "\n")
} else {
	b.WriteString("\n" + dimHintStyle.Render("  Connect to NetBird daemon for connection details") + "\n")
}
```

- [ ] **Step 5: Build and test**

Run: `cd /Users/jack/Development/netbird-management-cli && go build ./cmd/netbird-manage/`
Test: View a peer's detail, verify connection section appears.

- [ ] **Step 6: Commit**

```bash
git add internal/tui/peers_page.go internal/tui/app.go
git commit -m "feat(tui): show peer connection detail from daemon"
```

---

## Phase 4: Live Monitoring

### Task 14: Add auto-refresh tick infrastructure

**Files:**
- Create: `internal/tui/live.go`
- Modify: `internal/tui/messages.go`

- [ ] **Step 1: Add tick message types**

In `messages.go`, add:

```go
// Auto-refresh tick messages (one per page that supports live monitoring)
type PeersTickMsg struct{}
type EventsTickMsg struct{}
type ConnectionTickMsg struct{}
```

- [ ] **Step 2: Create live.go with tick helpers**

```go
// internal/tui/live.go
package tui

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

const (
	peersRefreshInterval      = 5 * time.Second
	eventsRefreshInterval     = 10 * time.Second
	connectionRefreshInterval = 5 * time.Second
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
```

- [ ] **Step 3: Build**

Run: `cd /Users/jack/Development/netbird-management-cli && go build ./cmd/netbird-manage/`

- [ ] **Step 4: Commit**

```bash
git add internal/tui/live.go internal/tui/messages.go
git commit -m "feat(tui): add auto-refresh tick infrastructure"
```

---

### Task 15: Add live monitoring to Peers list

**Files:**
- Modify: `internal/tui/peers_page.go`

- [ ] **Step 1: Add auto-refresh state to PeersPage**

Add fields:

```go
autoRefresh bool      // whether auto-refresh is active
lastRefresh time.Time // when data was last fetched
```

Initialize `autoRefresh: true` in `NewPeersPage()`.

- [ ] **Step 2: Start tick on Init**

In `Init`, return a batch command that includes the tick:

```go
func (p *PeersPage) Init(c *client.Client) tea.Cmd {
	p.loading = true
	cmds := []tea.Cmd{FetchPeers(c)}
	if p.autoRefresh {
		cmds = append(cmds, peersTickCmd())
	}
	return tea.Batch(cmds...)
}
```

- [ ] **Step 3: Handle tick message**

In `Update`, handle `PeersTickMsg`:

```go
case PeersTickMsg:
	if p.focused && p.autoRefresh && p.state == peersViewList {
		p.lastRefresh = time.Now()
		return p, tea.Batch(FetchPeers(c), peersTickCmd())
	}
	if p.autoRefresh {
		return p, peersTickCmd() // keep ticking, refresh when focused
	}
	return p, nil
```

- [ ] **Step 4: Preserve cursor on refresh**

In the `PeersLoadedMsg` handler, preserve cursor position and search:

```go
case PeersLoadedMsg:
	p.loading = false
	if msg.Err != nil {
		p.err = msg.Err
		return p, nil
	}
	p.peers = msg.Peers
	p.lastRefresh = time.Now()
	p.applyFilter() // this already clamps cursor
	return p, nil
```

- [ ] **Step 5: Add pause/resume with `p` key**

In the list view key handling:

```go
case "p":
	p.autoRefresh = !p.autoRefresh
	if p.autoRefresh {
		return p, peersTickCmd()
	}
```

- [ ] **Step 6: Show refresh indicator in view**

In the list view, add a "last updated" line:

```go
if p.autoRefresh && !p.lastRefresh.IsZero() {
	ago := time.Since(p.lastRefresh).Truncate(time.Second)
	b.WriteString(dimHintStyle.Render(fmt.Sprintf("  auto-refresh: %s ago  (p to pause)", ago)) + "\n")
} else if !p.autoRefresh {
	b.WriteString(dimHintStyle.Render("  auto-refresh paused  (p to resume)") + "\n")
}
```

- [ ] **Step 7: Build and test**

Run: `cd /Users/jack/Development/netbird-management-cli && go build ./cmd/netbird-manage/`
Test: Watch peers list auto-update. Press `p` to pause/resume. Verify cursor stays put.

- [ ] **Step 8: Commit**

```bash
git add internal/tui/peers_page.go
git commit -m "feat(tui): add live auto-refresh to peers list"
```

---

### Task 16: Add live monitoring to Events and connection detail

**Files:**
- Modify: `internal/tui/events_page.go`
- Modify: `internal/tui/peers_page.go`

- [ ] **Step 1: Add auto-refresh to EventsPage**

Same pattern as peers. Add `autoRefresh bool`, `lastRefresh time.Time` fields. Start tick in `Init`. Handle `EventsTickMsg` — only refresh when focused and in list view. Add `p` key to pause/resume. Show "last updated" indicator.

```go
case EventsTickMsg:
	if e.focused && e.autoRefresh {
		e.lastRefresh = time.Now()
		return e, tea.Batch(FetchEvents(c), eventsTickCmd())
	}
	if e.autoRefresh {
		return e, eventsTickCmd()
	}
	return e, nil
```

- [ ] **Step 2: Add connection auto-refresh in peer detail view**

In `PeersPage`, when entering detail view and daemon is available, start a connection tick. Handle `ConnectionTickMsg`:

```go
case ConnectionTickMsg:
	if p.focused && p.state == peersViewDetail && p.daemon != nil && p.autoRefresh {
		return p, tea.Batch(FetchPeerConnections(p.daemon), connectionTickCmd())
	}
	return p, nil
```

Start the connection tick when entering detail view:

```go
case "enter":
	if len(p.filtered) > 0 {
		p.state = peersViewDetail
		cmds := []tea.Cmd{}
		if p.daemon != nil {
			cmds = append(cmds, FetchPeerConnections(p.daemon), connectionTickCmd())
		}
		return p, tea.Batch(cmds...)
	}
```

Stop connection ticks when leaving detail view (they naturally stop since the `ConnectionTickMsg` handler checks `p.state == peersViewDetail`).

- [ ] **Step 3: Build and test**

Run: `cd /Users/jack/Development/netbird-management-cli && go build ./cmd/netbird-manage/`

- [ ] **Step 4: Commit**

```bash
git add internal/tui/events_page.go internal/tui/peers_page.go
git commit -m "feat(tui): add live monitoring to events and peer connection detail"
```

---

## Phase 5: Bulk Group Assignment

### Task 17: Create bulk select state machine

**Files:**
- Create: `internal/tui/bulk.go`

- [ ] **Step 1: Write failing tests for bulk selection logic**

```go
// internal/tui/bulk_test.go
package tui

import "testing"

func TestBulkSelection(t *testing.T) {
	t.Run("toggle selection", func(t *testing.T) {
		bs := newBulkState(5)
		bs.toggle(2)
		if !bs.isSelected(2) {
			t.Error("expected index 2 to be selected")
		}
		bs.toggle(2)
		if bs.isSelected(2) {
			t.Error("expected index 2 to be deselected")
		}
	})

	t.Run("select all", func(t *testing.T) {
		bs := newBulkState(3)
		bs.selectAll()
		if bs.count() != 3 {
			t.Errorf("expected 3 selected, got %d", bs.count())
		}
	})

	t.Run("deselect all when all selected", func(t *testing.T) {
		bs := newBulkState(3)
		bs.selectAll()
		bs.selectAll() // toggles to deselect
		if bs.count() != 0 {
			t.Errorf("expected 0 selected, got %d", bs.count())
		}
	})

	t.Run("selected indices", func(t *testing.T) {
		bs := newBulkState(5)
		bs.toggle(1)
		bs.toggle(3)
		indices := bs.selectedIndices()
		if len(indices) != 2 || indices[0] != 1 || indices[1] != 3 {
			t.Errorf("expected [1 3], got %v", indices)
		}
	})
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/jack/Development/netbird-management-cli && go test -race ./internal/tui/ -run TestBulkSelection -v`
Expected: FAIL

- [ ] **Step 3: Implement bulk state machine**

```go
// internal/tui/bulk.go
package tui

import "sort"

// bulkState tracks multi-selection state for list items.
type bulkState struct {
	selected map[int]bool
	total    int
}

func newBulkState(total int) *bulkState {
	return &bulkState{
		selected: make(map[int]bool),
		total:    total,
	}
}

func (b *bulkState) toggle(index int) {
	if b.selected[index] {
		delete(b.selected, index)
	} else {
		b.selected[index] = true
	}
}

func (b *bulkState) isSelected(index int) bool {
	return b.selected[index]
}

func (b *bulkState) selectAll() {
	if b.count() == b.total {
		b.selected = make(map[int]bool)
	} else {
		for i := 0; i < b.total; i++ {
			b.selected[i] = true
		}
	}
}

func (b *bulkState) count() int {
	return len(b.selected)
}

func (b *bulkState) selectedIndices() []int {
	indices := make([]int, 0, len(b.selected))
	for i := range b.selected {
		indices = append(indices, i)
	}
	sort.Ints(indices)
	return indices
}

func (b *bulkState) reset() {
	b.selected = make(map[int]bool)
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/jack/Development/netbird-management-cli && go test -race ./internal/tui/ -run TestBulkSelection -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/tui/bulk.go internal/tui/bulk_test.go
git commit -m "feat(tui): add bulk selection state machine"
```

---

### Task 18: Wire bulk group assignment into Peers page

**Files:**
- Modify: `internal/tui/peers_page.go`
- Modify: `internal/tui/messages.go`
- Modify: `internal/tui/api.go`

- [ ] **Step 1: Add bulk state and view modes to PeersPage**

Add to the view state enum:

```go
peersViewBulkSelect    // multi-select mode
peersViewBulkAction    // choosing add/remove
peersViewBulkGroup     // picking target group
peersViewBulkConfirm   // confirming the operation
```

Add fields to `PeersPage`:

```go
bulk          *bulkState
bulkAction    string // "add" or "remove"
bulkGroups    []models.PolicyGroup // available groups for picker
bulkGroupIdx  int    // cursor in group picker
bulkTargetGrp models.PolicyGroup // selected target group
```

- [ ] **Step 2: Add bulk-related messages**

In `messages.go`:

```go
type BulkGroupsLoadedMsg struct {
	Groups []models.PolicyGroup
	Err    error
}

type BulkGroupAssignMsg struct {
	Err error
}
```

- [ ] **Step 3: Add bulk group assignment API command**

In `api.go`:

```go
// BulkAssignGroup adds or removes peers from a group
func BulkAssignGroup(c *client.Client, group models.PolicyGroup, peerIDs []string, action string) tea.Cmd {
	return func() tea.Msg {
		// Fetch full group detail to get current peers
		resp, err := c.MakeRequest("GET", "/groups/"+url.PathEscape(group.ID), nil)
		if err != nil {
			return BulkGroupAssignMsg{Err: fmt.Errorf("fetch group: %w", err)}
		}
		defer resp.Body.Close()

		var detail struct {
			ID    string   `json:"id"`
			Name  string   `json:"name"`
			Peers []struct {
				ID string `json:"id"`
			} `json:"peers"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&detail); err != nil {
			return BulkGroupAssignMsg{Err: fmt.Errorf("decode group: %w", err)}
		}

		// Build updated peer list
		existingIDs := make(map[string]bool)
		for _, p := range detail.Peers {
			existingIDs[p.ID] = true
		}

		if action == "add" {
			for _, id := range peerIDs {
				existingIDs[id] = true
			}
		} else {
			for _, id := range peerIDs {
				delete(existingIDs, id)
			}
		}

		updatedPeerIDs := make([]string, 0, len(existingIDs))
		for id := range existingIDs {
			updatedPeerIDs = append(updatedPeerIDs, id)
		}

		// PUT updated group
		updateReq := struct {
			Name  string   `json:"name"`
			Peers []struct {
				ID string `json:"id"`
			} `json:"peers"`
		}{
			Name: detail.Name,
		}
		for _, id := range updatedPeerIDs {
			updateReq.Peers = append(updateReq.Peers, struct {
				ID string `json:"id"`
			}{ID: id})
		}

		body, err := json.Marshal(updateReq)
		if err != nil {
			return BulkGroupAssignMsg{Err: fmt.Errorf("marshal update: %w", err)}
		}
		resp2, err := c.MakeRequest("PUT", "/groups/"+url.PathEscape(group.ID), bytes.NewReader(body))
		if err != nil {
			return BulkGroupAssignMsg{Err: err}
		}
		defer resp2.Body.Close()

		return BulkGroupAssignMsg{}
	}
}
```

- [ ] **Step 4: Handle `g` key to enter bulk select mode**

In peers list key handling:

```go
case "g":
	if len(p.filtered) > 0 {
		p.bulk = newBulkState(len(p.filtered))
		p.state = peersViewBulkSelect
	}
```

- [ ] **Step 5: Handle bulk select mode keys**

Add a new section in `handleKey` for `peersViewBulkSelect`:

```go
if p.state == peersViewBulkSelect {
	switch key {
	case "space":
		p.bulk.toggle(p.cursor)
	case "a":
		p.bulk.selectAll()
	case "enter":
		if p.bulk.count() > 0 {
			p.state = peersViewBulkAction
		}
	case "esc":
		p.bulk = nil
		p.state = peersViewList
	case "up", "k":
		if p.cursor > 0 {
			p.cursor--
		}
	case "down", "j":
		if p.cursor < len(p.filtered)-1 {
			p.cursor++
		}
	}
	return p, nil
}
```

- [ ] **Step 6: Handle action picker and group picker**

For `peersViewBulkAction`:

```go
if p.state == peersViewBulkAction {
	switch key {
	case "1":
		p.bulkAction = "add"
		p.state = peersViewBulkGroup
		return p, FetchGroups(c) // reuse existing fetch, handle BulkGroupsLoadedMsg
	case "2":
		p.bulkAction = "remove"
		p.state = peersViewBulkGroup
		return p, FetchGroups(c)
	case "esc":
		p.state = peersViewBulkSelect
	}
	return p, nil
}
```

Handle `GroupsLoadedMsg` when in bulk mode to populate `bulkGroups`:

```go
case GroupsLoadedMsg:
	if p.state == peersViewBulkGroup {
		if msg.Err != nil {
			p.err = msg.Err
			p.state = peersViewBulkSelect
			return p, nil
		}
		p.bulkGroups = msg.Groups
		p.bulkGroupIdx = 0
	}
	return p, nil
```

For `peersViewBulkGroup`:

```go
if p.state == peersViewBulkGroup {
	switch key {
	case "up", "k":
		if p.bulkGroupIdx > 0 {
			p.bulkGroupIdx--
		}
	case "down", "j":
		if p.bulkGroupIdx < len(p.bulkGroups)-1 {
			p.bulkGroupIdx++
		}
	case "enter":
		p.bulkTargetGrp = p.bulkGroups[p.bulkGroupIdx]
		p.state = peersViewBulkConfirm
	case "esc":
		p.state = peersViewBulkAction
	}
	return p, nil
}
```

- [ ] **Step 7: Handle bulk confirm and execution**

For `peersViewBulkConfirm`:

```go
if p.state == peersViewBulkConfirm {
	switch key {
	case "y":
		peerIDs := make([]string, 0)
		for _, idx := range p.bulk.selectedIndices() {
			if idx < len(p.filtered) {
				peerIDs = append(peerIDs, p.filtered[idx].ID)
			}
		}
		p.state = peersViewList
		p.bulk = nil
		return p, BulkAssignGroup(c, p.bulkTargetGrp, peerIDs, p.bulkAction)
	case "n", "esc":
		p.state = peersViewBulkSelect
	}
	return p, nil
}
```

Handle `BulkGroupAssignMsg`:

```go
case BulkGroupAssignMsg:
	if msg.Err != nil {
		p.err = msg.Err
		return p, nil
	}
	return p, p.Init(c) // refresh
```

- [ ] **Step 8: Add bulk mode views**

Add rendering for each bulk state:
- **Bulk select**: Same table as list but with `[x]`/`[ ]` markers on selected rows. Status bar shows: "space: toggle  a: all  enter: assign  esc: cancel"
- **Bulk action**: Simple "1) Add to group  2) Remove from group" prompt
- **Bulk group**: Scrollable list of groups with cursor
- **Bulk confirm**: "Add 5 peers to group 'Production'? (y/n)" using `RenderConfirm`

- [ ] **Step 9: Build and test**

Run: `cd /Users/jack/Development/netbird-management-cli && go build ./cmd/netbird-manage/`
Test: Go to peers, press `g`, select peers with space, press enter, choose action, pick group, confirm.

- [ ] **Step 10: Commit**

```bash
git add internal/tui/peers_page.go internal/tui/messages.go internal/tui/api.go
git commit -m "feat(tui): add bulk group assignment for peers"
```

---

## Final: Run all tests

### Task 19: Verify everything builds and tests pass

- [ ] **Step 1: Run all tests**

Run: `cd /Users/jack/Development/netbird-management-cli && go test -race ./internal/tui/ -v`
Expected: All tests pass.

- [ ] **Step 2: Run full build**

Run: `cd /Users/jack/Development/netbird-management-cli && go build ./cmd/netbird-manage/`
Expected: Builds cleanly.

- [ ] **Step 3: Run vet**

Run: `cd /Users/jack/Development/netbird-management-cli && go vet ./...`
Expected: No issues.

- [ ] **Step 4: Final commit if any cleanup needed**

```bash
git add -A
git commit -m "chore(tui): final cleanup for new features"
```
