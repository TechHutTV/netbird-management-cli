# Ingress Ports

[Home](../README.md) | [Getting Started](getting-started.md) | [Geo-Locations](geo-locations.md) | [Accounts](accounts.md) | **Ingress Ports** | [Export & Import](export-import.md) | [More...](#documentation)

---

Manage port forwarding and ingress peers. **Cloud-only feature** - only available on NetBird Cloud. Running `netbird-manage ingress-port` by itself will display the help menu.

## Port Allocation Operations

```bash
# List port allocations for a peer
netbird-manage ingress-port --list --peer <peer-id>

# Inspect a port allocation
netbird-manage ingress-port --inspect <allocation-id> --peer <peer-id>

# Create a port allocation
netbird-manage ingress-port --create --peer <peer-id> \
  --target-port 8080 \
  --protocol tcp \
  --description "Web Server"

# Update a port allocation
netbird-manage ingress-port --update <allocation-id> --peer <peer-id> \
  --target-port 8443

# Delete a port allocation
netbird-manage ingress-port --delete <allocation-id> --peer <peer-id>
```

## Ingress Peer Operations

```bash
# List all ingress peers
netbird-manage ingress-peer --list

# Inspect an ingress peer
netbird-manage ingress-peer --inspect <ingress-peer-id>

# Create an ingress peer
netbird-manage ingress-peer --create \
  --name "US West" \
  --location us-west-1

# Update an ingress peer
netbird-manage ingress-peer --update <ingress-peer-id> \
  --enabled false

# Delete an ingress peer
netbird-manage ingress-peer --delete <ingress-peer-id>
```

## Examples

```bash
# Forward port 8080 on a peer to a public port
netbird-manage ingress-port --create --peer d41uqobl0ubs73bkuhqg \
  --target-port 8080 \
  --protocol tcp \
  --description "Production Web Server"

# List all port allocations for a specific peer
netbird-manage ingress-port --list --peer d41uqobl0ubs73bkuhqg

# Create a new ingress peer for EU region
netbird-manage ingress-peer --create \
  --name "EU Central" \
  --location eu-central-1

# Disable an ingress peer
netbird-manage ingress-peer --update ing-001 --enabled false
```

## Notes

- Ingress ports are **Cloud-only** - not available on self-hosted instances
- Public ports are automatically assigned by NetBird Cloud
- Protocol options: `tcp` (default) or `udp`
- Target ports must be between 1-65535

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
| [Export & Import](export-import.md) | YAML/JSON configuration management |
| [Migrate](migrate.md) | Migration between NetBird accounts |

---

[Home](../README.md) | [Accounts](accounts.md) | **Ingress Ports** | [Export & Import](export-import.md) | [Migrate](migrate.md)
