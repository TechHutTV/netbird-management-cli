# DNS Zones

[Home](../README.md) | [Getting Started](getting-started.md) | [DNS](dns.md) | **DNS Zones** | [Identity Providers](identity-providers.md) | [More...](#documentation)

---

Manage custom DNS zones and their records. This is separate from DNS nameserver groups (`dns` command) - DNS zones let you define custom zones with A, AAAA, and CNAME records. Running `netbird-manage dns-zone` by itself will display the help menu.

## Zone Operations

### Query

```bash
# List all DNS zones
netbird-manage dns-zone --list

# Inspect a specific zone
netbird-manage dns-zone --inspect <zone-id>

# JSON output
netbird-manage dns-zone --list --output json
```

### Create/Update/Delete

```bash
# Create a DNS zone
netbird-manage dns-zone --create "internal-zone" \
  --domain "internal.example.com" \
  --groups "group-id-1,group-id-2" \
  --search-domain \
  --enabled

# Update a zone
netbird-manage dns-zone --update <zone-id> \
  --name "new-name" \
  --domain "new.example.com" \
  --groups "group-id-3"

# Disable a zone
netbird-manage dns-zone --update <zone-id> --disabled

# Delete a zone (with confirmation)
netbird-manage dns-zone --delete <zone-id>
```

## Record Operations

### Query

```bash
# List records in a zone
netbird-manage dns-zone --list-records <zone-id>

# Inspect a specific record
netbird-manage dns-zone --inspect-record --zone-id <zone-id> --record-id <record-id>
```

### Create/Update/Delete

```bash
# Add an A record
netbird-manage dns-zone --add-record <zone-id> \
  --name "app.internal.example.com" \
  --type A \
  --content "10.0.0.5" \
  --ttl 300

# Add a CNAME record
netbird-manage dns-zone --add-record <zone-id> \
  --name "www.internal.example.com" \
  --type CNAME \
  --content "app.internal.example.com"

# Update a record
netbird-manage dns-zone --update-record \
  --zone-id <zone-id> \
  --record-id <record-id> \
  --content "10.0.0.10"

# Delete a record (with confirmation)
netbird-manage dns-zone --delete-record \
  --zone-id <zone-id> \
  --record-id <record-id>
```

## Record Types

| Type | Description | Content Format |
|------|-------------|----------------|
| `A` | IPv4 address record | `10.0.0.5` |
| `AAAA` | IPv6 address record | `2001:db8::1` |
| `CNAME` | Canonical name record | `target.example.com` |

## Notes

- DNS zones are for custom domain resolution within your NetBird network
- Distribution groups control which peers can resolve the zone
- Search domain enables the zone to be used for short hostname resolution
- Record names must be FQDNs within the zone's domain

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
| [Accounts](accounts.md) | Account settings and configuration |
| [Ingress Ports](ingress-ports.md) | Port forwarding (Cloud-only) |
| [Identity Providers](identity-providers.md) | OIDC/OAuth2 identity provider management |
| [Instance](instance.md) | Self-hosted instance setup |
| [Jobs](jobs.md) | Peer bundle collection and debugging |
| [Export & Import](export-import.md) | YAML/JSON configuration management |
| [Migrate](migrate.md) | Migration between NetBird accounts |

---

[Home](../README.md) | [DNS](dns.md) | **DNS Zones** | [Identity Providers](identity-providers.md) | [Instance](instance.md) | [Jobs](jobs.md)
