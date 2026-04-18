# Settings Page Design

## Overview

Replace the read-only Account page with an interactive Settings page that mirrors the real NetBird dashboard's layout. The page uses a vertical sidebar for section navigation and inline editing for modifying settings, with a single save action to commit all changes via `PUT /accounts/{id}`.

## Layout

```
┌─────────────────────────────────────────────────────┐
│ [1] Dashboard [2] Peers ... [0] Settings            │  ← top nav (existing)
├───────────────┬─────────────────────────────────────┤
│               │                                     │
│ Authentication│  Peer Session Expiration             │
│ Groups        │  ● Enabled         [30d]            │
│ Permissions   │                                     │
│ Networks      │  Peer Inactivity Expiration          │
│ Clients       │  ○ Disabled        [-]               │
│               │                                     │
│               │  Peer Approval Required  (cloud)     │
│               │  ○ Disabled                          │
│               │                                     │
│               │                                     │
│               │              [s] Save  [esc] Cancel  │
├───────────────┴─────────────────────────────────────┤
│ Settings > Authentication     ↑↓ navigate  s: save  │  ← status bar
└─────────────────────────────────────────────────────┘
```

- **Sidebar:** ~20 characters wide, left-aligned section list with highlight on active section.
- **Content area:** Fills remaining width. Shows settings for the active section.
- **Status bar:** Shows breadcrumb (`Settings > {Section}`) and key hints.

## Sections

### 1. Authentication

| Setting | Type | Field | API Field | Notes |
|---|---|---|---|---|
| Peer Session Expiration | toggle + duration | on/off + days/hours input | `peer_login_expiration_enabled`, `peer_login_expiration` | Duration in seconds, display as human-friendly |
| Peer Inactivity Expiration | toggle + duration | on/off + days/hours input | `peer_inactivity_expiration_enabled`, `peer_inactivity_expiration` | Duration in seconds |
| Peer Approval Required | toggle | on/off | `peer_approval_enabled` | Cloud-only, grayed out on self-hosted |

### 2. Groups

| Setting | Type | Field | API Field | Notes |
|---|---|---|---|---|
| User Group Propagation | toggle | on/off | `groups_propagation_enabled` | |
| JWT Group Sync | toggle | on/off | `jwt_groups_enabled` | |
| JWT Claim | text | string input | `jwt_groups_claim` | Only visible when JWT Group Sync is enabled |
| JWT Allow Groups | text | comma-separated | `jwt_allow_groups` | Only visible when JWT Group Sync is enabled |

### 3. Permissions

| Setting | Type | Field | API Field | Notes |
|---|---|---|---|---|
| Restrict Dashboard for Regular Users | toggle | on/off | `regular_users_view_blocked` | |

### 4. Networks

| Setting | Type | Field | API Field | Notes |
|---|---|---|---|---|
| DNS Domain | text | string input | `dns_domain` | |
| Network Range | text | CIDR input | `network_range` | Changing this re-allocates all peer IPs |

### 5. Clients

| Setting | Type | Field | API Field | Notes |
|---|---|---|---|---|
| Traffic Logging | toggle | on/off | `traffic_logging` | Cloud-only, grayed out on self-hosted |

## Model Updates

The `AccountSettings` model needs two new boolean fields to support the expiration toggles:

```go
PeerLoginExpirationEnabled      bool `json:"peer_login_expiration_enabled"`
PeerInactivityExpirationEnabled bool `json:"peer_inactivity_expiration_enabled"`
```

These are returned by the API but not currently captured in the model.

## Interaction Model

### Navigation

- **Up/Down** in sidebar: switch active section (content updates immediately)
- **Right arrow or Enter** from sidebar: focus moves into the content area
- **Up/Down** in content: navigate between settings fields
- **Left arrow or Esc** in content: return focus to sidebar
- **Esc** in sidebar: return focus to top nav bar

### Editing

- **Enter** on a toggle: flip the value immediately (in local state, not yet saved)
- **Enter** on a text field: enter inline edit mode — field becomes editable, type new value, Enter to confirm, Esc to cancel edit
- **Enter** on a duration field: cycle through preset values or enter custom value

### Saving

- **`s`** key: triggers save — shows confirmation screen listing all modified fields with old → new values, then `y` to confirm PUT, `n` to cancel
- **Esc** (when unsaved changes exist): prompt "Discard unsaved changes? (y/n)"
- Modified fields display a `*` marker next to them to indicate unsaved changes

### Cloud-Only Settings

- Rendered with dimmed/grayed style
- Show `(cloud)` label suffix
- Enter/toggle key presses are ignored (no-op)
- Detection: attempt to save with cloud field changed → if API returns error, mark as cloud-only for the session

## State Management

The settings page maintains two copies of settings:
1. **`original`** — the last-fetched state from the API (immutable reference)
2. **`draft`** — the working copy being edited

On save: `draft` is sent via PUT, on success `original` is updated to match `draft`.
On discard: `draft` is reset to match `original`.
Modified detection: compare `draft` fields against `original`.

## File Structure

- **`internal/tui/settings_page.go`** — main page: sidebar, section routing, save/discard logic
- **`internal/tui/settings_sections.go`** — section content renderers and field definitions
- Remove **`internal/tui/accounts_page.go`** — replaced by settings page

## API Integration

- **Load:** `GET /accounts` → populate `original` and `draft`
- **Save:** `PUT /accounts/{id}` with `AccountUpdateRequest{Settings: draft}`
- **Error handling:** display API error in status bar, keep draft intact so user can retry

## Navigation Registration

Replace `"Account"` tab in `app.go` page list with `"Settings"`. Same tab position, same hotkey slot.
