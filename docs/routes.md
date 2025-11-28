# Routes

[Home](../README.md) | [Getting Started](getting-started.md) | [Networks](networks.md) | [Policies](policies.md) | **Routes** | [DNS](dns.md) | [More...](#documentation)

---

Manage network routes and routing configuration. Routes define how traffic flows through your NetBird network. Running `netbird-manage route` by itself will display the help menu.

## Query Operations

```bash
# List all routes
netbird-manage route --list

# Filter routes by network CIDR pattern
netbird-manage route --list --filter-network "10.0"

# Filter by routing peer
netbird-manage route --list --filter-peer <peer-id>

# Show only enabled routes
netbird-manage route --list --enabled-only

# Show only disabled routes
netbird-manage route --list --disabled-only

# Inspect a specific route
netbird-manage route --inspect <route-id>
```

## Modification Operations

```bash
# Create a route for 10.0.0.0/16 network
netbird-manage route --create "10.0.0.0/16" \
  --network-id <network-id> \
  --peer <peer-id> \
  --groups <group-id> \
  --metric 100 \
  --masquerade

# Create a route using peer groups instead of single peer
netbird-manage route --create "192.168.0.0/16" \
  --network-id <network-id> \
  --peer-groups "router-group-1,router-group-2" \
  --groups <group-id> \
  --metric 50

# Create a disabled route with description
netbird-manage route --create "172.16.0.0/12" \
  --network-id <network-id> \
  --peer <peer-id> \
  --groups <group-id> \
  --description "Private network route" \
  --disabled

# Update route metric (priority)
netbird-manage route --update <route-id> --metric 50

# Enable/disable a route
netbird-manage route --enable <route-id>
netbird-manage route --disable <route-id>

# Delete a route
netbird-manage route --delete <route-id>
```

## Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `--network-id` | Target network ID (required) | - |
| `--peer` | Single routing peer ID (use OR `--peer-groups`) | - |
| `--peer-groups` | Peer group IDs for high-availability routing (use OR `--peer`) | - |
| `--metric` | Route priority (1-9999, lower = higher priority) | 100 |
| `--masquerade` | Enable masquerading/NAT | false |
| `--no-masquerade` | Disable masquerading | true |
| `--groups` | Access group IDs (required, comma-separated) | - |
| `--description` | Route description text | - |

## Notes

- Network must be in valid CIDR notation (e.g., `10.0.0.0/16`)
- Lower metric values have higher priority (metric 10 > metric 100)
- Masquerading enables NAT for outbound traffic
- Routes can use either a single peer or peer groups for redundancy

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
| [DNS](dns.md) | DNS nameserver groups and settings |
| [Posture Checks](posture-checks.md) | Device compliance validation |
| [Events](events.md) | Audit logs and traffic monitoring |
| [Geo-Locations](geo-locations.md) | Geographic location data |
| [Accounts](accounts.md) | Account settings and configuration |
| [Ingress Ports](ingress-ports.md) | Port forwarding (Cloud-only) |
| [Export & Import](export-import.md) | YAML/JSON configuration management |
| [Migrate](migrate.md) | Migration between NetBird accounts |

---

[Home](../README.md) | [Policies](policies.md) | **Routes** | [DNS](dns.md) | [Posture Checks](posture-checks.md) | [Events](events.md)
