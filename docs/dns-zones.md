# DNS Zones

[Home](../README.md) | [Getting Started](getting-started.md) | [Routes](routes.md) | [DNS](dns.md) | **DNS Zones** | [Posture Checks](posture-checks.md) | [More...](#documentation)

---

Manage custom DNS zones and their records. DNS zones let you define your own domains with A, AAAA, and CNAME records that resolve inside your NetBird network. Running `netbird-manage dns-zone` by itself will display the help menu.

## Zone Operations

```bash
# List all DNS zones
netbird-manage dns-zone --list

# Inspect a zone (including its records)
netbird-manage dns-zone --inspect <zone-id>

# Create a DNS zone
netbird-manage dns-zone --create \
  --name "internal" \
  --domain "internal.example.com" \
  --distribution-groups <group-id-1>,<group-id-2>

# Create a zone with search domain enabled
netbird-manage dns-zone --create \
  --name "corp" \
  --domain "corp.example.com" \
  --distribution-groups <group-id> \
  --search-domain true

# Create a disabled zone
netbird-manage dns-zone --create \
  --name "staging" \
  --domain "staging.example.com" \
  --distribution-groups <group-id> \
  --enabled false

# Update a zone (same parameter flags as --create)
netbird-manage dns-zone --update <zone-id> --enabled false

# Delete a zone
netbird-manage dns-zone --delete <zone-id>
```

## Record Operations

```bash
# List records in a zone
netbird-manage dns-zone --list-records <zone-id>

# Inspect a record
netbird-manage dns-zone --inspect-record <record-id> --zone <zone-id>

# Add an A record
netbird-manage dns-zone --add-record <zone-id> \
  --record-name "host.internal.example.com" \
  --record-type A \
  --content "100.64.0.10" \
  --ttl 300

# Add a CNAME record
netbird-manage dns-zone --add-record <zone-id> \
  --record-name "www.internal.example.com" \
  --record-type CNAME \
  --content "host.internal.example.com"

# Update a record
netbird-manage dns-zone --update-record <record-id> --zone <zone-id> \
  --content "100.64.0.20"

# Delete a record
netbird-manage dns-zone --delete-record <record-id> --zone <zone-id>
```

## Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `--name` | Zone name (required for `--create`) | - |
| `--domain` | Zone domain, e.g. `internal.example.com` (required for `--create`) | - |
| `--distribution-groups` | Distribution group IDs (comma-separated, required for `--create`) | - |
| `--enabled` | Enable/disable the zone: `true` or `false` | true |
| `--search-domain` | Add the zone domain as a search domain: `true` or `false` | false |
| `--record-name` | Record name (required for `--add-record`) | - |
| `--record-type` | Record type: `A`, `AAAA`, or `CNAME` (required for `--add-record`) | - |
| `--content` | Record content: IP address or hostname (required for `--add-record`) | - |
| `--ttl` | Record TTL in seconds | 300 |
| `--zone` | Zone ID for record operations | - |

## Notes

- `--inspect-record`, `--update-record`, and `--delete-record` require `--zone <zone-id>`
- Record types are limited to `A`, `AAAA`, and `CNAME`
- The zone's domain is resolved for peers in the zone's distribution groups
- For forwarding queries to external nameservers instead, see [DNS](dns.md) nameserver groups

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
| [Jobs](jobs.md) | Peer jobs and debug bundles |
| [Notifications](notifications.md) | Notification channels (email/webhook) |
| [Geo-Locations](geo-locations.md) | Geographic location data |
| [Accounts](accounts.md) | Account settings and configuration |
| [Ingress Ports](ingress-ports.md) | Port forwarding (Cloud-only) |
| [Export & Import](export-import.md) | YAML/JSON configuration management |
| [Migrate](migrate.md) | Migration between NetBird accounts |

---

[Home](../README.md) | [DNS](dns.md) | **DNS Zones** | [Posture Checks](posture-checks.md) | [Events](events.md) | [Jobs](jobs.md)
