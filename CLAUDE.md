# CLAUDE.md - NetBird Management CLI

## Project Overview

Unofficial Go CLI + TUI for managing NetBird networks via the REST API and local daemon gRPC.

- **Language:** Go 1.25+
- **Binary:** `netbird-manage` (single binary, `cmd/netbird-manage/main.go`)
- **API:** NetBird REST API with Bearer token auth (`Token <pat>`)
- **TUI:** Charm ecosystem (bubbletea v2, lipgloss v2, huh v2) — `netbird-manage tui`
- **Daemon:** Optional gRPC connection to local NetBird daemon (`unix:///var/run/netbird.sock`)

## Codebase Structure

```
cmd/netbird-manage/main.go          — Entry point, command router, global flags
internal/
  client/client.go                   — HTTP API client (MakeRequest, auth, debug logging)
  config/config.go                   — Config file management, auto-detect /api suffix
  models/models.go                   — All API data types (~630 lines, 40+ types)
  helpers/helpers.go                 — Validation, confirmations, duration parsing
  commands/                          — CLI command handlers (one file per resource)
    service.go, usage.go, peers.go, groups.go, networks.go, policies.go,
    setup_keys.go, users.go, tokens.go, routes.go, dns.go, posture_checks.go,
    events.go, geo_locations.go, accounts.go,
    migrate.go, export.go, import.go
  tui/                               — Terminal UI (Charm/bubbletea v2)
    app.go                           — Root model, navigation, focus management
    nav.go                           — Top bar with numbered hotkeys [1]-[0]
    layout.go                        — Header + nav bar + content + status bar layout
    theme.go                         — Adaptive color palette (dark/light detection)
    daemon.go                        — gRPC client wrapper for local NetBird daemon
    dashboard_page.go                — Status page (daemon status + management API counts)
    peers_page.go                    — Peers (list, detail, search, accessible peers)
    groups_page.go, networks_page.go, policies_page.go, routes_page.go,
    setup_keys_page.go, users_page.go, service_users_page.go, dns_page.go,
    posture_checks_page.go, events_page.go, accounts_page.go,
    export_import_page.go            — Entity pages (list + detail + create forms)
    reverse_proxy_page.go            — Reverse Proxy list + outer state machine
    reverse_proxy_wizard.go          — Multi-screen create/edit wizard (service → targets → auth → access → advanced → confirm)
    reverse_proxy_forms.go           — huh/v2 form constructors for every wizard leaf screen
    reverse_proxy_editors.go         — Reusable list-editor renderer (targets, access rules, header auths, custom headers)
    reverse_proxy_detail.go          — Tabbed read-only detail view (Service / Auth / Access / Advanced / Meta)
    reverse_proxy_domains.go         — Custom domains sub-view (add / validate / delete)
    reverse_proxy_events.go          — Proxy access-log sub-view with date-range filter
    forms.go                         — huh/v2 form constructors for all create operations
    confirm.go                       — Confirmation screen renderer (y/n before submit)
    messages.go                      — All custom tea.Msg types
    api.go                           — tea.Cmd factories for API calls
    helpers.go, keys.go, run.go, placeholder.go
    components/                      — Reusable UI components (statusbar, detail)
```

## Building & Running

```bash
go build -o netbird-manage ./cmd/netbird-manage/   # Build
./netbird-manage connect --token "nbp_..." --management-url "https://example.com"
./netbird-manage tui                                # Launch TUI
./netbird-manage peer --list                        # CLI mode
```

## TUI Architecture

- **Elm Architecture** via bubbletea v2: Model → Update → View
- **Import paths:** `charm.land/bubbletea/v2`, `charm.land/bubbles/v2`, `charm.land/lipgloss/v2`, `charm.land/huh/v2`
- **Key v2 API:** `Init() tea.Cmd` (not `(Model, Cmd)`), `View() tea.View` (use `tea.NewView(s)`), `tea.KeyPressMsg` (not `KeyMsg`)
- **Page interface:** `Init`, `Update`, `View`, `Title`, `CursorPosition`, `SetFocused`
- **All pages in same package** (`internal/tui/`) to avoid import cycles
- **Navigation:** numbered hotkeys `[1]`-`[9]` then `[0]`, left/right arrows, enter to load, down to focus content, up at cursor 0 returns to nav. Hotkey order (left to right): Status, Peers, Policies, Networks, DNS, Routes, Users, Groups, Posture, **Proxy**. Arrow-only tabs after: Events, Service Users, Keys, Settings, Export.
- **Focus model:** nav bar vs content — `focused` field on each page controls cursor highlight visibility
- **Forms:** huh/v2 with `MultiSelect` for groups, `esc` to cancel, confirmation screen before submit
- **Daemon:** optional gRPC to `unix:///var/run/netbird.sock`, graceful nil if unavailable

## Key Conventions

- **Auth header:** `Authorization: Token <pat>` (not Bearer)
- **Config:** `$HOME/.netbird-manage.json` with `0600` permissions
- **URL auto-detect:** `config.TestAndSave` checks Content-Type is JSON, auto-appends `/api` for self-hosted
- **Group updates require full PUT:** fetch group → modify → PUT entire object
- **Always close response bodies:** `defer resp.Body.Close()` after error check
- **Always PathEscape IDs in URLs:** `url.PathEscape(id)` on every URL concatenation
- **Always check json.Marshal errors:** return `APIErrorMsg` on failure
- **Confirmation prompts:** all CLI destructive operations use `confirmSingleDeletion`/`confirmBulkDeletion`
- **TUI forms use confirm screen:** form → review summary → y/n before API call

## API Endpoints

All through `client.MakeRequest(method, endpoint, body)`:

| Resource | Endpoints | CLI File | TUI Page |
|----------|-----------|----------|----------|
| Peers | GET/PUT/DELETE `/peers`, `/peers/{id}/accessible-peers` | peers.go | peers_page.go |
| Groups | GET/POST/PUT/DELETE `/groups` | groups.go | groups_page.go |
| Networks | `/networks`, `/{id}/resources`, `/{id}/routers` | networks.go | networks_page.go |
| Policies | GET/POST/PUT/DELETE `/policies` | policies.go | policies_page.go |
| Routes | GET/POST/PUT/DELETE `/routes` | routes.go | routes_page.go |
| Setup Keys | GET/POST/PUT/DELETE `/setup-keys` | setup_keys.go | setup_keys_page.go |
| Users | GET/POST/PUT/DELETE `/users`, `/users/{id}/invite` | users.go | users_page.go |
| DNS | `/dns/nameservers`, `/dns/settings` | dns.go | dns_page.go |
| Posture | GET/POST/PUT/DELETE `/posture-checks` | posture_checks.go | posture_checks_page.go |
| Events | GET `/events/audit`, GET `/events/proxy` | events.go | events_page.go |
| Accounts | GET/PUT/DELETE `/accounts` | accounts.go | accounts_page.go |
| Reverse Proxy | GET/POST/PUT/DELETE `/reverse-proxies/services`, GET `/reverse-proxies/clusters`, GET/POST/DELETE `/reverse-proxies/services/{id}/domains`, GET `/reverse-proxies/services/{id}/domains/{domain_id}/validate` | — (TUI only) | reverse_proxy_page.go + reverse_proxy_{wizard,forms,editors,detail,domains,events}.go |

## Reverse Proxy feature notes

- **TUI-only** — no CLI parity yet. Lives under the `Proxy` tab (`[0]`).
- **Full PUT required** on update (matches groups pattern). Use `ReverseProxyUpdateFromService` to rebuild the request from a cached service.
- **Mode is immutable** post-create; edit form shows it read-only.
- **L4 modes (TCP/UDP/TLS)** allow exactly one target and no auth.
- **Access rules** — single IPs are stored as `/32` CIDR on write, displayed without the suffix on read. See `accessBucketFor` / `accessBucketOfFlattened` in `reverse_proxy_wizard.go` for the flattened-index routing.
- **Cluster gating** — `listenPortValidator` rejects custom listen ports when the selected cluster reports `supports_custom_ports=false`.
- **Unprotected warning** — HTTP services with no auth AND no access rules trigger a y/n gate before submit.
- **Deferred:** certificate status polling (dashboard polls every 3.5s; we refetch on demand via `r`), CLI command parity, inline target expansion on list rows.

## Dependencies

- `gopkg.in/yaml.v3` — YAML export/import
- `charm.land/bubbletea/v2` — TUI framework
- `charm.land/bubbles/v2` — TUI components
- `charm.land/lipgloss/v2` — TUI styling (including `lipgloss/v2/table`)
- `charm.land/huh/v2` — TUI forms
- `github.com/netbirdio/netbird` — gRPC proto definitions for daemon
- `google.golang.org/grpc` — gRPC client

## Security

- Never commit tokens, API keys, or test results. `.netbird-manage.json` is in `.gitignore`.
- Config file uses `0600` permissions; export files and directories use `0600/0700` (not world-readable).
- All user IDs in URLs must use `url.PathEscape` — convention enforced in both `internal/tui/api.go` and `internal/commands/*.go`.
- `json.Marshal` errors must be checked (not discarded with `_`).
- **Debug-log redaction:** `client.redactSensitiveJSON` masks `password`, `pin`, `token`, `secret`, `api_key`, `access_token`, `refresh_token`, `client_secret`, and `header_auths[].value` in `--debug` JSON dumps. Callers still receive unredacted bytes — redaction applies to stderr output only.
- **Path traversal:** YAML directory imports use `commands.safeJoinUnder` to reject any `import_order` entry that resolves outside the import directory. Covers `..` escapes; leading `/` is absorbed by `filepath.Join` (safe).
- **Reverse-proxy secrets:** user-supplied passwords, PINs, and header-auth values never appear in list/summary views — only the edit form echoes them (passwords and PINs masked via `huh.EchoModePassword`). Header values are visible in the Auth detail tab.
