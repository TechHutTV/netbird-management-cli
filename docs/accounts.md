# Accounts

[Home](../README.md) | [Getting Started](getting-started.md) | [Events](events.md) | [Geo-Locations](geo-locations.md) | **Accounts** | [Ingress Ports](ingress-ports.md) | [More...](#documentation)

---

Manage account settings and configuration. Running `netbird-manage account` by itself will display the help menu.

## Query Operations

```bash
# List all accounts (returns current user's account)
netbird-manage account --list

# Inspect account details
netbird-manage account --inspect <account-id>
```

## Update Operations

```bash
# Update peer login expiration
netbird-manage account --update <account-id> --peer-login-expiration 48h

# Enable/disable peer login expiration
netbird-manage account --update <account-id> --peer-login-expiration-enabled true

# Update peer inactivity expiration
netbird-manage account --update <account-id> --peer-inactivity-expiration 30d

# Enable/disable peer inactivity expiration
netbird-manage account --update <account-id> --peer-inactivity-expiration-enabled true

# Update DNS domain
netbird-manage account --update <account-id> --dns-domain nb.local

# Update network range
netbird-manage account --update <account-id> --network-range 100.64.0.0/10

# Update IPv6 network range
netbird-manage account --update <account-id> --network-range-v6 "fd00:b14d::/48"

# Enable DNS resolution on routing peers
netbird-manage account --update <account-id> --routing-peer-dns-resolution-enabled true

# Enable lazy connections
netbird-manage account --update <account-id> --lazy-connection-enabled true

# Enable JWT groups
netbird-manage account --update <account-id> --jwt-groups-enabled true

# Update multiple settings at once
netbird-manage account --update <account-id> \
  --peer-login-expiration 24h \
  --dns-domain company.local \
  --jwt-groups-enabled true
```

### Update Flags

| Flag | Description |
|------|-------------|
| `--peer-login-expiration <dur>` | Peer login expiration (e.g., `24h`, `7d`) |
| `--peer-login-expiration-enabled <true\|false>` | Enable/disable peer login expiration |
| `--peer-inactivity-expiration <dur>` | Peer inactivity timeout (e.g., `30d`) |
| `--peer-inactivity-expiration-enabled <true\|false>` | Enable/disable peer inactivity expiration |
| `--dns-domain <domain>` | Network DNS domain |
| `--network-range <cidr>` | Network IP range (e.g., `100.64.0.0/10`) |
| `--network-range-v6 <cidr>` | IPv6 network range |
| `--routing-peer-dns-resolution-enabled <true\|false>` | Enable DNS resolution on routing peers |
| `--jwt-groups-enabled <true\|false>` | Enable JWT group claims |
| `--jwt-groups-claim <name>` | JWT claim name for groups (API field `jwt_groups_claim_name`) |
| `--jwt-allow-groups <groups>` | Comma-separated allowed groups |
| `--groups-propagation-enabled <true\|false>` | Enable groups propagation |
| `--regular-users-view-blocked <true\|false>` | Block regular users view |
| `--lazy-connection-enabled <true\|false>` | Enable lazy connections |
| `--peer-approval-enabled <true\|false>` | Enable peer approval (Cloud-only) |
| `--user-approval-required <true\|false>` | Require user approval (Cloud-only) |
| `--traffic-logging <true\|false>` | Enable network traffic logging (Cloud-only) |

## Delete Operations

```bash
# Delete account (requires confirmation, deletes ALL resources)
netbird-manage account --delete <account-id>
```

## Examples

```bash
# View current account settings
netbird-manage account --list

# Set peer login expiration to 2 days
netbird-manage account --update d10vfhbl0ubs73e6p8ig --peer-login-expiration 48h

# Configure JWT group claims
netbird-manage account --update d10vfhbl0ubs73e6p8ig \
  --jwt-groups-enabled true \
  --jwt-groups-claim groups \
  --jwt-allow-groups "engineering,ops,security"
```

## Notes

- Duration format: `24h` (hours), `7d` (days), `30d` (days)
- `--jwt-groups-claim` maps to the API's `jwt_groups_claim_name` settings field
- The Cloud-only settings (`--peer-approval-enabled`, `--user-approval-required`, `--traffic-logging`) live in the account's nested `extra` settings object; the CLI updates them there automatically
- `account --list` and `account --inspect` also display the `extra` Cloud settings and onboarding status when present
- Deleting an account is permanent and removes ALL associated resources

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
| [Notifications](notifications.md) | Notification channels (email/webhook) |
| [Geo-Locations](geo-locations.md) | Geographic location data |
| [Ingress Ports](ingress-ports.md) | Port forwarding (Cloud-only) |
| [Export & Import](export-import.md) | YAML/JSON configuration management |
| [Migrate](migrate.md) | Migration between NetBird accounts |

---

[Home](../README.md) | [Geo-Locations](geo-locations.md) | **Accounts** | [Ingress Ports](ingress-ports.md) | [Export & Import](export-import.md) | [Migrate](migrate.md)
