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

# Create a domain-based route (comma-separated domain list instead of a CIDR)
netbird-manage route --create "app.example.com,api.example.com" \
  --network-id <network-id> \
  --peer <peer-id> \
  --groups <group-id> \
  --keep-route true

# Create a route with access control groups
netbird-manage route --create "10.10.0.0/16" \
  --network-id <network-id> \
  --peer <peer-id> \
  --groups <group-id> \
  --access-control-groups "acl-group-1,acl-group-2"

# Create an exit node route that clients must opt into
netbird-manage route --create "0.0.0.0/0" \
  --network-id <network-id> \
  --peer <peer-id> \
  --groups <group-id> \
  --skip-auto-apply true

# Update route metric (priority)
netbird-manage route --update <route-id> --metric 50

# Replace the domain list on a domain-based route
netbird-manage route --update <route-id> --domains "app.example.com,new.example.com"

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
| `--groups` | Distribution group IDs (required, comma-separated) | - |
| `--access-control-groups` | Access control group IDs (optional, comma-separated) | - |
| `--keep-route` | Keep routes for resolved domain IPs: `true` or `false` | - |
| `--skip-auto-apply` | Skip auto-applying an exit node route on clients: `true` or `false` | - |
| `--domains` | Replace a route's domain list (comma-separated, update only) | - |
| `--description` | Route description text | - |

## Notes

- `--create` accepts either a network in CIDR notation (e.g., `10.0.0.0/16`) or a comma-separated domain list (e.g., `app.example.com,api.example.com`) for domain-based routes
- `--update` supports the same flags as `--create`, plus `--domains` to replace a domain-based route's domain list
- `--keep-route` keeps routes for previously resolved domain IPs even after the DNS answer changes
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
| [DNS Zones](dns-zones.md) | Custom DNS zones and records |
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

[Home](../README.md) | [Policies](policies.md) | **Routes** | [DNS](dns.md) | [Posture Checks](posture-checks.md) | [Events](events.md)
