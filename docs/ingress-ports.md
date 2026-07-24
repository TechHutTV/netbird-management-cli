# Ingress Ports

[Home](../README.md) | [Getting Started](getting-started.md) | [Geo-Locations](geo-locations.md) | [Accounts](accounts.md) | **Ingress Ports** | [Export & Import](export-import.md) | [More...](#documentation)

---

Manage port forwarding and ingress peers. **Cloud-only feature** - only available on NetBird Cloud. Running `netbird-manage ingress-port` by itself will display the help menu.

## Port Allocation Operations

Port allocations are named and forward one or more port ranges on a peer. All allocation operations require `--peer <peer-id>`.

```bash
# List port allocations for a peer
netbird-manage ingress-port --list --peer <peer-id>

# Filter allocations by name (server-side)
netbird-manage ingress-port --list --peer <peer-id> --filter-name "web"

# Inspect a port allocation
netbird-manage ingress-port --inspect <allocation-id> --peer <peer-id>

# Create a port allocation with port ranges
netbird-manage ingress-port --create --peer <peer-id> \
  --name "web-server" \
  --port-ranges "80:tcp,443:tcp,1000-2000:udp"

# Create a port allocation with direct ports
netbird-manage ingress-port --create --peer <peer-id> \
  --name "game-server" \
  --direct-ports "3:tcp"

# Create a disabled allocation
netbird-manage ingress-port --create --peer <peer-id> \
  --name "staging" \
  --port-ranges "8080:tcp" \
  --enabled false

# Update a port allocation (unset flags keep current values)
netbird-manage ingress-port --update <allocation-id> --peer <peer-id> \
  --port-ranges "8443:tcp" \
  --enabled true

# Delete a port allocation
netbird-manage ingress-port --delete <allocation-id> --peer <peer-id>
```

### Allocation Options

| Option | Description |
|--------|-------------|
| `--peer` | Peer ID (required for all operations) |
| `--name` | Allocation name (required for `--create`) |
| `--port-ranges` | Comma-separated ranges as `start-end:protocol` (e.g., `80:tcp,1000-2000:udp,443:tcp/udp`) |
| `--direct-ports` | Direct port mapping as `count:protocol` (e.g., `3:tcp`) |
| `--enabled` | Enable/disable the allocation: `true` or `false` (default: true) |
| `--filter-name` | Filter allocations by name (with `--list`) |

## Ingress Peer Operations

```bash
# List all ingress peers
netbird-manage ingress-peer --list

# Inspect an ingress peer
netbird-manage ingress-peer --inspect <ingress-peer-id>

# Convert an existing peer into an ingress peer
netbird-manage ingress-peer --create --peer <peer-id>

# Create a disabled fallback ingress peer
netbird-manage ingress-peer --create --peer <peer-id> \
  --enabled false \
  --fallback true

# Update an ingress peer
netbird-manage ingress-peer --update <ingress-peer-id> \
  --enabled false \
  --fallback true

# Delete an ingress peer
netbird-manage ingress-peer --delete <ingress-peer-id>
```

## Examples

```bash
# Forward web ports on a peer
netbird-manage ingress-port --create --peer d41uqobl0ubs73bkuhqg \
  --name "production-web" \
  --port-ranges "80:tcp,443:tcp"

# Allocate 3 direct TCP ports
netbird-manage ingress-port --create --peer d41uqobl0ubs73bkuhqg \
  --name "direct-tcp" \
  --direct-ports "3:tcp"

# List all port allocations for a specific peer
netbird-manage ingress-port --list --peer d41uqobl0ubs73bkuhqg

# Convert a peer into an ingress peer
netbird-manage ingress-peer --create --peer d41uqobl0ubs73bkuhqg

# Disable an ingress peer
netbird-manage ingress-peer --update ing-001 --enabled false
```

## Notes

- Ingress ports are **Cloud-only** - not available on self-hosted instances
- `--create` requires `--name` and at least one of `--port-ranges` or `--direct-ports`
- Protocol options: `tcp`, `udp`, or `tcp/udp`
- Ports must be between 1-65535 and range end must be >= start
- Inspect output shows the assigned ingress-to-translated port mappings (e.g., `40000->80/tcp`)
- Ingress peers are created from existing peers with `--peer <peer-id>`; a fallback ingress peer (`--fallback true`) is used when primary ingress peers are unavailable

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
| [Accounts](accounts.md) | Account settings and configuration |
| [Export & Import](export-import.md) | YAML/JSON configuration management |
| [Migrate](migrate.md) | Migration between NetBird accounts |

---

[Home](../README.md) | [Accounts](accounts.md) | **Ingress Ports** | [Export & Import](export-import.md) | [Migrate](migrate.md)
