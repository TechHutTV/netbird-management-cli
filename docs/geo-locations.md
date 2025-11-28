# Geo-Locations

[Home](../README.md) | [Getting Started](getting-started.md) | [Posture Checks](posture-checks.md) | [Events](events.md) | **Geo-Locations** | [Accounts](accounts.md) | [More...](#documentation)

---

Retrieve geographic location data for use in posture checks and access policies. Running `netbird-manage geo` by itself will display the help menu.

## Query Operations

```bash
# List all country codes
netbird-manage geo --countries

# List cities in a specific country
netbird-manage geo --cities --country DE
netbird-manage geo --cities --country US

# Export to JSON
netbird-manage geo --countries --output json
netbird-manage geo --cities --country FR --output json
```

## Examples

```bash
# Get all available country codes
netbird-manage geo --countries

# Find cities in Germany for geo-location posture checks
netbird-manage geo --cities --country DE

# Export US cities to JSON for automation
netbird-manage geo --cities --country US --output json > us-cities.json
```

## Notes

- Country codes follow ISO 3166-1 alpha-2 standard (e.g., US, GB, DE, FR)
- City data includes geoname IDs for precise location matching
- Use geo-location data when creating posture checks with `--type geo-location`

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
| [Accounts](accounts.md) | Account settings and configuration |
| [Ingress Ports](ingress-ports.md) | Port forwarding (Cloud-only) |
| [Export & Import](export-import.md) | YAML/JSON configuration management |
| [Migrate](migrate.md) | Migration between NetBird accounts |

---

[Home](../README.md) | [Events](events.md) | **Geo-Locations** | [Accounts](accounts.md) | [Ingress Ports](ingress-ports.md) | [Export & Import](export-import.md)
