# Events

[Home](../README.md) | [Getting Started](getting-started.md) | [DNS](dns.md) | [Posture Checks](posture-checks.md) | **Events** | [Geo-Locations](geo-locations.md) | [More...](#documentation)

---

Monitor audit logs, network traffic events, and reverse proxy access logs. Events provide visibility into network activity and user actions. Running `netbird-manage event` by itself will display the help menu.

## Audit Events

```bash
# List all audit events
netbird-manage event --audit

# Filter by user ID
netbird-manage event --audit --user-id <user-id>

# Filter by target resource
netbird-manage event --audit --target-id <resource-id>

# Filter by activity type
netbird-manage event --audit --activity-code peer.create

# Filter by date range
netbird-manage event --audit --start-date "2025-01-01T00:00:00Z" --end-date "2025-01-31T23:59:59Z"

# Search in initiator/target names
netbird-manage event --audit --search "laptop"

# Export to JSON
netbird-manage event --audit --output json > audit.json
```

## Network Traffic Events (Cloud-only)

Traffic events use the flow schema: each entry is a traffic flow with `source` and `destination` endpoint objects (name, DNS label, address) plus transferred byte counters. The table output shows the last event timestamp, user, source, destination, protocol, direction, and TX/RX bytes.

```bash
# List network traffic events
netbird-manage event --traffic

# Filter by protocol (6=TCP, 17=UDP, 1=ICMP)
netbird-manage event --traffic --protocol 6

# Filter by direction
netbird-manage event --traffic --direction incoming

# Filter by event type or connection type
netbird-manage event --traffic --type <type>
netbird-manage event --traffic --connection-type <type>

# Filter by reporting peer
netbird-manage event --traffic --reporter-id <peer-id>

# Pagination
netbird-manage event --traffic --page 2 --page-size 50

# Export to JSON (full flow objects with source/destination details)
netbird-manage event --traffic --output json > traffic.json
```

## Reverse Proxy Access Logs

```bash
# List reverse proxy access logs
netbird-manage event --proxy

# Filter by source IP
netbird-manage event --proxy --source-ip 203.0.113.10

# Filter by host and path
netbird-manage event --proxy --host "app.example.com" --path "/api"

# Filter by HTTP method and status code
netbird-manage event --proxy --method GET --status-code 404

# Sort results
netbird-manage event --proxy --sort-by timestamp --sort-order desc

# Pagination
netbird-manage event --proxy --page 2 --page-size 50

# Export to JSON
netbird-manage event --proxy --output json > proxy.json
```

Proxy log filters also include `--user-id`, `--status`, `--search`, `--start-date`, and `--end-date`.

## Examples

```bash
# View recent peer creation events
netbird-manage event --audit --activity-code peer.create

# Monitor TCP traffic from the last 24 hours
netbird-manage event --traffic --protocol 6 --start-date "2025-01-15T00:00:00Z"

# Search for events related to a specific user
netbird-manage event --audit --search "admin@example.com"

# Find failed requests through the reverse proxy
netbird-manage event --proxy --status-code 502 --sort-by timestamp --sort-order desc
```

## Notes

- Audit events track all management actions (create, update, delete)
- Traffic events are an experimental feature available only on NetBird Cloud
- Traffic events follow the flow schema: source/destination objects with name, DNS label, and address, plus TX/RX byte counters per flow
- Events support both table and JSON output formats for integration with other tools

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
| [Jobs](jobs.md) | Peer jobs and debug bundles |
| [Notifications](notifications.md) | Notification channels (email/webhook) |
| [Geo-Locations](geo-locations.md) | Geographic location data |
| [Accounts](accounts.md) | Account settings and configuration |
| [Ingress Ports](ingress-ports.md) | Port forwarding (Cloud-only) |
| [Export & Import](export-import.md) | YAML/JSON configuration management |
| [Migrate](migrate.md) | Migration between NetBird accounts |

---

[Home](../README.md) | [Posture Checks](posture-checks.md) | **Events** | [Geo-Locations](geo-locations.md) | [Accounts](accounts.md) | [Ingress Ports](ingress-ports.md)
