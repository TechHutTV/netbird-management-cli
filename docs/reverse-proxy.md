# Reverse Proxy

Manage NetBird reverse-proxy services — a public-facing domain pointing at a
private backend inside your NetBird network, with optional authentication,
access control, and custom domains.

The feature lives under the `Proxy` tab (hotkey `[0]`) in the TUI.

> **Note:** This feature is TUI-only in this CLI. There is no `netbird-manage
> reverse-proxy` subcommand yet — use the dashboard at
> `https://app.netbird.io/reverse-proxy/services` or the `Proxy` tab.

## REST API

All endpoints sit under `/api/reverse-proxies/...` on the NetBird Management
server and accept the standard `Authorization: Token <PAT>` header.

| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/reverse-proxies/clusters` | List proxy clusters (region, supports_custom_ports flag) |
| GET | `/reverse-proxies/services` | List all services |
| POST | `/reverse-proxies/services` | Create a service |
| GET | `/reverse-proxies/services/{id}` | Retrieve one service |
| PUT | `/reverse-proxies/services/{id}` | Update a service (send the full object) |
| DELETE | `/reverse-proxies/services/{id}` | Delete a service |
| GET | `/reverse-proxies/services/{id}/domains` | List custom domains |
| POST | `/reverse-proxies/services/{id}/domains` | Attach a custom domain |
| GET | `/reverse-proxies/services/{id}/domains/{domain_id}/validate` | Run DNS validation |
| DELETE | `/reverse-proxies/services/{id}/domains/{domain_id}` | Remove a custom domain |
| GET | `/events/proxy` | Access-log events (filter by `service_id`, `start_date`, `end_date`) |

## Concepts

- **Mode.** Each service is HTTP, TCP, UDP, or TLS. HTTP supports multiple
  targets with per-target paths, auth, and HTTP header rules. L4 modes
  (TCP/UDP/TLS) support exactly one target and no auth.
- **Target.** A backend destination. Either a NetBird peer (by peer ID), a
  raw host, a domain, or a subnet — plus protocol and port. HTTP targets can
  also include custom request headers, TLS skip, and a request timeout.
- **Auth.** HTTP services can require password, PIN, magic-link, SSO (bearer
  via user groups), or header-match. Stack is additive — any enabled method
  satisfies the gate.
- **Access Restrictions.** Allow/block lists of CIDRs or ISO country codes.
  Single IPs are stored as `/32` CIDR; the UI hides the `/32` when showing them.
- **Custom Domains.** Bring your own domain instead of using a NetBird-issued
  subdomain. Validation is DNS-based (TXT record).

## TUI walkthrough

### List view

`Proxy` tab. Columns: domain, mode, status, enabled, target count, auth
summary, access summary, cluster. Shortcuts:

| Key | Action |
|-----|--------|
| `↑ ↓` / `j k` | Move cursor |
| `enter` | Open detail view |
| `c` | Create a new service (launches the wizard) |
| `e` | Edit the selected service (launches the wizard) |
| `t` | Toggle enable/disable |
| `d` | Delete the selected service |
| `/` | Filter by name / domain / mode / cluster |
| `r` | Refresh from the API |

### Detail view

Tabbed read-only renderer: `Service` / `Auth` / `Access` / `Advanced` /
`Meta`. `←`/`→` switch tabs. Extra shortcuts while in detail:

| Key | Action |
|-----|--------|
| `D` | Drill into custom domains for this service |
| `E` | Drill into proxy events for this service |
| `e` | Open the edit wizard |
| `t` | Toggle enable/disable |
| `d` | Delete |

### Create / Edit wizard

A multi-screen flow, one step at a time:

1. **Service** — name, domain, mode (locked on edit), listen port, cluster, enabled.
2. **Targets** — list editor. `enter` on a row edits it; `enter` on `[+ Add]`
   adds a new one. `h` on a target row opens a custom-headers list editor.
   `c` continues to the next step.
3. **Authentication** (HTTP only) — 4 toggleable methods plus a header-rules
   sub-list. `enter` configures a method; `c` continues.
4. **Access Control** — list of rules (action × type × value). `c` continues.
5. **Advanced** — mode-gated. HTTP: pass host header, rewrite redirects.
   TCP/TLS: proxy protocol, request timeout. UDP: session idle timeout.
6. **Review & Submit** — summary + `y/n`.

An extra **Unprotected Service** warning screen appears if an HTTP service is
being submitted with no auth AND no access rules; `y` accepts, `n`/`esc` goes
back to Access so protection can be added.

### Custom domains

Reached via `D` from the detail view. Shortcuts: `a` add, `v` validate, `d`
delete, `r` refresh.

### Proxy events

Reached via `E` from the detail view. Columns: time, method, host+path,
status, auth method, location, bytes, duration. `1` / `2` / `3` change the
range to 1d / 7d / 30d; `r` refreshes.

## Notes and caveats

- **Full-object PUT.** Toggling `enabled` or editing any field re-sends the
  complete service object, including all targets and auth config. This
  matches the public API contract.
- **Mode is immutable** after create; the service form locks it on edit.
- **L4 constraints** (TCP/UDP/TLS). Exactly one target; no HTTP-specific
  settings (path, custom headers, `pass_host_header`, `rewrite_redirects`).
- **Durations** use Go format — `30s`, `5m`, `1h`. Empty means "no timeout".
- **Certificate status polling** is not implemented. Refresh with `r`.
