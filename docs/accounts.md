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

# Update peer inactivity expiration
netbird-manage account --update <account-id> --peer-inactivity-expiration 30d

# Update DNS domain
netbird-manage account --update <account-id> --dns-domain nb.local

# Update network range
netbird-manage account --update <account-id> --network-range 100.64.0.0/10

# Enable JWT groups
netbird-manage account --update <account-id> --jwt-groups-enabled true

# Update multiple settings at once
netbird-manage account --update <account-id> \
  --peer-login-expiration 24h \
  --dns-domain company.local \
  --jwt-groups-enabled true
```

## Delete Operations

```bash
# Delete account (requires confirmation, deletes ALL resources)
netbird-manage account --delete <account-id> --confirm
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
- Some settings like `peer-approval-enabled` and `traffic-logging` are Cloud-only
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
| [Posture Checks](posture-checks.md) | Device compliance validation |
| [Events](events.md) | Audit logs and traffic monitoring |
| [Geo-Locations](geo-locations.md) | Geographic location data |
| [Ingress Ports](ingress-ports.md) | Port forwarding (Cloud-only) |
| [Export & Import](export-import.md) | YAML/JSON configuration management |
| [Migrate](migrate.md) | Migration between NetBird accounts |

---

[Home](../README.md) | [Geo-Locations](geo-locations.md) | **Accounts** | [Ingress Ports](ingress-ports.md) | [Export & Import](export-import.md) | [Migrate](migrate.md)
