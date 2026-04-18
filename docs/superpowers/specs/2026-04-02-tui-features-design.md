# TUI New Features Design

**Date:** 2026-04-02
**Branch:** feat/tui
**Implementation order:** Search → Edit Forms → Connection Detail → Live Monitoring → Bulk Group Assignment

---

## 1. Search on All List Pages

### Scope

Add `/` search filtering to all list pages except DNS, Dashboard, Accounts, and Export/Import.

**Pages:** Groups, Networks, Policies, Routes, Setup Keys, Users, Service Users, Posture Checks, Events, Ingress (Peers already has search).

### Interaction

- `/` activates search input (same pattern as Peers today)
- Type to filter — list updates in real-time
- `esc` clears search and returns to full list
- Case-insensitive substring matching

### Searchable Fields

| Page | Fields |
|------|--------|
| Groups | name |
| Networks | name, description |
| Policies | name |
| Routes | network identifier, description |
| Setup Keys | name, state (valid/expired/revoked) |
| Users | name, email, role |
| Service Users | name |
| Posture Checks | name |
| Events | activity, initiator name, target |
| Ingress | peer name |

### Implementation

Extract the search pattern from `peers_page.go` into a reusable filter helper. Each page provides a `matchesSearch(query string) bool` function on its items. The search input renders inline above the table or in the status bar area, consistent with the existing Peers implementation.

---

## 2. Edit Forms

### Trigger

`e` key from the detail view opens a pre-populated huh/v2 form.

### Flow

Detail view → `e` → pre-populated form → confirmation screen (y/n) → API PUT → refresh → return to updated detail view.

### Editable Resources

| Page | Editable Fields |
|------|----------------|
| Peers | name, SSH enabled, login expiration enabled |
| Groups | name |
| Policies | name, description, enabled, rules (sources, destinations, protocol, ports, action) |
| Routes | description, network, peer groups, distribution groups, metric, masquerade, enabled |
| DNS | name, nameservers, groups, match domains, search domains, primary, enabled |
| Networks | name, description (plus sub-edit for resources and routers) |
| Users | role, auto-groups |
| Posture Checks | name, description, all type-specific check parameters |

### Implementation

Add `editForm` variants in `forms.go` that accept existing resource data to pre-populate fields. Reuse the same huh/v2 form constructors where possible, adding a parameter for initial values. Same confirmation screen pattern as create forms.

**Not editable (by design):** Setup Keys (immutable after creation), Accounts (read-only settings display), Events (audit log), Export/Import (CLI-only).

### Posture Checks Edit Complexity

Posture checks have type-specific parameters:
- **NBVersionCheck:** min version
- **OSVersionCheck:** OS, min version
- **GeoLocationCheck:** allowed locations list
- **PeerNetworkRangeCheck:** allowed ranges, action
- **ProcessCheck:** processes list with path/signature checks

The edit form detects the check type and renders the appropriate fields. Type cannot be changed after creation — only the parameters within the existing type.

### API Pattern

All edits use PUT with the full resource object:
1. Current resource data is already available in the detail view
2. Form pre-populates from that data
3. User modifies fields
4. Confirmation screen shows all fields (same format as create confirmation)
5. PUT sends the complete updated object

---

## 3. Connection Detail Per Peer

### Location

New section in the peer detail view, below the existing fields. Only visible when daemon is connected.

### Data Source

NetBird daemon gRPC — peer status already fetched for dashboard, extended to include per-peer connection info.

### Fields Displayed

| Field | Source |
|-------|--------|
| Connection type | P2P / Relayed |
| Remote endpoint | IP:port |
| Local ICE candidate type | host / srflx / relay |
| Remote ICE candidate type | host / srflx / relay |
| Latency (RTT) | Last measured round-trip time |
| Bytes sent | Cumulative transfer out |
| Bytes received | Cumulative transfer in |
| Last handshake | Timestamp of last WireGuard handshake |
| Relay address | Relay server address (if relayed) |

### Fallback

If daemon is not connected, display: "Connect to NetBird daemon for connection details" in a muted style.

### Matching Peers

Match management API peers to daemon peers by NetBird IP or public key. If no daemon match is found for a peer, show "Peer not connected locally" instead of connection fields.

---

## 4. Live Monitoring

### Auto-Refresh Behaviors

| Context | Interval | What Updates |
|---------|----------|--------------|
| Peers list | 5s | Online/offline status badges, connected count |
| Events page | 10s | New events prepended to top of list |
| Peer detail (connection section) | 5s | Latency, transfer bytes, connection type |

### UX Details

- **Status bar indicator:** "Last updated: Xs ago" shown when auto-refresh is active
- **Pause/resume:** `p` key toggles auto-refresh on any live page
- **Manual refresh:** `r` still works alongside auto-refresh
- **Focus-only:** Auto-refresh only fires when the page is currently focused (not background tabs)
- **Cursor preservation:** Refresh updates data in-place without resetting cursor position or scroll offset
- **Search interaction:** If search is active during auto-refresh, the filter re-applies to updated data

### Implementation

Add a `tickMsg` pattern using `tea.Tick` commands. Each page with live monitoring manages its own tick interval. The tick command is only sent when the page is focused and auto-refresh is not paused.

---

## 5. Bulk Group Assignment

### Trigger

`g` key from the Peers list enters bulk select mode.

### Flow

1. `g` on peers list → bulk select mode activates
   - Status bar updates to show bulk mode keybindings
   - Visual indicator that bulk mode is active
2. `space` toggles selection on current row
3. `a` selects/deselects all visible peers (respects active search filter)
4. Selected peers shown with a `[x]` marker or highlight
5. `enter` with selections → action picker:
   - **Add to group**
   - **Remove from group**
6. Group picker — searchable list of all groups
7. Confirmation screen: "Add 5 peers to group 'Production'?" (y/n)
8. API execution:
   - For "Add to group": GET group → append peer IDs → PUT group
   - For "Remove from group": GET group → remove peer IDs → PUT group
9. Refresh peers list, exit bulk mode
10. `esc` exits bulk mode without action

### API Pattern

Group updates require full PUT (fetch → modify → PUT entire object). For bulk operations, this means:
1. Fetch the target group
2. Add/remove the selected peer IDs from the group's peers list
3. PUT the complete updated group

Single API call per group (not per peer).

---

## Architecture Notes

### File Organization

New code fits within the existing `internal/tui/` package structure:

- **Search:** Add `searchable` helper in `helpers.go` or a new `search.go` (~50 lines). Each page adds `matchesSearch` and search state fields.
- **Edit forms:** Extend `forms.go` with edit variants. Each page adds `e` key handling in its `Update` method.
- **Connection detail:** Extend `peers_page.go` detail view section. May need new daemon helper in `daemon.go`.
- **Live monitoring:** Add tick command pattern. Each page manages its own interval. Shared `tickMsg` type in `messages.go`.
- **Bulk group assignment:** New `bulk.go` (~150 lines) for the selection state machine, integrated into `peers_page.go`.

### New Message Types (messages.go)

- `EditFormMsg` — triggers edit form open with pre-populated data
- `EditSuccessMsg` / `EditErrorMsg` — edit form submission results
- `BulkSelectMsg` — enters/exits bulk select mode
- `BulkActionMsg` — bulk action completed
- `TickMsg` — auto-refresh tick per page

### Key Bindings Summary

| Key | Context | Action |
|-----|---------|--------|
| `/` | List view | Activate search |
| `e` | Detail view | Open edit form |
| `g` | Peers list | Enter bulk select mode |
| `space` | Bulk select mode | Toggle peer selection |
| `a` | Bulk select mode | Select/deselect all |
| `p` | Live pages | Pause/resume auto-refresh |
