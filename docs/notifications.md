# Notifications

[Home](../README.md) | [Getting Started](getting-started.md) | [Events](events.md) | [Jobs](jobs.md) | **Notifications** | [Geo-Locations](geo-locations.md) | [More...](#documentation)

---

Manage notification channels that send email or webhook alerts when account events occur. Running `netbird-manage notification` by itself will display the help menu.

## Query Operations

```bash
# List all notification channels
netbird-manage notification --list

# List available event types to subscribe to
netbird-manage notification --list-types

# Inspect a notification channel
netbird-manage notification --inspect <channel-id>

# Export to JSON
netbird-manage notification --list --output json
```

## Modification Operations

```bash
# Create an email notification channel
netbird-manage notification --create \
  --type email \
  --emails "admin@example.com,ops@example.com" \
  --event-types "<event-type-1>,<event-type-2>"

# Create a webhook notification channel
netbird-manage notification --create \
  --type webhook \
  --url "https://hooks.example.com/netbird" \
  --event-types "<event-type-1>"

# Create a disabled channel
netbird-manage notification --create \
  --type email \
  --emails "admin@example.com" \
  --event-types "<event-type>" \
  --enabled false

# Update a channel (same parameter flags as --create)
netbird-manage notification --update <channel-id> --enabled false

# Delete a channel
netbird-manage notification --delete <channel-id>
```

## Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `--type` | Channel type: `email` or `webhook` (required for `--create`) | - |
| `--emails` | Recipient emails (comma-separated, required for `--type email`) | - |
| `--url` | Webhook URL (required for `--type webhook`) | - |
| `--event-types` | Event types to notify on (comma-separated, required for `--create`) | - |
| `--enabled` | Enable/disable the channel: `true` or `false` | true |

## Examples

```bash
# See which event types are available
netbird-manage notification --list-types

# Alert the ops team by email
netbird-manage notification --create \
  --type email \
  --emails "ops@example.com" \
  --event-types "<event-type>"

# Temporarily disable a channel
netbird-manage notification --update ch-001 --enabled false
```

## Notes

- Use `--list-types` to discover the event types you can subscribe channels to
- `--type email` requires `--emails`; `--type webhook` requires `--url`
- Channels can be disabled without deleting them via `--update <channel-id> --enabled false`

---

## Documentation

| Section | Description |
|---------|-------------|
| [Getting Started](getting-started.md) | Installation, safety features, debug mode |
| [Peers](peers.md) | Manage network peers |
| [Setup Keys](setup-keys.md) | Device registration and onboarding keys |
| [Users](users.md) | User management and invitations |
| [Tokens](tokens.md) | Personal access token management |
| [Groups](groups.md) | Peer group management |
| [Networks](networks.md) | Networks, resources, and routers |
| [Policies](policies.md) | Access control policies and firewall rules |
| [Routes](routes.md) | Network routing configuration |
| [DNS](dns.md) | DNS nameserver groups and settings |
| [DNS Zones](dns-zones.md) | Custom DNS zones and records |
| [Posture Checks](posture-checks.md) | Device compliance validation |
| [Events](events.md) | Audit logs and traffic monitoring |
| [Jobs](jobs.md) | Peer jobs and debug bundles |
| [Geo-Locations](geo-locations.md) | Geographic location data |
| [Accounts](accounts.md) | Account settings and configuration |
| [Ingress Ports](ingress-ports.md) | Port forwarding (Cloud-only) |
| [Export & Import](export-import.md) | YAML/JSON configuration management |
| [Migrate](migrate.md) | Migration between NetBird accounts |

---

[Home](../README.md) | [Jobs](jobs.md) | **Notifications** | [Geo-Locations](geo-locations.md) | [Accounts](accounts.md) | [Ingress Ports](ingress-ports.md)
